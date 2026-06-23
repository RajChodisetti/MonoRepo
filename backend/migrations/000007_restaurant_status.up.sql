-- P1-010: restaurant lead lifecycle status.

ALTER TABLE restaurants
  ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'lead';

ALTER TABLE restaurants
  DROP CONSTRAINT IF EXISTS restaurants_status_check;

ALTER TABLE restaurants
  ADD CONSTRAINT restaurants_status_check CHECK (
    status IN (
      'lead', 'demo_ready', 'emailed', 'interested',
      'client_onboarding', 'active_client', 'lost', 'archived'
    )
  );

CREATE INDEX IF NOT EXISTS idx_restaurants_status
  ON restaurants (status);
