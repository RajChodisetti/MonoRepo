BEGIN;

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
    last_error = 'Cancelled by migration 53 rollback; verify the prior signature before re-enabling outreach.',
    updated_at = now()
WHERE job_type = 'outreach.bulk_send'
  AND status IN ('queued', 'running');

DO $migration$
DECLARE
  draft_sequence_id constant uuid := '00000000-0000-4000-8000-000000000053';
BEGIN
  IF EXISTS (
    SELECT 1
    FROM outreach_email_sequences
    WHERE id = draft_sequence_id
      AND (
        status <> 'draft' OR is_active OR approved_at IS NOT NULL OR approved_by IS NOT NULL OR
        updated_at <> created_at OR signature_name <> 'Praveen Maurya' OR
        signature_title <> 'Business Development Manager' OR signature_details <> ''
      )
  ) OR EXISTS (
    SELECT 1
    FROM outreach_email_sequence_steps
    WHERE sequence_id = draft_sequence_id
      AND updated_at <> created_at
  ) THEN
    RAISE EXCEPTION 'refusing to remove migration 53 draft because it was activated or changed';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM outreach_email_sequences
    WHERE id <> draft_sequence_id
      AND (
        signature_name <> 'Praveen Maurya' OR
        signature_title <> 'Business Development Manager' OR
        signature_details <> ''
      )
  ) THEN
    RAISE EXCEPTION 'refusing to drop migration 53 signature fields because signature details were customized';
  END IF;

  DELETE FROM outreach_email_sequences WHERE id = draft_sequence_id;
END
$migration$;

ALTER TABLE outreach_email_sequences
  DROP CONSTRAINT outreach_email_sequences_signature_details_check,
  DROP CONSTRAINT outreach_email_sequences_signature_title_check,
  DROP CONSTRAINT outreach_email_sequences_signature_name_check,
  DROP COLUMN signature_details,
  DROP COLUMN signature_title,
  DROP COLUMN signature_name;

COMMIT;
