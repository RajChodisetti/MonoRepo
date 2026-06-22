package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

var (
	ErrNotConfigured = errors.New("database url is not configured")
	ErrNotReady      = errors.New("database is not ready")
)

type DB struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return &DB{}, nil
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("open database pool: %w", err)
	}

	return &DB{pool: pool}, nil
}

func ConnectRequired(ctx context.Context, cfg config.DatabaseConfig, readyTimeout time.Duration) (*DB, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("%w: DATABASE_URL is required", ErrNotConfigured)
	}

	database, err := Connect(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := waitUntilReady(ctx, database, readyTimeout); err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}

func waitUntilReady(ctx context.Context, database *DB, timeout time.Duration) error {
	if timeout <= 0 {
		return pingOrNotReady(database.Ping(ctx))
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastErr error
	for {
		if err := database.Ping(waitCtx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-waitCtx.Done():
			if lastErr == nil {
				lastErr = waitCtx.Err()
			}
			return fmt.Errorf("%w: %v", ErrNotReady, lastErr)
		case <-ticker.C:
		}
	}
}

func pingOrNotReady(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrNotReady, err)
}

func (database *DB) Configured() bool {
	return database != nil && database.pool != nil
}

func (database *DB) Ping(ctx context.Context) error {
	if !database.Configured() {
		return ErrNotConfigured
	}
	return database.pool.Ping(ctx)
}

func (database *DB) Close() {
	if database.Configured() {
		database.pool.Close()
	}
}

func (database *DB) Pool() *pgxpool.Pool {
	if database == nil {
		return nil
	}
	return database.pool
}
