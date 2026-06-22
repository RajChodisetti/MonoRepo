package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
)

type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

func NewRouter(log *slog.Logger, readiness ReadinessChecker, cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz(cfg))
	mux.HandleFunc("GET /readyz", readyz(readiness, cfg))

	var handler http.Handler = mux
	handler = Recovery(log)(handler)
	handler = CORS(cfg.HTTP.CORSAllowedOrigins)(handler)
	handler = AccessLog(log)(handler)
	handler = RequestID()(handler)

	return handler
}

func healthz(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !allowHealthEndpoints(cfg, w) {
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"service": cfg.App.Name,
			"env":     cfg.App.Env,
			"version": cfg.App.Version,
		})
	}
}

func readyz(readiness ReadinessChecker, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !allowHealthEndpoints(cfg, w) {
			return
		}

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
