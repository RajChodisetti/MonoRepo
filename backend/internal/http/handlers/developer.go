package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/developer"
)

type DeveloperHandler struct {
	service    *developer.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

type executeSQLRequest struct {
	Query string `json:"query"`
}

func NewDeveloperHandler(
	service *developer.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *DeveloperHandler {
	return &DeveloperHandler{
		service:    service,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

func (handler *DeveloperHandler) Schema(w http.ResponseWriter, r *http.Request) {
	tables, err := handler.service.Schema(r.Context())
	if err != nil {
		handler.mapError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, map[string]any{"tables": tables})
}

func (handler *DeveloperHandler) ExecuteSQL(w http.ResponseWriter, r *http.Request) {
	var req executeSQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}

	result, err := handler.service.Execute(r.Context(), req.Query)
	if err != nil {
		handler.mapError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *DeveloperHandler) mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, developer.ErrDatabaseUnavailable):
		handler.writeError(w, http.StatusServiceUnavailable, "database_unavailable", "Database is not configured for developer queries.")
	case errors.Is(err, developer.ErrQueryRequired):
		handler.writeError(w, http.StatusBadRequest, "query_required", "SQL query is required.")
	case errors.Is(err, developer.ErrQueryTooLong):
		handler.writeError(w, http.StatusBadRequest, "query_too_long", "SQL query is too long.")
	case errors.Is(err, developer.ErrReadOnlyRequired):
		handler.writeError(w, http.StatusBadRequest, "read_only_required", "Only read-only SELECT, WITH, SHOW, and EXPLAIN queries are allowed.")
	default:
		message := err.Error()
		if len(message) > 600 {
			message = message[:600] + "..."
		}
		handler.writeError(w, http.StatusBadRequest, "query_failed", message)
	}
}
