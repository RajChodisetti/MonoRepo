# ADR: Sydney scheduled outreach window

Date: 2026-08-15
Status: Accepted (deployed; inbox freshness/detail hardening local/unreleased)

## Context

Durable restaurant outreach used rolling per-mailbox windows. With multiple
mailboxes, the scheduler preferred finishing one mailbox before starting the
next, so the aggregate allowance could extend beyond the desired operating
hours. Direct admin messages and health checks shared providers but were never
intended to be constrained by the restaurant-outreach schedule.

The admin inbox also grouped inbound and outbound snapshots into one thread but
only exposed its latest message ID. After an administrator replied, that ID
pointed to the new outbound snapshot and the Reply action disappeared even
though the thread still had an authoritative inbound message and receiving
mailbox.

## Decision

- Persist one internal-admin-managed, start-inclusive/end-exclusive scheduled
  outreach window in PostgreSQL. Migration 54 seeds 07:00 through 12:00 in
  `Australia/Sydney`; the Outreach UI can change the start and end while the
  email job is disabled and no bulk run is active. Use the IANA timezone so
  daylight-saving transitions are applied automatically.
- Treat the saved period as a local send day. Reset per-mailbox usage at the
  next Sydney window. Advance the gradual 5-to-40 ramp only when the prior
  window's allowance was fully consumed.
- Pace all enabled sending mailboxes concurrently by their persisted
  `available_at` values. Cap the cross-account random delay by aggregate final
  quota capacity, while retaining the configured minimum delay, so the daily
  allowance can finish before the saved end time when enough eligible
  recipients exist. Reject UI updates and runtime account configurations whose
  full daily quota cannot fit at the configured minimum pacing interval.
- Defer queued work to the next saved start boundary outside the window. Quota
  claims lock and read the persisted window transactionally, so saved changes
  take effect without a process restart.
- Keep template tests, inbox replies, health checks, and other direct/manual
  email paths outside the scheduled window and daily quota, including quota
  reconciliation, synchronization, claims, and pacing. Existing global
  send-disable and admin confirmation controls still apply; provider limits do
  not change.
- Return the latest inbound message ID for each inbox thread separately from the
  latest overall message ID. The UI always replies to that inbound snapshot and
  the backend selects the Gmail provider using its captured `mailbox_key`, so a
  reply is sent from the address that received the email.
- Build each inbox row from its latest received message: subject, bounded text
  preview, sender, receiving address, received timestamp, and inbound message ID
  all come from the same deterministic snapshot. Order rows by that received
  timestamp with a stable message-ID tie-breaker; unread count is a filter/badge
  and neither unread state nor a later admin reply moves older received mail
  above newer mail.
- Keep complete message bodies out of the 15-second paginated response. Opening
  a subject, preview, or Open action marks only that received message read and
  returns its complete stored text for an escaped, pre-wrapped detail dialog.
  Refresh the mounted inbox every 15 seconds, retain the current page, filters,
  and open detail for in-app manual/automatic refreshes, and default inbound
  Gmail polling to the same interval.

## Options Considered

- Keep rolling windows and only change their duration to five hours: rejected
  because serial mailbox use can still carry aggregate quota past noon and a
  late worker start can shift a window into the afternoon.
- Apply the time guard inside the Gmail provider: rejected because it would also
  block explicitly exempt test, reply, health, and direct messages.
- Store the window only in protected environment variables: rejected because an
  administrator could not safely change it from the UI and workers would need a
  restart to observe it.
- Reply using the current sender rotation: rejected because it can move an
  existing conversation to a different mailbox and identity.

## Consequences

- Migration 54 adds the singleton `outreach_send_schedule` table, defaults it to
  07:00-12:00 Sydney time, and fails the outreach job closed during rollout.
- `OUTREACH_SEND_WINDOW` remains a legacy per-mailbox pacing field for schema
  compatibility; it no longer defines the scheduled send boundary. No protected
  environment mutation or worker restart is required for an admin window edit.
- The quota target assumes sufficient eligible recipients and successful
  provider responses. A shortage, explicit disable, or delivery failure can
  leave that day's allowance unused; the scheduler never sends outside the
  window to compensate.
- Inbox replies remain available after an earlier admin response and continue
  to preserve Gmail thread and RFC reply headers.
- Inbox detail shows the complete stored text body, not attachments or a
  faithful rendered MIME message. HTML fallback content is displayed literally
  rather than executed as markup.
- Inbound Gmail polling runs four times as often as the previous 60-second
  default. Existing deployments with an explicit interval retain that override
  until their protected configuration is changed and the worker is recreated.
- Production deployment, configuration mutation, and enabling real outreach
  remain explicit approval gates. This decision does not authorize any send.

## Rollback / Revisit Trigger

Rollback uses the prior binary and migration 54 down after restoring the saved
window to its 07:00-12:00 default. The down migration refuses to discard a
customized window and leaves outreach disabled. Revisit the timezone, minimum
pacing interval, or quota ramp only with sender-reputation evidence and an
approved operating plan.
