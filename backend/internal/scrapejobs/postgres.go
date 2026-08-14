package scrapejobs

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

const jobSelect = `
	SELECT
		j.id, j.city, j.city_key, j.niche, j.status, j.cycle_number,
		j.max_requests_per_window, j.requests_used_window, j.requests_used_total,
		j.window_started_at, j.resume_at, COALESCE(j.waiting_reason, ''),
		j.last_cycle_completed_at, j.current_cell_id, COALESCE(j.last_error, ''),
			j.created_at, j.updated_at,
			(SELECT count(*) FROM scrape_job_cells c WHERE c.scrape_job_id = j.id),
			(SELECT count(*) FROM scrape_job_cells c WHERE c.scrape_job_id = j.id AND c.status IN ('pending', 'running')),
			(SELECT count(*) FROM scrape_job_cells c WHERE c.scrape_job_id = j.id AND c.status = 'completed'),
			(SELECT count(*) FROM scrape_job_cells c WHERE c.scrape_job_id = j.id AND c.status = 'subdivided'),
			(SELECT count(*) FROM scrape_job_cells c WHERE c.scrape_job_id = j.id AND c.status = 'failed'),
			(SELECT count(*) FROM scrape_job_cells c WHERE c.scrape_job_id = j.id AND c.saturated = true AND c.status <> 'subdivided'),
			(SELECT count(*) FROM scrape_job_candidates c WHERE c.scrape_job_id = j.id),
		(SELECT count(*) FROM scrape_job_candidates c WHERE c.scrape_job_id = j.id AND c.status IN ('discovered', 'details_ready', 'enriched')),
		(SELECT count(*) FROM scrape_job_candidates c WHERE c.scrape_job_id = j.id AND c.status = 'imported'),
		(SELECT count(*) FROM scrape_job_candidates c WHERE c.scrape_job_id = j.id AND c.status = 'duplicate'),
		(SELECT count(*) FROM scrape_job_candidates c WHERE c.scrape_job_id = j.id AND c.status = 'failed')
	FROM scrape_jobs j`

func (repo *Postgres) CreateOrGetActive(
	ctx context.Context,
	city, cityKey, niche string,
	createdBy uuid.UUID,
) (Job, bool, error) {
	if repo.pool == nil {
		return Job{}, false, fmt.Errorf("database pool is not configured")
	}

	const insert = `
		INSERT INTO scrape_jobs (
			city, city_key, niche, status, max_requests_per_window, created_by
		)
		VALUES ($1, $2, $3, 'queued', $4, $5)
		ON CONFLICT (city_key, niche)
			WHERE status IN ('queued', 'running', 'waiting')
		DO NOTHING
		RETURNING id`

	var id uuid.UUID
	err := repo.pool.QueryRow(
		ctx,
		insert,
		city,
		cityKey,
		niche,
		DefaultMaxRequestsWindow,
		createdBy,
	).Scan(&id)
	created := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Job{}, false, fmt.Errorf("create scrape job: %w", err)
	}

	if !created {
		const active = `
			SELECT id
			FROM scrape_jobs
			WHERE city_key = $1 AND niche = $2
			  AND status IN ('queued', 'running', 'waiting')
			ORDER BY created_at DESC
			LIMIT 1`
		if err := repo.pool.QueryRow(ctx, active, cityKey, niche).Scan(&id); err != nil {
			return Job{}, false, fmt.Errorf("load active scrape job: %w", err)
		}
	}

	job, err := repo.GetByID(ctx, id)
	if err != nil {
		return Job{}, false, err
	}
	return job, created, nil
}

func (repo *Postgres) GetByID(ctx context.Context, id uuid.UUID) (Job, error) {
	if repo.pool == nil {
		return Job{}, fmt.Errorf("database pool is not configured")
	}
	row := repo.pool.QueryRow(ctx, jobSelect+` WHERE j.id = $1`, id)
	job, err := scanJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Job{}, ErrNotFound
	}
	if err != nil {
		return Job{}, fmt.Errorf("get scrape job: %w", err)
	}
	return job, nil
}

