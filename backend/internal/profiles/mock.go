package profiles

import (
	"context"
	"time"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Mock struct {
	RestaurantIDByPlaceID map[string]uuid.UUID
	GooglePlaceIDs        map[uuid.UUID]string
	MenuImages            map[uuid.UUID][]MenuImage
	GalleryImages         map[uuid.UUID][]GalleryImage
	SiteRestaurants       []SiteRestaurantSummary
}

func (mock *Mock) ListMenuImagesAdmin(ctx context.Context, restaurantID uuid.UUID) ([]MenuImage, error) {
	return mock.ListMenuImages(ctx, restaurantID)
}

func (mock *Mock) ListGalleryImagesAdmin(ctx context.Context, restaurantID uuid.UUID) ([]GalleryImage, error) {
	return mock.ListGalleryImages(ctx, restaurantID)
}

func (mock *Mock) HideMenuImage(_ context.Context, restaurantID, imageID, hiddenBy uuid.UUID) error {
	for i, img := range mock.MenuImages[restaurantID] {
		if img.ID == imageID {
			now := time.Now()
			mock.MenuImages[restaurantID][i].HiddenAt = &now
			mock.MenuImages[restaurantID][i].HiddenBy = &hiddenBy
			return nil
		}
	}
	return repository.ErrNotFound
}

func (mock *Mock) HideGalleryImage(_ context.Context, restaurantID, imageID, hiddenBy uuid.UUID) error {
	for i, img := range mock.GalleryImages[restaurantID] {
		if img.ID == imageID {
			now := time.Now()
			mock.GalleryImages[restaurantID][i].HiddenAt = &now
			mock.GalleryImages[restaurantID][i].HiddenBy = &hiddenBy
			return nil
		}
	}
	return repository.ErrNotFound
}

func (mock *Mock) UnhideMenuImage(_ context.Context, restaurantID, imageID uuid.UUID) error {
	for i, img := range mock.MenuImages[restaurantID] {
		if img.ID == imageID {
			mock.MenuImages[restaurantID][i].HiddenAt = nil
			mock.MenuImages[restaurantID][i].HiddenBy = nil
			return nil
		}
	}
	return repository.ErrNotFound
}

func (mock *Mock) UnhideGalleryImage(_ context.Context, restaurantID, imageID uuid.UUID) error {
	for i, img := range mock.GalleryImages[restaurantID] {
		if img.ID == imageID {
			mock.GalleryImages[restaurantID][i].HiddenAt = nil
			mock.GalleryImages[restaurantID][i].HiddenBy = nil
			return nil
		}
	}
	return repository.ErrNotFound
}

func (mock *Mock) GetRestaurantIDByPlaceID(_ context.Context, placeID string) (uuid.UUID, error) {
	if mock.RestaurantIDByPlaceID == nil {
		return uuid.Nil, repository.ErrNotFound
	}
	id, ok := mock.RestaurantIDByPlaceID[placeID]
	if !ok {
		return uuid.Nil, repository.ErrNotFound
	}
	return id, nil
}

func (mock *Mock) GetGooglePlaceID(_ context.Context, restaurantID uuid.UUID) (string, error) {
	placeID := mock.GooglePlaceIDs[restaurantID]
	if placeID == "" {
		return "", repository.ErrNotFound
	}
	return placeID, nil
}

func (mock *Mock) ListMenuImages(_ context.Context, restaurantID uuid.UUID) ([]MenuImage, error) {
	if mock.MenuImages == nil {
		return nil, nil
	}
	return mock.MenuImages[restaurantID], nil
}

func (mock *Mock) ListGalleryImages(_ context.Context, restaurantID uuid.UUID) ([]GalleryImage, error) {
	if mock.GalleryImages == nil {
		return nil, nil
	}
	return mock.GalleryImages[restaurantID], nil
}

func (mock *Mock) GetSiteImages(_ context.Context, restaurantID uuid.UUID) (SiteImages, error) {
	menuImages, _ := mock.ListMenuImages(context.Background(), restaurantID)
	galleryImages, _ := mock.ListGalleryImages(context.Background(), restaurantID)
	return SiteImages{
		RestaurantID:  restaurantID,
		MenuImages:    menuImages,
		GalleryImages: galleryImages,
	}, nil
}

func (mock *Mock) GetSiteImagesByPlaceID(ctx context.Context, placeID string) (SiteImages, error) {
	restaurantID, err := mock.GetRestaurantIDByPlaceID(ctx, placeID)
	if err != nil {
		return SiteImages{}, err
	}
	return mock.GetSiteImages(ctx, restaurantID)
}

func (mock *Mock) ListSiteRestaurants(_ context.Context) ([]SiteRestaurantSummary, error) {
	return mock.SiteRestaurants, nil
}

func (mock *Mock) GetSiteRestaurantByID(_ context.Context, restaurantID uuid.UUID) (SiteRestaurantSummary, error) {
	for _, summary := range mock.SiteRestaurants {
		if summary.ID == restaurantID {
			return summary, nil
		}
	}
	return SiteRestaurantSummary{}, repository.ErrNotFound
}

func (mock *Mock) GetSiteContentByIndex(_ context.Context, index int) (SiteContent, error) {
	return SiteContent{Index: index}, nil
}

func (mock *Mock) GetSiteContentByPlaceID(_ context.Context, placeID string) (SiteContent, error) {
	return SiteContent{PlaceID: placeID}, nil
}
