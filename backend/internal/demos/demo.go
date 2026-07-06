package demos

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

type Site struct {
	ID            uuid.UUID
	RestaurantID  uuid.UUID
	Slug          string
	TokenHash     string
	Status        string
	PublicPayload json.RawMessage
	ExpiresAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateInput struct {
	RestaurantID  uuid.UUID
	Slug          string
	TokenHash     string
	Status        string
	PublicPayload json.RawMessage
	ExpiresAt     *time.Time
}

type Repository interface {
	GetBySlug(ctx context.Context, slug string) (Site, error)
	GetByID(ctx context.Context, id uuid.UUID) (Site, error)
	Create(ctx context.Context, input CreateInput) (Site, error)
}
