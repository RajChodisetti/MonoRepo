package campaigns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	platformdb "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

const campaignSelectColumns = `
	id, restaurant_id, demo_site_id, campaign_type, status, current_step,
	subject, body_html, body_text, demo_token,
	approved_at, approved_by, last_sent_at, stopped_reason, created_at, updated_at`

func scanCampaign(row pgx.Row) (Campaign, error) {
	var record Campaign
	err := row.Scan(
		&record.ID,
		&record.RestaurantID,
		&record.DemoSiteID,
		&record.CampaignType,
		&record.Status,
		&record.CurrentStep,
		&record.Subject,
		&record.BodyHTML,
		&record.BodyText,
		&record.DemoToken,
		&record.ApprovedAt,
		&record.ApprovedBy,
		&record.LastSentAt,
		&record.StoppedReason,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	return record, err
}

func (repo *Postgres) Create(ctx context.Context, input CreateInput, draft DraftContent) (Campaign, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return Campaign{}, fmt.Errorf("begin campaign creation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := platformdb.LockRestaurantWorkflow(ctx, tx, input.RestaurantID); err != nil {
		return Campaign{}, err
	}

	var demoRestaurantID uuid.UUID
	var demoTokenHash string
	var demoExpiresAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT restaurant_id, token_hash, expires_at
		FROM demo_sites
		WHERE id = $1
		FOR SHARE`, input.DemoSiteID).Scan(
		&demoRestaurantID,
		&demoTokenHash,
		&demoExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("lock demo for campaign creation: %w", err)
	}
	if demoRestaurantID != input.RestaurantID {
		return Campaign{}, repository.ErrNotFound
	}
	if strings.TrimSpace(input.DemoToken) == "" || demos.CheckDemoToken(demoTokenHash, input.DemoToken) != nil {
		return Campaign{}, fmt.Errorf("%w: demo token changed before campaign creation", ErrNotEligible)
	}
	if demoExpiresAt != nil && !demoExpiresAt.After(time.Now().UTC()) {
		return Campaign{}, fmt.Errorf("%w: demo link expired before campaign creation", ErrNotEligible)
	}

	const query = `
		INSERT INTO email_campaigns (
			restaurant_id, demo_site_id, campaign_type, status,
			subject, body_html, body_text, demo_token
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING` + campaignSelectColumns

	record, err := scanCampaign(tx.QueryRow(
		ctx,
		query,
		input.RestaurantID,
		input.DemoSiteID,
		input.CampaignType,
		StatusDraft,
		draft.Subject,
		draft.BodyHTML,
		draft.BodyText,
		input.DemoToken,
	))
	if err != nil {
		return Campaign{}, fmt.Errorf("create campaign: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Campaign{}, fmt.Errorf("commit campaign creation: %w", err)
	}
	return record, nil
}

func (repo *Postgres) GetByID(ctx context.Context, id uuid.UUID) (Campaign, error) {
	query := `SELECT` + campaignSelectColumns + ` FROM email_campaigns WHERE id = $1`
	record, err := scanCampaign(repo.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("get campaign: %w", err)
	}
	return record, nil
}

func (repo *Postgres) ListByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]Campaign, error) {
	query := `SELECT` + campaignSelectColumns + `
		FROM email_campaigns
		WHERE restaurant_id = $1
		ORDER BY created_at DESC`

	rows, err := repo.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}
	defer rows.Close()

	var records []Campaign
	for rows.Next() {
		record, err := scanCampaign(rows)
		if err != nil {
			return nil, fmt.Errorf("scan campaign: %w", err)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (repo *Postgres) beginCampaignWorkflow(ctx context.Context, campaignID uuid.UUID) (pgx.Tx, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	var restaurantID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT restaurant_id FROM email_campaigns WHERE id = $1`, campaignID).Scan(&restaurantID); errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		return nil, repository.ErrNotFound
	} else if err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}
	if err := platformdb.LockRestaurantWorkflow(ctx, tx, restaurantID); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}
	return tx, nil
}

