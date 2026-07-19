package profiles

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SiteRestaurantSummary struct {
	Index   int       `json:"index"`
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	PlaceID string    `json:"place_id,omitempty"`
	City    string    `json:"city,omitempty"`
}

type SiteMenuItem struct {
	Name         string          `json:"name"`
	Category     string          `json:"category"`
	Description  string          `json:"description"`
	Price        string          `json:"price"`
	PriceNumeric *float64        `json:"price_numeric,omitempty"`
	ImageURL     string          `json:"image_url,omitempty"`
	Images       json.RawMessage `json:"images,omitempty"`
}

type SiteReview struct {
	Reviewer string   `json:"reviewer"`
	Review   string   `json:"review"`
	Stars    *float64 `json:"stars,omitempty"`
	Date     string   `json:"date,omitempty"`
}

type SiteContent struct {
	Index         int             `json:"index"`
	RestaurantID  uuid.UUID       `json:"restaurant_id"`
	PlaceID       string          `json:"place_id,omitempty"`
	Name          string          `json:"name"`
	Cuisines      json.RawMessage `json:"cuisines"`
	Rating        *float64        `json:"rating,omitempty"`
	ReviewsCount  *int            `json:"reviews_count,omitempty"`
	PriceLevel    string          `json:"price_level,omitempty"`
	Phone         string          `json:"phone,omitempty"`
	Email         string          `json:"email,omitempty"`
	Website       string          `json:"website,omitempty"`
	Address       string          `json:"address,omitempty"`
	City          string          `json:"city,omitempty"`
	State         string          `json:"state,omitempty"`
	Country       string          `json:"country,omitempty"`
	Latitude      *float64        `json:"latitude,omitempty"`
	Longitude     *float64        `json:"longitude,omitempty"`
	Hours         json.RawMessage `json:"hours"`
	MenuItems     []SiteMenuItem  `json:"menu_items"`
	GalleryImages []GalleryImage  `json:"gallery_images"`
	Reviews       []SiteReview    `json:"reviews"`
	Thumbnail     string          `json:"thumbnail,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type SiteRepository interface {
	ListSiteRestaurants(ctx context.Context) ([]SiteRestaurantSummary, error)
	GetSiteRestaurantByID(ctx context.Context, restaurantID uuid.UUID) (SiteRestaurantSummary, error)
	GetSiteContentByID(ctx context.Context, restaurantID uuid.UUID) (SiteContent, error)
	GetSiteContentByIndex(ctx context.Context, index int) (SiteContent, error)
	GetSiteContentByPlaceID(ctx context.Context, placeID string) (SiteContent, error)
}
