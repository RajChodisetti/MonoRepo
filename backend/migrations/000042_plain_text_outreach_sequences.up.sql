-- Versioned plain-text outreach sequences and consent-backed lead enrollment.
-- Historical OCR/demo columns remain in place for rollback, but outreach no
-- longer depends on them.

-- A deploy must never inherit an enabled sender from the previous workflow.
-- Operators must explicitly enable the new sequence sender after deployment
-- and verification. This happens before any sequence enrollment is created.
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
    last_error = 'Cancelled by migration 42; explicitly enable the plain-text sequence sender after deployment verification.',
    updated_at = now()
WHERE job_type = 'outreach.bulk_send'
  AND status IN ('queued', 'running');

-- Retire queued work before the worker registrations disappear. Running jobs
-- are cancelled and unfenced because their handlers no longer have a valid
-- lifecycle to complete; no provider send is attempted by this migration.
UPDATE job_runs
SET status = 'cancelled',
    locked_at = NULL,
    locked_by = NULL,
    lease_expires_at = NULL,
    last_error = 'Retired by plain-text outreach and OCR shutdown migration 42.',
    updated_at = now()
WHERE job_type IN ('lead.prepare', 'email.send')
  AND status IN ('queued', 'running');

CREATE TABLE outreach_email_sequences (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  version integer NOT NULL,
  status text NOT NULL DEFAULT 'draft',
  is_active boolean NOT NULL DEFAULT false,
  approved_at timestamptz,
  approved_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT outreach_email_sequences_name_check CHECK (length(trim(name)) BETWEEN 1 AND 120),
  CONSTRAINT outreach_email_sequences_version_check CHECK (version >= 1),
  CONSTRAINT outreach_email_sequences_status_check CHECK (status IN ('draft', 'approved', 'archived')),
  CONSTRAINT outreach_email_sequences_approval_check CHECK (
    (status = 'draft' AND approved_at IS NULL AND is_active = false)
    OR (status = 'approved' AND approved_at IS NOT NULL)
    OR (status = 'archived' AND approved_at IS NOT NULL AND is_active = false)
  ),
  CONSTRAINT outreach_email_sequences_name_version_unique UNIQUE (name, version)
);

CREATE UNIQUE INDEX outreach_email_sequences_one_active
  ON outreach_email_sequences ((1))
  WHERE is_active = true;

CREATE TABLE outreach_email_sequence_steps (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  sequence_id uuid NOT NULL REFERENCES outreach_email_sequences(id) ON DELETE CASCADE,
  position integer NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  delay_hours integer NOT NULL DEFAULT 72,
  subject_template text NOT NULL,
  body_text_template text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT outreach_email_sequence_steps_position_check CHECK (position >= 1),
  CONSTRAINT outreach_email_sequence_steps_delay_check CHECK (delay_hours BETWEEN 0 AND 8760),
  CONSTRAINT outreach_email_sequence_steps_subject_check CHECK (length(trim(subject_template)) BETWEEN 1 AND 200),
  CONSTRAINT outreach_email_sequence_steps_body_check CHECK (length(trim(body_text_template)) BETWEEN 1 AND 10000),
  CONSTRAINT outreach_email_sequence_steps_plain_text_check CHECK (
    body_text_template !~* '<[a-z][^>]*>'
    AND subject_template !~* 'https?://'
  ),
  CONSTRAINT outreach_email_sequence_steps_sequence_position_unique UNIQUE (sequence_id, position)
);

ALTER TABLE restaurants
  ADD COLUMN outreach_consent_basis text NOT NULL DEFAULT 'inferred_business',
  ADD COLUMN outreach_consent_source text NOT NULL DEFAULT 'business_contact_import',
  ADD COLUMN outreach_consent_recorded_at timestamptz NOT NULL DEFAULT now(),
  ADD COLUMN outreach_consent_evidence jsonb NOT NULL DEFAULT
    '{"policy":"inferred_business","source":"business_contact_import"}'::jsonb;

UPDATE restaurants
SET outreach_consent_recorded_at = COALESCE(outreach_consent_recorded_at, created_at),
    outreach_consent_evidence = outreach_consent_evidence || jsonb_build_object(
      'source', 'existing_business_lead_backfill',
      'recorded_by_migration', 42
    );

-- OCR-derived demo_ready is retired as a lead lifecycle. Preserve provenance
-- in the consent evidence while returning those unsent prospects to `lead`.
UPDATE restaurants
SET status = 'lead',
    outreach_consent_evidence = outreach_consent_evidence || jsonb_build_object(
      'previous_lifecycle', 'demo_ready',
      'lifecycle_normalized_by_migration', 42
    ),
    updated_at = now()
WHERE status = 'demo_ready';

ALTER TABLE restaurants
  ADD CONSTRAINT restaurants_outreach_consent_basis_check CHECK (
    outreach_consent_basis IN ('inferred_business', 'express_interest', 'withdrawn')
  ),
  ADD CONSTRAINT restaurants_outreach_consent_source_check CHECK (length(trim(outreach_consent_source)) > 0);

