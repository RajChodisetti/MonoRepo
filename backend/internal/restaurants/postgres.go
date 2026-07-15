package restaurants

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	platformdb "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

const restaurantSelectColumns = `
	id, name, email, status, is_contacted, shown_interest,
	email_sent, email_send_count, last_email_sent_at, last_email_send_sequence,
	created_at, updated_at`

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func scanRestaurant(scanner interface {
	Scan(dest ...any) error
}) (Restaurant, error) {
	var record Restaurant
	err := scanner.Scan(
		&record.ID,
		&record.Name,
		&record.Email,
		&record.Status,
		&record.IsContacted,
		&record.ShownInterest,
		&record.EmailSent,
		&record.EmailSendCount,
		&record.LastEmailSentAt,
		&record.LastEmailSendSequence,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	return record, err
}

func (repo *Postgres) GetByID(ctx context.Context, id uuid.UUID) (Restaurant, error) {
	if repo.pool == nil {
		return Restaurant{}, fmt.Errorf("database pool is not configured")
	}

	query := `SELECT` + restaurantSelectColumns + ` FROM restaurants WHERE id = $1`

	record, err := scanRestaurant(repo.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Restaurant{}, repository.ErrNotFound
	}
	if err != nil {
		return Restaurant{}, fmt.Errorf("get restaurant: %w", err)
	}

	return record, nil
}

func (repo *Postgres) List(ctx context.Context, filter ListFilter) ([]Restaurant, error) {
	return repo.queryList(ctx, nil, filter)
}

func (repo *Postgres) ListByIDs(ctx context.Context, ids []uuid.UUID, filter ListFilter) ([]Restaurant, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return repo.queryList(ctx, ids, filter)
}

func (repo *Postgres) queryList(ctx context.Context, ids []uuid.UUID, filter ListFilter) ([]Restaurant, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	query := `SELECT` + restaurantSelectColumns + ` FROM restaurants WHERE 1=1`
	args := make([]any, 0, 8)
	argPos := 1

	if len(ids) > 0 {
		query += fmt.Sprintf(" AND id = ANY($%d)", argPos)
		args = append(args, ids)
		argPos++
	}

	if filter.Restaurant != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argPos)
		args = append(args, "%"+filter.Restaurant+"%")
		argPos++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, filter.Status)
		argPos++
	}

	if filter.IsContacted != nil {
		query += fmt.Sprintf(" AND is_contacted = $%d", argPos)
		args = append(args, *filter.IsContacted)
		argPos++
	}

	if filter.ShownInterest != nil {
		query += fmt.Sprintf(" AND shown_interest = $%d", argPos)
		args = append(args, *filter.ShownInterest)
		argPos++
	}

	if !filter.IncludeArchived {
		query += fmt.Sprintf(" AND status <> $%d", argPos)
		args = append(args, StatusArchived)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list restaurants: %w", err)
	}
	defer rows.Close()

	var records []Restaurant
	for rows.Next() {
		record, err := scanRestaurant(rows)
		if err != nil {
			return nil, fmt.Errorf("scan restaurant: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list restaurants rows: %w", err)
	}

	return records, nil
}

