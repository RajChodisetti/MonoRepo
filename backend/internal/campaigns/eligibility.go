package campaigns

import (
	"fmt"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
)

type EligibilityInput struct {
	RestaurantEmail string
	ReviewStatus    string
	DemoStatus      string
	CampaignStatus  string
	Suppressed      bool
}

func CheckEligibility(input EligibilityInput) error {
	if strings.TrimSpace(input.RestaurantEmail) == "" {
		return fmt.Errorf("%w: restaurant has no contact email", ErrNotEligible)
	}
	if input.Suppressed {
		return fmt.Errorf("%w: recipient is suppressed", ErrNotEligible)
	}
	if input.DemoStatus != demos.StatusPublished {
		return fmt.Errorf("%w: demo site is not published", ErrNotEligible)
	}
	if input.CampaignStatus != StatusApproved && input.CampaignStatus != StatusSending {
		return fmt.Errorf("%w: campaign must be approved before sending", ErrNotEligible)
	}
	reviewStatus := strings.TrimSpace(input.ReviewStatus)
	if reviewStatus == "" {
		reviewStatus = "draft"
	}
	if reviewStatus != "approved" {
		return fmt.Errorf("%w: restaurant profile is not approved for outreach", ErrNotEligible)
	}
	return nil
}
