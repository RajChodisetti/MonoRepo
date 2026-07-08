package handlers

import (
	"errors"
	"net/http"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
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

	result, err := handler.service.TriggerBulkSend(r.Context(), principal)
	if errors.Is(err, outreach.ErrBulkJobActive) {
		handler.writeError(w, http.StatusConflict, "bulk_job_active", "A bulk outreach job is already queued or running.")
		return
	}
	if errors.Is(err, outreach.ErrNotConfigured) {
		handler.writeError(w, http.StatusServiceUnavailable, "outreach_not_configured", "Bulk outreach email accounts are not configured.")
		return
	}
	if err != nil {
		handler.writeError(w, http.StatusInternalServerError, "bulk_send_failed", err.Error())
		return
	}

	handler.writeJSON(w, http.StatusAccepted, result)
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
