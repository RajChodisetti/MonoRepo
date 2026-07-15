DROP TABLE IF EXISTS email_suppressions;
DROP TABLE IF EXISTS email_tracking_tokens;
DROP TABLE IF EXISTS email_events;
DROP TABLE IF EXISTS email_campaigns;
ALTER TABLE job_runs DROP COLUMN IF EXISTS locked_at;
