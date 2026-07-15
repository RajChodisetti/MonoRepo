package calendar

import (
	"context"
	"time"
)

type BusyPeriod struct {
	Start time.Time
	End   time.Time
}

type CreateEventInput struct {
	Title       string
	Description string
	Start       time.Time
	End         time.Time
	Attendee    string
}

type CreateEventResult struct {
	EventID  string
	HTMLLink string
}

type Provider interface {
	FreeBusy(ctx context.Context, start, end time.Time) ([]BusyPeriod, error)
	CreateEvent(ctx context.Context, input CreateEventInput) (CreateEventResult, error)
	DeleteEvent(ctx context.Context, eventID string) error
}
