package profiles

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MenuImage struct {
	ID           uuid.UUID       `json:"id"`
	RestaurantID uuid.UUID       `json:"restaurant_id"`
	URL          string          `json:"url"`
	ThumbnailURL string          `json:"thumbnail_url"`
	ImageType    string          `json:"image_type"`
	Confidence   *float64        `json:"confidence,omitempty"`
	Title        string          `json:"title"`
	Source       string          `json:"source"`
	SortOrder    int             `json:"sort_order"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	HiddenAt     *time.Time      `json:"hidden_at,omitempty"`
	HiddenBy     *uuid.UUID      `json:"hidden_by,omitempty"`
}

type GalleryImage struct {
	ID           uuid.UUID       `json:"id"`
	RestaurantID uuid.UUID       `json:"restaurant_id"`
	URL          string          `json:"url"`
	ThumbnailURL string          `json:"thumbnail_url"`
	ImageType    string          `json:"image_type"`
	Confidence   *float64        `json:"confidence,omitempty"`
	Title        string          `json:"title"`
	Source       string          `json:"source"`
	SortOrder    int             `json:"sort_order"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	HiddenAt     *time.Time      `json:"hidden_at,omitempty"`
	HiddenBy     *uuid.UUID      `json:"hidden_by,omitempty"`
}

type SiteImages struct {
	RestaurantID  uuid.UUID      `json:"restaurant_id"`
	MenuImages    []MenuImage    `json:"menu_images"`
	GalleryImages []GalleryImage `json:"gallery_images"`
}

type Repository interface {
	GetRestaurantIDByPlaceID(ctx context.Context, placeID string) (uuid.UUID, error)
	GetGooglePlaceID(ctx context.Context, restaurantID uuid.UUID) (string, error)
	ListMenuImages(ctx context.Context, restaurantID uuid.UUID) ([]MenuImage, error)
	ListGalleryImages(ctx context.Context, restaurantID uuid.UUID) ([]GalleryImage, error)
	GetSiteImages(ctx context.Context, restaurantID uuid.UUID) (SiteImages, error)
	GetSiteImagesByPlaceID(ctx context.Context, placeID string) (SiteImages, error)

	// Admin variants include hidden images; the plain List* methods above do not.
	ListMenuImagesAdmin(ctx context.Context, restaurantID uuid.UUID) ([]MenuImage, error)
	ListGalleryImagesAdmin(ctx context.Context, restaurantID uuid.UUID) ([]GalleryImage, error)
	HideMenuImage(ctx context.Context, restaurantID, imageID, hiddenBy uuid.UUID) error
	HideGalleryImage(ctx context.Context, restaurantID, imageID, hiddenBy uuid.UUID) error
	UnhideMenuImage(ctx context.Context, restaurantID, imageID uuid.UUID) error
	UnhideGalleryImage(ctx context.Context, restaurantID, imageID uuid.UUID) error

	SiteRepository
}
