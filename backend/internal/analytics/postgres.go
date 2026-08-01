package analytics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (repo *Postgres) CreateSession(
	ctx context.Context,
	demoSiteID *uuid.UUID,
	restaurantID uuid.UUID,
	templateID, sessionTokenHash string,
) (Session, error) {
	if repo.pool == nil {
		return Session{}, fmt.Errorf("database pool is not configured")
	}
	const query = `
		INSERT INTO demo_sessions (demo_site_id, restaurant_id, template_id, session_token_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, demo_site_id, restaurant_id, template_id, started_at, last_seen_at, ended_at, duration_seconds`
	var session Session
	err := repo.pool.QueryRow(ctx, query, demoSiteID, restaurantID, templateID, sessionTokenHash).Scan(
		&session.ID,
		&session.DemoSiteID,
		&session.RestaurantID,
		&session.TemplateID,
		&session.StartedAt,
		&session.LastSeenAt,
		&session.EndedAt,
		&session.DurationSeconds,
	)
	if err != nil {
		return Session{}, fmt.Errorf("create demo session: %w", err)
	}
	session.Transcript = []Transcript{}
	return session, nil
}

func (repo *Postgres) TouchSession(ctx context.Context, sessionID uuid.UUID, sessionToken string, activeSeconds int, ended bool) error {
	if repo.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin demo session touch: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var tokenHash string
	if err := tx.QueryRow(ctx, `
		SELECT session_token_hash
		FROM demo_sessions
		WHERE id = $1
		FOR UPDATE`, sessionID).Scan(&tokenHash); errors.Is(err, pgx.ErrNoRows) {
		return ErrSessionNotFound
	} else if err != nil {
		return fmt.Errorf("lock demo session: %w", err)
	}
	if demos.CheckDemoToken(tokenHash, sessionToken) != nil {
		return ErrSessionNotFound
	}
	now := time.Now().UTC()
	var endedAt any
	if ended {
		endedAt = now
	}
	_, err = tx.Exec(ctx, `
		UPDATE demo_sessions
		SET last_seen_at = $2,
		    ended_at = CASE WHEN $3::timestamptz IS NOT NULL THEN $3 ELSE ended_at END,
		    duration_seconds = GREATEST(duration_seconds, $4),
		    updated_at = $2
		WHERE id = $1`, sessionID, now, endedAt, activeSeconds)
	if err != nil {
		return fmt.Errorf("touch demo session: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit demo session touch: %w", err)
	}
	return nil
}

func (repo *Postgres) AddTranscript(
	ctx context.Context,
	sessionID uuid.UUID,
	sessionToken, role, content string,
) error {
	if err := repo.TouchSession(ctx, sessionID, sessionToken, 0, false); err != nil {
		return err
	}
	_, err := repo.pool.Exec(ctx, `
		INSERT INTO demo_session_transcripts (session_id, role, content)
		VALUES ($1, $2, $3)`, sessionID, role, content)
	if err != nil {
		return fmt.Errorf("add demo session transcript: %w", err)
	}
	return nil
}

func (repo *Postgres) ListSessions(ctx context.Context, restaurantID uuid.UUID) ([]Session, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}
	rows, err := repo.pool.Query(ctx, `
		SELECT id, demo_site_id, restaurant_id, template_id, started_at, last_seen_at, ended_at, duration_seconds
		FROM demo_sessions
		WHERE restaurant_id = $1
		ORDER BY started_at DESC
		LIMIT 200`, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("list restaurant demo sessions: %w", err)
	}
	defer rows.Close()
	sessions := make([]Session, 0)
	byID := make(map[uuid.UUID]int)
	for rows.Next() {
		var session Session
		if err := rows.Scan(
			&session.ID,
			&session.DemoSiteID,
			&session.RestaurantID,
			&session.TemplateID,
			&session.StartedAt,
			&session.LastSeenAt,
			&session.EndedAt,
			&session.DurationSeconds,
		); err != nil {
			return nil, fmt.Errorf("scan restaurant demo session: %w", err)
		}
		session.Transcript = []Transcript{}
		byID[session.ID] = len(sessions)
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list restaurant demo session rows: %w", err)
	}
	if len(sessions) == 0 {
		return sessions, nil
	}
	ids := make([]uuid.UUID, 0, len(sessions))
	for _, session := range sessions {
		ids = append(ids, session.ID)
	}
	turnRows, err := repo.pool.Query(ctx, `
		SELECT id, session_id, role, content, occurred_at
		FROM demo_session_transcripts
		WHERE session_id = ANY($1::uuid[])
		ORDER BY occurred_at ASC, id ASC`, ids)
	if err != nil {
		return nil, fmt.Errorf("list demo session transcripts: %w", err)
	}
	defer turnRows.Close()
	for turnRows.Next() {
		var turn Transcript
		var sessionID uuid.UUID
		if err := turnRows.Scan(&turn.ID, &sessionID, &turn.Role, &turn.Content, &turn.OccurredAt); err != nil {
			return nil, fmt.Errorf("scan demo session transcript: %w", err)
		}
		if index, ok := byID[sessionID]; ok {
			sessions[index].Transcript = append(sessions[index].Transcript, turn)
		}
	}
	return sessions, turnRows.Err()
}

var _ Repository = (*Postgres)(nil)
