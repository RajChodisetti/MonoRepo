-- Outbound send snapshots and inbound reply capture for the admin inbox.

CREATE TABLE IF NOT EXISTS email_messages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid REFERENCES restaurants(id) ON DELETE CASCADE,
  campaign_id uuid REFERENCES email_campaigns(id) ON DELETE SET NULL,
  delivery_attempt_id uuid REFERENCES email_delivery_attempts(id) ON DELETE SET NULL,
  reply_token uuid,
  direction text NOT NULL,
  from_email text NOT NULL DEFAULT '',
  to_email text NOT NULL DEFAULT '',
  reply_to text NOT NULL DEFAULT '',
  subject text NOT NULL DEFAULT '',
  body_text text NOT NULL DEFAULT '',
  gmail_message_id text NOT NULL DEFAULT '',
  gmail_thread_id text NOT NULL DEFAULT '',
  rfc_message_id text NOT NULL DEFAULT '',
  mailbox_key text NOT NULL DEFAULT '',
  unmatched boolean NOT NULL DEFAULT false,
  read_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT email_messages_direction_check CHECK (direction IN ('outbound', 'inbound'))
);

CREATE UNIQUE INDEX IF NOT EXISTS email_messages_outbound_reply_token_unique
  ON email_messages (reply_token)
  WHERE direction = 'outbound' AND reply_token IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS email_messages_gmail_message_id_unique
  ON email_messages (gmail_message_id)
  WHERE gmail_message_id <> '';

CREATE UNIQUE INDEX IF NOT EXISTS email_messages_outbound_attempt_unique
  ON email_messages (delivery_attempt_id)
  WHERE direction = 'outbound' AND delivery_attempt_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS email_messages_restaurant_created
  ON email_messages (restaurant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS email_messages_inbox_unread
  ON email_messages (created_at DESC)
  WHERE direction = 'inbound' AND read_at IS NULL;

CREATE INDEX IF NOT EXISTS email_messages_rfc_message_id
  ON email_messages (rfc_message_id)
  WHERE rfc_message_id <> '';

CREATE INDEX IF NOT EXISTS email_messages_gmail_thread_id
  ON email_messages (gmail_thread_id)
  WHERE gmail_thread_id <> '';

CREATE TABLE IF NOT EXISTS outreach_inbound_sync (
  mailbox_key text PRIMARY KEY,
  history_id text NOT NULL DEFAULT '',
  updated_at timestamptz NOT NULL DEFAULT now()
);
