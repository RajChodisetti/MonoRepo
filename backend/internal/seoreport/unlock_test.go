package seoreport

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

const unlockTestSecret = "unlock-test-secret-at-least-32-characters"

type fakeInterestedRepository struct {
	mu     sync.Mutex
	record InterestedRecord
}

func (repo *fakeInterestedRepository) UpsertPending(
	_ context.Context,
	email, placeID, restaurantName, contactName, contactPhone, otpHash, unlockToken string,
	otpExpires time.Time,
) (InterestedRecord, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	now := time.Now().UTC()
	if repo.record.ID != uuid.Nil && now.Sub(repo.record.OTPRequestedAt) < otpResendCooldown {
		return InterestedRecord{}, ErrUnlockRateLimit
	}
	if repo.record.ID == uuid.Nil {
		repo.record.ID = uuid.New()
	}
	repo.record.Email = email
	repo.record.PlaceID = placeID
	repo.record.RestaurantName = restaurantName
	repo.record.ContactName = contactName
	repo.record.ContactPhone = contactPhone
	repo.record.OTPHash = otpHash
	repo.record.OTPExpiresAt = &otpExpires
	repo.record.OTPRequestedAt = now
	repo.record.OTPAttempts = 0
	repo.record.VerifiedAt = nil
	repo.record.UnlockToken = unlockToken
	repo.record.UnlockExpiresAt = otpExpires
	return repo.record, nil
}

func (repo *fakeInterestedRepository) GetByEmailPlace(_ context.Context, email, placeID string) (InterestedRecord, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.record.ID == uuid.Nil || repo.record.Email != email || repo.record.PlaceID != placeID {
		return InterestedRecord{}, repository.ErrNotFound
	}
	return repo.record, nil
}

func (repo *fakeInterestedRepository) GetByUnlockToken(_ context.Context, token string) (InterestedRecord, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.record.ID == uuid.Nil || repo.record.UnlockToken != token {
		return InterestedRecord{}, repository.ErrNotFound
	}
	return repo.record, nil
}

func (repo *fakeInterestedRepository) RecordFailedOTP(_ context.Context, id uuid.UUID) (InterestedRecord, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.record.ID != id || repo.record.VerifiedAt != nil || repo.record.OTPAttempts >= maxOTPAttempts {
		return InterestedRecord{}, ErrInvalidOTP
	}
	repo.record.OTPAttempts++
	if repo.record.OTPAttempts >= maxOTPAttempts {
		repo.record.OTPHash = ""
		repo.record.OTPExpiresAt = nil
	}
	return repo.record, nil
}

func (repo *fakeInterestedRepository) MarkVerified(
	_ context.Context,
	id uuid.UUID,
	leadRestaurantID uuid.UUID,
	unlockExpires time.Time,
) (InterestedRecord, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.record.ID != id || repo.record.VerifiedAt != nil || repo.record.OTPHash == "" ||
		repo.record.OTPExpiresAt == nil || time.Now().UTC().After(*repo.record.OTPExpiresAt) ||
		repo.record.OTPAttempts >= maxOTPAttempts {
		return InterestedRecord{}, ErrInvalidOTP
	}
	now := time.Now().UTC()
	repo.record.VerifiedAt = &now
	repo.record.LeadRestaurantID = &leadRestaurantID
	repo.record.OTPHash = ""
	repo.record.OTPExpiresAt = nil
	repo.record.UnlockExpiresAt = unlockExpires
	return repo.record, nil
}

func (repo *fakeInterestedRepository) MarkEmailVerified(
	_ context.Context,
	id uuid.UUID,
	unlockExpires time.Time,
) (InterestedRecord, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.record.ID != id || time.Now().UTC().After(repo.record.UnlockExpiresAt) {
		return InterestedRecord{}, ErrInvalidUnlock
	}
	if repo.record.VerifiedAt == nil {
		now := time.Now().UTC()
		repo.record.VerifiedAt = &now
		repo.record.UnlockExpiresAt = unlockExpires
	}
	repo.record.OTPHash = ""
	repo.record.OTPExpiresAt = nil
	return repo.record, nil
}

func (repo *fakeInterestedRepository) snapshot() InterestedRecord {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	return repo.record
}

func (repo *fakeInterestedRepository) update(mutator func(*InterestedRecord)) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	mutator(&repo.record)
}

type fakeUnlockLeadUpserter struct {
	calls     int
	email     string
	lastPlace PlaceDetails
	id        uuid.UUID
}

