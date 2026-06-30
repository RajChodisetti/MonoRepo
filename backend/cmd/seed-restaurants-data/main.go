package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
)

const (
	databaseReadyTimeout = 30 * time.Second
	defaultDataFile      = "data/restaurants_data.json"
	importedMenuName     = "Imported Menu"
)

type dataFile struct {
	Meta        json.RawMessage     `json:"meta"`
	Restaurants []scrapedRestaurant `json:"restaurants"`
}

type scrapedRestaurant struct {
	Name         string          `json:"name"`
	Cuisines     json.RawMessage `json:"cuisines"`
	Rating       *float64        `json:"rating"`
	ReviewsCount *int            `json:"reviews_count"`
	Contact      contact         `json:"contact"`
	Owners       json.RawMessage `json:"owners"`
	Location     location        `json:"location"`
	MenuItems    []menuItem      `json:"menu_items"`
	Reviews      []review        `json:"reviews"`
	Images       json.RawMessage `json:"images"`
	Hours        json.RawMessage `json:"hours"`
	Google       googleMeta      `json:"google"`
	ApolloLead   json.RawMessage `json:"apollo_lead"`
	Errors       json.RawMessage `json:"errors"`
	ScrapeStatus string          `json:"scrape_status"`
	PriceLevel   string          `json:"price_level"`
}

type contact struct {
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website"`
}

type location struct {
	Address     string      `json:"address"`
	City        string      `json:"city"`
	State       string      `json:"state"`
	Country     string      `json:"country"`
	Coordinates coordinates `json:"coordinates"`
}

type coordinates struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type googleMeta struct {
	PlaceID string `json:"place_id"`
	DataID  string `json:"data_id"`
}

type menuItem struct {
	Name         string          `json:"name"`
	Category     string          `json:"category"`
	Description  string          `json:"description"`
	Price        string          `json:"price"`
	PriceNumeric *float64        `json:"price_numeric"`
	Images       json.RawMessage `json:"images"`
}

