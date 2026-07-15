package consultations

import (
	"context"
	"testing"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func TestAvailabilityExcludesConfirmedDatabaseSlot(t *testing.T) {
	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}

	date := nextWeekday(time.Now().In(loc).AddDate(0, 0, 2))
	blocked := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, loc)
	repo := &Mock{Booked: map[time.Time]struct{}{blocked: {}}}
	service := NewService(config.ConsultationConfig{
		Timezone:            loc,
		BusinessHourStart:   9,
		BusinessHourEnd:     17,
		SlotDurationMinutes: 30,
		DefaultAvailDays:    5,
		AvailabilityHorizon: 14,
	}, repo, nil, nil)

	result, err := service.GetAvailability(context.Background(), date.Format("2006-01-02"), 1)
	if err != nil {
		t.Fatalf("GetAvailability() error = %v", err)
	}
	for _, slot := range result.Slots {
		if slot.ISO == blocked.Format(time.RFC3339) {
			t.Fatalf("confirmed database slot %s was returned as available", slot.ISO)
		}
	}
	if len(result.Slots) != 15 {
		t.Fatalf("available slots = %d, want 15 after blocking one 30-minute slot", len(result.Slots))
	}

	check, err := service.CheckSlot(
		context.Background(),
		date.Format("2006-01-02"),
		"09:00",
	)
	if err != nil {
		t.Fatalf("CheckSlot() error = %v", err)
	}
	if check.Available {
		t.Fatal("CheckSlot() available = true for confirmed database booking")
	}
}

func nextWeekday(date time.Time) time.Time {
	for date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, 1)
	}
	return date
}
