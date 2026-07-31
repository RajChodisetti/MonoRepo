package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/seoreport"
)

// SEOPublicHandler serves public SEO search and report endpoints.
type SEOPublicHandler struct {
	service      *seoreport.Service
	publicWebURL string
	writeJSON    func(http.ResponseWriter, int, any)
	writeError   func(http.ResponseWriter, int, string, string)
}

// NewSEOPublicHandler constructs the public SEO handler.
func NewSEOPublicHandler(
	service *seoreport.Service,
	publicWebURL string,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *SEOPublicHandler {
	return &SEOPublicHandler{
		service:      service,
		publicWebURL: strings.TrimRight(strings.TrimSpace(publicWebURL), "/"),
		writeJSON:    writeJSON,
		writeError:   writeError,
	}
}

// Search handles GET /api/public/v1/seo/search?q=&location=
func (handler *SEOPublicHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	location := strings.TrimSpace(r.URL.Query().Get("location"))
	payload, err := handler.service.SearchRestaurants(r.Context(), q, location, 8)
	if err != nil {
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "Restaurant search failed.")
		return
	}
	handler.writeJSON(w, http.StatusOK, payload)
}

// Report handles GET /api/public/v1/seo/report/{place_id}
func (handler *SEOPublicHandler) Report(w http.ResponseWriter, r *http.Request) {
	placeID := strings.TrimSpace(r.PathValue("place_id"))
	if placeID == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_place_id", "Google place id is required.")
		return
	}

	unlockToken := strings.TrimSpace(r.URL.Query().Get("unlock"))
	var (
		payload seoreport.ReportResponse
		err     error
	)
	if unlockToken != "" {
		payload, err = handler.service.GetReportUnlocked(r.Context(), placeID, unlockToken)
	} else {
		payload, err = handler.service.GetReport(r.Context(), placeID)
	}
	if err != nil {
		if errors.Is(err, seoreport.ErrNotFound) {
			handler.writeError(w, http.StatusNotFound, "not_found", "Restaurant was not found.")
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to build SEO report.")
		return
	}
	handler.writeJSON(w, http.StatusOK, payload)
}

type seoUnlockRequestBody struct {
	Email   string `json:"email"`
	PlaceID string `json:"placeId"`
	OTP     string `json:"otp"`
}

// RequestUnlock handles POST /api/public/v1/seo/unlock/request
func (handler *SEOPublicHandler) RequestUnlock(w http.ResponseWriter, r *http.Request) {
	var body seoUnlockRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_body", "Request body must be JSON.")
		return
	}
	payload, err := handler.service.RequestUnlockEmail(r.Context(), body.Email, body.PlaceID)
	if err != nil {
		handler.writeUnlockError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, payload)
}

// VerifyUnlock handles POST /api/public/v1/seo/unlock/verify
func (handler *SEOPublicHandler) VerifyUnlock(w http.ResponseWriter, r *http.Request) {
	var body seoUnlockRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_body", "Request body must be JSON.")
		return
	}
	payload, err := handler.service.VerifyUnlockOTP(r.Context(), body.Email, body.PlaceID, body.OTP)
	if err != nil {
		handler.writeUnlockError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, payload)
}

// ClickUnlock handles GET /api/public/v1/seo/unlock/click/{token}
// Marks interested=true then redirects to the marketing report page.
func (handler *SEOPublicHandler) ClickUnlock(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.PathValue("token"))
	rec, place, err := handler.service.ConfirmUnlockClick(r.Context(), token)
	if err != nil {
		handler.writeUnlockError(w, err)
		return
	}

	base := handler.publicWebURL
	if base == "" {
		base = "http://localhost:3000"
	}
	dest := base + "/report/" + url.PathEscape(place.PlaceID) + "?unlock=" + url.QueryEscape(rec.UnlockToken)
	http.Redirect(w, r, dest, http.StatusFound)
}

func (handler *SEOPublicHandler) writeUnlockError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, seoreport.ErrInvalidEmail):
		handler.writeError(w, http.StatusBadRequest, "invalid_email", "A valid email is required.")
	case errors.Is(err, seoreport.ErrInvalidOTP):
		handler.writeError(w, http.StatusUnauthorized, "invalid_otp", "Invalid or expired verification code.")
	case errors.Is(err, seoreport.ErrInvalidUnlock):
		handler.writeError(w, http.StatusNotFound, "invalid_unlock", "Unlock link is invalid or expired.")
	case errors.Is(err, seoreport.ErrNotFound):
		handler.writeError(w, http.StatusNotFound, "not_found", "Restaurant was not found.")
	case errors.Is(err, seoreport.ErrEmailSendFailed):
		handler.writeError(w, http.StatusBadGateway, "email_failed", "Could not send verification email.")
	default:
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "Unlock request failed.")
	}
}
