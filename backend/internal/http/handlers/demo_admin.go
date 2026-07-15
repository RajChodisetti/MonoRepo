package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type DemoAdminHandler struct {
	demoService *demos.Service
	writeJSON   func(http.ResponseWriter, int, any)
	writeError  func(http.ResponseWriter, int, string, string)
}

func NewDemoAdminHandler(
	demoService *demos.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *DemoAdminHandler {
	return &DemoAdminHandler{
		demoService: demoService,
		writeJSON:   writeJSON,
		writeError:  writeError,
	}
}

type createDemoSiteRequest struct {
	Slug          string          `json:"slug"`
	Status        string          `json:"status"`
	PublicPayload json.RawMessage `json:"public_payload"`
}

func (handler *DemoAdminHandler) ReviewPreview(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	demoSiteID, err := uuid.Parse(strings.TrimSpace(r.PathValue("id")))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Demo site id must be a valid UUID.")
		return
	}

	result, err := handler.demoService.GetReviewPreview(r.Context(), principal, demoSiteID)
	if err != nil {
		handler.mapDemoError(w, err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *DemoAdminHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var request createDemoSiteRequest
	if len(body) > 0 {
		if err := json.Unmarshal(body, &request); err != nil {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
			return
		}
	}

	status := strings.TrimSpace(request.Status)
	if status == "" {
		status = demos.StatusDraft
	}

	result, err := handler.demoService.CreateDemoSite(r.Context(), principal, restaurantID, demos.CreateDemoInput{
		Slug:          request.Slug,
		Status:        status,
		PublicPayload: request.PublicPayload,
	})
	if err != nil {
		handler.mapDemoError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusCreated, result)
}

func (handler *DemoAdminHandler) mapDemoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, restaurants.ErrForbidden):
		handler.writeError(w, http.StatusForbidden, "forbidden", "You do not have access to this restaurants.")
	case errors.Is(err, repository.ErrNotFound):
		handler.writeError(w, http.StatusNotFound, "not_found", "Demo site was not found.")
	default:
		if strings.Contains(err.Error(), "slug is required") ||
			strings.Contains(err.Error(), "must be created as drafts") ||
			strings.Contains(err.Error(), "unsupported demo status") {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error()+".")
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
	}
}