ALTER TABLE email_campaigns
  ALTER COLUMN demo_site_id DROP NOT NULL,
  ADD COLUMN sequence_id uuid REFERENCES outreach_email_sequences(id) ON DELETE RESTRICT,
  ADD COLUMN next_step integer,
  ADD COLUMN next_send_at timestamptz,
  ADD COLUMN completed_at timestamptz;

ALTER TABLE email_campaigns
  ADD CONSTRAINT email_campaigns_sequence_progress_check CHECK (
    (sequence_id IS NULL)
    OR (
      current_step >= 0
      AND (next_step IS NULL OR next_step > current_step)
      AND ((next_step IS NULL AND next_send_at IS NULL) OR (next_step IS NOT NULL AND next_send_at IS NOT NULL))
    )
  );

CREATE UNIQUE INDEX email_campaigns_one_sequence_enrollment
  ON email_campaigns (restaurant_id)
  WHERE campaign_type = 'outreach' AND sequence_id IS NOT NULL;

CREATE INDEX email_campaigns_due_sequence
  ON email_campaigns (next_send_at, current_step DESC, created_at)
  WHERE campaign_type = 'outreach'
    AND sequence_id IS NOT NULL
    AND status = 'approved'
    AND next_step IS NOT NULL;

INSERT INTO outreach_email_sequences (
  id, name, version, status, is_active, approved_at
) VALUES (
  '00000000-0000-4000-8000-000000000042',
  'Tuvi restaurant introduction',
  1,
  'approved',
  true,
  now()
);

INSERT INTO outreach_email_sequence_steps (
  sequence_id, position, enabled, delay_hours, subject_template, body_text_template
) VALUES
(
  '00000000-0000-4000-8000-000000000042',
  1,
  true,
  0,
  'A practical idea for {{restaurant_name}}',
  $email_1${{greeting}}

Tuvi Solutions helps restaurants make it easier for guests to find what they need, request a reservation, and reach the team when staff are busy.

The goal is simple: fewer missed opportunities and a smoother guest experience without adding more work for your team.

See how we help restaurants:
{{website_url}}

Best,
The Tuvi Solutions team

Business outreach from Tuvi Solutions
Opt out: {{unsubscribe_url}}$email_1$
),
(
  '00000000-0000-4000-8000-000000000042',
  2,
  true,
  72,
  'What Tuvi can help {{restaurant_name}} with',
  $email_2${{greeting}}

Tuvi Solutions provides practical tools that help restaurants improve guest service and capture more business:

- Modern mobile-friendly websites
- An AI review of your restaurant's digital footprint
- QR ordering and rewards
- Online reservation-request capture
- An AI voice receptionist for missed and busy calls
- Reviewed marketing campaigns

Together, these tools can reduce missed enquiries, make bookings easier, and give your team more time to focus on guests.

Explore Tuvi Solutions:
{{website_url}}

Best,
The Tuvi Solutions team

Business outreach from Tuvi Solutions
Opt out: {{unsubscribe_url}}$email_2$
),
(
  '00000000-0000-4000-8000-000000000042',
  3,
  true,
  72,
  'Worth a quick conversation about {{restaurant_name}}?',
  $email_3${{greeting}}

I wanted to follow up in case improving {{restaurant_name}}'s guest experience or reducing missed enquiries is on your list.

You can schedule a free consultation through our website. We will look at where Tuvi could help and explain the practical benefits for your restaurant.

There is no pressure and no obligation.

Schedule a conversation:
{{website_url}}

Best,
The Tuvi Solutions team

Business outreach from Tuvi Solutions
Opt out: {{unsubscribe_url}}$email_3$
);

-- This is the stable ingestion seam. Importers can call it after an upsert;
-- it is idempotent and never sends or approves an unapproved sequence.
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
      AND NOT EXISTS (
        SELECT 1
        FROM email_suppressions suppression
        WHERE suppression.email = lower(trim(restaurant.email))
      )
  ) THEN
    RETURN NULL;
  END IF;

  INSERT INTO email_campaigns (
    restaurant_id,
    demo_site_id,
    campaign_type,
    status,
    current_step,
    subject,
    body_html,
    body_text,
    demo_token,
    approved_at,
    sequence_id,
    next_step,
    next_send_at
  ) VALUES (
    target_restaurant_id,
    NULL,
    'outreach',
    'approved',
    0,
    '',
    '',
    '',
    '',
    now(),
    active_sequence_id,
    first_step,
    now()
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

CREATE OR REPLACE FUNCTION enroll_restaurant_in_outreach_sequence()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  PERFORM ensure_outreach_sequence_enrollment(NEW.id);
  RETURN NEW;
END;
$$;

CREATE TRIGGER restaurants_outreach_sequence_enrollment
AFTER INSERT OR UPDATE OF name, email, status, shown_interest,
  outreach_consent_basis, outreach_consent_source, outreach_consent_recorded_at
ON restaurants
FOR EACH ROW
EXECUTE FUNCTION enroll_restaurant_in_outreach_sequence();

SELECT ensure_outreach_sequence_enrollment(id)
FROM restaurants;
