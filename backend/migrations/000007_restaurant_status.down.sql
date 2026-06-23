DROP INDEX IF EXISTS idx_restaurants_status;

ALTER TABLE restaurants
  DROP CONSTRAINT IF EXISTS restaurants_status_check;

ALTER TABLE restaurants
  DROP COLUMN IF EXISTS status;
