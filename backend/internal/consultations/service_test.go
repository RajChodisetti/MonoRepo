package consultations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func TestAvailabilityAndCheckExcludeEveryCandidateOverlappingConfirmedCall(t *testing.T) {
	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}

	date := nextWeekday(time.Now().In(loc).AddDate(0, 0, 2))
	bookedStart := time.Date(date.Year(), date.Month(), date.Day(), 9, 17, 0, 0, loc)
	bookedEnd := bookedStart.Add(30 * time.Minute)
	repo := &Mock{Booked: []BookedInterval{{SlotStart: bookedStart, SlotEnd: bookedEnd}}}
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
		if slot.Time == "09:00" || slot.Time == "09:30" {
			t.Fatalf("candidate overlapping 09:17-09:47 was returned as available: %+v", slot)
		}
	}
	if len(result.Slots) != 14 {
		t.Fatalf("available slots = %d, want 14 after one call overlaps two candidates", len(result.Slots))
	}

	for _, clock := range []string{"09:00", "09:30"} {
		check, err := service.CheckSlot(context.Background(), date.Format("2006-01-02"), clock)
		if err != nil {
			t.Fatalf("CheckSlot(%s) error = %v", clock, err)
		}
		if check.Available {
			t.Fatalf("CheckSlot(%s) available = true for overlapping confirmed call", clock)
		}
	}
}

func TestAvailabilityAndCheckRespectDisabledOverride(t *testing.T) {
	loc := loadSydney(t)
	date := nextWeekday(time.Now().In(loc).AddDate(0, 0, 2))
	disabled := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, loc)
	repo := &Mock{Overrides: map[time.Time]bool{disabled: false}}
	service := newTestService(loc, repo)

	result, err := service.GetAvailability(context.Background(), date.Format("2006-01-02"), 1)
	if err != nil {
		t.Fatalf("GetAvailability() error = %v", err)
	}
	if len(result.Slots) != 15 {
		t.Fatalf("available slots = %d, want 15 after disabling one slot", len(result.Slots))
	}
	for _, slot := range result.Slots {
		if slot.ISO == disabled.Format(time.RFC3339) {
			t.Fatalf("disabled override %s was returned as available", slot.ISO)
		}
	}

	check, err := service.CheckSlot(context.Background(), date.Format("2006-01-02"), "09:00")
	if err != nil {
		t.Fatalf("CheckSlot() error = %v", err)
	}
	if check.Available {
		t.Fatal("CheckSlot() available = true for disabled override")
	}
}

func TestCheckSlotRejectsOffGridTime(t *testing.T) {
	loc := loadSydney(t)
	date := nextWeekday(time.Now().In(loc).AddDate(0, 0, 2))
	service := newTestService(loc, &Mock{})

	_, err := service.CheckSlot(context.Background(), date.Format("2006-01-02"), "09:17")
	if err == nil {
		t.Fatal("CheckSlot() error = nil for off-grid time")
	}
}

func TestCalendarIncludesOverridesAndOffGridBookings(t *testing.T) {
	loc := loadSydney(t)
	monthStart := firstFutureMonth(loc, 2)
	candidates := GenerateMonthCandidateSlots(testSlotConfig(loc), monthStart)
	if len(candidates) < 2 {
		t.Fatal("GenerateMonthCandidateSlots() returned fewer than two candidates")
	}
	disabled := candidates[3]
	offGridBooked := candidates[0].Add(17 * time.Minute)
	repo := &Mock{
		Booked: []BookedInterval{{
			SlotStart: offGridBooked,
			SlotEnd:   offGridBooked.Add(30 * time.Minute),
		}},
		Overrides: map[time.Time]bool{disabled: false},
	}
	service := newTestService(loc, repo)

	calendar, err := service.GetCalendar(context.Background(), monthStart.Format("2006-01"))
	if err != nil {
		t.Fatalf("GetCalendar() error = %v", err)
	}
	if calendar.Timezone != "Australia/Sydney" || calendar.SlotDurationMinutes != 30 {
		t.Fatalf("calendar metadata = %+v", calendar)
	}
	if len(calendar.Slots) != len(candidates)+1 {
		t.Fatalf("calendar slots = %d, want %d", len(calendar.Slots), len(candidates)+1)
	}
	if calendar.BookedCallCount != 1 {
		t.Fatalf("booked_call_count = %d, want one unique call", calendar.BookedCallCount)
	}

	disabledSlot := findCalendarSlot(t, calendar.Slots, disabled)
	if disabledSlot.IsAvailable || disabledSlot.EffectiveAvailable {
		t.Fatalf("disabled slot = %+v, want unavailable", disabledSlot)
	}
	for _, blockedCandidate := range candidates[:2] {
		bookedSlot := findCalendarSlot(t, calendar.Slots, blockedCandidate)
		if !bookedSlot.Booked || bookedSlot.EffectiveAvailable {
			t.Fatalf("overlapping candidate = %+v, want booked and not effectively available", bookedSlot)
		}
	}
	offGridSlot := findCalendarSlot(t, calendar.Slots, offGridBooked)
	if !offGridSlot.Booked || offGridSlot.OnGrid {
		t.Fatalf("off-grid booked slot = %+v, want booked and off-grid", offGridSlot)
	}
}

