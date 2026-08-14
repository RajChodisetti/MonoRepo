package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreachaccounts"
)

type OutreachEmailAccountsHandler struct {
	service    *outreachaccounts.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewOutreachEmailAccountsHandler(
	service *outreachaccounts.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *OutreachEmailAccountsHandler {
	return &OutreachEmailAccountsHandler{service: service, writeJSON: writeJSON, writeError: writeError}
}

func (handler *OutreachEmailAccountsHandler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	result, err := handler.service.List(r.Context(), principal)
	if err != nil {
		handler.writeAccountError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *OutreachEmailAccountsHandler) Create(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	var input outreachaccounts.CreateInput
	if err := decodeEmailAccountJSON(r, &input); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := handler.service.Create(r.Context(), principal, input)
	if err != nil {
		handler.writeAccountError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusCreated, result)
}

func (handler *OutreachEmailAccountsHandler) Update(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Email account id must be a valid UUID.")
		return
	}
	var input outreachaccounts.UpdateInput
	if err := decodeEmailAccountJSON(r, &input); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := handler.service.Update(r.Context(), principal, id, input)
	if err != nil {
		handler.writeAccountError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, result)
}

func decodeEmailAccountJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errors.New("request body must be valid JSON with only supported fields")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func (handler *OutreachEmailAccountsHandler) writeAccountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, outreachaccounts.ErrForbidden):
		handler.writeError(w, http.StatusForbidden, "forbidden", "Internal administrator access is required.")
	case errors.Is(err, outreachaccounts.ErrInvalid):
		handler.writeError(w, http.StatusBadRequest, "invalid_email_account", err.Error())
	case errors.Is(err, outreachaccounts.ErrDuplicate):
		handler.writeError(w, http.StatusConflict, "duplicate_email_account", "That account key or mailbox is already configured. Environment accounts take precedence.")
	case errors.Is(err, outreachaccounts.ErrNotFound):
		handler.writeError(w, http.StatusNotFound, "email_account_not_found", "Email account was not found.")
	case errors.Is(err, outreachaccounts.ErrEncryptionUnavailable):
		handler.writeError(w, http.StatusServiceUnavailable, "credential_encryption_unavailable", "Secure credential storage is not configured.")
	default:
		handler.writeError(w, http.StatusInternalServerError, "email_accounts_unavailable", "Email accounts are unavailable.")
	}
}