type review struct {
	Reviewer string          `json:"reviewer"`
	Review   string          `json:"review"`
	Stars    *float64        `json:"stars"`
	Date     string          `json:"date"`
	Images   json.RawMessage `json:"images"`
	Source   string          `json:"source"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	dataPath, err := resolveDataFile()
	if err != nil {
		log.Fatalf("resolve data file: %v", err)
	}

	raw, err := os.ReadFile(dataPath)
	if err != nil {
		log.Fatalf("read data file: %v", err)
	}

	var payload dataFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Fatalf("parse data file: %v", err)
	}
	if len(payload.Restaurants) == 0 {
		log.Fatal("no restaurants found in data file")
	}

	ctx := context.Background()
	appLog := logger.New(cfg.Logging)
	database, err := db.ConnectRequiredLogged(ctx, appLog, cfg.Database, databaseReadyTimeout)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.CloseLogged(ctx, appLog, database)

	pool := database.Pool()
	imported, skipped, err := importRestaurants(ctx, pool, dataPath, payload)
	if err != nil {
		log.Fatalf("import restaurants: %v", err)
	}

	fmt.Printf("import complete from %s\n", dataPath)
	fmt.Printf("restaurants imported/updated: %d\n", imported)
	fmt.Printf("duplicate place_ids skipped: %d\n", skipped)
}

func resolveDataFile() (string, error) {
	if path := strings.TrimSpace(os.Getenv("RESTAURANTS_DATA_FILE")); path != "" {
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("RESTAURANTS_DATA_FILE %q: %w", path, err)
		}
		return path, nil
	}

	for _, candidate := range []string{
		defaultDataFile,
		filepath.Join("..", "..", "data", "restaurants_data.json"),
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("restaurants data file not found; set RESTAURANTS_DATA_FILE")
}

func importRestaurants(ctx context.Context, pool *pgxpool.Pool, sourceFile string, payload dataFile) (int, int, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	seenPlaceIDs := make(map[string]struct{})
	imported := 0
	skipped := 0

	for _, record := range payload.Restaurants {
		placeID := strings.TrimSpace(record.Google.PlaceID)
		if placeID != "" {
			if _, exists := seenPlaceIDs[placeID]; exists {
				skipped++
				continue
			}
			seenPlaceIDs[placeID] = struct{}{}
		}

		restaurantID, err := upsertRestaurant(ctx, tx, record)
		if err != nil {
			return 0, 0, err
		}

		rawRecord, err := json.Marshal(record)
		if err != nil {
			return 0, 0, fmt.Errorf("marshal restaurant record: %w", err)
		}

		if err := upsertProfile(ctx, tx, restaurantID, record, rawRecord); err != nil {
			return 0, 0, err
		}

		menuID, err := upsertMenu(ctx, tx, restaurantID)
		if err != nil {
			return 0, 0, err
		}

		if err := replaceMenuItems(ctx, tx, menuID, record.MenuItems); err != nil {
			return 0, 0, err
		}

		if err := replaceReviews(ctx, tx, restaurantID, record.Reviews); err != nil {
			return 0, 0, err
		}

		imported++
	}

	var importID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO restaurant_data_imports (source_file, meta, restaurants_imported, restaurants_skipped)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		sourceFile,
		defaultJSON(payload.Meta),
		imported,
		skipped,
	).Scan(&importID)
	if err != nil {
		return 0, 0, fmt.Errorf("record import: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("commit transaction: %w", err)
	}

	return imported, skipped, nil
}

func upsertRestaurant(ctx context.Context, tx pgx.Tx, record scrapedRestaurant) (uuid.UUID, error) {
	email := strings.TrimSpace(record.Contact.Email)
	name := strings.TrimSpace(record.Name)
	placeID := strings.TrimSpace(record.Google.PlaceID)

	if placeID != "" {
		var restaurantID uuid.UUID
		err := tx.QueryRow(ctx, `
			SELECT restaurant_id
			FROM restaurant_profiles
			WHERE google_place_id = $1`,
			placeID,
		).Scan(&restaurantID)
		if err == nil {
			_, err = tx.Exec(ctx, `
				UPDATE restaurants
				SET name = $2, email = $3, updated_at = now()
				WHERE id = $1`,
				restaurantID, name, email,
			)
			if err != nil {
				return uuid.Nil, fmt.Errorf("update restaurant %q: %w", name, err)
			}
			return restaurantID, nil
		}
		if err != nil && err != pgx.ErrNoRows {
			return uuid.Nil, fmt.Errorf("lookup restaurant by place_id: %w", err)
		}
	}

	var restaurantID uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO restaurants (name, email, status)
		VALUES ($1, $2, 'lead')
		RETURNING id`,
		name, email,
	).Scan(&restaurantID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create restaurant %q: %w", name, err)
	}

	return restaurantID, nil
}

func upsertProfile(ctx context.Context, tx pgx.Tx, restaurantID uuid.UUID, record scrapedRestaurant, rawRecord []byte) error {
	cuisines := defaultJSON(record.Cuisines)
	owners := defaultJSON(record.Owners)
	images := defaultJSON(record.Images)
	hours := defaultJSON(record.Hours)
	apolloLead := defaultJSON(record.ApolloLead)
	scrapeErrors := defaultJSON(record.Errors)

	_, err := tx.Exec(ctx, `
		INSERT INTO restaurant_profiles (
			restaurant_id, opening_hours, phone, website, address, city, state, country,
			latitude, longitude, google_place_id, google_data_id, rating, reviews_count,
			price_level, cuisines, owners, images, apollo_lead, scrape_status, scrape_errors,
			dietary_options, raw_public_data
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21,
			$16, $22
		)
		ON CONFLICT (restaurant_id) DO UPDATE SET
			opening_hours = EXCLUDED.opening_hours,
			phone = EXCLUDED.phone,
			website = EXCLUDED.website,
			address = EXCLUDED.address,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			country = EXCLUDED.country,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			google_place_id = EXCLUDED.google_place_id,
			google_data_id = EXCLUDED.google_data_id,
			rating = EXCLUDED.rating,
			reviews_count = EXCLUDED.reviews_count,
			price_level = EXCLUDED.price_level,
			cuisines = EXCLUDED.cuisines,
			owners = EXCLUDED.owners,
			images = EXCLUDED.images,
			apollo_lead = EXCLUDED.apollo_lead,
			scrape_status = EXCLUDED.scrape_status,
			scrape_errors = EXCLUDED.scrape_errors,
			dietary_options = EXCLUDED.dietary_options,
			raw_public_data = EXCLUDED.raw_public_data,
			updated_at = now()`,
		restaurantID,
		hours,
		strings.TrimSpace(record.Contact.Phone),
		strings.TrimSpace(record.Contact.Website),
		strings.TrimSpace(record.Location.Address),
		strings.TrimSpace(record.Location.City),
		strings.TrimSpace(record.Location.State),
		strings.TrimSpace(record.Location.Country),
		record.Location.Coordinates.Latitude,
		record.Location.Coordinates.Longitude,
		nullIfEmpty(strings.TrimSpace(record.Google.PlaceID)),
		nullIfEmpty(strings.TrimSpace(record.Google.DataID)),
		record.Rating,
		record.ReviewsCount,
		strings.TrimSpace(record.PriceLevel),
		cuisines,
		owners,
		images,
		apolloLead,
		defaultString(record.ScrapeStatus, "unknown"),
		scrapeErrors,
		rawRecord,
	)
	if err != nil {
		return fmt.Errorf("upsert profile for %q: %w", record.Name, err)
	}

	return nil
}

