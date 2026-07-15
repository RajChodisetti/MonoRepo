# ADR: Durable outreach email quota and delivery ledger

Date: 2026-07-14
Status: Accepted

## Context

The outreach account rotation counter previously lived in process memory. A
restart, concurrent worker, or reordered credentials could reset or misapply a
40-message allowance. Provider success and campaign/restaurant updates also
needed one durable reconciliation record. Production templates contained local
link defaults that must never reach real recipients.

## Decision

- Each configured HTTP email account has a stable non-secret key and a row in
  `outreach_email_accounts`.
- A transaction locks the next available account, reserves one of at most 40
  cycle slots, advances to the next account after slot 40, and sets a 24-hour
  `available_at` after that final slot.
- Each account's 40 slots are spread over a persisted 8-hour window: an
  approximately 12-minute base slot plus a cryptographically sampled 2-5 minute
  offset. On-time adjacent slots therefore normally land about 9-15 minutes
  apart. A separate persisted global 2-5 minute minimum-delay gate spans account
  transitions and prevents delayed or restarted workers from sending catch-up
  bursts.
- Every reservation creates an `email_delivery_attempts` record with a global
  `send_sequence`, account cycle, account sequence, and normalized immutable
  recipient. Ambiguous provider outcomes become `unknown` and are never retried
  automatically.
- Each provider-boundary attempt has a five-minute lease. An abandoned lease is
  reconciled to `unknown`/`send_unknown`, while a late worker is fenced from
  recording a confirmed result.
- The quota-claim transaction locks and rechecks the current OCR, profile,
  demo, campaign, prior-send, and suppression gates immediately before it
  consumes an account slot. Unsubscribe writes share an advisory recipient
  lock with that final check.
- Confirmed provider acceptance atomically marks the attempt/campaign/event and
  increments `restaurants.email_send_count`; `email_sent` remains a
  compatibility flag.
- Outreach campaigns cannot bypass the quota through the generic manual
  individual `send-step` endpoint; only the bulk workflow is registered.
- A durable job activation crosses the provider boundary at most once. It then
  releases its lease and requeues for the next PostgreSQL `available_at`; no
  worker sleeps through the pacing interval. One human-triggered job continues
  across account slots and cooldowns until no currently eligible leads remain or
  an operator-reconcilable delivery error stops it.
- Production/staging sends require absolute HTTPS links from the approved Tuvi
  hosts and fail before the provider call if any placeholder or local link
  remains.
- Providers remain HTTP based: Google Workspace Gmail OAuth/API or Zoho Mail
  OAuth/API for the quota-managed account pool, and Zoho or Resend HTTPS for
  generic sends. SMTP is explicitly rejected.
- Normalized Zoho `(region, account_id)` is the provider identity; duplicate
  aliases are rejected so one credential cannot receive multiple 40-send
  quotas.
- A Google Workspace mailbox uses `gmail|<primary mailbox email>` as its provider
  identity. Per-mailbox OAuth refresh tokens use the least-privilege
  `gmail.send` scope; optional send-as aliases do not create a second quota.
- Tracking tokens store the immutable normalized recipient used for the send.
  Unsubscribe tokens do not expire with the demo and suppress that stored
  address, never the restaurant row's later mutable email.

## Options Considered

- Keep per-process counters: rejected because restarts and concurrency break the
  allowance.
- Decrement quota after provider success: rejected because timeouts can be
  ambiguous and could cause a duplicate external send.
- Automatically retry provider timeouts: rejected because at-most-once outreach
  is safer than duplicate contact.
- Sleep for minutes inside the worker: rejected because a restart loses pacing
  state and a sleeping job needlessly holds worker capacity.

## Consequences

- Account configuration changes do not erase accumulated usage because rows are
  keyed independently of array position.
- Redirected, skipped, failed-ambiguous, and sent attempts conservatively consume
  a reserved slot.
- Forty is an upper bound on reserved attempts, not a guarantee of 40
  provider-accepted or delivered emails. Insufficient eligible leads and
  provider outcomes can yield fewer confirmed sends.
- The schedule and global guard survive restarts and concurrent workers. A late
  worker resumes from the current time guard instead of replaying expired slots
  rapidly.
- A provider/delivery error stops the current bulk job instead of draining the
  remaining account allowance through a broken credential or endpoint.
- `send_unknown` requires operator/provider reconciliation before a campaign can
  be retried.
- Previously created drafts are not silently rewritten. The administrator must
  use the regeneration endpoint, review the new content, republish the rotated
  token-gated demo, and re-approve the campaign if links are stale.

## Rollback / Revisit Trigger

Disable email sending before rollback. Keep the ledger for reconciliation even
if a provider adapter is changed. Revisit the policy if a provider supplies a
reliable idempotency key and delivery-status API.
