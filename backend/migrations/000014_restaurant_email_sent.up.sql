ALTER TABLE restaurants
  ADD COLUMN IF NOT EXISTS email_sent boolean NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_restaurants_email_sent_pending
  ON restaurants (email_sent)
  WHERE email_sent = false;
