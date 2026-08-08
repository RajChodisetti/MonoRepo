ALTER TABLE demo_sessions
  DROP CONSTRAINT IF EXISTS demo_sessions_template_check;

UPDATE demo_sessions
SET template_id = '3'
WHERE template_id = '4';

ALTER TABLE demo_sessions
  ADD CONSTRAINT demo_sessions_template_check
  CHECK (template_id IN ('1', '2', '3'));
