package outreach

import (
	"errors"
	"testing"
	"time"
)

func TestValidateEmailSendSchedule(t *testing.T) {
	t.Parallel()

	start, end, err := validateEmailSendSchedule(UpdateEmailSendScheduleInput{
		StartTime: "08:15",
		EndTime:   "13:45",
	})
	if err != nil {
		t.Fatalf("validateEmailSendSchedule() error = %v", err)
	}
	if start != 8*60+15 || end != 13*60+45 {
		t.Fatalf("minutes = %d-%d, want 495-825", start, end)
	}
}

func TestValidateEmailSendScheduleRejectsInvalidOrOvernightWindows(t *testing.T) {
	t.Parallel()

	for _, input := range []UpdateEmailSendScheduleInput{
		{StartTime: "7:00", EndTime: "12:00"},
		{StartTime: "12:00", EndTime: "07:00"},
		{StartTime: "07:00", EndTime: "07:30"},
	} {
		if _, _, err := validateEmailSendSchedule(input); !errors.Is(err, ErrInvalidSendSchedule) {
			t.Fatalf("validateEmailSendSchedule(%#v) error = %v, want ErrInvalidSendSchedule", input, err)
		}
	}
}

func TestEmailWindowAtUsesSavedMinutePrecision(t *testing.T) {
	t.Parallel()

	schedule := newStoredEmailSendSchedule(8*60+15, 13*60+45, nil, time.Time{})
	inside := time.Date(2026, time.December, 15, 8, 15, 0, 0, scheduledSendLocation)
	window := emailWindowAt(inside.UTC(), schedule)
	if !window.Open {
		t.Fatal("saved window is closed at its inclusive start")
	}
	if got := window.Start.In(scheduledSendLocation).Format("2006-01-02 15:04 MST"); got != "2026-12-15 08:15 AEDT" {
		t.Fatalf("start = %s", got)
	}
	if got := window.End.In(scheduledSendLocation).Format("2006-01-02 15:04 MST"); got != "2026-12-15 13:45 AEDT" {
		t.Fatalf("end = %s", got)
	}
}

func TestValidateEmailScheduleCapacity(t *testing.T) {
	t.Parallel()

	if err := validateEmailScheduleCapacity(5*time.Hour, 120, 2*time.Minute); err != nil {
		t.Fatalf("five-hour capacity error = %v", err)
	}
	if err := validateEmailScheduleCapacity(3*time.Hour, 120, 2*time.Minute); !errors.Is(err, ErrInvalidSendSchedule) {
		t.Fatalf("three-hour capacity error = %v, want ErrInvalidSendSchedule", err)
	}
}
