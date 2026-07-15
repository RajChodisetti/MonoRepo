DROP INDEX IF EXISTS idx_email_events_delivery_attempt_event;

ALTER TABLE email_events
  DROP COLUMN IF EXISTS delivery_attempt_id;

DROP TABLE IF EXISTS email_delivery_attempts;
DROP TABLE IF EXISTS outreach_email_accounts;

ALTER TABLE restaurants
  DROP CONSTRAINT IF EXISTS restaurants_email_send_count_check,
  DROP COLUMN IF EXISTS last_email_send_sequence,
  DROP COLUMN IF EXISTS last_email_sent_at,
  DROP COLUMN IF EXISTS email_send_count;

DROP SEQUENCE IF EXISTS email_send_sequence;
