package media

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	placesprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/places"
)

type repositoryStub struct {
	assets []Asset
}

func (stub repositoryStub) ListPublic(context.Context, uuid.UUID) ([]Asset, error) {
	return stub.assets, nil
}
func (stub repositoryStub) ListAdmin(context.Context, uuid.UUID) ([]Asset, error) {
	return stub.assets, nil
}
func (repositoryStub) Create(context.Context, CreateAssetInput) (Asset, error) { return Asset{}, nil }
func (repositoryStub) SetHidden(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID) error {
	return nil
}
func (repositoryStub) SetApproval(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, string) (Asset, error) {
	return Asset{}, nil
}

type profilesStub struct{ placeID string }

func (stub profilesStub) GetGooglePlaceID(context.Context, uuid.UUID) (string, error) {
	return stub.placeID, nil
}

type photosStub struct{ photos []placesprovider.Photo }

func (stub photosStub) ListPhotoURLs(context.Context, string, int) ([]placesprovider.Photo, error) {
	return stub.photos, nil
}

type objectsStub struct{}

func (objectsStub) Configured() bool                                            { return true }
func (objectsStub) Put(context.Context, string, string, io.Reader, int64) error { return nil }
func (objectsStub) Delete(context.Context, string) error                        { return nil }
func (objectsStub) PublicURL(key string) string                                 { return "https://cdn.example.test/" + key }

func TestPublicForRestaurantIncludesFreshAttributedGooglePhotos(t *testing.T) {
	restaurantID := uuid.New()
	service := NewService(
		repositoryStub{},
		profilesStub{placeID: "place-1"},
		photosStub{photos: []placesprovider.Photo{
			{
				URL:                "https://images.example.test/room.jpg",
				AuthorAttributions: []placesprovider.Attribution{{DisplayName: "Restaurant owner"}},
			},
		}},
		objectsStub{},
		slog.Default(),
	)

	items := service.PublicForRestaurant(context.Background(), restaurantID, 6)
	if len(items) != 1 {
		t.Fatalf("items = %#v, want one live photo", items)
	}
	if items[0].URL != "https://images.example.test/room.jpg" ||
		items[0].SourceKind != SourceGoogleLive ||
		!items[0].Unoptimized ||
		len(items[0].AuthorAttributions) != 1 {
		t.Fatalf("item = %#v, want attributed no-store Google media", items[0])
	}
}

func TestPublicOwnedAssetsSkipsInvalidRowsWithoutDroppingLaterAssets(t *testing.T) {
	validID := uuid.New()
	service := NewService(
		repositoryStub{assets: []Asset{
			{ID: uuid.New(), StorageKey: "menu.jpg", MediaType: "menu_document"},
			{ID: validID, StorageKey: "room.jpg", MediaType: "interior"},
		}},
		nil,
		nil,
		objectsStub{},
		slog.Default(),
	)

	items := service.PublicForRestaurant(context.Background(), uuid.New(), 6)
	if len(items) != 1 || items[0].ID == nil || *items[0].ID != validID {
		t.Fatalf("items = %#v, want later valid owned asset", items)
	}
}

func TestPreviewForRestaurantFallsBackToLiveGooglePhotos(t *testing.T) {
	restaurantID := uuid.New()
	service := NewService(
		repositoryStub{},
		profilesStub{placeID: "place-1"},
		photosStub{photos: []placesprovider.Photo{
			{
				URL:           "https://lh3.googleusercontent.com/photo",
				WidthPx:       1200,
				HeightPx:      800,
				GoogleMapsURI: "https://maps.google.com/place/1",
				AuthorAttributions: []placesprovider.Attribution{
					{DisplayName: "Ava", URI: "https://maps.google.com/contrib/ava"},
				},
			},
		}},
		objectsStub{},
		slog.Default(),
	)

	items := service.PreviewForRestaurant(context.Background(), restaurantID, 6)
	if len(items) != 1 {
		t.Fatalf("items = %#v, want one live preview photo", items)
	}
	if items[0].SourceKind != SourceGoogleLive || !items[0].Unoptimized || len(items[0].AuthorAttributions) != 1 {
		t.Fatalf("item = %#v, want attributed live Google preview media", items[0])
	}
}

func TestPreviewForRestaurantKeepsApprovedMediaWhenAvailable(t *testing.T) {
	validID := uuid.New()
	service := NewService(
		repositoryStub{assets: []Asset{
			{ID: validID, StorageKey: "room.jpg", MediaType: "interior", ApprovalStatus: "approved"},
		}},
		profilesStub{placeID: "place-1"},
		photosStub{photos: []placesprovider.Photo{{URL: "https://lh3.googleusercontent.com/photo"}}},
		objectsStub{},
		slog.Default(),
	)

	items := service.PreviewForRestaurant(context.Background(), uuid.New(), 6)
	if len(items) != 2 || items[0].ID == nil || *items[0].ID != validID || items[1].SourceKind != SourceGoogleLive {
		t.Fatalf("items = %#v, want approved owned media followed by live Google media", items)
	}
}
