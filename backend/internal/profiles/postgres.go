package profiles

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (repo *Postgres) GetRestaurantIDByPlaceID(ctx context.Context, placeID string) (uuid.UUID, error) {
	if repo.pool == nil {
		return uuid.Nil, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT restaurant_id
		FROM restaurant_profiles
		WHERE google_place_id = $1
		LIMIT 1`

	var restaurantID uuid.UUID
	err := repo.pool.QueryRow(ctx, query, placeID).Scan(&restaurantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, repository.ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("lookup restaurant by place_id: %w", err)
	}

	return restaurantID, nil
}

func (repo *Postgres) GetGooglePlaceID(ctx context.Context, restaurantID uuid.UUID) (string, error) {
	if repo.pool == nil {
		return "", fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT COALESCE(google_place_id, '')
		FROM restaurant_profiles
		WHERE restaurant_id = $1`

	var placeID string
	err := repo.pool.QueryRow(ctx, query, restaurantID).Scan(&placeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", repository.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("load restaurant google place id: %w", err)
	}
	if placeID == "" {
		return "", repository.ErrNotFound
	}
	return placeID, nil
}

func (repo *Postgres) ListMenuImages(ctx context.Context, restaurantID uuid.UUID) ([]MenuImage, error) {
	// Legacy scraped/menu-document images are retained for protected admin
	// review only. Public callers receive media through the live Google or
	// manually approved owner/licensed media service.
	return []MenuImage{}, nil
}

func (repo *Postgres) ListMenuImagesAdmin(ctx context.Context, restaurantID uuid.UUID) ([]MenuImage, error) {
	const query = `
		SELECT id, restaurant_id, url, thumbnail_url, image_type, confidence,
		       title, source, sort_order, metadata, created_at, updated_at, hidden_at, hidden_by
		FROM menu_images
		WHERE restaurant_id = $1
		ORDER BY sort_order ASC, created_at ASC`
	return repo.queryMenuImages(ctx, query, restaurantID)
}

func (repo *Postgres) queryMenuImages(ctx context.Context, query string, restaurantID uuid.UUID) ([]MenuImage, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	rows, err := repo.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("list menu images: %w", err)
	}
	defer rows.Close()

	images := make([]MenuImage, 0)
	for rows.Next() {
		record, scanErr := scanMenuImage(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		images = append(images, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list menu images rows: %w", err)
	}

	return images, nil
}

func (repo *Postgres) ListGalleryImages(ctx context.Context, restaurantID uuid.UUID) ([]GalleryImage, error) {
	return []GalleryImage{}, nil
}

func (repo *Postgres) ListGalleryImagesAdmin(ctx context.Context, restaurantID uuid.UUID) ([]GalleryImage, error) {
	const query = `
		SELECT id, restaurant_id, url, thumbnail_url, image_type, confidence,
		       title, source, sort_order, metadata, created_at, updated_at, hidden_at, hidden_by
		FROM gallery_images
		WHERE restaurant_id = $1
		ORDER BY sort_order ASC, created_at ASC`
	return repo.queryGalleryImages(ctx, query, restaurantID)
}

func (repo *Postgres) queryGalleryImages(ctx context.Context, query string, restaurantID uuid.UUID) ([]GalleryImage, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	rows, err := repo.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("list gallery images: %w", err)
	}
	defer rows.Close()

	images := make([]GalleryImage, 0)
	for rows.Next() {
		record, scanErr := scanGalleryImage(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		images = append(images, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list gallery images rows: %w", err)
	}

	return images, nil
}

func (repo *Postgres) HideMenuImage(ctx context.Context, restaurantID, imageID, hiddenBy uuid.UUID) error {
	return repo.setImageHidden(ctx, "menu_images", restaurantID, imageID, &hiddenBy)
}

func (repo *Postgres) HideGalleryImage(ctx context.Context, restaurantID, imageID, hiddenBy uuid.UUID) error {
	return repo.setImageHidden(ctx, "gallery_images", restaurantID, imageID, &hiddenBy)
}

func (repo *Postgres) UnhideMenuImage(ctx context.Context, restaurantID, imageID uuid.UUID) error {
	return repo.setImageHidden(ctx, "menu_images", restaurantID, imageID, nil)
}

func (repo *Postgres) UnhideGalleryImage(ctx context.Context, restaurantID, imageID uuid.UUID) error {
	return repo.setImageHidden(ctx, "gallery_images", restaurantID, imageID, nil)
}

func (repo *Postgres) setImageHidden(ctx context.Context, table string, restaurantID, imageID uuid.UUID, hiddenBy *uuid.UUID) error {
	if repo.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}

	var query string
	var args []any
	if hiddenBy != nil {
		query = fmt.Sprintf(`UPDATE %s SET hidden_at = now(), hidden_by = $3, updated_at = now() WHERE id = $1 AND restaurant_id = $2`, table)
		args = []any{imageID, restaurantID, *hiddenBy}
	} else {
		query = fmt.Sprintf(`UPDATE %s SET hidden_at = NULL, hidden_by = NULL, updated_at = now() WHERE id = $1 AND restaurant_id = $2`, table)
		args = []any{imageID, restaurantID}
	}

	result, err := repo.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update %s visibility: %w", table, err)
	}
	if result.RowsAffected() != 1 {
		return repository.ErrNotFound
	}
	return nil
}

func (repo *Postgres) GetSiteImages(ctx context.Context, restaurantID uuid.UUID) (SiteImages, error) {
	menuImages, err := repo.ListMenuImages(ctx, restaurantID)
	if err != nil {
		return SiteImages{}, err
	}
	galleryImages, err := repo.ListGalleryImages(ctx, restaurantID)
	if err != nil {
		return SiteImages{}, err
	}

	return SiteImages{
		RestaurantID:  restaurantID,
		MenuImages:    menuImages,
		GalleryImages: galleryImages,
	}, nil
}

func (repo *Postgres) GetSiteImagesByPlaceID(ctx context.Context, placeID string) (SiteImages, error) {
	restaurantID, err := repo.GetRestaurantIDByPlaceID(ctx, placeID)
	if err != nil {
		return SiteImages{}, err
	}
	return repo.GetSiteImages(ctx, restaurantID)
}

func scanMenuImage(scanner interface {
	Scan(dest ...any) error
}) (MenuImage, error) {
	var record MenuImage
	err := scanner.Scan(
		&record.ID,
		&record.RestaurantID,
		&record.URL,
		&record.ThumbnailURL,
		&record.ImageType,
		&record.Confidence,
		&record.Title,
		&record.Source,
		&record.SortOrder,
		&record.Metadata,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.HiddenAt,
		&record.HiddenBy,
	)
	if err != nil {
		return MenuImage{}, fmt.Errorf("scan menu image: %w", err)
	}
	return record, nil
}

func scanGalleryImage(scanner interface {
	Scan(dest ...any) error
}) (GalleryImage, error) {
	var record GalleryImage
	err := scanner.Scan(
		&record.ID,
		&record.RestaurantID,
		&record.URL,
		&record.ThumbnailURL,
		&record.ImageType,
		&record.Confidence,
		&record.Title,
		&record.Source,
		&record.SortOrder,
		&record.Metadata,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.HiddenAt,
		&record.HiddenBy,
	)
	if err != nil {
		return GalleryImage{}, fmt.Errorf("scan gallery image: %w", err)
	}
	return record, nil
}
