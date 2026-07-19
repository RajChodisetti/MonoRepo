package demos

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

func TestMapPublicPayloadFiltersInternalFields(t *testing.T) {
	raw := json.RawMessage(`{
		"restaurant_name": "Cafe",
		"cuisine": "Indian",
		"lead_notes": "private",
		"raw_enrichment": {"score": 1}
	}`)

	payload := MapPublicPayload(raw)
	if payload.RestaurantName != "Cafe" {
		t.Fatalf("RestaurantName = %q, want Cafe", payload.RestaurantName)
	}
	if payload.Cuisine != "Indian" {
		t.Fatalf("Cuisine = %q, want Indian", payload.Cuisine)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	body := string(encoded)
	if strings.Contains(body, "lead_notes") || strings.Contains(body, "raw_enrichment") {
		t.Fatalf("public payload leaked internal fields: %s", body)
	}
}

func TestResolvePublicDemoReturnsPayloadForValidSlugAndToken(t *testing.T) {
	slug := "cafe-demo"
	token := "valid-token"
	tokenHash, err := HashDemoToken(token)
	if err != nil {
		t.Fatalf("HashDemoToken() error = %v", err)
	}

	demosRepo := &Mock{
		Sites: map[string]Site{
			slug: {
				ID:            uuid.New(),
				Slug:          slug,
				TokenHash:     tokenHash,
				Status:        StatusPublished,
				PublicPayload: DefaultPublicPayload(),
			},
		},
	}
	service := NewService(demosRepo, restaurants.NewService(&restaurants.Mock{}, &restaurants.MembershipMock{}), 24*time.Hour)

	payload, err := service.ResolvePublicDemo(context.Background(), slug, token)
	if err != nil {
		t.Fatalf("ResolvePublicDemo() error = %v", err)
	}
	if payload.RestaurantName == "" {
		t.Fatalf("expected public payload, got empty restaurant name")
	}
}

func TestCreateDemoSiteRequiresInternalAdmin(t *testing.T) {
	service := NewService(&Mock{}, restaurants.NewService(&restaurants.Mock{}, &restaurants.MembershipMock{}), time.Hour)

	_, err := service.CreateDemoSite(context.Background(), auth.Principal{
		UserID: uuid.New(),
		Role:   auth.RoleRestaurantOwner,
	}, uuid.New(), CreateDemoInput{Slug: "demo"})
	if err == nil {
		t.Fatal("expected forbidden error for restaurant owner")
	}
}

func TestCreateDemoSiteBuildsRestaurantSpecificPayload(t *testing.T) {
	restaurantID := uuid.New()
	payload := json.RawMessage(`{"restaurant_name":"Real Restaurant","cuisine":"Indian"}`)
	demosRepo := &Mock{PublicPayloads: map[uuid.UUID]json.RawMessage{restaurantID: payload}}
	restaurantsRepo := &restaurants.Mock{Restaurants: map[uuid.UUID]restaurants.Restaurant{
		restaurantID: {ID: restaurantID, Name: "Real Restaurant", Status: restaurants.StatusLead},
	}}
	service := NewService(
		demosRepo,
		restaurants.NewService(restaurantsRepo, &restaurants.MembershipMock{}),
		time.Hour,
	)

	result, err := service.CreateDemoSite(context.Background(), auth.Principal{
		UserID: uuid.New(),
		Role:   auth.RoleInternalAdmin,
	}, restaurantID, CreateDemoInput{Slug: "real-restaurant"})
	if err != nil {
		t.Fatalf("CreateDemoSite() error = %v", err)
	}
	record, err := demosRepo.GetByID(context.Background(), result.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if string(record.PublicPayload) != string(payload) {
		t.Fatalf("PublicPayload = %s, want %s", record.PublicPayload, payload)
	}
	restaurant, err := restaurantsRepo.GetByID(context.Background(), restaurantID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if restaurant.Status != restaurants.StatusDemoReady {
		t.Fatalf("restaurant status = %q, want %q", restaurant.Status, restaurants.StatusDemoReady)
	}
}
