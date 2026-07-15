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

	timezone, location, err := service.restaurantLocation(ctx, restaurantID)
	if err != nil {
		return AvailabilityResult{}, err
	}

	date, err := time.ParseInLocation("2006-01-02", dateStr, location)
	if err != nil {
		return AvailabilityResult{}, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}

	today := time.Now().In(location)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, location)
	if date.Before(todayDate) {
		return AvailabilityResult{}, fmt.Errorf("date cannot be in the past")
	}

	return service.computeAvailability(ctx, restaurantID, date, partySize, timezone, location)
}

func (service *Service) restaurantLocation(ctx context.Context, restaurantID uuid.UUID) (string, *time.Location, error) {
	timezone := DefaultTimezone
	if timezoneRepo, ok := service.repo.(TimezoneRepository); ok {
		resolved, err := timezoneRepo.GetRestaurantTimezone(ctx, restaurantID)
		if err != nil {
			return "", nil, err
		}
		if strings.TrimSpace(resolved) != "" {
			timezone = resolved
		}
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return "", nil, fmt.Errorf("load restaurant timezone %q: %w", timezone, err)
	}
	return timezone, location, nil
}

func (service *Service) computeAvailability(
	ctx context.Context,
	restaurantID uuid.UUID,
	serviceDate time.Time,
	partySize int,
	timezone string,
	location *time.Location,
) (AvailabilityResult, error) {
	pgRepo, ok := service.repo.(*Postgres)
	if !ok {
		return AvailabilityResult{}, fmt.Errorf("availability requires postgres repository")
	}
	return computeAvailability(ctx, pgRepo, restaurantID, serviceDate, partySize, timezone, location)
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

	timezone, location, err := service.restaurantLocation(ctx, restaurantID)
	if err != nil {
		return PutReservationResponse{}, err
	}
	resDate, resTime, err := ParseReservationSlotInLocation(req.Slot, location)
	if err != nil {
		return PutReservationResponse{}, err
	}
	input := CreateInput{
		GuestName:       req.GuestName,
		GuestPhone:      req.GuestPhone,
		GuestEmail:      req.GuestEmail,
		PartySize:       req.PartySize,
		ReservationDate: resDate,
		ReservationTime: resTime,
		Source:          source,
		Notes:           req.Notes,
		ClientRequestID: req.ClientRequestID,
	}

	if req.ClientRequestID != "" {
		existing, err := service.repo.GetByClientRequestID(ctx, restaurantID, req.ClientRequestID)
		if err == nil {
			if !reservationMatchesInput(existing, input) {
				return PutReservationResponse{}, ErrClientRequestConflict
			}
			return duplicateReservationResponse(existing), nil
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return PutReservationResponse{}, err
		}
	}

	today := time.Now().In(location)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, location)
	if resDate.Before(todayDate) {
		return PutReservationResponse{}, fmt.Errorf("date cannot be in the past")
	}

	availability, err := service.computeAvailability(
		ctx,
		restaurantID,
		resDate,
		req.PartySize,
		timezone,
		location,
	)
	if err != nil {
		return PutReservationResponse{}, err
	}
	allowed := SlotAllowed(availability.AvailableSlots, req.Slot)
	if !allowed {
		// An overnight service window is anchored to the previous opening day,
		// while the persisted reservation date is the slot's actual local date.
		previousServiceDate := resDate.AddDate(0, 0, -1)
		previousAvailability, previousErr := service.computeAvailability(
			ctx,
			restaurantID,
			previousServiceDate,
			req.PartySize,
			timezone,
			location,
		)
		if previousErr == nil {
			allowed = SlotAllowed(previousAvailability.AvailableSlots, req.Slot)
		}
	}
	if !allowed {
		return PutReservationResponse{}, ErrSlotUnavailable
	}

	record, err := service.repo.Create(ctx, restaurantID, input)
	if err != nil {
		return PutReservationResponse{}, err
	}

	slotTime, _ := time.Parse("15:04:05", record.ReservationTime)
	displayTime := slotTime.Format("3:04 PM")
	message := fmt.Sprintf(
		"Reservation request received for %d guests at %s. Status is pending confirmation.",
		record.PartySize,
		displayTime,
	)
	if record.Status == StatusConfirmed {
		message = fmt.Sprintf("Reservation is confirmed for %d guests at %s.", record.PartySize, displayTime)
	} else if record.Status == StatusCancelled {
		message = fmt.Sprintf("The existing reservation request for %d guests at %s is cancelled.", record.PartySize, displayTime)
	}

	return PutReservationResponse{
		Status:        record.Status,
		ReservationID: record.ID.String(),
		Message:       message,
	}, nil
}

func duplicateReservationResponse(existing Reservation) PutReservationResponse {
	return PutReservationResponse{
		Status:        existing.Status,
		ReservationID: existing.ID.String(),
		Message:       fmt.Sprintf("Reservation already recorded for %d guests.", existing.PartySize),
	}
}
