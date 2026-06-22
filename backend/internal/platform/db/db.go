package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

var ErrNotConfigured = errors.New("database url is not configured")

type DB struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	if cfg.URL == "" {
		return &DB{}, nil
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	return &DB{pool: pool}, nil
}

func (d *DB) Configured() bool {
	return d != nil && d.pool != nil
}

func (d *DB) Ping(ctx context.Context) error {
	if !d.Configured() {
		return ErrNotConfigured
	}
	return d.pool.Ping(ctx)
}

func (d *DB) Close() {
	if d.Configured() {
		d.pool.Close()
	}
}

func (d *DB) Pool() *pgxpool.Pool {
	if d == nil {
		return nil
	}
	return d.pool
}
