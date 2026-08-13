# ADR: Outreach outbound inbox and plus-address reply capture

Date: 2026-08-13
Status: Accepted

## Context

Restaurant outreach sends reviewed Gmail messages, but the admin UI only showed
send counts and timestamps. `email_campaigns.subject` / `body_text` are overwritten
per send, so there was no durable copy of what a restaurant actually received.
Owner replies were invisible: sending mailboxes use `gmail.send` only, and the
privacy pages promised that tuvi does not read those inboxes.

## Decision

- Persist a rendered outbound snapshot in `email_messages` when Gmail accepts a
  send (bulk quota-managed and ad hoc). Store subject, body, from/to, Reply-To,
  Gmail message/thread ids, and RFC `Message-ID`.
- Set `Reply-To` to `outreach+{reply_token}@{domain}` after the delivery attempt
  is claimed so a reply can be matched without reading sales mailboxes.
- Capture inbound mail from one dedicated mailbox (`OUTREACH_INBOUND_MAILBOX_JSON`)
  with `gmail.readonly`. The worker polls Gmail history (fallback: recent inbox
  list) and inserts inbound `email_messages` rows.
- Match replies in order: plus-address token, `In-Reply-To` / `References`, Gmail
  `threadId`, then sender vs restaurant email / `last_email_recipient`. Unmatched
  mail is stored and listed, not dropped.
- The first matched inbound reply stops that campaign with
  `stopped_reason = inbound_reply`. Do not auto-set `shown_interest`.
- Admin Outreach has an Inbox tab; restaurant detail has a Messages tab.

Sending mailboxes remain `gmail.send` only.

## Options Considered

- Read every sales mailbox with `gmail.readonly`: rejected because it expands
  OAuth consent and contradicts the existing Gmail send-only disclosure.
- GCP Pub/Sub `users.watch`: deferred; a worker poller is enough for v1.
- Admin compose/reply from the UI: out of scope.

## Consequences

- Migration `000039` is required.
- Production needs `outreach@` (or the configured plus-address mailbox),
  `OUTREACH_INBOUND_ENABLED=true`, and a readonly refresh token for that mailbox
  only.
- Historical sends before this change have no body snapshot unless they are
  re-sent.
