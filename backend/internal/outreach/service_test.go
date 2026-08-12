package outreach_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

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

func (repo *mockRepo) RecordAdHocEmailSent(ctx context.Context, restaurantID uuid.UUID, recipientEmail string) error {
	return nil
}

type mockEmailProvider struct {
	sent []emailprovider.SendRequest
}

func (provider *mockEmailProvider) Send(ctx context.Context, req emailprovider.SendRequest) (emailprovider.SendResult, error) {
	provider.sent = append(provider.sent, req)
	return emailprovider.SendResult{ProviderMessageID: "mock"}, nil
}

func testAccountPool(t *testing.T, providers ...emailprovider.Provider) *emailprovider.AccountPool {
	t.Helper()
	if len(providers) == 0 {
		providers = []emailprovider.Provider{&mockEmailProvider{}}
	}
	pool, err := emailprovider.NewAccountPool(providers, 50, 150)
	if err != nil {
		t.Fatalf("NewAccountPool() error = %v", err)
	}
	return pool
}

func TestRunBulkSendUsesExistingApprovedCampaign(t *testing.T) {
	restaurantID := uuid.New()
	demoSiteID := uuid.New()
	campaignID := uuid.New()
	approvedAt := time.Now().UTC()
	approvedBy := uuid.New()
	demoToken := "approved-demo-token"
	demoTokenHash, err := demos.HashDemoToken(demoToken)
	if err != nil {
		t.Fatalf("HashDemoToken() error = %v", err)
	}
	demoRepo := &demos.Mock{Sites: map[string]demos.Site{
		"approved-demo": {
			ID:           demoSiteID,
			RestaurantID: restaurantID,
			Slug:         "approved-demo",
			TokenHash:    demoTokenHash,
			Status:       demos.StatusPublished,
		},
	}}
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
				DemoToken:    demoToken,
				ApprovedAt:   &approvedAt,
				ApprovedBy:   &approvedBy,
			},
		},
		SendContexts: map[uuid.UUID]campaigns.SendContext{
			campaignID: {
				RestaurantEmail:      "owner@example.com",
				OCRStatus:            "verified",
				ReviewStatus:         "approved",
				ProfileReviewAudited: true,
				DemoStatus:           demos.StatusPublished,
				DemoPublishAudited:   true,
			},
		},
		SiteIndices: map[uuid.UUID]int{restaurantID: 0},
	}
	campaignService := campaigns.NewService(campaignRepo, demoRepo, nil, nil, config.AppURLsConfig{
		PublicBaseURL: "https://api.example.com",
		PublicWebURL:  "https://example.com",
	})
	provider := &mockEmailProvider{}
	service := outreach.NewService(
		&mockRepo{leads: []outreach.EligibleLead{{
			CampaignID:   campaignID,
			RestaurantID: restaurantID,
			DemoSiteID:   demoSiteID,
		}}},
		nil,
		campaignRepo,
		campaignService,
		nil,
		outreach.DemoTokenResolver{},
		testAccountPool(t, provider),
		nil,
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
	if len(provider.sent) != 1 {
		t.Fatalf("provider sent count = %d, want 1", len(provider.sent))
	}
	if !strings.Contains(provider.sent[0].HTMLBody, "tuvi-solutions-logo.png") {
		t.Fatal("bulk HTML body missing Tuvi logo signature")
	}
	if !strings.Contains(provider.sent[0].HTMLBody, "Team Tuvi") {
		t.Fatal("bulk HTML body missing Tuvi signature")
	}
	if !strings.Contains(provider.sent[0].TextBody, "https://tuvisolutions.com") {
		t.Fatal("bulk text body missing Tuvi website link")
	}
}
