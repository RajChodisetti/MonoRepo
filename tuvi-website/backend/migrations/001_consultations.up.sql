CREATE TABLE IF NOT EXISTS consultations (
    id UUID PRIMARY KEY,
    confirmation_code TEXT NOT NULL UNIQUE,
    slot_start TIMESTAMPTZ NOT NULL UNIQUE,
    slot_end TIMESTAMPTZ NOT NULL,
    prospect_name TEXT NOT NULL,
    prospect_email TEXT NOT NULL DEFAULT '',
    prospect_phone TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled')),
    google_event_id TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'voice' CHECK (source IN ('voice', 'web')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consultations_status_slot ON consultations (status, slot_start);
