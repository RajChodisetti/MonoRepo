-- Bind delivery audit and unsubscribe behavior to the address actually used
-- for the send, rather than the restaurant's mutable current email.

ALTER TABLE email_tracking_tokens
  ADD COLUMN IF NOT EXISTS recipient_email text NOT NULL DEFAULT '';

ALTER TABLE email_delivery_attempts
  ADD COLUMN IF NOT EXISTS recipient_email text NOT NULL DEFAULT '';

ALTER TABLE restaurants
  ADD COLUMN IF NOT EXISTS last_email_recipient text NOT NULL DEFAULT '';

-- Existing rows predate recipient snapshots and cannot be backfilled safely
-- from the restaurant's mutable current email. NOT VALID preserves those rows
-- for the documented legacy opt-out fallback/provider reconciliation while
-- enforcing normalized, non-empty recipients on every new tracking token and
-- delivery attempt.
ALTER TABLE email_tracking_tokens
  DROP CONSTRAINT IF EXISTS email_tracking_tokens_recipient_email_check;

ALTER TABLE email_tracking_tokens
  ADD CONSTRAINT email_tracking_tokens_recipient_email_check
  CHECK (
    recipient_email <> ''
    AND recipient_email = lower(trim(recipient_email))
  ) NOT VALID;

ALTER TABLE email_delivery_attempts
  DROP CONSTRAINT IF EXISTS email_delivery_attempts_recipient_email_check;

ALTER TABLE email_delivery_attempts
  ADD CONSTRAINT email_delivery_attempts_recipient_email_check
  CHECK (
    recipient_email <> ''
    AND recipient_email = lower(trim(recipient_email))
  ) NOT VALID;

ALTER TABLE restaurants
  DROP CONSTRAINT IF EXISTS restaurants_last_email_recipient_check;

ALTER TABLE restaurants
  ADD CONSTRAINT restaurants_last_email_recipient_check
  CHECK (
    last_email_recipient = ''
    OR last_email_recipient = lower(trim(last_email_recipient))
  ) NOT VALID;

CREATE INDEX IF NOT EXISTS idx_email_tracking_tokens_recipient
  ON email_tracking_tokens (recipient_email)
  WHERE recipient_email <> '';

CREATE INDEX IF NOT EXISTS idx_email_delivery_attempts_recipient
  ON email_delivery_attempts (recipient_email, created_at DESC)
  WHERE recipient_email <> '';
