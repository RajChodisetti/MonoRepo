package outreach

import "testing"

func TestRampedSendLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		maxLimit  int
		rampDay   int
		wantLimit int
	}{
		{name: "day zero normalizes to day one", maxLimit: 40, rampDay: 0, wantLimit: 5},
		{name: "day one", maxLimit: 40, rampDay: 1, wantLimit: 5},
		{name: "day two", maxLimit: 40, rampDay: 2, wantLimit: 10},
		{name: "day seven", maxLimit: 40, rampDay: 7, wantLimit: 35},
		{name: "day eight", maxLimit: 40, rampDay: 8, wantLimit: 40},
		{name: "stays at forty", maxLimit: 40, rampDay: 20, wantLimit: 40},
		{name: "respects lower configured cap", maxLimit: 17, rampDay: 4, wantLimit: 17},
		{name: "respects cap below first step", maxLimit: 3, rampDay: 1, wantLimit: 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := rampedSendLimit(test.maxLimit, test.rampDay); got != test.wantLimit {
				t.Fatalf("rampedSendLimit(%d, %d) = %d, want %d", test.maxLimit, test.rampDay, got, test.wantLimit)
			}
		})
	}
}

func TestNextEmailRampDayStopsAtConfiguredLimit(t *testing.T) {
	t.Parallel()

	if got := nextEmailRampDay(40, 7); got != 8 {
		t.Fatalf("nextEmailRampDay(40, 7) = %d, want 8", got)
	}
	if got := nextEmailRampDay(40, 8); got != 8 {
		t.Fatalf("nextEmailRampDay(40, 8) = %d, want 8", got)
	}
	if got := nextEmailRampDay(20, 4); got != 4 {
		t.Fatalf("nextEmailRampDay(20, 4) = %d, want 4", got)
	}
}
