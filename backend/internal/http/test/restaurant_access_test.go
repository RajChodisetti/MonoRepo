package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

func TestAuthMeExposesUserIDAndRole(t *testing.T) {
	userID := uuid.New()
	router := testRouter(t, fakeReadiness{})
	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(userID, "owner@example.com", auth.RoleRestaurantOwner)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), userID.String()) || !strings.Contains(rec.Body.String(), auth.RoleRestaurantOwner) {
		t.Fatalf("body = %s, want user id and role", rec.Body.String())
	}
}

func TestRestaurantOwnerCannotAccessAnotherRestaurantHTTP(t *testing.T) {
	ownerID := uuid.New()
	ownedRestaurantID := uuid.New()
	otherRestaurantID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				ownedRestaurantID: {ID: ownedRestaurantID, Name: "Owned"},
				otherRestaurantID: {ID: otherRestaurantID, Name: "Other"},
			},
		},
		&restaurants.MembershipMock{
			Members: map[uuid.UUID]map[uuid.UUID]bool{
				ownerID: {ownedRestaurantID: true},
			},
		},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(ownerID, "owner@example.com", auth.RoleRestaurantOwner)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/"+otherRestaurantID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestRestaurantOwnerCanAccessOwnRestaurantHTTP(t *testing.T) {
	ownerID := uuid.New()
	ownedRestaurantID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				ownedRestaurantID: {ID: ownedRestaurantID, Name: "Owned"},
			},
		},
		&restaurants.MembershipMock{
			Members: map[uuid.UUID]map[uuid.UUID]bool{
				ownerID: {ownedRestaurantID: true},
			},
		},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(ownerID, "owner@example.com", auth.RoleRestaurantOwner)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/"+ownedRestaurantID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestAllRestaurantScopedRoutesDenyWithoutMembership(t *testing.T) {
	ownerID := uuid.New()
	restaurantID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				restaurantID: {ID: restaurantID, Name: "Foreign"},
			},
		},
		&restaurants.MembershipMock{},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(ownerID, "owner@example.com", auth.RoleRestaurantOwner)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/restaurants/" + restaurantID.String()},
		{http.MethodGet, "/api/v1/restaurants/" + restaurantID.String() + "/members"},
		{http.MethodPost, "/api/v1/restaurants/" + restaurantID.String() + "/members"},
		{http.MethodPost, "/api/v1/restaurants/" + restaurantID.String() + "/demo-sites"},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"user_id":"`+uuid.New().String()+`"}`))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
			}
		})
	}
}

func TestRestaurantOwnerListOnlyOwnedRestaurantsHTTP(t *testing.T) {
	ownerID := uuid.New()
	ownedRestaurantID := uuid.New()
	otherRestaurantID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				ownedRestaurantID: {ID: ownedRestaurantID, Name: "Owned"},
				otherRestaurantID: {ID: otherRestaurantID, Name: "Other"},
			},
		},
		&restaurants.MembershipMock{
			Members: map[uuid.UUID]map[uuid.UUID]bool{
				ownerID: {ownedRestaurantID: true},
			},
		},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(ownerID, "owner@example.com", auth.RoleRestaurantOwner)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, ownedRestaurantID.String()) {
		t.Fatalf("body = %s, want owned restaurant id", body)
	}
	if strings.Contains(body, otherRestaurantID.String()) {
		t.Fatalf("body = %s, should not include other restaurant id", body)
	}
}

func TestInternalAdminCanAccessAnyRestaurantHTTP(t *testing.T) {
	adminID := uuid.New()
	restaurantID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				restaurantID: {ID: restaurantID, Name: "Any"},
			},
		},
		&restaurants.MembershipMock{},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(adminID, "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/"+restaurantID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}
