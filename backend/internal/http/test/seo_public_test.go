package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSEOSearchEmptyQuery(t *testing.T) {
	router := testRouter(t, fakeReadiness{})
	req := httptest.NewRequest(http.MethodGet, "/api/public/v1/seo/search?q=", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	results, _ := payload["results"].([]any)
	if len(results) != 0 {
		t.Fatalf("expected empty results, got %#v", results)
	}
}

func TestSEOReportMissingPlace(t *testing.T) {
	router := testRouter(t, fakeReadiness{})
	req := httptest.NewRequest(http.MethodGet, "/api/public/v1/seo/report/missing-place-id-xyz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "not_found") {
		t.Fatalf("body = %s, want not_found", rec.Body.String())
	}
}

func TestSEOReportRejectsEmptyPlaceID(t *testing.T) {
	router := testRouter(t, fakeReadiness{})
	// Path with empty segment should not match; ServeMux returns 404.
	req := httptest.NewRequest(http.MethodGet, "/api/public/v1/seo/report/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}
