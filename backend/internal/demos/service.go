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

	payload := MapPublicPayload(record.PublicPayload)
	// The reservation identity comes from the resolved demo row, never from
	// mutable public_payload supplied by an administrator or ingestion job.
	payload.RestaurantID = record.RestaurantID.String()
	return payload, nil
}

type CreateDemoResult struct {
	ID     uuid.UUID `json:"id"`
	Slug   string    `json:"slug"`
	Token  string    `json:"token"`
	Status string    `json:"status"`
}

// ReviewPreview is the exact allowlisted payload an administrator reviews
// before publication. It deliberately excludes both the stored token hash and
// the campaign's long-lived bearer token.
type ReviewPreview struct {
	DemoSiteID    uuid.UUID         `json:"demo_site_id"`
	RestaurantID  uuid.UUID         `json:"restaurant_id"`
	Slug          string            `json:"slug"`
	Status        string            `json:"status"`
	ExpiresAt     *time.Time        `json:"expires_at,omitempty"`
	PublicPayload PublicDemoPayload `json:"public_payload"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type CreateDemoInput struct {
	Slug          string
	PublicPayload json.RawMessage
	Status        string
}

func (service *Service) GetReviewPreview(
	ctx context.Context,
	principal auth.Principal,
	demoSiteID uuid.UUID,
) (ReviewPreview, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return ReviewPreview{}, restaurants.ErrForbidden
	}

	record, err := service.demos.GetByID(ctx, demoSiteID)
	if err != nil {
		return ReviewPreview{}, err
	}
	if err := service.access.CanAccessRestaurant(ctx, principal, record.RestaurantID); err != nil {
		return ReviewPreview{}, err
	}

	payload := MapPublicPayload(record.PublicPayload)
	payload.RestaurantID = record.RestaurantID.String()
	return ReviewPreview{
		DemoSiteID:    record.ID,
		RestaurantID:  record.RestaurantID,
		Slug:          record.Slug,
		Status:        record.Status,
		ExpiresAt:     record.ExpiresAt,
		PublicPayload: payload,
		CreatedAt:     record.CreatedAt,
		UpdatedAt:     record.UpdatedAt,
	}, nil
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
	if status == StatusPublished {
		return CreateDemoResult{}, fmt.Errorf("demo sites must be created as drafts and published through the audited status endpoint")
	}
	if status != StatusDraft {
		return CreateDemoResult{}, fmt.Errorf("unsupported demo status")
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
