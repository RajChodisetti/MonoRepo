package media

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (repo *Postgres) ListPublic(ctx context.Context, restaurantID uuid.UUID) ([]Asset, error) {
	return repo.list(ctx, restaurantID, false)
}

func (repo *Postgres) ListAdmin(ctx context.Context, restaurantID uuid.UUID) ([]Asset, error) {
	return repo.list(ctx, restaurantID, true)
}

func (repo *Postgres) list(ctx context.Context, restaurantID uuid.UUID, includeHidden bool) ([]Asset, error) {
	if repo.pool == nil {
		return []Asset{}, nil
	}
	query := `
		SELECT id, restaurant_id, source_kind, storage_key, media_type,
		       caption, alt_text, tags, quality_score, hero_score, orientation,
		       subject_position, contains_people, contains_text, placement_role,
		       approval_status, rights_status, mime_type, width_px, height_px,
		       byte_size, sha256, sort_order, vision_status, vision_attempts,
		       vision_last_error, vision_result, vision_analyzed_at,
		       hidden_at, hidden_by, created_by,
		       created_at, updated_at
		FROM restaurant_media_assets
		WHERE restaurant_id = $1`
	if includeHidden {
		query += ` ORDER BY hidden_at NULLS FIRST, placement_role, sort_order, created_at, id`
	} else {
		query += `
			AND approval_status = 'approved'
			AND hidden_at IS NULL
			AND media_type <> 'menu_document'
			ORDER BY
			  CASE placement_role WHEN 'hero' THEN 0 WHEN 'about' THEN 1 ELSE 2 END,
			  hero_score DESC NULLS LAST,
			  sort_order, created_at, id`
	}

	rows, err := repo.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("list restaurant media assets: %w", err)
	}
	defer rows.Close()

	assets := make([]Asset, 0)
	for rows.Next() {
		var asset Asset
		if err := rows.Scan(
			&asset.ID,
			&asset.RestaurantID,
			&asset.SourceKind,
			&asset.StorageKey,
			&asset.MediaType,
			&asset.Caption,
			&asset.AltText,
			&asset.Tags,
			&asset.QualityScore,
			&asset.HeroScore,
			&asset.Orientation,
			&asset.SubjectPosition,
			&asset.ContainsPeople,
			&asset.ContainsText,
			&asset.PlacementRole,
			&asset.ApprovalStatus,
			&asset.RightsStatus,
			&asset.MimeType,
			&asset.WidthPx,
			&asset.HeightPx,
			&asset.ByteSize,
			&asset.SHA256,
			&asset.SortOrder,
			&asset.VisionStatus,
			&asset.VisionAttempts,
			&asset.VisionLastError,
			&asset.VisionResult,
			&asset.VisionAnalyzedAt,
			&asset.HiddenAt,
			&asset.HiddenBy,
			&asset.CreatedBy,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan restaurant media asset: %w", err)
		}
		assets = append(assets, asset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list restaurant media asset rows: %w", err)
	}
	return assets, nil
}

func (repo *Postgres) ListClassificationHints(ctx context.Context, restaurantID uuid.UUID) ([]ClassificationHint, error) {
	if repo.pool == nil {
		return []ClassificationHint{}, nil
	}
	rows, err := repo.pool.Query(ctx, `
		SELECT
		  COALESCE(
		    CASE WHEN classification.value ? 'source_index'
		      THEN (classification.value->>'source_index')::int
		    END,
		    classification.ordinality::int - 1
		  ) AS source_index,
		  COALESCE(classification.value->>'source_fingerprint', ''),
		  COALESCE(classification.value->>'image_type', 'other'),
		  COALESCE((classification.value->>'confidence')::double precision, 0),
		  COALESCE((classification.value->>'public_eligible')::boolean, false)
		FROM restaurant_profiles rp
		CROSS JOIN LATERAL jsonb_array_elements(
		  CASE
		    WHEN jsonb_typeof(rp.raw_public_data->'menu_ocr'->'classifications') = 'array'
		    THEN rp.raw_public_data->'menu_ocr'->'classifications'
		    ELSE '[]'::jsonb
		  END
		) WITH ORDINALITY AS classification(value, ordinality)
		WHERE rp.restaurant_id = $1
		ORDER BY source_index`, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("list media classification hints: %w", err)
	}
	defer rows.Close()

	hints := make([]ClassificationHint, 0)
	for rows.Next() {
		var hint ClassificationHint
		if err := rows.Scan(&hint.SourceIndex, &hint.SourceFingerprint, &hint.MediaType, &hint.Confidence, &hint.PublicEligible); err != nil {
			return nil, fmt.Errorf("scan media classification hint: %w", err)
		}
		hints = append(hints, hint)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list media classification hint rows: %w", err)
	}
	return hints, nil
}

