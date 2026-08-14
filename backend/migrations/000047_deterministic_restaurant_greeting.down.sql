-- Remove only the untouched, inactive draft created by migration 47.

DO $migration$
DECLARE
  draft_sequence_id constant uuid := '00000000-0000-4000-8000-000000000047';
  draft_status text;
  draft_active boolean;
  draft_approved_at timestamptz;
  draft_approved_by uuid;
  draft_created_at timestamptz;
  draft_updated_at timestamptz;
  first_step_id uuid;
  first_body text;
BEGIN
  SELECT status, is_active, approved_at, approved_by, created_at, updated_at
  INTO draft_status, draft_active, draft_approved_at, draft_approved_by,
       draft_created_at, draft_updated_at
  FROM outreach_email_sequences
  WHERE id = draft_sequence_id;

  IF NOT FOUND THEN
    RETURN;
  END IF;

  IF draft_status <> 'draft' OR draft_active OR draft_approved_at IS NOT NULL OR
     draft_approved_by IS NOT NULL OR draft_updated_at <> draft_created_at THEN
    RAISE EXCEPTION 'refusing to remove migration 47 draft because it was activated or changed';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM outreach_email_sequence_steps
    WHERE sequence_id = draft_sequence_id
      AND updated_at <> created_at
  ) THEN
    RAISE EXCEPTION 'refusing to remove migration 47 draft because its steps changed';
  END IF;

  SELECT id, body_text_template
  INTO first_step_id, first_body
  FROM outreach_email_sequence_steps
  WHERE sequence_id = draft_sequence_id
    AND enabled = true
  ORDER BY position
  LIMIT 1;

  IF first_step_id IS NULL OR
     (length(first_body) - length(replace(first_body, '{{greeting01}}', ''))) /
       length('{{greeting01}}') <> 1 OR
     strpos(first_body, 'The online flow could probably make it easier for guests to get answers and request a table without adding more work for the team.') = 0 OR
     EXISTS (
       SELECT 1
       FROM outreach_email_sequence_steps
       WHERE sequence_id = draft_sequence_id
         AND id <> first_step_id
         AND body_text_template LIKE '%{{greeting01}}%'
     ) THEN
    RAISE EXCEPTION 'refusing to remove migration 47 draft because its deterministic greeting copy changed';
  END IF;

  DELETE FROM outreach_email_sequences
  WHERE id = draft_sequence_id;
END
$migration$;
