package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/migrations"
)

const databaseReadyTimeout = 30 * time.Second

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: migrate up|down")
	}

	direction := os.Args[1]
	if direction != "up" && direction != "down" {
		log.Fatalf("invalid migration direction %q; use up or down", direction)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	appLog := logger.New(cfg.Logging)
	database, err := db.ConnectRequiredLogged(ctx, appLog, cfg.Database, databaseReadyTimeout)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.CloseLogged(ctx, appLog, database)

	dir, err := migrationsDir()
	if err != nil {
		log.Fatalf("resolve migrations dir: %v", err)
	}

	switch direction {
	case "up":
		err = migrations.ApplyUp(ctx, database.Pool(), dir)
	case "down":
		err = migrations.ApplyDown(ctx, database.Pool(), dir)
	}
	if err != nil {
		log.Fatalf("migration %s failed: %v", direction, err)
	}

	fmt.Printf("migration %s complete\n", direction)
}

func migrationsDir() (string, error) {
	for _, candidate := range []string{
		filepath.Join("backend", "migrations"),
		"migrations",
		filepath.Join("..", "migrations"),
	} {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("migrations directory not found")
}
