package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/consultations"
)

func TestCompanyConsultationVoiceRoutesRequireStaticTokenAndAreDocumented(t *testing.T) {
	t.Setenv("TUVI_API_TOKEN", "test-tuvi-token")
	router := testRouter(t, fakeReadiness{})
	routes := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/company/consultations/availability"},
		{method: http.MethodGet, path: "/api/v1/company/consultations/availability/check"},
		{method: http.MethodPost, path: "/api/v1/company/consultations"},
		{method: http.MethodGet, path: "/api/v1/consultations/availability"},
		{method: http.MethodGet, path: "/api/v1/consultations/availability/check"},
		{method: http.MethodPost, path: "/api/v1/consultations"},
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() did not return this test file")
	}
	openAPIPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "docs", "openapi", "openapi.yaml")
	openAPI, err := os.ReadFile(openAPIPath)
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	contract := string(openAPI)
	if !strings.Contains(contract, "    consultationStaticBearer:") {
		t.Fatal("OpenAPI contract does not define consultationStaticBearer")
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("route status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
			}

			operation := fmt.Sprintf("  %s:\n    %s:", route.path, strings.ToLower(route.method))
			if !strings.Contains(contract, operation) {
				t.Fatalf("OpenAPI contract missing %s %s", route.method, route.path)
			}
		})
	}
}

