// Package leadprep creates reviewable demo and outreach draft artifacts after
// OCR verification. It never publishes, approves, enqueues, or sends outreach.
package leadprep

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

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	platformdb "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

var ErrLeadNotOCRVerified = errors.New("lead is not OCR verified")

type Result struct {
	RestaurantID uuid.UUID
	DemoSiteID   uuid.UUID
	CampaignID   uuid.UUID
	Created      bool
}

type Service struct {
	pool     *pgxpool.Pool
	tokenTTL time.Duration
	appURLs  config.AppURLsConfig
}

func NewService(pool *pgxpool.Pool, tokenTTL time.Duration, appURLs config.AppURLsConfig) *Service {
	return &Service{pool: pool, tokenTTL: tokenTTL, appURLs: appURLs}
}

func (service *Service) Prepare(ctx context.Context, restaurantID uuid.UUID) (Result, error) {
	if service.pool == nil {
		return Result{}, fmt.Errorf("database pool is not configured")
	}

	tx, err := service.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Result{}, fmt.Errorf("begin lead preparation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Serialize preparation with import, review, campaign, and delivery state.
	if err := platformdb.LockRestaurantWorkflow(ctx, tx, restaurantID); err != nil {
		return Result{}, err
	}

	profile, err := loadVerifiedProfile(ctx, tx, restaurantID)
	if err != nil {
		return Result{}, err
	}
	payload, profileFingerprint, err := buildPublicPayload(ctx, tx, profile.RestaurantID)
	if err != nil {
		return Result{}, err
	}

	// Never create a second outreach campaign automatically. Existing manual or
	// non-draft campaigns are operator-owned. Only a stale, still-unapproved
	// automatic draft may be refreshed for new OCR or profile provenance.
	var existingCampaignID uuid.UUID
	var existingDemoID uuid.UUID
	var existingCampaignStatus string
	var existingCampaignAuto bool
	var existingCampaignOCRFingerprint string
	var existingCampaignProfileFingerprint string
	var existingDemoOCRFingerprint string
	var existingDemoProfileFingerprint string
	err = tx.QueryRow(ctx, `
			SELECT c.id,
			       c.demo_site_id,
			       c.status,
			       c.auto_generated,
			       c.source_ocr_fingerprint,
			       c.source_profile_fingerprint,
			       COALESCE(d.source_ocr_fingerprint, ''),
			       COALESCE(d.source_profile_fingerprint, '')
			FROM email_campaigns c
			LEFT JOIN demo_sites d ON d.id = c.demo_site_id
			WHERE c.restaurant_id = $1 AND c.campaign_type = $2
			ORDER BY c.created_at DESC
			LIMIT 1`, restaurantID, campaigns.TypeOutreach).Scan(
		&existingCampaignID,
		&existingDemoID,
		&existingCampaignStatus,
		&existingCampaignAuto,
		&existingCampaignOCRFingerprint,
		&existingCampaignProfileFingerprint,
		&existingDemoOCRFingerprint,
		&existingDemoProfileFingerprint,
	)
	refreshExistingDraft := false
	if err == nil {
		refreshExistingDraft = existingCampaignAuto &&
			existingCampaignStatus == campaigns.StatusDraft &&
			(existingCampaignOCRFingerprint != profile.OCRFingerprint ||
				existingCampaignProfileFingerprint != profileFingerprint ||
				existingDemoOCRFingerprint != profile.OCRFingerprint ||
				existingDemoProfileFingerprint != profileFingerprint)
		if !refreshExistingDraft {
			if err := markRestaurantDemoReady(ctx, tx, restaurantID); err != nil {
				return Result{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return Result{}, fmt.Errorf("commit existing lead preparation: %w", err)
			}
			return Result{
				RestaurantID: restaurantID,
				DemoSiteID:   existingDemoID,
				CampaignID:   existingCampaignID,
				Created:      false,
			}, nil
		}
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Result{}, fmt.Errorf("find existing outreach draft: %w", err)
	}

	token, err := demos.GenerateDemoToken()
	if err != nil {
		return Result{}, err
	}
	tokenHash, err := demos.HashDemoToken(token)
	if err != nil {
		return Result{}, err
	}

	slug := automaticDemoSlug(profile.Name, restaurantID)
	var demoSiteID uuid.UUID
	var demoRestaurantID uuid.UUID
	var demoStatus string
	var demoAuto bool
	if refreshExistingDraft {
		err = tx.QueryRow(ctx, `
			SELECT id, restaurant_id, status, auto_generated
			FROM demo_sites
			WHERE id = $1
			FOR UPDATE`, existingDemoID).Scan(
			&demoSiteID,
			&demoRestaurantID,
			&demoStatus,
			&demoAuto,
		)
	} else {
		err = tx.QueryRow(ctx, `
			SELECT id, restaurant_id, status, auto_generated
			FROM demo_sites
			WHERE slug = $1
			FOR UPDATE`, slug).Scan(
			&demoSiteID,
			&demoRestaurantID,
			&demoStatus,
			&demoAuto,
		)
	}

	var expiresAt *time.Time
	if service.tokenTTL > 0 {
		expiry := time.Now().UTC().Add(service.tokenTTL)
		expiresAt = &expiry
	}

	switch {
	case errors.Is(err, pgx.ErrNoRows) && refreshExistingDraft:
		return Result{}, fmt.Errorf("automatic campaign draft references a missing demo draft")
	case errors.Is(err, pgx.ErrNoRows):
		err = tx.QueryRow(ctx, `
				INSERT INTO demo_sites (
					restaurant_id, slug, token_hash, status, public_payload, expires_at,
					auto_generated, source_ocr_fingerprint, source_profile_fingerprint
				) VALUES ($1, $2, $3, $4, $5, $6, true, $7, $8)
				RETURNING id`,
			restaurantID,
			slug,
			tokenHash,
			demos.StatusDraft,
			payload,
			expiresAt,
			profile.OCRFingerprint,
			profileFingerprint,
		).Scan(&demoSiteID)
		if err != nil {
			return Result{}, fmt.Errorf("create automatic demo draft: %w", err)
		}
	case err != nil:
		return Result{}, fmt.Errorf("find automatic demo draft: %w", err)
	case demoRestaurantID != restaurantID:
		return Result{}, fmt.Errorf("automatic demo slug collision")
	case !demoAuto || demoStatus != demos.StatusDraft:
		return Result{}, fmt.Errorf("automatic campaign draft does not own an editable demo draft")
	default:
		// Only provenance-marked draft artifacts may be rotated automatically.
		result, updateErr := tx.Exec(ctx, `
				UPDATE demo_sites
				SET slug = $2,
					token_hash = $3,
					status = $4,
					public_payload = $5,
					expires_at = $6,
					source_ocr_fingerprint = $7,
					source_profile_fingerprint = $8,
					published_at = NULL,
					published_by = NULL,
					updated_at = now()
				WHERE id = $1
				  AND auto_generated = true
				  AND status = $4`,
			demoSiteID,
			slug,
			tokenHash,
			demos.StatusDraft,
			payload,
			expiresAt,
			profile.OCRFingerprint,
			profileFingerprint,
		)
		if updateErr != nil {
			return Result{}, fmt.Errorf("refresh automatic demo draft: %w", updateErr)
		}
		if result.RowsAffected() != 1 {
			return Result{}, fmt.Errorf("automatic demo draft changed while it was being refreshed")
		}
	}

	draft := campaigns.BuildDraft(campaigns.DraftInput{
		RestaurantName:      profile.Name,
		PresentationSiteURL: service.appURLs.PresentationSiteURL,
		MarketingSiteURL:    service.appURLs.PublicMarketingURL,
	})
	campaignID := existingCampaignID
	if refreshExistingDraft {
		result, updateErr := tx.Exec(ctx, `
			UPDATE email_campaigns
			SET demo_site_id = $2,
				subject = $3,
				body_html = $4,
				body_text = $5,
				demo_token = $6,
				source_ocr_fingerprint = $7,
				source_profile_fingerprint = $8,
				approved_at = NULL,
				approved_by = NULL,
				updated_at = now()
			WHERE id = $1
			  AND status = $9
			  AND auto_generated = true`,
			existingCampaignID,
			demoSiteID,
			draft.Subject,
			draft.BodyHTML,
			draft.BodyText,
			token,
			profile.OCRFingerprint,
			profileFingerprint,
			campaigns.StatusDraft,
		)
		if updateErr != nil {
			return Result{}, fmt.Errorf("refresh automatic campaign draft: %w", updateErr)
		}
		if result.RowsAffected() != 1 {
			return Result{}, fmt.Errorf("automatic campaign draft changed while it was being refreshed")
		}
	} else {
		err = tx.QueryRow(ctx, `
			INSERT INTO email_campaigns (
				restaurant_id, demo_site_id, campaign_type, status,
			subject, body_html, body_text, demo_token,
			auto_generated, source_ocr_fingerprint, source_profile_fingerprint
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, $9, $10)
			RETURNING id`,
			restaurantID,
			demoSiteID,
			campaigns.TypeOutreach,
			campaigns.StatusDraft,
			draft.Subject,
			draft.BodyHTML,
			draft.BodyText,
			token,
			profile.OCRFingerprint,
			profileFingerprint,
		).Scan(&campaignID)
		if err != nil {
			return Result{}, fmt.Errorf("create automatic campaign draft: %w", err)
		}
	}

	if err := markRestaurantDemoReady(ctx, tx, restaurantID); err != nil {
		return Result{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit lead preparation: %w", err)
	}
	return Result{
		RestaurantID: restaurantID,
		DemoSiteID:   demoSiteID,
		CampaignID:   campaignID,
		Created:      !refreshExistingDraft,
	}, nil
}

func markRestaurantDemoReady(ctx context.Context, tx pgx.Tx, restaurantID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
		UPDATE restaurants
		SET status = $2,
		    updated_at = now()
		WHERE id = $1
		  AND status = $3`,
		restaurantID,
		restaurants.StatusDemoReady,
		restaurants.StatusLead,
	); err != nil {
		return fmt.Errorf("mark restaurant demo ready: %w", err)
	}
	return nil
}

type verifiedProfile struct {
	RestaurantID uuid.UUID
	Name         string

	OCRFingerprint string
}

func loadVerifiedProfile(ctx context.Context, tx pgx.Tx, restaurantID uuid.UUID) (verifiedProfile, error) {
	var profile verifiedProfile
	profile.RestaurantID = restaurantID
	// Fence identity first and profile second, matching import/review lock order.
	// The importer holds the profile write lock through menu synchronization, so
	// these locks also prevent a mixed identity/profile/menu payload snapshot.
	err := tx.QueryRow(ctx, `
		SELECT name
		FROM restaurants
		WHERE id = $1
		FOR SHARE`, restaurantID).Scan(&profile.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return verifiedProfile{}, ErrLeadNotOCRVerified
	}
	if err != nil {
		return verifiedProfile{}, fmt.Errorf("lock lead identity for preparation: %w", err)
	}
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(NULLIF(ocr_input_fingerprint, ''), 'legacy')
		FROM restaurant_profiles
		WHERE restaurant_id = $1 AND ocr_status = 'verified'
		FOR SHARE`, restaurantID).Scan(&profile.OCRFingerprint)
	if errors.Is(err, pgx.ErrNoRows) {
		return verifiedProfile{}, ErrLeadNotOCRVerified
	}
	if err != nil {
		return verifiedProfile{}, fmt.Errorf("lock OCR-verified lead profile: %w", err)
	}
	return profile, nil
}

func buildPublicPayload(
	ctx context.Context,
	tx pgx.Tx,
	restaurantID uuid.UUID,
) (json.RawMessage, string, error) {
	var payload []byte
	var profileFingerprint string
	err := tx.QueryRow(ctx, `
		SELECT lead_artifact_current_public_payload($1),
		       lead_artifact_current_profile_fingerprint($1)`, restaurantID).Scan(
		&payload,
		&profileFingerprint,
	)
	if err != nil {
		return nil, "", fmt.Errorf("build demo payload provenance: %w", err)
	}
	if len(payload) == 0 || strings.TrimSpace(profileFingerprint) == "" {
		return nil, "", fmt.Errorf("build demo payload provenance: restaurant profile is unavailable")
	}
	return json.RawMessage(payload), profileFingerprint, nil
}

func automaticDemoSlug(name string, restaurantID uuid.UUID) string {
	var builder strings.Builder
	lastDash := false
	for _, char := range strings.ToLower(strings.TrimSpace(name)) {
		valid := (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
		if valid {
			builder.WriteRune(char)
			lastDash = false
			continue
		}
		if builder.Len() > 0 && !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	base := strings.Trim(builder.String(), "-")
	if base == "" {
		base = "restaurant"
	}
	if len(base) > 60 {
		base = strings.Trim(base[:60], "-")
	}
	shortID := strings.ReplaceAll(restaurantID.String(), "-", "")[:8]
	return base + "-" + shortID + "-auto"
}
