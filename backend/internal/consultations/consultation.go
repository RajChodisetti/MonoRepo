package consultations

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"

	SourceVoice = "voice"
	SourceWeb   = "web"
)

type Consultation struct {
	ID               uuid.UUID `json:"id"`
	ConfirmationCode string    `json:"confirmation_code"`
	SlotStart        time.Time `json:"slot_start"`
	SlotEnd          time.Time `json:"slot_end"`
	ProspectName     string    `json:"prospect_name"`
	ProspectEmail    string    `json:"prospect_email"`
	ProspectPhone    string    `json:"prospect_phone"`
	Status           string    `json:"status"`
	GoogleEventID    string    `json:"google_event_id"`
	Source           string    `json:"source"`
	CreatedAt        time.Time `json:"created_at"`
}

type InsertInput struct {
	ID               uuid.UUID
	ConfirmationCode string
	SlotStart        time.Time
	SlotEnd          time.Time
	ProspectName     string
	ProspectEmail    string
	ProspectPhone    string
	GoogleEventID    string
	Source           string
}

type SlotOverride struct {
	SlotStart   time.Time
	IsAvailable bool
}

type SlotOverrideInput struct {
	SlotStart   time.Time
	IsAvailable bool
}

type BookedInterval struct {
	SlotStart time.Time
	SlotEnd   time.Time
}

type CalendarSlotUpdate struct {
	ISO         string `json:"iso"`
	IsAvailable bool   `json:"is_available"`
}

type CalendarSlot struct {
	Date               string `json:"date"`
	Time               string `json:"time"`
	ISO                string `json:"iso"`
	IsAvailable        bool   `json:"is_available"`
	Booked             bool   `json:"booked"`
	Past               bool   `json:"past"`
	EffectiveAvailable bool   `json:"effective_available"`
	OnGrid             bool   `json:"on_grid"`
}

type CalendarResult struct {
	Month               string         `json:"month"`
	Revision            int64          `json:"revision"`
	BookedCallCount     int            `json:"booked_call_count"`
	Timezone            string         `json:"timezone"`
	SlotDurationMinutes int            `json:"slot_duration_minutes"`
	BusinessHourStart   string         `json:"business_hour_start"`
	BusinessHourEnd     string         `json:"business_hour_end"`
	Slots               []CalendarSlot `json:"slots"`
}

type Repository interface {
	BookedIntervals(ctx context.Context, from, to time.Time) ([]BookedInterval, error)
	HasConfirmedOverlap(ctx context.Context, slotStart, slotEnd time.Time) (bool, error)
	SlotOverrides(ctx context.Context, from, to time.Time) ([]SlotOverride, error)
	IsSlotEnabled(ctx context.Context, slotStart time.Time) (bool, error)
	CalendarRevision(ctx context.Context, monthStart time.Time) (int64, error)
	ReplaceMonthSlotOverrides(
		ctx context.Context,
		monthStart, monthEnd time.Time,
		inputs []SlotOverrideInput,
		expectedRevision int64,
		updatedBy uuid.UUID,
	) (int64, error)
	InsertIfAvailable(ctx context.Context, input InsertInput) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
