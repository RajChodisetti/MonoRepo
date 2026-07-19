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

func TestListRestaurantsFilterByRestaurantQueryParam(t *testing.T) {
	adminID := uuid.New()
	thaiID := uuid.New()
	indianID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				thaiID: {
					ID:     thaiID,
					Name:   "Thai Rama",
					Email:  "thai@example.com",
					Status: restaurants.StatusLead,
				},
				indianID: {
					ID:     indianID,
					Name:   "Indian Spice",
					Email:  "indian@example.com",
					Status: restaurants.StatusLead,
				},
			},
		},
		&restaurants.MembershipMock{},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(adminID, "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants?restaurant=thai", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Thai Rama") {
		t.Fatalf("body = %s, want Thai Rama", body)
	}
	if strings.Contains(body, "Indian Spice") {
		t.Fatalf("body = %s, should not include Indian Spice", body)
	}
}

func TestListRestaurantsFilterByDemoReadyStatus(t *testing.T) {
	adminID := uuid.New()
	demoReadyID := uuid.New()
	leadID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				demoReadyID: {
					ID:     demoReadyID,
					Name:   "Demo Ready Cafe",
					Email:  "demo@example.com",
					Status: restaurants.StatusDemoReady,
				},
				leadID: {
					ID:     leadID,
					Name:   "Lead Cafe",
					Email:  "lead@example.com",
					Status: restaurants.StatusLead,
				},
			},
		},
		&restaurants.MembershipMock{},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(adminID, "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants?status=demo_ready", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Demo Ready Cafe") {
		t.Fatalf("body = %s, want Demo Ready Cafe", body)
	}
	if strings.Contains(body, "Lead Cafe") {
		t.Fatalf("body = %s, should not include Lead Cafe", body)
	}
}

func TestListRestaurantsFilterByOCRVerifiedStatus(t *testing.T) {
	adminID := uuid.New()
	verifiedID := uuid.New()
	pendingID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				verifiedID: {
					ID:        verifiedID,
					Name:      "Verified Cafe",
					Email:     "verified@example.com",
					Status:    restaurants.StatusLead,
					OCRStatus: "verified",
				},
				pendingID: {
					ID:        pendingID,
					Name:      "Pending Cafe",
					Email:     "pending@example.com",
					Status:    restaurants.StatusLead,
					OCRStatus: "pending",
				},
			},
		},
		&restaurants.MembershipMock{},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(adminID, "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants?ocr_status=verified", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Verified Cafe") {
		t.Fatalf("body = %s, want Verified Cafe", body)
	}
	if strings.Contains(body, "Pending Cafe") {
		t.Fatalf("body = %s, should not include Pending Cafe", body)
	}
}

func TestListRestaurantsInvalidStatusReturnsBadRequest(t *testing.T) {
	adminID := uuid.New()
	router := testRouterWithStores(t, fakeReadiness{}, &auth.Mock{}, &restaurants.Mock{}, &restaurants.MembershipMock{}, &demos.Mock{})

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(adminID, "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants?status=invalid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPatchRestaurantUpdatesFields(t *testing.T) {
	adminID := uuid.New()
	restaurantID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				restaurantID: {
					ID:     restaurantID,
					Name:   "Old Name",
					Email:  "old@example.com",
					Status: restaurants.StatusLead,
				},
			},
		},
		&restaurants.MembershipMock{},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(adminID, "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/restaurants/"+restaurantID.String(), strings.NewReader(`{"name":"New Name","is_contacted":true}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"name":"New Name"`) || !strings.Contains(body, `"is_contacted":true`) {
		t.Fatalf("body = %s, want updated fields", body)
	}
	if !strings.Contains(body, `"status":"emailed"`) {
		t.Fatalf("body = %s, want status emailed after contact", body)
	}
}

func TestDeleteRestaurantSoftArchives(t *testing.T) {
	adminID := uuid.New()
	restaurantID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				restaurantID: {
					ID:     restaurantID,
					Name:   "Archive Me",
					Email:  "archive@example.com",
					Status: restaurants.StatusLead,
				},
			},
		},
		&restaurants.MembershipMock{},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(adminID, "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/restaurants/"+restaurantID.String(), nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want %d; body=%s", deleteRec.Code, http.StatusOK, deleteRec.Body.String())
	}
	if !strings.Contains(deleteRec.Body.String(), `"status":"archived"`) {
		t.Fatalf("delete body = %s, want archived status", deleteRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRec.Code, http.StatusOK)
	}
	if strings.Contains(listRec.Body.String(), restaurantID.String()) {
		t.Fatalf("archived restaurant should be hidden from default list: %s", listRec.Body.String())
	}
}

func TestOwnerListWithRestaurantFilterStillScoped(t *testing.T) {
	ownerID := uuid.New()
	ownedID := uuid.New()
	otherID := uuid.New()

	router := testRouterWithStores(
		t,
		fakeReadiness{},
		&auth.Mock{},
		&restaurants.Mock{
			Restaurants: map[uuid.UUID]restaurants.Restaurant{
				ownedID: {ID: ownedID, Name: "Owned Thai", Email: "owned@example.com", Status: restaurants.StatusLead},
				otherID: {ID: otherID, Name: "Other Thai", Email: "other@example.com", Status: restaurants.StatusLead},
			},
		},
		&restaurants.MembershipMock{
			Members: map[uuid.UUID]map[uuid.UUID]bool{
				ownerID: {ownedID: true},
			},
		},
		&demos.Mock{},
	)

	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(ownerID, "owner@example.com", auth.RoleRestaurantOwner)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/restaurants?restaurant=thai", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Owned Thai") {
		t.Fatalf("body = %s, want owned restaurant", body)
	}
	if strings.Contains(body, otherID.String()) {
		t.Fatalf("body should not include unassigned restaurant")
	}
}
