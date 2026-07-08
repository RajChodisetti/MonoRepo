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
	const query = `
		INSERT INTO email_campaigns (
			restaurant_id, demo_site_id, campaign_type, status,
			subject, body_html, body_text, demo_token
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING` + campaignSelectColumns

	return scanCampaign(repo.pool.QueryRow(
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

func (repo *Postgres) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) (Campaign, error) {
	const query = `
		UPDATE email_campaigns
		SET status = $2, approved_at = now(), approved_by = $3, updated_at = now()
		WHERE id = $1 AND status <> $4
		RETURNING` + campaignSelectColumns

	record, err := scanCampaign(repo.pool.QueryRow(ctx, query, id, StatusApproved, approvedBy, StatusStopped))
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("approve campaign: %w", err)
	}
	return record, nil
}

func (repo *Postgres) MarkSending(ctx context.Context, id uuid.UUID, step int) (Campaign, error) {
	const query = `
		UPDATE email_campaigns
		SET status = $2, current_step = $3, updated_at = now()
		WHERE id = $1
		RETURNING` + campaignSelectColumns

	record, err := scanCampaign(repo.pool.QueryRow(ctx, query, id, StatusSending, step))
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("mark campaign sending: %w", err)
	}
	return record, nil
}

func (repo *Postgres) MarkSent(ctx context.Context, id uuid.UUID, step int) (Campaign, error) {
	const query = `
		UPDATE email_campaigns
		SET status = $2, current_step = $3, last_sent_at = now(), updated_at = now()
		WHERE id = $1
		RETURNING` + campaignSelectColumns

	record, err := scanCampaign(repo.pool.QueryRow(ctx, query, id, StatusSent, step))
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("mark campaign sent: %w", err)
	}
	return record, nil
}

func (repo *Postgres) Stop(ctx context.Context, id uuid.UUID, reason string) (Campaign, error) {
	const query = `
		UPDATE email_campaigns
		SET status = $2, stopped_reason = $3, updated_at = now()
		WHERE id = $1
		RETURNING` + campaignSelectColumns

	record, err := scanCampaign(repo.pool.QueryRow(ctx, query, id, StatusStopped, reason))
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, repository.ErrNotFound
	}
	if err != nil {
		return Campaign{}, fmt.Errorf("stop campaign: %w", err)
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
			token, campaign_id, restaurant_id, demo_site_id, token_type, target_url, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := repo.pool.Exec(
		ctx,
		query,
		token.Token,
		token.CampaignID,
		token.RestaurantID,
		token.DemoSiteID,
		token.TokenType,
		token.TargetURL,
		token.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create tracking token: %w", err)
	}
	return nil
}

func (repo *Postgres) GetTrackingToken(ctx context.Context, token string) (TrackingToken, error) {
	const query = `
		SELECT token, campaign_id, restaurant_id, demo_site_id, token_type, target_url, expires_at, created_at
		FROM email_tracking_tokens
		WHERE token = $1`

	var record TrackingToken
	err := repo.pool.QueryRow(ctx, query, token).Scan(
		&record.Token,
		&record.CampaignID,
		&record.RestaurantID,
		&record.DemoSiteID,
		&record.TokenType,
		&record.TargetURL,
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
	const query = `
		INSERT INTO email_suppressions (email, reason)
		VALUES (lower($1), $2)
		ON CONFLICT (email) DO NOTHING`
	_, err := repo.pool.Exec(ctx, query, strings.TrimSpace(email), reason)
	if err != nil {
		return fmt.Errorf("add suppression: %w", err)
	}
	return nil
}

func (repo *Postgres) GetSendContext(ctx context.Context, campaignID uuid.UUID) (SendContext, error) {
	const query = `
		SELECT
			r.email,
			r.name,
			COALESCE(p.review_status, 'draft'),
			d.status,
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
		&ctxData.DemoStatus,
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
			'',
			''
		FROM restaurants r
		LEFT JOIN restaurant_profiles p ON p.restaurant_id = r.id
		WHERE r.id = $1`

	var ctxData SendContext
	err := repo.pool.QueryRow(ctx, query, restaurantID).Scan(
		&ctxData.RestaurantEmail,
		&ctxData.RestaurantName,
		&ctxData.ReviewStatus,
		&ctxData.DemoStatus,
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
		    status = 'emailed',
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
