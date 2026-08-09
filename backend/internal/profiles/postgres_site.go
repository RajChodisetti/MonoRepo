package profiles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

const importedMenuName = "Imported Menu"

func (repo *Postgres) ListSiteRestaurants(ctx context.Context) ([]SiteRestaurantSummary, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT r.id, r.name, COALESCE(rp.google_place_id, ''), COALESCE(rp.city, '')
		FROM restaurants r
		JOIN restaurant_profiles rp ON rp.restaurant_id = r.id
		WHERE rp.google_place_id IS NOT NULL AND rp.google_place_id <> ''
		ORDER BY r.created_at ASC, r.name ASC`

	rows, err := repo.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list site restaurants: %w", err)
	}
	defer rows.Close()

	summaries := make([]SiteRestaurantSummary, 0)
	index := 0
	for rows.Next() {
		var summary SiteRestaurantSummary
		if err := rows.Scan(&summary.ID, &summary.Name, &summary.PlaceID, &summary.City); err != nil {
			return nil, fmt.Errorf("scan site restaurant: %w", err)
		}
		summary.Index = index
		index++
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list site restaurants rows: %w", err)
	}

	return summaries, nil
}

func (repo *Postgres) GetSiteRestaurantByID(ctx context.Context, restaurantID uuid.UUID) (SiteRestaurantSummary, error) {
	if repo.pool == nil {
		return SiteRestaurantSummary{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		WITH ranked AS (
			SELECT r.id, r.name, COALESCE(rp.google_place_id, '') AS place_id,
			       COALESCE(rp.city, '') AS city,
			       row_number() OVER (ORDER BY r.created_at ASC, r.name ASC) - 1 AS site_index
			FROM restaurants r
			JOIN restaurant_profiles rp ON rp.restaurant_id = r.id
			WHERE rp.google_place_id IS NOT NULL AND rp.google_place_id <> ''
		)
		SELECT id, name, place_id, city, site_index
		FROM ranked
		WHERE id = $1`

	var summary SiteRestaurantSummary
	err := repo.pool.QueryRow(ctx, query, restaurantID).Scan(
		&summary.ID,
		&summary.Name,
		&summary.PlaceID,
		&summary.City,
		&summary.Index,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SiteRestaurantSummary{}, repository.ErrNotFound
	}
	if err != nil {
		return SiteRestaurantSummary{}, fmt.Errorf("load generated site restaurant: %w", err)
	}
	return summary, nil
}

func (repo *Postgres) GetSiteContentByIndex(ctx context.Context, index int) (SiteContent, error) {
	summaries, err := repo.ListSiteRestaurants(ctx)
	if err != nil {
		return SiteContent{}, err
	}
	if index < 0 || index >= len(summaries) {
		return SiteContent{}, repository.ErrNotFound
	}
	return repo.buildSiteContent(ctx, summaries[index])
}

func (repo *Postgres) GetSiteContentByID(ctx context.Context, restaurantID uuid.UUID) (SiteContent, error) {
	summary, err := repo.GetSiteRestaurantByID(ctx, restaurantID)
	if err != nil {
		return SiteContent{}, err
	}
	return repo.buildSiteContent(ctx, summary)
}

func (repo *Postgres) GetSiteContentByPlaceID(ctx context.Context, placeID string) (SiteContent, error) {
	summaries, err := repo.ListSiteRestaurants(ctx)
	if err != nil {
		return SiteContent{}, err
	}
	for _, summary := range summaries {
		if summary.PlaceID == placeID {
			return repo.buildSiteContent(ctx, summary)
		}
	}
	return SiteContent{}, repository.ErrNotFound
}

