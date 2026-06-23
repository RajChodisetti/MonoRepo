DROP INDEX IF EXISTS idx_restaurants_shown_interest;
DROP INDEX IF EXISTS idx_restaurants_email;

ALTER TABLE restaurants
  DROP COLUMN IF EXISTS shown_interest,
  DROP COLUMN IF EXISTS is_contacted,
  DROP COLUMN IF EXISTS email;