func (repo *Postgres) Create(ctx context.Context, input CreateInput) (Restaurant, error) {
	if repo.pool == nil {
		return Restaurant{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		INSERT INTO restaurants (name, email, status)
		VALUES ($1, $2, $3)
		RETURNING` + restaurantSelectColumns

	record, err := scanRestaurant(repo.pool.QueryRow(ctx, query, input.Name, input.Email, StatusLead))
	if err != nil {
		return Restaurant{}, fmt.Errorf("create restaurant: %w", err)
	}

	return record, nil
}

func (repo *Postgres) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (Restaurant, error) {
	if repo.pool == nil {
		return Restaurant{}, fmt.Errorf("database pool is not configured")
	}

	tx, err := repo.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Restaurant{}, fmt.Errorf("begin restaurant update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := platformdb.LockRestaurantWorkflow(ctx, tx, id); err != nil {
		return Restaurant{}, err
	}

	currentQuery := `SELECT` + restaurantSelectColumns + ` FROM restaurants WHERE id = $1 FOR UPDATE`
	current, err := scanRestaurant(tx.QueryRow(ctx, currentQuery, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Restaurant{}, repository.ErrNotFound
	}
	if err != nil {
		return Restaurant{}, fmt.Errorf("lock restaurant for update: %w", err)
	}

	updated := ApplyUpdateInput(current, input)
	identityChanged := current.Name != updated.Name || current.Email != updated.Email

	const query = `
		UPDATE restaurants
		SET name = $2,
		    email = $3,
		    status = $4,
		    is_contacted = $5,
		    shown_interest = $6,
		    updated_at = now()
		WHERE id = $1
		RETURNING` + restaurantSelectColumns

	record, err := scanRestaurant(tx.QueryRow(
		ctx,
		query,
		id,
		updated.Name,
		updated.Email,
		updated.Status,
		updated.IsContacted,
		updated.ShownInterest,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return Restaurant{}, repository.ErrNotFound
	}
	if err != nil {
		return Restaurant{}, fmt.Errorf("update restaurant: %w", err)
	}

	if identityChanged {
		if _, err := tx.Exec(ctx, `
			UPDATE restaurant_profiles
			SET review_status = 'draft',
			    reviewed_at = NULL,
			    reviewed_by = NULL,
			    updated_at = now()
			WHERE restaurant_id = $1`, id); err != nil {
			return Restaurant{}, fmt.Errorf("invalidate profile review after identity update: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE demo_sites
			SET status = 'draft',
			    published_at = NULL,
			    published_by = NULL,
			    updated_at = now()
			WHERE restaurant_id = $1 AND status = 'published'`, id); err != nil {
			return Restaurant{}, fmt.Errorf("invalidate demo publication after identity update: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE email_campaigns
			SET status = 'draft',
			    approved_at = NULL,
			    approved_by = NULL,
			    updated_at = now()
			WHERE restaurant_id = $1 AND status = 'approved'`, id); err != nil {
			return Restaurant{}, fmt.Errorf("invalidate campaign approval after identity update: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO job_runs (job_type, status, payload, idempotency_key, max_attempts)
			SELECT 'lead.prepare',
			       'queued',
			       jsonb_build_object('restaurant_id', rp.restaurant_id::text),
			       'lead.prepare:' || rp.restaurant_id::text || ':' ||
			         rp.ocr_input_fingerprint || ':' ||
			         lead_artifact_current_profile_fingerprint(rp.restaurant_id),
			       3
			FROM restaurant_profiles rp
			WHERE rp.restaurant_id = $1
			  AND rp.ocr_status = 'verified'
			  AND lead_artifact_current_profile_fingerprint(rp.restaurant_id) IS NOT NULL
			ON CONFLICT (idempotency_key) WHERE idempotency_key IS NOT NULL
			DO UPDATE SET
			    status = 'queued', payload = EXCLUDED.payload, attempts = 0,
			    max_attempts = EXCLUDED.max_attempts, last_error = NULL,
			    available_at = now(), locked_at = NULL, locked_by = NULL,
			    lease_expires_at = NULL, updated_at = now()
			WHERE job_runs.job_type = 'lead.prepare'
			  AND job_runs.status IN ('completed', 'failed')`, id); err != nil {
			return Restaurant{}, fmt.Errorf("enqueue lead preparation after identity update: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Restaurant{}, fmt.Errorf("commit restaurant update: %w", err)
	}

	return record, nil
}

func (repo *Postgres) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (Restaurant, error) {
	if repo.pool == nil {
		return Restaurant{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		UPDATE restaurants
		SET status = $2, updated_at = now()
		WHERE id = $1
		RETURNING` + restaurantSelectColumns

	record, err := scanRestaurant(repo.pool.QueryRow(ctx, query, id, status))
	if errors.Is(err, pgx.ErrNoRows) {
		return Restaurant{}, repository.ErrNotFound
	}
	if err != nil {
		return Restaurant{}, fmt.Errorf("update restaurant status: %w", err)
	}

	return record, nil
}

func (repo *Postgres) Archive(ctx context.Context, id uuid.UUID) (Restaurant, error) {
	return repo.UpdateStatus(ctx, id, StatusArchived)
}

func (repo *Postgres) MarkShownInterest(ctx context.Context, id uuid.UUID) (Restaurant, error) {
	if repo.pool == nil {
		return Restaurant{}, fmt.Errorf("database pool is not configured")
	}

	current, err := repo.GetByID(ctx, id)
	if err != nil {
		return Restaurant{}, err
	}

	status := StatusAfterShownInterest(current.Status)
	shownInterest := true

	const query = `
		UPDATE restaurants
		SET shown_interest = $2,
		    status = $3,
		    updated_at = now()
		WHERE id = $1
		RETURNING` + restaurantSelectColumns

	record, err := scanRestaurant(repo.pool.QueryRow(ctx, query, id, shownInterest, status))
	if errors.Is(err, pgx.ErrNoRows) {
		return Restaurant{}, repository.ErrNotFound
	}
	if err != nil {
		return Restaurant{}, fmt.Errorf("mark shown interest: %w", err)
	}

	return record, nil
}

type MembershipPostgres struct {
	pool *pgxpool.Pool
}

func NewMembershipPostgres(pool *pgxpool.Pool) *MembershipPostgres {
	return &MembershipPostgres{pool: pool}
}

func (repo *MembershipPostgres) HasMembership(ctx context.Context, userID, restaurantID uuid.UUID) (bool, error) {
	if repo.pool == nil {
		return false, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT 1
		FROM restaurant_members
		WHERE user_id = $1 AND restaurant_id = $2`

	var exists int
	err := repo.pool.QueryRow(ctx, query, userID, restaurantID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check restaurant membership: %w", err)
	}

	return true, nil
}

func (repo *MembershipPostgres) ListMembershipsByUser(ctx context.Context, userID uuid.UUID) ([]Member, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT id, restaurant_id, user_id, member_role, created_at
		FROM restaurant_members
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := repo.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list memberships by user: %w", err)
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var member Member
		if err := rows.Scan(&member.ID, &member.RestaurantID, &member.UserID, &member.MemberRole, &member.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan membership: %w", err)
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list memberships by user rows: %w", err)
	}

	return members, nil
}

func (repo *MembershipPostgres) ListRestaurantIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT restaurant_id
		FROM restaurant_members
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := repo.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list restaurant memberships: %w", err)
	}
	defer rows.Close()

	var restaurantIDs []uuid.UUID
	for rows.Next() {
		var restaurantID uuid.UUID
		if err := rows.Scan(&restaurantID); err != nil {
			return nil, fmt.Errorf("scan restaurant membership: %w", err)
		}
		restaurantIDs = append(restaurantIDs, restaurantID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list restaurant memberships rows: %w", err)
	}

	return restaurantIDs, nil
}

func (repo *MembershipPostgres) ListMembersByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]Member, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT id, restaurant_id, user_id, member_role, created_at
		FROM restaurant_members
		WHERE restaurant_id = $1
		ORDER BY created_at DESC`

	rows, err := repo.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("list restaurant members: %w", err)
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var member Member
		if err := rows.Scan(&member.ID, &member.RestaurantID, &member.UserID, &member.MemberRole, &member.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan restaurant member: %w", err)
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list restaurant members rows: %w", err)
	}

	return members, nil
}

func (repo *MembershipPostgres) AddMember(ctx context.Context, restaurantID, userID uuid.UUID, memberRole string) (Member, error) {
	if repo.pool == nil {
		return Member{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		INSERT INTO restaurant_members (restaurant_id, user_id, member_role)
		VALUES ($1, $2, $3)
		RETURNING id, restaurant_id, user_id, member_role, created_at`

	var member Member
	err := repo.pool.QueryRow(ctx, query, restaurantID, userID, memberRole).Scan(
		&member.ID,
		&member.RestaurantID,
		&member.UserID,
		&member.MemberRole,
		&member.CreatedAt,
	)
	if err != nil {
		return Member{}, fmt.Errorf("add restaurant member: %w", err)
	}

	return member, nil
}

var (
	_ Repository           = (*Postgres)(nil)
	_ MembershipRepository = (*MembershipPostgres)(nil)
)
