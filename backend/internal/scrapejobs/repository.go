package scrapejobs

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateOrGetActive(ctx context.Context, city, cityKey, niche string, createdBy uuid.UUID) (Job, bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (Job, error)
	ListRecent(ctx context.Context, limit int) ([]Job, error)
	RetryFailed(ctx context.Context, id uuid.UUID) (Job, error)
}
