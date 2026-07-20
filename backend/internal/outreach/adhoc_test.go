package outreach_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type suppressibleMockRepo struct {
	mockRepo
	suppressed  map[string]bool
	recorded    []uuid.UUID
	recordError error
}

func (repo *suppressibleMockRepo) IsEmailSuppressed(ctx context.Context, email string) (bool, error) {
	return repo.suppressed[email], nil
}

func (repo *suppressibleMockRepo) RecordAdHocEmailSent(ctx context.Context, restaurantID uuid.UUID, recipientEmail string) error {
	if repo.recordError != nil {
		return repo.recordError
	}
	repo.recorded = append(repo.recorded, restaurantID)
	return nil
}

type recordingEmailProvider struct {
	sent   []emailprovider.SendRequest
	result emailprovider.SendResult
	err    error
}

func (provider *recordingEmailProvider) Send(ctx context.Context, req emailprovider.SendRequest) (emailprovider.SendResult, error) {
	if provider.err != nil {
		return emailprovider.SendResult{}, provider.err
	}
	provider.sent = append(provider.sent, req)
	if provider.result.ProviderMessageID != "" || provider.result.Skipped || provider.result.RedirectedTo != "" {
		return provider.result, nil
	}
	return emailprovider.SendResult{ProviderMessageID: "mock-adhoc"}, nil
}

func newAdHocTestService(t *testing.T, repo *suppressibleMockRepo, restaurantID uuid.UUID, email string, campaign *campaigns.Campaign, emailProvider emailprovider.Provider, disableSending bool, emailPools ...*emailprovider.AccountPool) *outreach.Service {
	t.Helper()

	restaurantsMock := &restaurants.Mock{Restaurants: map[uuid.UUID]restaurants.Restaurant{
		restaurantID: {ID: restaurantID, Name: "Test Cafe", Email: email, Status: "lead"},
	}}
	accessService := restaurants.NewService(restaurantsMock, &restaurants.MembershipMock{})

	campaignRepo := &campaigns.Mock{Campaigns: map[uuid.UUID]campaigns.Campaign{}}
	if campaign != nil {
		campaignRepo.Campaigns[campaign.ID] = *campaign
	}
	campaignService := campaigns.NewService(campaignRepo, nil, accessService, nil, config.AppURLsConfig{
		PublicBaseURL: "https://api.example.com",
		PublicWebURL:  "https://example.com",
	})

	var emailPool *emailprovider.AccountPool
	if len(emailPools) > 0 {
		emailPool = emailPools[0]
	}

	return outreach.NewService(
		repo,
		nil,
		campaignRepo,
		campaignService,
		accessService,
		outreach.DemoTokenResolver{},
		emailPool,
		emailProvider,
		config.EmailConfig{Provider: "zoho", DisableSending: disableSending},
		config.OutreachConfig{BulkMax: 150},
		nil,
		nil,
	)
}

func adminPrincipal() auth.Principal {
	return auth.Principal{Role: auth.RoleInternalAdmin, UserID: uuid.New()}
}

func TestSendAdHocSendsWhenGenericSendingDisabled(t *testing.T) {
	restaurantID := uuid.New()
	repo := &suppressibleMockRepo{}
	provider := &recordingEmailProvider{}
	campaign := &campaigns.Campaign{ID: uuid.New(), RestaurantID: restaurantID, Subject: "Hi", BodyHTML: "<p>hi</p>", BodyText: "hi"}
	service := newAdHocTestService(t, repo, restaurantID, "owner@example.com", campaign, provider, true)

	result, err := service.SendAdHoc(context.Background(), adminPrincipal(), restaurantID)
	if err != nil {
		t.Fatalf("SendAdHoc() error = %v, want nil", err)
	}
	if !result.Sent {
		t.Fatalf("result.Sent = false, want true")
	}
	if len(provider.sent) != 1 {
		t.Fatalf("expected one send attempt, got %d", len(provider.sent))
	}
}

func TestSendAdHocPrefersOutreachPoolWhenConfigured(t *testing.T) {
	restaurantID := uuid.New()
	repo := &suppressibleMockRepo{}
	genericProvider := &recordingEmailProvider{}
	outreachProvider := &recordingEmailProvider{}
	pool, err := emailprovider.NewAccountPool([]emailprovider.Provider{outreachProvider}, 25, 25)
	if err != nil {
		t.Fatalf("NewAccountPool() error = %v", err)
	}
	campaign := &campaigns.Campaign{ID: uuid.New(), RestaurantID: restaurantID, Subject: "Hi", BodyHTML: "<p>hi</p>", BodyText: "hi"}
	service := newAdHocTestService(t, repo, restaurantID, "owner@example.com", campaign, genericProvider, true, pool)

	result, err := service.SendAdHoc(context.Background(), adminPrincipal(), restaurantID)
	if err != nil {
		t.Fatalf("SendAdHoc() error = %v, want nil", err)
	}
	if !result.Sent {
		t.Fatalf("result.Sent = false, want true")
	}
	if len(outreachProvider.sent) != 1 {
		t.Fatalf("outreachProvider.sent = %d, want 1", len(outreachProvider.sent))
	}
	if len(genericProvider.sent) != 0 {
		t.Fatalf("genericProvider.sent = %d, want 0", len(genericProvider.sent))
	}
}

