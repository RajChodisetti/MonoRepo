package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func ConnectRequiredLogged(
	ctx context.Context,
	log *slog.Logger,
	cfg config.DatabaseConfig,
	readyTimeout time.Duration,
) (*DB, error) {
	log.InfoContext(ctx, "database_connecting")

	database, err := ConnectRequired(ctx, cfg, readyTimeout)
	if err != nil {
		log.ErrorContext(ctx, "database_connection_failed", "error", err)
		return nil, err
	}

	log.InfoContext(ctx, "database_connected_successfully",
		"max_conns", cfg.MaxConns,
		"min_conns", cfg.MinConns,
	)

	return database, nil
}

func CloseLogged(ctx context.Context, log *slog.Logger, database *DB) {
	if database == nil || !database.Configured() {
		return
	}
	database.Close()
	log.InfoContext(ctx, "database_disconnected")
}
