package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	demos "github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

func TestPublicDemoDoesNotExposeInternalFields(t *testing.T) {
	token := "public-demo-token-value-32chars-min"
	tokenHash, err := demos.HashDemoToken(token)
	if err != nil {
		t.Fatalf("HashDemoToken() error = %v", err)
	}

	payload := json.RawMessage(`{
		"restaurant_name": "Sample Cafe",
		"cuisine": "Thai",
		"hero": "Welcome",
		"lead_notes": "secret lead note",
		"raw_enrichment": {"score": 99}
	}`)

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{},
		&restaurants.MembershipMock{},
		&demos.Mock{
			Sites: map[string]demos.Site{
				"sample-cafe": {
					ID:            uuid.New(),
					Slug:          "sample-cafe",
					TokenHash:     tokenHash,
					Status:        demos.StatusPublished,
					PublicPayload: payload,
				},
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/public/v1/demo/sample-cafe?token="+token, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	body := rec.Body.String()
	if strings.Contains(body, "lead_notes") || strings.Contains(body, "raw_enrichment") || strings.Contains(body, "secret lead note") {
		t.Fatalf("body leaked internal fields: %s", body)
	}
	if !strings.Contains(body, "Sample Cafe") {
		t.Fatalf("body = %s, want public restaurant name", body)
	}
}

func TestPublicDemoInvalidTokenReturnsNotFound(t *testing.T) {
	tokenHash, err := demos.HashDemoToken("valid-token-value-32chars-minimum")
	if err != nil {
		t.Fatalf("HashDemoToken() error = %v", err)
	}

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{},
		&restaurants.MembershipMock{},
		&demos.Mock{
			Sites: map[string]demos.Site{
				"sample-cafe": {
					ID:            uuid.New(),
					Slug:          "sample-cafe",
					TokenHash:     tokenHash,
					Status:        demos.StatusPublished,
					PublicPayload: demos.DefaultPublicPayload(),
				},
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/public/v1/demo/sample-cafe?token=wrong-token", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestPublicDemoDraftStatusReturnsNotFound(t *testing.T) {
	token := "draft-demo-token-value-32chars-min"
	tokenHash, err := demos.HashDemoToken(token)
	if err != nil {
		t.Fatalf("HashDemoToken() error = %v", err)
	}

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{},
		&restaurants.MembershipMock{},
		&demos.Mock{
			Sites: map[string]demos.Site{
				"draft-cafe": {
					ID:            uuid.New(),
					Slug:          "draft-cafe",
					TokenHash:     tokenHash,
					Status:        demos.StatusDraft,
					PublicPayload: demos.DefaultPublicPayload(),
				},
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/public/v1/demo/draft-cafe?token="+token, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestAdminCanCreateDemoSiteForRestaurant(t *testing.T) {
	adminID := uuid.New()
	restaurantID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				restaurantID: {ID: restaurantID, Name: "Demo Cafe", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
		},
		&restaurants.MembershipMock{},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(adminID, "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/"+restaurantID.String()+"/demo-sites", strings.NewReader(`{"slug":"demo-cafe","status":"draft"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"slug":"demo-cafe"`) {
		t.Fatalf("body = %s, want slug in response", rec.Body.String())
	}
}
