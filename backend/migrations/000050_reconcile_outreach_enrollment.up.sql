-- Align the stable enrollment seam with the current outreach policy. Legacy
-- suppression rows no longer determine sequence eligibility, but deployment
-- must still fail closed until an administrator explicitly enables sending.
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
    last_error = 'Cancelled by migration 50; verify reconciled outreach enrollments before explicitly enabling sending.',
    updated_at = now()
WHERE job_type = 'outreach.bulk_send'
  AND status = 'queued';

CREATE INDEX IF NOT EXISTS restaurants_normalized_email
  ON restaurants (lower(trim(email)));

CREATE OR REPLACE FUNCTION ensure_outreach_sequence_enrollment(target_restaurant_id uuid)
RETURNS uuid
LANGUAGE plpgsql
AS $$
DECLARE
  active_sequence_id uuid;
  first_step integer;
  enrollment_id uuid;
BEGIN
  SELECT sequence.id
  INTO active_sequence_id
  FROM outreach_email_sequences sequence
  WHERE sequence.is_active = true
    AND sequence.status = 'approved'
    AND sequence.approved_at IS NOT NULL
  LIMIT 1;

  IF active_sequence_id IS NULL THEN
    RETURN NULL;
  END IF;

  SELECT min(step.position)
  INTO first_step
  FROM outreach_email_sequence_steps step
  WHERE step.sequence_id = active_sequence_id
    AND step.enabled = true;

  IF first_step IS NULL OR NOT EXISTS (
    SELECT 1
    FROM restaurants restaurant
    WHERE restaurant.id = target_restaurant_id
      AND length(trim(restaurant.name)) > 0
      AND lower(trim(restaurant.email)) ~ '^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$'
      AND restaurant.status IN ('lead', 'emailed')
      AND restaurant.shown_interest = false
      AND restaurant.outreach_consent_basis = 'inferred_business'
      AND restaurant.outreach_consent_recorded_at IS NOT NULL
      AND length(trim(restaurant.outreach_consent_source)) > 0
      AND jsonb_typeof(restaurant.outreach_consent_evidence) = 'object'
      AND restaurant.outreach_consent_evidence <> '{}'::jsonb
  ) THEN
    RETURN NULL;
  END IF;

  INSERT INTO email_campaigns (
    restaurant_id, demo_site_id, campaign_type, status, current_step,
    subject, body_html, body_text, demo_token, approved_at,
    sequence_id, next_step, next_send_at
  ) VALUES (
    target_restaurant_id, NULL, 'outreach', 'approved', 0,
    '', '', '', '', now(), active_sequence_id, first_step, now()
  )
  ON CONFLICT (restaurant_id)
    WHERE campaign_type = 'outreach' AND sequence_id IS NOT NULL
  DO NOTHING
  RETURNING id INTO enrollment_id;

  IF enrollment_id IS NULL THEN
    SELECT campaign.id
    INTO enrollment_id
    FROM email_campaigns campaign
    WHERE campaign.restaurant_id = target_restaurant_id
      AND campaign.campaign_type = 'outreach'
      AND campaign.sequence_id IS NOT NULL
    LIMIT 1;
  END IF;

  RETURN enrollment_id;
END;
$$;

DROP TRIGGER IF EXISTS restaurants_outreach_sequence_enrollment ON restaurants;
CREATE TRIGGER restaurants_outreach_sequence_enrollment
AFTER INSERT OR UPDATE OF name, email, status, shown_interest,
  outreach_consent_basis, outreach_consent_source, outreach_consent_recorded_at,
  outreach_consent_evidence
ON restaurants
FOR EACH ROW
EXECUTE FUNCTION enroll_restaurant_in_outreach_sequence();

SELECT ensure_outreach_sequence_enrollment(id)
FROM restaurants;
