BEGIN;

-- This migration changes the signature used at the provider boundary. Keep
-- real outreach fail-closed until an administrator reviews and activates the
-- new draft explicitly.
INSERT INTO outreach_runtime_control (
  control_key, enabled, enabled_at, enabled_by, updated_at
)
VALUES ('email_job', false, NULL, NULL, now())
ON CONFLICT (control_key) DO UPDATE
SET enabled = false,
    enabled_at = NULL,
    enabled_by = NULL,
    updated_at = now();

UPDATE job_runs
SET status = 'cancelled',
    locked_at = NULL,
    locked_by = NULL,
    lease_expires_at = NULL,
    last_error = 'Cancelled by migration 53; review the placeholder copy and Praveen signature before re-enabling outreach.',
    updated_at = now()
WHERE job_type = 'outreach.bulk_send'
  AND status IN ('queued', 'running');

ALTER TABLE outreach_email_sequences
  ADD COLUMN signature_name text NOT NULL DEFAULT 'Praveen Maurya',
  ADD COLUMN signature_title text NOT NULL DEFAULT 'Business Development Manager',
  ADD COLUMN signature_details text NOT NULL DEFAULT '',
  ADD CONSTRAINT outreach_email_sequences_signature_name_check
    CHECK (length(trim(signature_name)) BETWEEN 1 AND 120 AND signature_name !~ E'[\r\n]'),
  ADD CONSTRAINT outreach_email_sequences_signature_title_check
    CHECK (length(signature_title) <= 160 AND signature_title !~ E'[\r\n]'),
  ADD CONSTRAINT outreach_email_sequences_signature_details_check
    CHECK (length(signature_details) <= 1000 AND signature_details !~* '<[a-z][^>]*>');

DO $migration$
DECLARE
  source_sequence_id uuid;
  draft_sequence_id constant uuid := '00000000-0000-4000-8000-000000000053';
  first_step_id uuid;
  copied_steps integer;
BEGIN
  SELECT id
  INTO source_sequence_id
  FROM outreach_email_sequences
  WHERE status = 'approved'
    AND is_active = true
  ORDER BY approved_at DESC NULLS LAST, created_at DESC
  LIMIT 1;

  IF source_sequence_id IS NULL THEN
    RAISE EXCEPTION 'migration 53 requires one active approved outreach sequence';
  END IF;

  IF EXISTS (SELECT 1 FROM outreach_email_sequences WHERE id = draft_sequence_id) THEN
    RAISE EXCEPTION 'migration 53 draft sequence already exists';
  END IF;

  INSERT INTO outreach_email_sequences (
    id, name, version, status, is_active, signature_name, signature_title,
    signature_details, approved_at, approved_by
  )
  SELECT
    draft_sequence_id,
    'Tuvi digital growth introduction',
    COALESCE((
      SELECT max(existing.version) + 1
      FROM outreach_email_sequences existing
      WHERE existing.name = 'Tuvi digital growth introduction'
    ), 1),
    'draft',
    false,
    'Praveen Maurya',
    'Business Development Manager',
    '',
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
    RAISE EXCEPTION 'migration 53 active sequence has no steps to clone';
  END IF;

  SELECT id
  INTO first_step_id
  FROM outreach_email_sequence_steps
  WHERE sequence_id = draft_sequence_id
    AND enabled = true
  ORDER BY position
  LIMIT 1;

  IF first_step_id IS NULL THEN
    RAISE EXCEPTION 'migration 53 active sequence has no enabled step';
  END IF;

  UPDATE outreach_email_sequence_steps
  SET subject_template = 'A digital growth idea for [RESTAURANT_NAME]',
      body_text_template = $template_1$[GREETING]

Did you know that optimizing your digital presence could lead to 30% more revenue?

I Praveen from Tuvi Solutions an Australian digital partner delivering excellence through custom solutions, modern websites, SEO, Google reviews engines, QR ordering, online reservations, AI voice receptionists, and reviewed marketing campaigns.

Check your digital score for free on our website (the website link is in my signature below), or reply to this email if you want to schedule a free call so we can map out a quick digital game plan for [RESTAURANT_NAME].

And if this isn't something you're looking at right now, just let me know and I won't follow up further.

Looking forward to your response. Cheers!$template_1$,
      updated_at = created_at
  WHERE id = first_step_id;

  UPDATE outreach_email_sequences
  SET updated_at = created_at
  WHERE id = draft_sequence_id;

  IF EXISTS (
    SELECT 1
    FROM outreach_email_sequences
    WHERE id = draft_sequence_id
      AND (status <> 'draft' OR is_active OR approved_at IS NOT NULL OR approved_by IS NOT NULL)
  ) THEN
    RAISE EXCEPTION 'migration 53 draft activation guard failed';
  END IF;
END
$migration$;

COMMIT;
