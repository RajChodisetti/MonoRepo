package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	httpapi "github.com/rajchodisetti/restaurant-platform/backend/internal/http"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
)

type API struct {
	cfg    config.Config
	log    *slog.Logger
	db     *db.DB
	server *http.Server
}

func NewAPI(ctx context.Context) (*API, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.Logging)
	database, err := db.Connect(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	router := httpapi.NewRouter(log, database, cfg)
	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &API{
		cfg:    cfg,
		log:    log,
		db:     database,
		server: server,
	}, nil
}

func (a *API) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() {
		a.log.InfoContext(ctx, "api_starting", "addr", a.cfg.HTTP.Addr, "env", a.cfg.App.Env)
		errs <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			a.log.ErrorContext(ctx, "api_shutdown_failed", "error", err)
			return err
		}
		a.db.Close()
		a.log.InfoContext(ctx, "api_stopped")
		return nil
	case err := <-errs:
		a.db.Close()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
