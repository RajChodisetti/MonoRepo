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
	JOIN LATERAL (
	  SELECT c.id AS campaign_id,
	         c.approved_at,
	         c.approved_by,
	         d.id AS demo_site_id,
	         d.status AS demo_status,
	         d.published_at,
	         d.published_by,
	         d.expires_at,
	         c.auto_generated AS campaign_auto_generated,
	         c.source_ocr_fingerprint AS campaign_source_ocr_fingerprint,
	         c.source_profile_fingerprint AS campaign_source_profile_fingerprint,
	         d.auto_generated AS demo_auto_generated,
	         d.source_ocr_fingerprint AS demo_source_ocr_fingerprint,
	         d.source_profile_fingerprint AS demo_source_profile_fingerprint
	  FROM email_campaigns c
	  JOIN demo_sites d ON d.id = c.demo_site_id
	  WHERE c.restaurant_id = r.id
	    AND d.restaurant_id = r.id
	    AND c.campaign_type = 'outreach'
	    AND c.status = 'approved'
	    AND d.status = 'published'
	    AND (d.expires_at IS NULL OR d.expires_at > now())
	  ORDER BY c.approved_at DESC NULLS LAST, c.created_at DESC
	  LIMIT 1
	) eligible ON true
	WHERE rp.ocr_status = 'verified'
	  AND rp.review_status = 'approved'
	  AND rp.reviewed_at IS NOT NULL
	  AND rp.reviewed_by IS NOT NULL
	  AND rp.reviewed_at >= rp.updated_at
	  AND rp.reviewed_at >= r.updated_at
	  AND eligible.demo_status = 'published'
	  AND eligible.published_at IS NOT NULL
	  AND eligible.published_by IS NOT NULL
	  AND (eligible.expires_at IS NULL OR eligible.expires_at > now())
	  AND eligible.approved_at IS NOT NULL
	  AND eligible.approved_by IS NOT NULL
	  AND (
	    NOT eligible.campaign_auto_generated
	    OR (
	      eligible.campaign_source_ocr_fingerprint <> ''
	      AND eligible.campaign_source_ocr_fingerprint = rp.ocr_input_fingerprint
	      AND eligible.campaign_source_profile_fingerprint <> ''
	      AND eligible.campaign_source_profile_fingerprint = COALESCE(lead_artifact_current_profile_fingerprint(r.id), '')
	    )
	  )
	  AND (
	    NOT eligible.demo_auto_generated
	    OR (
	      eligible.demo_source_ocr_fingerprint <> ''
	      AND eligible.demo_source_ocr_fingerprint = rp.ocr_input_fingerprint
	      AND eligible.demo_source_profile_fingerprint <> ''
	      AND eligible.demo_source_profile_fingerprint = COALESCE(lead_artifact_current_profile_fingerprint(r.id), '')
	    )
	  )
	  AND r.email_sent = false
	  AND r.email_send_count = 0
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
		SELECT eligible.campaign_id, r.id, eligible.demo_site_id` + eligibleLeadsBaseQuery + `
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
		if err := rows.Scan(&lead.CampaignID, &lead.RestaurantID, &lead.DemoSiteID); err != nil {
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
