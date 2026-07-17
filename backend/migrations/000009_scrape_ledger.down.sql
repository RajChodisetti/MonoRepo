DROP TABLE IF EXISTS scrape_runs;

DROP INDEX IF EXISTS idx_restaurant_profiles_last_scraped_at;
DROP INDEX IF EXISTS idx_restaurant_profiles_identity_hash;

ALTER TABLE restaurant_profiles
  DROP COLUMN IF EXISTS discovery_rank,
  DROP COLUMN IF EXISTS content_hash,
  DROP COLUMN IF EXISTS last_scraped_at,
  DROP COLUMN IF EXISTS identity_hash;
