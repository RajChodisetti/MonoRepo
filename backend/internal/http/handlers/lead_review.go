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
	"github.com/rajchodisetti/restaurant-platform/backend/internal/leadreview"
)

type LeadReviewHandler struct {
	service    *leadreview.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewLeadReviewHandler(
	service *leadreview.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *LeadReviewHandler {
	return &LeadReviewHandler{service: service, writeJSON: writeJSON, writeError: writeError}
}

type leadReviewRequest struct {
	Status                      string     `json:"status"`
	ExpectedUpdatedAt           *time.Time `json:"expected_updated_at"`
	ExpectedRestaurantUpdatedAt *time.Time `json:"expected_restaurant_updated_at"`
	ExpectedProfileUpdatedAt    *time.Time `json:"expected_profile_updated_at"`
}

func (handler *LeadReviewHandler) GetProfileReviewPreview(w http.ResponseWriter, r *http.Request) {
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

	result, err := handler.service.GetProfileReviewPreview(r.Context(), principal, restaurantID)
	if err != nil {
		handler.mapError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *LeadReviewHandler) ReviewProfile(w http.ResponseWriter, r *http.Request) {
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
	request, ok := handler.decodeRequest(w, r)
	if !ok {
		return
	}
	if !strings.EqualFold(strings.TrimSpace(request.Status), leadreview.ProfileDraft) &&
		(request.ExpectedRestaurantUpdatedAt == nil || request.ExpectedProfileUpdatedAt == nil) {
		handler.writeError(w, http.StatusBadRequest, "expected_updated_at_required", "Both reviewed restaurant and profile versions are required.")
		return
	}

	result, err := handler.service.ReviewProfile(
		r.Context(),
		principal,
		restaurantID,
		request.Status,
		request.ExpectedRestaurantUpdatedAt,
		request.ExpectedProfileUpdatedAt,
	)
	if err != nil {
		handler.mapError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *LeadReviewHandler) SetDemoStatus(w http.ResponseWriter, r *http.Request) {
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
	request, ok := handler.decodeRequest(w, r)
	if !ok {
		return
	}
	if strings.EqualFold(strings.TrimSpace(request.Status), leadreview.DemoPublished) && request.ExpectedUpdatedAt == nil {
		handler.writeError(w, http.StatusBadRequest, "expected_updated_at_required", "expected_updated_at from the demo review preview is required for publication.")
		return
	}

	result, err := handler.service.SetDemoStatus(r.Context(), principal, demoSiteID, request.Status, request.ExpectedUpdatedAt)
	if err != nil {
		handler.mapError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *LeadReviewHandler) decodeRequest(w http.ResponseWriter, r *http.Request) (leadReviewRequest, bool) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 16<<10))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Could not read request body.")
		return leadReviewRequest{}, false
	}
	var request leadReviewRequest
	if len(body) == 0 || json.Unmarshal(body, &request) != nil || strings.TrimSpace(request.Status) == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must contain a valid status.")
		return leadReviewRequest{}, false
	}
	return request, true
}

func (handler *LeadReviewHandler) mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, leadreview.ErrForbidden):
		handler.writeError(w, http.StatusForbidden, "forbidden", "Internal administrator access is required.")
	case errors.Is(err, leadreview.ErrInvalidStatus):
		handler.writeError(w, http.StatusBadRequest, "invalid_status", "The requested review status is not supported.")
	case errors.Is(err, leadreview.ErrExpectedUpdatedAt):
		handler.writeError(w, http.StatusBadRequest, "expected_updated_at_required", "The reviewed artifact version is required.")
	case errors.Is(err, leadreview.ErrStaleReview):
		handler.writeError(w, http.StatusConflict, "stale_review", "The artifact changed after review; inspect it again before approving or publishing.")
	case errors.Is(err, leadreview.ErrNotFound):
		handler.writeError(w, http.StatusNotFound, "not_found", "The requested review target was not found.")
	case errors.Is(err, leadreview.ErrOCRNotVerified):
		handler.writeError(w, http.StatusConflict, "ocr_not_verified", "OCR must be verified before approval or publishing.")
	case errors.Is(err, leadreview.ErrProfileNotApproved):
		handler.writeError(w, http.StatusConflict, "profile_not_approved", "The restaurant profile must be approved before publishing.")
	case errors.Is(err, leadreview.ErrDemoExpired):
		handler.writeError(w, http.StatusConflict, "demo_expired", "The demo link has expired; regenerate the draft before publishing.")
	default:
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
	}
}
