ALTER TABLE company_consultations
  DROP CONSTRAINT IF EXISTS company_consultations_slot_start_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_company_consultations_confirmed_slot
  ON company_consultations (slot_start)
  WHERE status = 'confirmed';
