package seoreport

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

var (
	ErrInvalidEmail   = errors.New("invalid email")
	ErrInvalidOTP     = errors.New("invalid or expired otp")
	ErrInvalidUnlock  = errors.New("invalid unlock token")
	ErrEmailSendFailed = errors.New("failed to send verification email")
)

const (
	otpTTL          = 15 * time.Minute
	unlockTokenBytes = 24
)

// InterestedRecord is a row in seo_interested.
type InterestedRecord struct {
	ID               uuid.UUID
	Email            string
	PlaceID          string
	RestaurantName   string
	OTPHash          string
	OTPExpiresAt     *time.Time
	VerifiedAt       *time.Time
	Interested       bool
	UnlockToken      string
	LeadRestaurantID *uuid.UUID
}

// InterestedRepository persists SEO unlock / interested rows.
type InterestedRepository interface {
	UpsertPending(ctx context.Context, email, placeID, restaurantName, otpHash, unlockToken string, otpExpires time.Time) (InterestedRecord, error)
	GetByEmailPlace(ctx context.Context, email, placeID string) (InterestedRecord, error)
	GetByUnlockToken(ctx context.Context, token string) (InterestedRecord, error)
	MarkVerified(ctx context.Context, id uuid.UUID, leadRestaurantID uuid.UUID) (InterestedRecord, error)
	MarkInterested(ctx context.Context, id uuid.UUID) (InterestedRecord, error)
}

// LeadUpserter creates/updates a restaurant lead from Places data.
type LeadUpserter interface {
	UpsertLeadFromPlace(ctx context.Context, place PlaceDetails, contactEmail string) (uuid.UUID, error)
}

type postgresInterested struct {
	pool *pgxpool.Pool
}

// NewInterestedPostgres creates the seo_interested repository.
func NewInterestedPostgres(pool *pgxpool.Pool) InterestedRepository {
	return &postgresInterested{pool: pool}
}

func (repo *postgresInterested) UpsertPending(
	ctx context.Context,
	email, placeID, restaurantName, otpHash, unlockToken string,
	otpExpires time.Time,
) (InterestedRecord, error) {
	const query = `
		INSERT INTO seo_interested (
			email, place_id, restaurant_name, otp_hash, otp_expires_at, unlock_token, interested, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, false, now())
		ON CONFLICT (email, place_id) DO UPDATE SET
			restaurant_name = EXCLUDED.restaurant_name,
			otp_hash = EXCLUDED.otp_hash,
			otp_expires_at = EXCLUDED.otp_expires_at,
			unlock_token = EXCLUDED.unlock_token,
			updated_at = now()
		RETURNING id, email, place_id, restaurant_name, otp_hash, otp_expires_at, verified_at,
		          interested, unlock_token, lead_restaurant_id`

	return scanInterested(repo.pool.QueryRow(ctx, query, email, placeID, restaurantName, otpHash, otpExpires, unlockToken))
}

func (repo *postgresInterested) GetByEmailPlace(ctx context.Context, email, placeID string) (InterestedRecord, error) {
	const query = `
		SELECT id, email, place_id, restaurant_name, otp_hash, otp_expires_at, verified_at,
		       interested, unlock_token, lead_restaurant_id
		FROM seo_interested
		WHERE email = $1 AND place_id = $2`
	rec, err := scanInterested(repo.pool.QueryRow(ctx, query, email, placeID))
	if errors.Is(err, pgx.ErrNoRows) {
		return InterestedRecord{}, repository.ErrNotFound
	}
	return rec, err
}

func (repo *postgresInterested) GetByUnlockToken(ctx context.Context, token string) (InterestedRecord, error) {
	const query = `
		SELECT id, email, place_id, restaurant_name, otp_hash, otp_expires_at, verified_at,
		       interested, unlock_token, lead_restaurant_id
		FROM seo_interested
		WHERE unlock_token = $1`
	rec, err := scanInterested(repo.pool.QueryRow(ctx, query, token))
	if errors.Is(err, pgx.ErrNoRows) {
		return InterestedRecord{}, repository.ErrNotFound
	}
	return rec, err
}

func (repo *postgresInterested) MarkVerified(ctx context.Context, id uuid.UUID, leadRestaurantID uuid.UUID) (InterestedRecord, error) {
	const query = `
		UPDATE seo_interested
		SET verified_at = COALESCE(verified_at, now()),
		    interested = true,
		    lead_restaurant_id = $2,
		    otp_hash = '',
		    otp_expires_at = NULL,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, email, place_id, restaurant_name, otp_hash, otp_expires_at, verified_at,
		          interested, unlock_token, lead_restaurant_id`
	return scanInterested(repo.pool.QueryRow(ctx, query, id, leadRestaurantID))
}

