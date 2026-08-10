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

	"github.com/google/uuid"

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
	appURLs config.AppURLsConfig,
	profilesRepo profiles.SiteRepository,
	interestedRepo InterestedRepository,
	leadUpserter LeadUpserter,
	mailer EmailSender,
	llmClient llmlib.Client,
	tokenSecret string,
	log *slog.Logger,
) *Service {
	svc := NewService(placesCfg, profilesRepo, llmClient, log)
	svc.appEnv = appCfg.Env
	svc.publicBaseURL = strings.TrimRight(strings.TrimSpace(appURLs.PublicBaseURL), "/")
	svc.publicWebURL = strings.TrimRight(strings.TrimSpace(appURLs.PublicWebURL), "/")
	if svc.publicWebURL == "" {
		svc.publicWebURL = "http://localhost:3000"
	}
	if svc.publicBaseURL == "" {
		svc.publicBaseURL = "http://localhost:8080"
	}
	svc.interested = interestedRepo
	svc.leads = leadUpserter
	svc.mailer = mailer
	serviceOTPSecrets.Store(svc, strings.TrimSpace(tokenSecret))
	return svc
}

// RequestUnlockEmail stores a bounded OTP and emails it with a View full report link.
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
	otpSecret, err := s.unlockOTPSecret()
	if err != nil {
		return UnlockRequestResult{}, err
	}

	reportPayload, err := s.GetReport(ctx, placeID)
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
		reportPayload.Place.Name,
		contactName,
		contactPhone,
		otpHash,
		token,
		expires,
	)
	if err != nil {
		return UnlockRequestResult{}, fmt.Errorf("store unlock otp: %w", err)
	}

	clickURL := fmt.Sprintf("%s/api/public/v1/seo/unlock/click/%s", s.publicBaseURL, token)
	subject := fmt.Sprintf("Your Tuvi SEO report code for %s", reportPayload.Place.Name)
	htmlBody := buildUnlockEmailHTML(reportPayload.Place.Name, otp, clickURL)
	textBody := fmt.Sprintf(
		"Your Tuvi verification code is %s (expires in %d minutes).\n\nView full report: %s\n",
		otp, int(otpTTL.Minutes()), clickURL,
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
	if sendResult.Skipped && s.isLocalEnv() {
		result.DevOTP = otp
		result.Message = "Email sending disabled — use the on-screen code (local only)."
	}

	return result, nil
}

// VerifyUnlockOTP validates the OTP, records the verified lead without inferring
// marketing interest, and returns the unlocked report.
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

	leadID, err := s.upsertLead(ctx, reportPayload.Place, email)
	if err != nil {
		s.log.ErrorContext(ctx, "seo_unlock_lead_upsert_failed", "error", err, "place_id", placeID)
		return UnlockVerifyResult{}, fmt.Errorf("scrape lead failed: %w", err)
	}

	rec, err = s.interested.MarkVerified(ctx, rec.ID, leadID, time.Now().UTC().Add(unlockTTL))
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

// ConfirmUnlockClick treats possession of the emailed, unexpired token as email
// verification. It intentionally does not create a lead or mark marketing
// interest; only an explicit OTP submission can create/update the lead record.
func (s *Service) ConfirmUnlockClick(ctx context.Context, token string) (InterestedRecord, PlaceDetails, error) {
	if s.interested == nil {
		return InterestedRecord{}, PlaceDetails{}, fmt.Errorf("interested repository unavailable")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return InterestedRecord{}, PlaceDetails{}, ErrInvalidUnlock
	}
	rec, err := s.interested.GetByUnlockToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return InterestedRecord{}, PlaceDetails{}, ErrInvalidUnlock
		}
		return InterestedRecord{}, PlaceDetails{}, err
	}
	if rec.UnlockExpiresAt.IsZero() || time.Now().UTC().After(rec.UnlockExpiresAt.UTC()) {
		return InterestedRecord{}, PlaceDetails{}, ErrInvalidUnlock
	}

	reportPayload, err := s.GetReport(ctx, rec.PlaceID)
	if err != nil {
		return InterestedRecord{}, PlaceDetails{}, err
	}

	if rec.VerifiedAt == nil {
		rec, err = s.interested.MarkEmailVerified(ctx, rec.ID, time.Now().UTC().Add(unlockTTL))
		if err != nil {
			return InterestedRecord{}, PlaceDetails{}, err
		}
	}

	return rec, reportPayload.Place, nil
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

func (s *Service) upsertLead(ctx context.Context, place PlaceDetails, email string) (uuid.UUID, error) {
	if s.leads == nil {
		return uuid.Nil, fmt.Errorf("lead upserter unavailable")
	}
	return s.leads.UpsertLeadFromPlace(ctx, place, email)
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

func buildUnlockEmailHTML(restaurantName, otp, clickURL string) string {
	name := html.EscapeString(restaurantName)
	code := html.EscapeString(otp)
	url := html.EscapeString(clickURL)
	return fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:Arial,sans-serif;line-height:1.5;color:#111;background:#f7f5f2;padding:24px;">
  <div style="max-width:520px;margin:0 auto;background:#fff;border-radius:16px;padding:28px;">
    <p style="margin:0 0 8px;font-size:13px;letter-spacing:.08em;text-transform:uppercase;color:#888;">Tuvi SEO Report</p>
    <h1 style="margin:0 0 16px;font-size:22px;">Verify your email</h1>
    <p style="margin:0 0 12px;">Use this code to unlock the full report for <strong>%s</strong>:</p>
    <p style="margin:0 0 20px;font-size:32px;font-weight:700;letter-spacing:8px;">%s</p>
    <p style="margin:0 0 20px;color:#555;">Code expires in %d minutes.</p>
    <p style="margin:0 0 12px;">Or open the full report directly:</p>
    <p style="margin:0 0 8px;">
      <a href="%s" style="display:inline-block;background:#e86a2d;color:#fff;text-decoration:none;padding:12px 20px;border-radius:999px;font-weight:600;">
        View full report
      </a>
    </p>
    <p style="margin:24px 0 0;font-size:12px;color:#999;">If you did not request this, you can ignore this email.</p>
  </div>
</body></html>`, name, code, int(otpTTL.Minutes()), url)
}
