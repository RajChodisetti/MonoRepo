package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	httpapi "github.com/rajchodisetti/restaurant-platform/backend/internal/http"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
)

func TestHealthzRequiresAuthorization(t *testing.T) {
	router := testRouter(t, fakeReadiness{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestHealthzReturnsServiceHealthForDeveloperToken(t *testing.T) {
	router := testRouter(t, fakeReadiness{})
	token := developerToken(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("X-Request-ID header is empty")
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("body = %s, want status ok", rec.Body.String())
	}
}

func TestHealthzForbiddenForNonDeveloperToken(t *testing.T) {
	router := testRouter(t, fakeReadiness{})
	tokens := auth.NewTokenManager(testTokenSecret, time.Hour)
	token, _, err := tokens.IssueToken(uuid.New(), "admin@example.com", auth.RoleInternalAdmin)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestReadyzReturnsOKWhenDatabasePings(t *testing.T) {
	router := testRouter(t, fakeReadiness{})
	token := developerToken(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestReadyzReturnsServiceUnavailableWhenDatabaseMissing(t *testing.T) {
	router := testRouter(t, fakeReadiness{err: db.ErrNotConfigured})
	token := developerToken(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(rec.Body.String(), "database_not_configured") {
		t.Fatalf("body = %s, want database_not_configured", rec.Body.String())
	}
}

func TestReadyzReturnsServiceUnavailableWhenDatabaseErrors(t *testing.T) {
	router := testRouter(t, fakeReadiness{err: errors.New("connection refused")})
	token := developerToken(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(rec.Body.String(), "database_unavailable") {
		t.Fatalf("body = %s, want database_unavailable", rec.Body.String())
	}
}

func TestSignupReturnsToken(t *testing.T) {
	userID := uuid.New()
	router := testRouterWithUserRepo(t, fakeReadiness{}, &auth.Mock{
		CreateFn: func(ctx context.Context, input auth.CreateInput) (auth.User, error) {
			return auth.User{
				ID:       userID,
				Email:    input.Email,
				FullName: input.FullName,
				Role:     input.Role,
				IsActive: true,
			}, nil
		},
	})

	body := `{"email":"dev@example.com","password":"password123","full_name":"Dev","role":"developer"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password_hash") {
		t.Fatalf("body leaks password hash: %s", rec.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response["access_token"] == "" {
		t.Fatal("access_token is empty")
	}
}

func TestLoginReturnsInvalidCredentials(t *testing.T) {
	router := testRouterWithUserRepo(t, fakeReadiness{}, &auth.Mock{})

	body := `{"email":"missing@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "invalid_credentials") {
		t.Fatalf("body = %s, want invalid_credentials", rec.Body.String())
	}
}

func TestRecoveryReturnsSafeInternalError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := httpapi.RequestID()(httpapi.Recovery(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("body exposes panic detail: %s", rec.Body.String())
	}
}
