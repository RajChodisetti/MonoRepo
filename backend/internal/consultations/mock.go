package consultations

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Mock struct {
	Booked map[time.Time]struct{}
}

func (mock *Mock) BookedSlotStarts(ctx context.Context, from, to time.Time) ([]time.Time, error) {
	var slots []time.Time
	for slot := range mock.Booked {
		if (slot.Equal(from) || slot.After(from)) && slot.Before(to) {
			slots = append(slots, slot)
		}
	}
	return slots, nil
}

func (mock *Mock) IsSlotBooked(ctx context.Context, slotStart time.Time) (bool, error) {
	_, ok := mock.Booked[slotStart]
	return ok, nil
}

func (mock *Mock) Insert(ctx context.Context, tx pgx.Tx, input InsertInput) error {
	return nil
}

func (mock *Mock) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (mock *Mock) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return nil, fmt.Errorf("mock consultation transaction is not configured")
}