func (lead *fakeUnlockLeadUpserter) UpsertLeadFromPlace(_ context.Context, place PlaceDetails, email string) (uuid.UUID, error) {
	lead.calls++
	lead.email = email
	lead.lastPlace = place
	if lead.id == uuid.Nil {
		lead.id = uuid.New()
	}
	return lead.id, nil
}

type fakeUnlockMailer struct {
	requests []emailprovider.SendRequest
	err      error
}

func (mailer *fakeUnlockMailer) Send(_ context.Context, request emailprovider.SendRequest) (emailprovider.SendResult, error) {
	mailer.requests = append(mailer.requests, request)
	if mailer.err != nil {
		return emailprovider.SendResult{}, mailer.err
	}
	return emailprovider.SendResult{ProviderMessageID: "message-1"}, nil
}

func newUnlockTestService(
	t *testing.T,
	repo *fakeInterestedRepository,
	leads *fakeUnlockLeadUpserter,
	mailer *fakeUnlockMailer,
) *Service {
	t.Helper()
	service := newReportTestService()
	service.appEnv = "test"
	service.publicBaseURL = "https://api.example.test"
	service.publicWebURL = "https://www.example.test"
	service.interested = repo
	service.leads = leads
	service.mailer = mailer
	service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
		return &placeSnapshot{Details: PlaceDetails{
			PlaceID: "place-1",
			Name:    "Security Test Thai",
			Address: "1 Test Street",
			Phone:   "+61 2 9000 0000",
			Source:  "places",
		}}, nil
	}
	service.fetchSiteContent = func(context.Context, string) (profiles.SiteContent, bool) {
		return profiles.SiteContent{}, false
	}
	serviceOTPSecrets.Store(service, unlockTestSecret)
	t.Cleanup(func() { serviceOTPSecrets.Delete(service) })
	return service
}

func otpFromMailer(t *testing.T, mailer *fakeUnlockMailer) string {
	t.Helper()
	if len(mailer.requests) == 0 {
		t.Fatal("verification email was not sent")
	}
	match := regexp.MustCompile(`verification code is ([0-9]{6})`).FindStringSubmatch(mailer.requests[len(mailer.requests)-1].TextBody)
	if len(match) != 2 {
		t.Fatalf("verification email did not contain one six-digit code: %q", mailer.requests[len(mailer.requests)-1].TextBody)
	}
	return match[1]
}

func requestUnlockForTest(t *testing.T, service *Service) (UnlockRequestResult, string) {
	t.Helper()
	result, err := service.RequestUnlockEmail(
		context.Background(),
		"  Sam   Owner  ",
		"OWNER@EXAMPLE.COM",
		"+61 (0)2 9123 4567",
		"place-1",
	)
	if err != nil {
		t.Fatalf("RequestUnlockEmail() error = %v", err)
	}
	mailer := service.mailer.(*fakeUnlockMailer)
	return result, otpFromMailer(t, mailer)
}

func TestRequestUnlockValidatesMandatoryContactFields(t *testing.T) {
	tests := []struct {
		name      string
		contact   string
		email     string
		phone     string
		placeID   string
		wantError error
	}{
		{name: "missing name", contact: "", email: "owner@example.com", phone: "+61 2 9123 4567", placeID: "place-1", wantError: ErrInvalidName},
		{name: "short name", contact: "A", email: "owner@example.com", phone: "+61 2 9123 4567", placeID: "place-1", wantError: ErrInvalidName},
		{name: "long name", contact: strings.Repeat("a", 101), email: "owner@example.com", phone: "+61 2 9123 4567", placeID: "place-1", wantError: ErrInvalidName},
		{name: "invalid email", contact: "Sam Owner", email: "not-an-email", phone: "+61 2 9123 4567", placeID: "place-1", wantError: ErrInvalidEmail},
		{name: "short phone", contact: "Sam Owner", email: "owner@example.com", phone: "123456", placeID: "place-1", wantError: ErrInvalidPhone},
		{name: "long phone", contact: "Sam Owner", email: "owner@example.com", phone: strings.Repeat("1", 21), placeID: "place-1", wantError: ErrInvalidPhone},
		{name: "phone letters", contact: "Sam Owner", email: "owner@example.com", phone: "call-me-1234567", placeID: "place-1", wantError: ErrInvalidPhone},
		{name: "phone control", contact: "Sam Owner", email: "owner@example.com", phone: "1234\n5678", placeID: "place-1", wantError: ErrInvalidPhone},
		{name: "missing place", contact: "Sam Owner", email: "owner@example.com", phone: "+61 2 9123 4567", placeID: "", wantError: ErrNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeInterestedRepository{}
			mailer := &fakeUnlockMailer{}
			service := newUnlockTestService(t, repo, &fakeUnlockLeadUpserter{}, mailer)
			_, err := service.RequestUnlockEmail(context.Background(), test.contact, test.email, test.phone, test.placeID)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("RequestUnlockEmail() error = %v, want %v", err, test.wantError)
			}
			if len(mailer.requests) != 0 {
				t.Fatalf("invalid request sent %d email(s)", len(mailer.requests))
			}
		})
	}
}

