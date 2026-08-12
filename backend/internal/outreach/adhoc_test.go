package outreach_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type adhocMockRepo struct {
	mockRepo
	recorded    []uuid.UUID
	recordError error
}

func (repo *adhocMockRepo) RecordAdHocEmailSent(ctx context.Context, restaurantID uuid.UUID, recipientEmail string) error {
	if repo.recordError != nil {
		return repo.recordError
	}
	repo.recorded = append(repo.recorded, restaurantID)
	return nil
}

type recordingEmailProvider struct {
	sent []emailprovider.SendRequest
	err  error
}

func (provider *recordingEmailProvider) Send(ctx context.Context, req emailprovider.SendRequest) (emailprovider.SendResult, error) {
	if provider.err != nil {
		return emailprovider.SendResult{}, provider.err
	}
	provider.sent = append(provider.sent, req)
	return emailprovider.SendResult{ProviderMessageID: "mock-adhoc"}, nil
}

func newAdHocTestService(t *testing.T, repo *adhocMockRepo, restaurantID uuid.UUID, email string, campaign *campaigns.Campaign, emailProvider emailprovider.Provider, disableSending bool) *outreach.Service {
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

	return outreach.NewService(
		repo,
		nil,
		campaignRepo,
		campaignService,
		accessService,
		outreach.DemoTokenResolver{},
		nil,
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

func TestSendAdHocRejectsWhenSendingDisabled(t *testing.T) {
	restaurantID := uuid.New()
	repo := &adhocMockRepo{}
	provider := &recordingEmailProvider{}
	service := newAdHocTestService(t, repo, restaurantID, "owner@example.com", nil, provider, true)

	_, err := service.SendAdHoc(context.Background(), adminPrincipal(), restaurantID)
	if !errors.Is(err, outreach.ErrSendingDisabled) {
		t.Fatalf("SendAdHoc() error = %v, want ErrSendingDisabled", err)
	}
	if len(provider.sent) != 0 {
		t.Fatalf("expected no send attempt, got %d", len(provider.sent))
	}
}

func TestSendAdHocRejectsWhenNoCampaignDraft(t *testing.T) {
	restaurantID := uuid.New()
	repo := &adhocMockRepo{}
	provider := &recordingEmailProvider{}
	service := newAdHocTestService(t, repo, restaurantID, "owner@example.com", nil, provider, false)

	_, err := service.SendAdHoc(context.Background(), adminPrincipal(), restaurantID)
	if !errors.Is(err, outreach.ErrNoCampaignDraft) {
		t.Fatalf("SendAdHoc() error = %v, want ErrNoCampaignDraft", err)
	}
}

func TestSendAdHocSendsAndRecords(t *testing.T) {
	restaurantID := uuid.New()
	repo := &adhocMockRepo{}
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
	if !strings.Contains(provider.sent[0].HTMLBody, "tuvi-solutions-logo.png") {
		t.Fatal("ad hoc HTML body missing Tuvi logo signature")
	}
	if !strings.Contains(provider.sent[0].HTMLBody, "Team Tuvi") {
		t.Fatal("ad hoc HTML body missing Tuvi signature")
	}
	if !strings.Contains(provider.sent[0].TextBody, "https://tuvisolutions.com") {
		t.Fatal("ad hoc text body missing Tuvi website link")
	}
	if len(repo.recorded) != 1 || repo.recorded[0] != restaurantID {
		t.Fatalf("repo.recorded = %v, want [%v]", repo.recorded, restaurantID)
	}
}

func TestSendAdHocBatchCollectsPerLeadResults(t *testing.T) {
	sentID := uuid.New()
	// blockedID intentionally gets no campaign entry below, so it fails with
	// ErrNoCampaignDraft — the point of this test is that a batch reports
	// mixed per-lead results rather than succeeding or failing as a whole.
	blockedID := uuid.New()
	repo := &adhocMockRepo{}
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
