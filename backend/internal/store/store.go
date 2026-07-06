package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/metadata"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/reservations"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

const schemaBaselineKey = "schema_baseline"

type Store struct {
	database    *db.DB
	Metadata    metadata.Repository
	Users       auth.Repository
	Restaurants restaurants.Repository
	Memberships restaurants.MembershipRepository
	Demos       demos.Repository
	Campaigns   campaigns.Repository
	Profiles      profiles.Repository
	Reservations  reservations.Repository
}

func New(database *db.DB) *Store {
	pool := database.Pool()
	return NewWithRepositories(
		database,
		metadata.NewPostgres(pool),
		auth.NewPostgres(pool),
		restaurants.NewPostgres(pool),
		restaurants.NewMembershipPostgres(pool),
		demos.NewPostgres(pool),
		campaigns.NewPostgres(pool),
		profiles.NewPostgres(pool),
		reservations.NewPostgres(pool),
	)
}

func NewWithRepositories(
	database *db.DB,
	metadataRepo metadata.Repository,
	usersRepo auth.Repository,
	restaurantsRepo restaurants.Repository,
	membershipsRepo restaurants.MembershipRepository,
	demosRepo demos.Repository,
	campaignsRepo campaigns.Repository,
	profilesRepo profiles.Repository,
	reservationsRepo reservations.Repository,
) *Store {
	return &Store{
		database:    database,
		Metadata:    metadataRepo,
		Users:       usersRepo,
		Restaurants: restaurantsRepo,
		Memberships: membershipsRepo,
		Demos:       demosRepo,
		Campaigns:   campaignsRepo,
		Profiles:    profilesRepo,
		Reservations: reservationsRepo,
	}
}

func (store *Store) Pool() *pgxpool.Pool {
	if store.database == nil {
		return nil
	}
	return store.database.Pool()
}

func (store *Store) VerifyStartup(ctx context.Context) error {
	if err := store.VerifyFoundation(ctx); err != nil {
		return err
	}
	if err := store.verifyUsersTable(ctx); err != nil {
		return err
	}
	if err := store.verifyRestaurantTables(ctx); err != nil {
		return err
	}
	if err := store.verifyTableExists(ctx, "demo_sites", "demo_sites migration not applied: run make migrate-up"); err != nil {
		return err
	}
	return store.VerifyEmailCampaigns(ctx)
}

func (store *Store) VerifyEmailCampaigns(ctx context.Context) error {
	return store.verifyTableExists(ctx, "email_campaigns", "email campaigns migration not applied: run make migrate-up")
}

func (store *Store) VerifyFoundation(ctx context.Context) error {
	_, err := store.Metadata.Get(ctx, schemaBaselineKey)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("foundation migration not applied: %w", err)
	}
	return err
}

func (store *Store) verifyUsersTable(ctx context.Context) error {
	return store.verifyTableExists(ctx, "users", "users migration not applied: run make migrate-up")
}

func (store *Store) verifyRestaurantTables(ctx context.Context) error {
	if err := store.verifyTableExists(ctx, "restaurants", "restaurants migration not applied: run make migrate-up"); err != nil {
		return err
	}
	return store.verifyTableExists(ctx, "restaurant_members", "restaurant_members migration not applied: run make migrate-up")
}

func (store *Store) verifyTableExists(ctx context.Context, tableName, message string) error {
	const query = `
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = $1`

	var exists int
	err := store.database.Pool().QueryRow(ctx, query, tableName).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", message, repository.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("verify %s table: %w", tableName, err)
	}

	return nil
}
