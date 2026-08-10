package seoreport

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
	llmlib "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/llm"
)

// A process has one SEO report service in production. Keeping the unlock-only
// secret alongside that service avoids broadening the report orchestration
// struct with credential material that no other report path should access.
var serviceOTPSecrets sync.Map // map[*Service]string

// UnlockRequestResult is returned after sending an OTP email.
type UnlockRequestResult struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	Email        string `json:"email"`
	PlaceID      string `json:"placeId"`
	ExpiresInSec int    `json:"expiresInSec"`
	// DevOTP is only populated when APP_ENV is local/test and email sending is disabled.
	DevOTP string `json:"devOtp,omitempty"`
}

// UnlockVerifyResult is returned after successful OTP verification.
type UnlockVerifyResult struct {
	Status      string       `json:"status"`
	UnlockToken string       `json:"unlockToken"`
	Interested  bool         `json:"interested"`
	Place       PlaceDetails `json:"place"`
	Report      Report       `json:"report"`
}

// EmailSender sends unlock verification emails.
type EmailSender interface {
	Send(ctx context.Context, req emailprovider.SendRequest) (emailprovider.SendResult, error)
}

// NewServiceFull constructs SEO report + unlock services.
func NewServiceFull(
	placesCfg config.PlacesConfig,
	appCfg config.AppConfig,
	profilesRepo profiles.SiteRepository,
	interestedRepo InterestedRepository,
	mailer EmailSender,
	llmClient llmlib.Client,
	tokenSecret string,
	log *slog.Logger,
) *Service {
	svc := NewService(placesCfg, profilesRepo, llmClient, log)
	svc.appEnv = appCfg.Env
	svc.interested = interestedRepo
	svc.mailer = mailer
	svc.unlockRate = newUnlockRequestLimiter()
	serviceOTPSecrets.Store(svc, strings.TrimSpace(tokenSecret))
	return svc
}

