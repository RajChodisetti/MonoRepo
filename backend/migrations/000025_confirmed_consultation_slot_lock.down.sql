DROP INDEX IF EXISTS idx_company_consultations_confirmed_slot;

ALTER TABLE company_consultations
  ADD CONSTRAINT company_consultations_slot_start_key UNIQUE (slot_start);
