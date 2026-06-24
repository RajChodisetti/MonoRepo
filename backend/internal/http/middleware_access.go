package httpapi

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type RestaurantAccessOptions struct {
	PathParam string
}

func RequireRestaurantAccess(accessService *restaurants.Service, opts ...RestaurantAccessOptions) func(http.Handler) http.Handler {
	pathParam := "id"
	if len(opts) > 0 && opts[0].PathParam != "" {
		pathParam = opts[0].PathParam
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := auth.PrincipalFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
				return
			}

			restaurantID, err := uuid.Parse(r.PathValue(pathParam))
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
				return
			}

			if err := accessService.CanAccessRestaurant(r.Context(), principal, restaurantID); err != nil {
				if errors.Is(err, restaurants.ErrForbidden) {
					writeError(w, http.StatusForbidden, "forbidden", "You do not have access to this restaurants.")
					return
				}
				writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
				return
			}

			ctx := auth.WithRestaurantID(r.Context(), restaurantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
