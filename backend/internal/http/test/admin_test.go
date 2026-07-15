package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

func TestAdminMeRequiresAuthorization(t *testing.T) {
	router := testRouter(t, fakeReadiness{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAdminMeForbiddenForRestaurantOwnerToken(t *testing.T) {
	userID := uuid.New()
	repo := &auth.Mock{
		Users: map[uuid.UUID]auth.User{
			userID: {
				ID:       userID,
				Email:    "owner@example.com",
				Role:     auth.RoleRestaurantOwner,
				IsActive: true,
			},
		},
	}
	router := testRouterWithUserRepo(t, fakeReadiness{}, repo)
	token := userToken(t, auth.RoleRestaurantOwner)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestAdminMeReturnsProfileForInternalAdminToken(t *testing.T) {
	userID := uuid.New()
	repo := &auth.Mock{
		Users: map[uuid.UUID]auth.User{
			userID: {
				ID:       userID,
				Email:    "admin@example.com",
				FullName: "Platform Admin",
				Role:     auth.RoleInternalAdmin,
				IsActive: true,
			},
		},
	}
	router := testRouterWithStores(t, fakeReadiness{}, repo, &restaurants.Mock{}, &restaurants.MembershipMock{}, &demos.Mock{}, &campaigns.Mock{})
	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(userID, "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"role":"internal_admin"`) {
		t.Fatalf("body = %s, want internal_admin role", rec.Body.String())
	}
}

func TestSignupRejectsInternalAdminRole(t *testing.T) {
	router := testRouterWithStores(t, fakeReadiness{}, &auth.Mock{}, &restaurants.Mock{}, &restaurants.MembershipMock{}, &demos.Mock{}, &campaigns.Mock{})

	body := `{"email":"admin@example.com","password":"password123","full_name":"Admin","role":"internal_admin"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func userToken(t *testing.T, role string) string {
	t.Helper()
	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(uuid.New(), "user@example.com", role)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	return token
}
