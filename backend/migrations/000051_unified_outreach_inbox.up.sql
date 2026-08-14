-- Read every configured Gmail outreach mailbox while keeping provider ids,
-- sync health, and the provider-received timestamp mailbox-scoped.

DROP INDEX IF EXISTS email_messages_gmail_message_id_unique;

CREATE UNIQUE INDEX IF NOT EXISTS email_messages_mailbox_provider_id_unique
  ON email_messages (mailbox_key, gmail_message_id)
  WHERE gmail_message_id <> '';

ALTER TABLE email_messages
  ADD COLUMN IF NOT EXISTS received_at timestamptz;

UPDATE email_messages
SET received_at = created_at
WHERE received_at IS NULL;

ALTER TABLE email_messages
  ALTER COLUMN received_at SET DEFAULT now(),
  ALTER COLUMN received_at SET NOT NULL;

CREATE INDEX IF NOT EXISTS email_messages_inbox_received
  ON email_messages (received_at DESC)
  WHERE direction = 'inbound';

ALTER TABLE outreach_inbound_sync
  ADD COLUMN IF NOT EXISTS last_attempt_at timestamptz,
  ADD COLUMN IF NOT EXISTS last_success_at timestamptz,
  ADD COLUMN IF NOT EXISTS last_error text NOT NULL DEFAULT '';