func TestCompanyConsultationsRequireStaticBearerToken(t *testing.T) {
	t.Setenv("TUVI_API_TOKEN", "test-tuvi-token")
	router := testRouter(t, fakeReadiness{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/company/consultations/availability?date="+nextWeekday().Format("2006-01-02"), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unauthorized") {
		t.Fatalf("body = %s, want unauthorized", rec.Body.String())
	}
}

func TestCompanyConsultationsAvailability(t *testing.T) {
	t.Setenv("TUVI_API_TOKEN", "test-tuvi-token")
	router := testRouter(t, fakeReadiness{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/company/consultations/availability?date="+nextWeekday().Format("2006-01-02")+"&days=1", nil)
	req.Header.Set("Authorization", "Bearer test-tuvi-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"success"`) {
		t.Fatalf("body = %s, want success", rec.Body.String())
	}
}

func TestCompanyConsultationsCheckAndBookUseSharedDatabaseAvailability(t *testing.T) {
	t.Setenv("TUVI_API_TOKEN", "test-tuvi-token")
	repo := &consultations.Mock{}
	router := testRouterWithConsultationRepo(t, fakeReadiness{}, repo)
	date := futureConsultationDate(t)
	checkPath := "/api/v1/company/consultations/availability/check?date=" + date + "&time=09:00"

	checkReq := httptest.NewRequest(http.MethodGet, checkPath, nil)
	checkReq.Header.Set("Authorization", "Bearer test-tuvi-token")
	checkRec := httptest.NewRecorder()
	router.ServeHTTP(checkRec, checkReq)
	if checkRec.Code != http.StatusOK || !strings.Contains(checkRec.Body.String(), `"available":true`) {
		t.Fatalf("initial check status=%d body=%s, want available", checkRec.Code, checkRec.Body.String())
	}

	bookBody := strings.NewReader(`{
		"date":"` + date + `",
		"time":"09:00",
		"prospect_name":"Voice Route Test",
		"prospect_email":"voice-route@example.test",
		"source":"voice"
	}`)
	bookReq := httptest.NewRequest(http.MethodPost, "/api/v1/company/consultations", bookBody)
	bookReq.Header.Set("Authorization", "Bearer test-tuvi-token")
	bookReq.Header.Set("Content-Type", "application/json")
	bookRec := httptest.NewRecorder()
	router.ServeHTTP(bookRec, bookReq)
	if bookRec.Code != http.StatusCreated || !strings.Contains(bookRec.Body.String(), `"status":"success"`) {
		t.Fatalf("book status=%d body=%s, want confirmed booking", bookRec.Code, bookRec.Body.String())
	}
	if len(repo.Inserted) != 1 {
		t.Fatalf("inserted consultations = %d, want 1", len(repo.Inserted))
	}

	checkAgainReq := httptest.NewRequest(http.MethodGet, checkPath, nil)
	checkAgainReq.Header.Set("Authorization", "Bearer test-tuvi-token")
	checkAgainRec := httptest.NewRecorder()
	router.ServeHTTP(checkAgainRec, checkAgainReq)
	if checkAgainRec.Code != http.StatusOK || !strings.Contains(checkAgainRec.Body.String(), `"available":false`) {
		t.Fatalf("post-booking check status=%d body=%s, want unavailable", checkAgainRec.Code, checkAgainRec.Body.String())
	}
}

func TestLegacyConsultationsAvailabilityAlias(t *testing.T) {
	t.Setenv("TUVI_API_TOKEN", "test-tuvi-token")
	router := testRouter(t, fakeReadiness{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/consultations/availability?date="+nextWeekday().Format("2006-01-02")+"&days=1", nil)
	req.Header.Set("Authorization", "Bearer test-tuvi-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestConsultationCalendarRequiresInternalAdmin(t *testing.T) {
	router := testRouterWithConsultationRepo(t, fakeReadiness{}, &consultations.Mock{})
	month := futureCalendarMonth(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/consultation-calendar/"+month, nil)
	req.Header.Set("Authorization", "Bearer "+userToken(t, auth.RoleRestaurantOwner))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestConsultationCalendarRequiresAuthentication(t *testing.T) {
	router := testRouterWithConsultationRepo(t, fakeReadiness{}, &consultations.Mock{})
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/consultation-calendar/"+futureCalendarMonth(t),
		nil,
	)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestConsultationCalendarGetAndPut(t *testing.T) {
	repo := &consultations.Mock{}
	router := testRouterWithConsultationRepo(t, fakeReadiness{}, repo)
	token := userToken(t, auth.RoleInternalAdmin)
	month := futureCalendarMonth(t)
	path := "/api/v1/admin/consultation-calendar/" + month

	getReq := httptest.NewRequest(http.MethodGet, path, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d; body=%s", getRec.Code, http.StatusOK, getRec.Body.String())
	}

	var calendar consultations.CalendarResult
	if err := json.Unmarshal(getRec.Body.Bytes(), &calendar); err != nil {
		t.Fatalf("decode GET calendar: %v", err)
	}
	updates := make([]consultations.CalendarSlotUpdate, 0, len(calendar.Slots))
	for _, slot := range calendar.Slots {
		if slot.Past || !slot.OnGrid {
			continue
		}
		updates = append(updates, consultations.CalendarSlotUpdate{
			ISO:         slot.ISO,
			IsAvailable: true,
		})
	}
	if len(updates) == 0 {
		t.Fatal("GET calendar returned no future on-grid slots")
	}
	updates[0].IsAvailable = false
	body, err := json.Marshal(struct {
		ExpectedRevision int64                              `json:"expected_revision"`
		Slots            []consultations.CalendarSlotUpdate `json:"slots"`
	}{ExpectedRevision: calendar.Revision, Slots: updates})
	if err != nil {
		t.Fatalf("marshal PUT calendar: %v", err)
	}

	putReq := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(body))
	putReq.Header.Set("Authorization", "Bearer "+token)
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want %d; body=%s", putRec.Code, http.StatusOK, putRec.Body.String())
	}
	if !strings.Contains(putRec.Body.String(), `"effective_available":false`) {
		t.Fatalf("PUT body = %s, want disabled effective availability", putRec.Body.String())
	}
	var saved consultations.CalendarResult
	if err := json.Unmarshal(putRec.Body.Bytes(), &saved); err != nil {
		t.Fatalf("decode PUT calendar: %v", err)
	}
	if saved.Revision != calendar.Revision+1 {
		t.Fatalf("PUT revision = %d, want %d", saved.Revision, calendar.Revision+1)
	}

	staleReq := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(body))
	staleReq.Header.Set("Authorization", "Bearer "+token)
	staleReq.Header.Set("Content-Type", "application/json")
	staleRec := httptest.NewRecorder()
	router.ServeHTTP(staleRec, staleReq)
	if staleRec.Code != http.StatusConflict {
		t.Fatalf("stale PUT status = %d, want %d; body=%s", staleRec.Code, http.StatusConflict, staleRec.Body.String())
	}
	if !strings.Contains(staleRec.Body.String(), `"error":{"code":"consultation_calendar_conflict"`) {
		t.Fatalf("stale PUT body = %s, want standard conflict error", staleRec.Body.String())
	}
}

func TestConsultationCalendarPutRequiresRevisionAndBoundsBody(t *testing.T) {
	router := testRouterWithConsultationRepo(t, fakeReadiness{}, &consultations.Mock{})
	token := userToken(t, auth.RoleInternalAdmin)
	path := "/api/v1/admin/consultation-calendar/" + futureCalendarMonth(t)

	missingReq := httptest.NewRequest(http.MethodPut, path, strings.NewReader(`{"slots":[]}`))
	missingReq.Header.Set("Authorization", "Bearer "+token)
	missingRec := httptest.NewRecorder()
	router.ServeHTTP(missingRec, missingReq)
	if missingRec.Code != http.StatusBadRequest {
		t.Fatalf("missing revision status = %d, want %d; body=%s", missingRec.Code, http.StatusBadRequest, missingRec.Body.String())
	}

	oversized := `{"expected_revision":0,"slots":[],"padding":"` + strings.Repeat("x", (1<<20)+1) + `"}`
	largeReq := httptest.NewRequest(http.MethodPut, path, strings.NewReader(oversized))
	largeReq.Header.Set("Authorization", "Bearer "+token)
	largeRec := httptest.NewRecorder()
	router.ServeHTTP(largeRec, largeReq)
	if largeRec.Code != http.StatusBadRequest {
		t.Fatalf("oversized body status = %d, want %d; body=%s", largeRec.Code, http.StatusBadRequest, largeRec.Body.String())
	}
}

func TestConsultationCalendarUsesStandardAdminErrorShape(t *testing.T) {
	router := testRouterWithConsultationRepo(t, fakeReadiness{}, &consultations.Mock{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/consultation-calendar/not-a-month", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(t, auth.RoleInternalAdmin))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"error":{"code":"invalid_consultation_calendar"`) {
		t.Fatalf("body = %s, want standard admin error", rec.Body.String())
	}
}

func nextWeekday() time.Time {
	now := time.Now()
	date := now.AddDate(0, 0, 1)
	for date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

func futureCalendarMonth(t *testing.T) string {
	t.Helper()
	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		t.Fatalf("load consultation timezone: %v", err)
	}
	return time.Now().In(loc).AddDate(0, 2, 0).Format("2006-01")
}

func futureConsultationDate(t *testing.T) string {
	t.Helper()
	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		t.Fatalf("load consultation timezone: %v", err)
	}
	date := time.Now().In(loc).AddDate(0, 0, 2)
	for date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, 1)
	}
	return date.Format("2006-01-02")
}
