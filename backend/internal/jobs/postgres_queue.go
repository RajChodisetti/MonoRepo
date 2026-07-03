package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresQueue struct {
	pool         *pgxpool.Pool
	jobs         chan Job
	pollInterval time.Duration
}

func NewPostgresQueue(pool *pgxpool.Pool, bufferSize int, pollInterval time.Duration) *PostgresQueue {
	if bufferSize < 1 {
		bufferSize = 1
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	return &PostgresQueue{
		pool:         pool,
		jobs:         make(chan Job, bufferSize),
		pollInterval: pollInterval,
	}
}

func (queue *PostgresQueue) Enqueue(ctx context.Context, job Job) (Job, error) {
	if job.Type == "" {
		return Job{}, fmt.Errorf("job type is required")
	}
	if len(job.Payload) == 0 {
		job.Payload = json.RawMessage(`{}`)
	}
	if job.MaxAttempts < 1 {
		job.MaxAttempts = 3
	}

	const query = `
		INSERT INTO job_runs (job_type, status, payload, idempotency_key, max_attempts)
		VALUES ($1, 'queued', $2, NULLIF($3, ''), $4)
		ON CONFLICT (idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING
		RETURNING id::text, attempts`

	var id string
	var attempts int
	err := queue.pool.QueryRow(ctx, query, job.Type, job.Payload, job.IdempotencyKey, job.MaxAttempts).Scan(&id, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return job, nil
	}
	if err != nil {
		return Job{}, fmt.Errorf("enqueue job: %w", err)
	}

	job.ID = id
	job.Attempts = attempts
	if job.EnqueuedAt.IsZero() {
		job.EnqueuedAt = time.Now().UTC()
	}
	return job, nil
}

func (queue *PostgresQueue) Jobs() <-chan Job {
	return queue.jobs
}

func (queue *PostgresQueue) StartPoller(ctx context.Context) {
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			job, err := queue.pollOne(ctx)
			if err != nil {
				time.Sleep(queue.pollInterval)
				continue
			}
			if job.Type == "" {
				time.Sleep(queue.pollInterval)
				continue
			}
			select {
			case queue.jobs <- job:
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (queue *PostgresQueue) pollOne(ctx context.Context) (Job, error) {
	const query = `
		UPDATE job_runs
		SET status = 'running',
			locked_at = now(),
			attempts = attempts + 1,
			updated_at = now()
		WHERE id = (
			SELECT id
			FROM job_runs
			WHERE status = 'queued'
				AND available_at <= now()
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id::text, job_type, payload, attempts, max_attempts, idempotency_key, created_at`

	var job Job
	var idempotencyKey *string
	err := queue.pool.QueryRow(ctx, query).Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Attempts,
		&job.MaxAttempts,
		&idempotencyKey,
		&job.EnqueuedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Job{}, nil
	}
	if err != nil {
		return Job{}, err
	}
	if idempotencyKey != nil {
		job.IdempotencyKey = *idempotencyKey
	}
	return job, nil
}

func (queue *PostgresQueue) Complete(ctx context.Context, job Job) error {
	const query = `
		UPDATE job_runs
		SET status = 'completed', locked_at = NULL, updated_at = now()
		WHERE id = $1::uuid`
	_, err := queue.pool.Exec(ctx, query, job.ID)
	return err
}

func (queue *PostgresQueue) Fail(ctx context.Context, job Job, jobErr error, retryDelay time.Duration) error {
	if job.Attempts < job.MaxAttempts {
		const query = `
			UPDATE job_runs
			SET status = 'queued',
				last_error = $2,
				available_at = now() + ($3 * interval '1 second'),
				locked_at = NULL,
				updated_at = now()
			WHERE id = $1::uuid`
		_, err := queue.pool.Exec(ctx, query, job.ID, jobErr.Error(), int(retryDelay.Seconds()))
		return err
	}

	const query = `
		UPDATE job_runs
		SET status = 'failed',
			last_error = $2,
			locked_at = NULL,
			updated_at = now()
		WHERE id = $1::uuid`
	_, err := queue.pool.Exec(ctx, query, job.ID, jobErr.Error())
	return err
}

var _ Queue = (*PostgresQueue)(nil)
