-- Audit the human review gates between OCR verification and outreach sending.

ALTER TABLE restaurant_profiles
  ADD COLUMN IF NOT EXISTS reviewed_at timestamptz,
  ADD COLUMN IF NOT EXISTS reviewed_by uuid REFERENCES users(id) ON DELETE SET NULL;

UPDATE restaurant_profiles
SET review_status = 'draft',
    reviewed_at = NULL,
    reviewed_by = NULL
WHERE review_status <> 'draft'
   OR reviewed_at IS NOT NULL
   OR reviewed_by IS NOT NULL;

ALTER TABLE restaurant_profiles
  DROP CONSTRAINT IF EXISTS restaurant_profiles_review_status_check;

ALTER TABLE restaurant_profiles
  ADD CONSTRAINT restaurant_profiles_review_status_check
  CHECK (review_status IN ('draft', 'approved', 'rejected'));

ALTER TABLE demo_sites
  ADD COLUMN IF NOT EXISTS published_at timestamptz,
  ADD COLUMN IF NOT EXISTS published_by uuid REFERENCES users(id) ON DELETE SET NULL;

UPDATE demo_sites
SET status = 'draft',
    published_at = NULL,
    published_by = NULL
WHERE status <> 'draft'
   OR published_at IS NOT NULL
   OR published_by IS NOT NULL;

ALTER TABLE demo_sites
  DROP CONSTRAINT IF EXISTS demo_sites_status_check;

ALTER TABLE demo_sites
  ADD CONSTRAINT demo_sites_status_check
  CHECK (status IN ('draft', 'published'));

-- Outreach is the only implemented campaign type. Enforce this in PostgreSQL
-- so a typo/custom value cannot bypass the quota-managed outreach path.
UPDATE email_campaigns
SET campaign_type = 'outreach',
    status = CASE
      WHEN status IN ('draft', 'approved') THEN 'draft'
      ELSE status
    END,
    approved_at = CASE
      WHEN status IN ('draft', 'approved') THEN NULL
      ELSE approved_at
    END,
    approved_by = CASE
      WHEN status IN ('draft', 'approved') THEN NULL
      ELSE approved_by
    END,
    updated_at = now()
WHERE campaign_type <> 'outreach';

ALTER TABLE email_campaigns
  DROP CONSTRAINT IF EXISTS email_campaigns_campaign_type_check;

ALTER TABLE email_campaigns
  ADD CONSTRAINT email_campaigns_campaign_type_check
  CHECK (campaign_type = 'outreach');
