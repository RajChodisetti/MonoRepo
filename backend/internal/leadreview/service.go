// Package leadreview owns explicit human review of restaurant profiles and
// demo publication. Outreach sequence eligibility is intentionally separate.
package leadreview

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

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	platformdb "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
)

const (
	ProfileDraft    = "draft"
	ProfileApproved = "approved"
	ProfileRejected = "rejected"

	DemoDraft     = "draft"
	DemoPublished = "published"
)

var (
	ErrForbidden          = errors.New("lead review requires an internal administrator")
	ErrInvalidStatus      = errors.New("invalid review status")
	ErrNotFound           = errors.New("review target not found")
	ErrProfileNotApproved = errors.New("restaurant profile is not approved")
	ErrDemoExpired        = errors.New("demo link has expired")
	ErrExpectedUpdatedAt  = errors.New("expected_updated_at is required for a review decision")
	ErrStaleReview        = errors.New("review target changed after it was inspected")
)

type ProfileReview struct {
	RestaurantID uuid.UUID  `json:"restaurant_id"`
	Status       string     `json:"status"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy   *uuid.UUID `json:"reviewed_by,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type DemoReview struct {
	DemoSiteID   uuid.UUID  `json:"demo_site_id"`
	RestaurantID uuid.UUID  `json:"restaurant_id"`
	Status       string     `json:"status"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	PublishedBy  *uuid.UUID `json:"published_by,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ProfileReviewPreview struct {
	RestaurantID        uuid.UUID       `json:"restaurant_id"`
	RestaurantName      string          `json:"restaurant_name"`
	ContactEmail        string          `json:"contact_email"`
	ApolloStatus        string          `json:"apollo_status"`
	ApolloEmailFound    bool            `json:"apollo_email_found"`
	ReviewStatus        string          `json:"review_status"`
	ReviewedAt          *time.Time      `json:"reviewed_at,omitempty"`
	ReviewedBy          *uuid.UUID      `json:"reviewed_by,omitempty"`
	RestaurantUpdatedAt time.Time       `json:"restaurant_updated_at"`
	ProfileUpdatedAt    time.Time       `json:"profile_updated_at"`
	Profile             json.RawMessage `json:"profile"`
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (service *Service) GetProfileReviewPreview(
	ctx context.Context,
	principal auth.Principal,
	restaurantID uuid.UUID,
) (ProfileReviewPreview, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return ProfileReviewPreview{}, ErrForbidden
	}
	if service.pool == nil {
		return ProfileReviewPreview{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT r.name,
		       r.email,
		       COALESCE(
		         NULLIF(rp.raw_public_data #>> '{apollo_enrichment,status}', ''),
		         CASE WHEN rp.apollo_lead <> '{}'::jsonb THEN 'enriched' ELSE 'not_recorded' END
		       ),
		       COALESCE(NULLIF(rp.apollo_lead #>> '{contact,email}', ''), '') <> '',
		       rp.review_status,
		       rp.reviewed_at,
		       rp.reviewed_by,
		       r.updated_at,
		       rp.updated_at,
		       jsonb_build_object(
		         'description', rp.description,
		         'opening_hours', rp.opening_hours,
		         'phone', rp.phone,
		         'website', rp.website,
		         'address', rp.address,
		         'city', rp.city,
		         'state', rp.state,
		         'country', rp.country,
		         'latitude', rp.latitude,
		         'longitude', rp.longitude,
		         'google_place_id', rp.google_place_id,
		         'rating', rp.rating,
		         'reviews_count', rp.reviews_count,
		         'price_level', rp.price_level,
		         'cuisines', rp.cuisines,
		         'owners', rp.owners,
		         'images', rp.images,
		         'dietary_options', rp.dietary_options,
		         'parking_info', rp.parking_info,
		         'reservation_policy', rp.reservation_policy,
		         'brand_tone', rp.brand_tone,
		         'scrape_status', rp.scrape_status,
		         'scrape_errors', rp.scrape_errors
		       )
		FROM restaurants r
		JOIN restaurant_profiles rp ON rp.restaurant_id = r.id
		WHERE r.id = $1`

	result := ProfileReviewPreview{RestaurantID: restaurantID}
	var profile []byte
	err := service.pool.QueryRow(ctx, query, restaurantID).Scan(
		&result.RestaurantName,
		&result.ContactEmail,
		&result.ApolloStatus,
		&result.ApolloEmailFound,
		&result.ReviewStatus,
		&result.ReviewedAt,
		&result.ReviewedBy,
		&result.RestaurantUpdatedAt,
		&result.ProfileUpdatedAt,
		&profile,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProfileReviewPreview{}, ErrNotFound
	}
	if err != nil {
		return ProfileReviewPreview{}, fmt.Errorf("load profile review preview: %w", err)
	}
	result.Profile = json.RawMessage(profile)
	return result, nil
}

// ReviewProfile records a human decision for demo/public content. Outreach
// eligibility depends on business identity, consent evidence and lifecycle.
func (service *Service) ReviewProfile(
	ctx context.Context,
	principal auth.Principal,
	restaurantID uuid.UUID,
	status string,
	expectedRestaurantUpdatedAt *time.Time,
	expectedProfileUpdatedAt *time.Time,
) (ProfileReview, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return ProfileReview{}, ErrForbidden
	}
	if service.pool == nil {
		return ProfileReview{}, fmt.Errorf("database pool is not configured")
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != ProfileDraft && status != ProfileApproved && status != ProfileRejected {
		return ProfileReview{}, ErrInvalidStatus
	}
	if status != ProfileDraft && (expectedRestaurantUpdatedAt == nil || expectedProfileUpdatedAt == nil) {
		return ProfileReview{}, ErrExpectedUpdatedAt
	}

	if status == ProfileDraft {
		tx, err := service.pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return ProfileReview{}, fmt.Errorf("begin clearing restaurant profile review: %w", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := platformdb.LockRestaurantWorkflow(ctx, tx, restaurantID); err != nil {
			return ProfileReview{}, err
		}
		const clearReview = `
			UPDATE restaurant_profiles
			SET review_status = 'draft',
			    reviewed_at = NULL,
			    reviewed_by = NULL,
			    updated_at = now()
			WHERE restaurant_id = $1
			RETURNING restaurant_id, review_status, reviewed_at, reviewed_by, updated_at`
		var result ProfileReview
		err = tx.QueryRow(ctx, clearReview, restaurantID).Scan(
			&result.RestaurantID,
			&result.Status,
			&result.ReviewedAt,
			&result.ReviewedBy,
			&result.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return ProfileReview{}, ErrNotFound
		}
		if err != nil {
			return ProfileReview{}, fmt.Errorf("clear restaurant profile review: %w", err)
		}
		if err := invalidateDownstreamApprovals(ctx, tx, restaurantID); err != nil {
			return ProfileReview{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return ProfileReview{}, fmt.Errorf("commit clearing restaurant profile review: %w", err)
		}
		return result, nil
	}

	tx, err := service.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ProfileReview{}, fmt.Errorf("begin restaurant profile review: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := platformdb.LockRestaurantWorkflow(ctx, tx, restaurantID); err != nil {
		return ProfileReview{}, err
	}

	// Import and admin update paths touch restaurant identity before profile
	// details, so the review gate uses the same lock order.
	var currentRestaurantUpdatedAt time.Time
	err = tx.QueryRow(ctx, `SELECT updated_at FROM restaurants WHERE id = $1 FOR UPDATE`, restaurantID).Scan(&currentRestaurantUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProfileReview{}, ErrNotFound
	}
	if err != nil {
		return ProfileReview{}, fmt.Errorf("lock restaurant identity for review: %w", err)
	}

	var currentProfileUpdatedAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT updated_at
		FROM restaurant_profiles
		WHERE restaurant_id = $1
		FOR UPDATE`, restaurantID).Scan(&currentProfileUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProfileReview{}, ErrNotFound
	}
	if err != nil {
		return ProfileReview{}, fmt.Errorf("lock restaurant profile for review: %w", err)
	}
	if !currentRestaurantUpdatedAt.Equal(*expectedRestaurantUpdatedAt) ||
		!currentProfileUpdatedAt.Equal(*expectedProfileUpdatedAt) {
		return ProfileReview{}, ErrStaleReview
	}
	const recordReview = `
		UPDATE restaurant_profiles
		SET review_status = $2,
		    reviewed_at = now(),
		    reviewed_by = $3,
		    updated_at = now()
		WHERE restaurant_id = $1 AND updated_at = $4
		RETURNING restaurant_id, review_status, reviewed_at, reviewed_by, updated_at`
	var result ProfileReview
	err = tx.QueryRow(ctx, recordReview, restaurantID, status, principal.UserID, *expectedProfileUpdatedAt).Scan(
		&result.RestaurantID,
		&result.Status,
		&result.ReviewedAt,
		&result.ReviewedBy,
		&result.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProfileReview{}, ErrStaleReview
	}
	if err != nil {
		return ProfileReview{}, fmt.Errorf("record restaurant profile review: %w", err)
	}
	// Every new profile decision invalidates downstream approvals. This keeps a
	// later identity/profile change from re-enabling an older campaign artifact.
	if err := invalidateDownstreamApprovals(ctx, tx, restaurantID); err != nil {
		return ProfileReview{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ProfileReview{}, fmt.Errorf("commit restaurant profile review: %w", err)
	}
	return result, nil
}

func invalidateDownstreamApprovals(ctx context.Context, tx pgx.Tx, restaurantID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
		UPDATE demo_sites
		SET status = 'draft',
		    published_at = NULL,
		    published_by = NULL,
		    updated_at = now()
		WHERE restaurant_id = $1 AND status = 'published'`, restaurantID); err != nil {
		return fmt.Errorf("invalidate published demos after profile review: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE email_campaigns
		SET status = 'draft',
		    approved_at = NULL,
		    approved_by = NULL,
		    updated_at = now()
		WHERE restaurant_id = $1 AND status = 'approved' AND sequence_id IS NULL`, restaurantID); err != nil {
		return fmt.Errorf("invalidate approved campaigns after profile review: %w", err)
	}
	return nil
}

// SetDemoStatus publishes or unpublishes an existing generated demo after
// human profile approval. It does not depend on retired image-analysis state.
func (service *Service) SetDemoStatus(
	ctx context.Context,
	principal auth.Principal,
	demoSiteID uuid.UUID,
	status string,
	expectedUpdatedAt *time.Time,
) (DemoReview, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return DemoReview{}, ErrForbidden
	}
	if service.pool == nil {
		return DemoReview{}, fmt.Errorf("database pool is not configured")
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != DemoDraft && status != DemoPublished {
		return DemoReview{}, ErrInvalidStatus
	}
	if status == DemoPublished && expectedUpdatedAt == nil {
		return DemoReview{}, ErrExpectedUpdatedAt
	}

	if status == DemoDraft {
		tx, err := service.pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return DemoReview{}, fmt.Errorf("begin returning demo to draft: %w", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		var restaurantID uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT restaurant_id FROM demo_sites WHERE id = $1`, demoSiteID).Scan(&restaurantID); errors.Is(err, pgx.ErrNoRows) {
			return DemoReview{}, ErrNotFound
		} else if err != nil {
			return DemoReview{}, fmt.Errorf("load demo restaurant: %w", err)
		}
		if err := platformdb.LockRestaurantWorkflow(ctx, tx, restaurantID); err != nil {
			return DemoReview{}, err
		}
		const unpublish = `
			UPDATE demo_sites
			SET status = 'draft',
			    published_at = NULL,
			    published_by = NULL,
			    updated_at = now()
			WHERE id = $1
			RETURNING id, restaurant_id, status, published_at, published_by, updated_at`
		var result DemoReview
		err = tx.QueryRow(ctx, unpublish, demoSiteID).Scan(
			&result.DemoSiteID,
			&result.RestaurantID,
			&result.Status,
			&result.PublishedAt,
			&result.PublishedBy,
			&result.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return DemoReview{}, ErrNotFound
		}
		if err != nil {
			return DemoReview{}, fmt.Errorf("return demo to draft: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return DemoReview{}, fmt.Errorf("commit returning demo to draft: %w", err)
		}
		return result, nil
	}

	tx, err := service.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DemoReview{}, fmt.Errorf("begin demo publication: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Lock restaurant identity, then profile, then demo. Import and review-reset
	// paths use the same order, preventing a concurrent change from leaving an
	// ineligible demo published.
	var restaurantID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT restaurant_id FROM demo_sites WHERE id = $1`, demoSiteID).Scan(&restaurantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return DemoReview{}, ErrNotFound
	}
	if err != nil {
		return DemoReview{}, fmt.Errorf("load demo restaurant: %w", err)
	}
	if err := platformdb.LockRestaurantWorkflow(ctx, tx, restaurantID); err != nil {
		return DemoReview{}, err
	}

	var restaurantUpdatedAt time.Time
	err = tx.QueryRow(ctx, `SELECT updated_at FROM restaurants WHERE id = $1 FOR UPDATE`, restaurantID).Scan(&restaurantUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return DemoReview{}, ErrNotFound
	}
	if err != nil {
		return DemoReview{}, fmt.Errorf("lock restaurant identity for demo publication: %w", err)
	}

	var reviewStatus string
	var reviewAudited bool
	var reviewedAt *time.Time
	var profileUpdatedAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT review_status,
		       (reviewed_at IS NOT NULL AND reviewed_by IS NOT NULL),
		       reviewed_at,
		       updated_at
		FROM restaurant_profiles
		WHERE restaurant_id = $1
		FOR UPDATE`, restaurantID).Scan(
		&reviewStatus,
		&reviewAudited,
		&reviewedAt,
		&profileUpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DemoReview{}, ErrProfileNotApproved
	}
	if err != nil {
		return DemoReview{}, fmt.Errorf("lock profile for demo publication: %w", err)
	}

	var expiresAt *time.Time
	var currentUpdatedAt time.Time
	var autoGenerated bool
	var sourceProfileFingerprint string
	var currentProfileFingerprint string
	err = tx.QueryRow(ctx, `
		SELECT expires_at,
		       updated_at,
		       auto_generated,
		       source_profile_fingerprint,
		       COALESCE(lead_artifact_current_profile_fingerprint($2), '')
		FROM demo_sites
		WHERE id = $1 AND restaurant_id = $2
		FOR UPDATE`, demoSiteID, restaurantID).Scan(
		&expiresAt,
		&currentUpdatedAt,
		&autoGenerated,
		&sourceProfileFingerprint,
		&currentProfileFingerprint,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DemoReview{}, ErrNotFound
	}
	if err != nil {
		return DemoReview{}, fmt.Errorf("lock demo for publication: %w", err)
	}

	if autoGenerated && (sourceProfileFingerprint == "" ||
		sourceProfileFingerprint != currentProfileFingerprint) {
		return DemoReview{}, ErrStaleReview
	}
	if reviewStatus != ProfileApproved || !reviewAudited || reviewedAt == nil ||
		reviewedAt.Before(restaurantUpdatedAt) || reviewedAt.Before(profileUpdatedAt) {
		return DemoReview{}, ErrProfileNotApproved
	}
	if expiresAt != nil && !expiresAt.After(time.Now().UTC()) {
		return DemoReview{}, ErrDemoExpired
	}
	if !currentUpdatedAt.Equal(*expectedUpdatedAt) {
		return DemoReview{}, ErrStaleReview
	}

	const publish = `
		UPDATE demo_sites
		SET status = 'published',
		    published_at = now(),
		    published_by = $2,
		    updated_at = now()
		WHERE id = $1 AND updated_at = $3
		RETURNING id, restaurant_id, status, published_at, published_by, updated_at`
	var result DemoReview
	err = tx.QueryRow(ctx, publish, demoSiteID, principal.UserID, *expectedUpdatedAt).Scan(
		&result.DemoSiteID,
		&result.RestaurantID,
		&result.Status,
		&result.PublishedAt,
		&result.PublishedBy,
		&result.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DemoReview{}, ErrStaleReview
	}
	if err != nil {
		return DemoReview{}, fmt.Errorf("publish demo: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return DemoReview{}, fmt.Errorf("commit demo publication: %w", err)
	}
	return result, nil
}
