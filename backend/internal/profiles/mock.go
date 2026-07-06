package profiles

import (
	"context"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Mock struct {
	RestaurantIDByPlaceID map[string]uuid.UUID
	MenuImages            map[uuid.UUID][]MenuImage
	GalleryImages         map[uuid.UUID][]GalleryImage
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
	return nil, nil
}

func (mock *Mock) GetSiteContentByIndex(_ context.Context, index int) (SiteContent, error) {
	return SiteContent{Index: index}, nil
}

func (mock *Mock) GetSiteContentByPlaceID(_ context.Context, placeID string) (SiteContent, error) {
	return SiteContent{PlaceID: placeID}, nil
}
