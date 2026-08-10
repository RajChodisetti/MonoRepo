package handlers

import (
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxSEOClientRateEntries = 4096

type seoRatePolicy struct {
	perIPLimit   int
	perIPWindow  time.Duration
	globalLimit  int
	globalWindow time.Duration
}

var seoRatePolicies = map[string]seoRatePolicy{
	"search":         {perIPLimit: 30, perIPWindow: time.Minute, globalLimit: 600, globalWindow: time.Minute},
	"report":         {perIPLimit: 10, perIPWindow: time.Minute, globalLimit: 100, globalWindow: time.Minute},
	"photo":          {perIPLimit: 120, perIPWindow: time.Minute, globalLimit: 1200, globalWindow: time.Minute},
	"unlock_request": {perIPLimit: 10, perIPWindow: 10 * time.Minute, globalLimit: 120, globalWindow: time.Minute},
	"unlock_verify":  {perIPLimit: 30, perIPWindow: 10 * time.Minute, globalLimit: 600, globalWindow: time.Minute},
}

type seoRateWindow struct {
	startedAt time.Time
	lastSeen  time.Time
	count     int
}

// seoPublicRateLimiter is deliberately process-local defense in depth. The
// unlock repository separately enforces the email/place cooldown so horizontal
// API replicas cannot bypass the costliest pre-report guard.
type seoPublicRateLimiter struct {
	mu      sync.Mutex
	clients map[string]seoRateWindow
	global  map[string]seoRateWindow
	now     func() time.Time
}

func newSEOPublicRateLimiter() *seoPublicRateLimiter {
	return &seoPublicRateLimiter{
		clients: make(map[string]seoRateWindow),
		global:  make(map[string]seoRateWindow),
		now:     time.Now,
	}
}

// allow returns zero when accepted, otherwise the number of seconds callers
// should wait. Fixed windows keep this dependency-free and cheap enough for all
// public SEO routes; the client map has a hard cap and evicts its stalest key.
func (limiter *seoPublicRateLimiter) allow(route, client string) int {
	policy, ok := seoRatePolicies[route]
	if !ok {
		return 0
	}
	if limiter == nil {
		return 0
	}
	now := limiter.now().UTC()
	client = strings.TrimSpace(client)
	if client == "" {
		client = "unknown"
	}
	clientKey := route + "\x00" + client

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	global := currentSEOWindow(limiter.global[route], now, policy.globalWindow)
	if global.count >= policy.globalLimit {
		return retryAfterSeconds(global.startedAt, policy.globalWindow, now)
	}
	entry, exists := limiter.clients[clientKey]
	entry = currentSEOWindow(entry, now, policy.perIPWindow)
	if entry.count >= policy.perIPLimit {
		return retryAfterSeconds(entry.startedAt, policy.perIPWindow, now)
	}

	if !exists && len(limiter.clients) >= maxSEOClientRateEntries {
		for len(limiter.clients) >= maxSEOClientRateEntries {
			limiter.evictStalestClient()
		}
	}
	entry.count++
	entry.lastSeen = now
	limiter.clients[clientKey] = entry
	global.count++
	global.lastSeen = now
	limiter.global[route] = global
	return 0
}

func currentSEOWindow(window seoRateWindow, now time.Time, duration time.Duration) seoRateWindow {
	if window.startedAt.IsZero() || !now.Before(window.startedAt.Add(duration)) {
		return seoRateWindow{startedAt: now, lastSeen: now}
	}
	return window
}

func retryAfterSeconds(startedAt time.Time, duration time.Duration, now time.Time) int {
	remaining := startedAt.Add(duration).Sub(now)
	seconds := int((remaining + time.Second - 1) / time.Second)
	if seconds < 1 {
		return 1
	}
	return seconds
}

func (limiter *seoPublicRateLimiter) evictStalestClient() {
	var oldestKey string
	var oldest time.Time
	for key, entry := range limiter.clients {
		if oldestKey == "" || entry.lastSeen.Before(oldest) {
			oldestKey = key
			oldest = entry.lastSeen
		}
	}
	if oldestKey != "" {
		delete(limiter.clients, oldestKey)
	}
}

func (handler *SEOPublicHandler) allowPublicSEORequest(w http.ResponseWriter, r *http.Request, route string) bool {
	retryAfter := handler.limiter.allow(route, publicSEOClientIP(r))
	if retryAfter == 0 {
		return true
	}
	setNoStore(w)
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	handler.writeError(w, http.StatusTooManyRequests, "seo_rate_limited", "Too many requests. Please retry later.")
	return false
}

// publicSEOClientIP trusts forwarding headers only when the immediate peer is
// loopback/private (the production Caddy or web BFF hop). The rightmost valid
// forwarded address is the one appended by the trusted edge, so spoofed
// left-side entries cannot mint independent rate-limit identities.
func publicSEOClientIP(r *http.Request) string {
	peer := parseSEOIP(r.RemoteAddr)
	if peer.IsValid() && (peer.IsLoopback() || peer.IsPrivate()) {
		forwarded := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		for index := len(forwarded) - 1; index >= 0; index-- {
			if candidate := parseSEOIP(forwarded[index]); candidate.IsValid() {
				return candidate.String()
			}
		}
		if candidate := parseSEOIP(r.Header.Get("X-Real-IP")); candidate.IsValid() {
			return candidate.String()
		}
	}
	if peer.IsValid() {
		return peer.String()
	}
	return "unknown"
}

func parseSEOIP(raw string) netip.Addr {
	value := strings.TrimSpace(raw)
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	value = strings.Trim(value, "[]")
	address, err := netip.ParseAddr(value)
	if err != nil {
		return netip.Addr{}
	}
	return address.Unmap()
}
