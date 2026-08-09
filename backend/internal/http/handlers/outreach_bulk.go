package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type OutreachBulkHandler struct {
	service    *outreach.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewOutreachBulkHandler(
	service *outreach.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *OutreachBulkHandler {
	return &OutreachBulkHandler{
		service:    service,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

func (handler *OutreachBulkHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	result, err := handler.service.SetEmailJob(r.Context(), principal, true)
	if errors.Is(err, outreach.ErrBulkJobActive) {
		handler.writeError(w, http.StatusConflict, "bulk_job_active", "A bulk outreach job is already queued or running.")
		return
	}
	if errors.Is(err, outreach.ErrNotConfigured) {
		handler.writeError(w, http.StatusServiceUnavailable, "outreach_not_configured", "Bulk outreach email accounts are not configured.")
		return
	}
	if errors.Is(err, outreach.ErrSendingDisabled) {
		handler.writeError(w, http.StatusServiceUnavailable, "email_sending_disabled", "Email sending is disabled.")
		return
	}
	if err != nil {
		handler.writeError(w, http.StatusInternalServerError, "bulk_send_failed", err.Error())
		return
	}

	handler.writeJSON(w, http.StatusAccepted, result)
}

type setEmailJobRequest struct {
	Enabled *bool `json:"enabled"`
}

func (handler *OutreachBulkHandler) SetEmailJob(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Could not read request body.")
		return
	}
	var request setEmailJobRequest
	if json.Unmarshal(body, &request) != nil || request.Enabled == nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "enabled must be true or false.")
		return
	}
	result, err := handler.service.SetEmailJob(r.Context(), principal, *request.Enabled)
	if errors.Is(err, outreach.ErrBulkJobActive) {
		handler.writeError(w, http.StatusConflict, "bulk_job_active", "A bulk outreach job is already queued or running.")
		return
	}
	if errors.Is(err, outreach.ErrNotConfigured) {
		handler.writeError(w, http.StatusServiceUnavailable, "outreach_not_configured", "Gmail OAuth outreach accounts are not configured.")
		return
	}
	if err != nil {
		handler.writeError(w, http.StatusInternalServerError, "email_job_control_failed", err.Error())
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *OutreachBulkHandler) Status(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	result, err := handler.service.GetStatus(r.Context(), principal)
	if err != nil {
		handler.writeError(w, http.StatusInternalServerError, "bulk_status_failed", err.Error())
		return
	}

	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *OutreachBulkHandler) ListSequences(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	result, err := handler.service.ListSequences(r.Context(), principal)
	if err != nil {
		handler.writeSequenceError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *OutreachBulkHandler) CreateSequence(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	var request outreach.CreateSequenceInput
	if err := decodeOutreachJSON(r, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := handler.service.CreateSequenceDraft(r.Context(), principal, request)
	if err != nil {
		handler.writeSequenceError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusCreated, result)
}

func (handler *OutreachBulkHandler) UpdateSequence(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	sequenceID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Sequence id must be a valid UUID.")
		return
	}
	var request outreach.UpdateSequenceInput
	if err := decodeOutreachJSON(r, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := handler.service.UpdateSequenceDraft(r.Context(), principal, sequenceID, request)
	if err != nil {
		handler.writeSequenceError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

type approveSequenceRequest struct {
	ExpectedUpdatedAt time.Time `json:"expected_updated_at"`
}

func (handler *OutreachBulkHandler) ApproveSequence(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	sequenceID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Sequence id must be a valid UUID.")
		return
	}
	var request approveSequenceRequest
	if err := decodeOutreachJSON(r, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := handler.service.ApproveSequence(r.Context(), principal, sequenceID, request.ExpectedUpdatedAt)
	if err != nil {
		handler.writeSequenceError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *OutreachBulkHandler) PreviewSequence(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	sequenceID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Sequence id must be a valid UUID.")
		return
	}
	var request outreach.PreviewSequenceInput
	if err := decodeOutreachJSON(r, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := handler.service.PreviewSequence(r.Context(), principal, sequenceID, request)
	if err != nil {
		handler.writeSequenceError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *OutreachBulkHandler) ListRecipients(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	result, err := handler.service.ListRecipientProgress(r.Context(), principal, limit, offset)
	if err != nil {
		handler.writeSequenceError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func decodeOutreachJSON(r *http.Request, target any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return errors.New("could not read request body")
	}
	if len(body) == 0 || json.Unmarshal(body, target) != nil {
		return errors.New("request body must be valid JSON")
	}
	return nil
}

func (handler *OutreachBulkHandler) writeSequenceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, restaurants.ErrForbidden):
		handler.writeError(w, http.StatusForbidden, "forbidden", "Internal administrator access is required.")
	case errors.Is(err, repository.ErrNotFound):
		handler.writeError(w, http.StatusNotFound, "not_found", "Outreach sequence was not found.")
	case errors.Is(err, outreach.ErrSequenceStale):
		handler.writeError(w, http.StatusConflict, "stale_sequence", err.Error())
	case errors.Is(err, outreach.ErrSequenceInvalid):
		handler.writeError(w, http.StatusBadRequest, "invalid_sequence", err.Error())
	default:
		handler.writeError(w, http.StatusInternalServerError, "sequence_request_failed", err.Error())
	}
}
