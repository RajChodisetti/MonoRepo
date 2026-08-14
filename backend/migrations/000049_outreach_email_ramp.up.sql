-- Persist a restart-safe per-mailbox warm-up allowance: 5, 10, ... 40.

ALTER TABLE outreach_email_accounts
  ADD COLUMN IF NOT EXISTS ramp_day integer NOT NULL DEFAULT 1;

ALTER TABLE outreach_email_accounts
  DROP CONSTRAINT IF EXISTS outreach_email_accounts_ramp_day_check;

ALTER TABLE outreach_email_accounts
  ADD CONSTRAINT outreach_email_accounts_ramp_day_check
  CHECK (ramp_day BETWEEN 1 AND 8);

-- Existing accounts begin conservatively on day one. If their current cycle
-- already used at least five slots, wait a full cooldown before day two.
UPDATE outreach_email_accounts
SET available_at = GREATEST(available_at, now() + interval '24 hours'),
    updated_at = now()
WHERE ramp_day = 1
  AND usage_count >= LEAST(send_limit, 5);
