package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/seoreport"
)

const maxSEOUnlockBodyBytes = 8 << 10

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
	if unlockToken != "" {
		setNoStore(w)
	}
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
		if errors.Is(err, seoreport.ErrReportBusy) {
			w.Header().Set("Retry-After", "3")
			handler.writeError(w, http.StatusServiceUnavailable, "report_busy", "The live scan is busy. Please retry shortly.")
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to build SEO report.")
		return
	}
	handler.writeJSON(w, http.StatusOK, payload)
}

type seoUnlockRequestBody struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	PlaceID string `json:"placeId"`
}

type seoUnlockVerifyBody struct {
	Email   string `json:"email"`
	PlaceID string `json:"placeId"`
	OTP     string `json:"otp"`
}

// RequestUnlock handles POST /api/public/v1/seo/unlock/request
func (handler *SEOPublicHandler) RequestUnlock(w http.ResponseWriter, r *http.Request) {
	setNoStore(w)
	var body seoUnlockRequestBody
	if err := decodeSEOUnlockBody(w, r, &body); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_body", "Request body must be JSON.")
		return
	}
	payload, err := handler.service.RequestUnlockEmail(r.Context(), body.Name, body.Email, body.Phone, body.PlaceID)
	if err != nil {
		handler.writeUnlockError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, payload)
}

// VerifyUnlock handles POST /api/public/v1/seo/unlock/verify
func (handler *SEOPublicHandler) VerifyUnlock(w http.ResponseWriter, r *http.Request) {
	setNoStore(w)
	var body seoUnlockVerifyBody
	if err := decodeSEOUnlockBody(w, r, &body); err != nil {
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
// Verifies possession of an unexpired emailed token without recording marketing
// interest, then redirects to the report page.
func (handler *SEOPublicHandler) ClickUnlock(w http.ResponseWriter, r *http.Request) {
	setNoStore(w)
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

// Photo handles GET /api/public/v1/seo/photo?name=&max=
// Proxies Google Places photo media so the API key stays server-side.
func (handler *SEOPublicHandler) Photo(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_photo", "Photo name is required.")
		return
	}
	maxPx := 720
	if raw := strings.TrimSpace(r.URL.Query().Get("max")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			maxPx = n
		}
	}

	body, contentType, err := handler.service.FetchPlacePhoto(r.Context(), name, maxPx)
	if err != nil {
		if errors.Is(err, seoreport.ErrNotFound) {
			handler.writeError(w, http.StatusNotFound, "not_found", "Photo was not found.")
			return
		}
		handler.writeError(w, http.StatusBadGateway, "photo_failed", "Could not load listing photo.")
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func (handler *SEOPublicHandler) writeUnlockError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, seoreport.ErrInvalidName):
		handler.writeError(w, http.StatusBadRequest, "invalid_name", "A name between 2 and 100 characters is required.")
	case errors.Is(err, seoreport.ErrInvalidEmail):
		handler.writeError(w, http.StatusBadRequest, "invalid_email", "A valid email is required.")
	case errors.Is(err, seoreport.ErrInvalidPhone):
		handler.writeError(w, http.StatusBadRequest, "invalid_phone", "A valid phone number is required.")
	case errors.Is(err, seoreport.ErrInvalidOTP):
		handler.writeError(w, http.StatusUnauthorized, "invalid_otp", "Invalid or expired verification code.")
	case errors.Is(err, seoreport.ErrInvalidUnlock):
		handler.writeError(w, http.StatusNotFound, "invalid_unlock", "Unlock link is invalid or expired.")
	case errors.Is(err, seoreport.ErrUnlockRateLimit):
		w.Header().Set("Retry-After", strconv.Itoa(60))
		handler.writeError(w, http.StatusTooManyRequests, "unlock_rate_limited", "Please wait before requesting another code.")
	case errors.Is(err, seoreport.ErrNotFound):
		handler.writeError(w, http.StatusNotFound, "not_found", "Restaurant was not found.")
	case errors.Is(err, seoreport.ErrEmailSendFailed):
		handler.writeError(w, http.StatusBadGateway, "email_failed", "Could not send verification email.")
	case errors.Is(err, seoreport.ErrReportBusy):
		w.Header().Set("Retry-After", "3")
		handler.writeError(w, http.StatusServiceUnavailable, "report_busy", "The live scan is busy. Please retry shortly.")
	default:
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "Unlock request failed.")
	}
}

func decodeSEOUnlockBody(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxSEOUnlockBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func setNoStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}
