package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type CampaignHandler struct {
	service    *campaigns.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewCampaignHandler(
	service *campaigns.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *CampaignHandler {
	return &CampaignHandler{
		service:    service,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

type createCampaignRequest struct {
	DemoSiteID   string `json:"demo_site_id"`
	DemoToken    string `json:"demo_token"`
	CampaignType string `json:"campaign_type"`
}

type stopCampaignRequest struct {
	Reason string `json:"reason"`
}

type approveCampaignRequest struct {
	ExpectedUpdatedAt *time.Time `json:"expected_updated_at"`
}

type sendStepRequest struct {
	Step *int `json:"step"`
}

func (handler *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Could not read request body.")
		return
	}

	var request createCampaignRequest
	if len(body) > 0 {
		if err := json.Unmarshal(body, &request); err != nil {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
			return
		}
	}

	demoSiteID, err := uuid.Parse(strings.TrimSpace(request.DemoSiteID))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "demo_site_id must be a valid UUID.")
		return
	}
	if strings.TrimSpace(request.DemoToken) == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "demo_token is required.")
		return
	}

	record, err := handler.service.CreateDraft(r.Context(), principal, campaigns.CreateInput{
		RestaurantID: restaurantID,
		DemoSiteID:   demoSiteID,
		DemoToken:    request.DemoToken,
		CampaignType: request.CampaignType,
	})
	if err != nil {
		handler.mapCampaignError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusCreated, campaignResponse(record))
}

func (handler *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
		return
	}

	records, err := handler.service.ListByRestaurant(r.Context(), principal, restaurantID)
	if err != nil {
		handler.mapCampaignError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		items = append(items, campaignResponse(record))
	}
	handler.writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (handler *CampaignHandler) Get(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	campaignID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Campaign id must be a valid UUID.")
		return
	}

	campaign, events, err := handler.service.GetByID(r.Context(), principal, campaignID)
	if err != nil {
		handler.mapCampaignError(w, err)
		return
	}

	eventItems := make([]map[string]any, 0, len(events))
	for _, event := range events {
		eventItems = append(eventItems, map[string]any{
			"id":            event.ID,
			"event_type":    event.EventType,
			"metadata":      json.RawMessage(event.Metadata),
			"event_time":    event.EventTime,
			"restaurant_id": event.RestaurantID,
		})
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{
		"campaign": campaignResponse(campaign),
		"events":   eventItems,
	})
}

func (handler *CampaignHandler) Approve(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	campaignID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Campaign id must be a valid UUID.")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 16<<10))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Could not read request body.")
		return
	}
	var request approveCampaignRequest
	if len(body) == 0 || json.Unmarshal(body, &request) != nil || request.ExpectedUpdatedAt == nil {
		handler.writeError(w, http.StatusBadRequest, "expected_updated_at_required", "expected_updated_at from the reviewed campaign detail is required.")
		return
	}

	record, err := handler.service.Approve(r.Context(), principal, campaignID, *request.ExpectedUpdatedAt)
	if err != nil {
		handler.mapCampaignError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, campaignResponse(record))
}

func (handler *CampaignHandler) Regenerate(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	campaignID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Campaign id must be a valid UUID.")
		return
	}

	record, err := handler.service.RegenerateDraft(r.Context(), principal, campaignID)
	if err != nil {
		handler.mapCampaignError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, campaignResponse(record))
}

func (handler *CampaignHandler) SendStep(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	campaignID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Campaign id must be a valid UUID.")
		return
	}

	step := 0
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Could not read request body.")
		return
	}
	if len(body) > 0 {
		var request sendStepRequest
		if err := json.Unmarshal(body, &request); err != nil {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
			return
		}
		if request.Step != nil {
			step = *request.Step
		}
	}

	record, err := handler.service.SendStep(r.Context(), principal, campaignID, step)
	if err != nil {
		handler.mapCampaignError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, campaignResponse(record))
}

func (handler *CampaignHandler) Stop(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	campaignID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Campaign id must be a valid UUID.")
		return
	}

	reason := ""
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Could not read request body.")
		return
	}
	if len(body) > 0 {
		var request stopCampaignRequest
		if err := json.Unmarshal(body, &request); err != nil {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
			return
		}
		reason = request.Reason
	}

	record, err := handler.service.Stop(r.Context(), principal, campaignID, reason)
	if err != nil {
		handler.mapCampaignError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, campaignResponse(record))
}

func campaignResponse(record campaigns.Campaign) map[string]any {
	return map[string]any{
		"id":             record.ID,
		"restaurant_id":  record.RestaurantID,
		"demo_site_id":   record.DemoSiteID,
		"campaign_type":  record.CampaignType,
		"status":         record.Status,
		"current_step":   record.CurrentStep,
		"subject":        record.Subject,
		"body_html":      record.BodyHTML,
		"body_text":      record.BodyText,
		"approved_at":    record.ApprovedAt,
		"approved_by":    record.ApprovedBy,
		"last_sent_at":   record.LastSentAt,
		"stopped_reason": record.StoppedReason,
		"created_at":     record.CreatedAt,
		"updated_at":     record.UpdatedAt,
	}
}

func (handler *CampaignHandler) mapCampaignError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, restaurants.ErrForbidden):
		handler.writeError(w, http.StatusForbidden, "forbidden", "You do not have access to this resource.")
	case errors.Is(err, repository.ErrNotFound):
		handler.writeError(w, http.StatusNotFound, "not_found", "Campaign was not found.")
	case errors.Is(err, campaigns.ErrAlreadyStopped):
		handler.writeError(w, http.StatusConflict, "campaign_stopped", err.Error())
	case errors.Is(err, campaigns.ErrStaleReview):
		handler.writeError(w, http.StatusConflict, "stale_review", "The campaign changed after review; inspect it again before approving.")
	default:
		var eligibility campaigns.ErrEligibility
		if errors.As(err, &eligibility) {
			handler.writeError(w, http.StatusConflict, "not_eligible", err.Error())
			return
		}
		if strings.Contains(err.Error(), "no contact email") {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
	}
}
