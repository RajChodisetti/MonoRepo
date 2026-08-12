package outreach

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
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

func DisableEmailJobControl(ctx context.Context, pool *pgxpool.Pool) (EmailJobControl, error) {
	if pool == nil {
		return EmailJobControl{}, fmt.Errorf("database pool is not configured")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return EmailJobControl{}, fmt.Errorf("begin disabling outreach email job: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var control EmailJobControl
	if err := tx.QueryRow(ctx, `
		INSERT INTO outreach_runtime_control (control_key, enabled, enabled_at, enabled_by, updated_at)
		VALUES ('email_job', false, NULL, NULL, now())
		ON CONFLICT (control_key) DO UPDATE
		SET enabled = false, enabled_at = NULL, enabled_by = NULL, updated_at = now()
		RETURNING enabled, enabled_at, enabled_by, updated_at`).Scan(
		&control.Enabled, &control.EnabledAt, &control.EnabledBy, &control.UpdatedAt,
	); err != nil {
		return EmailJobControl{}, fmt.Errorf("disable outreach email job control: %w", err)
	}
	// Deferred work has not crossed a provider boundary, so it is safe to
	// cancel. A running lease is left intact to finalize any in-flight provider
	// outcome; RunBulkSend rechecks the disabled control before another send.
	if _, err := tx.Exec(ctx, `
		UPDATE job_runs
		SET status = 'cancelled',
		    locked_at = NULL,
		    locked_by = NULL,
		    lease_expires_at = NULL,
		    last_error = 'Cancelled when outreach email job was disabled.',
		    updated_at = now()
		WHERE job_type = $1 AND status = 'queued'`, BulkSendJobType); err != nil {
		return EmailJobControl{}, fmt.Errorf("cancel deferred outreach email job: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return EmailJobControl{}, fmt.Errorf("commit disabling outreach email job: %w", err)
	}
	return control, nil
}

func EnableEmailJobControl(
	ctx context.Context,
	pool *pgxpool.Pool,
	enabledBy uuid.UUID,
) (EmailJobControl, string, error) {
	if pool == nil {
		return EmailJobControl{}, "", fmt.Errorf("database pool is not configured")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return EmailJobControl{}, "", fmt.Errorf("begin enabling outreach email job: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var control EmailJobControl
	if err := tx.QueryRow(ctx, `
		INSERT INTO outreach_runtime_control (control_key, enabled, enabled_at, enabled_by, updated_at)
		VALUES ('email_job', true, now(), $1, now())
		ON CONFLICT (control_key) DO UPDATE
		SET enabled = true, enabled_at = now(), enabled_by = EXCLUDED.enabled_by, updated_at = now()
		RETURNING enabled, enabled_at, enabled_by, updated_at`, enabledBy).Scan(
		&control.Enabled, &control.EnabledAt, &control.EnabledBy, &control.UpdatedAt,
	); err != nil {
		return EmailJobControl{}, "", fmt.Errorf("enable outreach email job control: %w", err)
	}
	var activeJobID string
	err = tx.QueryRow(ctx, `
		SELECT id::text
		FROM job_runs
		WHERE job_type = $1 AND status IN ('queued', 'running')
		ORDER BY created_at
		LIMIT 1
		FOR UPDATE`, BulkSendJobType).Scan(&activeJobID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return EmailJobControl{}, "", fmt.Errorf("lock active outreach job while enabling: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return EmailJobControl{}, "", fmt.Errorf("commit enabling outreach email job: %w", err)
	}
	return control, activeJobID, nil
}

const eligibleLeadsBaseQuery = `
	FROM email_campaigns campaign
	JOIN restaurants restaurant ON restaurant.id = campaign.restaurant_id
	JOIN outreach_email_sequences sequence ON sequence.id = campaign.sequence_id
	JOIN outreach_email_sequence_steps step
	  ON step.sequence_id = campaign.sequence_id
	 AND step.position = campaign.next_step
	WHERE campaign.campaign_type = 'outreach'
	  AND campaign.sequence_id IS NOT NULL
	  AND campaign.status = 'approved'
	  AND campaign.next_step IS NOT NULL
	  AND campaign.next_send_at IS NOT NULL
	  AND campaign.next_send_at <= now()
	  AND sequence.status IN ('approved', 'archived')
	  AND sequence.approved_at IS NOT NULL
	  AND step.enabled = true
	  AND length(trim(restaurant.name)) > 0
	  AND lower(trim(restaurant.email)) ~ '^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$'
	  AND restaurant.status IN ('lead', 'emailed')
	  AND restaurant.shown_interest = false
	  AND restaurant.outreach_consent_basis = 'inferred_business'
	  AND restaurant.outreach_consent_recorded_at IS NOT NULL
	  AND length(trim(restaurant.outreach_consent_source)) > 0
	  AND jsonb_typeof(restaurant.outreach_consent_evidence) = 'object'
	  AND restaurant.outreach_consent_evidence <> '{}'::jsonb
	  AND (
	    campaign.current_step > 0
	    OR NOT EXISTS (
	      SELECT 1
	      FROM email_campaigns existing_campaign
	      JOIN restaurants existing_restaurant
	        ON existing_restaurant.id = existing_campaign.restaurant_id
	      JOIN outreach_email_sequences existing_sequence
	        ON existing_sequence.id = existing_campaign.sequence_id
	      JOIN outreach_email_sequence_steps existing_step
	        ON existing_step.sequence_id = existing_campaign.sequence_id
	       AND existing_step.position = existing_campaign.next_step
	      WHERE existing_campaign.campaign_type = 'outreach'
	        AND existing_campaign.sequence_id IS NOT NULL
	        AND existing_campaign.status = 'approved'
	        AND existing_campaign.current_step > 0
	        AND existing_campaign.next_step IS NOT NULL
	        AND existing_sequence.status IN ('approved', 'archived')
	        AND existing_sequence.approved_at IS NOT NULL
	        AND existing_step.enabled = true
	        AND length(trim(existing_restaurant.name)) > 0
	        AND lower(trim(existing_restaurant.email)) ~ '^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$'
	        AND existing_restaurant.status IN ('lead', 'emailed')
	        AND existing_restaurant.shown_interest = false
	        AND existing_restaurant.outreach_consent_basis = 'inferred_business'
	        AND existing_restaurant.outreach_consent_recorded_at IS NOT NULL
	        AND length(trim(existing_restaurant.outreach_consent_source)) > 0
	        AND jsonb_typeof(existing_restaurant.outreach_consent_evidence) = 'object'
	        AND existing_restaurant.outreach_consent_evidence <> '{}'::jsonb
	    )
	  )`

func (repo *Postgres) ListEligibleLeads(ctx context.Context, limit int) ([]EligibleLead, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}
	if limit < 1 {
		return nil, fmt.Errorf("limit must be at least 1")
	}

	query := `
		SELECT campaign.id, restaurant.id, COALESCE(campaign.demo_site_id, '00000000-0000-0000-0000-000000000000'::uuid), campaign.next_step` + eligibleLeadsBaseQuery + `
		ORDER BY (campaign.current_step > 0) DESC,
		         campaign.next_send_at ASC,
		         restaurant.created_at ASC
		LIMIT $1`

	rows, err := repo.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list eligible leads: %w", err)
	}
	defer rows.Close()

	var leads []EligibleLead
	for rows.Next() {
		var lead EligibleLead
		if err := rows.Scan(&lead.CampaignID, &lead.RestaurantID, &lead.DemoSiteID, &lead.Step); err != nil {
			return nil, fmt.Errorf("scan eligible lead: %w", err)
		}
		leads = append(leads, lead)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate eligible leads: %w", err)
	}
	return leads, nil
}

type sequenceDeliveryRepository interface {
	GetSequenceDelivery(ctx context.Context, campaignID uuid.UUID, step int) (SequenceDelivery, error)
	PrepareSequenceDelivery(ctx context.Context, campaignID uuid.UUID, step int, subject, bodyHTML, bodyText string) error
	FinalizeSequenceDelivery(ctx context.Context, finalization SequenceDeliveryFinalization) error
	NextSequenceDueAt(ctx context.Context) (*time.Time, error)
}

func (repo *Postgres) GetSequenceDelivery(ctx context.Context, campaignID uuid.UUID, position int) (SequenceDelivery, error) {
	if repo.pool == nil {
		return SequenceDelivery{}, fmt.Errorf("database pool is not configured")
	}
	const query = `
		SELECT campaign.id,
		       restaurant.id,
		       restaurant.name,
		       COALESCE(
		         NULLIF(trim(profile.apollo_lead #>> '{contact,first_name}'), ''),
		         NULLIF(split_part(trim(profile.apollo_lead #>> '{contact,name}'), ' ', 1), ''),
		         NULLIF(split_part(trim(profile.owners ->> 0), ' ', 1), ''),
		         ''
		       ),
		       lower(trim(restaurant.email)),
		       restaurant.status,
		       restaurant.shown_interest,
		       restaurant.outreach_consent_basis,
		       restaurant.outreach_consent_source,
		       restaurant.outreach_consent_recorded_at,
		       restaurant.outreach_consent_evidence,
		       sequence.status,
		       step.id,
		       step.sequence_id,
		       step.position,
		       step.enabled,
		       step.delay_hours,
		       step.subject_template,
		       step.body_text_template,
		       step.created_at,
		       step.updated_at
		FROM email_campaigns campaign
		JOIN restaurants restaurant ON restaurant.id = campaign.restaurant_id
		LEFT JOIN restaurant_profiles profile ON profile.restaurant_id = restaurant.id
		JOIN outreach_email_sequences sequence ON sequence.id = campaign.sequence_id
		JOIN outreach_email_sequence_steps step
		  ON step.sequence_id = campaign.sequence_id
		 AND step.position = $2
		WHERE campaign.id = $1
		  AND campaign.campaign_type = 'outreach'
		  AND campaign.sequence_id IS NOT NULL
		  AND campaign.status = 'approved'
		  AND campaign.next_step = $2
		  AND campaign.next_send_at <= now()`
	var delivery SequenceDelivery
	if err := repo.pool.QueryRow(ctx, query, campaignID, position).Scan(
		&delivery.CampaignID,
		&delivery.RestaurantID,
		&delivery.RestaurantName,
		&delivery.OwnerFirstName,
		&delivery.RecipientEmail,
		&delivery.LifecycleStatus,
		&delivery.ShownInterest,
		&delivery.ConsentBasis,
		&delivery.ConsentSource,
		&delivery.ConsentRecordedAt,
		&delivery.ConsentEvidence,
		&delivery.SequenceStatus,
		&delivery.Step.ID,
		&delivery.Step.SequenceID,
		&delivery.Step.Position,
		&delivery.Step.Enabled,
		&delivery.Step.DelayHours,
		&delivery.Step.SubjectTemplate,
		&delivery.Step.BodyTextTemplate,
		&delivery.Step.CreatedAt,
		&delivery.Step.UpdatedAt,
	); errors.Is(err, pgx.ErrNoRows) {
		return SequenceDelivery{}, fmt.Errorf("%w: sequence step is not due", campaigns.ErrNotEligible)
	} else if err != nil {
		return SequenceDelivery{}, fmt.Errorf("load sequence delivery: %w", err)
	}
	return delivery, nil
}

func (repo *Postgres) PrepareSequenceDelivery(ctx context.Context, campaignID uuid.UUID, step int, subject, bodyHTML, bodyText string) error {
	if repo.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}
	result, err := repo.pool.Exec(ctx, `
		UPDATE email_campaigns
		SET subject = $3,
		    body_html = $4,
		    body_text = $5,
		    demo_token = '',
		    updated_at = now()
		WHERE id = $1
		  AND next_step = $2
		  AND status = 'approved'
		  AND next_send_at <= now()`, campaignID, step, subject, bodyHTML, bodyText)
	if err != nil {
		return fmt.Errorf("prepare sequence delivery: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("%w: sequence step changed before delivery", campaigns.ErrNotEligible)
	}
	return nil
}

func (repo *Postgres) NextSequenceDueAt(ctx context.Context) (*time.Time, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}
	var next *time.Time
	if err := repo.pool.QueryRow(ctx, `
		SELECT min(campaign.next_send_at)
		FROM email_campaigns campaign
		JOIN restaurants restaurant ON restaurant.id = campaign.restaurant_id
		JOIN outreach_email_sequences sequence ON sequence.id = campaign.sequence_id
		JOIN outreach_email_sequence_steps step
		  ON step.sequence_id = campaign.sequence_id
		 AND step.position = campaign.next_step
		WHERE campaign.campaign_type = 'outreach'
		  AND campaign.sequence_id IS NOT NULL
		  AND campaign.status = 'approved'
		  AND campaign.current_step > 0
		  AND campaign.next_step IS NOT NULL
		  AND campaign.next_send_at IS NOT NULL
		  AND sequence.status IN ('approved', 'archived')
		  AND sequence.approved_at IS NOT NULL
		  AND step.enabled = true
		  AND length(trim(restaurant.name)) > 0
		  AND lower(trim(restaurant.email)) ~ '^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$'
		  AND restaurant.status IN ('lead', 'emailed')
		  AND restaurant.shown_interest = false
		  AND restaurant.outreach_consent_basis = 'inferred_business'
		  AND restaurant.outreach_consent_recorded_at IS NOT NULL
		  AND length(trim(restaurant.outreach_consent_source)) > 0
		  AND jsonb_typeof(restaurant.outreach_consent_evidence) = 'object'
		  AND restaurant.outreach_consent_evidence <> '{}'::jsonb`).Scan(&next); err != nil {
		return nil, fmt.Errorf("get next sequence due time: %w", err)
	}
	return next, nil
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

func (repo *Postgres) CountRecipientStatuses(ctx context.Context) (RecipientStatusCounts, error) {
	if repo.pool == nil {
		return RecipientStatusCounts{}, fmt.Errorf("database pool is not configured")
	}
	const query = `
		WITH recipient_state AS MATERIALIZED (
		  SELECT campaign.current_step,
		         campaign.next_step,
		         campaign.next_send_at,
		         campaign.completed_at,
		         campaign.status AS campaign_status,
		         sequence.status AS sequence_status,
		         sequence.approved_at IS NOT NULL AS sequence_approved,
		         COALESCE(step.enabled, false) AS step_enabled,
		         length(trim(restaurant.name)) > 0 AS has_name,
		         lower(trim(restaurant.email)) ~ '^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$' AS has_email,
		         restaurant.status IN ('lead', 'emailed') AS lifecycle_eligible,
		         restaurant.shown_interest = false AS has_no_interest,
		         restaurant.outreach_consent_basis = 'inferred_business'
		           AND restaurant.outreach_consent_recorded_at IS NOT NULL
		           AND length(trim(restaurant.outreach_consent_source)) > 0
		           AND jsonb_typeof(restaurant.outreach_consent_evidence) = 'object'
		           AND restaurant.outreach_consent_evidence <> '{}'::jsonb AS has_consent_evidence
		  FROM email_campaigns campaign
		  JOIN restaurants restaurant ON restaurant.id = campaign.restaurant_id
		  JOIN outreach_email_sequences sequence ON sequence.id = campaign.sequence_id
		  LEFT JOIN outreach_email_sequence_steps step
		    ON step.sequence_id = campaign.sequence_id
		   AND step.position = campaign.next_step
		  WHERE campaign.campaign_type = 'outreach'
		    AND campaign.sequence_id IS NOT NULL
		), policy_state AS MATERIALIZED (
		  SELECT *,
		         campaign_status = 'approved'
		           AND sequence_status IN ('approved', 'archived')
		           AND sequence_approved
		           AND step_enabled
		           AND has_name
		           AND has_email
		           AND lifecycle_eligible
		           AND has_no_interest
		           AND has_consent_evidence
		           AND next_step IS NOT NULL
		           AND next_send_at IS NOT NULL AS policy_eligible
		  FROM recipient_state
		), phase AS (
		  SELECT EXISTS (
		    SELECT 1 FROM policy_state
		    WHERE policy_eligible AND current_step > 0
		  ) AS unfinished_followups
		)
		SELECT count(*) FILTER (
		         WHERE policy_eligible AND current_step > 0 AND next_send_at <= now()
		       ),
		       count(*) FILTER (
		         WHERE policy_eligible AND current_step = 0 AND next_send_at <= now()
		           AND NOT (SELECT unfinished_followups FROM phase)
		       ),
		       count(*) FILTER (
		         WHERE completed_at IS NULL AND NOT policy_eligible
		       ),
		       count(*) FILTER (
		         WHERE completed_at IS NOT NULL
		       )
		FROM policy_state`
	var counts RecipientStatusCounts
	if err := repo.pool.QueryRow(ctx, query).Scan(
		&counts.DueFollowups, &counts.NewRecipients, &counts.Paused, &counts.Completed,
	); err != nil {
		return RecipientStatusCounts{}, fmt.Errorf("count outreach recipient states: %w", err)
	}
	return counts, nil
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
