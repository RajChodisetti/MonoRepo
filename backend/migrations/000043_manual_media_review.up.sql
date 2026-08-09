-- Replace automated vision approval with explicit administrator review while
-- retaining historical vision columns for rollback/audit.

ALTER TABLE restaurant_media_assets
  ADD COLUMN reviewed_at timestamptz,
  ADD COLUMN reviewed_by uuid REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN review_note text NOT NULL DEFAULT '';

-- Pre-deploy impact query (read-only). Run this against each target before the
-- migration; it reports the rows that will be grandfathered or have an active
-- automated claim retired without changing their existing approval decision.
-- SELECT approval_status, source_kind, rights_status, vision_status, count(*)
-- FROM restaurant_media_assets
-- WHERE approval_status IN ('approved', 'rejected') OR vision_status = 'running'
-- GROUP BY approval_status, source_kind, rights_status, vision_status
-- ORDER BY approval_status, source_kind, rights_status, vision_status;

-- Existing approval decisions remain effective. Their provenance cannot be
-- safely reclassified as manual, so mark them explicitly as grandfathered.
UPDATE restaurant_media_assets
SET metadata = metadata || jsonb_build_object(
      'manual_review_grandfathered', true,
      'manual_review_grandfathered_migration', 43
    ),
    updated_at = now()
WHERE approval_status IN ('approved', 'rejected');

-- Retire only claims that were actively running. Historical analysis fields
-- are retained for audit/rollback and no existing approval is demoted.
UPDATE restaurant_media_assets
SET metadata = metadata || jsonb_build_object(
      'pre_manual_review_vision_status', vision_status,
      'pre_manual_review_vision_claim_id', vision_claim_id,
      'pre_manual_review_vision_claimed_at', vision_claimed_at,
      'pre_manual_review_vision_last_error', vision_last_error,
      'manual_review_vision_retired_migration', 43
    ),
    vision_status = 'failed',
    vision_claim_id = NULL,
    vision_claimed_at = NULL,
    vision_last_error = 'Automated media review retired during an active claim.',
    updated_at = now()
WHERE vision_status = 'running';

ALTER TABLE restaurant_media_assets
  ADD CONSTRAINT restaurant_media_assets_manual_review_check CHECK (
    (approval_status = 'draft' AND reviewed_at IS NULL AND reviewed_by IS NULL)
    OR (
      approval_status IN ('approved', 'rejected')
      AND (
        (reviewed_at IS NOT NULL AND reviewed_by IS NOT NULL)
        OR (
          reviewed_at IS NULL
          AND reviewed_by IS NULL
          AND metadata ->> 'manual_review_grandfathered' = 'true'
        )
      )
    )
  );

DROP INDEX IF EXISTS idx_restaurant_media_assets_vision_queue;

CREATE INDEX restaurant_media_assets_manual_review_queue
  ON restaurant_media_assets (created_at, id)
  WHERE approval_status = 'draft' AND hidden_at IS NULL;
