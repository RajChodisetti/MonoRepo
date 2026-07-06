package consultations

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

type Repository interface {
	BookedSlotStarts(ctx context.Context, from, to time.Time) ([]time.Time, error)
	IsSlotBooked(ctx context.Context, slotStart time.Time) (bool, error)
	Insert(ctx context.Context, tx pgx.Tx, input InsertInput) error
	Delete(ctx context.Context, id uuid.UUID) error
	BeginTx(ctx context.Context) (pgx.Tx, error)
}
