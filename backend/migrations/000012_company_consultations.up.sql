-- Company sales consultations booked from the Tuvi website and corporate voice agent.

CREATE TABLE IF NOT EXISTS company_consultations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  confirmation_code text NOT NULL UNIQUE,
  slot_start timestamptz NOT NULL UNIQUE,
  slot_end timestamptz NOT NULL,
  prospect_name text NOT NULL,
  prospect_email text NOT NULL DEFAULT '',
  prospect_phone text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled')),
  google_event_id text NOT NULL DEFAULT '',
  source text NOT NULL DEFAULT 'voice' CHECK (source IN ('voice', 'web')),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_company_consultations_status_slot
  ON company_consultations (status, slot_start);
