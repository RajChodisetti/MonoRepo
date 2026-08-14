package app

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/jobs"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreachaccounts"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/store"
)

type WorkerApp struct {
	cfg           config.Config
	log           *slog.Logger
	db            *db.DB
	store         *store.Store
	queue         *jobs.PostgresQueue
	worker        *jobs.Worker
	emailHealth   emailprovider.HealthMonitor
	accountLoader emailprovider.OutreachConfigLoader
	outreachRepo  *outreach.Postgres
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
	campaignService := campaigns.NewService(dataStore.Campaigns, dataStore.Demos, accessService, &jobs.CampaignEnqueuer{Queue: queue}, cfg.AppURLs, cfg.Demo.TokenTTL)
	emailProvider, err := emailprovider.NewFromConfig(cfg.Email, cfg.ZohoMail)
	if err != nil {
		return nil, err
	}
	outreachRepo := outreach.NewPostgres(dataStore.Pool())
	accountStore := outreachaccounts.NewPostgres(dataStore.Pool())
	accountLoader := outreachaccounts.NewService(accountStore, cfg.Outreach, cfg.Outreach.CredentialEncryptionKey, log)
	emailHealth := emailprovider.NewReloadingHealthService(cfg.Email, accountLoader, outreachRepo)
	outreachAccountPool := emailprovider.NewReloadingPersistentAccountPool(cfg.Email, accountLoader, outreachRepo)
	outreachService := outreach.NewService(
		outreachRepo,
		dataStore.Pool(),
		dataStore.Campaigns,
		campaignService,
		accessService,
		outreach.DemoTokenResolver{Campaigns: dataStore.Campaigns, Demos: dataStore.Demos},
		outreachAccountPool,
		emailProvider,
		cfg.Email,
		cfg.Outreach,
		cfg.AppURLs,
		nil,
		log,
	)
	if err := worker.Register(jobs.OutreachBulkSendJobType, jobs.OutreachBulkSendHandler(jobs.OutreachBulkSendDeps{
		Outreach: outreachService,
	}, log)); err != nil {
		return nil, err
	}

	return &WorkerApp{
		cfg:           cfg,
		log:           log,
		db:            database,
		store:         dataStore,
		queue:         queue,
		worker:        worker,
		emailHealth:   emailHealth,
		accountLoader: accountLoader,
		outreachRepo:  outreachRepo,
	}, nil
}

func (w *WorkerApp) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	w.queue.StartPoller(ctx)
	go w.runEmailHealthChecks(ctx)
	if w.cfg.Outreach.InboundEnabled {
		go w.runInboundPoll(ctx)
	}

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

func (w *WorkerApp) runEmailHealthChecks(ctx context.Context) {
	if w.emailHealth == nil {
		return
	}
	run := func() {
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		if err := w.emailHealth.RunDue(checkCtx); err != nil {
			w.log.ErrorContext(ctx, "gmail_health_check_failed", "error", err)
		}
	}
	run()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func (w *WorkerApp) runInboundPoll(ctx context.Context) {
	if w.accountLoader == nil || w.outreachRepo == nil {
		return
	}
	interval := w.cfg.Outreach.InboundPollInterval
	if interval <= 0 {
		interval = time.Minute
	}
	run := func() {
		w.pollInboundMailboxes(ctx)
	}
	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func (w *WorkerApp) pollInboundMailboxes(ctx context.Context) {
	effective, err := w.accountLoader.Load(ctx)
	if err != nil {
		w.log.ErrorContext(ctx, "outreach_inbound_accounts_unavailable", "error", err)
		return
	}
	if !effective.InboundEnabled {
		return
	}

	var wait sync.WaitGroup
	for _, mailbox := range effective.InboundMailboxes {
		mailbox := mailbox
		wait.Add(1)
		go func() {
			defer wait.Done()
			reader, inboxErr := emailprovider.NewGmailInbox(w.cfg.Email, mailbox)
			if inboxErr != nil {
				w.log.WarnContext(ctx, "outreach_inbound_mailbox_unavailable", "mailbox_key", mailbox.AccountKey, "error", inboxErr)
				return
			}
			inbound := outreach.NewInboundService(
				w.outreachRepo,
				w.store.Campaigns,
				reader,
				mailbox.AccountKey,
				effective,
				w.log.With("mailbox_key", mailbox.AccountKey),
			)
			checkCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			defer cancel()
			if pollErr := inbound.Poll(checkCtx); pollErr != nil {
				w.log.ErrorContext(ctx, "outreach_inbound_poll_failed", "mailbox_key", mailbox.AccountKey, "error", pollErr)
			}
		}()
	}
	wait.Wait()
}
