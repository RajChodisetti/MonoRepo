DROP TABLE IF EXISTS company_consultation_calendar_months;

ALTER TABLE company_consultations
  DROP CONSTRAINT IF EXISTS company_consultations_confirmed_no_overlap;

ALTER TABLE company_consultations
  DROP CONSTRAINT IF EXISTS company_consultations_valid_interval;
