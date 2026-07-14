-- Bind automatically generated demo/campaign drafts to the OCR input that
-- produced them so a later verified input can refresh only agent-owned drafts.

ALTER TABLE demo_sites
  ADD COLUMN IF NOT EXISTS auto_generated boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS source_ocr_fingerprint text NOT NULL DEFAULT '';

ALTER TABLE demo_sites
  DROP CONSTRAINT IF EXISTS demo_sites_auto_ocr_fingerprint_check;

ALTER TABLE demo_sites
  ADD CONSTRAINT demo_sites_auto_ocr_fingerprint_check
  CHECK (auto_generated = false OR source_ocr_fingerprint <> '');

ALTER TABLE email_campaigns
  ADD COLUMN IF NOT EXISTS auto_generated boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS source_ocr_fingerprint text NOT NULL DEFAULT '';

ALTER TABLE email_campaigns
  DROP CONSTRAINT IF EXISTS email_campaigns_auto_ocr_fingerprint_check;

ALTER TABLE email_campaigns
  ADD CONSTRAINT email_campaigns_auto_ocr_fingerprint_check
  CHECK (auto_generated = false OR source_ocr_fingerprint <> '');
