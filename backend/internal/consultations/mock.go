package consultations

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Mock struct {
	Booked    []BookedInterval
	Overrides map[time.Time]bool
	Revisions map[string]int64
	Inserted  []InsertInput
	InsertErr error
}

func (mock *Mock) BookedIntervals(ctx context.Context, from, to time.Time) ([]BookedInterval, error) {
	var intervals []BookedInterval
	for _, booked := range mock.Booked {
		if intervalsOverlap(booked.SlotStart, booked.SlotEnd, from, to) {
			intervals = append(intervals, booked)
		}
	}
	return intervals, nil
}

func (mock *Mock) HasConfirmedOverlap(
	ctx context.Context,
	slotStart, slotEnd time.Time,
) (bool, error) {
	for _, booked := range mock.Booked {
		if intervalsOverlap(booked.SlotStart, booked.SlotEnd, slotStart, slotEnd) {
			return true, nil
		}
	}
	return false, nil
}

func (mock *Mock) CalendarRevision(ctx context.Context, monthStart time.Time) (int64, error) {
	return mock.Revisions[monthStart.Format("2006-01")], nil
}

func (mock *Mock) SlotOverrides(ctx context.Context, from, to time.Time) ([]SlotOverride, error) {
	var overrides []SlotOverride
	for slot, available := range mock.Overrides {
		if (slot.Equal(from) || slot.After(from)) && slot.Before(to) {
			overrides = append(overrides, SlotOverride{
				SlotStart:   slot,
				IsAvailable: available,
			})
		}
	}
	return overrides, nil
}

func (mock *Mock) IsSlotEnabled(ctx context.Context, slotStart time.Time) (bool, error) {
	for override, available := range mock.Overrides {
		if override.Equal(slotStart) {
			return available, nil
		}
	}
	return true, nil
}

func (mock *Mock) ReplaceMonthSlotOverrides(
	ctx context.Context,
	monthStart, monthEnd time.Time,
	inputs []SlotOverrideInput,
	expectedRevision int64,
	updatedBy uuid.UUID,
) (int64, error) {
	monthKey := monthStart.Format("2006-01")
	currentRevision := mock.Revisions[monthKey]
	if currentRevision != expectedRevision {
		return 0, ErrCalendarRevisionConflict
	}
	if mock.Overrides == nil {
		mock.Overrides = map[time.Time]bool{}
	}
	for existing := range mock.Overrides {
		if !existing.Before(monthStart) && existing.Before(monthEnd) {
			delete(mock.Overrides, existing)
		}
	}
	for _, input := range inputs {
		for existing := range mock.Overrides {
			if existing.Equal(input.SlotStart) {
				delete(mock.Overrides, existing)
			}
		}
		mock.Overrides[input.SlotStart] = input.IsAvailable
	}
	if mock.Revisions == nil {
		mock.Revisions = map[string]int64{}
	}
	newRevision := currentRevision + 1
	mock.Revisions[monthKey] = newRevision
	return newRevision, nil
}

func (mock *Mock) InsertIfAvailable(ctx context.Context, input InsertInput) (bool, error) {
	if mock.InsertErr != nil {
		return false, mock.InsertErr
	}
	for slot, enabled := range mock.Overrides {
		if slot.Equal(input.SlotStart) && !enabled {
			return false, nil
		}
	}
	for _, booked := range mock.Booked {
		if intervalsOverlap(booked.SlotStart, booked.SlotEnd, input.SlotStart, input.SlotEnd) {
			return false, ErrConflict
		}
	}
	mock.Booked = append(mock.Booked, BookedInterval{
		SlotStart: input.SlotStart,
		SlotEnd:   input.SlotEnd,
	})
	mock.Inserted = append(mock.Inserted, input)
	return true, nil
}

func (mock *Mock) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func intervalsOverlap(firstStart, firstEnd, secondStart, secondEnd time.Time) bool {
	return firstStart.Before(secondEnd) && firstEnd.After(secondStart)
}
