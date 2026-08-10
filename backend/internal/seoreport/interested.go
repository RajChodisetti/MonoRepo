package seoreport

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

var (
	ErrInvalidName     = errors.New("invalid contact name")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPhone    = errors.New("invalid phone")
	ErrInvalidOTP      = errors.New("invalid or expired otp")
	ErrInvalidUnlock   = errors.New("invalid unlock token")
	ErrUnlockRateLimit = errors.New("unlock verification requested too recently")
	ErrEmailSendFailed = errors.New("failed to send verification email")
)

const (
	otpTTL            = 15 * time.Minute
	otpResendCooldown = 60 * time.Second
	unlockTTL         = 7 * 24 * time.Hour
	maxOTPAttempts    = 5
	unlockTokenBytes  = 24
)

// InterestedRecord is a row in seo_interested.
type InterestedRecord struct {
	ID               uuid.UUID
	Email            string
	PlaceID          string
	RestaurantName   string
	ContactName      string
	ContactPhone     string
	OTPHash          string
	OTPExpiresAt     *time.Time
	OTPRequestedAt   time.Time
	OTPAttempts      int
	VerifiedAt       *time.Time
	Interested       bool
	UnlockToken      string
	UnlockExpiresAt  time.Time
	LeadRestaurantID *uuid.UUID
}

