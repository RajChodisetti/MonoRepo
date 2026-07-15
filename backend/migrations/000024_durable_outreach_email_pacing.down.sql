DROP TABLE IF EXISTS outreach_email_pacing;

ALTER TABLE outreach_email_accounts
  DROP CONSTRAINT IF EXISTS outreach_email_accounts_slot_width_check,
  DROP CONSTRAINT IF EXISTS outreach_email_accounts_send_jitter_check,
  DROP CONSTRAINT IF EXISTS outreach_email_accounts_send_window_check,
  DROP COLUMN IF EXISTS send_jitter_max_seconds,
  DROP COLUMN IF EXISTS send_jitter_min_seconds,
  DROP COLUMN IF EXISTS send_window_seconds;

