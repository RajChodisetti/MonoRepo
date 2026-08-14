# Lead scrape and outreach operations

This runbook covers the production path from Australian restaurant discovery
to the approved plain-text email sequence. Image OCR is retired and is not a
lead-eligibility or outreach dependency.

## Runtime topology

- `scrape-worker` claims durable city jobs and uses Google Places first.
- Apollo may add missing owner and work-email details, but missing credentials,
  no-match, and provider failures do not mark the job failed or stop verified
  Places data from importing. The shared request ceiling still pauses all
  provider work when exhausted.
- `import_to_db.py` upserts the restaurant/profile and records the lead as
  `inferred_business` with source evidence.
- `ensure_outreach_sequence_enrollment(uuid)` enrolls any restaurant with a
  non-empty name, valid email, eligible lifecycle, and recorded inferred-business
  consent in the active approved sequence. It does not consult legacy suppressions.
- The Go worker sends only when the persisted outreach job is enabled.
- Gmail mailbox quotas, idempotency, and delivery-attempt records remain
  authoritative.

There is no OCR container, one-shot OCR job, host cron, or OCR provider key.

## Eligibility and lifecycle

Eligible lifecycle states are `lead` and `emailed`. A restaurant is excluded
or paused when it has expressed interest or is in `lost`, `archived`,
`client_onboarding`, or `active_client`. The application does not consult the
legacy email-suppression table.

Outreach does not require a generated profile, published demo, approved
restaurant-specific campaign, or media review. Media remains a separate safety
gate: Google media is resolved live with attribution, and owner/licensed files
need explicit admin approval before public use.

## Sequence behavior

- The active approved version may contain any positive number of steps.
- Seed data contains three approved Tuvi Solutions messages.
- Each step is plain text and is rendered from its saved PostgreSQL template.
- `{{greeting01}}` is optional, deterministic, and allowed exactly once in the
  first enabled email body. It is forbidden in subjects and later emails;
  `{{greeting}}` remains supported for follow-ups and legacy active sequences.
- A selected restaurant greeting uses city, the first safe cuisine ending in
  `Restaurant`, rating, and review count only when its profile has a Google
  place id and `scrape_status = 'success'`. Missing or unsafe optional facts
  always select a generic fallback and never block delivery.
- Preview and test-send accept an optional `restaurant_id`. The server then
  ignores synthetic name/owner fields, renders authoritative facts, and returns
  only `greeting01` plus non-sensitive fact-category names for review.
- Any unsubscribe copy or URL must be authored in that saved template. The
  application does not append, require, validate, or render a specialized
  unsubscribe merge tag.
- Delay is measured from the previous confirmed delivery; seed delays are 0,
  3, and 3 days.
- Due follow-ups are ordered before new recipients, but neither due nor future
  follow-ups remove new recipients from the eligible set.
- A normalized email used by more than three restaurant records is excluded at
  selection and rechecked immediately before delivery. The Restaurants admin
  page lists every shared-email group and its restaurant records.
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

Migration `000047_deterministic_restaurant_greeting` creates an inactive draft
cloned from the active sequence. Review and explicitly approve that draft in a
separate administrator action; applying the migration alone does not change the
active version or enable sending. Its down migration refuses to remove a draft
that has been edited or activated.

Migration `000050_reconcile_outreach_enrollment` replaces the stale
suppression-gated enrollment function, backfills missing eligible enrollments,
and leaves the email job disabled. Applying it never activates sending.

## Unified inbox across configured sending mailboxes

Set `OUTREACH_INBOUND_ENABLED=true` and optionally
`OUTREACH_INBOUND_ACCOUNT_KEY`. The key must select an entry already present in
`OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON`; when omitted, the first entry defines
the canonical plus-address Reply-To. Every configured Google Workspace account
is polled independently and every refresh token needs both `gmail.send` and
`gmail.readonly`.

The initial and fallback sync uses `in:inbox newer_than:10d`; the API also
filters on Gmail's provider-received timestamp, so only the last 10 days are
shown. Older stored snapshots are retained. Each mailbox has its own history
cursor and last-attempt/success/error state. The admin Inbox can combine all
mailboxes or filter by stable account key, and unmatched messages remain
replyable without pausing a restaurant campaign.

The admin Inbox reply action sends plain text from the mailbox that captured the
message, preserves the Gmail thread and RFC reply headers, and stores the
accepted outbound snapshot. It does not resume the stopped sequence.

## Resume a failed scrape job

The Scrape jobs admin screen shows a deliberate **Resume** action only for a
failed job. Confirming it calls `POST /api/v1/scrape-jobs/{id}/resume` and
requeues the same job. It preserves completed/subdivided cells, imported and
duplicate candidates, Places detail checkpoints, total request accounting, and
the current request window when it is less than 24 hours old. Only interrupted
or explicitly failed work is made pending again. If another active job exists
for the same city and niche, resume fails closed with HTTP 409.

Do not use Resume as a provider smoke test. It is an explicit production
mutation and can cause the worker to make real Places/Apollo calls once it
claims the queued job.

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
