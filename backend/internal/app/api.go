package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	httpapi "github.com/rajchodisetti/restaurant-platform/backend/internal/http"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/store"
)

type API struct {
	cfg      config.Config
	log      *slog.Logger
	database *db.DB
	store    *store.Store
	fiberApp *fiber.App
}

func NewAPI(ctx context.Context) (*API, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.Logging)
	database, err := db.ConnectRequiredLogged(ctx, log, cfg.Database, databaseReadyTimeout)
	if err != nil {
		return nil, err
	}

	dataStore := store.New(database)
	if err := dataStore.VerifyStartup(ctx); err != nil {
		db.CloseLogged(ctx, log, database)
		return nil, err
	}

	router := httpapi.NewRouter(log, database, cfg)
	fiberApp := fiber.New()
	fiberApp.Use(adaptor.HTTPHandler(router))

	return &API{
		cfg:      cfg,
		log:      log,
		database: database,
		store:    dataStore,
		fiberApp: fiberApp,
	}, nil
}

func (api *API) Run(ctx context.Context) error {
	listenErr := make(chan error, 1)
	go func() {
		api.log.InfoContext(ctx, "api_starting", "addr", api.cfg.HTTP.Addr, "env", api.cfg.App.Env)
		listenErr <- api.fiberApp.Listen(api.cfg.HTTP.Addr)
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-shutdownSignals:
		api.log.InfoContext(ctx, "api_shutting_down", "signal", sig.String())
	case err := <-listenErr:
		db.CloseLogged(ctx, api.log, api.database)
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := api.fiberApp.ShutdownWithContext(shutdownCtx); err != nil {
		api.log.ErrorContext(ctx, "api_shutdown_failed", "error", err)
		db.CloseLogged(ctx, api.log, api.database)
		return err
	}

	db.CloseLogged(ctx, api.log, api.database)
	api.log.InfoContext(ctx, "api_stopped")
	return nil
}
