ALTER TABLE email_campaigns
  DROP CONSTRAINT IF EXISTS email_campaigns_campaign_type_check;

ALTER TABLE demo_sites
  DROP CONSTRAINT IF EXISTS demo_sites_status_check;

ALTER TABLE demo_sites
  DROP COLUMN IF EXISTS published_by,
  DROP COLUMN IF EXISTS published_at;

ALTER TABLE restaurant_profiles
  DROP CONSTRAINT IF EXISTS restaurant_profiles_review_status_check;

ALTER TABLE restaurant_profiles
  DROP COLUMN IF EXISTS reviewed_by,
  DROP COLUMN IF EXISTS reviewed_at;
