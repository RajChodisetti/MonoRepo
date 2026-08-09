-- Per-slot consultation availability managed from the internal admin calendar.
-- Missing rows intentionally retain the configured business-hours default.

CREATE TABLE IF NOT EXISTS company_consultation_slot_overrides (
  slot_start timestamptz PRIMARY KEY,
  is_available boolean NOT NULL,
  updated_by uuid REFERENCES users(id) ON DELETE SET NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);
