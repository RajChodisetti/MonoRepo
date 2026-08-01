package httpapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

func TestDeveloperSchemaRequiresInternalAdmin(t *testing.T) {
	router := testRouter(t, fakeReadiness{})
	token := developerToken(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/developer/schema", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestDeveloperSchemaReportsDatabaseUnavailable(t *testing.T) {
	router := testRouter(t, fakeReadiness{})
	token := internalAdminToken(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/developer/schema", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusServiceUnavailable, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "database_unavailable") {
		t.Fatalf("body = %s, want database_unavailable", rec.Body.String())
	}
}

func TestDeveloperSQLRejectsMutationBeforeDatabaseAccess(t *testing.T) {
	router := testRouter(t, fakeReadiness{})
	token := internalAdminToken(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/developer/sql", bytes.NewBufferString(`{"query":"delete from restaurants"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "read_only_required") {
		t.Fatalf("body = %s, want read_only_required", rec.Body.String())
	}
}

func internalAdminToken(t *testing.T) string {
	t.Helper()
	token, _, err := auth.NewTokenManager(testTokenSecret, time.Hour).IssueToken(uuid.New(), "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	return token
}
