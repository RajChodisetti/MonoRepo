package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/scrapejobs"
)

type ScrapeJobHandler struct {
	service    *scrapejobs.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewScrapeJobHandler(
	service *scrapejobs.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *ScrapeJobHandler {
	return &ScrapeJobHandler{
		service:    service,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

type createScrapeJobRequest struct {
	City  string `json:"city"`
	Niche string `json:"niche"`
}

func (handler *ScrapeJobHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 16<<10))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Could not read request body.")
		return
	}
	var request createScrapeJobRequest
	if len(body) == 0 || json.Unmarshal(body, &request) != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}
	if strings.TrimSpace(request.City) == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "city is required.")
		return
	}

	result, err := handler.service.Trigger(r.Context(), principal, scrapejobs.CreateInput{
		City:  request.City,
		Niche: request.Niche,
	})
	if err != nil {
		handler.mapError(w, err)
		return
	}

	status := http.StatusAccepted
	if !result.Created {
		status = http.StatusOK
	}
	handler.writeJSON(w, status, result)
}

func (handler *ScrapeJobHandler) Get(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(r.PathValue("id")))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Scrape job id must be a valid UUID.")
		return
	}

	job, err := handler.service.Get(r.Context(), principal, id)
	if err != nil {
		handler.mapError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, job)
}

func (handler *ScrapeJobHandler) Retry(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(r.PathValue("id")))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Scrape job id must be a valid UUID.")
		return
	}

	job, err := handler.service.RetryFailed(r.Context(), principal, id)
	if err != nil {
		handler.mapError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusAccepted, job)
}

func (handler *ScrapeJobHandler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	limit := 25
	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", "limit must be a positive integer.")
			return
		}
		limit = parsed
	}

	jobs, err := handler.service.List(r.Context(), principal, limit)
	if err != nil {
		handler.mapError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
}

func (handler *ScrapeJobHandler) mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, scrapejobs.ErrForbidden):
		handler.writeError(w, http.StatusForbidden, "forbidden", "Internal administrator access is required.")
	case errors.Is(err, scrapejobs.ErrInvalidCity):
		handler.writeError(w, http.StatusUnprocessableEntity, "unsupported_city", "Supported cities are Adelaide, Brisbane, Melbourne, Perth, and Sydney.")
	case errors.Is(err, scrapejobs.ErrInvalidNiche):
		handler.writeError(w, http.StatusUnprocessableEntity, "unsupported_niche", "Supported niches are restaurant, dentist, and plumber.")
	case errors.Is(err, scrapejobs.ErrNotFound):
		handler.writeError(w, http.StatusNotFound, "scrape_job_not_found", "Scrape job was not found.")
	case errors.Is(err, scrapejobs.ErrNotFailed):
		handler.writeError(w, http.StatusConflict, "scrape_job_not_failed", "Only a failed scrape job can be retried.")
	case errors.Is(err, scrapejobs.ErrActiveJobExists):
		handler.writeError(w, http.StatusConflict, "active_scrape_job_exists", "Another active scrape job already exists for this city and niche.")
	default:
		handler.writeError(w, http.StatusInternalServerError, "scrape_job_failed", "The scrape job request could not be completed.")
	}
}