func upsertMenu(ctx context.Context, tx pgx.Tx, restaurantID uuid.UUID) (uuid.UUID, error) {
	var menuID uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT id FROM menus
		WHERE restaurant_id = $1 AND name = $2
		LIMIT 1`,
		restaurantID, importedMenuName,
	).Scan(&menuID)
	if err == nil {
		return menuID, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, fmt.Errorf("lookup menu: %w", err)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO menus (restaurant_id, name, status)
		VALUES ($1, $2, 'active')
		RETURNING id`,
		restaurantID, importedMenuName,
	).Scan(&menuID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create menu: %w", err)
	}

	return menuID, nil
}

func replaceMenuItems(ctx context.Context, tx pgx.Tx, menuID uuid.UUID, items []menuItem) error {
	if _, err := tx.Exec(ctx, `DELETE FROM menu_items WHERE menu_id = $1`, menuID); err != nil {
		return fmt.Errorf("delete menu items: %w", err)
	}

	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" && strings.TrimSpace(item.Category) == "" {
			continue
		}
		if name == "" {
			name = strings.TrimSpace(item.Category)
		}

		imageURL := firstMenuImageURL(item.Images)

		_, err := tx.Exec(ctx, `
			INSERT INTO menu_items (
				menu_id, name, description, price, price_text, category,
				image_url, images, sort_order
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			menuID,
			name,
			strings.TrimSpace(item.Description),
			item.PriceNumeric,
			strings.TrimSpace(item.Price),
			strings.TrimSpace(item.Category),
			imageURL,
			defaultJSON(item.Images),
			i,
		)
		if err != nil {
			return fmt.Errorf("insert menu item %q: %w", name, err)
		}
	}

	return nil
}

func replaceReviews(ctx context.Context, tx pgx.Tx, restaurantID uuid.UUID, reviews []review) error {
	if _, err := tx.Exec(ctx, `DELETE FROM restaurant_reviews WHERE restaurant_id = $1`, restaurantID); err != nil {
		return fmt.Errorf("delete reviews: %w", err)
	}

	for i, item := range reviews {
		_, err := tx.Exec(ctx, `
			INSERT INTO restaurant_reviews (
				restaurant_id, reviewer, review_text, stars, review_date, images, source, sort_order
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			restaurantID,
			strings.TrimSpace(item.Reviewer),
			strings.TrimSpace(item.Review),
			item.Stars,
			strings.TrimSpace(item.Date),
			defaultJSON(item.Images),
			strings.TrimSpace(item.Source),
			i,
		)
		if err != nil {
			return fmt.Errorf("insert review: %w", err)
		}
	}

	return nil
}

func firstMenuImageURL(images json.RawMessage) string {
	if len(images) == 0 {
		return ""
	}

	var objects []struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(images, &objects); err == nil && len(objects) > 0 {
		return strings.TrimSpace(objects[0].URL)
	}

	var urls []string
	if err := json.Unmarshal(images, &urls); err == nil && len(urls) > 0 {
		return strings.TrimSpace(urls[0])
	}

	return ""
}

func defaultJSON(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("null")
	}
	return raw
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
