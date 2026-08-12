-- Make active outreach sequence copy read like a human conversation while
-- preserving the required one-click opt-out path.

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
    last_error = 'Cancelled by migration 45; verify conversational outreach copy before re-enabling the sender.',
    updated_at = now()
WHERE job_type = 'outreach.bulk_send'
  AND status IN ('queued', 'running');

UPDATE outreach_email_sequences
SET name = 'Tuvi conversational restaurant introduction',
    version = version + 1,
    updated_at = now()
WHERE id = '00000000-0000-4000-8000-000000000042';

UPDATE outreach_email_sequence_steps
SET subject_template = 'Quick question about {{restaurant_name}}',
    body_text_template = $email_1${{greeting}}

I was looking at {{restaurant_name}} and noticed a couple of places where the online experience could probably make it easier for guests to get answers and request a table.

I put together a quick Tuvi preview so you can see what I mean:
{{website_url}}

Is it okay if I send over a few notes on what stood out?

No more emails: {{unsubscribe_url}}$email_1$,
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 1;

UPDATE outreach_email_sequence_steps
SET subject_template = 'One thought for {{restaurant_name}}',
    body_text_template = $email_2${{greeting}}

Quick follow-up. The main thing Tuvi helps with is simple: guests should be able to understand the restaurant, request a booking, and get basic questions answered without needing someone on the team to stop what they are doing.

For {{restaurant_name}}, I think the useful pieces would be a cleaner website flow and an AI receptionist for busy or missed calls.

Worth a look?
{{website_url}}

No more emails: {{unsubscribe_url}}$email_2$,
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 2;

UPDATE outreach_email_sequence_steps
SET subject_template = 'Should I close the loop?',
    body_text_template = $email_3${{greeting}}

I do not want to keep chasing you if this is not useful.

If improving the website, reservations, or missed-call handling is something you are considering, I can share the Tuvi preview and the notes behind it.

If not, no worries.

Tuvi overview:
{{website_url}}

No more emails: {{unsubscribe_url}}$email_3$,
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 3;
