package auth

import (
	"context"

	"github.com/google/uuid"
)

type Principal struct {
	UserID uuid.UUID
	Email  string
	Role   string
}

type principalKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalKey{}).(Principal)
	return principal, ok
}

type restaurantIDKey struct{}

func WithRestaurantID(ctx context.Context, restaurantID uuid.UUID) context.Context {
	return context.WithValue(ctx, restaurantIDKey{}, restaurantID)
}

func RestaurantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	restaurantID, ok := ctx.Value(restaurantIDKey{}).(uuid.UUID)
	return restaurantID, ok
}