func (repo *postgresInterested) MarkInterested(ctx context.Context, id uuid.UUID) (InterestedRecord, error) {
	const query = `
		UPDATE seo_interested
		SET interested = true,
		    verified_at = COALESCE(verified_at, now()),
		    updated_at = now()
		WHERE id = $1
		RETURNING id, email, place_id, restaurant_name, otp_hash, otp_expires_at, verified_at,
		          interested, unlock_token, lead_restaurant_id`
	return scanInterested(repo.pool.QueryRow(ctx, query, id))
}

func scanInterested(row pgx.Row) (InterestedRecord, error) {
	var rec InterestedRecord
	err := row.Scan(
		&rec.ID,
		&rec.Email,
		&rec.PlaceID,
		&rec.RestaurantName,
		&rec.OTPHash,
		&rec.OTPExpiresAt,
		&rec.VerifiedAt,
		&rec.Interested,
		&rec.UnlockToken,
		&rec.LeadRestaurantID,
	)
	if err != nil {
		return InterestedRecord{}, err
	}
	return rec, nil
}

type postgresLeadUpserter struct {
	pool *pgxpool.Pool
}

// NewLeadUpserter creates a Places→restaurants lead writer.
func NewLeadUpserter(pool *pgxpool.Pool) LeadUpserter {
	return &postgresLeadUpserter{pool: pool}
}

func (u *postgresLeadUpserter) UpsertLeadFromPlace(ctx context.Context, place PlaceDetails, contactEmail string) (uuid.UUID, error) {
	placeID := sanitizePlaceID(place.PlaceID)
	name := firstNonEmpty(strings.TrimSpace(place.Name), "Restaurant")
	email := strings.ToLower(strings.TrimSpace(contactEmail))
	if email == "" {
		email = strings.ToLower(strings.TrimSpace(place.Email))
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin lead upsert: %w", err)
	}
	defer tx.Rollback(ctx)

	var restaurantID uuid.UUID
	if placeID != "" {
		err = tx.QueryRow(ctx, `
			SELECT restaurant_id FROM restaurant_profiles WHERE google_place_id = $1`, placeID,
		).Scan(&restaurantID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("lookup lead by place: %w", err)
		}
	}

	if restaurantID != uuid.Nil {
		_, err = tx.Exec(ctx, `
			UPDATE restaurants
			SET name = $2,
			    email = CASE WHEN $3 <> '' THEN $3 ELSE email END,
			    shown_interest = true,
			    status = CASE
			      WHEN status IN ('lead', 'demo_ready', 'emailed') THEN 'interested'
			      ELSE status
			    END,
			    updated_at = now()
			WHERE id = $1`, restaurantID, name, email)
		if err != nil {
			return uuid.Nil, fmt.Errorf("update lead restaurant: %w", err)
		}
	} else {
		err = tx.QueryRow(ctx, `
			INSERT INTO restaurants (name, email, status, shown_interest)
			VALUES ($1, $2, 'interested', true)
			RETURNING id`, name, email,
		).Scan(&restaurantID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert lead restaurant: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO restaurant_profiles (
			restaurant_id, phone, website, address, google_place_id, rating, reviews_count,
			price_level, scrape_status, raw_public_data, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, 'seo_unlock', '{}'::jsonb, now()
		)
		ON CONFLICT (restaurant_id) DO UPDATE SET
			phone = CASE WHEN EXCLUDED.phone <> '' THEN EXCLUDED.phone ELSE restaurant_profiles.phone END,
			website = CASE WHEN EXCLUDED.website <> '' THEN EXCLUDED.website ELSE restaurant_profiles.website END,
			address = CASE WHEN EXCLUDED.address <> '' THEN EXCLUDED.address ELSE restaurant_profiles.address END,
			google_place_id = CASE
				WHEN EXCLUDED.google_place_id <> '' THEN EXCLUDED.google_place_id
				ELSE restaurant_profiles.google_place_id
			END,
			rating = COALESCE(EXCLUDED.rating, restaurant_profiles.rating),
			reviews_count = COALESCE(EXCLUDED.reviews_count, restaurant_profiles.reviews_count),
			price_level = CASE WHEN EXCLUDED.price_level <> '' THEN EXCLUDED.price_level ELSE restaurant_profiles.price_level END,
			scrape_status = 'seo_unlock',
			updated_at = now()`,
		restaurantID,
		nullString(place.Phone),
		nullString(place.Website),
		nullString(place.Address),
		nullString(placeID),
		place.Rating,
		place.UserRatingCount,
		nullString(place.PriceLevel),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert lead profile: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit lead upsert: %w", err)
	}
	return restaurantID, nil
}

func nullString(v string) string {
	return strings.TrimSpace(v)
}

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address == "" || !strings.Contains(addr.Address, "@") {
		return "", ErrInvalidEmail
	}
	return strings.ToLower(addr.Address), nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashOTP(otp string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(otp)))
	return hex.EncodeToString(sum[:])
}

func generateUnlockToken() (string, error) {
	buf := make([]byte, unlockTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