func TestRequestUnlockPersistsContactAndHMACOTP(t *testing.T) {
	repo := &fakeInterestedRepository{}
	mailer := &fakeUnlockMailer{}
	service := newUnlockTestService(t, repo, &fakeUnlockLeadUpserter{}, mailer)

	started := time.Now().UTC()
	result, otp := requestUnlockForTest(t, service)
	record := repo.snapshot()

	if result.Email != "owner@example.com" || result.PlaceID != "place-1" || result.ExpiresInSec != int(otpTTL.Seconds()) {
		t.Fatalf("unexpected request result: %+v", result)
	}
	if record.ContactName != "Sam Owner" || record.ContactPhone != "+61 (0)2 9123 4567" {
		t.Fatalf("stored contact = %q / %q", record.ContactName, record.ContactPhone)
	}
	if record.VerifiedAt != nil || record.Interested {
		t.Fatalf("pending request mutated verification/interest: %+v", record)
	}
	if !otpHashMatches(unlockTestSecret, record.Email, record.PlaceID, otp, record.OTPHash) {
		t.Fatal("stored OTP is not the expected context-bound HMAC")
	}
	if otpHashMatches("different-secret", record.Email, record.PlaceID, otp, record.OTPHash) ||
		otpHashMatches(unlockTestSecret, "other@example.com", record.PlaceID, otp, record.OTPHash) ||
		otpHashMatches(unlockTestSecret, record.Email, "other-place", otp, record.OTPHash) {
		t.Fatal("OTP HMAC was not bound to secret, email, and place")
	}
	if record.UnlockExpiresAt.Before(started.Add(otpTTL-time.Second)) || record.UnlockExpiresAt.After(time.Now().UTC().Add(otpTTL+time.Second)) {
		t.Fatalf("pending unlock expiry = %v, want about %v", record.UnlockExpiresAt, otpTTL)
	}
	if got := mailer.requests[0].To; got != "owner@example.com" {
		t.Fatalf("email recipient = %q", got)
	}
}

func TestRequestUnlockCooldownAndPendingReset(t *testing.T) {
	repo := &fakeInterestedRepository{}
	mailer := &fakeUnlockMailer{}
	service := newUnlockTestService(t, repo, &fakeUnlockLeadUpserter{}, mailer)
	requestUnlockForTest(t, service)

	_, err := service.RequestUnlockEmail(context.Background(), "Sam Owner", "owner@example.com", "+61 2 9123 4567", "place-1")
	if !errors.Is(err, ErrUnlockRateLimit) {
		t.Fatalf("immediate resend error = %v, want rate limit", err)
	}
	if len(mailer.requests) != 1 {
		t.Fatalf("cooldown sent %d emails, want 1", len(mailer.requests))
	}

	verifiedAt := time.Now().UTC().Add(-time.Hour)
	repo.update(func(record *InterestedRecord) {
		record.VerifiedAt = &verifiedAt
		record.OTPRequestedAt = time.Now().UTC().Add(-otpResendCooldown - time.Second)
	})
	requestUnlockForTest(t, service)
	if repo.snapshot().VerifiedAt != nil {
		t.Fatal("new pending request did not reset verified_at")
	}
}

