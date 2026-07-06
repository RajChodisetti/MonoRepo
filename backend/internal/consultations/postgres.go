package consultations

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (repo *Postgres) BookedSlotStarts(ctx context.Context, from, to time.Time) ([]time.Time, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	rows, err := repo.pool.Query(ctx, `
		SELECT slot_start
		FROM company_consultations
		WHERE status = 'confirmed'
		  AND slot_start >= $1
		  AND slot_start < $2
		ORDER BY slot_start`,
		from,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("booked consultation slots: %w", err)
	}
	defer rows.Close()

	var slots []time.Time
	for rows.Next() {
		var slot time.Time
		if err := rows.Scan(&slot); err != nil {
			return nil, fmt.Errorf("scan consultation slot: %w", err)
		}
		slots = append(slots, slot)
	}
	return slots, rows.Err()
}

func (repo *Postgres) IsSlotBooked(ctx context.Context, slotStart time.Time) (bool, error) {
	if repo.pool == nil {
		return false, fmt.Errorf("database pool is not configured")
	}

	var exists bool
	err := repo.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM company_consultations
			WHERE status = 'confirmed' AND slot_start = $1
		)`,
		slotStart,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("consultation slot booked: %w", err)
	}
	return exists, nil
}

func (repo *Postgres) Insert(ctx context.Context, tx pgx.Tx, input InsertInput) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO company_consultations (
			id, confirmation_code, slot_start, slot_end,
			prospect_name, prospect_email, prospect_phone,
			status, google_event_id, source
		) VALUES ($1,$2,$3,$4,$5,$6,$7,'confirmed',$8,$9)`,
		input.ID,
		input.ConfirmationCode,
		input.SlotStart,
		input.SlotEnd,
		input.ProspectName,
		input.ProspectEmail,
		input.ProspectPhone,
		input.GoogleEventID,
		input.Source,
	)
	if err != nil {
		return fmt.Errorf("insert consultation: %w", err)
	}
	return nil
}

func (repo *Postgres) Delete(ctx context.Context, id uuid.UUID) error {
	if repo.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}
	_, err := repo.pool.Exec(ctx, `DELETE FROM company_consultations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete consultation: %w", err)
	}
	return nil
}

func (repo *Postgres) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}
	return repo.pool.Begin(ctx)
}
