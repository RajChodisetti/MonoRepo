DROP INDEX IF EXISTS idx_restaurants_email_sent_pending;

ALTER TABLE restaurants
  DROP COLUMN IF EXISTS email_sent;
