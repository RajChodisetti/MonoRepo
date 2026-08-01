package demos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

func (repo *Postgres) GetByID(ctx context.Context, id uuid.UUID) (Site, error) {
	if repo.pool == nil {
		return Site{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT id, restaurant_id, slug, token_hash, status, public_payload, expires_at, created_at, updated_at
		FROM demo_sites
		WHERE id = $1`

	var record Site
	err := repo.pool.QueryRow(ctx, query, id).Scan(
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

func (repo *Postgres) UpdateTokenHash(ctx context.Context, id uuid.UUID, tokenHash string) error {
	if repo.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}

	const query = `
		UPDATE demo_sites
		SET token_hash = $2, updated_at = now()
		WHERE id = $1 AND status = 'draft'`

	tag, err := repo.pool.Exec(ctx, query, id, tokenHash)
	if err != nil {
		return fmt.Errorf("update demo token hash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (repo *Postgres) ListByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]Site, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT id, restaurant_id, slug, token_hash, status, public_payload, expires_at, created_at, updated_at
		FROM demo_sites
		WHERE restaurant_id = $1
		ORDER BY created_at DESC`

	rows, err := repo.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("list demo sites: %w", err)
	}
	defer rows.Close()

	sites := make([]Site, 0)
	for rows.Next() {
		var record Site
		if err := rows.Scan(
			&record.ID,
			&record.RestaurantID,
			&record.Slug,
			&record.TokenHash,
			&record.Status,
			&record.PublicPayload,
			&record.ExpiresAt,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan demo site: %w", err)
		}
		sites = append(sites, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list demo sites rows: %w", err)
	}

	return sites, nil
}

func (repo *Postgres) BuildPublicPayload(ctx context.Context, restaurantID uuid.UUID) (json.RawMessage, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	var payload []byte
	err := repo.pool.QueryRow(ctx, `SELECT lead_artifact_current_public_payload($1)`, restaurantID).Scan(&payload)
	if err != nil {
		return nil, fmt.Errorf("build restaurant demo payload: %w", err)
	}
	if len(payload) == 0 {
		return nil, repository.ErrNotFound
	}
	return json.RawMessage(payload), nil
}

var _ Repository = (*Postgres)(nil)
