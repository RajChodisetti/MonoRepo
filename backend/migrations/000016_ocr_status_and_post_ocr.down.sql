DROP INDEX IF EXISTS idx_restaurant_profiles_ocr_verified_status;
DROP INDEX IF EXISTS idx_restaurant_profiles_ocr_running;
DROP INDEX IF EXISTS idx_restaurant_profiles_ocr_pending;

ALTER TABLE restaurant_profiles
  DROP CONSTRAINT IF EXISTS restaurant_profiles_ocr_status_check,
  DROP CONSTRAINT IF EXISTS restaurant_profiles_ocr_attempts_check,
  DROP CONSTRAINT IF EXISTS restaurant_profiles_ocr_verified_projection_check,
  DROP CONSTRAINT IF EXISTS restaurant_profiles_ocr_claim_state_check;

ALTER TABLE restaurant_profiles
  DROP COLUMN IF EXISTS ocr_input_fingerprint,
  DROP COLUMN IF EXISTS ocr_attempts,
  DROP COLUMN IF EXISTS ocr_completed_at,
  DROP COLUMN IF EXISTS ocr_started_at,
  DROP COLUMN IF EXISTS ocr_claim_fingerprint,
  DROP COLUMN IF EXISTS ocr_claim_id,
  DROP COLUMN IF EXISTS ocr_status;

CREATE INDEX IF NOT EXISTS idx_restaurant_profiles_ocr_verified
  ON restaurant_profiles (ocr_verified)
  WHERE ocr_verified = false;
