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

func TestUserMeRequiresAuthorization(t *testing.T) {
	router := testRouter(t, fakeReadiness{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestUserMeForbiddenForInternalAdminToken(t *testing.T) {
	userID := uuid.New()
	repo := &auth.Mock{
		Users: map[uuid.UUID]auth.User{
			userID: {
				ID:       userID,
				Email:    "admin@example.com",
				Role:     auth.RoleInternalAdmin,
				IsActive: true,
			},
		},
	}
	router := testRouterWithUserRepo(t, fakeReadiness{}, repo)
	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(userID, "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestUserMeReturnsProfileForRestaurantOwnerToken(t *testing.T) {
	userID := uuid.New()
	restaurantID := uuid.New()
	repo := &auth.Mock{
		Users: map[uuid.UUID]auth.User{
			userID: {
				ID:       userID,
				Email:    "owner@example.com",
				FullName: "Restaurant Owner",
				Role:     auth.RoleRestaurantOwner,
				IsActive: true,
			},
		},
	}
	router := testRouterWithStores(
		t,
		fakeReadiness{},
		repo,
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				restaurantID: {
					ID:     restaurantID,
					Name:   "Demo Cafe",
					Email:  "cafe@example.com",
					Status: restaurants.StatusLead,
				},
			},
		},
		&restaurants.MembershipMock{
			Members: map[uuid.UUID]map[uuid.UUID]bool{
				userID: {restaurantID: true},
			},
		},
		&demos.Mock{},
	)
	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(userID, "owner@example.com", auth.RoleRestaurantOwner)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	body := rec.Body.String()
	for _, want := range []string{
		`"role":"restaurant_owner"`,
		`"full_name":"Restaurant Owner"`,
		`"email":"owner@example.com"`,
		`"Demo Cafe"`,
		`"member_role":"owner"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %s, want substring %s", body, want)
		}
	}
}
