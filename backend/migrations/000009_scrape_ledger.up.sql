-- Scrape ledger: identity hash dedup + cron audit log

ALTER TABLE restaurant_profiles
  ADD COLUMN IF NOT EXISTS identity_hash text,
  ADD COLUMN IF NOT EXISTS last_scraped_at timestamptz,
  ADD COLUMN IF NOT EXISTS content_hash text,
  ADD COLUMN IF NOT EXISTS discovery_rank int;

CREATE UNIQUE INDEX IF NOT EXISTS idx_restaurant_profiles_identity_hash
  ON restaurant_profiles (identity_hash)
  WHERE identity_hash IS NOT NULL AND identity_hash <> '';

CREATE INDEX IF NOT EXISTS idx_restaurant_profiles_last_scraped_at
  ON restaurant_profiles (last_scraped_at)
  WHERE last_scraped_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS scrape_runs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  started_at timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz,
  city text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'running',
  leads_seen int NOT NULL DEFAULT 0,
  skipped_existing int NOT NULL DEFAULT 0,
  scraped_new int NOT NULL DEFAULT 0,
  refreshed_stale int NOT NULL DEFAULT 0,
  failed int NOT NULL DEFAULT 0,
  meta jsonb NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_scrape_runs_started_at ON scrape_runs (started_at DESC);
CREATE INDEX IF NOT EXISTS idx_scrape_runs_city ON scrape_runs (city);
