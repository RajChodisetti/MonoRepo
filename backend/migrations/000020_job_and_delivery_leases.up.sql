-- Recoverable worker leases and durable uniqueness for outreach scheduling.

ALTER TABLE job_runs
  ADD COLUMN IF NOT EXISTS locked_by text,
  ADD COLUMN IF NOT EXISTS lease_expires_at timestamptz;

-- Lead preparation is transactionally idempotent and safe to retry. Other
-- legacy running jobs may have crossed an external boundary, so fail closed
-- for explicit operator reconciliation instead of risking a duplicate action.
UPDATE job_runs
SET status = 'queued',
    available_at = now(),
    attempts = 0,
    locked_at = NULL,
    locked_by = NULL,
    lease_expires_at = NULL,
    last_error = 'Recovered an unfenced lead preparation job during lease migration.',
    updated_at = now()
WHERE status = 'running'
  AND job_type = 'lead.prepare';

UPDATE job_runs
SET status = 'failed',
    locked_at = NULL,
    locked_by = NULL,
    lease_expires_at = NULL,
    last_error = 'Legacy unfenced running job requires operator reconciliation.',
    updated_at = now()
WHERE status = 'running';

CREATE INDEX IF NOT EXISTS idx_job_runs_expired_lease
  ON job_runs (lease_expires_at)
  WHERE status = 'running';

-- Collapse any pre-existing duplicate active bulk rows before enforcing the
-- one-workflow invariant.
WITH ranked AS (
  SELECT id,
         row_number() OVER (ORDER BY created_at, id) AS row_number
  FROM job_runs
  WHERE job_type = 'outreach.bulk_send'
    AND status IN ('queued', 'running')
)
UPDATE job_runs AS job
SET status = 'failed',
    last_error = 'Superseded duplicate active bulk outreach job.',
    locked_at = NULL,
    locked_by = NULL,
    lease_expires_at = NULL,
    updated_at = now()
FROM ranked
WHERE job.id = ranked.id
  AND ranked.row_number > 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_job_runs_one_active_bulk_outreach
  ON job_runs ((1))
  WHERE job_type = 'outreach.bulk_send'
    AND status IN ('queued', 'running');

ALTER TABLE email_delivery_attempts
  ADD COLUMN IF NOT EXISTS lease_expires_at timestamptz;

-- A pre-migration sending row has an ambiguous external outcome. Reconcile it
-- to unknown and never retry it automatically.
WITH candidates AS (
  SELECT id, error_code AS prior_error_code
  FROM email_delivery_attempts
  WHERE status = 'sending'
  FOR UPDATE
), stale AS (
  UPDATE email_delivery_attempts AS attempt
  SET status = 'unknown',
      error_code = 'legacy_sending_attempt_reconciled',
      lease_expires_at = NULL,
      updated_at = now()
  FROM candidates
  WHERE attempt.id = candidates.id
  RETURNING attempt.id, attempt.campaign_id, attempt.restaurant_id,
            candidates.prior_error_code
)
INSERT INTO email_events (
  campaign_id, restaurant_id, event_type, metadata, delivery_attempt_id
)
SELECT campaign_id,
       restaurant_id,
       'failed',
       jsonb_build_object(
         'error_code', 'legacy_sending_attempt_reconciled',
         'prior_error_code', prior_error_code,
         'delivery_attempt_id', id,
         'bulk_outreach', true
       ),
       id
FROM stale
ON CONFLICT (delivery_attempt_id, event_type)
WHERE delivery_attempt_id IS NOT NULL
DO NOTHING;

UPDATE email_campaigns AS campaign
SET status = 'send_unknown',
    updated_at = now()
WHERE campaign.status = 'sending'
  AND EXISTS (
    SELECT 1
    FROM email_delivery_attempts attempt
    WHERE attempt.campaign_id = campaign.id
      AND attempt.status = 'unknown'
      AND attempt.error_code = 'legacy_sending_attempt_reconciled'
  );

-- A pre-lease campaign can also be stuck at sending without a delivery-attempt
-- row (for example, an old queued email.send job). Its external outcome cannot
-- be proven, so fail closed instead of reopening it for an automatic retry.
WITH unresolved AS (
  UPDATE email_campaigns
  SET status = 'send_unknown',
      updated_at = now()
  WHERE status = 'sending'
  RETURNING id, restaurant_id
)
INSERT INTO email_events (
  campaign_id, restaurant_id, event_type, metadata
)
SELECT id,
       restaurant_id,
       'failed',
       jsonb_build_object(
         'error_code', 'legacy_sending_campaign_reconciled',
         'legacy_email_send', true
       )
FROM unresolved;

UPDATE job_runs
SET status = 'failed',
    last_error = 'Legacy email.send job disabled; use quota-managed bulk outreach.',
    locked_at = NULL,
    locked_by = NULL,
    lease_expires_at = NULL,
    updated_at = now()
WHERE job_type = 'email.send'
  AND status = 'queued';

ALTER TABLE email_delivery_attempts
  DROP CONSTRAINT IF EXISTS email_delivery_attempts_lease_state_check;

ALTER TABLE email_delivery_attempts
  ADD CONSTRAINT email_delivery_attempts_lease_state_check CHECK (
    (status = 'sending' AND lease_expires_at IS NOT NULL)
    OR (status <> 'sending' AND lease_expires_at IS NULL)
  );

CREATE INDEX IF NOT EXISTS idx_email_delivery_attempts_expired_lease
  ON email_delivery_attempts (lease_expires_at)
  WHERE status = 'sending';

CREATE UNIQUE INDEX IF NOT EXISTS idx_email_delivery_attempts_account_slot
  ON email_delivery_attempts (account_id, account_cycle, account_sequence);
