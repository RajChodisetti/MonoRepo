-- Email outreach campaigns, tracking, and suppressions.

ALTER TABLE job_runs ADD COLUMN IF NOT EXISTS locked_at timestamptz;

CREATE TABLE IF NOT EXISTS email_campaigns (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  demo_site_id uuid NOT NULL REFERENCES demo_sites(id) ON DELETE RESTRICT,
  campaign_type text NOT NULL DEFAULT 'outreach',
  status text NOT NULL DEFAULT 'draft',
  current_step integer NOT NULL DEFAULT 0,
  subject text NOT NULL DEFAULT '',
  body_html text NOT NULL DEFAULT '',
  body_text text NOT NULL DEFAULT '',
  demo_token text NOT NULL DEFAULT '',
  approved_at timestamptz,
  approved_by uuid REFERENCES users(id) ON DELETE SET NULL,
  last_sent_at timestamptz,
  stopped_reason text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_email_campaigns_restaurant_id ON email_campaigns (restaurant_id);
CREATE INDEX IF NOT EXISTS idx_email_campaigns_status ON email_campaigns (status);

CREATE TABLE IF NOT EXISTS email_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id uuid NOT NULL REFERENCES email_campaigns(id) ON DELETE CASCADE,
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  event_type text NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  event_time timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_email_events_campaign_id_event_type ON email_events (campaign_id, event_type);
CREATE INDEX IF NOT EXISTS idx_email_events_restaurant_id ON email_events (restaurant_id);

CREATE TABLE IF NOT EXISTS email_tracking_tokens (
  token text PRIMARY KEY,
  campaign_id uuid NOT NULL REFERENCES email_campaigns(id) ON DELETE CASCADE,
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  demo_site_id uuid REFERENCES demo_sites(id) ON DELETE SET NULL,
  token_type text NOT NULL DEFAULT 'click',
  target_url text NOT NULL DEFAULT '',
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_email_tracking_tokens_campaign_id ON email_tracking_tokens (campaign_id);

CREATE TABLE IF NOT EXISTS email_suppressions (
  email text PRIMARY KEY,
  reason text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);
