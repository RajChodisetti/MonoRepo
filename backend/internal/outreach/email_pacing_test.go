package outreach

import (
	"testing"
	"time"
)

func TestEmailWindowAtUsesSydneyClockAndExclusiveNoonBoundary(t *testing.T) {
	t.Parallel()
	schedule := defaultEmailSendSchedule()

	tests := []struct {
		name          string
		localNow      time.Time
		open          bool
		wantStartDate string
	}{
		{
			name:          "before window",
			localNow:      time.Date(2026, time.August, 15, 6, 59, 0, 0, scheduledSendLocation),
			open:          false,
			wantStartDate: "2026-08-15 07:00 AEST",
		},
		{
			name:          "start inclusive",
			localNow:      time.Date(2026, time.August, 15, 7, 0, 0, 0, scheduledSendLocation),
			open:          true,
			wantStartDate: "2026-08-15 07:00 AEST",
		},
		{
			name:          "before noon",
			localNow:      time.Date(2026, time.August, 15, 11, 59, 59, 0, scheduledSendLocation),
			open:          true,
			wantStartDate: "2026-08-15 07:00 AEST",
		},
		{
			name:          "noon exclusive",
			localNow:      time.Date(2026, time.August, 15, 12, 0, 0, 0, scheduledSendLocation),
			open:          false,
			wantStartDate: "2026-08-16 07:00 AEST",
		},
		{
			name:          "daylight saving",
			localNow:      time.Date(2026, time.December, 15, 8, 0, 0, 0, scheduledSendLocation),
			open:          true,
			wantStartDate: "2026-12-15 07:00 AEDT",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			window := emailWindowAt(test.localNow.UTC(), schedule)
			if window.Open != test.open {
				t.Fatalf("Open = %v, want %v", window.Open, test.open)
			}
			if got := window.Start.In(scheduledSendLocation).Format("2006-01-02 15:04 MST"); got != test.wantStartDate {
				t.Fatalf("start = %s, want %s", got, test.wantStartDate)
			}
			if got := window.End.Sub(window.Start); got != 5*time.Hour {
				t.Fatalf("window duration = %s, want %s", got, 5*time.Hour)
			}
		})
	}
}

func TestScheduledDailyAccountSendAtDistributesFortySlotsBeforeNoon(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.August, 15, 7, 0, 0, 0, scheduledSendLocation).UTC()
	windowEnd := time.Date(2026, time.August, 15, 12, 0, 0, 0, scheduledSendLocation).UTC()
	const sendLimit = 40
	slotWidth := 5 * time.Hour / sendLimit
	previous := windowStart

	for usedSlots := 1; usedSlots < sendLimit; usedSlots++ {
		next, err := scheduledDailyAccountSendAt(
			previous,
			windowStart,
			windowEnd,
			usedSlots,
			sendLimit,
		)
		if err != nil {
			t.Fatalf("scheduledDailyAccountSendAt(%d) error = %v", usedSlots, err)
		}
		want := windowStart.Add(time.Duration(usedSlots) * slotWidth)
		if !next.Equal(want) {
			t.Fatalf("slot %d = %s, want %s", usedSlots+1, next, want)
		}
		if !next.After(previous) {
			t.Fatalf("slot %d = %s, previous = %s", usedSlots+1, next, previous)
		}
		previous = next
	}

	if !previous.Before(windowEnd) {
		t.Fatalf("last slot = %s, want before window end %s", previous, windowEnd)
	}
}

func TestNextSydneyWindowKeepsSevenAMAcrossDaylightSavingStart(t *testing.T) {
	t.Parallel()

	currentEnd := time.Date(2026, time.October, 3, 12, 0, 0, 0, scheduledSendLocation).UTC()
	nextStart := emailWindowAt(currentEnd, defaultEmailSendSchedule()).Start.In(scheduledSendLocation)
	if got := nextStart.Format("2006-01-02 15:04 MST"); got != "2026-10-04 07:00 AEDT" {
		t.Fatalf("next Sydney window = %s", got)
	}
}

func TestScheduledDailyAccountSendAtUsesRemainingWindowAfterDelay(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.August, 15, 7, 0, 0, 0, scheduledSendLocation).UTC()
	windowEnd := time.Date(2026, time.August, 15, 12, 0, 0, 0, scheduledSendLocation).UTC()
	now := time.Date(2026, time.August, 15, 10, 0, 0, 0, scheduledSendLocation).UTC()

	next, err := scheduledDailyAccountSendAt(now, windowStart, windowEnd, 20, 40)
	if err != nil {
		t.Fatalf("scheduledDailyAccountSendAt() error = %v", err)
	}
	want := now.Add(windowEnd.Sub(now) / 21)
	if !next.Equal(want) {
		t.Fatalf("next = %s, want remaining-window slot %s", next, want)
	}
	if !next.Before(windowEnd) {
		t.Fatalf("next = %s, want before noon %s", next, windowEnd)
	}
}

func TestRandomPacingJitterForDailyCapacityCapsAggregateDelay(t *testing.T) {
	t.Parallel()

	for range 100 {
		got, err := randomPacingJitterForDailyCapacity(
			2*time.Minute,
			5*time.Minute,
			5*time.Hour,
			120,
			5*time.Hour,
			120,
		)
		if err != nil {
			t.Fatalf("randomPacingJitterForDailyCapacity() error = %v", err)
		}
		if got < 2*time.Minute || got > 150*time.Second {
			t.Fatalf("jitter = %s, want 2m..2m30s", got)
		}
	}
}

func TestRandomPacingJitterForDailyCapacityTightensAfterLateStart(t *testing.T) {
	t.Parallel()

	for range 100 {
		got, err := randomPacingJitterForDailyCapacity(
			2*time.Minute,
			5*time.Minute,
			5*time.Hour,
			120,
			2*time.Hour,
			60,
		)
		if err != nil {
			t.Fatalf("randomPacingJitterForDailyCapacity() error = %v", err)
		}
		if got < 2*time.Minute || got > 2*time.Hour/59 {
			t.Fatalf("jitter = %s, want remaining-window capacity", got)
		}
	}
}

func TestRandomPacingJitterForDailyCapacityRejectsQuotaThatCannotFit(t *testing.T) {
	t.Parallel()

	_, err := randomPacingJitterForDailyCapacity(
		2*time.Minute,
		5*time.Minute,
		5*time.Hour,
		160,
		5*time.Hour,
		160,
	)
	if err == nil {
		t.Fatal("randomPacingJitterForDailyCapacity() error = nil, want capacity error")
	}
}

func TestScheduledDailyAccountSendAtRejectsOutsideWindow(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, time.August, 15, 7, 0, 0, 0, scheduledSendLocation).UTC()
	windowEnd := time.Date(2026, time.August, 15, 12, 0, 0, 0, scheduledSendLocation).UTC()
	_, err := scheduledDailyAccountSendAt(windowEnd, windowStart, windowEnd, 1, 40)
	if err == nil {
		t.Fatal("scheduledDailyAccountSendAt() error = nil, want outside-window error")
	}
}
