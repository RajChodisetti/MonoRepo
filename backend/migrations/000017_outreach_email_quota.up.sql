-- Durable outreach email account quotas, delivery audit, and confirmed-send counters.

CREATE SEQUENCE IF NOT EXISTS email_send_sequence;

ALTER TABLE restaurants
  ADD COLUMN IF NOT EXISTS email_send_count integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_email_sent_at timestamptz,
  ADD COLUMN IF NOT EXISTS last_email_send_sequence bigint;

ALTER TABLE restaurants
  DROP CONSTRAINT IF EXISTS restaurants_email_send_count_check;

ALTER TABLE restaurants
  ADD CONSTRAINT restaurants_email_send_count_check CHECK (email_send_count >= 0);

WITH sent_events AS (
  SELECT restaurant_id, count(*)::integer AS sent_count, max(event_time) AS last_sent_at
  FROM email_events
  WHERE event_type = 'sent'
  GROUP BY restaurant_id
)
UPDATE restaurants AS r
SET email_send_count = GREATEST(
      r.email_send_count,
      CASE WHEN r.email_sent THEN 1 ELSE 0 END,
      COALESCE(sent_events.sent_count, 0)
    ),
    last_email_sent_at = COALESCE(r.last_email_sent_at, sent_events.last_sent_at),
    email_sent = r.email_sent OR sent_events.sent_count > 0
FROM sent_events
WHERE sent_events.restaurant_id = r.id;

UPDATE restaurants
SET email_send_count = 1
WHERE email_sent = true AND email_send_count = 0;

CREATE TABLE IF NOT EXISTS outreach_email_accounts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  account_key text NOT NULL UNIQUE,
  provider text NOT NULL DEFAULT 'zoho',
  provider_identity text NOT NULL UNIQUE,
  from_email text NOT NULL DEFAULT '',
  position integer NOT NULL DEFAULT 0,
  enabled boolean NOT NULL DEFAULT true,
  send_limit integer NOT NULL DEFAULT 40,
  cycle_number bigint NOT NULL DEFAULT 1,
  usage_count integer NOT NULL DEFAULT 0,
  cycle_started_at timestamptz,
  available_at timestamptz NOT NULL DEFAULT now(),
  last_used_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT outreach_email_accounts_position_check CHECK (position >= 0),
  CONSTRAINT outreach_email_accounts_send_limit_check CHECK (send_limit BETWEEN 1 AND 40),
  CONSTRAINT outreach_email_accounts_cycle_number_check CHECK (cycle_number >= 1),
  CONSTRAINT outreach_email_accounts_usage_count_check CHECK (usage_count BETWEEN 0 AND send_limit)
);

CREATE INDEX IF NOT EXISTS idx_outreach_email_accounts_available
  ON outreach_email_accounts (enabled, available_at, position);

CREATE TABLE IF NOT EXISTS email_delivery_attempts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  send_sequence bigint NOT NULL DEFAULT nextval('email_send_sequence'),
  campaign_id uuid NOT NULL REFERENCES email_campaigns(id) ON DELETE CASCADE,
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  account_id uuid NOT NULL REFERENCES outreach_email_accounts(id) ON DELETE RESTRICT,
  bulk_job_id uuid REFERENCES job_runs(id) ON DELETE SET NULL,
  campaign_step integer NOT NULL DEFAULT 0,
  account_cycle bigint NOT NULL,
  account_sequence integer NOT NULL,
  status text NOT NULL DEFAULT 'sending',
  provider_message_id text NOT NULL DEFAULT '',
  error_code text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  sent_at timestamptz,
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT email_delivery_attempts_send_sequence_unique UNIQUE (send_sequence),
  CONSTRAINT email_delivery_attempts_campaign_step_check CHECK (campaign_step >= 0),
  CONSTRAINT email_delivery_attempts_account_cycle_check CHECK (account_cycle >= 1),
  CONSTRAINT email_delivery_attempts_account_sequence_check CHECK (account_sequence BETWEEN 1 AND 40),
  CONSTRAINT email_delivery_attempts_status_check CHECK (
    status IN ('sending', 'sent', 'skipped', 'failed', 'unknown')
  )
);

CREATE INDEX IF NOT EXISTS idx_email_delivery_attempts_campaign
  ON email_delivery_attempts (campaign_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_email_delivery_attempts_restaurant
  ON email_delivery_attempts (restaurant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_email_delivery_attempts_account
  ON email_delivery_attempts (account_id, account_cycle, account_sequence);

CREATE UNIQUE INDEX IF NOT EXISTS idx_email_delivery_attempts_active_campaign_step
  ON email_delivery_attempts (campaign_id, campaign_step)
  WHERE status IN ('sending', 'sent');

ALTER TABLE email_events
  ADD COLUMN IF NOT EXISTS delivery_attempt_id uuid REFERENCES email_delivery_attempts(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_email_events_delivery_attempt_event
  ON email_events (delivery_attempt_id, event_type)
  WHERE delivery_attempt_id IS NOT NULL;
