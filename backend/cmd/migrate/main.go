package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/migrations"
)

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
	if cfg.Database.URL == "" {
		log.Fatal("DATABASE_URL is required to run migrations")
	}

	ctx := context.Background()
	database, err := db.Connect(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer database.Close()

	dir := filepath.Join("backend", "migrations")
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
