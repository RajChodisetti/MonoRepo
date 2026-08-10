package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSEOPublicRateLimiterEnforcesPerIPAndGlobalWindows(t *testing.T) {
	now := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
	limiter := newSEOPublicRateLimiter()
	limiter.now = func() time.Time { return now }

	for request := 1; request <= seoRatePolicies["search"].perIPLimit; request++ {
		if retry := limiter.allow("search", "203.0.113.1"); retry != 0 {
			t.Fatalf("search request %d rejected with retry=%d", request, retry)
		}
	}
	if retry := limiter.allow("search", "203.0.113.1"); retry != 60 {
		t.Fatalf("per-IP retry = %d, want 60", retry)
	}
	now = now.Add(time.Minute)
	if retry := limiter.allow("search", "203.0.113.1"); retry != 0 {
		t.Fatalf("new window rejected with retry=%d", retry)
	}

	global := newSEOPublicRateLimiter()
	global.now = func() time.Time { return now }
	for request := 0; request < seoRatePolicies["report"].globalLimit; request++ {
		client := fmt.Sprintf("198.51.%d.%d", request/250, request%250+1)
		if retry := global.allow("report", client); retry != 0 {
			t.Fatalf("global report request %d rejected with retry=%d", request+1, retry)
		}
	}
	if retry := global.allow("report", "203.0.113.254"); retry != 60 {
		t.Fatalf("global retry = %d, want 60", retry)
	}
}

func TestSEORateLimiterAllowsNormalGalleryAndBoundsMemory(t *testing.T) {
	now := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
	limiter := newSEOPublicRateLimiter()
	limiter.now = func() time.Time { return now }
	for image := 1; image <= 20; image++ {
		if retry := limiter.allow("photo", "203.0.113.10"); retry != 0 {
			t.Fatalf("normal 20-image gallery rejected image %d with retry=%d", image, retry)
		}
	}

	old := now.Add(-time.Hour)
	for index := 0; index < maxSEOClientRateEntries; index++ {
		key := fmt.Sprintf("photo\x00old-%d", index)
		limiter.clients[key] = seoRateWindow{startedAt: old, lastSeen: old, count: 1}
	}
	if retry := limiter.allow("photo", "203.0.113.11"); retry != 0 {
		t.Fatalf("new client after eviction rejected with retry=%d", retry)
	}
	if len(limiter.clients) > maxSEOClientRateEntries {
		t.Fatalf("client map length = %d, cap = %d", len(limiter.clients), maxSEOClientRateEntries)
	}
}

func TestPublicSEOClientIPUsesTrustedEdgeRightmostAddress(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		forwarded  string
		realIP     string
		want       string
	}{
		{
			name:       "trusted private BFF forwards Caddy chain",
			remoteAddr: "172.18.0.4:4321",
			forwarded:  "192.0.2.20, 203.0.113.20",
			want:       "203.0.113.20",
		},
		{
			name:       "trusted loopback uses real IP fallback",
			remoteAddr: "127.0.0.1:4321",
			realIP:     "203.0.113.30",
			want:       "203.0.113.30",
		},
		{
			name:       "public peer cannot spoof forwarding header",
			remoteAddr: "198.51.100.40:4321",
			forwarded:  "203.0.113.99",
			want:       "198.51.100.40",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/public/v1/seo/search", nil)
			request.RemoteAddr = test.remoteAddr
			request.Header.Set("X-Forwarded-For", test.forwarded)
			request.Header.Set("X-Real-IP", test.realIP)
			if got := publicSEOClientIP(request); got != test.want {
				t.Fatalf("publicSEOClientIP() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSEORateLimitResponseHasRetryAfterAndNoStore(t *testing.T) {
	handler := newSEOUnlockHandlerForTest()
	for request := 0; request < seoRatePolicies["unlock_verify"].perIPLimit; request++ {
		if retry := handler.limiter.allow("unlock_verify", "192.0.2.1"); retry != 0 {
			t.Fatalf("setup request %d rejected", request+1)
		}
	}
	request := httptest.NewRequest(http.MethodPost, "/api/public/v1/seo/unlock/verify", strings.NewReader(`{}`))
	request.RemoteAddr = "192.0.2.1:1234"
	response := httptest.NewRecorder()
	handler.VerifyUnlock(response, request)

	if response.Code != http.StatusTooManyRequests || !strings.Contains(response.Body.String(), "seo_rate_limited") {
		t.Fatalf("status/body = %d/%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Retry-After") == "" {
		t.Fatal("rate-limited response omitted Retry-After")
	}
	assertSEOUnlockNoStore(t, response)
}
