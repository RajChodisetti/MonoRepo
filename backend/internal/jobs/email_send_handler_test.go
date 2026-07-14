package jobs

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

func TestEmailSendHandlerLeavesSkippedCampaignApproved(t *testing.T) {
	restaurantID := uuid.New()
	demoSiteID := uuid.New()
	campaignID := uuid.New()
	repo := &campaigns.Mock{
		Campaigns: map[uuid.UUID]campaigns.Campaign{
			campaignID: {
				ID:           campaignID,
				RestaurantID: restaurantID,
				DemoSiteID:   demoSiteID,
				Status:       campaigns.StatusSending,
				Subject:      "Test",
				BodyHTML:     "<p>Test</p>",
				BodyText:     "Test",
				DemoToken:    "demo-token",
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
	campaignService := campaigns.NewService(repo, nil, nil, nil, config.AppURLsConfig{
		PublicBaseURL: "https://api.example.com",
		PublicWebURL:  "https://example.com",
	})
	handler := EmailSendHandler(EmailSendDeps{
		Campaigns:        repo,
		CampaignsService: campaignService,
		Email:            emailprovider.NewDisabled(),
		EmailCfg:         config.EmailConfig{DisableSending: true},
	}, testLogger())
	job, err := NewEmailSendJob(campaignID, 0)
	if err != nil {
		t.Fatalf("NewEmailSendJob() error = %v", err)
	}

	if err := handler(context.Background(), job); err != nil {
		t.Fatalf("EmailSendHandler() error = %v", err)
	}
	if got := repo.Campaigns[campaignID].Status; got != campaigns.StatusApproved {
		t.Fatalf("campaign status = %q, want %q", got, campaigns.StatusApproved)
	}
	if repo.Campaigns[campaignID].LastSentAt != nil {
		t.Fatal("campaign LastSentAt was set for a skipped send")
	}
	if len(repo.Events) != 1 || repo.Events[0].EventType != campaigns.EventSkipped {
		t.Fatalf("events = %#v, want one skipped event", repo.Events)
	}
}

func TestEmailSendJobDoesNotRetryProviderSends(t *testing.T) {
	job, err := NewEmailSendJob(uuid.New(), 0)
	if err != nil {
		t.Fatalf("NewEmailSendJob() error = %v", err)
	}
	if job.MaxAttempts != 1 {
		t.Fatalf("MaxAttempts = %d, want 1", job.MaxAttempts)
	}
}
