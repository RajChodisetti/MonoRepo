package handlers

import (
	"errors"
	"net/http"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type AdminHandler struct {
	users      auth.Repository
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewAdminHandler(
	users auth.Repository,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *AdminHandler {
	return &AdminHandler{
		users:      users,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

func (handler *AdminHandler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	record, err := handler.users.GetByID(r.Context(), principal.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			handler.writeError(w, http.StatusUnauthorized, "unauthorized", "User account was not found.")
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
		return
	}

	if !record.IsActive {
		handler.writeError(w, http.StatusForbidden, "forbidden", "User account is inactive.")
		return
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{
		"id":        record.ID,
		"email":     record.Email,
		"full_name": record.FullName,
		"role":      record.Role,
	})
}
