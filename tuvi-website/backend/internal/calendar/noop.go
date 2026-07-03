package calendar

import (
	"context"
	"time"
)

// Noop skips Google Calendar integration (local dev only).
type Noop struct{}

func NewNoop() *Noop {
	return &Noop{}
}

func (n *Noop) FreeBusy(ctx context.Context, start, end time.Time) ([]BusyPeriod, error) {
	return nil, nil
}

func (n *Noop) CreateEvent(ctx context.Context, input CreateEventInput) (CreateEventResult, error) {
	return CreateEventResult{EventID: "noop", HTMLLink: ""}, nil
}

func (n *Noop) DeleteEvent(ctx context.Context, eventID string) error {
	return nil
}
