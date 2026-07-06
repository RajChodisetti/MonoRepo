package reservations

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (service *Service) GetAvailability(ctx context.Context, restaurantID uuid.UUID, dateStr string, partySize int) (AvailabilityResult, error) {
	if partySize < 1 || partySize > 20 {
		return AvailabilityResult{}, fmt.Errorf("party_size must be between 1 and 20")
	}

	exists, err := service.repo.RestaurantExists(ctx, restaurantID)
	if err != nil {
		return AvailabilityResult{}, err
	}
	if !exists {
		return AvailabilityResult{}, repository.ErrNotFound
	}

	date, err := time.ParseInLocation("2006-01-02", dateStr, SydneyLocation)
	if err != nil {
		return AvailabilityResult{}, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}

	today := time.Now().In(SydneyLocation)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, SydneyLocation)
	if date.Before(todayDate) {
		return AvailabilityResult{}, fmt.Errorf("date cannot be in the past")
	}

	pgRepo, ok := service.repo.(*Postgres)
	if !ok {
		return AvailabilityResult{}, fmt.Errorf("availability requires postgres repository")
	}

	return ComputeAvailability(ctx, pgRepo, restaurantID, date, partySize)
}

type PutReservationRequest struct {
	GuestName       string `json:"guest_name"`
	GuestPhone      string `json:"guest_phone"`
	GuestEmail      string `json:"guest_email,omitempty"`
	PartySize       int    `json:"party_size"`
	Slot            string `json:"slot"`
	Source          string `json:"source,omitempty"`
	Notes           string `json:"notes,omitempty"`
	ClientRequestID string `json:"client_request_id,omitempty"`
}

type PutReservationResponse struct {
	Status        string `json:"status"`
	ReservationID string `json:"reservation_id"`
	Message       string `json:"message"`
}

func (service *Service) PutReservation(ctx context.Context, restaurantID uuid.UUID, req PutReservationRequest) (PutReservationResponse, error) {
	req.GuestName = strings.TrimSpace(req.GuestName)
	req.GuestPhone = strings.TrimSpace(req.GuestPhone)
	req.GuestEmail = strings.TrimSpace(req.GuestEmail)
	req.Slot = strings.TrimSpace(req.Slot)
	req.Source = strings.TrimSpace(req.Source)
	req.ClientRequestID = strings.TrimSpace(req.ClientRequestID)

	if req.GuestName == "" {
		return PutReservationResponse{}, fmt.Errorf("guest_name is required")
	}
	if req.GuestPhone == "" {
		return PutReservationResponse{}, fmt.Errorf("guest_phone is required")
	}
	if req.PartySize < 1 || req.PartySize > 20 {
		return PutReservationResponse{}, fmt.Errorf("party_size must be between 1 and 20")
	}
	if req.Slot == "" {
		return PutReservationResponse{}, fmt.Errorf("slot is required")
	}
	source := SourceVoiceAgent
	if req.Source != "" {
		switch req.Source {
		case SourceVoiceAgent, SourceWebForm:
			source = req.Source
		default:
			return PutReservationResponse{}, fmt.Errorf("source must be voice_agent or web_form")
		}
	}

	exists, err := service.repo.RestaurantExists(ctx, restaurantID)
	if err != nil {
		return PutReservationResponse{}, err
	}
	if !exists {
		return PutReservationResponse{}, repository.ErrNotFound
	}

	if req.ClientRequestID != "" {
		existing, err := service.repo.GetByClientRequestID(ctx, restaurantID, req.ClientRequestID)
		if err == nil {
			return PutReservationResponse{
				Status:        existing.Status,
				ReservationID: existing.ID.String(),
				Message:       fmt.Sprintf("Reservation already recorded for %d guests.", existing.PartySize),
			}, nil
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return PutReservationResponse{}, err
		}
	}

	resDate, resTime, err := ParseReservationSlot(req.Slot)
	if err != nil {
		return PutReservationResponse{}, err
	}

	availability, err := service.GetAvailability(ctx, restaurantID, resDate.Format("2006-01-02"), req.PartySize)
	if err != nil {
		return PutReservationResponse{}, err
	}
	if !SlotAllowed(availability.AvailableSlots, req.Slot) {
		return PutReservationResponse{}, fmt.Errorf("selected slot is not available")
	}

	record, err := service.repo.Create(ctx, restaurantID, CreateInput{
		GuestName:       req.GuestName,
		GuestPhone:      req.GuestPhone,
		GuestEmail:      req.GuestEmail,
		PartySize:       req.PartySize,
		ReservationDate: resDate,
		ReservationTime: resTime,
		Source:          source,
		Notes:           req.Notes,
		ClientRequestID: req.ClientRequestID,
	})
	if err != nil {
		return PutReservationResponse{}, err
	}

	slotTime, _ := time.Parse("15:04:05", resTime)
	displayTime := slotTime.Format("3:04 PM")

	return PutReservationResponse{
		Status:        record.Status,
		ReservationID: record.ID.String(),
		Message:       fmt.Sprintf("Reservation request received for %d guests at %s. Status is pending confirmation.", req.PartySize, displayTime),
	}, nil
}
