package calendar

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
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

func NewFromConfig(ctx context.Context, cfg config.ConsultationConfig, log *slog.Logger) Provider {
	if cfg.GoogleCalendarDisabled {
		if log != nil {
			log.InfoContext(ctx, "consultation_calendar_disabled")
		}
		return NewNoop()
	}

	googleProvider, activeID, err := EnsureWritable(ctx, cfg.GoogleCalendarID, cfg.GoogleServiceAccountJSON)
	if err != nil {
		if log != nil {
			log.ErrorContext(ctx, "consultation_calendar_init_failed", "error", err)
		}
		return NewUnavailable(fmt.Errorf("consultation calendar unavailable: %w", err))
	}
	if log != nil {
		if activeID != cfg.GoogleCalendarID {
			log.WarnContext(ctx, "consultation_calendar_fallback", "configured_calendar_id", cfg.GoogleCalendarID, "active_calendar_id", activeID)
		} else {
			log.InfoContext(ctx, "consultation_calendar_connected", "calendar_id", activeID)
		}
	}
	return googleProvider
}
