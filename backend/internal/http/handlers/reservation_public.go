package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/reservations"
)

type ReservationPublicHandler struct {
	service    *reservations.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewReservationPublicHandler(
	service *reservations.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *ReservationPublicHandler {
	return &ReservationPublicHandler{
		service:    service,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

func (handler *ReservationPublicHandler) GetTableAvailability(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_id", "Restaurant id must be a valid UUID.")
		return
	}

	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_date", "Query parameter date is required (YYYY-MM-DD).")
		return
	}

	partySize := 2
	if raw := strings.TrimSpace(r.URL.Query().Get("party_size")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			handler.writeError(w, http.StatusBadRequest, "invalid_party_size", "party_size must be a positive integer.")
			return
		}
		partySize = parsed
	}

	result, err := handler.service.GetAvailability(r.Context(), restaurantID, date, partySize)
	if err != nil {
		handler.writeAvailabilityError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *ReservationPublicHandler) PutReservation(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_id", "Restaurant id must be a valid UUID.")
		return
	}

	var req reservations.PutReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_body", "Request body must be valid JSON.")
		return
	}

	result, err := handler.service.PutReservation(r.Context(), restaurantID, req)
	if err != nil {
		handler.writeAvailabilityError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *ReservationPublicHandler) writeAvailabilityError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		handler.writeError(w, http.StatusNotFound, "not_found", "Restaurant was not found.")
		return
	}
	msg := err.Error()
	if strings.Contains(msg, "party_size") || strings.Contains(msg, "date") || strings.Contains(msg, "guest_") || strings.Contains(msg, "slot") {
		handler.writeError(w, http.StatusBadRequest, "validation_error", msg)
		return
	}
	if strings.Contains(msg, "not available") {
		handler.writeError(w, http.StatusConflict, "slot_unavailable", msg)
		return
	}
	handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
}