func (repo *Postgres) Approve(
	ctx context.Context,
	id uuid.UUID,
	approvedBy uuid.UUID,
	expectedUpdatedAt time.Time,
) (Campaign, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return Campaign{}, fmt.Errorf("begin campaign approval: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var workflowRestaurantID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT restaurant_id FROM email_campaigns WHERE id = $1`, id).Scan(&workflowRestaurantID); errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	} else if err != nil {
		return Campaign{}, fmt.Errorf("load campaign restaurant for approval: %w", err)
	}
	if err := platformdb.LockRestaurantWorkflow(ctx, tx, workflowRestaurantID); err != nil {
		return Campaign{}, err
	}

	var campaignRestaurantID uuid.UUID
	var demoRestaurantID uuid.UUID
	var campaignStatus string
	var demoToken string
	var currentUpdatedAt time.Time
	var demoTokenHash string
	var demoExpiresAt *time.Time
	var campaignProvenanceCurrent bool
	var demoProvenanceCurrent bool
	err = tx.QueryRow(ctx, `
		SELECT c.restaurant_id,
		       d.restaurant_id,
		       c.status,
		       c.demo_token,
		       c.updated_at,
		       d.token_hash,
		       d.expires_at,
		       (
		         NOT c.auto_generated
		         OR (
		           c.source_profile_fingerprint <> ''
		           AND c.source_profile_fingerprint = COALESCE(lead_artifact_current_profile_fingerprint(c.restaurant_id), '')
		         )
		       ),
		       (
		         NOT d.auto_generated
		         OR (
		           d.source_profile_fingerprint <> ''
		           AND d.source_profile_fingerprint = COALESCE(lead_artifact_current_profile_fingerprint(d.restaurant_id), '')
		         )
		       )
		FROM email_campaigns c
		JOIN demo_sites d ON d.id = c.demo_site_id
		WHERE c.id = $1
		FOR UPDATE OF c, d`, id).Scan(
		&campaignRestaurantID,
		&demoRestaurantID,
		&campaignStatus,
		&demoToken,
		&currentUpdatedAt,
		&demoTokenHash,
		&demoExpiresAt,
		&campaignProvenanceCurrent,
		&demoProvenanceCurrent,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("lock campaign approval: %w", err)
	}
	if campaignRestaurantID != demoRestaurantID {
		return Campaign{}, fmt.Errorf("%w: campaign demo belongs to another restaurant", ErrNotEligible)
	}
	if campaignStatus != StatusDraft && campaignStatus != StatusApproved {
		return Campaign{}, fmt.Errorf("%w: campaign is not approvable", ErrNotEligible)
	}
	if !campaignProvenanceCurrent || !demoProvenanceCurrent {
		return Campaign{}, fmt.Errorf("%w: automatic draft provenance is stale; prepare and review the current profile again", ErrNotEligible)
	}
	if !currentUpdatedAt.Equal(expectedUpdatedAt) {
		return Campaign{}, ErrStaleReview
	}
	if strings.TrimSpace(demoToken) == "" || demos.CheckDemoToken(demoTokenHash, demoToken) != nil {
		return Campaign{}, fmt.Errorf("%w: campaign demo token is no longer valid", ErrNotEligible)
	}
	if demoExpiresAt != nil && !demoExpiresAt.After(time.Now().UTC()) {
		return Campaign{}, fmt.Errorf("%w: demo link expired before approval", ErrNotEligible)
	}

	const query = `
		UPDATE email_campaigns
		SET status = $2, approved_at = now(), approved_by = $3, updated_at = now()
		WHERE id = $1 AND status IN ($4, $2)
		RETURNING` + campaignSelectColumns

	record, err := scanCampaign(tx.QueryRow(ctx, query, id, StatusApproved, approvedBy, StatusDraft))
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("approve campaign: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Campaign{}, fmt.Errorf("commit campaign approval: %w", err)
	}
	return record, nil
}

