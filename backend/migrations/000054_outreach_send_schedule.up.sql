-- Persist the admin-managed daily send window used only by quota-managed
-- scheduled outreach. Direct/test/reply/health email paths remain independent.

-- Schema changes to delivery policy fail closed. An administrator must review
-- the saved window and explicitly re-enable outreach after deployment.
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
    last_error = 'Cancelled while the admin-managed outreach send window was deployed.',
    updated_at = now()
WHERE job_type = 'outreach.bulk_send'
  AND status IN ('queued', 'running');

CREATE TABLE outreach_send_schedule (
  singleton smallint PRIMARY KEY DEFAULT 1,
  timezone text NOT NULL DEFAULT 'Australia/Sydney',
  start_minute smallint NOT NULL DEFAULT 420,
  end_minute smallint NOT NULL DEFAULT 720,
  updated_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT outreach_send_schedule_singleton_check CHECK (singleton = 1),
  CONSTRAINT outreach_send_schedule_timezone_check CHECK (timezone = 'Australia/Sydney'),
  CONSTRAINT outreach_send_schedule_start_check CHECK (start_minute BETWEEN 0 AND 1438),
  CONSTRAINT outreach_send_schedule_end_check CHECK (end_minute BETWEEN 1 AND 1439),
  CONSTRAINT outreach_send_schedule_order_check CHECK (end_minute > start_minute),
  CONSTRAINT outreach_send_schedule_duration_check CHECK (end_minute - start_minute >= 60)
);

INSERT INTO outreach_send_schedule (singleton, timezone, start_minute, end_minute)
VALUES (1, 'Australia/Sydney', 420, 720);
