-- Restart-safe outreach pacing: one global velocity gate plus forty durable
-- account slots distributed across an eight-hour window by default.

ALTER TABLE outreach_email_accounts
  ADD COLUMN IF NOT EXISTS send_window_seconds integer NOT NULL DEFAULT 28800,
  ADD COLUMN IF NOT EXISTS send_jitter_min_seconds integer NOT NULL DEFAULT 120,
  ADD COLUMN IF NOT EXISTS send_jitter_max_seconds integer NOT NULL DEFAULT 300;

ALTER TABLE outreach_email_accounts
  DROP CONSTRAINT IF EXISTS outreach_email_accounts_send_window_check,
  DROP CONSTRAINT IF EXISTS outreach_email_accounts_send_jitter_check,
  DROP CONSTRAINT IF EXISTS outreach_email_accounts_slot_width_check;

ALTER TABLE outreach_email_accounts
  ADD CONSTRAINT outreach_email_accounts_send_window_check
    CHECK (send_window_seconds >= 28800),
  ADD CONSTRAINT outreach_email_accounts_send_jitter_check
    CHECK (
      send_jitter_min_seconds >= 120
      AND send_jitter_max_seconds >= send_jitter_min_seconds
    ),
  ADD CONSTRAINT outreach_email_accounts_slot_width_check
    CHECK (send_window_seconds > send_limit * send_jitter_max_seconds);

-- A partially used account created by the old immediate-loop sender must not
-- burst its remaining allowance as soon as this migration is deployed. Anchor
-- its existing usage to the new slot width and require at least the minimum
-- delay before another claim.
UPDATE outreach_email_accounts
SET cycle_started_at = clock_timestamp()
      - make_interval(secs => usage_count * send_window_seconds / send_limit),
    available_at = GREATEST(
      available_at,
      clock_timestamp() + (send_jitter_min_seconds * interval '1 second')
    ),
    updated_at = clock_timestamp()
WHERE usage_count > 0
  AND usage_count < send_limit;

CREATE TABLE IF NOT EXISTS outreach_email_pacing (
  singleton smallint PRIMARY KEY DEFAULT 1,
  next_send_at timestamptz NOT NULL DEFAULT now(),
  last_reserved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT outreach_email_pacing_singleton_check CHECK (singleton = 1)
);

INSERT INTO outreach_email_pacing (singleton)
VALUES (1)
ON CONFLICT (singleton) DO NOTHING;
