package consultations

import (
	"context"
	"errors"
	"fmt"
	"sort"
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

func (repo *Postgres) BookedIntervals(ctx context.Context, from, to time.Time) ([]BookedInterval, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	rows, err := repo.pool.Query(ctx, `
		SELECT slot_start, slot_end
		FROM company_consultations
		WHERE status = 'confirmed'
		  AND slot_start < $2
		  AND slot_end > $1
		ORDER BY slot_start`,
		from,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("booked consultation intervals: %w", err)
	}
	defer rows.Close()

	var intervals []BookedInterval
	for rows.Next() {
		var interval BookedInterval
		if err := rows.Scan(&interval.SlotStart, &interval.SlotEnd); err != nil {
			return nil, fmt.Errorf("scan consultation interval: %w", err)
		}
		intervals = append(intervals, interval)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("booked consultation interval rows: %w", err)
	}
	return intervals, nil
}

func (repo *Postgres) HasConfirmedOverlap(
	ctx context.Context,
	slotStart, slotEnd time.Time,
) (bool, error) {
	if repo.pool == nil {
		return false, fmt.Errorf("database pool is not configured")
	}

	var exists bool
	err := repo.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM company_consultations
			WHERE status = 'confirmed'
			  AND slot_start < $2
			  AND slot_end > $1
		)`,
		slotStart,
		slotEnd,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("consultation interval overlap: %w", err)
	}
	return exists, nil
}

func (repo *Postgres) CalendarRevision(ctx context.Context, monthStart time.Time) (int64, error) {
	if repo.pool == nil {
		return 0, fmt.Errorf("database pool is not configured")
	}

	var revision int64
	err := repo.pool.QueryRow(ctx, `
		SELECT revision
		FROM company_consultation_calendar_months
		WHERE month = $1::date`,
		monthStart.Format("2006-01-02"),
	).Scan(&revision)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("consultation calendar revision: %w", err)
	}
	return revision, nil
}

func (repo *Postgres) SlotOverrides(ctx context.Context, from, to time.Time) ([]SlotOverride, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	rows, err := repo.pool.Query(ctx, `
		SELECT slot_start, is_available
		FROM company_consultation_slot_overrides
		WHERE slot_start >= $1 AND slot_start < $2
		ORDER BY slot_start`,
		from,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("consultation slot overrides: %w", err)
	}
	defer rows.Close()

	var overrides []SlotOverride
	for rows.Next() {
		var override SlotOverride
		if err := rows.Scan(&override.SlotStart, &override.IsAvailable); err != nil {
			return nil, fmt.Errorf("scan consultation slot override: %w", err)
		}
		overrides = append(overrides, override)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("consultation slot overrides rows: %w", err)
	}
	return overrides, nil
}

func (repo *Postgres) IsSlotEnabled(ctx context.Context, slotStart time.Time) (bool, error) {
	if repo.pool == nil {
		return false, fmt.Errorf("database pool is not configured")
	}

	var enabled bool
	err := repo.pool.QueryRow(ctx, `
		SELECT COALESCE(
			(SELECT is_available
			 FROM company_consultation_slot_overrides
			 WHERE slot_start = $1),
			true
		)`,
		slotStart,
	).Scan(&enabled)
	if err != nil {
		return false, fmt.Errorf("consultation slot enabled: %w", err)
	}
	return enabled, nil
}

func (repo *Postgres) ReplaceMonthSlotOverrides(
	ctx context.Context,
	monthStart, monthEnd time.Time,
	inputs []SlotOverrideInput,
	expectedRevision int64,
	updatedBy uuid.UUID,
) (int64, error) {
	if repo.pool == nil {
		return 0, fmt.Errorf("database pool is not configured")
	}

	ordered := append([]SlotOverrideInput(nil), inputs...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].SlotStart.Before(ordered[j].SlotStart)
	})

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin consultation calendar transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	monthDate := monthStart.Format("2006-01-02")
	if _, err := tx.Exec(ctx, `
		INSERT INTO company_consultation_calendar_months (month, revision)
		VALUES ($1::date, 0)
		ON CONFLICT (month) DO NOTHING`,
		monthDate,
	); err != nil {
		return 0, fmt.Errorf("materialize consultation calendar month: %w", err)
	}

	var currentRevision int64
	if err := tx.QueryRow(ctx, `
		SELECT revision
		FROM company_consultation_calendar_months
		WHERE month = $1::date
		FOR UPDATE`,
		monthDate,
	).Scan(&currentRevision); err != nil {
		return 0, fmt.Errorf("lock consultation calendar month: %w", err)
	}
	if currentRevision != expectedRevision {
		return 0, ErrCalendarRevisionConflict
	}

	// The month-row lock coordinates this replacement with bookings. Removing
	// the whole month first also eliminates overrides left behind by an older
	// business-hours grid before the current future grid is reinserted below.
	if _, err := tx.Exec(ctx, `
		DELETE FROM company_consultation_slot_overrides
		WHERE slot_start >= $1 AND slot_start < $2`,
		monthStart,
		monthEnd,
	); err != nil {
		return 0, fmt.Errorf("delete obsolete consultation slot overrides: %w", err)
	}

	for _, input := range ordered {
		_, err := tx.Exec(ctx, `
			INSERT INTO company_consultation_slot_overrides (
				slot_start, is_available, updated_by, updated_at
			) VALUES ($1, $2, $3, now())
			ON CONFLICT (slot_start) DO UPDATE SET
				is_available = EXCLUDED.is_available,
				updated_by = EXCLUDED.updated_by,
				updated_at = now()`,
			input.SlotStart,
			input.IsAvailable,
			updatedBy,
		)
		if err != nil {
			return 0, fmt.Errorf("upsert consultation slot override: %w", err)
		}
	}

	var newRevision int64
	if err := tx.QueryRow(ctx, `
		UPDATE company_consultation_calendar_months
		SET revision = revision + 1,
		    updated_by = $2,
		    updated_at = now()
		WHERE month = $1::date
		RETURNING revision`,
		monthDate,
		updatedBy,
	).Scan(&newRevision); err != nil {
		return 0, fmt.Errorf("increment consultation calendar revision: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit consultation calendar: %w", err)
	}
	return newRevision, nil
}

func (repo *Postgres) InsertIfAvailable(ctx context.Context, input InsertInput) (bool, error) {
	if repo.pool == nil {
		return false, fmt.Errorf("database pool is not configured")
	}

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin consultation booking transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	monthDate := input.SlotStart.Format("2006-01") + "-01"
	if _, err := tx.Exec(ctx, `
		INSERT INTO company_consultation_calendar_months (month, revision)
		VALUES ($1::date, 0)
		ON CONFLICT (month) DO NOTHING`,
		monthDate,
	); err != nil {
		return false, fmt.Errorf("materialize consultation booking month: %w", err)
	}
	var revision int64
	if err := tx.QueryRow(ctx, `
		SELECT revision
		FROM company_consultation_calendar_months
		WHERE month = $1::date
		FOR SHARE`,
		monthDate,
	).Scan(&revision); err != nil {
		return false, fmt.Errorf("lock consultation booking month: %w", err)
	}

	// Materializing the legacy default makes the row lock coordinate a booking
	// with a concurrent admin calendar save, including the first save for a month.
	if _, err := tx.Exec(ctx, `
		INSERT INTO company_consultation_slot_overrides (slot_start, is_available)
		VALUES ($1, true)
		ON CONFLICT (slot_start) DO NOTHING`,
		input.SlotStart,
	); err != nil {
		return false, fmt.Errorf("materialize consultation slot override: %w", err)
	}

	var enabled bool
	if err := tx.QueryRow(ctx, `
		SELECT is_available
		FROM company_consultation_slot_overrides
		WHERE slot_start = $1
		FOR UPDATE`,
		input.SlotStart,
	).Scan(&enabled); err != nil {
		return false, fmt.Errorf("lock consultation slot override: %w", err)
	}
	if !enabled {
		return false, nil
	}

	if err := insertConsultation(ctx, tx, input); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit consultation booking: %w", err)
	}
	return true, nil
}

func insertConsultation(ctx context.Context, tx pgx.Tx, input InsertInput) error {
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
