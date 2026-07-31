-- SEO report unlock: email OTP verification + interested tracking.

CREATE TABLE IF NOT EXISTS seo_interested (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text NOT NULL,
  place_id text NOT NULL,
  restaurant_name text NOT NULL DEFAULT '',
  otp_hash text NOT NULL DEFAULT '',
  otp_expires_at timestamptz,
  verified_at timestamptz,
  interested boolean NOT NULL DEFAULT false,
  unlock_token text NOT NULL,
  lead_restaurant_id uuid REFERENCES restaurants(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT seo_interested_email_place_unique UNIQUE (email, place_id),
  CONSTRAINT seo_interested_unlock_token_unique UNIQUE (unlock_token)
);

CREATE INDEX IF NOT EXISTS idx_seo_interested_place_id
  ON seo_interested (place_id);

CREATE INDEX IF NOT EXISTS idx_seo_interested_interested
  ON seo_interested (interested)
  WHERE interested = true;

CREATE INDEX IF NOT EXISTS idx_seo_interested_email
  ON seo_interested (email);
