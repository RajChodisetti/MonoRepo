package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (repo *Postgres) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	if repo.pool == nil {
		return User{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1`

	return repo.scanUser(repo.pool.QueryRow(ctx, query, id))
}

func (repo *Postgres) GetByEmail(ctx context.Context, email string) (User, error) {
	if repo.pool == nil {
		return User{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)`

	return repo.scanUser(repo.pool.QueryRow(ctx, query, strings.TrimSpace(email)))
}

func (repo *Postgres) Create(ctx context.Context, input CreateInput) (User, error) {
	if repo.pool == nil {
		return User{}, fmt.Errorf("database pool is not configured")
	}
	if !ValidRole(input.Role) {
		return User{}, fmt.Errorf("invalid user role %q", input.Role)
	}

	const query = `
		INSERT INTO users (email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, password_hash, full_name, role, is_active, created_at, updated_at`

	record, err := repo.scanUser(repo.pool.QueryRow(
		ctx,
		query,
		strings.TrimSpace(input.Email),
		input.PasswordHash,
		strings.TrimSpace(input.FullName),
		input.Role,
	))
	if err != nil {
		return User{}, err
	}

	return record, nil
}

func (repo *Postgres) scanUser(row pgx.Row) (User, error) {
	var record User
	err := row.Scan(
		&record.ID,
		&record.Email,
		&record.PasswordHash,
		&record.FullName,
		&record.Role,
		&record.IsActive,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, repository.ErrNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, fmt.Errorf("email already exists: %w", repository.ErrConflict)
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}

	return record, nil
}

var _ Repository = (*Postgres)(nil)