func (repo *Postgres) ListRecent(ctx context.Context, limit int) ([]Job, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}
	rows, err := repo.pool.Query(ctx, jobSelect+` ORDER BY j.created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list scrape jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]Job, 0)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, fmt.Errorf("scan scrape job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scrape jobs: %w", err)
	}
	return jobs, nil
}

func (repo *Postgres) ResumeFailed(ctx context.Context, id uuid.UUID) (Job, error) {
	if repo.pool == nil {
		return Job{}, fmt.Errorf("database pool is not configured")
	}

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return Job{}, fmt.Errorf("begin scrape job resume: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	var cityKey string
	var niche string
	if err := tx.QueryRow(ctx, `SELECT status, city_key, niche FROM scrape_jobs WHERE id = $1 FOR UPDATE`, id).Scan(&status, &cityKey, &niche); errors.Is(err, pgx.ErrNoRows) {
		return Job{}, ErrNotFound
	} else if err != nil {
		return Job{}, fmt.Errorf("lock scrape job for resume: %w", err)
	}
	if status != StatusFailed {
		return Job{}, ErrNotFailed
	}
	var activeExists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM scrape_jobs
		  WHERE id <> $1 AND city_key = $2 AND niche = $3
		    AND status IN ('queued', 'running', 'waiting')
		)`, id, cityKey, niche).Scan(&activeExists); err != nil {
		return Job{}, fmt.Errorf("check active scrape job before resume: %w", err)
	}
	if activeExists {
		return Job{}, ErrActiveJobExists
	}

	if _, err := tx.Exec(ctx, `
		UPDATE scrape_job_cells
		SET status = 'pending', page_token = NULL, page_number = 0,
		    results_seen_cycle = 0, saturated = false, last_error = NULL,
		    started_at = NULL, completed_at = NULL,
		    updated_at = now()
		WHERE scrape_job_id = $1 AND status IN ('failed', 'running')`, id); err != nil {
		return Job{}, fmt.Errorf("reset failed scrape cells: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE scrape_job_candidates
		SET status = 'discovered', attempts = 0, last_error = NULL, updated_at = now()
		WHERE scrape_job_id = $1 AND status = 'failed'`, id); err != nil {
		return Job{}, fmt.Errorf("reset failed scrape candidates: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE scrape_jobs
		SET status = 'queued',
		    requests_used_window = CASE
		      WHEN window_started_at IS NULL OR window_started_at <= now() - interval '24 hours' THEN 0
		      ELSE requests_used_window
		    END,
		    window_started_at = CASE
		      WHEN window_started_at IS NULL OR window_started_at <= now() - interval '24 hours' THEN now()
		      ELSE window_started_at
		    END,
		    resume_at = NULL, waiting_reason = NULL, current_cell_id = NULL,
		    last_error = NULL, locked_by = NULL, locked_at = NULL,
		    lease_expires_at = NULL, updated_at = now()
		WHERE id = $1`, id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Job{}, ErrActiveJobExists
		}
		return Job{}, fmt.Errorf("queue failed scrape job: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Job{}, fmt.Errorf("commit scrape job resume: %w", err)
	}
	return repo.GetByID(ctx, id)
}

// RetryFailed keeps compatibility for internal callers that still use the old
// action name. Both paths preserve completed cells and imported candidates.
func (repo *Postgres) RetryFailed(ctx context.Context, id uuid.UUID) (Job, error) {
	return repo.ResumeFailed(ctx, id)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(row rowScanner) (Job, error) {
	var job Job
	err := row.Scan(
		&job.ID,
		&job.City,
		&job.CityKey,
		&job.Niche,
		&job.Status,
		&job.CycleNumber,
		&job.MaxRequestsPerWindow,
		&job.RequestsUsedWindow,
		&job.RequestsUsedTotal,
		&job.WindowStartedAt,
		&job.ResumeAt,
		&job.WaitingReason,
		&job.LastCycleCompletedAt,
		&job.CurrentCellID,
		&job.LastError,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.Progress.CellsTotal,
		&job.Progress.CellsPending,
		&job.Progress.CellsCompleted,
		&job.Progress.CellsSubdivided,
		&job.Progress.CellsFailed,
		&job.Progress.CellsSaturated,
		&job.Progress.CandidatesTotal,
		&job.Progress.CandidatesPending,
		&job.Progress.CandidatesImported,
		&job.Progress.CandidatesDuplicate,
		&job.Progress.CandidatesFailed,
	)
	return job, err
}

var _ Repository = (*Postgres)(nil)
