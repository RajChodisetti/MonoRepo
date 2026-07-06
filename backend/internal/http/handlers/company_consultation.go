package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/consultations"
)

type CompanyConsultationHandler struct {
	service   *consultations.Service
	writeJSON func(http.ResponseWriter, int, any)
}

func NewCompanyConsultationHandler(
	service *consultations.Service,
	writeJSON func(http.ResponseWriter, int, any),
) *CompanyConsultationHandler {
	return &CompanyConsultationHandler{
		service:   service,
		writeJSON: writeJSON,
	}
}

func (handler *CompanyConsultationHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	days := parseIntDefault(r.URL.Query().Get("days"), 0)

	result, err := handler.service.GetAvailability(r.Context(), date, days)
	if err != nil {
		handler.writeStatusError(w, http.StatusBadRequest, err.Error())
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *CompanyConsultationHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	timeStr := strings.TrimSpace(r.URL.Query().Get("time"))
	if date == "" || timeStr == "" {
		handler.writeStatusError(w, http.StatusBadRequest, "date and time query params are required")
		return
	}

	result, err := handler.service.CheckSlot(r.Context(), date, timeStr)
	if err != nil {
		handler.writeStatusError(w, http.StatusBadRequest, err.Error())
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *CompanyConsultationHandler) Book(w http.ResponseWriter, r *http.Request) {
	var req consultations.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.writeStatusError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	success, conflict, err := handler.service.Book(r.Context(), req)
	if conflict != nil {
		handler.writeJSON(w, http.StatusConflict, conflict)
		return
	}
	if err != nil {
		handler.writeStatusError(w, http.StatusBadRequest, err.Error())
		return
	}
	handler.writeJSON(w, http.StatusCreated, success)
}

func (handler *CompanyConsultationHandler) writeStatusError(w http.ResponseWriter, status int, message string) {
	handler.writeJSON(w, status, map[string]any{
		"status":  "error",
		"message": message,
	})
}

func parseIntDefault(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}
