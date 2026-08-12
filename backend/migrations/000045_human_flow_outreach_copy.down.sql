-- Restore the original migration-42 active outreach sequence copy.

UPDATE outreach_email_sequences
SET name = 'Tuvi restaurant introduction',
    version = GREATEST(version - 1, 1),
    updated_at = now()
WHERE id = '00000000-0000-4000-8000-000000000042';

UPDATE outreach_email_sequence_steps
SET subject_template = 'A practical idea for {{restaurant_name}}',
    body_text_template = $email_1${{greeting}}

Tuvi Solutions helps restaurants make it easier for guests to find what they need, request a reservation, and reach the team when staff are busy.

The goal is simple: fewer missed opportunities and a smoother guest experience without adding more work for your team.

See how we help restaurants:
{{website_url}}

Best,
The Tuvi Solutions team

Business outreach from Tuvi Solutions
Opt out: {{unsubscribe_url}}$email_1$,
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 1;

UPDATE outreach_email_sequence_steps
SET subject_template = 'What Tuvi can help {{restaurant_name}} with',
    body_text_template = $email_2${{greeting}}

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
Opt out: {{unsubscribe_url}}$email_2$,
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 2;

UPDATE outreach_email_sequence_steps
SET subject_template = 'Worth a quick conversation about {{restaurant_name}}?',
    body_text_template = $email_3${{greeting}}

I wanted to follow up in case improving {{restaurant_name}}'s guest experience or reducing missed enquiries is on your list.

You can schedule a free consultation through our website. We will look at where Tuvi could help and explain the practical benefits for your restaurant.

There is no pressure and no obligation.

Schedule a conversation:
{{website_url}}

Best,
The Tuvi Solutions team

Business outreach from Tuvi Solutions
Opt out: {{unsubscribe_url}}$email_3$,
    updated_at = now()
WHERE sequence_id = '00000000-0000-4000-8000-000000000042'
  AND position = 3;
