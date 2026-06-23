-- Restaurant lead fields (sales MVP scope): email, contact flags, interest tracking.

ALTER TABLE restaurants
  ADD COLUMN IF NOT EXISTS email text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS is_contacted boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS shown_interest boolean NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_restaurants_email
  ON restaurants (email);

CREATE INDEX IF NOT EXISTS idx_restaurants_shown_interest
  ON restaurants (shown_interest);
