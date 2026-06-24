package main

import (
	"context"
	"encoding/json"
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
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/store"
)

const databaseReadyTimeout = 30 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	restaurantName := strings.TrimSpace(os.Getenv("FIXTURE_RESTAURANT_NAME"))
	if restaurantName == "" {
		restaurantName = "Demo Fixture Cafe"
	}

	ownerEmail := strings.TrimSpace(os.Getenv("FIXTURE_OWNER_EMAIL"))
	if ownerEmail == "" {
		ownerEmail = "owner@local.test"
	}

	leadEmail := strings.TrimSpace(os.Getenv("FIXTURE_RESTAURANT_EMAIL"))
	if leadEmail == "" {
		leadEmail = ownerEmail
	}

	ownerPassword := strings.TrimSpace(os.Getenv("FIXTURE_OWNER_PASSWORD"))
	if ownerPassword == "" {
		ownerPassword = "password123"
	}

	ownerFullName := strings.TrimSpace(os.Getenv("FIXTURE_OWNER_FULL_NAME"))
	if ownerFullName == "" {
		ownerFullName = "Demo Restaurant Owner"
	}

	demoSlug := strings.TrimSpace(os.Getenv("FIXTURE_DEMO_SLUG"))
	if demoSlug == "" {
		demoSlug = "demo-fixture-cafe"
	}

	demoToken := strings.TrimSpace(os.Getenv("FIXTURE_DEMO_TOKEN"))
	if demoToken == "" {
		demoToken = "demo-fixture-token-value-32chars"
	}

	if len(ownerPassword) < 8 {
		log.Fatal("FIXTURE_OWNER_PASSWORD must be at least 8 characters")
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

	restaurantRecord, err := dataStore.Restaurants.Create(ctx, restaurants.CreateInput{
		Name:  restaurantName,
		Email: leadEmail,
	})
	if err != nil {
		log.Fatalf("create restaurant: %v", err)
	}

	ownerRecord, err := ensureOwnerUser(ctx, dataStore, ownerEmail, ownerPassword, ownerFullName)
	if err != nil {
		log.Fatalf("ensure owner user: %v", err)
	}

	if _, err := dataStore.Memberships.AddMember(ctx, restaurantRecord.ID, ownerRecord.ID, "owner"); err != nil {
		log.Fatalf("add restaurant membership: %v", err)
	}

	tokenHash, err := demos.HashDemoToken(demoToken)
	if err != nil {
		log.Fatalf("hash demo token: %v", err)
	}

	publicPayload := json.RawMessage(`{
		"restaurant_name": "` + restaurantName + `",
		"cuisine": "Thai",
		"hero": "Welcome to ` + restaurantName + `",
		"hours": {"monday": "9am-9pm"},
		"address": "123 Demo Street",
		"phone": "+1-555-0199",
		"menu_sections": [{"name": "Mains", "items": ["Pad Thai", "Green Curry"]}],
		"reservation_cta": "Book a table",
		"ai_receptionist_cta": "Call our AI assistant",
		"lead_notes": "internal-only fixture note"
	}`)

	var expiresAt *time.Time
	if cfg.Demo.TokenTTL > 0 {
		expiry := time.Now().Add(cfg.Demo.TokenTTL)
		expiresAt = &expiry
	}

	demoRecord, err := dataStore.Demos.Create(ctx, demos.CreateInput{
		RestaurantID:  restaurantRecord.ID,
		Slug:          demoSlug,
		TokenHash:     tokenHash,
		Status:        demos.StatusPublished,
		PublicPayload: publicPayload,
		ExpiresAt:     expiresAt,
	})
	if err != nil {
		log.Fatalf("create demo site: %v", err)
	}

	appLog.InfoContext(ctx, "seed_demo_fixture_created",
		"restaurant_id", restaurantRecord.ID.String(),
		"owner_email", ownerRecord.Email,
		"demo_slug", demoRecord.Slug,
	)

	fmt.Printf("restaurant created: %s (%s)\n", restaurantRecord.Name, restaurantRecord.ID)
	fmt.Printf("owner user: %s (%s)\n", ownerRecord.Email, ownerRecord.ID)
	fmt.Printf("public demo URL: /api/public/v1/demo/%s?token=%s\n", demoRecord.Slug, demoToken)
}

func ensureOwnerUser(ctx context.Context, dataStore *store.Store, email, password, fullName string) (auth.User, error) {
	existing, err := dataStore.Users.GetByEmail(ctx, email)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return auth.User{}, err
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return auth.User{}, err
	}

	return dataStore.Users.Create(ctx, auth.CreateInput{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Role:         auth.RoleRestaurantOwner,
	})
}
