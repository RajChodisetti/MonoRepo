package metadata

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

func (repo *Postgres) Get(ctx context.Context, key string) (Record, error) {
	if repo.pool == nil {
		return Record{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT key, value, created_at, updated_at
		FROM app_metadata
		WHERE key = $1`

	var record Record
	err := repo.pool.QueryRow(ctx, query, key).Scan(
		&record.Key,
		&record.Value,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, repository.ErrNotFound
	}
	if err != nil {
		return Record{}, fmt.Errorf("get app_metadata %q: %w", key, err)
	}

	return record, nil
}

var _ Repository = (*Postgres)(nil)
