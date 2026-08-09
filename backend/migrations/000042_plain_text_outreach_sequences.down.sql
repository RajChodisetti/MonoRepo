-- Rollback also fails closed. A previously active sender or cancelled job is
-- intentionally not restored; an operator must explicitly re-enable outreach
-- after verifying the rolled-back runtime.
INSERT INTO outreach_runtime_control (
  control_key, enabled, enabled_at, enabled_by, updated_at
)
VALUES ('email_job', false, NULL, NULL, now())
ON CONFLICT (control_key) DO UPDATE
SET enabled = false,
    enabled_at = NULL,
    enabled_by = NULL,
    updated_at = now();

UPDATE job_runs
SET status = 'cancelled',
    locked_at = NULL,
    locked_by = NULL,
    lease_expires_at = NULL,
    last_error = 'Cancelled by migration 42 rollback; explicitly enable outreach after rollback verification.',
    updated_at = now()
WHERE job_type = 'outreach.bulk_send'
  AND status IN ('queued', 'running');

DROP TRIGGER IF EXISTS restaurants_outreach_sequence_enrollment ON restaurants;
DROP FUNCTION IF EXISTS enroll_restaurant_in_outreach_sequence();
DROP FUNCTION IF EXISTS ensure_outreach_sequence_enrollment(uuid);

DELETE FROM email_campaigns
WHERE campaign_type = 'outreach' AND sequence_id IS NOT NULL;

UPDATE restaurants
SET status = 'demo_ready', updated_at = now()
WHERE status = 'lead'
  AND outreach_consent_evidence ->> 'previous_lifecycle' = 'demo_ready'
  AND outreach_consent_evidence ->> 'lifecycle_normalized_by_migration' = '42';

DROP INDEX IF EXISTS email_campaigns_due_sequence;
DROP INDEX IF EXISTS email_campaigns_one_sequence_enrollment;

ALTER TABLE email_campaigns
  DROP CONSTRAINT IF EXISTS email_campaigns_sequence_progress_check,
  DROP COLUMN IF EXISTS completed_at,
  DROP COLUMN IF EXISTS next_send_at,
  DROP COLUMN IF EXISTS next_step,
  DROP COLUMN IF EXISTS sequence_id;

ALTER TABLE email_campaigns
  ALTER COLUMN demo_site_id SET NOT NULL;

ALTER TABLE restaurants
  DROP CONSTRAINT IF EXISTS restaurants_outreach_consent_source_check,
  DROP CONSTRAINT IF EXISTS restaurants_outreach_consent_basis_check,
  DROP COLUMN IF EXISTS outreach_consent_evidence,
  DROP COLUMN IF EXISTS outreach_consent_recorded_at,
  DROP COLUMN IF EXISTS outreach_consent_source,
  DROP COLUMN IF EXISTS outreach_consent_basis;

DROP TABLE IF EXISTS outreach_email_sequence_steps;
DROP TABLE IF EXISTS outreach_email_sequences;