func TestUpdateCalendarRequiresAndPersistsFullFutureGrid(t *testing.T) {
	loc := loadSydney(t)
	monthStart := firstFutureMonth(loc, 2)
	obsolete := GenerateMonthCandidateSlots(testSlotConfig(loc), monthStart)[0].Add(17 * time.Minute)
	repo := &Mock{Overrides: map[time.Time]bool{obsolete: false}}
	service := newTestService(loc, repo)

	calendar, err := service.GetCalendar(context.Background(), monthStart.Format("2006-01"))
	if err != nil {
		t.Fatalf("GetCalendar() error = %v", err)
	}
	updates := make([]CalendarSlotUpdate, 0, len(calendar.Slots))
	for _, slot := range calendar.Slots {
		if slot.Past || !slot.OnGrid {
			continue
		}
		updates = append(updates, CalendarSlotUpdate{
			ISO:         slot.ISO,
			IsAvailable: true,
		})
	}
	if len(updates) == 0 {
		t.Fatal("future calendar has no editable slots")
	}
	updates[0].IsAvailable = false

	updated, err := service.UpdateCalendar(
		context.Background(),
		monthStart.Format("2006-01"),
		updates,
		calendar.Revision,
		uuid.New(),
	)
	if err != nil {
		t.Fatalf("UpdateCalendar() error = %v", err)
	}
	disabledTime, err := time.Parse(time.RFC3339, updates[0].ISO)
	if err != nil {
		t.Fatalf("parse saved slot: %v", err)
	}
	if available, ok := lookupOverride(repo.Overrides, disabledTime); !ok || available {
		t.Fatalf("saved override found=%v available=%v, want found and false", ok, available)
	}
	if got := findCalendarSlot(t, updated.Slots, disabledTime); got.EffectiveAvailable {
		t.Fatalf("updated disabled slot = %+v, want effectively unavailable", got)
	}
	if updated.Revision != 1 {
		t.Fatalf("updated revision = %d, want 1", updated.Revision)
	}
	if _, ok := lookupOverride(repo.Overrides, obsolete); ok {
		t.Fatalf("obsolete off-grid override %s survived full-month replacement", obsolete.Format(time.RFC3339))
	}

	staleUpdates := append([]CalendarSlotUpdate(nil), updates...)
	staleUpdates[0].IsAvailable = true
	_, err = service.UpdateCalendar(
		context.Background(),
		monthStart.Format("2006-01"),
		staleUpdates,
		calendar.Revision,
		uuid.New(),
	)
	if !errors.Is(err, ErrCalendarRevisionConflict) {
		t.Fatalf("stale UpdateCalendar() error = %v, want ErrCalendarRevisionConflict", err)
	}
	if available, ok := lookupOverride(repo.Overrides, disabledTime); !ok || available {
		t.Fatalf("stale save mutated override found=%v available=%v", ok, available)
	}

	_, err = service.UpdateCalendar(
		context.Background(),
		monthStart.Format("2006-01"),
		updates[:len(updates)-1],
		updated.Revision,
		uuid.New(),
	)
	if !errors.Is(err, ErrInvalidCalendar) {
		t.Fatalf("incomplete UpdateCalendar() error = %v, want ErrInvalidCalendar", err)
	}
}

