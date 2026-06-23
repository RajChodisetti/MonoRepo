package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/store"
)

const databaseReadyTimeout = 30 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	password := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
	fullName := strings.TrimSpace(os.Getenv("ADMIN_FULL_NAME"))
	if fullName == "" {
		fullName = "Platform Admin"
	}

	if email == "" || password == "" {
		log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD are required")
	}
	if len(password) < 8 {
		log.Fatal("ADMIN_PASSWORD must be at least 8 characters")
	}

	ctx := context.Background()
	appLog := logger.New(cfg.Logging)
	database, err := db.ConnectRequiredLogged(ctx, appLog, cfg.Database, databaseReadyTimeout)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.CloseLogged(ctx, appLog, database)

	dataStore := store.New(database)
	if err := dataStore.VerifyStartup(ctx); err != nil {
		log.Fatalf("verify startup: %v", err)
	}

	existing, err := dataStore.Users.GetByEmail(ctx, email)
	if err == nil {
		appLog.InfoContext(ctx, "seed_admin_skipped", "email", existing.Email, "reason", "already_exists")
		return
	}
	if !errors.Is(err, repository.ErrNotFound) {
		log.Fatalf("lookup admin user: %v", err)
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	record, err := dataStore.Users.Create(ctx, auth.CreateInput{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Role:         auth.RoleInternalAdmin,
	})
	if err != nil {
		log.Fatalf("create admin user: %v", err)
	}

	appLog.InfoContext(ctx, "seed_admin_created",
		"user_id", record.ID.String(),
		"email", record.Email,
		"role", record.Role,
	)
	fmt.Printf("admin user created: %s (%s)\n", record.Email, record.ID)
}