// RequestUnlockEmail validates and throttles the requester before any report or
// provider work, stores a bounded OTP, and emails the code without a bearer link.
func (s *Service) RequestUnlockEmail(
	ctx context.Context,
	nameRaw, emailRaw, phoneRaw, placeID string,
) (UnlockRequestResult, error) {
	contactName, err := normalizeContactName(nameRaw)
	if err != nil {
		return UnlockRequestResult{}, err
	}
	email, err := normalizeEmail(emailRaw)
	if err != nil {
		return UnlockRequestResult{}, err
	}
	contactPhone, err := normalizeContactPhone(phoneRaw)
	if err != nil {
		return UnlockRequestResult{}, err
	}
	placeID = sanitizePlaceID(placeID)
	if placeID == "" {
		return UnlockRequestResult{}, ErrNotFound
	}
	if s.interested == nil {
		return UnlockRequestResult{}, fmt.Errorf("interested repository unavailable")
	}
	if s.unlockRate == nil {
		return UnlockRequestResult{}, fmt.Errorf("unlock rate limiter unavailable")
	}
	if retryAfter := s.unlockRate.allow(time.Now().UTC(), email, placeID); retryAfter > 0 {
		return UnlockRequestResult{}, &unlockRateLimitError{retryAfter: retryAfter}
	}
	// The PostgreSQL timestamp is the cross-process source of truth. This check
	// deliberately occurs before GetReport so a resend cannot repeatedly trigger
	// Google Places, screenshot, competitor, or AI work.
	previous, previousErr := s.interested.GetByEmailPlace(ctx, email, placeID)
	if previousErr == nil && time.Now().UTC().Before(previous.OTPRequestedAt.Add(otpResendCooldown)) {
		return UnlockRequestResult{}, &unlockRateLimitError{
			retryAfter: time.Until(previous.OTPRequestedAt.Add(otpResendCooldown)),
		}
	}
	if previousErr != nil && !errors.Is(previousErr, repository.ErrNotFound) {
		return UnlockRequestResult{}, previousErr
	}
	otpSecret, err := s.unlockOTPSecret()
	if err != nil {
		return UnlockRequestResult{}, err
	}

	restaurantName, err := s.unlockRestaurantName(ctx, placeID)
	if err != nil {
		return UnlockRequestResult{}, err
	}

	otp, err := generateOTP()
	if err != nil {
		return UnlockRequestResult{}, fmt.Errorf("generate otp: %w", err)
	}
	token, err := generateUnlockToken()
	if err != nil {
		return UnlockRequestResult{}, fmt.Errorf("generate unlock token: %w", err)
	}
	expires := time.Now().UTC().Add(otpTTL)
	otpHash, err := hashOTP(otpSecret, email, placeID, otp)
	if err != nil {
		return UnlockRequestResult{}, fmt.Errorf("hash otp: %w", err)
	}

	_, err = s.interested.UpsertPending(
		ctx,
		email,
		placeID,
		restaurantName,
		contactName,
		contactPhone,
		otpHash,
		token,
		expires,
	)
	if err != nil {
		return UnlockRequestResult{}, fmt.Errorf("store unlock otp: %w", err)
	}

	subjectName := strings.Join(strings.Fields(restaurantName), " ")
	subject := fmt.Sprintf("Your Tuvi SEO report code for %s", subjectName)
	htmlBody := buildUnlockEmailHTML(restaurantName, otp)
	textBody := fmt.Sprintf(
		"Your Tuvi verification code is %s (expires in %d minutes).\n\nReturn to the Tuvi report and enter this code.\n",
		otp, int(otpTTL.Minutes()),
	)

	result := UnlockRequestResult{
		Status:       "ok",
		Message:      "Verification code sent. Check your inbox.",
		Email:        email,
		PlaceID:      placeID,
		ExpiresInSec: int(otpTTL.Seconds()),
	}

	if s.mailer == nil {
		if s.isLocalEnv() {
			result.DevOTP = otp
			result.Message = "Email sending disabled — use the on-screen code (local only)."
			return result, nil
		}
		return UnlockRequestResult{}, ErrEmailSendFailed
	}

	sendResult, sendErr := s.mailer.Send(ctx, emailprovider.SendRequest{
		To:       email,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
		Metadata: map[string]string{
			"kind":     "seo_unlock_otp",
			"place_id": placeID,
		},
	})
	if sendErr != nil {
		s.log.ErrorContext(ctx, "seo_unlock_email_failed", "error", sendErr)
		if s.isLocalEnv() {
			result.DevOTP = otp
			result.Message = "Email send failed — use the on-screen code (local only)."
			return result, nil
		}
		return UnlockRequestResult{}, ErrEmailSendFailed
	}
	if sendResult.Skipped || strings.TrimSpace(sendResult.RedirectedTo) != "" {
		if s.isLocalEnv() {
			result.DevOTP = otp
			result.Message = "Email sending disabled — use the on-screen code (local only)."
			return result, nil
		}
		s.log.ErrorContext(ctx, "seo_unlock_email_not_delivered", "skipped", sendResult.Skipped, "redirected", sendResult.RedirectedTo != "")
		return UnlockRequestResult{}, ErrEmailSendFailed
	}

	return result, nil
}

func (s *Service) unlockRestaurantName(ctx context.Context, placeID string) (string, error) {
	fetchPlaceDetails := s.fetchPlaceDetails
	if fetchPlaceDetails == nil && s.places != nil {
		fetchPlaceDetails = s.places.GetPlaceDetails
	}
	if fetchPlaceDetails == nil {
		return "", ErrNotFound
	}
	snapshot, err := fetchPlaceDetails(ctx, placeID)
	if err != nil {
		return "", err
	}
	if snapshot == nil {
		return "", ErrNotFound
	}
	name := strings.Join(strings.Fields(snapshot.Details.Name), " ")
	if name == "" {
		name = "your restaurant"
	}
	return name, nil
}

