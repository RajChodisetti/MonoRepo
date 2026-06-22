package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
)

func TestHealthzReturnsServiceHealth(t *testing.T) {
	router := testRouter(t, fakeReadiness{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
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

func TestReadyzReturnsOKWhenDatabasePings(t *testing.T) {
	router := testRouter(t, fakeReadiness{})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestReadyzReturnsServiceUnavailableWhenDatabaseMissing(t *testing.T) {
	router := testRouter(t, fakeReadiness{err: db.ErrNotConfigured})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(rec.Body.String(), "database_not_configured") {
		t.Fatalf("body = %s, want database_not_configured", rec.Body.String())
	}
}

func TestRecoveryReturnsSafeInternalError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := RequestID()(Recovery(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func testRouter(t *testing.T, readiness fakeReadiness) http.Handler {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	cfg.App.Env = config.EnvTest
	cfg.Logging.Format = "text"
	cfg.Logging.Level = "error"

	return NewRouter(logger.NewWithWriter(cfg.Logging, io.Discard), readiness, cfg)
}

type fakeReadiness struct {
	err error
}

func (f fakeReadiness) Ping(context.Context) error {
	if f.err != nil {
		return f.err
	}
	return nil
}

var _ ReadinessChecker = fakeReadiness{}

func TestReadyzReturnsServiceUnavailableWhenDatabaseErrors(t *testing.T) {
	router := testRouter(t, fakeReadiness{err: errors.New("connection refused")})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(rec.Body.String(), "database_unavailable") {
		t.Fatalf("body = %s, want database_unavailable", rec.Body.String())
	}
}
