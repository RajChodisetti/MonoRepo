package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
	placesprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/places"
)

type fakePhotoResolver struct {
	photos []placesprovider.Photo
	err    error
}

func (resolver fakePhotoResolver) ListPhotoURLs(context.Context, string) ([]placesprovider.Photo, error) {
	return resolver.photos, resolver.err
}

func TestListGoogleReturnsFreshURLsWithoutCaching(t *testing.T) {
	restaurantID := uuid.New()
	profilesRepo := &profiles.Mock{GooglePlaceIDs: map[uuid.UUID]string{
		restaurantID: "place-1",
	}}
	handler := NewRestaurantImagesAdminHandler(
		profilesRepo,
		fakePhotoResolver{photos: []placesprovider.Photo{{URL: "https://images.example.test/photo.jpg"}}},
		func(w http.ResponseWriter, status int, value any) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(value)
		},
		func(w http.ResponseWriter, status int, code string, message string) {
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": message})
		},
	)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/"+restaurantID.String()+"/images/google", nil)
	request.SetPathValue("id", restaurantID.String())
	response := httptest.NewRecorder()
	handler.ListGoogle(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", response.Header().Get("Cache-Control"))
	}
	var payload struct {
		GooglePlaceID string                 `json:"google_place_id"`
		Photos        []placesprovider.Photo `json:"photos"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.GooglePlaceID != "place-1" || len(payload.Photos) != 1 {
		t.Fatalf("payload = %#v", payload)
	}
}
