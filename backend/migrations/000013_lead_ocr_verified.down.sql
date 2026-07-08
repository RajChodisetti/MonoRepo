DROP INDEX IF EXISTS idx_restaurant_profiles_ocr_verified;

ALTER TABLE restaurant_profiles
  DROP COLUMN IF EXISTS ocr_verification_errors,
  DROP COLUMN IF EXISTS ocr_verified_at,
  DROP COLUMN IF EXISTS ocr_verified;
