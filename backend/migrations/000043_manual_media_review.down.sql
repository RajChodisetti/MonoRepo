DROP INDEX IF EXISTS restaurant_media_assets_manual_review_queue;

ALTER TABLE restaurant_media_assets
  DROP CONSTRAINT IF EXISTS restaurant_media_assets_manual_review_check;

UPDATE restaurant_media_assets
SET vision_status = metadata ->> 'pre_manual_review_vision_status',
    vision_claim_id = (metadata ->> 'pre_manual_review_vision_claim_id')::uuid,
    vision_claimed_at = (metadata ->> 'pre_manual_review_vision_claimed_at')::timestamptz,
    vision_last_error = COALESCE(metadata ->> 'pre_manual_review_vision_last_error', ''),
    metadata = metadata
      - 'pre_manual_review_vision_status'
      - 'pre_manual_review_vision_claim_id'
      - 'pre_manual_review_vision_claimed_at'
      - 'pre_manual_review_vision_last_error'
      - 'manual_review_vision_retired_migration',
    updated_at = now()
WHERE metadata ->> 'manual_review_vision_retired_migration' = '43';

UPDATE restaurant_media_assets
SET metadata = metadata - 'manual_review_grandfathered' - 'manual_review_grandfathered_migration',
    updated_at = now()
WHERE metadata ->> 'manual_review_grandfathered_migration' = '43'
  AND metadata ->> 'manual_review_grandfathered' = 'true';

ALTER TABLE restaurant_media_assets
  DROP COLUMN IF EXISTS review_note,
  DROP COLUMN IF EXISTS reviewed_by,
  DROP COLUMN IF EXISTS reviewed_at;

CREATE INDEX IF NOT EXISTS idx_restaurant_media_assets_vision_queue
  ON restaurant_media_assets (vision_status, vision_claimed_at, created_at)
  WHERE approval_status = 'draft' AND hidden_at IS NULL;
