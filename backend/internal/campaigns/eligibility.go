package campaigns

import (
	"fmt"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
)

type EligibilityInput struct {
	RestaurantEmail         string
	ReviewStatus            string
	ProfileReviewAudited    bool
	DemoStatus              string
	DemoPublishAudited      bool
	DemoExpired             bool
	CampaignStatus          string
	CampaignApprovalAudited bool
	Suppressed              bool
}

type BulkEligibilityInput struct {
	RestaurantEmail         string
	ReviewStatus            string
	ProfileReviewAudited    bool
	DemoStatus              string
	DemoPublishAudited      bool
	DemoExpired             bool
	CampaignStatus          string
	CampaignApprovalAudited bool
	Suppressed              bool
}

func CheckBulkEligibility(input BulkEligibilityInput) error {
	return CheckEligibility(EligibilityInput{
		RestaurantEmail:         input.RestaurantEmail,
		ReviewStatus:            input.ReviewStatus,
		ProfileReviewAudited:    input.ProfileReviewAudited,
		DemoStatus:              input.DemoStatus,
		DemoPublishAudited:      input.DemoPublishAudited,
		DemoExpired:             input.DemoExpired,
		CampaignStatus:          input.CampaignStatus,
		CampaignApprovalAudited: input.CampaignApprovalAudited,
		Suppressed:              input.Suppressed,
	})
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
	if !input.DemoPublishAudited {
		return fmt.Errorf("%w: demo publication is not attributable to an administrator", ErrNotEligible)
	}
	if input.DemoExpired {
		return fmt.Errorf("%w: demo link has expired and must be regenerated and republished", ErrNotEligible)
	}
	if input.CampaignStatus != StatusApproved && input.CampaignStatus != StatusSending {
		return fmt.Errorf("%w: campaign must be approved before sending", ErrNotEligible)
	}
	if !input.CampaignApprovalAudited {
		return fmt.Errorf("%w: campaign approval is not attributable to an administrator", ErrNotEligible)
	}
	reviewStatus := strings.TrimSpace(input.ReviewStatus)
	if reviewStatus == "" {
		reviewStatus = "draft"
	}
	if reviewStatus != "approved" {
		return fmt.Errorf("%w: restaurant profile is not approved for outreach", ErrNotEligible)
	}
	if !input.ProfileReviewAudited {
		return fmt.Errorf("%w: restaurant profile approval is not attributable to an administrator", ErrNotEligible)
	}
	return nil
}
