-- Daily Gmail sender verification plus restaurant demo engagement evidence.

CREATE TABLE IF NOT EXISTS outreach_email_account_health (
  account_key text PRIMARY KEY,
  provider text NOT NULL,
  provider_identity text NOT NULL UNIQUE,
  from_email text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  health_status text NOT NULL DEFAULT 'pending',
  last_checked_at timestamptz,
  next_check_at timestamptz,
  provider_message_id text NOT NULL DEFAULT '',
  last_error text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT outreach_email_account_health_status_check
    CHECK (health_status IN ('pending', 'checking', 'healthy', 'failed', 'disabled'))
);

CREATE INDEX IF NOT EXISTS idx_outreach_email_account_health_due
  ON outreach_email_account_health (next_check_at, account_key)
  WHERE enabled = true;

CREATE TABLE IF NOT EXISTS outreach_runtime_control (
  control_key text PRIMARY KEY,
  enabled boolean NOT NULL DEFAULT false,
  enabled_at timestamptz,
  enabled_by uuid REFERENCES users(id) ON DELETE SET NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT outreach_runtime_control_key_check
    CHECK (control_key = 'email_job')
);

INSERT INTO outreach_runtime_control (control_key, enabled)
VALUES ('email_job', false)
ON CONFLICT (control_key) DO NOTHING;

CREATE TABLE IF NOT EXISTS demo_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  demo_site_id uuid REFERENCES demo_sites(id) ON DELETE CASCADE,
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  template_id text NOT NULL,
  session_token_hash text NOT NULL,
  started_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  ended_at timestamptz,
  duration_seconds integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT demo_sessions_duration_check
    CHECK (duration_seconds >= 0 AND duration_seconds <= 86400),
  CONSTRAINT demo_sessions_template_check
    CHECK (template_id IN ('1', '2', '3'))
);

CREATE INDEX IF NOT EXISTS idx_demo_sessions_restaurant_started
  ON demo_sessions (restaurant_id, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_demo_sessions_demo_site_started
  ON demo_sessions (demo_site_id, started_at DESC);

CREATE TABLE IF NOT EXISTS demo_session_transcripts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id uuid NOT NULL REFERENCES demo_sessions(id) ON DELETE CASCADE,
  role text NOT NULL,
  content text NOT NULL,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT demo_session_transcripts_role_check
    CHECK (role IN ('user', 'assistant', 'system')),
  CONSTRAINT demo_session_transcripts_content_check
    CHECK (length(trim(content)) BETWEEN 1 AND 4000)
);

CREATE INDEX IF NOT EXISTS idx_demo_session_transcripts_session_time
  ON demo_session_transcripts (session_id, occurred_at ASC);
