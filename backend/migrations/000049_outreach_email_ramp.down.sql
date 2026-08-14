ALTER TABLE outreach_email_accounts
  DROP CONSTRAINT IF EXISTS outreach_email_accounts_ramp_day_check;

ALTER TABLE outreach_email_accounts
  DROP COLUMN IF EXISTS ramp_day;
