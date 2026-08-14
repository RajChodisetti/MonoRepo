package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/consultations"
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
	database      *db.DB
	Metadata      metadata.Repository
	Users         auth.Repository
	Restaurants   restaurants.Repository
	Memberships   restaurants.MembershipRepository
	Demos         demos.Repository
	Campaigns     campaigns.Repository
	Profiles      profiles.Repository
	Reservations  reservations.Repository
	Consultations consultations.Repository
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
		consultations.NewPostgres(pool),
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
	consultationsRepo ...consultations.Repository,
) *Store {
	consultationRepo := consultations.Repository(&consultations.Mock{})
	if len(consultationsRepo) > 0 && consultationsRepo[0] != nil {
		consultationRepo = consultationsRepo[0]
	}

	return &Store{
		database:      database,
		Metadata:      metadataRepo,
		Users:         usersRepo,
		Restaurants:   restaurantsRepo,
		Memberships:   membershipsRepo,
		Demos:         demosRepo,
		Campaigns:     campaignsRepo,
		Profiles:      profilesRepo,
		Reservations:  reservationsRepo,
		Consultations: consultationRepo,
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
	if err := store.VerifyEmailCampaigns(ctx); err != nil {
		return err
	}
	if err := store.VerifyLeadWorkflow(ctx); err != nil {
		return err
	}
	if err := store.verifyTableExists(ctx, "company_consultations", "company_consultations migration not applied: run make migrate-up"); err != nil {
		return err
	}
	if err := store.verifyTableExists(
		ctx,
		"company_consultation_slot_overrides",
		"company_consultation_slot_overrides migration not applied: run make migrate-up",
	); err != nil {
		return err
	}
	if err := store.verifyTableExists(
		ctx,
		"company_consultation_calendar_months",
		"company_consultation_calendar_months migration not applied: run make migrate-up",
	); err != nil {
		return err
	}
	if err := store.verifyConstraintExists(
		ctx,
		"company_consultations",
		"company_consultations_confirmed_no_overlap",
		"company consultation overlap migration not applied: run make migrate-up",
	); err != nil {
		return err
	}
	return store.verifyConstraintExists(
		ctx,
		"company_consultations",
		"company_consultations_valid_interval",
		"company consultation valid-interval migration not applied: run make migrate-up",
	)
}

func (store *Store) VerifyEmailCampaigns(ctx context.Context) error {
	if err := store.verifyTableExists(ctx, "email_campaigns", "email campaigns migration not applied: run make migrate-up"); err != nil {
		return err
	}
	if err := store.verifyTableExists(ctx, "email_delivery_attempts", "outreach email quota migration not applied: run make migrate-up"); err != nil {
		return err
	}
	if err := store.verifyTableExists(ctx, "email_messages", "email messages migration not applied: run make migrate-up"); err != nil {
		return err
	}
	if err := store.verifyColumnExists(ctx, "email_messages", "received_at"); err != nil {
		return err
	}
	if err := store.verifyColumnExists(ctx, "outreach_inbound_sync", "last_success_at"); err != nil {
		return err
	}
	return store.verifyColumnExists(ctx, "outreach_email_accounts", "ramp_day")
}

func (store *Store) VerifyLeadWorkflow(ctx context.Context) error {
	if err := store.verifyTableExists(ctx, "scrape_jobs", "city scrape migration not applied: run make migrate-up"); err != nil {
		return err
	}
	if err := store.verifyTableExists(ctx, "restaurant_media_assets", "restaurant media migration not applied: run make migrate-up"); err != nil {
		return err
	}
	requiredColumns := []struct {
		table  string
		column string
	}{
		{table: "restaurant_profiles", column: "reviewed_by"},
		{table: "demo_sites", column: "published_by"},
		{table: "demo_sites", column: "source_profile_fingerprint"},
		{table: "email_campaigns", column: "source_profile_fingerprint"},
		{table: "job_runs", column: "locked_by"},
		{table: "job_runs", column: "lease_expires_at"},
		{table: "email_delivery_attempts", column: "lease_expires_at"},
		{table: "email_delivery_attempts", column: "recipient_email"},
		{table: "email_tracking_tokens", column: "recipient_email"},
		{table: "restaurants", column: "last_email_recipient"},
	}
	for _, required := range requiredColumns {
		if err := store.verifyColumnExists(ctx, required.table, required.column); err != nil {
			return err
		}
	}
	for _, functionName := range []string{
		"lead_artifact_current_public_payload(uuid)",
		"lead_artifact_current_profile_fingerprint(uuid)",
	} {
		if err := store.verifyFunctionExists(ctx, functionName); err != nil {
			return err
		}
	}
	for _, constraint := range []struct {
		table string
		name  string
	}{
		{table: "scrape_jobs", name: "scrape_jobs_waiting_reason_check"},
		{table: "scrape_job_cells", name: "scrape_job_cells_status_check"},
	} {
		var coverageStateSupported bool
		err := store.Pool().QueryRow(ctx, `
			SELECT position('coverage_incomplete' in pg_get_constraintdef(constraint_row.oid)) > 0
			FROM pg_constraint AS constraint_row
			JOIN pg_class AS table_row ON table_row.oid = constraint_row.conrelid
			WHERE table_row.relname = $1 AND constraint_row.conname = $2`,
			constraint.table,
			constraint.name,
		).Scan(&coverageStateSupported)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("lead workflow migration not applied: %s.%s missing; run make migrate-up", constraint.table, constraint.name)
		}
		if err != nil {
			return fmt.Errorf("verify constraint %s.%s: %w", constraint.table, constraint.name, err)
		}
		if !coverageStateSupported {
			return fmt.Errorf("lead workflow migration 000023 not applied: coverage_incomplete missing from %s; run make migrate-up", constraint.name)
		}
	}
	return nil
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

func (store *Store) verifyConstraintExists(
	ctx context.Context,
	tableName, constraintName, message string,
) error {
	const query = `
		SELECT 1
		FROM pg_constraint AS constraint_row
		JOIN pg_class AS table_row ON table_row.oid = constraint_row.conrelid
		JOIN pg_namespace AS namespace_row ON namespace_row.oid = table_row.relnamespace
		WHERE namespace_row.nspname = 'public'
		  AND table_row.relname = $1
		  AND constraint_row.conname = $2`

	var exists int
	err := store.Pool().QueryRow(ctx, query, tableName, constraintName).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", message, repository.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("verify constraint %s.%s: %w", tableName, constraintName, err)
	}
	return nil
}

func (store *Store) verifyColumnExists(ctx context.Context, tableName, columnName string) error {
	const query = `
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2`
	var one int
	err := store.Pool().QueryRow(ctx, query, tableName, columnName).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("lead workflow migration not applied: missing %s.%s; run make migrate-up", tableName, columnName)
	}
	if err != nil {
		return fmt.Errorf("verify column %s.%s: %w", tableName, columnName, err)
	}
	return nil
}

func (store *Store) verifyFunctionExists(ctx context.Context, functionName string) error {
	const query = `SELECT to_regprocedure($1) IS NOT NULL`
	var exists bool
	if err := store.Pool().QueryRow(ctx, query, functionName).Scan(&exists); err != nil {
		return fmt.Errorf("verify function %s: %w", functionName, err)
	}
	if !exists {
		return fmt.Errorf("lead workflow migration not applied: missing function %s; run make migrate-up", functionName)
	}
	return nil
}
