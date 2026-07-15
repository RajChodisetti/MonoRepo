DROP INDEX IF EXISTS idx_email_delivery_attempts_recipient;
DROP INDEX IF EXISTS idx_email_tracking_tokens_recipient;

ALTER TABLE email_delivery_attempts
  DROP CONSTRAINT IF EXISTS email_delivery_attempts_recipient_email_check;

ALTER TABLE email_tracking_tokens
  DROP CONSTRAINT IF EXISTS email_tracking_tokens_recipient_email_check;

ALTER TABLE restaurants
  DROP CONSTRAINT IF EXISTS restaurants_last_email_recipient_check,
  DROP COLUMN IF EXISTS last_email_recipient;

ALTER TABLE email_delivery_attempts
  DROP COLUMN IF EXISTS recipient_email;

ALTER TABLE email_tracking_tokens
  DROP COLUMN IF EXISTS recipient_email;
