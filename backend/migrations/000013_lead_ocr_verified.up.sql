ALTER TABLE restaurant_profiles
  ADD COLUMN IF NOT EXISTS ocr_verified boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS ocr_verified_at timestamptz,
  ADD COLUMN IF NOT EXISTS ocr_verification_errors jsonb NOT NULL DEFAULT '[]';

CREATE INDEX IF NOT EXISTS idx_restaurant_profiles_ocr_verified
  ON restaurant_profiles (ocr_verified)
  WHERE ocr_verified = false;
