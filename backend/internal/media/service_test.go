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
	hints  []ClassificationHint
}

func (stub repositoryStub) ListPublic(context.Context, uuid.UUID) ([]Asset, error) {
	return stub.assets, nil
}
func (stub repositoryStub) ListAdmin(context.Context, uuid.UUID) ([]Asset, error) {
	return stub.assets, nil
}
func (stub repositoryStub) ListClassificationHints(context.Context, uuid.UUID) ([]ClassificationHint, error) {
	return stub.hints, nil
}
func (repositoryStub) Create(context.Context, CreateAssetInput) (Asset, error) { return Asset{}, nil }
func (repositoryStub) SetHidden(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID) error {
	return nil
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

func TestPublicForRestaurantExcludesMenuIneligibleAndUnclassifiedGooglePhotos(t *testing.T) {
	restaurantID := uuid.New()
	service := NewService(
		repositoryStub{hints: []ClassificationHint{
			{SourceIndex: 0, SourceFingerprint: "menu-fingerprint", MediaType: "menu_document", Confidence: 0.99, PublicEligible: false},
			{SourceIndex: 1, SourceFingerprint: "room-fingerprint", MediaType: "interior", Confidence: 0.92, PublicEligible: true},
			{SourceIndex: 2, SourceFingerprint: "uncertain-fingerprint", MediaType: "exterior", Confidence: 0.3, PublicEligible: false},
		}},
		profilesStub{placeID: "place-1"},
		photosStub{photos: []placesprovider.Photo{
			{URL: "https://images.example.test/menu.jpg", SourceIndex: 0, SourceFingerprint: "menu-fingerprint"},
			{URL: "https://images.example.test/room.jpg", SourceIndex: 1, SourceFingerprint: "room-fingerprint"},
			{URL: "https://images.example.test/uncertain.jpg", SourceIndex: 2, SourceFingerprint: "uncertain-fingerprint"},
			{URL: "https://images.example.test/unknown.jpg", SourceIndex: 3, SourceFingerprint: "unknown-fingerprint"},
		}},
		objectsStub{},
		slog.Default(),
	)

	items := service.PublicForRestaurant(context.Background(), restaurantID, 6)
	if len(items) != 1 {
		t.Fatalf("items = %#v, want one OCR-approved photo", items)
	}
	if items[0].URL != "https://images.example.test/room.jpg" || items[0].MediaType != "interior" {
		t.Fatalf("item = %#v", items[0])
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
