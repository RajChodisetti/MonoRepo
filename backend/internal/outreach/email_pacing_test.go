package outreach

import (
	"testing"
	"time"
)

func TestScheduledAccountSendAtDistributesFortySlotsAcrossEightHours(t *testing.T) {
	t.Parallel()

	cycleStart := time.Date(2026, time.July, 14, 9, 0, 0, 0, time.UTC)
	const sendLimit = 40
	const window = 8 * time.Hour
	const jitter = 3 * time.Minute
	slotWidth := window / sendLimit
	previous := cycleStart

	for usedSlots := 1; usedSlots < sendLimit; usedSlots++ {
		schedule, err := scheduledAccountSendAt(
			previous,
			cycleStart,
			usedSlots,
			sendLimit,
			window,
			jitter,
		)
		if err != nil {
			t.Fatalf("scheduledAccountSendAt(%d) error = %v", usedSlots, err)
		}
		next := schedule.NextSendAt
		want := cycleStart.Add(time.Duration(usedSlots)*slotWidth + jitter)
		if !next.Equal(want) {
			t.Fatalf("slot %d = %s, want %s", usedSlots+1, next, want)
		}
		if !next.After(previous) {
			t.Fatalf("slot %d = %s, previous = %s", usedSlots+1, next, previous)
		}
		previous = next
	}

	if !previous.Before(cycleStart.Add(window)) {
		t.Fatalf("last slot = %s, want before window end %s", previous, cycleStart.Add(window))
	}
}

func TestScheduledAccountSendAtPreventsCatchUpBurst(t *testing.T) {
	t.Parallel()

	cycleStart := time.Date(2026, time.July, 14, 9, 0, 0, 0, time.UTC)
	now := cycleStart.Add(6 * time.Hour)
	jitter := 4 * time.Minute

	schedule, err := scheduledAccountSendAt(now, cycleStart, 2, 40, 8*time.Hour, jitter)
	if err != nil {
		t.Fatalf("scheduledAccountSendAt() error = %v", err)
	}
	if want := now.Add(12*time.Minute + jitter); !schedule.NextSendAt.Equal(want) {
		t.Fatalf("next = %s, want restart-safe slot %s", schedule.NextSendAt, want)
	}
	if want := now.Add(-12 * time.Minute); !schedule.CycleStartedAt.Equal(want) {
		t.Fatalf("cycle start = %s, want re-anchored %s", schedule.CycleStartedAt, want)
	}
}

func TestRandomPacingJitterStaysWithinConfiguredBounds(t *testing.T) {
	t.Parallel()

	for range 100 {
		got, err := randomPacingJitter(2*time.Minute, 5*time.Minute)
		if err != nil {
			t.Fatalf("randomPacingJitter() error = %v", err)
		}
		if got < 2*time.Minute || got > 5*time.Minute {
			t.Fatalf("randomPacingJitter() = %s, want 2m..5m", got)
		}
		if got%time.Second != 0 {
			t.Fatalf("randomPacingJitter() = %s, want whole seconds", got)
		}
	}
}

func TestScheduledAccountSendAtRejectsJitterWiderThanSlot(t *testing.T) {
	t.Parallel()

	_, err := scheduledAccountSendAt(time.Now(), time.Now(), 1, 40, 8*time.Hour, 12*time.Minute)
	if err == nil {
		t.Fatal("scheduledAccountSendAt() error = nil, want slot-width validation error")
	}
}
