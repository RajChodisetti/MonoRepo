package app

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/jobs"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/store"
)

type WorkerApp struct {
	cfg    config.Config
	log    *slog.Logger
	db     *db.DB
	store  *store.Store
	queue  *jobs.PostgresQueue
	worker *jobs.Worker
}

func NewWorker(ctx context.Context) (*WorkerApp, error) {
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

	queue := jobs.NewPostgresQueue(dataStore.Pool(), cfg.Jobs.BufferSize, cfg.Jobs.RetryDelay)
	worker := jobs.NewWorker(queue, log, cfg.Jobs.RetryDelay)

	if err := worker.Register(jobs.SampleJobType, jobs.SampleHandler(log)); err != nil {
		return nil, err
	}

	accessService := restaurants.NewService(dataStore.Restaurants, dataStore.Memberships)
	campaignService := campaigns.NewService(dataStore.Campaigns, dataStore.Demos, accessService, &jobs.CampaignEnqueuer{Queue: queue}, cfg.AppURLs)
	emailProvider, err := emailprovider.NewFromConfig(cfg.Email, cfg.SMTP)
	if err != nil {
		return nil, err
	}
	if err := worker.Register(jobs.EmailSendJobType, jobs.EmailSendHandler(jobs.EmailSendDeps{
		Campaigns:        dataStore.Campaigns,
		CampaignsService: campaignService,
		Email:            emailProvider,
		EmailCfg:         cfg.Email,
		AppURLs:          cfg.AppURLs,
	}, log)); err != nil {
		return nil, err
	}

	return &WorkerApp{
		cfg:    cfg,
		log:    log,
		db:     database,
		store:  dataStore,
		queue:  queue,
		worker: worker,
	}, nil
}

func (w *WorkerApp) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	w.queue.StartPoller(ctx)

	sample, err := jobs.NewSampleJob("worker booted")
	if err != nil {
		return err
	}
	if _, err := w.queue.Enqueue(ctx, sample); err != nil {
		return err
	}

	w.log.InfoContext(ctx, "worker_starting", "env", w.cfg.App.Env)
	err = w.worker.Run(ctx)
	db.CloseLogged(ctx, w.log, w.db)
	if errors.Is(err, context.Canceled) {
		w.log.InfoContext(ctx, "worker_stopped")
		return nil
	}
	return err
}
