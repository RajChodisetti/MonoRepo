package demos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

var ErrDemoNotFound = errors.New("demo not found")

type Service struct {
	demos    Repository
	access   *restaurants.Service
	tokenTTL time.Duration
}

func NewService(demos Repository, accessService *restaurants.Service, tokenTTL time.Duration) *Service {
	return &Service{
		demos:    demos,
		access:   accessService,
		tokenTTL: tokenTTL,
	}
}

func (service *Service) ResolvePublicDemo(ctx context.Context, slug, token string) (PublicDemoPayload, error) {
	slug = strings.TrimSpace(slug)
	token = strings.TrimSpace(token)
	if slug == "" || token == "" {
		return PublicDemoPayload{}, ErrDemoNotFound
	}

	record, err := service.demos.GetBySlug(ctx, slug)
	if err != nil {
		return PublicDemoPayload{}, ErrDemoNotFound
	}

	if record.Status != StatusPublished {
		return PublicDemoPayload{}, ErrDemoNotFound
	}

	if record.ExpiresAt != nil && time.Now().After(*record.ExpiresAt) {
		return PublicDemoPayload{}, ErrDemoNotFound
	}

	if err := CheckDemoToken(record.TokenHash, token); err != nil {
		return PublicDemoPayload{}, ErrDemoNotFound
	}

	return MapPublicPayload(record.PublicPayload), nil
}

type CreateDemoResult struct {
	ID     uuid.UUID `json:"id"`
	Slug   string    `json:"slug"`
	Token  string    `json:"token"`
	Status string    `json:"status"`
}

type CreateDemoInput struct {
	Slug          string
	PublicPayload json.RawMessage
	Status        string
}

func (service *Service) CreateDemoSite(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID, input CreateDemoInput) (CreateDemoResult, error) {
	if _, err := service.access.MustAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return CreateDemoResult{}, err
	}
	if !auth.IsInternalAdmin(principal.Role) {
		return CreateDemoResult{}, restaurants.ErrForbidden
	}

	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		return CreateDemoResult{}, fmt.Errorf("slug is required")
	}

	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = StatusDraft
	}

	token, err := GenerateDemoToken()
	if err != nil {
		return CreateDemoResult{}, err
	}

	tokenHash, err := HashDemoToken(token)
	if err != nil {
		return CreateDemoResult{}, err
	}

	publicPayload := input.PublicPayload
	if len(publicPayload) == 0 {
		publicPayload = DefaultPublicPayload()
	}

	var expiresAt *time.Time
	if service.tokenTTL > 0 {
		expiry := time.Now().Add(service.tokenTTL)
		expiresAt = &expiry
	}

	record, err := service.demos.Create(ctx, CreateInput{
		RestaurantID:  restaurantID,
		Slug:          slug,
		TokenHash:     tokenHash,
		Status:        status,
		PublicPayload: publicPayload,
		ExpiresAt:     expiresAt,
	})
	if err != nil {
		return CreateDemoResult{}, err
	}

	return CreateDemoResult{
		ID:     record.ID,
		Slug:   record.Slug,
		Token:  token,
		Status: record.Status,
	}, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrDemoNotFound)
}
