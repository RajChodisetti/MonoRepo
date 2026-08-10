package seoreport

import (
	"strings"
	"sync"
	"time"
)

const maxUnlockRateEntries = 4096

type unlockRateWindow struct {
	startedAt time.Time
	lastSeen  time.Time
	count     int
}

// unlockRequestLimiter bounds novel email/place combinations before the live
// report is generated. PostgreSQL remains the cross-process source of truth for
// the one-minute resend cooldown; this process-local layer limits broader spray
// patterns and has a hard memory cap.
type unlockRequestLimiter struct {
	mu         sync.Mutex
	emailPlace map[string]unlockRateWindow
	emails     map[string]unlockRateWindow
	global     unlockRateWindow
}

func newUnlockRequestLimiter() *unlockRequestLimiter {
	return &unlockRequestLimiter{
		emailPlace: make(map[string]unlockRateWindow),
		emails:     make(map[string]unlockRateWindow),
	}
}

func (limiter *unlockRequestLimiter) allow(now time.Time, email, placeID string) time.Duration {
	if limiter == nil {
		return time.Hour
	}
	now = now.UTC()
	email = strings.ToLower(strings.TrimSpace(email))
	placeID = sanitizePlaceID(placeID)
	pairKey := email + "\x00" + placeID

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	global := currentUnlockWindow(limiter.global, now, time.Minute)
	pair, pairExists := limiter.emailPlace[pairKey]
	pair = currentUnlockWindow(pair, now, time.Hour)
	emailWindow, emailExists := limiter.emails[email]
	emailWindow = currentUnlockWindow(emailWindow, now, time.Hour)
	retryAfter := time.Duration(0)
	if global.count >= 120 {
		retryAfter = maxUnlockRetry(retryAfter, global.startedAt.Add(time.Minute).Sub(now))
	}
	if pair.count >= 3 {
		retryAfter = maxUnlockRetry(retryAfter, pair.startedAt.Add(time.Hour).Sub(now))
	}
	if emailWindow.count >= 6 {
		retryAfter = maxUnlockRetry(retryAfter, emailWindow.startedAt.Add(time.Hour).Sub(now))
	}
	if retryAfter > 0 {
		return retryAfter
	}

	if !pairExists && len(limiter.emailPlace) >= maxUnlockRateEntries {
		for len(limiter.emailPlace) >= maxUnlockRateEntries {
			evictOldestUnlockWindow(limiter.emailPlace)
		}
	}
	if !emailExists && len(limiter.emails) >= maxUnlockRateEntries {
		for len(limiter.emails) >= maxUnlockRateEntries {
			evictOldestUnlockWindow(limiter.emails)
		}
	}
	pair.count++
	pair.lastSeen = now
	limiter.emailPlace[pairKey] = pair
	emailWindow.count++
	emailWindow.lastSeen = now
	limiter.emails[email] = emailWindow
	global.count++
	global.lastSeen = now
	limiter.global = global
	return 0
}

func maxUnlockRetry(current, candidate time.Duration) time.Duration {
	if candidate > current {
		return candidate
	}
	return current
}

func currentUnlockWindow(window unlockRateWindow, now time.Time, duration time.Duration) unlockRateWindow {
	if window.startedAt.IsZero() || !now.Before(window.startedAt.Add(duration)) {
		return unlockRateWindow{startedAt: now, lastSeen: now}
	}
	return window
}

func evictOldestUnlockWindow(entries map[string]unlockRateWindow) {
	var oldestKey string
	var oldest time.Time
	for key, entry := range entries {
		if oldestKey == "" || entry.lastSeen.Before(oldest) {
			oldestKey = key
			oldest = entry.lastSeen
		}
	}
	if oldestKey != "" {
		delete(entries, oldestKey)
	}
}
