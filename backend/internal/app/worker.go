package app

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/jobs"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/leadprep"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
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

	leadPreparationService := leadprep.NewService(dataStore.Pool(), cfg.Demo.TokenTTL, cfg.AppURLs)
	if err := worker.Register(jobs.LeadPrepareJobType, jobs.LeadPrepareHandler(leadPreparationService, log)); err != nil {
		return nil, err
	}

	accessService := restaurants.NewService(dataStore.Restaurants, dataStore.Memberships)
	campaignService := campaigns.NewService(dataStore.Campaigns, dataStore.Demos, accessService, &jobs.CampaignEnqueuer{Queue: queue}, cfg.AppURLs, cfg.Demo.TokenTTL)
	emailProvider, err := emailprovider.NewFromConfig(cfg.Email, cfg.ZohoMail)
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

	outreachRepo := outreach.NewPostgres(dataStore.Pool())
	outreachAccountPool, outreachPoolErr := emailprovider.NewPersistentAccountPoolFromConfig(
		ctx,
		cfg.Email,
		cfg.Outreach,
		outreachRepo,
	)
	if outreachPoolErr != nil {
		log.WarnContext(ctx, "outreach_account_pool_unavailable", "error", outreachPoolErr)
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
		nil,
		log,
	)
	if err := worker.Register(jobs.OutreachBulkSendJobType, jobs.OutreachBulkSendHandler(jobs.OutreachBulkSendDeps{
		Outreach: outreachService,
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