func TestUpdateCalendarAcceptsCanonicalCandidatesThatAreAlreadyPast(t *testing.T) {
	loc := loadSydney(t)
	now := time.Now().In(loc)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, -1, 0)
	candidates := GenerateMonthCandidateSlots(testSlotConfig(loc), monthStart)
	if len(candidates) == 0 {
		t.Fatal("past month has no candidates")
	}
	repo := &Mock{Overrides: map[time.Time]bool{candidates[0]: false}}
	service := newTestService(loc, repo)
	updates := make([]CalendarSlotUpdate, 0, len(candidates))
	for _, candidate := range candidates {
		updates = append(updates, CalendarSlotUpdate{
			ISO:         candidate.Format(time.RFC3339),
			IsAvailable: true,
		})
	}

	updated, err := service.UpdateCalendar(
		context.Background(),
		monthStart.Format("2006-01"),
		updates,
		0,
		uuid.New(),
	)
	if err != nil {
		t.Fatalf("UpdateCalendar(past rollover slots) error = %v", err)
	}
	if updated.Revision != 1 {
		t.Fatalf("revision = %d, want 1", updated.Revision)
	}
	if len(repo.Overrides) != 0 {
		t.Fatalf("past rollover inputs persisted overrides: %+v", repo.Overrides)
	}
}

func TestUpdateCalendarRejectsDuplicateOffGridAndWrongOffsetSlots(t *testing.T) {
	loc := loadSydney(t)
	monthStart := firstFutureMonth(loc, 2)
	candidates := GenerateMonthCandidateSlots(testSlotConfig(loc), monthStart)
	base := make([]CalendarSlotUpdate, 0, len(candidates))
	for _, candidate := range candidates {
		base = append(base, CalendarSlotUpdate{ISO: candidate.Format(time.RFC3339), IsAvailable: true})
	}

	tests := []struct {
		name   string
		mutate func([]CalendarSlotUpdate) []CalendarSlotUpdate
	}{
		{
			name: "duplicate",
			mutate: func(updates []CalendarSlotUpdate) []CalendarSlotUpdate {
				return append(updates, updates[0])
			},
		},
		{
			name: "off grid",
			mutate: func(updates []CalendarSlotUpdate) []CalendarSlotUpdate {
				parsed, _ := time.Parse(time.RFC3339, updates[0].ISO)
				updates[0].ISO = parsed.Add(17 * time.Minute).Format(time.RFC3339)
				return updates
			},
		},
		{
			name: "wrong offset",
			mutate: func(updates []CalendarSlotUpdate) []CalendarSlotUpdate {
				parsed, _ := time.Parse(time.RFC3339, updates[0].ISO)
				updates[0].ISO = parsed.UTC().Format(time.RFC3339)
				return updates
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updates := test.mutate(append([]CalendarSlotUpdate(nil), base...))
			_, err := newTestService(loc, &Mock{}).UpdateCalendar(
				context.Background(),
				monthStart.Format("2006-01"),
				updates,
				0,
				uuid.New(),
			)
			if !errors.Is(err, ErrInvalidCalendar) {
				t.Fatalf("UpdateCalendar() error = %v, want ErrInvalidCalendar", err)
			}
		})
	}
}

func TestBookReturnsConflictForDisabledOverride(t *testing.T) {
	loc := loadSydney(t)
	date := nextWeekday(time.Now().In(loc).AddDate(0, 0, 2))
	disabled := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, loc)
	repo := &Mock{Overrides: map[time.Time]bool{disabled: false}}
	service := newTestService(loc, repo)

	_, conflict, err := service.Book(context.Background(), BookRequest{
		Date:          date.Format("2006-01-02"),
		Time:          "09:00",
		ProspectName:  "Test Prospect",
		ProspectEmail: "prospect@example.test",
		Source:        SourceWeb,
	})
	if err != nil {
		t.Fatalf("Book() error = %v", err)
	}
	if conflict == nil || conflict.Message != "That slot is not available." {
		t.Fatalf("Book() conflict = %+v, want unavailable conflict", conflict)
	}
	if len(repo.Inserted) != 0 {
		t.Fatalf("inserted consultations = %d, want 0", len(repo.Inserted))
	}
}

func TestBookMapsPostgresExclusionViolationToConflict(t *testing.T) {
	loc := loadSydney(t)
	date := nextWeekday(time.Now().In(loc).AddDate(0, 0, 2))
	repo := &Mock{InsertErr: &pgconn.PgError{Code: "23P01"}}
	service := newTestService(loc, repo)

	_, conflict, err := service.Book(context.Background(), BookRequest{
		Date:          date.Format("2006-01-02"),
		Time:          "09:00",
		ProspectName:  "Test Prospect",
		ProspectEmail: "prospect@example.test",
		Source:        SourceVoice,
	})
	if err != nil {
		t.Fatalf("Book() error = %v", err)
	}
	if conflict == nil || conflict.Message != "That slot is already booked." {
		t.Fatalf("Book() conflict = %+v, want overlap conflict", conflict)
	}
}

