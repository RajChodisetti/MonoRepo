package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
)

func TestGeneratedSiteURL(t *testing.T) {
	got := generatedSiteURL("https://demo.example.test", "d419296d-17f6-4bc8-b522-943933451d74", 42, "3")
	want := "https://demo.example.test?id=42&restaurant_id=d419296d-17f6-4bc8-b522-943933451d74&template=3"
	if got != want {
		t.Fatalf("generatedSiteURL() = %q, want %q", got, want)
	}
}

func TestRestaurantSiteAdminTemplatesPreferElysian(t *testing.T) {
	restaurantID := uuid.New()
	handler := NewRestaurantSiteAdminHandler(
		&profiles.Mock{SiteRestaurants: []profiles.SiteRestaurantSummary{{
			ID:      restaurantID,
			Index:   42,
			Name:    "Test Restaurant",
			PlaceID: "place-1",
		}}},
		"https://demo.example.test",
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

	request := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/"+restaurantID.String()+"/generated-site", nil)
	request.SetPathValue("id", restaurantID.String())
	response := httptest.NewRecorder()
	handler.Get(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	var payload struct {
		Templates []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"templates"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Templates) != 3 {
		t.Fatalf("template count = %d, want 3", len(payload.Templates))
	}

	wantIDs := []string{"3", "2", "1"}
	wantNames := []string{"Elysian reservations", "Aurora", "Cinematic"}
	for index, template := range payload.Templates {
		if template.ID != wantIDs[index] || template.Name != wantNames[index] {
			t.Fatalf("template[%d] = %#v, want id %q and name %q", index, template, wantIDs[index], wantNames[index])
		}
	}
}
