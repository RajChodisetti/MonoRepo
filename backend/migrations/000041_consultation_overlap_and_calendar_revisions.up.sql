-- Prevent confirmed consultation intervals from overlapping, including legacy
-- calls whose starts do not align with the current configured slot grid.
ALTER TABLE company_consultations
  ADD CONSTRAINT company_consultations_valid_interval
  CHECK (slot_end > slot_start);

ALTER TABLE company_consultations
  ADD CONSTRAINT company_consultations_confirmed_no_overlap
  EXCLUDE USING gist (
    tstzrange(slot_start, slot_end, '[)') WITH &&
  )
  WHERE (status = 'confirmed');

-- A month-scoped revision gives the admin calendar optimistic concurrency:
-- every save compares and increments this value in the same transaction as
-- the slot override replacement.
CREATE TABLE company_consultation_calendar_months (
  month date PRIMARY KEY,
  revision bigint NOT NULL DEFAULT 0 CHECK (revision >= 0),
  updated_by uuid REFERENCES users(id) ON DELETE SET NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT company_consultation_calendar_months_first_day
    CHECK (month = date_trunc('month', month)::date)
);