func (repo *Postgres) buildSiteContent(ctx context.Context, summary SiteRestaurantSummary) (SiteContent, error) {
	if repo.pool == nil {
		return SiteContent{}, fmt.Errorf("database pool is not configured")
	}

	const profileQuery = `
		SELECT
			r.name,
			COALESCE(r.email, ''),
			COALESCE(rp.google_place_id, ''),
			COALESCE(rp.cuisines, '[]'::jsonb),
			rp.rating,
			rp.reviews_count,
			COALESCE(rp.price_level, ''),
			COALESCE(rp.phone, ''),
			COALESCE(rp.website, ''),
			COALESCE(rp.address, ''),
			COALESCE(rp.city, ''),
			COALESCE(rp.state, ''),
			COALESCE(rp.country, ''),
			rp.latitude,
			rp.longitude,
			COALESCE(rp.opening_hours, '{}'::jsonb),
			'',
			rp.updated_at
		FROM restaurants r
		JOIN restaurant_profiles rp ON rp.restaurant_id = r.id
		WHERE r.id = $1`

	content := SiteContent{
		Index:        summary.Index,
		RestaurantID: summary.ID,
		PlaceID:      summary.PlaceID,
		Hours:        json.RawMessage("{}"),
		Cuisines:     json.RawMessage("[]"),
	}

	err := repo.pool.QueryRow(ctx, profileQuery, summary.ID).Scan(
		&content.Name,
		&content.Email,
		&content.PlaceID,
		&content.Cuisines,
		&content.Rating,
		&content.ReviewsCount,
		&content.PriceLevel,
		&content.Phone,
		&content.Website,
		&content.Address,
		&content.City,
		&content.State,
		&content.Country,
		&content.Latitude,
		&content.Longitude,
		&content.Hours,
		&content.Thumbnail,
		&content.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SiteContent{}, repository.ErrNotFound
	}
	if err != nil {
		return SiteContent{}, fmt.Errorf("load site profile: %w", err)
	}
	menuItems, err := repo.listSiteMenuItems(ctx, summary.ID)
	if err != nil {
		return SiteContent{}, err
	}

	reviews, err := repo.listSiteReviews(ctx, summary.ID)
	if err != nil {
		return SiteContent{}, err
	}

	content.MenuItems = menuItems
	content.GalleryImages = []GalleryImage{}
	content.Reviews = reviews

	return content, nil
}

func (repo *Postgres) listSiteMenuItems(ctx context.Context, restaurantID uuid.UUID) ([]SiteMenuItem, error) {
	const query = `
		SELECT mi.name, mi.category, mi.description,
		       COALESCE(mi.price_text, ''), mi.price
		FROM menu_items mi
		JOIN menus m ON m.id = mi.menu_id
		WHERE m.restaurant_id = $1 AND m.name = $2
		ORDER BY mi.sort_order ASC, mi.name ASC`

	rows, err := repo.pool.Query(ctx, query, restaurantID, importedMenuName)
	if err != nil {
		return nil, fmt.Errorf("list site menu items: %w", err)
	}
	defer rows.Close()

	items := make([]SiteMenuItem, 0)
	for rows.Next() {
		var item SiteMenuItem
		var priceNumeric *float64
		if err := rows.Scan(
			&item.Name,
			&item.Category,
			&item.Description,
			&item.Price,
			&priceNumeric,
		); err != nil {
			return nil, fmt.Errorf("scan site menu item: %w", err)
		}
		if priceNumeric != nil {
			rounded := math.Round(*priceNumeric*100) / 100
			item.PriceNumeric = &rounded
			if item.Price == "" {
				item.Price = fmt.Sprintf("$%.2f", rounded)
			}
		}

		// Legacy scraped image fields are intentionally not read into the public
		// menu model. Text and pricing remain available.
		item.ImageURL = ""
		item.Images = nil
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list site menu items rows: %w", err)
	}

	return items, nil
}

func (repo *Postgres) listSiteReviews(ctx context.Context, restaurantID uuid.UUID) ([]SiteReview, error) {
	const query = `
		SELECT reviewer, review_text, stars, review_date
		FROM restaurant_reviews
		WHERE restaurant_id = $1
		ORDER BY sort_order ASC
		LIMIT 12`

	rows, err := repo.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("list site reviews: %w", err)
	}
	defer rows.Close()

	reviews := make([]SiteReview, 0)
	for rows.Next() {
		var review SiteReview
		if err := rows.Scan(&review.Reviewer, &review.Review, &review.Stars, &review.Date); err != nil {
			return nil, fmt.Errorf("scan site review: %w", err)
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list site reviews rows: %w", err)
	}

	return reviews, nil
}
