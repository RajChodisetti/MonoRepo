# ADR: Gmail-only restaurant outreach and daily sender health

Date: 2026-07-18
Status: Accepted

## Context

Restaurant outreach needs multiple independently authorized sender mailboxes,
durable low-volume pacing, and visible proof that every sender can still submit
mail. Adding a mailbox must be configuration-only. SMTP and a single shared API
key do not establish authority to send as multiple Gmail mailboxes.

## Decision

Quota-managed restaurant outreach uses the Gmail API with one offline OAuth
refresh token per mailbox. Accounts are loaded from
`OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON`; each entry has a stable non-secret key,
mailbox/from address, OAuth client ID/secret, and refresh token. Adding an entry
requires no code change.

PostgreSQL keeps the per-account delivery schedule: 40 attempts spread across an
eight-hour window with persisted jitter and a 24-hour cooldown. The worker also
sends one real health-check message from each enabled Gmail account to the
configured internal recipient when first registered and at least 24 hours apart.
The admin UI shows the last/next check, provider acceptance, and safe failure.

OAuth credentials remain secret configuration and are never returned through
the API or stored in health tables. PostgreSQL stores the restaurant outreach
job's operational enabled state, and only an internal admin can change it from
the Outreach UI. Enabling immediately queues the durable eligible-lead run;
disabling prevents the next Gmail provider request from starting. Daily health
checks are controlled separately by `OUTREACH_EMAIL_HEALTH_ENABLED`.

## Options Considered

- Gmail API with per-mailbox OAuth: selected for explicit mailbox authority and
  HTTP API delivery.
- Mixed Gmail and Zoho outreach pool: rejected for restaurant outreach so one
  provider contract and health model remain authoritative.
- SMTP: rejected by repository security policy and weaker delivery/audit control.

## Consequences

- Deployment requires migrations `000024` and `000029`, plus valid Gmail OAuth
  records in the production environment.
- A Google API key alone is insufficient; each mailbox needs offline OAuth
  consent and a refresh token with Gmail send permission.
- Real daily health messages consume a small amount of mailbox sending capacity.
- `EMAIL_DISABLE_SENDING` continues to gate the legacy/generic adapter, not the
  quota-managed Gmail outreach job.

## Rollback / Revisit Trigger

Disable `OUTREACH_EMAIL_HEALTH_ENABLED` and restart the worker to stop health
messages. Disable the outreach job in the admin UI to stop new restaurant sends.
Revisit the provider decision if Gmail delivery limits, account policy, or
measurable deliverability makes another HTTPS provider safer.