func TestVerifyUnlockRequiresExactCodeAndCapsFailures(t *testing.T) {
	repo := &fakeInterestedRepository{}
	mailer := &fakeUnlockMailer{}
	leads := &fakeUnlockLeadUpserter{}
	service := newUnlockTestService(t, repo, leads, mailer)
	_, otp := requestUnlockForTest(t, service)

	for _, malformed := range []string{"1234", "12345", "1234567", "12345x"} {
		_, err := service.VerifyUnlockOTP(context.Background(), "owner@example.com", "place-1", malformed)
		if !errors.Is(err, ErrInvalidOTP) {
			t.Fatalf("VerifyUnlockOTP(%q) error = %v", malformed, err)
		}
	}
	if repo.snapshot().OTPAttempts != 0 {
		t.Fatal("malformed codes must not consume stored comparison attempts")
	}

	wrong := "000000"
	if otp == wrong {
		wrong = "999999"
	}
	for attempt := 1; attempt <= maxOTPAttempts; attempt++ {
		_, err := service.VerifyUnlockOTP(context.Background(), "owner@example.com", "place-1", wrong)
		if !errors.Is(err, ErrInvalidOTP) {
			t.Fatalf("wrong attempt %d error = %v", attempt, err)
		}
		if got := repo.snapshot().OTPAttempts; got != attempt {
			t.Fatalf("wrong attempt count = %d, want %d", got, attempt)
		}
	}
	if repo.snapshot().OTPHash != "" || repo.snapshot().OTPExpiresAt != nil {
		t.Fatal("fifth failure did not invalidate the OTP")
	}
	if _, err := service.VerifyUnlockOTP(context.Background(), "owner@example.com", "place-1", otp); !errors.Is(err, ErrInvalidOTP) {
		t.Fatalf("correct code after lockout error = %v", err)
	}
	if leads.calls != 0 {
		t.Fatalf("failed verification created %d lead(s)", leads.calls)
	}
}

func TestVerifyUnlockCreatesLeadWithoutMarketingInterest(t *testing.T) {
	repo := &fakeInterestedRepository{}
	mailer := &fakeUnlockMailer{}
	leads := &fakeUnlockLeadUpserter{}
	service := newUnlockTestService(t, repo, leads, mailer)
	_, otp := requestUnlockForTest(t, service)

	verifiedAt := time.Now().UTC()
	result, err := service.VerifyUnlockOTP(context.Background(), "OWNER@example.com", "place-1", otp)
	if err != nil {
		t.Fatalf("VerifyUnlockOTP() error = %v", err)
	}
	record := repo.snapshot()
	if leads.calls != 1 || leads.email != "owner@example.com" {
		t.Fatalf("lead calls/email = %d/%q", leads.calls, leads.email)
	}
	if leads.lastPlace.Phone != "+61 2 9000 0000" {
		t.Fatalf("requester phone overwrote venue phone: %q", leads.lastPlace.Phone)
	}
	if record.ContactPhone != "+61 (0)2 9123 4567" {
		t.Fatalf("requester phone was not retained separately: %q", record.ContactPhone)
	}
	if record.VerifiedAt == nil || record.LeadRestaurantID == nil {
		t.Fatalf("verification was not persisted: %+v", record)
	}
	if record.Interested || result.Interested {
		t.Fatal("email verification must not opt the contact into marketing interest")
	}
	if record.UnlockExpiresAt.Before(verifiedAt.Add(unlockTTL-time.Second)) || record.UnlockExpiresAt.After(time.Now().UTC().Add(unlockTTL+time.Second)) {
		t.Fatalf("verified unlock expiry = %v, want about %v", record.UnlockExpiresAt, unlockTTL)
	}
	if result.Report.FullReportLocked {
		t.Fatal("verified OTP did not unlock returned report")
	}

	payload, err := service.GetReportUnlocked(context.Background(), "place-1", record.UnlockToken)
	if err != nil || payload.Report.FullReportLocked {
		t.Fatalf("GetReportUnlocked(valid) locked=%v err=%v", payload.Report.FullReportLocked, err)
	}
	payload, err = service.GetReportUnlocked(context.Background(), "other-place", record.UnlockToken)
	if err != nil || !payload.Report.FullReportLocked {
		t.Fatalf("GetReportUnlocked(other place) locked=%v err=%v", payload.Report.FullReportLocked, err)
	}
	repo.update(func(record *InterestedRecord) { record.UnlockExpiresAt = time.Now().UTC().Add(-time.Second) })
	payload, err = service.GetReportUnlocked(context.Background(), "place-1", record.UnlockToken)
	if err != nil || !payload.Report.FullReportLocked {
		t.Fatalf("GetReportUnlocked(expired) locked=%v err=%v", payload.Report.FullReportLocked, err)
	}
}

