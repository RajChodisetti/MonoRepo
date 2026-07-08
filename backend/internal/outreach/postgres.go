package outreach

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListEligibleLeads(ctx context.Context, limit int) ([]EligibleLead, error)
	CountEligibleLeads(ctx context.Context) (int, error)
}

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

const eligibleLeadsBaseQuery = `
	FROM restaurants r
	JOIN restaurant_profiles rp ON rp.restaurant_id = r.id
	JOIN demo_sites d ON d.restaurant_id = r.id AND d.status = 'published'
	WHERE rp.ocr_verified = true
	  AND r.email_sent = false
	  AND trim(r.email) <> ''
	  AND NOT EXISTS (
	    SELECT 1 FROM email_suppressions s WHERE s.email = lower(trim(r.email))
	  )`

func (repo *Postgres) ListEligibleLeads(ctx context.Context, limit int) ([]EligibleLead, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}
	if limit < 1 {
		return nil, fmt.Errorf("limit must be at least 1")
	}

	query := `
		SELECT r.id, r.email, r.name, d.id, d.slug` + eligibleLeadsBaseQuery + `
		ORDER BY r.created_at ASC
		LIMIT $1`

	rows, err := repo.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list eligible leads: %w", err)
	}
	defer rows.Close()

	var leads []EligibleLead
	for rows.Next() {
		var lead EligibleLead
		if err := rows.Scan(&lead.RestaurantID, &lead.Email, &lead.Name, &lead.DemoSiteID, &lead.DemoSlug); err != nil {
			return nil, fmt.Errorf("scan eligible lead: %w", err)
		}
		leads = append(leads, lead)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate eligible leads: %w", err)
	}
	return leads, nil
}

func (repo *Postgres) CountEligibleLeads(ctx context.Context) (int, error) {
	if repo.pool == nil {
		return 0, fmt.Errorf("database pool is not configured")
	}

	query := `SELECT COUNT(*) ` + eligibleLeadsBaseQuery
	var count int
	if err := repo.pool.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count eligible leads: %w", err)
	}
	return count, nil
}

func HasActiveBulkJob(ctx context.Context, pool *pgxpool.Pool, jobType string) (bool, string, error) {
	if pool == nil {
		return false, "", nil
	}

	const query = `
		SELECT id::text
		FROM job_runs
		WHERE job_type = $1 AND status IN ('queued', 'running')
		ORDER BY created_at DESC
		LIMIT 1`

	var jobID string
	err := pool.QueryRow(ctx, query, jobType).Scan(&jobID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, "", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("check active bulk job: %w", err)
	}
	return true, jobID, nil
}

func GetLatestBulkJobSummary(ctx context.Context, pool *pgxpool.Pool, jobType string) (*CompletedJobStatus, error) {
	if pool == nil {
		return nil, nil
	}

	const query = `
		SELECT id::text, status, payload
		FROM job_runs
		WHERE job_type = $1 AND status IN ('completed', 'failed')
		ORDER BY updated_at DESC
		LIMIT 1`

	var jobID string
	var status string
	var payload []byte
	err := pool.QueryRow(ctx, query, jobType).Scan(&jobID, &status, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest bulk job: %w", err)
	}

	summary, err := decodeBulkSummary(payload)
	if err != nil {
		return nil, err
	}

	return &CompletedJobStatus{
		JobID:   jobID,
		Status:  status,
		Summary: summary,
	}, nil
}

var _ Repository = (*Postgres)(nil)