// InterestedRepository persists SEO unlock / interested rows.
type InterestedRepository interface {
	UpsertPending(ctx context.Context, email, placeID, restaurantName, contactName, contactPhone, otpHash, unlockToken string, otpExpires time.Time) (InterestedRecord, error)
	GetByEmailPlace(ctx context.Context, email, placeID string) (InterestedRecord, error)
	GetByUnlockToken(ctx context.Context, token string) (InterestedRecord, error)
	RecordFailedOTP(ctx context.Context, id uuid.UUID) (InterestedRecord, error)
	MarkVerified(ctx context.Context, id uuid.UUID, leadRestaurantID uuid.UUID, unlockExpires time.Time) (InterestedRecord, error)
	MarkEmailVerified(ctx context.Context, id uuid.UUID, unlockExpires time.Time) (InterestedRecord, error)
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
	email, placeID, restaurantName, contactName, contactPhone, otpHash, unlockToken string,
	otpExpires time.Time,
) (InterestedRecord, error) {
	const query = `
			INSERT INTO seo_interested (
				email, place_id, restaurant_name, contact_name, contact_phone,
				otp_hash, otp_expires_at, otp_requested_at, otp_attempts,
				unlock_token, unlock_expires_at, interested, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, now(), 0, $8, $7, false, now())
			ON CONFLICT (email, place_id) DO UPDATE SET
				restaurant_name = EXCLUDED.restaurant_name,
				contact_name = EXCLUDED.contact_name,
				contact_phone = EXCLUDED.contact_phone,
				otp_hash = EXCLUDED.otp_hash,
				otp_expires_at = EXCLUDED.otp_expires_at,
				otp_requested_at = now(),
				otp_attempts = 0,
				verified_at = NULL,
				unlock_token = EXCLUDED.unlock_token,
				unlock_expires_at = EXCLUDED.unlock_expires_at,
				updated_at = now()
			WHERE seo_interested.otp_requested_at <= now() - ($9 * interval '1 second')
			RETURNING id, email, place_id, restaurant_name, contact_name, contact_phone,
			          otp_hash, otp_expires_at, otp_requested_at, otp_attempts, verified_at,
			          interested, unlock_token, unlock_expires_at, lead_restaurant_id`

	record, err := scanInterested(repo.pool.QueryRow(
		ctx,
		query,
		email,
		placeID,
		restaurantName,
		contactName,
		contactPhone,
		otpHash,
		otpExpires,
		unlockToken,
		int(otpResendCooldown.Seconds()),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return InterestedRecord{}, ErrUnlockRateLimit
	}
	return record, err
}

func (repo *postgresInterested) GetByEmailPlace(ctx context.Context, email, placeID string) (InterestedRecord, error) {
	const query = `
			SELECT id, email, place_id, restaurant_name, contact_name, contact_phone,
			       otp_hash, otp_expires_at, otp_requested_at, otp_attempts, verified_at,
			       interested, unlock_token, unlock_expires_at, lead_restaurant_id
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
			SELECT id, email, place_id, restaurant_name, contact_name, contact_phone,
			       otp_hash, otp_expires_at, otp_requested_at, otp_attempts, verified_at,
			       interested, unlock_token, unlock_expires_at, lead_restaurant_id
			FROM seo_interested
			WHERE unlock_token = $1`
	rec, err := scanInterested(repo.pool.QueryRow(ctx, query, token))
	if errors.Is(err, pgx.ErrNoRows) {
		return InterestedRecord{}, repository.ErrNotFound
	}
	return rec, err
}

func (repo *postgresInterested) RecordFailedOTP(ctx context.Context, id uuid.UUID) (InterestedRecord, error) {
	const query = `
			UPDATE seo_interested
			SET otp_attempts = LEAST(otp_attempts + 1, $2),
			    otp_hash = CASE WHEN otp_attempts + 1 >= $2 THEN '' ELSE otp_hash END,
			    otp_expires_at = CASE WHEN otp_attempts + 1 >= $2 THEN NULL ELSE otp_expires_at END,
			    updated_at = now()
			WHERE id = $1 AND verified_at IS NULL AND otp_attempts < $2
			RETURNING id, email, place_id, restaurant_name, contact_name, contact_phone,
			          otp_hash, otp_expires_at, otp_requested_at, otp_attempts, verified_at,
			          interested, unlock_token, unlock_expires_at, lead_restaurant_id`
	record, err := scanInterested(repo.pool.QueryRow(ctx, query, id, maxOTPAttempts))
	if errors.Is(err, pgx.ErrNoRows) {
		return InterestedRecord{}, ErrInvalidOTP
	}
	return record, err
}

func (repo *postgresInterested) MarkVerified(
	ctx context.Context,
	id uuid.UUID,
	leadRestaurantID uuid.UUID,
	unlockExpires time.Time,
) (InterestedRecord, error) {
	const query = `
			UPDATE seo_interested
			SET verified_at = now(),
			    lead_restaurant_id = $2,
			    otp_hash = '',
			    otp_expires_at = NULL,
			    unlock_expires_at = $3,
			    updated_at = now()
			WHERE id = $1
			  AND verified_at IS NULL
			  AND otp_hash <> ''
			  AND otp_expires_at >= now()
			  AND otp_attempts < $4
			RETURNING id, email, place_id, restaurant_name, contact_name, contact_phone,
			          otp_hash, otp_expires_at, otp_requested_at, otp_attempts, verified_at,
			          interested, unlock_token, unlock_expires_at, lead_restaurant_id`
	record, err := scanInterested(repo.pool.QueryRow(ctx, query, id, leadRestaurantID, unlockExpires, maxOTPAttempts))
	if errors.Is(err, pgx.ErrNoRows) {
		return InterestedRecord{}, ErrInvalidOTP
	}
	return record, err
}

func (repo *postgresInterested) MarkEmailVerified(
	ctx context.Context,
	id uuid.UUID,
	unlockExpires time.Time,
) (InterestedRecord, error) {
	const query = `
			UPDATE seo_interested
			SET verified_at = COALESCE(verified_at, now()),
			    otp_hash = '',
			    otp_expires_at = NULL,
			    unlock_expires_at = CASE WHEN verified_at IS NULL THEN $2 ELSE unlock_expires_at END,
			    updated_at = now()
			WHERE id = $1 AND unlock_expires_at >= now()
			RETURNING id, email, place_id, restaurant_name, contact_name, contact_phone,
			          otp_hash, otp_expires_at, otp_requested_at, otp_attempts, verified_at,
			          interested, unlock_token, unlock_expires_at, lead_restaurant_id`
	record, err := scanInterested(repo.pool.QueryRow(ctx, query, id, unlockExpires))
	if errors.Is(err, pgx.ErrNoRows) {
		return InterestedRecord{}, ErrInvalidUnlock
	}
	return record, err
}

func scanInterested(row pgx.Row) (InterestedRecord, error) {
	var rec InterestedRecord
	err := row.Scan(
		&rec.ID,
		&rec.Email,
		&rec.PlaceID,
		&rec.RestaurantName,
		&rec.ContactName,
		&rec.ContactPhone,
		&rec.OTPHash,
		&rec.OTPExpiresAt,
		&rec.OTPRequestedAt,
		&rec.OTPAttempts,
		&rec.VerifiedAt,
		&rec.Interested,
		&rec.UnlockToken,
		&rec.UnlockExpiresAt,
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
				    updated_at = now()
				WHERE id = $1`, restaurantID, name, email)
		if err != nil {
			return uuid.Nil, fmt.Errorf("update lead restaurant: %w", err)
		}
	} else {
		err = tx.QueryRow(ctx, `
				INSERT INTO restaurants (name, email, status, shown_interest)
				VALUES ($1, $2, 'lead', false)
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
	if email == "" || len(email) > 254 {
		return "", ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address == "" || len(addr.Address) > 254 || !strings.Contains(addr.Address, "@") {
		return "", ErrInvalidEmail
	}
	return strings.ToLower(addr.Address), nil
}

func normalizeContactName(raw string) (string, error) {
	name := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	length := utf8.RuneCountInString(name)
	if length < 2 || length > 100 {
		return "", ErrInvalidName
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", ErrInvalidName
		}
	}
	return name, nil
}

func normalizeContactPhone(raw string) (string, error) {
	phone := strings.TrimSpace(raw)
	if phone == "" || utf8.RuneCountInString(phone) > 40 {
		return "", ErrInvalidPhone
	}
	digits := 0
	for _, r := range phone {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == '+', r == '-', r == '(', r == ')', r == '.', unicode.IsSpace(r) && !unicode.IsControl(r):
			// Common international phone formatting is accepted and preserved.
		default:
			return "", ErrInvalidPhone
		}
	}
	if digits < 7 || digits > 20 {
		return "", ErrInvalidPhone
	}
	return phone, nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func isSixDigitOTP(otp string) bool {
	if len(otp) != 6 {
		return false
	}
	for _, char := range otp {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func hashOTP(secret, email, placeID, otp string) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("otp secret is unavailable")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strings.ToLower(strings.TrimSpace(email))))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(sanitizePlaceID(placeID)))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(strings.TrimSpace(otp)))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func otpHashMatches(secret, email, placeID, otp, storedHash string) bool {
	expectedHex, err := hashOTP(secret, email, placeID, otp)
	if err != nil {
		return false
	}
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return false
	}
	stored, err := hex.DecodeString(strings.TrimSpace(storedHash))
	if err != nil || len(stored) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare(expected, stored) == 1
}

func generateUnlockToken() (string, error) {
	buf := make([]byte, unlockTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
