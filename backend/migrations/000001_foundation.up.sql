CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS app_metadata (
  key text PRIMARY KEY,
  value jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO app_metadata (key, value)
VALUES ('schema_baseline', '{"phase":"p1-e01"}'::jsonb)
ON CONFLICT (key)
DO UPDATE SET value = EXCLUDED.value, updated_at = now();

CREATE TABLE IF NOT EXISTS job_runs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  job_type text NOT NULL,
  status text NOT NULL,
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  idempotency_key text,
  attempts integer NOT NULL DEFAULT 0,
  max_attempts integer NOT NULL DEFAULT 1,
  last_error text,
  available_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT job_runs_status_check CHECK (status IN ('queued', 'running', 'completed', 'failed', 'cancelled')),
  CONSTRAINT job_runs_attempts_check CHECK (attempts >= 0 AND max_attempts >= 1)
);

CREATE INDEX IF NOT EXISTS idx_job_runs_status_available_at
  ON job_runs (status, available_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_job_runs_idempotency_key
  ON job_runs (idempotency_key)
  WHERE idempotency_key IS NOT NULL;
