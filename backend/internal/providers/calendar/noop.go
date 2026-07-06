package calendar

import (
	"context"
	"time"
)

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

type Unavailable struct {
	err error
}

func NewUnavailable(err error) *Unavailable {
	return &Unavailable{err: err}
}

func (u *Unavailable) FreeBusy(ctx context.Context, start, end time.Time) ([]BusyPeriod, error) {
	return nil, u.err
}

func (u *Unavailable) CreateEvent(ctx context.Context, input CreateEventInput) (CreateEventResult, error) {
	return CreateEventResult{}, u.err
}

func (u *Unavailable) DeleteEvent(ctx context.Context, eventID string) error {
	return u.err
}
