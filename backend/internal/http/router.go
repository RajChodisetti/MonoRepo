package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/consultations"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/http/handlers"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/jobs"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/leadreview"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	calendarprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/calendar"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
	llmprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/llm"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/reservations"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/scrapejobs"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/seoreport"
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

	jobQueue := jobs.NewPostgresQueue(dataStore.Pool(), cfg.Jobs.BufferSize, cfg.Jobs.RetryDelay)
	var jobEnqueuer campaigns.SendJobEnqueuer = &jobs.CampaignEnqueuer{Queue: jobQueue}
	if dataStore.Pool() == nil {
		jobEnqueuer = &jobs.CampaignEnqueuer{Queue: jobs.NewInMemoryQueue(cfg.Jobs.BufferSize)}
	}
	campaignService := campaigns.NewService(dataStore.Campaigns, dataStore.Demos, accessService, jobEnqueuer, cfg.AppURLs, cfg.Demo.TokenTTL)
	calendarProvider := calendarprovider.NewFromConfig(context.Background(), cfg.Consultations, log)
	emailProvider, err := emailprovider.NewFromConfig(cfg.Email, cfg.ZohoMail)
	if err != nil {
		log.ErrorContext(context.Background(), "consultation_email_provider_unavailable", "error", err)
		emailProvider = emailprovider.NewDisabled()
	}
	consultationService := consultations.NewService(cfg.Consultations, dataStore.Consultations, calendarProvider, emailProvider, log)

	authHandler := handlers.NewAuthHandler(authService, cfg.App.Env, writeJSON, writeError)
	adminHandler := handlers.NewAdminHandler(dataStore.Users, writeJSON, writeError)
	userHandler := handlers.NewUserHandler(dataStore.Users, dataStore.Restaurants, dataStore.Memberships, writeJSON, writeError)
	restaurantHandler := handlers.NewRestaurantHandler(accessService, writeJSON, writeError)
	demoPublicHandler := handlers.NewDemoPublicHandler(demoService, writeJSON, writeError)
	demoAdminHandler := handlers.NewDemoAdminHandler(demoService, writeJSON, writeError)
	campaignHandler := handlers.NewCampaignHandler(campaignService, writeJSON, writeError)
	outreachRepo := outreach.NewPostgres(dataStore.Pool())
	outreachAccountPool, outreachPoolErr := emailprovider.NewPersistentAccountPoolFromConfig(
		context.Background(),
		cfg.Email,
		cfg.Outreach,
		outreachRepo,
	)
	if outreachPoolErr != nil {
		log.WarnContext(context.Background(), "outreach_account_pool_unavailable", "error", outreachPoolErr)
	}
	outreachService := outreach.NewService(
		outreachRepo,
		dataStore.Pool(),
		dataStore.Campaigns,
		campaignService,
		outreach.DemoTokenResolver{Campaigns: dataStore.Campaigns, Demos: dataStore.Demos},
		outreachAccountPool,
		cfg.Email,
		cfg.Outreach,
		&jobs.OutreachBulkEnqueuer{Queue: jobQueue},
		log,
	)
	outreachBulkHandler := handlers.NewOutreachBulkHandler(outreachService, writeJSON, writeError)
	scrapeJobRepo := scrapejobs.NewPostgres(dataStore.Pool())
	scrapeJobService := scrapejobs.NewService(scrapeJobRepo)
	scrapeJobHandler := handlers.NewScrapeJobHandler(scrapeJobService, writeJSON, writeError)
	leadReviewService := leadreview.NewService(dataStore.Pool())
	leadReviewHandler := handlers.NewLeadReviewHandler(leadReviewService, writeJSON, writeError)
	trackingHandler := handlers.NewTrackingHandler(dataStore.Campaigns, writeError)
	restaurantPublicHandler := handlers.NewRestaurantPublicHandler(dataStore.Profiles, writeJSON, writeError)
	reservationService := reservations.NewService(dataStore.Reservations)
	reservationPublicHandler := handlers.NewReservationPublicHandler(reservationService, writeJSON, writeError)
	companyConsultationHandler := handlers.NewCompanyConsultationHandler(consultationService, writeJSON)
	var interestedRepo seoreport.InterestedRepository
	var leadUpserter seoreport.LeadUpserter
	if pool := dataStore.Pool(); pool != nil {
		interestedRepo = seoreport.NewInterestedPostgres(pool)
		leadUpserter = seoreport.NewLeadUpserter(pool)
	}
	// Prefer the same Google Workspace mailbox used for outreach sending.
	seoMailer := emailProvider
	if len(cfg.Outreach.GoogleWorkspaceAccounts) > 0 {
		gmailMailer, gmailErr := emailprovider.NewGmail(cfg.Email, cfg.Outreach.GoogleWorkspaceAccounts[0])
		if gmailErr != nil {
			log.WarnContext(context.Background(), "seo_gmail_mailer_unavailable", "error", gmailErr)
		} else {
			seoMailer = gmailMailer
			log.InfoContext(context.Background(), "seo_unlock_mailer", "provider", "gmail", "from", cfg.Outreach.GoogleWorkspaceAccounts[0].FromEmail)
		}
	}
	seoService := seoreport.NewServiceFull(
		cfg.Places,
		cfg.App,
		cfg.AppURLs,
		dataStore.Profiles,
		interestedRepo,
		leadUpserter,
		seoMailer,
		llmprovider.NewFromConfig(cfg.LLM),
		log,
	)
	seoPublicHandler := handlers.NewSEOPublicHandler(seoService, cfg.AppURLs.PublicWebURL, writeJSON, writeError)

	mux.HandleFunc("POST /api/v1/auth/signup", authHandler.Signup)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	protectAuthenticated := RequireAuth(tokenManager)
	protectDeveloper := func(next http.Handler) http.Handler {
		return protectAuthenticated(RequireRole(auth.RoleDeveloper)(next))
	}
	protectInternalAdmin := func(next http.Handler) http.Handler {
		return protectAuthenticated(RequireRole(auth.RoleInternalAdmin)(next))
	}
	protectRestaurantUser := func(next http.Handler) http.Handler {
		return protectAuthenticated(RequireAnyRole(auth.RoleRestaurantOwner, auth.RoleDeveloper)(next))
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
	protectConsultationAPI := RequireStaticBearerToken(cfg.Consultations.APIToken)

	mux.Handle("GET /api/v1/auth/me", protectAuthenticated(http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /healthz", protectDeveloper(http.HandlerFunc(healthz(cfg))))
	mux.Handle("GET /readyz", protectDeveloper(http.HandlerFunc(readyz(readiness))))
	mux.Handle("GET /api/v1/admin/me", protectInternalAdmin(http.HandlerFunc(adminHandler.Me)))
	mux.Handle("GET /api/v1/user/me", protectRestaurantUser(http.HandlerFunc(userHandler.Me)))

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
	mux.Handle("GET /api/v1/demo-sites/{id}/review-preview", protectInternalAdmin(http.HandlerFunc(demoAdminHandler.ReviewPreview)))
	mux.Handle("GET /api/v1/restaurants/{id}/profile/review-preview", protectRestaurantAdmin(http.HandlerFunc(leadReviewHandler.GetProfileReviewPreview)))
	mux.Handle("PATCH /api/v1/restaurants/{id}/profile/review", protectRestaurantAdmin(http.HandlerFunc(leadReviewHandler.ReviewProfile)))
	mux.Handle("PATCH /api/v1/demo-sites/{id}/status", protectInternalAdmin(http.HandlerFunc(leadReviewHandler.SetDemoStatus)))
	mux.Handle("POST /api/v1/restaurants/{id}/campaigns", protectRestaurantAdmin(http.HandlerFunc(campaignHandler.Create)))
	mux.Handle("GET /api/v1/restaurants/{id}/campaigns", protectRestaurantAdmin(http.HandlerFunc(campaignHandler.List)))
	mux.Handle("GET /api/v1/campaigns/{id}", protectInternalAdmin(http.HandlerFunc(campaignHandler.Get)))
	mux.Handle("POST /api/v1/campaigns/{id}/approve", protectInternalAdmin(http.HandlerFunc(campaignHandler.Approve)))
	mux.Handle("POST /api/v1/campaigns/{id}/regenerate", protectInternalAdmin(http.HandlerFunc(campaignHandler.Regenerate)))
	mux.Handle("POST /api/v1/campaigns/{id}/stop", protectInternalAdmin(http.HandlerFunc(campaignHandler.Stop)))
	mux.Handle("POST /api/v1/outreach/bulk-send", protectInternalAdmin(http.HandlerFunc(outreachBulkHandler.Trigger)))
	mux.Handle("GET /api/v1/outreach/bulk-send/status", protectInternalAdmin(http.HandlerFunc(outreachBulkHandler.Status)))
	mux.Handle("POST /api/v1/scrape-jobs", protectInternalAdmin(http.HandlerFunc(scrapeJobHandler.Trigger)))
	mux.Handle("GET /api/v1/scrape-jobs", protectInternalAdmin(http.HandlerFunc(scrapeJobHandler.List)))
	mux.Handle("GET /api/v1/scrape-jobs/{id}", protectInternalAdmin(http.HandlerFunc(scrapeJobHandler.Get)))
	mux.Handle("POST /api/v1/scrape-jobs/{id}/retry", protectInternalAdmin(http.HandlerFunc(scrapeJobHandler.Retry)))

	mux.HandleFunc("GET /t/click/{token}", trackingHandler.Click)
	mux.HandleFunc("GET /t/open/{token}", trackingHandler.Open)
	mux.HandleFunc("GET /t/unsubscribe/{token}", trackingHandler.Unsubscribe)

	mux.HandleFunc("GET /api/public/v1/demo/{slug}", demoPublicHandler.Get)
	mux.HandleFunc("GET /api/public/v1/restaurants/{id}/site-images", restaurantPublicHandler.GetSiteImagesByID)
	mux.HandleFunc("GET /api/public/v1/restaurants/by-place/{place_id}/site-images", restaurantPublicHandler.GetSiteImagesByPlaceID)
	mux.HandleFunc("GET /api/public/v1/site/restaurants", restaurantPublicHandler.ListSiteRestaurants)
	mux.HandleFunc("GET /api/public/v1/site/restaurants/{index}", restaurantPublicHandler.GetSiteContentByIndex)
	mux.HandleFunc("GET /api/public/v1/site/by-place/{place_id}", restaurantPublicHandler.GetSiteContentByPlaceID)
	mux.HandleFunc("GET /api/public/v1/seo/search", seoPublicHandler.Search)
	mux.HandleFunc("GET /api/public/v1/seo/report/{place_id}", seoPublicHandler.Report)
	mux.HandleFunc("POST /api/public/v1/seo/unlock/request", seoPublicHandler.RequestUnlock)
	mux.HandleFunc("POST /api/public/v1/seo/unlock/verify", seoPublicHandler.VerifyUnlock)
	mux.HandleFunc("GET /api/public/v1/seo/unlock/click/{token}", seoPublicHandler.ClickUnlock)
	mux.HandleFunc("GET /api/public/v1/restaurants/{id}/table-availability", reservationPublicHandler.GetTableAvailability)
	mux.HandleFunc("PUT /api/public/v1/restaurants/{id}/reservations", reservationPublicHandler.PutReservation)

	registerCompanyConsultationRoutes(mux, "/api/v1/company/consultations", protectConsultationAPI, companyConsultationHandler)
	registerCompanyConsultationRoutes(mux, "/api/v1/consultations", protectConsultationAPI, companyConsultationHandler)

	var handler http.Handler = mux
	handler = Recovery(log)(handler)
	handler = CORS(cfg.HTTP.CORSAllowedOrigins)(handler)
	handler = AccessLog(log)(handler)
	handler = RequestID()(handler)

	return handler
}

func registerCompanyConsultationRoutes(
	mux *http.ServeMux,
	prefix string,
	protect func(http.Handler) http.Handler,
	handler *handlers.CompanyConsultationHandler,
) {
	mux.Handle("GET "+prefix+"/availability", protect(http.HandlerFunc(handler.GetAvailability)))
	mux.Handle("GET "+prefix+"/availability/check", protect(http.HandlerFunc(handler.CheckAvailability)))
	mux.Handle("POST "+prefix, protect(http.HandlerFunc(handler.Book)))
}

func RequireStaticBearerToken(token string) func(http.Handler) http.Handler {
	expected := "Bearer " + token
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != expected {
				writeJSON(w, http.StatusUnauthorized, map[string]any{
					"status":  "error",
					"message": "unauthorized",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
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
