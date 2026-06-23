package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	httpapi "github.com/rajchodisetti/restaurant-platform/backend/internal/http"
)

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	tokens := auth.NewTokenManager(testTokenSecret, time.Hour)
	handler := httpapi.RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthRejectsInvalidBearerScheme(t *testing.T) {
	tokens := auth.NewTokenManager(testTokenSecret, time.Hour)
	handler := httpapi.RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abc.def.ghi")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireRoleRejectsNonDeveloper(t *testing.T) {
	tokens := auth.NewTokenManager(testTokenSecret, time.Hour)
	token, _, err := tokens.IssueToken(uuid.New(), "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	handler := httpapi.RequireAuth(tokens)(httpapi.RequireRole(auth.RoleDeveloper)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRequireRoleAllowsDeveloper(t *testing.T) {
	tokens := auth.NewTokenManager(testTokenSecret, time.Hour)
	token, _, err := tokens.IssueToken(uuid.New(), "dev@example.com", auth.RoleDeveloper)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	handler := httpapi.RequireAuth(tokens)(httpapi.RequireRole(auth.RoleDeveloper)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok || principal.Role != auth.RoleDeveloper {
			t.Fatal("developer principal missing from context")
		}
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestRequireAuthSetsPrincipalInContext(t *testing.T) {
	tokens := auth.NewTokenManager(testTokenSecret, time.Hour)
	userID := uuid.New()
	token, _, err := tokens.IssueToken(userID, "dev@example.com", auth.RoleDeveloper)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	var gotEmail string
	handler := httpapi.RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("principal missing from context")
		}
		gotEmail = principal.Email
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotEmail != "dev@example.com" {
		t.Fatalf("email = %q, want dev@example.com", gotEmail)
	}
}