func (repo *Postgres) RegenerateDraft(
	ctx context.Context,
	id uuid.UUID,
	draft DraftContent,
	demoToken string,
	demoTokenHash string,
	demoExpiresAt *time.Time,
	regeneratedBy uuid.UUID,
) (Campaign, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return Campaign{}, fmt.Errorf("begin campaign regeneration: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var workflowRestaurantID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT restaurant_id FROM email_campaigns WHERE id = $1`, id).Scan(&workflowRestaurantID); errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	} else if err != nil {
		return Campaign{}, fmt.Errorf("load campaign restaurant for regeneration: %w", err)
	}
	if err := platformdb.LockRestaurantWorkflow(ctx, tx, workflowRestaurantID); err != nil {
		return Campaign{}, err
	}

	var restaurantID, demoSiteID uuid.UUID
	var campaignStatus, campaignType, demoStatus string
	var campaignAutoGenerated, demoAutoGenerated bool
	var currentProfileFingerprint string
	if err := tx.QueryRow(ctx, `
		SELECT c.restaurant_id,
		       c.demo_site_id,
		       c.status,
		       c.campaign_type,
		       d.status,
		       c.auto_generated,
		       d.auto_generated,
		       COALESCE(lead_artifact_current_profile_fingerprint(c.restaurant_id), '')
		FROM email_campaigns c
		JOIN demo_sites d ON d.id = c.demo_site_id
		WHERE c.id = $1
		FOR UPDATE OF c, d`, id).Scan(
		&restaurantID,
		&demoSiteID,
		&campaignStatus,
		&campaignType,
		&demoStatus,
		&campaignAutoGenerated,
		&demoAutoGenerated,
		&currentProfileFingerprint,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Campaign{}, repository.ErrNotFound
		}
		return Campaign{}, fmt.Errorf("lock campaign regeneration: %w", err)
	}
	if campaignType != TypeOutreach {
		return Campaign{}, ErrUnsupportedType
	}
	if campaignStatus != StatusDraft && campaignStatus != StatusApproved {
		return Campaign{}, fmt.Errorf("%w: campaign is not regenerable", ErrNotEligible)
	}
	if demoStatus != "draft" {
		return Campaign{}, fmt.Errorf("%w: demo must be a draft before regeneration", ErrNotEligible)
	}
	if (campaignAutoGenerated || demoAutoGenerated) && currentProfileFingerprint == "" {
		return Campaign{}, fmt.Errorf("%w: automatic draft requires current profile provenance", ErrNotEligible)
	}

	result, err := tx.Exec(ctx, `
		UPDATE demo_sites
		SET token_hash = $2,
		    expires_at = $3,
		    source_profile_fingerprint = CASE WHEN auto_generated THEN $4 ELSE source_profile_fingerprint END,
		    published_at = NULL,
		    published_by = NULL,
		    updated_at = now()
		WHERE id = $1 AND status = 'draft'`,
		demoSiteID,
		demoTokenHash,
		demoExpiresAt,
		currentProfileFingerprint,
	)
	if err != nil {
		return Campaign{}, fmt.Errorf("rotate demo token: %w", err)
	}
	if result.RowsAffected() != 1 {
		return Campaign{}, fmt.Errorf("%w: demo changed during regeneration", ErrNotEligible)
	}

	const updateCampaign = `
		UPDATE email_campaigns
		SET status = $2,
		    subject = $3,
		    body_html = $4,
		    body_text = $5,
		    demo_token = $6,
		    source_profile_fingerprint = CASE WHEN auto_generated THEN $7 ELSE source_profile_fingerprint END,
		    approved_at = NULL,
		    approved_by = NULL,
		    updated_at = now()
		WHERE id = $1 AND status IN ($8, $9)
		RETURNING` + campaignSelectColumns
	record, err := scanCampaign(tx.QueryRow(
		ctx,
		updateCampaign,
		id,
		StatusDraft,
		draft.Subject,
		draft.BodyHTML,
		draft.BodyText,
		demoToken,
		currentProfileFingerprint,
		StatusDraft,
		StatusApproved,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Campaign{}, fmt.Errorf("%w: campaign changed during regeneration", ErrNotEligible)
		}
		return Campaign{}, fmt.Errorf("regenerate campaign draft: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO email_events (campaign_id, restaurant_id, event_type, metadata)
		VALUES ($1, $2, $3, jsonb_build_object('regenerated_by', $4::text))`,
		id,
		restaurantID,
		EventDraftRegenerated,
		regeneratedBy,
	); err != nil {
		return Campaign{}, fmt.Errorf("audit campaign regeneration: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Campaign{}, fmt.Errorf("commit campaign regeneration: %w", err)
	}
	return record, nil
}

func (repo *Postgres) MarkSending(ctx context.Context, id uuid.UUID, step int) (Campaign, error) {
	tx, err := repo.beginCampaignWorkflow(ctx, id)
	if err != nil {
		return Campaign{}, fmt.Errorf("begin mark campaign sending: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE email_campaigns
		SET status = $2, current_step = $3, updated_at = now()
		WHERE id = $1
		RETURNING` + campaignSelectColumns

	record, err := scanCampaign(tx.QueryRow(ctx, query, id, StatusSending, step))
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("mark campaign sending: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Campaign{}, fmt.Errorf("commit mark campaign sending: %w", err)
	}
	return record, nil
}

func (repo *Postgres) MarkSent(ctx context.Context, id uuid.UUID, step int) (Campaign, error) {
	tx, err := repo.beginCampaignWorkflow(ctx, id)
	if err != nil {
		return Campaign{}, fmt.Errorf("begin mark campaign sent: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE email_campaigns
		SET status = $2, current_step = $3, last_sent_at = now(), updated_at = now()
		WHERE id = $1
		RETURNING` + campaignSelectColumns

	record, err := scanCampaign(tx.QueryRow(ctx, query, id, StatusSent, step))
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("mark campaign sent: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Campaign{}, fmt.Errorf("commit mark campaign sent: %w", err)
	}
	return record, nil
}

func (repo *Postgres) MarkSendSkipped(ctx context.Context, id uuid.UUID, step int) (Campaign, error) {
	tx, err := repo.beginCampaignWorkflow(ctx, id)
	if err != nil {
		return Campaign{}, fmt.Errorf("begin mark campaign skipped: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE email_campaigns
		SET status = CASE WHEN status = $4 THEN status ELSE $2 END,
		    current_step = $3,
		    updated_at = now()
		WHERE id = $1 AND status IN ($4, $5, $6)
		RETURNING` + campaignSelectColumns

	record, err := scanCampaign(tx.QueryRow(
		ctx,
		query,
		id,
		StatusApproved,
		step,
		StatusStopped,
		StatusSending,
		StatusApproved,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("mark campaign send skipped: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Campaign{}, fmt.Errorf("commit mark campaign skipped: %w", err)
	}
	return record, nil
}

func (repo *Postgres) Stop(ctx context.Context, id uuid.UUID, reason string) (Campaign, error) {
	tx, err := repo.beginCampaignWorkflow(ctx, id)
	if err != nil {
		return Campaign{}, fmt.Errorf("begin stop campaign: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE email_campaigns
		SET status = $2, stopped_reason = $3, updated_at = now()
		WHERE id = $1
		RETURNING` + campaignSelectColumns

	record, err := scanCampaign(tx.QueryRow(ctx, query, id, StatusStopped, reason))
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("stop campaign: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Campaign{}, fmt.Errorf("commit stop campaign: %w", err)
	}
	return record, nil
}

func (repo *Postgres) ListEvents(ctx context.Context, campaignID uuid.UUID) ([]Event, error) {
	const query = `
		SELECT id, campaign_id, restaurant_id, event_type, metadata, event_time
		FROM email_events
		WHERE campaign_id = $1
		ORDER BY event_time ASC`

	rows, err := repo.pool.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var records []Event
	for rows.Next() {
		var record Event
		if err := rows.Scan(
			&record.ID,
			&record.CampaignID,
			&record.RestaurantID,
			&record.EventType,
			&record.Metadata,
			&record.EventTime,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (repo *Postgres) InsertEvent(ctx context.Context, campaignID, restaurantID uuid.UUID, eventType string, metadata json.RawMessage) error {
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}
	const query = `
		INSERT INTO email_events (campaign_id, restaurant_id, event_type, metadata)
		VALUES ($1, $2, $3, $4)`
	_, err := repo.pool.Exec(ctx, query, campaignID, restaurantID, eventType, metadata)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

func (repo *Postgres) CreateTrackingToken(ctx context.Context, token TrackingToken) error {
	const query = `
		INSERT INTO email_tracking_tokens (
			token, campaign_id, restaurant_id, demo_site_id, token_type,
			target_url, recipient_email, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := repo.pool.Exec(
		ctx,
		query,
		token.Token,
		token.CampaignID,
		token.RestaurantID,
		token.DemoSiteID,
		token.TokenType,
		token.TargetURL,
		token.RecipientEmail,
		token.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create tracking token: %w", err)
	}
	return nil
}

func (repo *Postgres) GetTrackingToken(ctx context.Context, token string) (TrackingToken, error) {
	const query = `
		SELECT tracking.token, tracking.campaign_id, tracking.restaurant_id,
		       tracking.demo_site_id, tracking.token_type, tracking.target_url,
		       COALESCE(NULLIF(tracking.recipient_email, ''), lower(trim(restaurant.email))),
		       tracking.recipient_email <> '', tracking.expires_at, tracking.created_at
		FROM email_tracking_tokens AS tracking
		JOIN restaurants AS restaurant ON restaurant.id = tracking.restaurant_id
		WHERE tracking.token = $1`

	var record TrackingToken
	err := repo.pool.QueryRow(ctx, query, token).Scan(
		&record.Token,
		&record.CampaignID,
		&record.RestaurantID,
		&record.DemoSiteID,
		&record.TokenType,
		&record.TargetURL,
		&record.RecipientEmail,
		&record.RecipientSnapshot,
		&record.ExpiresAt,
		&record.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return TrackingToken{}, repository.ErrNotFound
	}
	if err != nil {
		return TrackingToken{}, fmt.Errorf("get tracking token: %w", err)
	}
	if record.ExpiresAt != nil && record.ExpiresAt.Before(time.Now().UTC()) {
		return TrackingToken{}, repository.ErrNotFound
	}
	return record, nil
}

func (repo *Postgres) IsSuppressed(ctx context.Context, email string) (bool, error) {
	const query = `SELECT 1 FROM email_suppressions WHERE lower(email) = lower($1) LIMIT 1`
	var exists int
	err := repo.pool.QueryRow(ctx, query, strings.TrimSpace(email)).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check suppression: %w", err)
	}
	return true, nil
}

func (repo *Postgres) AddSuppression(ctx context.Context, email, reason string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("add suppression: email is required")
	}
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin suppression: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(
		ctx,
		`SELECT pg_advisory_xact_lock(hashtextextended(lower(trim($1)), 0))`,
		email,
	); err != nil {
		return fmt.Errorf("lock suppression recipient: %w", err)
	}
	const query = `
		INSERT INTO email_suppressions (email, reason)
		VALUES (lower($1), $2)
		ON CONFLICT (email) DO NOTHING`
	_, err = tx.Exec(ctx, query, email, reason)
	if err != nil {
		return fmt.Errorf("add suppression: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit suppression: %w", err)
	}
	return nil
}

func (repo *Postgres) GetSendContext(ctx context.Context, campaignID uuid.UUID) (SendContext, error) {
	const query = `
		SELECT
			r.email,
			r.name,
			COALESCE(p.review_status, 'draft'),
			(p.reviewed_at IS NOT NULL
			 AND p.reviewed_by IS NOT NULL
			 AND p.reviewed_at >= p.updated_at
			 AND p.reviewed_at >= r.updated_at
			 AND (
			   NOT c.auto_generated
			   OR (
			     c.source_profile_fingerprint <> ''
			     AND c.source_profile_fingerprint = COALESCE(lead_artifact_current_profile_fingerprint(r.id), '')
			   )
			 )
			 AND (
			   NOT d.auto_generated
			   OR (
			     d.source_profile_fingerprint <> ''
			     AND d.source_profile_fingerprint = COALESCE(lead_artifact_current_profile_fingerprint(r.id), '')
			   )
			 )),
			d.status,
			(d.published_at IS NOT NULL AND d.published_by IS NOT NULL),
			(d.expires_at IS NOT NULL AND d.expires_at <= now()),
			d.slug
		FROM email_campaigns c
		JOIN restaurants r ON r.id = c.restaurant_id
		JOIN demo_sites d ON d.id = c.demo_site_id
		LEFT JOIN restaurant_profiles p ON p.restaurant_id = c.restaurant_id
		WHERE c.id = $1`

	var ctxData SendContext
	err := repo.pool.QueryRow(ctx, query, campaignID).Scan(
		&ctxData.RestaurantEmail,
		&ctxData.RestaurantName,
		&ctxData.ReviewStatus,
		&ctxData.ProfileReviewAudited,
		&ctxData.DemoStatus,
		&ctxData.DemoPublishAudited,
		&ctxData.DemoExpired,
		&ctxData.DemoSlug,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SendContext{}, repository.ErrNotFound
	}
	if err != nil {
		return SendContext{}, fmt.Errorf("get send context: %w", err)
	}
	return ctxData, nil
}

func (repo *Postgres) GetRestaurantContext(ctx context.Context, restaurantID uuid.UUID) (SendContext, error) {
	const query = `
		SELECT
			r.email,
			r.name,
			COALESCE(p.review_status, 'draft'),
			(p.reviewed_at IS NOT NULL
			 AND p.reviewed_by IS NOT NULL
			 AND p.reviewed_at >= p.updated_at
			 AND p.reviewed_at >= r.updated_at),
			'',
			false,
			false,
			''
		FROM restaurants r
		LEFT JOIN restaurant_profiles p ON p.restaurant_id = r.id
		WHERE r.id = $1`

	var ctxData SendContext
	err := repo.pool.QueryRow(ctx, query, restaurantID).Scan(
		&ctxData.RestaurantEmail,
		&ctxData.RestaurantName,
		&ctxData.ReviewStatus,
		&ctxData.ProfileReviewAudited,
		&ctxData.DemoStatus,
		&ctxData.DemoPublishAudited,
		&ctxData.DemoExpired,
		&ctxData.DemoSlug,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SendContext{}, repository.ErrNotFound
	}
	if err != nil {
		return SendContext{}, fmt.Errorf("get restaurant context: %w", err)
	}
	return ctxData, nil
}

func (repo *Postgres) GetSiteIndexByRestaurantID(ctx context.Context, restaurantID uuid.UUID) (int, error) {
	const query = `
		WITH ranked AS (
			SELECT
				r.id,
				ROW_NUMBER() OVER (ORDER BY r.created_at ASC, r.name ASC) - 1 AS site_index
			FROM restaurants r
			JOIN restaurant_profiles rp ON rp.restaurant_id = r.id
			WHERE rp.google_place_id IS NOT NULL AND rp.google_place_id <> ''
		)
		SELECT site_index
		FROM ranked
		WHERE id = $1`

	var siteIndex int
	err := repo.pool.QueryRow(ctx, query, restaurantID).Scan(&siteIndex)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoSiteIndex
	}
	if err != nil {
		return 0, fmt.Errorf("get site index: %w", err)
	}
	return siteIndex, nil
}

func (repo *Postgres) MarkRestaurantEmailed(ctx context.Context, restaurantID uuid.UUID) error {
	const query = `
		UPDATE restaurants
		SET is_contacted = true,
		    email_sent = true,
		    email_send_count = email_send_count + 1,
		    last_email_sent_at = now(),
		    last_email_send_sequence = nextval('email_send_sequence'),
		    last_email_recipient = lower(trim(email)),
		    status = CASE WHEN status IN ('lead', 'demo_ready') THEN 'emailed' ELSE status END,
		    updated_at = now()
		WHERE id = $1`
	_, err := repo.pool.Exec(ctx, query, restaurantID)
	if err != nil {
		return fmt.Errorf("mark restaurant emailed: %w", err)
	}
	return nil
}

func (repo *Postgres) GetLatestDemoTokenByDemoSiteID(ctx context.Context, demoSiteID uuid.UUID) (string, error) {
	const query = `
		SELECT demo_token
		FROM email_campaigns
		WHERE demo_site_id = $1 AND trim(demo_token) <> ''
		ORDER BY created_at DESC
		LIMIT 1`

	var token string
	err := repo.pool.QueryRow(ctx, query, demoSiteID).Scan(&token)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get latest demo token: %w", err)
	}
	return strings.TrimSpace(token), nil
}

var _ Repository = (*Postgres)(nil)
