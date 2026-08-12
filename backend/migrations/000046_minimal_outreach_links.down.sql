BEGIN;

UPDATE outreach_email_sequence_steps
SET body_text_template = '{{greeting}}

I was looking at {{restaurant_name}} and noticed a couple of places where the online experience could probably make it easier for guests to get answers and request a table.

I put together a quick Tuvi preview so you can see what I mean:
{{website_url}}

Is it okay if I send over a few notes on what stood out?

No more emails: {{unsubscribe_url}}',
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 1;

UPDATE outreach_email_sequence_steps
SET body_text_template = '{{greeting}}

Quick follow-up. The main thing Tuvi helps with is simple: guests should be able to understand the restaurant, request a booking, and get basic questions answered without needing someone on the team to stop what they are doing.

For {{restaurant_name}}, I think the useful pieces would be a cleaner website flow and an AI receptionist for busy or missed calls.

Worth a look?
{{website_url}}

No more emails: {{unsubscribe_url}}',
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 2;

UPDATE outreach_email_sequence_steps
SET body_text_template = '{{greeting}}

I do not want to keep chasing you if this is not useful.

If improving the website, reservations, or missed-call handling is something you are considering, I can share the Tuvi preview and the notes behind it.

If not, no worries.

Tuvi overview:
{{website_url}}

No more emails: {{unsubscribe_url}}',
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 3;

COMMIT;
