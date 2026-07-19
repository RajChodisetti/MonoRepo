package outreach

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListEligibleLeads(ctx context.Context, limit int) ([]EligibleLead, error)
	CountEligibleLeads(ctx context.Context) (int, error)
	IsEmailSuppressed(ctx context.Context, email string) (bool, error)
	RecordAdHocEmailSent(ctx context.Context, restaurantID uuid.UUID, recipientEmail string) error
}

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func GetEmailJobControl(ctx context.Context, pool *pgxpool.Pool) (EmailJobControl, error) {
	if pool == nil {
		return EmailJobControl{}, nil
	}
	const query = `
		SELECT enabled, enabled_at, enabled_by, updated_at
		FROM outreach_runtime_control
		WHERE control_key = 'email_job'`
	var control EmailJobControl
	err := pool.QueryRow(ctx, query).Scan(
		&control.Enabled,
		&control.EnabledAt,
		&control.EnabledBy,
		&control.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return EmailJobControl{}, nil
	}
	if err != nil {
		return EmailJobControl{}, fmt.Errorf("get outreach email job control: %w", err)
	}
	return control, nil
}

func SetEmailJobControl(ctx context.Context, pool *pgxpool.Pool, enabled bool, enabledBy *uuid.UUID) (EmailJobControl, error) {
	if pool == nil {
		return EmailJobControl{}, fmt.Errorf("database pool is not configured")
	}
	const query = `
		INSERT INTO outreach_runtime_control (control_key, enabled, enabled_at, enabled_by, updated_at)
		VALUES ('email_job', $1, CASE WHEN $1 THEN now() ELSE NULL END, CASE WHEN $1 THEN $2 ELSE NULL END, now())
		ON CONFLICT (control_key) DO UPDATE
		SET enabled = EXCLUDED.enabled,
		    enabled_at = EXCLUDED.enabled_at,
		    enabled_by = EXCLUDED.enabled_by,
		    updated_at = now()
		RETURNING enabled, enabled_at, enabled_by, updated_at`
	var control EmailJobControl
	if err := pool.QueryRow(ctx, query, enabled, enabledBy).Scan(
		&control.Enabled,
		&control.EnabledAt,
		&control.EnabledBy,
		&control.UpdatedAt,
	); err != nil {
		return EmailJobControl{}, fmt.Errorf("set outreach email job control: %w", err)
	}
	return control, nil
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

func (repo *Postgres) IsEmailSuppressed(ctx context.Context, email string) (bool, error) {
	if repo.pool == nil {
		return false, fmt.Errorf("database pool is not configured")
	}

	const query = `SELECT EXISTS (SELECT 1 FROM email_suppressions WHERE email = lower(trim($1)))`
	var suppressed bool
	if err := repo.pool.QueryRow(ctx, query, strings.ToLower(strings.TrimSpace(email))).Scan(&suppressed); err != nil {
		return false, fmt.Errorf("check email suppression: %w", err)
	}
	return suppressed, nil
}

func (repo *Postgres) RecordAdHocEmailSent(ctx context.Context, restaurantID uuid.UUID, recipientEmail string) error {
	if repo.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}

	const query = `
		UPDATE restaurants
		SET is_contacted = true,
		    email_sent = true,
		    email_send_count = email_send_count + 1,
		    last_email_sent_at = now(),
		    last_email_recipient = $2,
		    status = CASE WHEN status IN ('lead', 'demo_ready') THEN 'emailed' ELSE status END,
		    updated_at = now()
		WHERE id = $1`

	result, err := repo.pool.Exec(ctx, query, restaurantID, recipientEmail)
	if err != nil {
		return fmt.Errorf("record ad hoc email sent: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("restaurant was not found while recording ad hoc email")
	}
	return nil
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
