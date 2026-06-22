package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/repository"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/repositories/metadata"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/repositories/user"
)

const schemaBaselineKey = "schema_baseline"

type Store struct {
	database *db.DB
	Metadata metadata.Repository
	Users    user.Repository
}

func New(database *db.DB) *Store {
	pool := database.Pool()
	return &Store{
		database: database,
		Metadata: metadata.NewPostgres(pool),
		Users:    user.NewPostgres(pool),
	}
}

func (store *Store) VerifyStartup(ctx context.Context) error {
	if err := store.VerifyFoundation(ctx); err != nil {
		return err
	}
	return store.verifyUsersTable(ctx)
}

func (store *Store) VerifyFoundation(ctx context.Context) error {
	_, err := store.Metadata.Get(ctx, schemaBaselineKey)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("foundation migration not applied: %w", err)
	}
	return err
}

func (store *Store) verifyUsersTable(ctx context.Context) error {
	const query = `
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'users'`

	var exists int
	err := store.database.Pool().QueryRow(ctx, query).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("users migration not applied: run make migrate-up: %w", repository.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("verify users table: %w", err)
	}

	return nil
}
