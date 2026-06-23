package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/http/handlers"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/store"
)

type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

func NewRouter(log *slog.Logger, readiness ReadinessChecker, dataStore *store.Store, cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	tokenManager := auth.NewTokenManager(cfg.Token.Secret, cfg.Token.AccessTokenTTL)
	authService := auth.NewService(dataStore.Users, tokenManager)
	accessService := restaurants.NewService(dataStore.Restaurants, dataStore.Memberships)
	demoService := demos.NewService(dataStore.Demos, accessService, cfg.Demo.TokenTTL)

	authHandler := handlers.NewAuthHandler(authService, cfg.App.Env, writeJSON, writeError)
	adminHandler := handlers.NewAdminHandler(dataStore.Users, writeJSON, writeError)
	restaurantHandler := handlers.NewRestaurantHandler(accessService, writeJSON, writeError)
	demoPublicHandler := handlers.NewDemoPublicHandler(demoService, writeJSON, writeError)
	demoAdminHandler := handlers.NewDemoAdminHandler(demoService, writeJSON, writeError)

	mux.HandleFunc("POST /api/v1/auth/signup", authHandler.Signup)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	protectAuthenticated := RequireAuth(tokenManager)
	protectDeveloper := func(next http.Handler) http.Handler {
		return protectAuthenticated(RequireRole(auth.RoleDeveloper)(next))
	}
	protectInternalAdmin := func(next http.Handler) http.Handler {
		return protectAuthenticated(RequireRole(auth.RoleInternalAdmin)(next))
	}
	protectRestaurantScoped := func(next http.Handler) http.Handler {
		return protectAuthenticated(RequireAnyRole(auth.RoleInternalAdmin, auth.RoleRestaurantOwner)(
			RequireRestaurantAccess(accessService)(next),
		))
	}
	protectRestaurantAdmin := func(next http.Handler) http.Handler {
		return protectAuthenticated(RequireRole(auth.RoleInternalAdmin)(
			RequireRestaurantAccess(accessService)(next),
		))
	}

	mux.Handle("GET /api/v1/auth/me", protectAuthenticated(http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /healthz", protectDeveloper(http.HandlerFunc(healthz(cfg))))
	mux.Handle("GET /readyz", protectDeveloper(http.HandlerFunc(readyz(readiness))))
	mux.Handle("GET /api/v1/admin/me", protectInternalAdmin(http.HandlerFunc(adminHandler.Me)))

	mux.Handle("GET /api/v1/restaurants", protectAuthenticated(RequireAnyRole(auth.RoleInternalAdmin, auth.RoleRestaurantOwner)(
		http.HandlerFunc(restaurantHandler.List),
	)))
	mux.Handle("POST /api/v1/restaurants", protectInternalAdmin(http.HandlerFunc(restaurantHandler.Create)))
	mux.Handle("GET /api/v1/restaurants/{id}", protectRestaurantScoped(http.HandlerFunc(restaurantHandler.Get)))
	mux.Handle("PATCH /api/v1/restaurants/{id}", protectRestaurantAdmin(http.HandlerFunc(restaurantHandler.Update)))
	mux.Handle("PATCH /api/v1/restaurants/{id}/status", protectRestaurantAdmin(http.HandlerFunc(restaurantHandler.UpdateStatus)))
	mux.Handle("DELETE /api/v1/restaurants/{id}", protectRestaurantAdmin(http.HandlerFunc(restaurantHandler.Archive)))
	mux.Handle("GET /api/v1/restaurants/{id}/members", protectRestaurantAdmin(http.HandlerFunc(restaurantHandler.ListMembers)))
	mux.Handle("POST /api/v1/restaurants/{id}/members", protectRestaurantAdmin(http.HandlerFunc(restaurantHandler.AddMember)))
	mux.Handle("POST /api/v1/restaurants/{id}/demo-sites", protectRestaurantAdmin(http.HandlerFunc(demoAdminHandler.Create)))

	mux.HandleFunc("GET /api/public/v1/demo/{slug}", demoPublicHandler.Get)

	var handler http.Handler = mux
	handler = Recovery(log)(handler)
	handler = CORS(cfg.HTTP.CORSAllowedOrigins)(handler)
	handler = AccessLog(log)(handler)
	handler = RequestID()(handler)

	return handler
}

func healthz(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"service": cfg.App.Name,
			"env":     cfg.App.Env,
			"version": cfg.App.Version,
		})
	}
}

func readyz(readiness ReadinessChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if readiness == nil {
			writeError(w, http.StatusServiceUnavailable, "database_not_configured", "Database readiness is not configured.")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := readiness.Ping(ctx); err != nil {
			code := "database_unavailable"
			message := "Database is not ready."
			if errors.Is(err, db.ErrNotConfigured) {
				code = "database_not_configured"
				message = "Database URL is not configured."
			}
			writeError(w, http.StatusServiceUnavailable, code, message)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status":   "ok",
			"database": "ready",
		})
	}
}
