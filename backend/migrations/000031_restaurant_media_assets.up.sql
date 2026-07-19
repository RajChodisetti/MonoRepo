-- Durable media is restricted to restaurant-owned or separately licensed
-- assets. Google Places photo names, media URLs, and bytes remain live-only and
-- are never persisted in this table.

CREATE TABLE IF NOT EXISTS restaurant_media_assets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  source_kind text NOT NULL,
  storage_key text NOT NULL,
  media_type text NOT NULL DEFAULT 'other',
  caption text NOT NULL DEFAULT '',
  alt_text text NOT NULL DEFAULT '',
  tags jsonb NOT NULL DEFAULT '[]'::jsonb,
  quality_score numeric(4,3),
  hero_score numeric(4,3),
  orientation text NOT NULL DEFAULT 'unknown',
  subject_position text NOT NULL DEFAULT 'center',
  contains_people boolean NOT NULL DEFAULT false,
  contains_text boolean NOT NULL DEFAULT false,
  placement_role text NOT NULL DEFAULT 'gallery',
  approval_status text NOT NULL DEFAULT 'draft',
  rights_status text NOT NULL,
  mime_type text NOT NULL,
  width_px int NOT NULL,
  height_px int NOT NULL,
  byte_size bigint NOT NULL,
  sha256 text NOT NULL,
  sort_order int NOT NULL DEFAULT 0,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  vision_status text NOT NULL DEFAULT 'pending',
  vision_attempts int NOT NULL DEFAULT 0,
  vision_claim_id uuid,
  vision_claimed_at timestamptz,
  vision_last_error text NOT NULL DEFAULT '',
  vision_result jsonb NOT NULL DEFAULT '{}'::jsonb,
  vision_analyzed_at timestamptz,
  hidden_at timestamptz,
  hidden_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT restaurant_media_assets_source_kind_check
    CHECK (source_kind IN ('owner_upload', 'licensed')),
  CONSTRAINT restaurant_media_assets_media_type_check
    CHECK (media_type IN ('exterior', 'interior', 'food', 'drink', 'logo', 'team', 'event', 'other')),
  CONSTRAINT restaurant_media_assets_orientation_check
    CHECK (orientation IN ('landscape', 'portrait', 'square', 'unknown')),
  CONSTRAINT restaurant_media_assets_subject_position_check
    CHECK (subject_position IN ('left', 'center', 'right')),
  CONSTRAINT restaurant_media_assets_placement_role_check
    CHECK (placement_role IN ('hero', 'about', 'gallery', 'food_gallery', 'ambience_gallery', 'logo')),
  CONSTRAINT restaurant_media_assets_approval_status_check
    CHECK (approval_status IN ('draft', 'approved', 'rejected')),
  CONSTRAINT restaurant_media_assets_rights_status_check
    CHECK (rights_status IN ('owner_granted', 'licensed')),
  CONSTRAINT restaurant_media_assets_vision_status_check
    CHECK (vision_status IN ('pending', 'running', 'verified', 'failed')),
  CONSTRAINT restaurant_media_assets_vision_attempts_check
    CHECK (vision_attempts >= 0),
  CONSTRAINT restaurant_media_assets_dimensions_check
    CHECK (width_px > 0 AND height_px > 0 AND byte_size > 0),
  CONSTRAINT restaurant_media_assets_sha256_check
    CHECK (sha256 ~ '^[0-9a-f]{64}$'),
  CONSTRAINT restaurant_media_assets_quality_score_check
    CHECK (quality_score IS NULL OR (quality_score >= 0 AND quality_score <= 1)),
  CONSTRAINT restaurant_media_assets_hero_score_check
    CHECK (hero_score IS NULL OR (hero_score >= 0 AND hero_score <= 1)),
  CONSTRAINT restaurant_media_assets_no_menu_photo_check
    CHECK (media_type <> 'menu_document')
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_restaurant_media_assets_storage_key
  ON restaurant_media_assets (storage_key);

CREATE INDEX IF NOT EXISTS idx_restaurant_media_assets_public
  ON restaurant_media_assets (
    restaurant_id,
    placement_role,
    sort_order,
    created_at
  )
  WHERE approval_status = 'approved' AND hidden_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_restaurant_media_assets_sha256
  ON restaurant_media_assets (restaurant_id, sha256);

CREATE INDEX IF NOT EXISTS idx_restaurant_media_assets_vision_queue
  ON restaurant_media_assets (vision_status, vision_claimed_at, created_at)
  WHERE approval_status = 'draft' AND hidden_at IS NULL;

-- Older Google OCR summaries were keyed only by array position. Return them to
-- pending so the worker can add a one-way provider-resource fingerprint before
-- any fresh photo becomes public. The expirable resource name is never stored.
UPDATE restaurant_profiles rp
SET ocr_status = 'pending',
    ocr_verified = false,
    ocr_verified_at = NULL,
    ocr_completed_at = NULL,
    updated_at = now()