// VerifyUnlockOTP validates the OTP and returns the unlocked report. Requester
// identity stays exclusively in seo_interested: verification never creates or
// updates a canonical restaurant, profile, scrape lifecycle, owner, or interest.
func (s *Service) VerifyUnlockOTP(ctx context.Context, emailRaw, placeID, otp string) (UnlockVerifyResult, error) {
	email, err := normalizeEmail(emailRaw)
	if err != nil {
		return UnlockVerifyResult{}, err
	}
	placeID = sanitizePlaceID(placeID)
	otp = strings.TrimSpace(otp)
	if placeID == "" || !isSixDigitOTP(otp) {
		return UnlockVerifyResult{}, ErrInvalidOTP
	}
	if s.interested == nil {
		return UnlockVerifyResult{}, fmt.Errorf("interested repository unavailable")
	}
	otpSecret, err := s.unlockOTPSecret()
	if err != nil {
		return UnlockVerifyResult{}, err
	}

	rec, err := s.interested.GetByEmailPlace(ctx, email, placeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return UnlockVerifyResult{}, ErrInvalidOTP
		}
		return UnlockVerifyResult{}, err
	}
	if rec.OTPHash == "" || rec.OTPExpiresAt == nil || rec.OTPAttempts >= maxOTPAttempts || time.Now().UTC().After(rec.OTPExpiresAt.UTC()) {
		return UnlockVerifyResult{}, ErrInvalidOTP
	}
	if !otpHashMatches(otpSecret, email, placeID, otp, rec.OTPHash) {
		if _, recordErr := s.interested.RecordFailedOTP(ctx, rec.ID); recordErr != nil && !errors.Is(recordErr, ErrInvalidOTP) {
			s.log.WarnContext(ctx, "seo_unlock_failed_attempt_record_failed", "error", recordErr, "place_id", placeID)
		}
		return UnlockVerifyResult{}, ErrInvalidOTP
	}

	reportPayload, err := s.GetReport(ctx, placeID)
	if err != nil {
		return UnlockVerifyResult{}, err
	}

	rec, err = s.interested.MarkVerified(ctx, rec.ID, time.Now().UTC().Add(unlockTTL))
	if err != nil {
		return UnlockVerifyResult{}, err
	}

	reportPayload.Report = unlockReport(reportPayload.Report)
	return UnlockVerifyResult{
		Status:      "ok",
		UnlockToken: rec.UnlockToken,
		Interested:  rec.Interested,
		Place:       reportPayload.Place,
		Report:      reportPayload.Report,
	}, nil
}

// GetReportUnlocked returns the report unlocked when a valid unlock token is provided.
func (s *Service) GetReportUnlocked(ctx context.Context, placeID, unlockToken string) (ReportResponse, error) {
	payload, err := s.GetReport(ctx, placeID)
	if err != nil {
		return ReportResponse{}, err
	}
	if strings.TrimSpace(unlockToken) == "" || s.interested == nil {
		return payload, nil
	}
	rec, err := s.interested.GetByUnlockToken(ctx, strings.TrimSpace(unlockToken))
	if err != nil {
		return payload, nil
	}
	if sanitizePlaceID(rec.PlaceID) != sanitizePlaceID(placeID) {
		return payload, nil
	}
	if rec.VerifiedAt != nil && !rec.UnlockExpiresAt.IsZero() && time.Now().UTC().Before(rec.UnlockExpiresAt.UTC()) {
		payload.Report = unlockReport(payload.Report)
	}
	return payload, nil
}

func (s *Service) isLocalEnv() bool {
	env := strings.ToLower(strings.TrimSpace(s.appEnv))
	return env == "" || env == config.EnvLocal || env == config.EnvTest
}

func (s *Service) unlockOTPSecret() (string, error) {
	value, ok := serviceOTPSecrets.Load(s)
	secret, isString := value.(string)
	if !ok || !isString || strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("unlock otp secret unavailable")
	}
	return secret, nil
}

func unlockReport(report Report) Report {
	report.FullReportLocked = false
	report.UnlockCTA = ""
	if len(report.Issues) == 0 {
		report.Issues = append(report.Issues, Issue{
			Title:       "Keep listing freshness high",
			Description: "Post weekly photos and reply to new reviews within 24 hours to hold Map Pack rank.",
		})
	}
	return report
}

func buildUnlockEmailHTML(restaurantName, otp string) string {
	name := html.EscapeString(restaurantName)
	code := html.EscapeString(otp)
	return fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:Arial,sans-serif;line-height:1.5;color:#111;background:#f7f5f2;padding:24px;">
  <div style="max-width:520px;margin:0 auto;background:#fff;border-radius:16px;padding:28px;">
    <p style="margin:0 0 8px;font-size:13px;letter-spacing:.08em;text-transform:uppercase;color:#888;">Tuvi SEO Report</p>
    <h1 style="margin:0 0 16px;font-size:22px;">Verify your email</h1>
    <p style="margin:0 0 12px;">Use this code to unlock the full report for <strong>%s</strong>:</p>
    <p style="margin:0 0 20px;font-size:32px;font-weight:700;letter-spacing:8px;">%s</p>
    <p style="margin:0 0 20px;color:#555;">Code expires in %d minutes.</p>
    <p style="margin:0 0 12px;">Return to the Tuvi report and enter this code.</p>
    <p style="margin:24px 0 0;font-size:12px;color:#999;">If you did not request this, you can ignore this email.</p>
  </div>
</body></html>`, name, code, int(otpTTL.Minutes()))
}
