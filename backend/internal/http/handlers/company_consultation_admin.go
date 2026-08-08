package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/consultations"
)

type CompanyConsultationAdminHandler struct {
	service    *consultations.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

type calendarSlotUpdateRequest struct {
	ISO         string `json:"iso"`
	IsAvailable *bool  `json:"is_available"`
}

type calendarUpdateRequest struct {
	ExpectedRevision *int64                       `json:"expected_revision"`
	Slots            *[]calendarSlotUpdateRequest `json:"slots"`
}

const maxConsultationCalendarUpdateBytes = 1 << 20

func NewCompanyConsultationAdminHandler(
	service *consultations.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *CompanyConsultationAdminHandler {
	return &CompanyConsultationAdminHandler{
		service:    service,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

func (handler *CompanyConsultationAdminHandler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.PrincipalFromContext(r.Context()); !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	result, err := handler.service.GetCalendar(r.Context(), r.PathValue("month"))
	if err != nil {
		handler.writeCalendarError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *CompanyConsultationAdminHandler) PutCalendar(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	var request calendarUpdateRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxConsultationCalendarUpdateBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body.")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body.")
		return
	}
	if request.Slots == nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Slots are required.")
		return
	}
	if request.ExpectedRevision == nil || *request.ExpectedRevision < 0 {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "A non-negative expected_revision is required.")
		return
	}
	updates := make([]consultations.CalendarSlotUpdate, 0, len(*request.Slots))
	for _, slot := range *request.Slots {
		if slot.IsAvailable == nil {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", "Every slot requires is_available.")
			return
		}
		updates = append(updates, consultations.CalendarSlotUpdate{
			ISO:         slot.ISO,
			IsAvailable: *slot.IsAvailable,
		})
	}

	result, err := handler.service.UpdateCalendar(
		r.Context(),
		r.PathValue("month"),
		updates,
		*request.ExpectedRevision,
		principal.UserID,
	)
	if err != nil {
		handler.writeCalendarError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *CompanyConsultationAdminHandler) writeCalendarError(w http.ResponseWriter, err error) {
	if errors.Is(err, consultations.ErrCalendarRevisionConflict) {
		handler.writeError(
			w,
			http.StatusConflict,
			"consultation_calendar_conflict",
			"The consultation calendar changed after it was loaded. Refresh the month and reapply your changes.",
		)
		return
	}
	if errors.Is(err, consultations.ErrInvalidCalendar) {
		handler.writeError(w, http.StatusBadRequest, "invalid_consultation_calendar", err.Error())
		return
	}
	handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
}