func TestVerifyUnlockRejectsExpiredOTPWithoutLead(t *testing.T) {
	repo := &fakeInterestedRepository{}
	mailer := &fakeUnlockMailer{}
	leads := &fakeUnlockLeadUpserter{}
	service := newUnlockTestService(t, repo, leads, mailer)
	_, otp := requestUnlockForTest(t, service)
	past := time.Now().UTC().Add(-time.Second)
	repo.update(func(record *InterestedRecord) { record.OTPExpiresAt = &past })

	_, err := service.VerifyUnlockOTP(context.Background(), "owner@example.com", "place-1", otp)
	if !errors.Is(err, ErrInvalidOTP) || leads.calls != 0 {
		t.Fatalf("expired OTP error/leads = %v/%d", err, leads.calls)
	}
}

func TestMagicLinkVerifiesEmailWithoutLeadOrInterest(t *testing.T) {
	repo := &fakeInterestedRepository{}
	mailer := &fakeUnlockMailer{}
	leads := &fakeUnlockLeadUpserter{}
	service := newUnlockTestService(t, repo, leads, mailer)
	request, _ := requestUnlockForTest(t, service)
	recordBefore := repo.snapshot()

	record, place, err := service.ConfirmUnlockClick(context.Background(), recordBefore.UnlockToken)
	if err != nil {
		t.Fatalf("ConfirmUnlockClick() error = %v", err)
	}
	if request.PlaceID != place.PlaceID || record.VerifiedAt == nil {
		t.Fatalf("magic-link result = place=%q record=%+v", place.PlaceID, record)
	}
	if leads.calls != 0 || record.LeadRestaurantID != nil || record.Interested {
		t.Fatalf("magic link mutated lead/interest: leads=%d record=%+v", leads.calls, record)
	}
	if record.UnlockExpiresAt.Before(time.Now().UTC().Add(unlockTTL - time.Second)) {
		t.Fatalf("magic link did not exchange to seven-day unlock: %v", record.UnlockExpiresAt)
	}

	payload, err := service.GetReportUnlocked(context.Background(), "place-1", record.UnlockToken)
	if err != nil || payload.Report.FullReportLocked {
		t.Fatalf("magic-link token did not unlock report: locked=%v err=%v", payload.Report.FullReportLocked, err)
	}
}

func TestMagicLinkRejectsExpiredTokenReadOnly(t *testing.T) {
	repo := &fakeInterestedRepository{}
	mailer := &fakeUnlockMailer{}
	leads := &fakeUnlockLeadUpserter{}
	service := newUnlockTestService(t, repo, leads, mailer)
	requestUnlockForTest(t, service)
	repo.update(func(record *InterestedRecord) { record.UnlockExpiresAt = time.Now().UTC().Add(-time.Second) })
	token := repo.snapshot().UnlockToken

	_, _, err := service.ConfirmUnlockClick(context.Background(), token)
	if !errors.Is(err, ErrInvalidUnlock) {
		t.Fatalf("expired magic-link error = %v", err)
	}
	if leads.calls != 0 || repo.snapshot().VerifiedAt != nil || repo.snapshot().Interested {
		t.Fatal("expired magic link mutated verification, lead, or interest")
	}
}

func TestUnlockRequiresConfiguredOTPSecret(t *testing.T) {
	repo := &fakeInterestedRepository{}
	mailer := &fakeUnlockMailer{}
	service := newUnlockTestService(t, repo, &fakeUnlockLeadUpserter{}, mailer)
	serviceOTPSecrets.Delete(service)

	_, err := service.RequestUnlockEmail(context.Background(), "Sam Owner", "owner@example.com", "+61 2 9123 4567", "place-1")
	if err == nil || !strings.Contains(err.Error(), "secret") {
		t.Fatalf("missing-secret error = %v", err)
	}
	if len(mailer.requests) != 0 {
		t.Fatal("missing secret sent an email")
	}
}

func TestProductionEmailFailureDoesNotReturnDevOTP(t *testing.T) {
	repo := &fakeInterestedRepository{}
	mailer := &fakeUnlockMailer{err: fmt.Errorf("provider unavailable")}
	service := newUnlockTestService(t, repo, &fakeUnlockLeadUpserter{}, mailer)
	service.appEnv = "production"

	result, err := service.RequestUnlockEmail(context.Background(), "Sam Owner", "owner@example.com", "+61 2 9123 4567", "place-1")
	if !errors.Is(err, ErrEmailSendFailed) {
		t.Fatalf("production email failure = %v", err)
	}
	if result.DevOTP != "" {
		t.Fatalf("production response exposed dev OTP %q", result.DevOTP)
	}
}
