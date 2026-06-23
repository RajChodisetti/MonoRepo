package demos

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (repo *Postgres) GetBySlug(ctx context.Context, slug string) (Site, error) {
	if repo.pool == nil {
		return Site{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT id, restaurant_id, slug, token_hash, status, public_payload, expires_at, created_at, updated_at
		FROM demo_sites
		WHERE slug = $1`

	var record Site
	err := repo.pool.QueryRow(ctx, query, slug).Scan(
		&record.ID,
		&record.RestaurantID,
		&record.Slug,
		&record.TokenHash,
		&record.Status,
		&record.PublicPayload,
		&record.ExpiresAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Site{}, repository.ErrNotFound
	}
	if err != nil {
		return Site{}, fmt.Errorf("get demo site: %w", err)
	}

	return record, nil
}

func (repo *Postgres) Create(ctx context.Context, input CreateInput) (Site, error) {
	if repo.pool == nil {
		return Site{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		INSERT INTO demo_sites (restaurant_id, slug, token_hash, status, public_payload, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, restaurant_id, slug, token_hash, status, public_payload, expires_at, created_at, updated_at`

	var record Site
	err := repo.pool.QueryRow(
		ctx,
		query,
		input.RestaurantID,
		input.Slug,
		input.TokenHash,
		input.Status,
		input.PublicPayload,
		input.ExpiresAt,
	).Scan(
		&record.ID,
		&record.RestaurantID,
		&record.Slug,
		&record.TokenHash,
		&record.Status,
		&record.PublicPayload,
		&record.ExpiresAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return Site{}, fmt.Errorf("create demo site: %w", err)
	}

	return record, nil
}

var _ Repository = (*Postgres)(nil)
