package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/seoreport"
)

func newSEOUnlockHandlerForTest() *SEOPublicHandler {
	return NewSEOPublicHandler(
		&seoreport.Service{},
		"https://www.example.test",
		func(w http.ResponseWriter, status int, value any) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(value)
		},
		func(w http.ResponseWriter, status int, code, message string) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{"code": code, "message": message},
			})
		},
	)
}

func assertSEOUnlockNoStore(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
	if got := response.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("Pragma = %q, want no-cache", got)
	}
}

func TestSEOUnlockRequestRequiresNameEmailPhoneAndPlace(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode string
	}{
		{
			name:     "name",
			body:     `{"email":"owner@example.com","phone":"+61 2 9123 4567","placeId":"place-1"}`,
			wantCode: "invalid_name",
		},
		{
			name:     "email",
			body:     `{"name":"Sam Owner","email":"bad","phone":"+61 2 9123 4567","placeId":"place-1"}`,
			wantCode: "invalid_email",
		},
		{
			name:     "phone",
			body:     `{"name":"Sam Owner","email":"owner@example.com","placeId":"place-1"}`,
			wantCode: "invalid_phone",
		},
		{
			name:     "place",
			body:     `{"name":"Sam Owner","email":"owner@example.com","phone":"+61 2 9123 4567","placeId":""}`,
			wantCode: "not_found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/public/v1/seo/unlock/request", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			newSEOUnlockHandlerForTest().RequestUnlock(response, request)

			if response.Code != http.StatusBadRequest && response.Code != http.StatusNotFound {
				t.Fatalf("status = %d; body=%s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), test.wantCode) {
				t.Fatalf("body = %s, want code %q", response.Body.String(), test.wantCode)
			}
			assertSEOUnlockNoStore(t, response)
		})
	}
}

func TestSEOUnlockRequestRejectsUnknownTrailingAndOversizeJSON(t *testing.T) {
	valid := `{"name":"Sam Owner","email":"owner@example.com","phone":"+61 2 9123 4567","placeId":"place-1"}`
	tests := []struct {
		name string
		body string
	}{
		{name: "unknown field", body: strings.TrimSuffix(valid, "}") + `,"marketingConsent":true}`},
		{name: "trailing value", body: valid + ` {}`},
		{name: "oversize", body: `{"name":"` + strings.Repeat("a", maxSEOUnlockBodyBytes) + `"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/public/v1/seo/unlock/request", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			newSEOUnlockHandlerForTest().RequestUnlock(response, request)

			if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "invalid_body") {
				t.Fatalf("status/body = %d/%s", response.Code, response.Body.String())
			}
			assertSEOUnlockNoStore(t, response)
		})
	}
}

func TestSEOUnlockVerifyUsesSeparateStrictBodyAndSixDigitOTP(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode string
		status   int
	}{
		{
			name:     "request-only field rejected",
			body:     `{"name":"Sam Owner","email":"owner@example.com","placeId":"place-1","otp":"123456"}`,
			wantCode: "invalid_body",
			status:   http.StatusBadRequest,
		},
		{
			name:     "four digit OTP rejected",
			body:     `{"email":"owner@example.com","placeId":"place-1","otp":"1234"}`,
			wantCode: "invalid_otp",
			status:   http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/public/v1/seo/unlock/verify", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			newSEOUnlockHandlerForTest().VerifyUnlock(response, request)

			if response.Code != test.status || !strings.Contains(response.Body.String(), test.wantCode) {
				t.Fatalf("status/body = %d/%s", response.Code, response.Body.String())
			}
			assertSEOUnlockNoStore(t, response)
		})
	}
}

func TestSEOUnlockRateLimitResponse(t *testing.T) {
	response := httptest.NewRecorder()
	newSEOUnlockHandlerForTest().writeUnlockError(response, seoreport.ErrUnlockRateLimit)

	if response.Code != http.StatusTooManyRequests || !strings.Contains(response.Body.String(), "unlock_rate_limited") {
		t.Fatalf("status/body = %d/%s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Retry-After"); got != "60" {
		t.Fatalf("Retry-After = %q, want 60", got)
	}
}

func TestSEOUnlockMagicLinkResponseIsNotCacheable(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/public/v1/seo/unlock/click/token", nil)
	request.SetPathValue("token", "token")
	response := httptest.NewRecorder()
	newSEOUnlockHandlerForTest().ClickUnlock(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want controlled unavailable response; body=%s", response.Code, response.Body.String())
	}
	assertSEOUnlockNoStore(t, response)
}
