package outreach

import "github.com/google/uuid"

type EligibleLead struct {
	RestaurantID uuid.UUID
	Email        string
	Name         string
	DemoSiteID   uuid.UUID
	DemoSlug     string
}
