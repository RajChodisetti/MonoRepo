package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultJobLease = 15 * time.Minute

var (
	ErrActiveBulkJob = errors.New("an active bulk outreach job already exists")
	ErrJobLeaseLost  = errors.New("job lease is no longer owned by this worker")
)

type PostgresQueue struct {
	pool         *pgxpool.Pool
	jobs         chan Job
	pollInterval time.Duration
	workerID     string
	lease        time.Duration
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
		workerID:     "go-worker-" + newJobID(),
		lease:        defaultJobLease,
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "idx_job_runs_one_active_bulk_outreach" {
			return Job{}, ErrActiveBulkJob
		}
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
	if queue.pool == nil {
		return Job{}, fmt.Errorf("database pool is not configured")
	}
	if strings.TrimSpace(queue.workerID) == "" {
		return Job{}, fmt.Errorf("job worker identity is not configured")
	}

	const expireExhausted = `
		UPDATE job_runs
		SET status = 'failed',
		    last_error = COALESCE(NULLIF(last_error, ''), 'Worker lease expired after maximum attempts.'),
		    locked_at = NULL,
		    locked_by = NULL,
		    lease_expires_at = NULL,
		    updated_at = now()
		WHERE status = 'running'
		  AND lease_expires_at <= now()
		  AND attempts >= max_attempts`
	if _, err := queue.pool.Exec(ctx, expireExhausted); err != nil {
		return Job{}, fmt.Errorf("expire exhausted job leases: %w", err)
	}

	const query = `
		UPDATE job_runs AS job
		SET status = 'running',
			locked_at = now(),
			locked_by = $1,
			lease_expires_at = now() + ($2 * interval '1 second'),
			attempts = attempts + 1,
			updated_at = now()
		WHERE job.id = (
			SELECT candidate.id
			FROM job_runs AS candidate
			WHERE (
				(candidate.status = 'queued' AND candidate.available_at <= now())
				OR (candidate.status = 'running' AND candidate.lease_expires_at <= now())
			)
				AND candidate.attempts < candidate.max_attempts
				AND NOT EXISTS (
					SELECT 1
					FROM job_runs AS owned
					WHERE owned.status = 'running'
					  AND owned.locked_by = $1
					  AND owned.lease_expires_at > now()
				)
			ORDER BY CASE candidate.status WHEN 'running' THEN 0 ELSE 1 END,
			         candidate.available_at,
			         candidate.created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING job.id::text, job.job_type, job.payload, job.attempts,
		          job.max_attempts, job.idempotency_key, job.created_at, job.locked_by`

	var job Job
	var idempotencyKey *string
	err := queue.pool.QueryRow(ctx, query, queue.workerID, int64(queue.lease/time.Second)).Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Attempts,
		&job.MaxAttempts,
		&idempotencyKey,
		&job.EnqueuedAt,
		&job.LockedBy,
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
		SET status = 'completed', locked_at = NULL, locked_by = NULL,
		    lease_expires_at = NULL, updated_at = now()
		WHERE id = $1::uuid AND status = 'running' AND locked_by = $2`
	result, err := queue.pool.Exec(ctx, query, job.ID, job.LockedBy)
	return leaseMutationResult(result.RowsAffected(), err)
}

func (queue *PostgresQueue) Fail(ctx context.Context, job Job, jobErr error, retryDelay time.Duration) error {
	if job.Attempts < job.MaxAttempts {
		const query = `
			UPDATE job_runs
			SET status = 'queued',
				last_error = $2,
				available_at = now() + ($3 * interval '1 second'),
				locked_at = NULL,
				locked_by = NULL,
				lease_expires_at = NULL,
				updated_at = now()
			WHERE id = $1::uuid AND status = 'running' AND locked_by = $4`
		result, err := queue.pool.Exec(ctx, query, job.ID, jobErr.Error(), int(retryDelay.Seconds()), job.LockedBy)
		return leaseMutationResult(result.RowsAffected(), err)
	}

	const query = `
		UPDATE job_runs
		SET status = 'failed',
			last_error = $2,
			locked_at = NULL,
			locked_by = NULL,
			lease_expires_at = NULL,
			updated_at = now()
		WHERE id = $1::uuid AND status = 'running' AND locked_by = $3`
	result, err := queue.pool.Exec(ctx, query, job.ID, jobErr.Error(), job.LockedBy)
	return leaseMutationResult(result.RowsAffected(), err)
}

func (queue *PostgresQueue) RenewLease(ctx context.Context, job Job) error {
	const query = `
		UPDATE job_runs
		SET locked_at = now(),
		    lease_expires_at = now() + ($3 * interval '1 second'),
		    updated_at = now()
		WHERE id = $1::uuid AND status = 'running' AND locked_by = $2`
	result, err := queue.pool.Exec(ctx, query, job.ID, job.LockedBy, int64(queue.lease/time.Second))
	return leaseMutationResult(result.RowsAffected(), err)
}

func (queue *PostgresQueue) LeaseHeartbeatInterval() time.Duration {
	return queue.lease / 3
}

func leaseMutationResult(rows int64, err error) error {
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrJobLeaseLost
	}
	return nil
}

var _ Queue = (*PostgresQueue)(nil)
var _ LeasingQueue = (*PostgresQueue)(nil)
