package outreach_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type mockRepo struct {
	count int
	leads []outreach.EligibleLead
}

func (repo *mockRepo) ListEligibleLeads(ctx context.Context, limit int) ([]outreach.EligibleLead, error) {
	if limit >= len(repo.leads) {
		return repo.leads, nil
	}
	return repo.leads[:limit], nil
}

func (repo *mockRepo) CountEligibleLeads(ctx context.Context) (int, error) {
	return repo.count, nil
}

type mockEnqueuer struct {
	jobID string
	err   error
}

func (enqueuer *mockEnqueuer) EnqueueBulkSend(ctx context.Context, triggeredBy uuid.UUID) (string, error) {
	if enqueuer.err != nil {
		return "", enqueuer.err
	}
	return enqueuer.jobID, nil
}

type mockEmailProvider struct{}

func (provider *mockEmailProvider) Send(ctx context.Context, req emailprovider.SendRequest) (emailprovider.SendResult, error) {
	return emailprovider.SendResult{ProviderMessageID: "mock"}, nil
}

func testAccountPool(t *testing.T) *emailprovider.AccountPool {
	t.Helper()
	pool, err := emailprovider.NewAccountPool([]emailprovider.Provider{&mockEmailProvider{}}, 50, 150)
	if err != nil {
		t.Fatalf("NewAccountPool() error = %v", err)
	}
	return pool
}

func TestTriggerBulkSendRequiresConfiguredAccounts(t *testing.T) {
	service := outreach.NewService(
		&mockRepo{count: 3},
		nil,
		nil,
		nil,
		outreach.DemoTokenResolver{},
		nil,
		config.EmailConfig{Provider: "zoho"},
		config.OutreachConfig{BulkMax: 150},
		&mockEnqueuer{jobID: "job-1"},
		nil,
	)

	_, err := service.TriggerBulkSend(context.Background(), auth.Principal{
		Role:   auth.RoleInternalAdmin,
		UserID: uuid.New(),
	})
	if !errors.Is(err, outreach.ErrNotConfigured) {
		t.Fatalf("TriggerBulkSend() error = %v, want ErrNotConfigured", err)
	}
}

func TestTriggerBulkSendEnqueuesJob(t *testing.T) {
	service := outreach.NewService(
		&mockRepo{count: 7},
		nil,
		nil,
		nil,
		outreach.DemoTokenResolver{},
		testAccountPool(t),
		config.EmailConfig{Provider: "zoho"},
		config.OutreachConfig{
			BulkMax:      150,
			ZohoAccounts: []config.ZohoMailConfig{{AccountID: "1", ClientID: "a", ClientSecret: "b", RefreshToken: "c"}},
		},
		&mockEnqueuer{jobID: "job-123"},
		nil,
	)

	result, err := service.TriggerBulkSend(context.Background(), auth.Principal{
		Role:   auth.RoleInternalAdmin,
		UserID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("TriggerBulkSend() error = %v, want nil", err)
	}
	if result.JobID != "job-123" {
		t.Fatalf("JobID = %q, want job-123", result.JobID)
	}
	if result.PendingEligibleCount != 7 {
		t.Fatalf("PendingEligibleCount = %d, want 7", result.PendingEligibleCount)
	}
}

func TestTriggerBulkSendRejectsDisabledSending(t *testing.T) {
	service := outreach.NewService(
		&mockRepo{count: 7},
		nil,
		nil,
		nil,
		outreach.DemoTokenResolver{},
		testAccountPool(t),
		config.EmailConfig{Provider: "zoho", DisableSending: true},
		config.OutreachConfig{
			BulkMax:      150,
			ZohoAccounts: []config.ZohoMailConfig{{AccountID: "1", ClientID: "a", ClientSecret: "b", RefreshToken: "c"}},
		},
		&mockEnqueuer{jobID: "job-123"},
		nil,
	)

	_, err := service.TriggerBulkSend(context.Background(), auth.Principal{
		Role:   auth.RoleInternalAdmin,
		UserID: uuid.New(),
	})
	if !errors.Is(err, outreach.ErrSendingDisabled) {
		t.Fatalf("TriggerBulkSend() error = %v, want ErrSendingDisabled", err)
	}
}

func TestRunBulkSendUsesExistingApprovedCampaign(t *testing.T) {
	restaurantID := uuid.New()
	demoSiteID := uuid.New()
	campaignID := uuid.New()
	campaignRepo := &campaigns.Mock{
		Campaigns: map[uuid.UUID]campaigns.Campaign{
			campaignID: {
				ID:           campaignID,
				RestaurantID: restaurantID,
				DemoSiteID:   demoSiteID,
				CampaignType: campaigns.TypeOutreach,
				Status:       campaigns.StatusApproved,
				Subject:      "Approved subject",
				BodyHTML:     "<p>Approved body</p>",
				BodyText:     "Approved body",
				DemoToken:    "approved-demo-token",
			},
		},
		SendContexts: map[uuid.UUID]campaigns.SendContext{
			campaignID: {
				RestaurantEmail: "owner@example.com",
				ReviewStatus:    "approved",
				DemoStatus:      demos.StatusPublished,
			},
		},
		SiteIndices: map[uuid.UUID]int{restaurantID: 0},
	}
	campaignService := campaigns.NewService(campaignRepo, nil, nil, nil, config.AppURLsConfig{
		PublicBaseURL: "https://api.example.com",
		PublicWebURL:  "https://example.com",
	})
	service := outreach.NewService(
		&mockRepo{leads: []outreach.EligibleLead{{
			CampaignID:   campaignID,
			RestaurantID: restaurantID,
			DemoSiteID:   demoSiteID,
		}}},
		nil,
		campaignRepo,
		campaignService,
		outreach.DemoTokenResolver{},
		testAccountPool(t),
		config.EmailConfig{Provider: "zoho"},
		config.OutreachConfig{
			BulkMax:      150,
			ZohoAccounts: []config.ZohoMailConfig{{AccountID: "1"}},
		},
		nil,
		nil,
	)

	summary, err := service.RunBulkSend(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("RunBulkSend() error = %v", err)
	}
	if summary.Sent != 1 || summary.Attempted != 1 {
		t.Fatalf("summary = %#v, want one sent attempt", summary)
	}
	if len(campaignRepo.Campaigns) != 1 {
		t.Fatalf("campaign count = %d, want existing campaign only", len(campaignRepo.Campaigns))
	}
	if got := campaignRepo.Campaigns[campaignID].Status; got != campaigns.StatusSent {
		t.Fatalf("campaign status = %q, want %q", got, campaigns.StatusSent)
	}
}
