-- Harden public SEO-report email verification and retain the submitted contact
-- details separately from the restaurant's public listing contact fields.

ALTER TABLE seo_interested
  ADD COLUMN IF NOT EXISTS contact_name text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS contact_phone text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS otp_requested_at timestamptz,
  ADD COLUMN IF NOT EXISTS otp_attempts integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS unlock_expires_at timestamptz;

-- Existing verified links receive one bounded compatibility window. Pending
-- links remain bounded by their original OTP expiry (or expire immediately when
-- no expiry was recorded).
UPDATE seo_interested
SET
  otp_requested_at = COALESCE(otp_requested_at, updated_at, created_at),
  unlock_expires_at = COALESCE(
    unlock_expires_at,
    CASE
      WHEN verified_at IS NOT NULL THEN now() + interval '7 days'
      ELSE COALESCE(otp_expires_at, now())
    END
  );

ALTER TABLE seo_interested
  ALTER COLUMN otp_requested_at SET NOT NULL,
  ALTER COLUMN otp_requested_at SET DEFAULT now(),
  ALTER COLUMN unlock_expires_at SET NOT NULL,
  ALTER COLUMN unlock_expires_at SET DEFAULT now(),
  ADD CONSTRAINT seo_interested_contact_name_length_check
    CHECK (contact_name = '' OR char_length(contact_name) BETWEEN 2 AND 100),
  ADD CONSTRAINT seo_interested_contact_phone_length_check
    CHECK (contact_phone = '' OR char_length(contact_phone) <= 40),
  ADD CONSTRAINT seo_interested_otp_attempts_check
    CHECK (otp_attempts BETWEEN 0 AND 5);
