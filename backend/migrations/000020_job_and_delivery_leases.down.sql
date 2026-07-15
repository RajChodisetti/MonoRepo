DROP INDEX IF EXISTS idx_email_delivery_attempts_account_slot;
DROP INDEX IF EXISTS idx_email_delivery_attempts_expired_lease;

ALTER TABLE email_delivery_attempts
  DROP CONSTRAINT IF EXISTS email_delivery_attempts_lease_state_check;

ALTER TABLE email_delivery_attempts
  DROP COLUMN IF EXISTS lease_expires_at;

DROP INDEX IF EXISTS idx_job_runs_one_active_bulk_outreach;
DROP INDEX IF EXISTS idx_job_runs_expired_lease;

ALTER TABLE job_runs
  DROP COLUMN IF EXISTS lease_expires_at,
  DROP COLUMN IF EXISTS locked_by;
