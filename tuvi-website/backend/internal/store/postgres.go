package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Consultation struct {
	ID               uuid.UUID
	ConfirmationCode string
	SlotStart        time.Time
	SlotEnd          time.Time
	ProspectName     string
	ProspectEmail    string
	ProspectPhone    string
	Status           string
	GoogleEventID    string
	Source           string
	CreatedAt        time.Time
}

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) BookedSlotStarts(ctx context.Context, from, to time.Time) ([]time.Time, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT slot_start
		FROM consultations
		WHERE status = 'confirmed'
		  AND slot_start >= $1
		  AND slot_start < $2
		ORDER BY slot_start
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []time.Time
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		slots = append(slots, t)
	}
	return slots, rows.Err()
}

func (s *Store) IsSlotBooked(ctx context.Context, slotStart time.Time) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM consultations
			WHERE status = 'confirmed' AND slot_start = $1
		)
	`, slotStart).Scan(&exists)
	return exists, err
}

type InsertConsultationInput struct {
	ID               uuid.UUID
	ConfirmationCode string
	SlotStart        time.Time
	SlotEnd          time.Time
	ProspectName     string
	ProspectEmail    string
	ProspectPhone    string
	GoogleEventID    string
	Source           string
}

func (s *Store) InsertConsultation(ctx context.Context, tx pgx.Tx, input InsertConsultationInput) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO consultations (
			id, confirmation_code, slot_start, slot_end,
			prospect_name, prospect_email, prospect_phone,
			status, google_event_id, source
		) VALUES ($1,$2,$3,$4,$5,$6,$7,'confirmed',$8,$9)
	`,
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
	return err
}

func (s *Store) DeleteConsultation(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM consultations WHERE id = $1`, id)
	return err
}

func (s *Store) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return s.pool.Begin(ctx)
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}