WHERE rp.ocr_status = 'verified'
  AND EXISTS (
    SELECT 1
    FROM jsonb_array_elements(
      CASE
        WHEN jsonb_typeof(rp.raw_public_data->'menu_ocr'->'classifications') = 'array'
        THEN rp.raw_public_data->'menu_ocr'->'classifications'
        ELSE '[]'::jsonb
      END
    ) AS classification(value)
    WHERE classification.value->>'source' = 'google_places_photo'
      AND COALESCE(classification.value->>'source_fingerprint', '') = ''
  );

-- Demo fingerprints include the exact durable media selection that an admin
-- reviewed. Live Google photos remain runtime-only and are deliberately
-- represented only by policy, never by a cached name, URL, or byte copy.
CREATE OR REPLACE FUNCTION lead_artifact_current_public_payload(p_restaurant_id uuid)
RETURNS jsonb
LANGUAGE sql
STABLE
AS $$
  SELECT jsonb_build_object(
    'restaurant_name', r.name,
    'cuisine', COALESCE(rp.cuisines->>0, ''),
    'hero', COALESCE(
      (
        SELECT NULLIF(gi.url, '')
        FROM gallery_images gi
        WHERE gi.restaurant_id = r.id
          AND gi.url <> ''
          AND gi.hidden_at IS NULL
          AND lower(COALESCE(gi.image_type, '')) NOT IN ('menu_document', 'menu_list', 'menu_ocr')
          AND gi.url !~* '^https://[^/]*(googleusercontent\.com|ggpht\.com)(/|$)'
        ORDER BY gi.sort_order ASC, gi.created_at ASC, gi.id ASC
        LIMIT 1
      ),
      ''
    ),
    'hours', COALESCE(rp.opening_hours, '{}'::jsonb),
    'address', COALESCE(rp.address, ''),
    'phone', COALESCE(rp.phone, ''),
    'menu_sections', COALESCE(
      (
        SELECT jsonb_agg(grouped.section ORDER BY grouped.section_name)
        FROM (
          SELECT
            COALESCE(NULLIF(mi.category, ''), 'Menu') AS section_name,
            jsonb_build_object(
              'name', COALESCE(NULLIF(mi.category, ''), 'Menu'),
              'items', jsonb_agg(
                jsonb_build_object(
                  'name', mi.name,
                  'description', mi.description,
                  'price', COALESCE(
                    NULLIF(mi.price_text, ''),
                    CASE
                      WHEN mi.price IS NULL THEN ''
                      ELSE '$' || trim(to_char(mi.price, 'FM999999990.00'))
                    END
                  ),
                  'image_url', CASE
                    WHEN mi.image_url ~* '^https://[^/]*(googleusercontent\.com|ggpht\.com)(/|$)'
                    THEN ''
                    WHEN EXISTS (
                      SELECT 1
                      FROM menu_images menu_photo
                      WHERE menu_photo.restaurant_id = r.id
                        AND (
                          menu_photo.url = mi.image_url
                          OR menu_photo.thumbnail_url = mi.image_url
                        )
                    ) THEN ''
                    ELSE mi.image_url
                  END
                )
                ORDER BY mi.sort_order ASC, mi.name ASC, mi.id ASC
              )
            ) AS section
          FROM menu_items mi
          JOIN menus m ON m.id = mi.menu_id
          WHERE m.restaurant_id = r.id AND m.name = 'Imported Menu'
          GROUP BY COALESCE(NULLIF(mi.category, ''), 'Menu')
        ) grouped
      ),
      '[]'::jsonb
    ),
    'media_manifest', jsonb_build_object(
      'version', 1,
      'google_places_policy', 'live_resolve_no_cache',
      'menu_documents_public', false,
      'owned_assets', COALESCE(
        (
          SELECT jsonb_agg(
            jsonb_build_object(
              'id', asset.id::text,
              'source_kind', asset.source_kind,
              'media_type', asset.media_type,
              'placement_role', asset.placement_role,
              'caption', asset.caption,
              'alt_text', asset.alt_text,
              'width_px', asset.width_px,
              'height_px', asset.height_px,
              'sort_order', asset.sort_order
            )
            ORDER BY
              CASE asset.placement_role WHEN 'hero' THEN 0 WHEN 'about' THEN 1 ELSE 2 END,
              asset.hero_score DESC NULLS LAST,
              asset.sort_order,
              asset.created_at,
              asset.id
          )
          FROM restaurant_media_assets asset
          WHERE asset.restaurant_id = r.id
            AND asset.approval_status = 'approved'
            AND asset.hidden_at IS NULL
            AND asset.media_type <> 'menu_document'
        ),
        '[]'::jsonb
      )
    ),
    'reservation_cta', 'Book a table',
    'ai_receptionist_cta', 'Call our AI assistant'
  )
  FROM restaurants r
  JOIN restaurant_profiles rp ON rp.restaurant_id = r.id
  WHERE r.id = p_restaurant_id
$$;
