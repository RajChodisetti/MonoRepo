package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/analytics"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
)

type DemoEngagementHandler struct {
	service    *analytics.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewDemoEngagementHandler(
	service *analytics.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *DemoEngagementHandler {
	return &DemoEngagementHandler{service: service, writeJSON: writeJSON, writeError: writeError}
}

type startDemoSessionRequest struct {
	DemoToken  string `json:"demo_token"`
	TemplateID string `json:"template_id"`
}

func (handler *DemoEngagementHandler) Start(w http.ResponseWriter, r *http.Request) {
	var request startDemoSessionRequest
	if err := decodeEngagementBody(r, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := handler.service.StartSession(r.Context(), r.PathValue("slug"), request.DemoToken, request.TemplateID)
	if err != nil {
		if errors.Is(err, analytics.ErrInvalidEvent) {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", "template_id must be 1, 2, 3, or 4.")
			return
		}
		if errors.Is(err, demos.ErrDemoNotFound) {
			handler.writeError(w, http.StatusNotFound, "demo_not_found", "Demo site was not found.")
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "engagement_unavailable", "Demo engagement could not be started.")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	handler.writeJSON(w, http.StatusCreated, result)
}

type startAdminPreviewRequest struct {
	TemplateID string `json:"template_id"`
}

func (handler *DemoEngagementHandler) StartAdminPreview(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
		return
	}
	var request startAdminPreviewRequest
	if err := decodeEngagementBody(r, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := handler.service.StartAdminPreview(r.Context(), restaurantID, request.TemplateID)
	if errors.Is(err, analytics.ErrInvalidEvent) {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "template_id must be 1, 2, 3, or 4.")
		return
	}
	if err != nil {
		handler.writeError(w, http.StatusInternalServerError, "engagement_unavailable", "Demo engagement could not be started.")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	handler.writeJSON(w, http.StatusCreated, result)
}

type touchDemoSessionRequest struct {
	SessionToken  string `json:"session_token"`
	Event         string `json:"event"`
	ActiveSeconds int    `json:"active_seconds"`
}

func (handler *DemoEngagementHandler) Touch(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(r.PathValue("session_id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "session_id must be a valid UUID.")
		return
	}
	var request touchDemoSessionRequest
	if err := decodeEngagementBody(r, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	event := strings.ToLower(strings.TrimSpace(request.Event))
	if event != "heartbeat" && event != "end" {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "event must be heartbeat or end.")
		return
	}
	if request.ActiveSeconds < 0 || request.ActiveSeconds > 86400 {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "active_seconds must be between 0 and 86400.")
		return
	}
	if err := handler.service.Touch(r.Context(), sessionID, request.SessionToken, request.ActiveSeconds, event == "end"); err != nil {
		handler.mapSessionError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type transcriptRequest struct {
	SessionToken string `json:"session_token"`
	Role         string `json:"role"`
	Content      string `json:"content"`
}

func (handler *DemoEngagementHandler) Transcript(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(r.PathValue("session_id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "session_id must be a valid UUID.")
		return
	}
	var request transcriptRequest
	if err := decodeEngagementBody(r, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := handler.service.AddTranscript(r.Context(), sessionID, request.SessionToken, request.Role, request.Content); err != nil {
		if errors.Is(err, analytics.ErrInvalidEvent) {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", "Transcript role or content is invalid.")
			return
		}
		handler.mapSessionError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *DemoEngagementHandler) ListByRestaurant(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
		return
	}
	sessions, err := handler.service.ListSessions(r.Context(), restaurantID)
	if err != nil {
		handler.writeError(w, http.StatusInternalServerError, "engagement_unavailable", "Demo engagement is unavailable.")
		return
	}
	handler.writeJSON(w, http.StatusOK, map[string]any{"items": sessions})
}

func (handler *DemoEngagementHandler) mapSessionError(w http.ResponseWriter, err error) {
	if errors.Is(err, analytics.ErrSessionNotFound) {
		handler.writeError(w, http.StatusNotFound, "session_not_found", "Demo session was not found.")
		return
	}
	handler.writeError(w, http.StatusInternalServerError, "engagement_unavailable", "Demo engagement could not be recorded.")
}

func decodeEngagementBody(r *http.Request, target any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		return errors.New("Could not read request body.")
	}
	if len(body) == 0 || json.Unmarshal(body, target) != nil {
		return errors.New("Request body must be valid JSON.")
	}
	return nil
}
