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
    last_error = 'Cancelled by migration 46; verify minimal outreach links before re-enabling the sender.',
    updated_at = now()
WHERE job_type = 'outreach.bulk_send'
  AND status IN ('queued', 'running');

UPDATE outreach_email_sequence_steps
SET body_text_template = '{{greeting}}

I had a quick thought after looking at {{restaurant_name}}. The online flow could probably make it easier for guests to get answers and request a table without adding more work for the team.

Would it be useful if I sent over a couple of notes on what stood out?

Unsubscribe: {{unsubscribe_url}}',
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 1;

UPDATE outreach_email_sequence_steps
SET body_text_template = '{{greeting}}

Quick follow-up. The main thing Tuvi helps with is simple: guests should be able to understand the restaurant, request a booking, and get basic questions answered without needing someone on the team to stop what they are doing.

For {{restaurant_name}}, I think the useful pieces would be a cleaner website flow and an AI receptionist for busy or missed calls.

Worth a quick note back?

Unsubscribe: {{unsubscribe_url}}',
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 2;

UPDATE outreach_email_sequence_steps
SET body_text_template = '{{greeting}}

I do not want to keep chasing you if this is not useful.

If improving the website, reservations, or missed-call handling is something you are considering, I can send a few notes in a reply.

If not, no worries.

Unsubscribe: {{unsubscribe_url}}',
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 3;

COMMIT;
