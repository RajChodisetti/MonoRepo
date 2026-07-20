package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

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

func (handler *OutreachBulkHandler) PreviewAdHoc(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
		return
	}

	preview, err := handler.service.PreviewAdHoc(r.Context(), principal, restaurantID)
	if err != nil {
		handler.writeAdHocError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, preview)
}

func (handler *OutreachBulkHandler) SendAdHoc(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
		return
	}

	result, err := handler.service.SendAdHoc(r.Context(), principal, restaurantID)
	if err != nil {
		handler.writeAdHocError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, result)
}

type sendAdHocBatchRequest struct {
	RestaurantIDs []string `json:"restaurant_ids"`
}

func (handler *OutreachBulkHandler) SendAdHocBatch(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Could not read request body.")
		return
	}
	var request sendAdHocBatchRequest
	if err := json.Unmarshal(body, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}
	if len(request.RestaurantIDs) == 0 {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "restaurant_ids must not be empty.")
		return
	}

	restaurantIDs := make([]uuid.UUID, 0, len(request.RestaurantIDs))
	for _, raw := range request.RestaurantIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", "restaurant_ids must all be valid UUIDs.")
			return
		}
		restaurantIDs = append(restaurantIDs, id)
	}

	results, err := handler.service.SendAdHocBatch(r.Context(), principal, restaurantIDs)
	if err != nil {
		handler.writeAdHocError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{"items": results})
}

func (handler *OutreachBulkHandler) writeAdHocError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, restaurants.ErrForbidden):
		handler.writeError(w, http.StatusForbidden, "forbidden", "You do not have access to this restaurant.")
	case errors.Is(err, repository.ErrNotFound):
		handler.writeError(w, http.StatusNotFound, "not_found", "Restaurant was not found.")
	case errors.Is(err, outreach.ErrSendingDisabled):
		handler.writeError(w, http.StatusServiceUnavailable, "email_sending_disabled", "Email sending is disabled.")
	case errors.Is(err, outreach.ErrNotConfigured):
		handler.writeError(w, http.StatusServiceUnavailable, "outreach_not_configured", "No email provider is configured.")
	case errors.Is(err, outreach.ErrNoContactEmail):
		handler.writeError(w, http.StatusBadRequest, "no_contact_email", "This restaurant has no valid contact email.")
	case errors.Is(err, outreach.ErrEmailSuppressed):
		handler.writeError(w, http.StatusBadRequest, "email_suppressed", "This recipient has opted out of outreach email.")
	case errors.Is(err, outreach.ErrNoCampaignDraft):
		handler.writeError(w, http.StatusBadRequest, "no_campaign_draft", "No campaign draft exists yet for this restaurant. Create one from the Campaign tab first.")
	case errors.Is(err, outreach.ErrDeliverySkipped):
		handler.writeError(w, http.StatusServiceUnavailable, "adhoc_delivery_skipped", "Email provider skipped or redirected the delivery; the restaurant was not marked contacted.")
	default:
		handler.writeError(w, http.StatusInternalServerError, "adhoc_send_failed", err.Error())
	}
}
