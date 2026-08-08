ALTER TABLE demo_sessions
  DROP CONSTRAINT IF EXISTS demo_sessions_template_check;

ALTER TABLE demo_sessions
  ADD CONSTRAINT demo_sessions_template_check
  CHECK (template_id IN ('1', '2', '3', '4'));
