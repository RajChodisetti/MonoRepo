package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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

func nextWeekday() time.Time {
	now := time.Now()
	date := now.AddDate(0, 0, 1)
	for date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, 1)
	}
	return date
}
