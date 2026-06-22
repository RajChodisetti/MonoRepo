package app

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/jobs"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
)

type WorkerApp struct {
	cfg    config.Config
	log    *slog.Logger
	db     *db.DB
	queue  *jobs.InMemoryQueue
	worker *jobs.Worker
}

func NewWorker(ctx context.Context) (*WorkerApp, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.Logging)
	database, err := db.Connect(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	queue := jobs.NewInMemoryQueue(cfg.Jobs.BufferSize)
	worker := jobs.NewWorker(queue, log, cfg.Jobs.RetryDelay)
	if err := worker.Register(jobs.SampleJobType, jobs.SampleHandler(log)); err != nil {
		return nil, err
	}

	return &WorkerApp{
		cfg:    cfg,
		log:    log,
		db:     database,
		queue:  queue,
		worker: worker,
	}, nil
}

func (w *WorkerApp) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sample, err := jobs.NewSampleJob("worker booted")
	if err != nil {
		return err
	}
	if _, err := w.queue.Enqueue(ctx, sample); err != nil {
		return err
	}

	w.log.InfoContext(ctx, "worker_starting", "env", w.cfg.App.Env)
	err = w.worker.Run(ctx)
	w.db.Close()
	if errors.Is(err, context.Canceled) {
		w.log.InfoContext(ctx, "worker_stopped")
		return nil
	}
	return err
}
