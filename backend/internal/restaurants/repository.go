package restaurants

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (Restaurant, error)
	List(ctx context.Context, filter ListFilter) ([]Restaurant, error)
	ListByIDs(ctx context.Context, ids []uuid.UUID, filter ListFilter) ([]Restaurant, error)
	Create(ctx context.Context, input CreateInput) (Restaurant, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (Restaurant, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (Restaurant, error)
	Archive(ctx context.Context, id uuid.UUID) (Restaurant, error)
	MarkShownInterest(ctx context.Context, id uuid.UUID) (Restaurant, error)
}

type Member struct {
	ID           uuid.UUID
	RestaurantID uuid.UUID
	UserID       uuid.UUID
	MemberRole   string
	CreatedAt    time.Time
}

type MembershipRepository interface {
	HasMembership(ctx context.Context, userID, restaurantID uuid.UUID) (bool, error)
	ListRestaurantIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	ListMembershipsByUser(ctx context.Context, userID uuid.UUID) ([]Member, error)
	ListMembersByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]Member, error)
	AddMember(ctx context.Context, restaurantID, userID uuid.UUID, memberRole string) (Member, error)
}
