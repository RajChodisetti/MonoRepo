# ADR: Outreach outbound inbox and plus-address reply capture

Date: 2026-08-13
Status: Accepted (local/unreleased)

## Context

Restaurant outreach sends reviewed Gmail messages, but the admin UI only showed
send counts and timestamps. `email_campaigns.subject` / `body_text` are overwritten
per send, so there was no durable copy of what a restaurant actually received.
Owner replies were invisible: sending mailboxes use `gmail.send` only, and the
privacy pages promised that tuvi does not read those inboxes.

## Decision

- Persist a rendered outbound snapshot in `email_messages` when Gmail accepts a
  quota-managed sequence send. Store subject, body, from/to, Reply-To,
  Gmail message/thread ids, and RFC `Message-ID`.
- Set `Reply-To` to `outreach+{reply_token}@{domain}` after the delivery attempt
  is claimed so a reply can be matched without reading sales mailboxes.
- Capture inbound mail from one account already present in
  `OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON`, selected by
  `OUTREACH_INBOUND_ACCOUNT_KEY` (or the first account by default). Its one OAuth
  refresh token carries both `gmail.send` and `gmail.readonly`. The worker polls
  Gmail history (fallback: recent inbox list) and inserts inbound `email_messages` rows.
- Match replies in order: plus-address token, `In-Reply-To` / `References`, Gmail
  `threadId`, then sender vs restaurant email / `last_email_recipient`. Unmatched
  mail is stored and listed, not dropped.
- The first matched inbound reply stops that campaign with
  `stopped_reason = inbound_reply`. Do not auto-set `shown_interest`.
- Admin Outreach has an Inbox tab; restaurant detail has a Messages tab.
- An internal administrator can compose a confirmed plain-text reply to the
  latest inbound message. Gmail receives the original thread id,
  `In-Reply-To`, and `References`; the reply uses the capturing mailbox and is
  stored as another outbound snapshot. It does not restart the paused campaign.

## Options Considered

- Read every sales mailbox with `gmail.readonly`: rejected. Only the selected
  single-point mailbox receives readonly scope; other sending accounts remain send-only.
- GCP Pub/Sub `users.watch`: deferred; a worker poller is enough for v1.
- A second inbound-only credential and mailbox: rejected because it creates two
  configuration entry points and cannot send a same-thread admin reply.

## Consequences

- Migration `000048` is required; the source branch's `000039` was renumbered
  during reconciliation because the deployed lineage already uses `000039`.
- Production sets `OUTREACH_INBOUND_ENABLED=true` and selects a configured
  account with `OUTREACH_INBOUND_ACCOUNT_KEY`. That account refresh token must
  include send and readonly scopes; other configured accounts remain send-only.
- Historical sends before this change have no body snapshot unless they are
  re-sent.
