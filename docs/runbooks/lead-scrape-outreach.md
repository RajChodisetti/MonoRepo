# Lead scrape and outreach operations

This runbook covers the production path from Australian restaurant discovery
to the approved plain-text email sequence. Image OCR is retired and is not a
lead-eligibility or outreach dependency.

## Runtime topology

- `scrape-worker` claims durable city jobs and uses Google Places first.
- Apollo may add missing owner and work-email details.
- `import_to_db.py` upserts the restaurant/profile and records the lead as
  `inferred_business` with source evidence.
- `ensure_outreach_sequence_enrollment(uuid)` enrolls any restaurant with a
  non-empty name and valid email in the active approved sequence.
- The Go worker sends only when the persisted outreach job is enabled.
- Gmail mailbox quotas, suppression, idempotency, and delivery-attempt records
  remain authoritative.

There is no OCR container, one-shot OCR job, host cron, or OCR provider key.

## Eligibility and lifecycle

Eligible lifecycle states are `lead` and `emailed`. A restaurant is excluded
or paused when it is suppressed, has expressed interest, or is in `lost`,
`archived`, `client_onboarding`, or `active_client`.

Outreach does not require a generated profile, published demo, approved
restaurant-specific campaign, or media review. Media remains a separate safety
gate: Google media is resolved live with attribution, and owner/licensed files
need explicit admin approval before public use.

## Sequence behavior

- The active approved version may contain any positive number of steps.
- Seed data contains three approved Tuvi Solutions messages.
- Each step is plain text and must render exactly two URLs: the direct
  `https://tuvisolutions.com` URL and the personalized unsubscribe URL.
- Delay is measured from the previous confirmed delivery; seed delays are 0,
  3, and 3 days.
- Due follow-ups are claimed before new recipients.
- A failure or unknown provider result never advances the integer step.
- A confirmed send advances the step and records sent/next-due timestamps.
- Future-due work is deferred; it does not disable the email job.

## Safe deployment order

1. Keep the production email job disabled.
2. Back up the database and protected environment files without printing
   secret values.
3. Confirm no retired OCR process/container/cron entry exists.
4. Apply the next sequential migration.
5. Deploy API, worker, admin, website, template, and scrape-worker images from
   the same release.
6. Verify migration version, health endpoints, Compose topology, and that no
   executable OCR references remain.
7. Run preview/fake-provider sequence tests. Do not send a real lead email.
8. Inspect eligible/follow-up counts and sequence rendering in the admin UI.
9. Enabling production outreach is a separate deliberate admin action.

## Operational checks

Use redacted/aggregate queries only:

```sql
SELECT status, count(*) FROM restaurants GROUP BY status ORDER BY status;

SELECT outreach_consent_basis, count(*)
FROM restaurants
GROUP BY outreach_consent_basis
ORDER BY outreach_consent_basis;

SELECT sequence.status, count(*)
FROM email_campaigns AS campaign
JOIN outreach_email_sequences AS sequence ON sequence.id = campaign.sequence_id
GROUP BY sequence.status
ORDER BY sequence.status;

SELECT count(*)
FROM email_campaigns AS campaign
JOIN restaurants AS restaurant ON restaurant.id = campaign.restaurant_id
WHERE campaign.next_send_at <= now()
  AND campaign.status = 'approved'
  AND campaign.next_step IS NOT NULL
  AND restaurant.status IN ('lead', 'emailed')
  AND restaurant.shown_interest = false;
```

Confirm the outreach job remains disabled after deployment unless a human has
explicitly approved enabling it. Do not use a live provider for smoke tests.

## Rollback

Roll back application images first, then apply the matching migration down only
if no sequence sends have advanced under the new schema. Historical OCR columns
remain during the stabilization window so the schema rollback is reversible;
the retired provider credentials are not restored automatically.
