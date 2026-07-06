package reservations

import (
	"context"
	"time"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Mock struct{}

func (mock *Mock) Create(_ context.Context, _ uuid.UUID, _ CreateInput) (Reservation, error) {
	return Reservation{}, nil
}

func (mock *Mock) GetByClientRequestID(_ context.Context, _ uuid.UUID, _ string) (Reservation, error) {
	return Reservation{}, repository.ErrNotFound
}

func (mock *Mock) CountBySlot(_ context.Context, _ uuid.UUID, _ time.Time, _ string) (int, error) {
	return 0, nil
}

func (mock *Mock) GetOpeningHours(_ context.Context, _ uuid.UUID) (map[string]string, error) {
	return defaultHours(), nil
}

func (mock *Mock) RestaurantExists(_ context.Context, _ uuid.UUID) (bool, error) {
	return true, nil
}