func (repo *Postgres) Create(ctx context.Context, input CreateAssetInput) (Asset, error) {
	if repo.pool == nil {
		return Asset{}, fmt.Errorf("database pool is not configured")
	}
	var asset Asset
	err := repo.pool.QueryRow(ctx, `
		INSERT INTO restaurant_media_assets (
		  restaurant_id, source_kind, storage_key, media_type, caption, alt_text,
		  orientation, subject_position, placement_role, approval_status,
		  rights_status, mime_type, width_px, height_px, byte_size, sha256,
		  created_by
		) VALUES (
		  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
		  $15, $16, $17
		)
		RETURNING id, restaurant_id, source_kind, storage_key, media_type,
		  caption, alt_text, tags, quality_score, hero_score, orientation,
		  subject_position, contains_people, contains_text, placement_role,
		  approval_status, rights_status, mime_type, width_px, height_px,
		  byte_size, sha256, sort_order, vision_status, vision_attempts,
		  vision_last_error, vision_result, vision_analyzed_at,
		  hidden_at, hidden_by, created_by,
		  created_at, updated_at`,
		input.RestaurantID,
		input.SourceKind,
		input.StorageKey,
		input.MediaType,
		input.Caption,
		input.AltText,
		input.Orientation,
		input.SubjectPosition,
		input.PlacementRole,
		input.ApprovalStatus,
		input.RightsStatus,
		input.MimeType,
		input.WidthPx,
		input.HeightPx,
		input.ByteSize,
		input.SHA256,
		input.CreatedBy,
	).Scan(
		&asset.ID,
		&asset.RestaurantID,
		&asset.SourceKind,
		&asset.StorageKey,
		&asset.MediaType,
		&asset.Caption,
		&asset.AltText,
		&asset.Tags,
		&asset.QualityScore,
		&asset.HeroScore,
		&asset.Orientation,
		&asset.SubjectPosition,
		&asset.ContainsPeople,
		&asset.ContainsText,
		&asset.PlacementRole,
		&asset.ApprovalStatus,
		&asset.RightsStatus,
		&asset.MimeType,
		&asset.WidthPx,
		&asset.HeightPx,
		&asset.ByteSize,
		&asset.SHA256,
		&asset.SortOrder,
		&asset.VisionStatus,
		&asset.VisionAttempts,
		&asset.VisionLastError,
		&asset.VisionResult,
		&asset.VisionAnalyzedAt,
		&asset.HiddenAt,
		&asset.HiddenBy,
		&asset.CreatedBy,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		return Asset{}, fmt.Errorf("create restaurant media asset: %w", err)
	}
	return asset, nil
}

func (repo *Postgres) SetHidden(
	ctx context.Context,
	restaurantID, assetID uuid.UUID,
	hiddenBy *uuid.UUID,
) error {
	if repo.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}
	var tag pgconn.CommandTag
	var err error
	if hiddenBy == nil {
		tag, err = repo.pool.Exec(ctx, `
			UPDATE restaurant_media_assets
			SET hidden_at = NULL, hidden_by = NULL, updated_at = now()
			WHERE restaurant_id = $1 AND id = $2`, restaurantID, assetID)
	} else {
		tag, err = repo.pool.Exec(ctx, `
			UPDATE restaurant_media_assets
			SET hidden_at = now(), hidden_by = $3, updated_at = now()
			WHERE restaurant_id = $1 AND id = $2`, restaurantID, assetID, *hiddenBy)
	}
	if err != nil {
		return fmt.Errorf("set restaurant media visibility: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

var _ Repository = (*Postgres)(nil)
