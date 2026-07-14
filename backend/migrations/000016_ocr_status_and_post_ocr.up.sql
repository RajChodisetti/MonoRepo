-- Replace the ambiguous OCR boolean with an explicit state machine while
-- retaining ocr_verified as a compatibility projection during rollout.

ALTER TABLE restaurant_profiles
  ADD COLUMN IF NOT EXISTS ocr_status text,
  ADD COLUMN IF NOT EXISTS ocr_started_at timestamptz,
  ADD COLUMN IF NOT EXISTS ocr_completed_at timestamptz,
  ADD COLUMN IF NOT EXISTS ocr_attempts integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS ocr_input_fingerprint text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS ocr_claim_id uuid,
  ADD COLUMN IF NOT EXISTS ocr_claim_fingerprint text;

UPDATE restaurant_profiles
SET ocr_status = CASE
  WHEN ocr_verified = true
    AND COALESCE(ocr_verification_errors, '[]'::jsonb) @> '["no_images"]'::jsonb
    THEN 'no_images'
  WHEN ocr_verified = true THEN 'verified'
  WHEN jsonb_typeof(COALESCE(ocr_verification_errors, '[]'::jsonb)) <> 'array'
    THEN 'failed'
  WHEN jsonb_array_length(COALESCE(ocr_verification_errors, '[]'::jsonb)) > 0
    THEN 'failed'
  ELSE 'pending'
END
WHERE ocr_status IS NULL;

ALTER TABLE restaurant_profiles
  ALTER COLUMN ocr_status SET DEFAULT 'pending',
  ALTER COLUMN ocr_status SET NOT NULL;

ALTER TABLE restaurant_profiles
  DROP CONSTRAINT IF EXISTS restaurant_profiles_ocr_status_check;

ALTER TABLE restaurant_profiles
  DROP CONSTRAINT IF EXISTS restaurant_profiles_ocr_attempts_check;

ALTER TABLE restaurant_profiles
  ADD CONSTRAINT restaurant_profiles_ocr_status_check
  CHECK (ocr_status IN ('pending', 'running', 'verified', 'no_images', 'failed'));

ALTER TABLE restaurant_profiles
  ADD CONSTRAINT restaurant_profiles_ocr_attempts_check
  CHECK (ocr_attempts >= 0);

-- A no_images result is terminal but is deliberately not verified/eligible.
UPDATE restaurant_profiles
SET
  ocr_verified = (ocr_status = 'verified'),
  ocr_verified_at = CASE
    WHEN ocr_status = 'verified' THEN COALESCE(ocr_verified_at, now())
    ELSE NULL
  END,
  ocr_completed_at = CASE
    WHEN ocr_status IN ('verified', 'no_images', 'failed')
      THEN COALESCE(ocr_completed_at, ocr_verified_at, now())
    ELSE NULL
  END,
  ocr_claim_id = NULL,
  ocr_claim_fingerprint = NULL;

ALTER TABLE restaurant_profiles
  DROP CONSTRAINT IF EXISTS restaurant_profiles_ocr_verified_projection_check;

ALTER TABLE restaurant_profiles
  DROP CONSTRAINT IF EXISTS restaurant_profiles_ocr_claim_state_check;

ALTER TABLE restaurant_profiles
  ADD CONSTRAINT restaurant_profiles_ocr_verified_projection_check
  CHECK (ocr_verified = (ocr_status = 'verified'));

ALTER TABLE restaurant_profiles
  ADD CONSTRAINT restaurant_profiles_ocr_claim_state_check
  CHECK (
    (ocr_status = 'running' AND ocr_claim_id IS NOT NULL AND ocr_claim_fingerprint IS NOT NULL)
    OR (ocr_status <> 'running' AND ocr_claim_id IS NULL AND ocr_claim_fingerprint IS NULL)
  );

DROP INDEX IF EXISTS idx_restaurant_profiles_ocr_verified;

CREATE INDEX IF NOT EXISTS idx_restaurant_profiles_ocr_pending
  ON restaurant_profiles (created_at, restaurant_id)
  WHERE ocr_status = 'pending';

CREATE INDEX IF NOT EXISTS idx_restaurant_profiles_ocr_running
  ON restaurant_profiles (ocr_started_at)
  WHERE ocr_status = 'running';

CREATE INDEX IF NOT EXISTS idx_restaurant_profiles_ocr_verified_status
  ON restaurant_profiles (restaurant_id)
  WHERE ocr_status = 'verified';