func TestSendAdHocRejectsSuppressedEmail(t *testing.T) {
	restaurantID := uuid.New()
	repo := &suppressibleMockRepo{suppressed: map[string]bool{"owner@example.com": true}}
	provider := &recordingEmailProvider{}
	campaign := &campaigns.Campaign{ID: uuid.New(), RestaurantID: restaurantID, Subject: "Hi", BodyHTML: "<p>hi</p>", BodyText: "hi"}
	service := newAdHocTestService(t, repo, restaurantID, "owner@example.com", campaign, provider, false)

	_, err := service.SendAdHoc(context.Background(), adminPrincipal(), restaurantID)
	if !errors.Is(err, outreach.ErrEmailSuppressed) {
		t.Fatalf("SendAdHoc() error = %v, want ErrEmailSuppressed", err)
	}
	if len(provider.sent) != 0 {
		t.Fatalf("expected no send attempt for suppressed recipient, got %d", len(provider.sent))
	}
}

func TestSendAdHocRejectsWhenNoCampaignDraft(t *testing.T) {
	restaurantID := uuid.New()
	repo := &suppressibleMockRepo{}
	provider := &recordingEmailProvider{}
	service := newAdHocTestService(t, repo, restaurantID, "owner@example.com", nil, provider, false)

	_, err := service.SendAdHoc(context.Background(), adminPrincipal(), restaurantID)
	if !errors.Is(err, outreach.ErrNoCampaignDraft) {
		t.Fatalf("SendAdHoc() error = %v, want ErrNoCampaignDraft", err)
	}
}

func TestSendAdHocSendsAndRecords(t *testing.T) {
	restaurantID := uuid.New()
	repo := &suppressibleMockRepo{}
	provider := &recordingEmailProvider{}
	campaign := &campaigns.Campaign{
		ID: uuid.New(), RestaurantID: restaurantID,
		Subject: "Hello Test Cafe", BodyHTML: "<p>hello</p>", BodyText: "hello",
	}
	service := newAdHocTestService(t, repo, restaurantID, "owner@example.com", campaign, provider, false)

	result, err := service.SendAdHoc(context.Background(), adminPrincipal(), restaurantID)
	if err != nil {
		t.Fatalf("SendAdHoc() error = %v, want nil", err)
	}
	if !result.Sent {
		t.Fatalf("result.Sent = false, want true")
	}
	if len(provider.sent) != 1 || provider.sent[0].To != "owner@example.com" {
		t.Fatalf("provider.sent = %+v, want one send to owner@example.com", provider.sent)
	}
	if len(repo.recorded) != 1 || repo.recorded[0] != restaurantID {
		t.Fatalf("repo.recorded = %v, want [%v]", repo.recorded, restaurantID)
	}
}

func TestSendAdHocDoesNotRecordSkippedDelivery(t *testing.T) {
	restaurantID := uuid.New()
	repo := &suppressibleMockRepo{}
	provider := &recordingEmailProvider{result: emailprovider.SendResult{Skipped: true}}
	campaign := &campaigns.Campaign{
		ID: uuid.New(), RestaurantID: restaurantID,
		Subject: "Hello Test Cafe", BodyHTML: "<p>hello</p>", BodyText: "hello",
	}
	service := newAdHocTestService(t, repo, restaurantID, "owner@example.com", campaign, provider, false)

	_, err := service.SendAdHoc(context.Background(), adminPrincipal(), restaurantID)
	if !errors.Is(err, outreach.ErrDeliverySkipped) {
		t.Fatalf("SendAdHoc() error = %v, want ErrDeliverySkipped", err)
	}
	if len(repo.recorded) != 0 {
		t.Fatalf("repo.recorded = %v, want no contacted record", repo.recorded)
	}
}

func TestSendAdHocBatchCollectsPerLeadResults(t *testing.T) {
	sentID := uuid.New()
	blockedID := uuid.New()
	repo := &suppressibleMockRepo{suppressed: map[string]bool{"blocked@example.com": true}}
	provider := &recordingEmailProvider{}

	restaurantsMock := &restaurants.Mock{Restaurants: map[uuid.UUID]restaurants.Restaurant{
		sentID:    {ID: sentID, Name: "Sendable Cafe", Email: "owner@example.com", Status: "lead"},
		blockedID: {ID: blockedID, Name: "Blocked Cafe", Email: "blocked@example.com", Status: "lead"},
	}}
	accessService := restaurants.NewService(restaurantsMock, &restaurants.MembershipMock{})
	sendableCampaignID := uuid.New()
	campaignRepo := &campaigns.Mock{Campaigns: map[uuid.UUID]campaigns.Campaign{
		sendableCampaignID: {ID: sendableCampaignID, RestaurantID: sentID, Subject: "Hi", BodyHTML: "<p>hi</p>", BodyText: "hi"},
	}}
	campaignService := campaigns.NewService(campaignRepo, nil, accessService, nil, config.AppURLsConfig{
		PublicBaseURL: "https://api.example.com",
		PublicWebURL:  "https://example.com",
	})
	service := outreach.NewService(
		repo, nil, campaignRepo, campaignService, accessService,
		outreach.DemoTokenResolver{}, nil, provider,
		config.EmailConfig{Provider: "zoho"}, config.OutreachConfig{BulkMax: 150}, nil, nil,
	)

	results, err := service.SendAdHocBatch(context.Background(), adminPrincipal(), []uuid.UUID{sentID, blockedID})
	if err != nil {
		t.Fatalf("SendAdHocBatch() error = %v, want nil", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	byID := map[uuid.UUID]outreach.AdHocSendResult{}
	for _, r := range results {
		byID[r.RestaurantID] = r
	}
	if !byID[sentID].Sent {
		t.Fatalf("expected %v to be sent, got %+v", sentID, byID[sentID])
	}
	if byID[blockedID].Sent || byID[blockedID].Error == "" {
		t.Fatalf("expected %v to be blocked with an error, got %+v", blockedID, byID[blockedID])
	}
}
