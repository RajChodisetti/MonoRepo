package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type UserHandler struct {
	users       auth.Repository
	restaurants restaurants.Repository
	memberships restaurants.MembershipRepository
	writeJSON   func(http.ResponseWriter, int, any)
	writeError  func(http.ResponseWriter, int, string, string)
}

func NewUserHandler(
	users auth.Repository,
	restaurantsRepo restaurants.Repository,
	memberships restaurants.MembershipRepository,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *UserHandler {
	return &UserHandler{
		users:       users,
		restaurants: restaurantsRepo,
		memberships: memberships,
		writeJSON:   writeJSON,
		writeError:  writeError,
	}
}

func (handler *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
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

	restaurantItems, err := handler.listRestaurantMemberships(r, principal)
	if err != nil {
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
		return
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{
		"id":          record.ID,
		"email":       record.Email,
		"full_name":   record.FullName,
		"role":        record.Role,
		"is_active":   record.IsActive,
		"created_at":  record.CreatedAt,
		"updated_at":  record.UpdatedAt,
		"restaurants": restaurantItems,
	})
}

func (handler *UserHandler) listRestaurantMemberships(r *http.Request, principal auth.Principal) ([]map[string]any, error) {
	memberships, err := handler.memberships.ListMembershipsByUser(r.Context(), principal.UserID)
	if err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		return []map[string]any{}, nil
	}

	restaurantIDs := make([]uuid.UUID, 0, len(memberships))
	memberRoles := make(map[uuid.UUID]string, len(memberships))
	for _, membership := range memberships {
		restaurantIDs = append(restaurantIDs, membership.RestaurantID)
		memberRoles[membership.RestaurantID] = membership.MemberRole
	}

	records, err := handler.restaurants.ListByIDs(r.Context(), restaurantIDs, restaurants.ListFilter{})
	if err != nil {
		return nil, err
	}

	recordsByID := make(map[uuid.UUID]restaurants.Restaurant, len(records))
	for _, record := range records {
		recordsByID[record.ID] = record
	}

	items := make([]map[string]any, 0, len(memberships))
	for _, membership := range memberships {
		record, ok := recordsByID[membership.RestaurantID]
		if !ok {
			continue
		}
		items = append(items, map[string]any{
			"id":          record.ID,
			"name":        record.Name,
			"email":       record.Email,
			"status":      record.Status,
			"member_role": memberRoles[record.ID],
		})
	}

	return items, nil
}
