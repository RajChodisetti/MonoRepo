package restaurants

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

var (
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidStatus  = errors.New("invalid status")
	ErrInvalidRequest = errors.New("invalid request")
)

type Service struct {
	restaurants Repository
	memberships MembershipRepository
}

func NewService(restaurants Repository, memberships MembershipRepository) *Service {
	return &Service{
		restaurants: restaurants,
		memberships: memberships,
	}
}

func (service *Service) CanAccessRestaurant(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) error {
	if auth.IsInternalAdmin(principal.Role) {
		return nil
	}

	if !auth.IsRestaurantOwner(principal.Role) {
		return ErrForbidden
	}

	allowed, err := service.memberships.HasMembership(ctx, principal.UserID, restaurantID)
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if !allowed {
		return ErrForbidden
	}

	return nil
}

func (service *Service) MustAccessRestaurant(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) (uuid.UUID, error) {
	if err := service.CanAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return uuid.Nil, err
	}
	return restaurantID, nil
}

func (service *Service) GetRestaurant(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) (Restaurant, error) {
	if _, err := service.MustAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return Restaurant{}, err
	}

	record, err := service.restaurants.GetByID(ctx, restaurantID)
	if err != nil {
		return Restaurant{}, err
	}

	if record.Status == StatusArchived && !auth.IsInternalAdmin(principal.Role) {
		return Restaurant{}, repository.ErrNotFound
	}

	return record, nil
}

func (service *Service) ListAccessibleRestaurants(ctx context.Context, principal auth.Principal, filter ListFilter) ([]Restaurant, error) {
	if filter.Status != "" && !IsValidStatus(filter.Status) {
		return nil, ErrInvalidStatus
	}

	if auth.IsInternalAdmin(principal.Role) {
		return service.restaurants.List(ctx, filter)
	}

	if !auth.IsRestaurantOwner(principal.Role) {
		return nil, ErrForbidden
	}

	restaurantIDs, err := service.memberships.ListRestaurantIDsByUser(ctx, principal.UserID)
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}

	return service.restaurants.ListByIDs(ctx, restaurantIDs, filter)
}

func (service *Service) CreateRestaurant(ctx context.Context, principal auth.Principal, input CreateInput) (Restaurant, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Restaurant{}, ErrForbidden
	}

	return service.restaurants.Create(ctx, input)
}

func (service *Service) UpdateRestaurant(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID, input UpdateInput) (Restaurant, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Restaurant{}, ErrForbidden
	}

	if _, err := service.MustAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return Restaurant{}, err
	}

	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Restaurant{}, ErrInvalidRequest
	}
	if input.Email != nil && strings.TrimSpace(*input.Email) == "" {
		return Restaurant{}, ErrInvalidRequest
	}

	current, err := service.restaurants.GetByID(ctx, restaurantID)
	if err != nil {
		return Restaurant{}, err
	}

	patch := input
	if input.IsContacted != nil && *input.IsContacted {
		status := StatusAfterContacted(current.Status)
		patch.Status = &status
	}
	if input.ShownInterest != nil && *input.ShownInterest {
		status := StatusAfterShownInterest(current.Status)
		patch.Status = &status
	}

	return service.restaurants.Update(ctx, restaurantID, patch)
}

func (service *Service) UpdateRestaurantStatus(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID, status string) (Restaurant, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Restaurant{}, ErrForbidden
	}

	if !IsValidStatus(status) {
		return Restaurant{}, ErrInvalidStatus
	}

	if _, err := service.MustAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return Restaurant{}, err
	}

	return service.restaurants.UpdateStatus(ctx, restaurantID, status)
}

func (service *Service) ArchiveRestaurant(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) (Restaurant, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Restaurant{}, ErrForbidden
	}

	if _, err := service.MustAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return Restaurant{}, err
	}

	return service.restaurants.Archive(ctx, restaurantID)
}

func (service *Service) ListMembers(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) ([]Member, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return nil, ErrForbidden
	}

	if _, err := service.MustAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return nil, err
	}

	return service.memberships.ListMembersByRestaurant(ctx, restaurantID)
}

func (service *Service) AddMember(ctx context.Context, principal auth.Principal, restaurantID, userID uuid.UUID, memberRole string) (Member, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Member{}, ErrForbidden
	}

	if _, err := service.MustAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return Member{}, err
	}

	if _, err := service.restaurants.GetByID(ctx, restaurantID); err != nil {
		return Member{}, err
	}

	return service.memberships.AddMember(ctx, restaurantID, userID, memberRole)
}

func IsNotFound(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}
