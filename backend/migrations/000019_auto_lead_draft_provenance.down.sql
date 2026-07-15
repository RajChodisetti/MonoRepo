ALTER TABLE email_campaigns
  DROP CONSTRAINT IF EXISTS email_campaigns_auto_ocr_fingerprint_check,
  DROP COLUMN IF EXISTS source_ocr_fingerprint,
  DROP COLUMN IF EXISTS auto_generated;

ALTER TABLE demo_sites
  DROP CONSTRAINT IF EXISTS demo_sites_auto_ocr_fingerprint_check,
  DROP COLUMN IF EXISTS source_ocr_fingerprint,
  DROP COLUMN IF EXISTS auto_generated;
