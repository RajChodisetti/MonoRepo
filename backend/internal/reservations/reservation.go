package reservations

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"

	SourceVoiceAgent = "voice_agent"
	SourceWebForm    = "web_form"

	DefaultMaxTablesPerSlot = 4
	SlotIntervalMinutes     = 30
)

type Reservation struct {
	ID              uuid.UUID `json:"id"`
	RestaurantID    uuid.UUID `json:"restaurant_id"`
	GuestName       string    `json:"guest_name"`
	GuestPhone      string    `json:"guest_phone"`
	GuestEmail      string    `json:"guest_email,omitempty"`
	PartySize       int       `json:"party_size"`
	ReservationDate time.Time `json:"reservation_date"`
	ReservationTime string    `json:"reservation_time"`
	Status          string    `json:"status"`
	Source          string    `json:"source"`
	Notes           string    `json:"notes,omitempty"`
	ClientRequestID string    `json:"client_request_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateInput struct {
	GuestName       string
	GuestPhone      string
	GuestEmail      string
	PartySize       int
	ReservationDate time.Time
	ReservationTime string
	Source          string
	Notes           string
	ClientRequestID string
}

type AvailabilityResult struct {
	AvailableSlots []string `json:"available_slots"`
	Timezone       string   `json:"timezone"`
	Date           string   `json:"date"`
	PartySize      int      `json:"party_size"`
}

type Repository interface {
	Create(ctx context.Context, restaurantID uuid.UUID, input CreateInput) (Reservation, error)
	GetByClientRequestID(ctx context.Context, restaurantID uuid.UUID, clientRequestID string) (Reservation, error)
	CountBySlot(ctx context.Context, restaurantID uuid.UUID, date time.Time, slotTime string) (int, error)
	GetOpeningHours(ctx context.Context, restaurantID uuid.UUID) (map[string]string, error)
	RestaurantExists(ctx context.Context, restaurantID uuid.UUID) (bool, error)
}
