package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type AuthHandler struct {
	service    *auth.Service
	appEnv     string
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewAuthHandler(
	service *auth.Service,
	appEnv string,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *AuthHandler {
	return &AuthHandler{
		service:    service,
		appEnv:     appEnv,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Signup allows restaurant_owner and developer (local/test) roles.
// Internal admin accounts must be created via make seed-admin.
func (handler *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var body signupRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}

	role := body.Role
	if role == "" {
		role = auth.RoleRestaurantOwner
	}
	if !auth.SignupAllowedRole(role, handler.appEnv) {
		handler.writeError(w, http.StatusForbidden, "forbidden", "This role cannot be assigned during signup.")
		return
	}

	result, err := handler.service.Signup(r.Context(), auth.SignupInput{
		Email:    body.Email,
		Password: body.Password,
		FullName: body.FullName,
		Role:     role,
	})
	if err != nil {
		handler.mapServiceError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusCreated, result)
}

func (handler *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}

	result, err := handler.service.Login(r.Context(), auth.LoginInput{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		handler.mapServiceError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, result)
}

func (handler *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{
		"user_id": principal.UserID,
		"email":   principal.Email,
		"role":    principal.Role,
	})
}

func (handler *AuthHandler) mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		handler.writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid credentials.")
	case errors.Is(err, auth.ErrInactiveUser):
		handler.writeError(w, http.StatusForbidden, "forbidden", "User account is inactive.")
	case errors.Is(err, auth.ErrInvalidInput):
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, repository.ErrConflict):
		handler.writeError(w, http.StatusConflict, "conflict", "Email is already registered.")
	default:
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
	}
}
