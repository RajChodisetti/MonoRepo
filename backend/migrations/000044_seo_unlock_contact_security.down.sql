ALTER TABLE seo_interested
  DROP CONSTRAINT IF EXISTS seo_interested_otp_attempts_check,
  DROP CONSTRAINT IF EXISTS seo_interested_contact_phone_length_check,
  DROP CONSTRAINT IF EXISTS seo_interested_contact_name_length_check,
  DROP COLUMN IF EXISTS unlock_expires_at,
  DROP COLUMN IF EXISTS otp_attempts,
  DROP COLUMN IF EXISTS otp_requested_at,
  DROP COLUMN IF EXISTS contact_phone,
  DROP COLUMN IF EXISTS contact_name;