func TestBookPersistsEnabledOnGridSlot(t *testing.T) {
	loc := loadSydney(t)
	date := nextWeekday(time.Now().In(loc).AddDate(0, 0, 2))
	slot := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, loc)
	repo := &Mock{Overrides: map[time.Time]bool{slot: true}}
	service := newTestService(loc, repo)

	success, conflict, err := service.Book(context.Background(), BookRequest{
		Date:          date.Format("2006-01-02"),
		Time:          "09:00",
		ProspectName:  "Test Prospect",
		ProspectEmail: "prospect@example.test",
		Source:        SourceWeb,
	})
	if err != nil {
		t.Fatalf("Book() error = %v", err)
	}
	if conflict != nil {
		t.Fatalf("Book() conflict = %+v, want nil", conflict)
	}
	if success.Status != "success" || success.Slot != slot.Format(time.RFC3339) {
		t.Fatalf("Book() success = %+v", success)
	}
	if len(repo.Inserted) != 1 || !repo.Inserted[0].SlotStart.Equal(slot) {
		t.Fatalf("inserted consultations = %+v, want one at %s", repo.Inserted, slot.Format(time.RFC3339))
	}
}

func TestMonthCandidateSlotsFollowSydneyDST(t *testing.T) {
	loc := loadSydney(t)
	monthStart, err := ParseMonth("2026-10", loc)
	if err != nil {
		t.Fatalf("ParseMonth() error = %v", err)
	}
	slots := GenerateMonthCandidateSlots(testSlotConfig(loc), monthStart)
	before := findTime(t, slots, "2026-10-02", "09:00")
	after := findTime(t, slots, "2026-10-05", "09:00")
	_, beforeOffset := before.Zone()
	_, afterOffset := after.Zone()
	if beforeOffset != 10*60*60 || afterOffset != 11*60*60 {
		t.Fatalf("Sydney offsets before=%d after=%d, want +10h then +11h", beforeOffset, afterOffset)
	}
}

func TestDayCandidateSlotsAdvanceContinuously(t *testing.T) {
	loc := loadSydney(t)
	date, err := ParseDate("2026-10-05", loc)
	if err != nil {
		t.Fatalf("ParseDate() error = %v", err)
	}
	cfg := testSlotConfig(loc)
	cfg.SlotDurationMinutes = 45
	slots := GenerateDayCandidateSlots(cfg, date)
	if len(slots) < 3 {
		t.Fatalf("GenerateDayCandidateSlots() slots = %d, want at least 3", len(slots))
	}
	if got := FormatSlotTime(slots[1]); got != "09:45" {
		t.Fatalf("second slot = %s, want 09:45", got)
	}
	if got := FormatSlotTime(slots[2]); got != "10:30" {
		t.Fatalf("third slot = %s, want 10:30", got)
	}
}

func nextWeekday(date time.Time) time.Time {
	for date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

func loadSydney(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}
	return loc
}

func newTestService(loc *time.Location, repo Repository) *Service {
	return NewService(config.ConsultationConfig{
		Timezone:            loc,
		BusinessHourStart:   9,
		BusinessHourEnd:     17,
		SlotDurationMinutes: 30,
		DefaultAvailDays:    5,
		AvailabilityHorizon: 14,
	}, repo, nil, nil)
}

func testSlotConfig(loc *time.Location) SlotConfig {
	return SlotConfig{
		Timezone:            loc,
		BusinessHourStart:   9,
		BusinessHourEnd:     17,
		SlotDurationMinutes: 30,
		HorizonDays:         14,
	}
}

func firstFutureMonth(loc *time.Location, monthsAhead int) time.Time {
	now := time.Now().In(loc).AddDate(0, monthsAhead, 0)
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
}

func findCalendarSlot(t *testing.T, slots []CalendarSlot, target time.Time) CalendarSlot {
	t.Helper()
	for _, slot := range slots {
		parsed, err := time.Parse(time.RFC3339, slot.ISO)
		if err == nil && parsed.Equal(target) {
			return slot
		}
	}
	t.Fatalf("calendar slot %s not found", target.Format(time.RFC3339))
	return CalendarSlot{}
}

func lookupOverride(overrides map[time.Time]bool, target time.Time) (bool, bool) {
	for slot, available := range overrides {
		if slot.Equal(target) {
			return available, true
		}
	}
	return false, false
}

func findTime(t *testing.T, slots []time.Time, date, clock string) time.Time {
	t.Helper()
	for _, slot := range slots {
		if slot.Format("2006-01-02") == date && FormatSlotTime(slot) == clock {
			return slot
		}
	}
	t.Fatalf("slot %s %s not found", date, clock)
	return time.Time{}
}
