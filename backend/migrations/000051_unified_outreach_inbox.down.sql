DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM email_messages
    WHERE gmail_message_id <> ''
    GROUP BY gmail_message_id
    HAVING count(*) > 1
  ) THEN
    RAISE EXCEPTION 'refusing to remove migration 51 because provider message ids overlap across mailboxes';
  END IF;
END $$;

DROP INDEX IF EXISTS email_messages_mailbox_provider_id_unique;

CREATE UNIQUE INDEX IF NOT EXISTS email_messages_gmail_message_id_unique
  ON email_messages (gmail_message_id)
  WHERE gmail_message_id <> '';

DROP INDEX IF EXISTS email_messages_inbox_received;

ALTER TABLE email_messages
  DROP COLUMN IF EXISTS received_at;

ALTER TABLE outreach_inbound_sync
  DROP COLUMN IF EXISTS last_attempt_at,
  DROP COLUMN IF EXISTS last_success_at,
  DROP COLUMN IF EXISTS last_error;
