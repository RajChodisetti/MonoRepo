package outreach

import "github.com/google/uuid"

type EligibleLead struct {
	CampaignID   uuid.UUID
	RestaurantID uuid.UUID
	DemoSiteID   uuid.UUID
}
