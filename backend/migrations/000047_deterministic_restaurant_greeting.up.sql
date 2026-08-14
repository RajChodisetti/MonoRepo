-- Clone the active sequence into an inactive draft that opts only the first
-- enabled email into the deterministic {{greeting01}} merge field.

DO $migration$
DECLARE
  source_sequence_id uuid;
  draft_sequence_id constant uuid := '00000000-0000-4000-8000-000000000047';
  first_step_id uuid;
  first_body text;
  greeting_count integer;
  copied_steps integer;
  repeated_paragraph constant text := 'I had a quick thought after looking at {{restaurant_name}}. The online flow could probably make it easier for guests to get answers and request a table without adding more work for the team.';
  revised_paragraph constant text := 'The online flow could probably make it easier for guests to get answers and request a table without adding more work for the team.';
BEGIN
  SELECT id
  INTO source_sequence_id
  FROM outreach_email_sequences
  WHERE status = 'approved'
    AND is_active = true
  ORDER BY approved_at DESC NULLS LAST, created_at DESC
  LIMIT 1;

  IF source_sequence_id IS NULL THEN
    RAISE EXCEPTION 'migration 47 requires one active approved outreach sequence';
  END IF;

  IF EXISTS (SELECT 1 FROM outreach_email_sequences WHERE id = draft_sequence_id) THEN
    RAISE EXCEPTION 'migration 47 draft sequence already exists';
  END IF;

  INSERT INTO outreach_email_sequences (
    id, name, version, status, is_active, approved_at, approved_by
  )
  SELECT
    draft_sequence_id,
    source.name,
    COALESCE((
      SELECT max(existing.version) + 1
      FROM outreach_email_sequences existing
      WHERE existing.name = source.name
    ), source.version + 1),
    'draft',
    false,
    NULL,
    NULL
  FROM outreach_email_sequences source
  WHERE source.id = source_sequence_id;

  INSERT INTO outreach_email_sequence_steps (
    sequence_id, position, enabled, delay_hours, subject_template, body_text_template
  )
  SELECT
    draft_sequence_id, position, enabled, delay_hours, subject_template, body_text_template
  FROM outreach_email_sequence_steps
  WHERE sequence_id = source_sequence_id
  ORDER BY position;

  GET DIAGNOSTICS copied_steps = ROW_COUNT;
  IF copied_steps = 0 THEN
    RAISE EXCEPTION 'migration 47 active sequence has no steps to clone';
  END IF;

  SELECT id, body_text_template
  INTO first_step_id, first_body
  FROM outreach_email_sequence_steps
  WHERE sequence_id = draft_sequence_id
    AND enabled = true
  ORDER BY position
  LIMIT 1;

  IF first_step_id IS NULL THEN
    RAISE EXCEPTION 'migration 47 active sequence has no enabled step';
  END IF;

  greeting_count := (
    length(first_body) - length(replace(first_body, '{{greeting}}', ''))
  ) / length('{{greeting}}');
  IF greeting_count <> 1 THEN
    RAISE EXCEPTION 'migration 47 expected exactly one {{greeting}} in the first enabled step';
  END IF;
  UPDATE outreach_email_sequence_steps
  SET body_text_template = CASE
        WHEN strpos(first_body, repeated_paragraph) > 0 THEN replace(
          regexp_replace(first_body, '\{\{greeting\}\}', '{{greeting01}}'),
          repeated_paragraph,
          revised_paragraph
        )
        ELSE regexp_replace(first_body, '\{\{greeting\}\}', '{{greeting01}}')
      END,
      updated_at = now()
  WHERE id = first_step_id;

  IF EXISTS (
    SELECT 1
    FROM outreach_email_sequences
    WHERE id = draft_sequence_id
      AND (status <> 'draft' OR is_active OR approved_at IS NOT NULL OR approved_by IS NOT NULL)
  ) THEN
    RAISE EXCEPTION 'migration 47 draft activation guard failed';
  END IF;
END
$migration$;
