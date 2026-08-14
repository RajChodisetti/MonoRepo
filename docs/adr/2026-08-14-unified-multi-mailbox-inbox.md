# ADR: Unified multi-mailbox outreach inbox

Date: 2026-08-14
Status: Accepted (local/unreleased)

## Context

The initial outreach inbox read one selected Google Workspace account and used
that account as a single reply entry point. Tuvi now needs one admin inbox that
shows recent mail received by every configured outreach account and lets an
administrator answer through the account that received the message. Unmatched
mail must remain visible; it must not break listing or pause unrelated outreach.

## Decision

- Start one Gmail inbox poller for every entry in
  `OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON`. Each refresh token must authorize
  both `gmail.send` and `gmail.readonly`.
- Retain `OUTREACH_INBOUND_ACCOUNT_KEY` only as the selector for the canonical
  plus-address Reply-To. It no longer limits which configured inboxes are read.
- On initial sync or an expired history cursor, capture the complete result of
  `in:inbox newer_than:10d`, with pagination. Read the profile history cursor
  before listing so mail arriving during the bootstrap cannot be skipped.
- On incremental sync, page through Gmail message-added history and fetch only
  messages that currently carry the `INBOX` label. Enforce the same 10-day
  window using Gmail's `internalDate` and again when listing stored threads.
- Store all qualifying inbox messages, whether or not they match a restaurant.
  Match in the existing order: reply token, RFC reply headers, Gmail thread,
  then restaurant sender address. Only a match that recovers a campaign pauses
  that campaign.
- Scope provider-message idempotency, sync cursors, sync health, and conversation
  grouping by stable mailbox key. Group conversations by mailbox plus Gmail
  thread id, and expose the receiving address and mailbox key to internal admins.
- A manual reply always selects the provider by the captured mailbox key and
  preserves the Gmail thread and RFC reply headers. The persisted bulk-email
  control remains independent; the global email safety shutdown still blocks
  every real send.

## Options Considered

- Keep one dedicated inbound mailbox: rejected because replies sent directly to
  the other configured sender accounts remain invisible.
- Forward every account into one mailbox: rejected because it loses authoritative
  receiving-account identity and can reply from the wrong account.
- Store only matched outreach replies: rejected because the administrator asked
  for every recent inbox message and unmatched correspondence still needs a
  safe reply path.
- Import unlimited historical mail: rejected in favor of an explicit 10-day
  operational window that limits private-data retention and initial API load.

## Consequences

- Migration `000051` records provider-received time, changes Gmail idempotency to
  `(mailbox_key, gmail_message_id)`, and adds per-mailbox poll health. Its down
  migration fails closed if provider ids now overlap across mailboxes.
- Existing refresh tokens issued with only `gmail.send` cannot read mail; those
  accounts show a redacted access error until an administrator completes OAuth
  consent with `gmail.readonly`. Other healthy accounts continue syncing.
- The admin surface contains private mailbox correspondence and remains restricted
  to `internal_admin`; no inbox content is added to public restaurant payloads.
- Adding another properly authorized Google Workspace account remains an
  environment-only change. Zoho is not part of the active outreach account pool
  and would require a separate inbox adapter if reintroduced.

## Rollback / Revisit Trigger

Revisit polling in favor of Gmail `users.watch` when mailbox volume or polling
latency warrants Pub/Sub. Roll back only after confirming there are no duplicate
provider message ids across mailboxes and after preserving any correspondence
needed for audit.
