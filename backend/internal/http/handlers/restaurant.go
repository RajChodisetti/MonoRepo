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
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type RestaurantHandler struct {
	accessService *restaurants.Service
	writeJSON     func(http.ResponseWriter, int, any)
	writeError    func(http.ResponseWriter, int, string, string)
}

func NewRestaurantHandler(
	accessService *restaurants.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *RestaurantHandler {
	return &RestaurantHandler{
		accessService: accessService,
		writeJSON:     writeJSON,
		writeError:    writeError,
	}
}

func (handler *RestaurantHandler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	filter, err := parseListFilter(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	records, err := handler.accessService.ListAccessibleRestaurants(r.Context(), principal, filter)
	if err != nil {
		handler.mapAccessError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		items = append(items, restaurantResponse(record))
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (handler *RestaurantHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	record, err := handler.accessService.GetRestaurant(r.Context(), principal, restaurantID)
	if err != nil {
		handler.mapAccessError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, restaurantResponse(record))
}

type createRestaurantRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (handler *RestaurantHandler) Create(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Could not read request body.")
		return
	}

	var request createRestaurantRequest
	if err := json.Unmarshal(body, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}

	name := strings.TrimSpace(request.Name)
	email := strings.TrimSpace(request.Email)
	if name == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant name is required.")
		return
	}
	if !isValidLeadEmail(email) {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "A valid restaurant email is required.")
		return
	}

	record, err := handler.accessService.CreateRestaurant(r.Context(), principal, restaurants.CreateInput{
		Name:  name,
		Email: email,
	})
	if err != nil {
		handler.mapAccessError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusCreated, restaurantResponse(record))
}

type updateRestaurantRequest struct {
	Name          *string `json:"name"`
	Email         *string `json:"email"`
	IsContacted   *bool   `json:"is_contacted"`
	ShownInterest *bool   `json:"shown_interest"`
}

func (handler *RestaurantHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var request updateRestaurantRequest
	if err := json.Unmarshal(body, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}

	if request.Name != nil {
		trimmed := strings.TrimSpace(*request.Name)
		request.Name = &trimmed
	}
	if request.Email != nil {
		trimmed := strings.TrimSpace(*request.Email)
		request.Email = &trimmed
		if !isValidLeadEmail(trimmed) {
			handler.writeError(w, http.StatusBadRequest, "invalid_request", "A valid restaurant email is required.")
			return
		}
	}

	record, err := handler.accessService.UpdateRestaurant(r.Context(), principal, restaurantID, restaurants.UpdateInput{
		Name:          request.Name,
		Email:         request.Email,
		IsContacted:   request.IsContacted,
		ShownInterest: request.ShownInterest,
	})
	if err != nil {
		handler.mapAccessError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, restaurantResponse(record))
}

type updateRestaurantStatusRequest struct {
	Status string `json:"status"`
}

func (handler *RestaurantHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	var request updateRestaurantStatusRequest
	if err := json.Unmarshal(body, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}

	status := strings.TrimSpace(request.Status)
	if status == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "status is required.")
		return
	}

	record, err := handler.accessService.UpdateRestaurantStatus(r.Context(), principal, restaurantID, status)
	if err != nil {
		handler.mapAccessError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, restaurantResponse(record))
}

func (handler *RestaurantHandler) Archive(w http.ResponseWriter, r *http.Request) {
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

	record, err := handler.accessService.ArchiveRestaurant(r.Context(), principal, restaurantID)
	if err != nil {
		handler.mapAccessError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, restaurantResponse(record))
}

type addMemberRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	MemberRole string    `json:"member_role"`
}

func (handler *RestaurantHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := handler.accessService.ListMembers(r.Context(), principal, restaurantID)
	if err != nil {
		handler.mapAccessError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(members))
	for _, member := range members {
		items = append(items, map[string]any{
			"id":            member.ID,
			"restaurant_id": member.RestaurantID,
			"user_id":       member.UserID,
			"member_role":   member.MemberRole,
			"created_at":    member.CreatedAt,
		})
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (handler *RestaurantHandler) AddMember(w http.ResponseWriter, r *http.Request) {
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

	var request addMemberRequest
	if err := json.Unmarshal(body, &request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}
	if request.UserID == uuid.Nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "user_id is required.")
		return
	}

	memberRole := strings.TrimSpace(request.MemberRole)
	if memberRole == "" {
		memberRole = "owner"
	}

	member, err := handler.accessService.AddMember(r.Context(), principal, restaurantID, request.UserID, memberRole)
	if err != nil {
		handler.mapAccessError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusCreated, map[string]any{
		"id":            member.ID,
		"restaurant_id": member.RestaurantID,
		"user_id":       member.UserID,
		"member_role":   member.MemberRole,
		"created_at":    member.CreatedAt,
	})
}

func parseListFilter(r *http.Request) (restaurants.ListFilter, error) {
	query := r.URL.Query()
	filter := restaurants.ListFilter{
		Restaurant: strings.TrimSpace(query.Get("restaurant")),
		Status:     strings.TrimSpace(query.Get("status")),
	}

	if value := strings.TrimSpace(query.Get("is_contacted")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return restaurants.ListFilter{}, errors.New("is_contacted must be true or false.")
		}
		filter.IsContacted = &parsed
	}

	if value := strings.TrimSpace(query.Get("shown_interest")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return restaurants.ListFilter{}, errors.New("shown_interest must be true or false.")
		}
		filter.ShownInterest = &parsed
	}

	if value := strings.TrimSpace(query.Get("include_archived")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return restaurants.ListFilter{}, errors.New("include_archived must be true or false.")
		}
		filter.IncludeArchived = parsed
	}

	if filter.Status != "" && !restaurants.IsValidStatus(filter.Status) {
		return restaurants.ListFilter{}, errors.New("status must be a valid restaurant lifecycle value.")
	}

	return filter, nil
}

func restaurantResponse(record restaurants.Restaurant) map[string]any {
	return map[string]any{
		"id":             record.ID,
		"name":           record.Name,
		"email":          record.Email,
		"status":         record.Status,
		"is_contacted":   record.IsContacted,
		"shown_interest": record.ShownInterest,
		"created_at":     record.CreatedAt,
		"updated_at":     record.UpdatedAt,
	}
}

func isValidLeadEmail(email string) bool {
	at := strings.LastIndex(email, "@")
	if at < 1 || at == len(email)-1 {
		return false
	}
	dot := strings.LastIndex(email[at:], ".")
	return dot > 1
}

func restaurantIDFromRequest(r *http.Request) (uuid.UUID, error) {
	if restaurantID, ok := auth.RestaurantIDFromContext(r.Context()); ok {
		return restaurantID, nil
	}
	return uuid.Parse(r.PathValue("id"))
}

func (handler *RestaurantHandler) mapAccessError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, restaurants.ErrForbidden):
		handler.writeError(w, http.StatusForbidden, "forbidden", "You do not have access to this restaurants.")
	case errors.Is(err, restaurants.ErrInvalidStatus):
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "status must be a valid restaurant lifecycle value.")
	case errors.Is(err, restaurants.ErrInvalidRequest):
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Request validation failed.")
	case errors.Is(err, repository.ErrNotFound):
		handler.writeError(w, http.StatusNotFound, "not_found", "Restaurant was not found.")
	default:
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
	}
}
