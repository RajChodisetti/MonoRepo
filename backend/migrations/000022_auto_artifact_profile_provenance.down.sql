ALTER TABLE email_campaigns
  DROP CONSTRAINT IF EXISTS email_campaigns_auto_profile_fingerprint_check,
  DROP COLUMN IF EXISTS source_profile_fingerprint;

ALTER TABLE demo_sites
  DROP CONSTRAINT IF EXISTS demo_sites_auto_profile_fingerprint_check,
  DROP COLUMN IF EXISTS source_profile_fingerprint;

DROP FUNCTION IF EXISTS lead_artifact_current_profile_fingerprint(uuid);
DROP FUNCTION IF EXISTS lead_artifact_current_public_payload(uuid);
