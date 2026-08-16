-- Rollback also fails closed. Preserve a customized schedule until an operator
-- intentionally restores the default 07:00-12:00 Sydney window.
INSERT INTO outreach_runtime_control (control_key, enabled, enabled_at, enabled_by, updated_at)
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
    last_error = 'Cancelled while the admin-managed outreach send window was rolled back.',
    updated_at = now()
WHERE job_type = 'outreach.bulk_send'
  AND status IN ('queued', 'running');

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM outreach_send_schedule
    WHERE singleton = 1
      AND (timezone <> 'Australia/Sydney' OR start_minute <> 420 OR end_minute <> 720)
  ) THEN
    RAISE EXCEPTION 'refusing to remove migration 54 while the outreach send schedule is customized';
  END IF;
END
$$;

DROP TABLE outreach_send_schedule;
