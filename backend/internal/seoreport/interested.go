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

type unlockRateLimitError struct {
	retryAfter time.Duration
}

func (err *unlockRateLimitError) Error() string {
	return ErrUnlockRateLimit.Error()
}

func (err *unlockRateLimitError) Unwrap() error {
	return ErrUnlockRateLimit
}

// UnlockRetryAfterSeconds returns the bounded delay associated with an unlock
// throttle. Repository race/cooldown errors without a richer value use 60s.
func UnlockRetryAfterSeconds(err error) int {
	retryAfter := otpResendCooldown
	var limited *unlockRateLimitError
	if errors.As(err, &limited) && limited.retryAfter > 0 {
		retryAfter = limited.retryAfter
	}
	seconds := int((retryAfter + time.Second - 1) / time.Second)
	if seconds < 1 {
		return 1
	}
	return seconds
}

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
	MarkVerified(ctx context.Context, id uuid.UUID, unlockExpires time.Time) (InterestedRecord, error)
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
	unlockExpires time.Time,
) (InterestedRecord, error) {
	const query = `
			UPDATE seo_interested
			SET verified_at = now(),
			    lead_restaurant_id = NULL,
			    otp_hash = '',
			    otp_expires_at = NULL,
			    unlock_expires_at = $2,
			    updated_at = now()
			WHERE id = $1
			  AND verified_at IS NULL
			  AND otp_hash <> ''
			  AND otp_expires_at >= now()
			  AND otp_attempts < $3
			RETURNING id, email, place_id, restaurant_name, contact_name, contact_phone,
			          otp_hash, otp_expires_at, otp_requested_at, otp_attempts, verified_at,
			          interested, unlock_token, unlock_expires_at, lead_restaurant_id`
	record, err := scanInterested(repo.pool.QueryRow(ctx, query, id, unlockExpires, maxOTPAttempts))
	if errors.Is(err, pgx.ErrNoRows) {
		return InterestedRecord{}, ErrInvalidOTP
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
