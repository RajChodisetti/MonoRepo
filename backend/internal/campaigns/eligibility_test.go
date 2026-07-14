package campaigns_test

import (
	"errors"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
)

func TestCheckEligibilityRequiresApprovedProfile(t *testing.T) {
	err := campaigns.CheckEligibility(campaigns.EligibilityInput{
		RestaurantEmail: "owner@example.com",
		ReviewStatus:    "draft",
		DemoStatus:      demos.StatusPublished,
		CampaignStatus:  campaigns.StatusApproved,
	})
	if err == nil {
		t.Fatal("CheckEligibility() error = nil, want eligibility error")
	}
	if !errors.Is(err, campaigns.ErrNotEligible) {
		t.Fatalf("CheckEligibility() error = %v, want ErrNotEligible", err)
	}
}

func TestCheckEligibilityPassesWhenReady(t *testing.T) {
	err := campaigns.CheckEligibility(campaigns.EligibilityInput{
		RestaurantEmail: "owner@example.com",
		ReviewStatus:    "approved",
		DemoStatus:      demos.StatusPublished,
		CampaignStatus:  campaigns.StatusApproved,
	})
	if err != nil {
		t.Fatalf("CheckEligibility() error = %v, want nil", err)
	}
}

func TestCheckBulkEligibilityRequiresApprovedReviewAndCampaign(t *testing.T) {
	err := campaigns.CheckBulkEligibility(campaigns.BulkEligibilityInput{
		RestaurantEmail: "owner@example.com",
		ReviewStatus:    "draft",
		DemoStatus:      demos.StatusPublished,
		CampaignStatus:  campaigns.StatusApproved,
	})
	if err == nil || !errors.Is(err, campaigns.ErrNotEligible) {
		t.Fatalf("CheckBulkEligibility() error = %v, want ErrNotEligible", err)
	}
}

func TestCheckBulkEligibilityRequiresPublishedDemo(t *testing.T) {
	err := campaigns.CheckBulkEligibility(campaigns.BulkEligibilityInput{
		RestaurantEmail: "owner@example.com",
		ReviewStatus:    "approved",
		DemoStatus:      demos.StatusDraft,
		CampaignStatus:  campaigns.StatusApproved,
	})
	if err == nil {
		t.Fatal("CheckBulkEligibility() error = nil, want eligibility error")
	}
	if !errors.Is(err, campaigns.ErrNotEligible) {
		t.Fatalf("CheckBulkEligibility() error = %v, want ErrNotEligible", err)
	}
}
