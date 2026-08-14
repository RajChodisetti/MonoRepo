# ADR: Gradual durable outreach email ramp

Date: 2026-08-14
Status: Accepted (local/unreleased)

## Context

Quota-managed outreach previously allowed each configured mailbox to reserve up
to `OUTREACH_EMAILS_PER_ACCOUNT` attempts, normally 40, during its first send
cycle. Account usage and cooldowns were already durable, but there was no
mailbox warm-up state. Starting immediately at the maximum creates avoidable
sender-reputation risk.

## Decision

- Persist `ramp_day` on each `outreach_email_accounts` row.
- Treat `OUTREACH_EMAILS_PER_ACCOUNT` as the final per-mailbox ceiling, still
  constrained to 1–40.
- Set the effective allowance to the lower of that ceiling and
  `ramp_day * 5`: 5, 10, 15, 20, 25, 30, 35, then 40 with the default ceiling.
- Advance `ramp_day` only after the mailbox fully uses its current allowance and
  completes the existing minimum 24-hour cooldown. Once the configured ceiling
  is reached, retain that ramp day and ceiling for later cycles.
- Distribute the effective allowance across the existing send window and retain
  the global random pacing gate. PostgreSQL remains the authority across worker
  restarts and concurrent claims.
- Start existing accounts conservatively on ramp day one. If an existing cycle
  already consumed five or more slots when migration `000049` is applied, hold
  that account for a fresh 24 hours before allowing the next ramp cycle.

Here, “day” means a fully used quota cycle followed by its cooldown, not a UTC
calendar boundary. A mailbox that does not consume its allowance does not warm
up merely because wall-clock days pass.

## Options Considered

- Set the environment limit to 5 and change it manually each day: rejected
  because restarts and operator timing could reset or skip the intended ramp.
- Derive the limit from account creation date: rejected because idle mailboxes
  could reach 40 without having established any sending history.
- Count all sends by UTC calendar day: deferred because it would replace the
  established rolling cooldown model and add timezone/boundary complexity.

## Consequences

- Migration `000049_outreach_email_ramp` is required before this code can start.
- A new mailbox needs eight fully used/cooldown cycles to reach 40. A lower
  configured ceiling is honored as soon as the next five-send step reaches it.
- Health checks and deliberate template-test sends remain outside durable
  campaign quota, as before; real sequence delivery remains approval-gated.
- The first production migration/deployment and any sender enablement require
  explicit operator approval. This ADR does not authorize an email send.

## Rollback / Revisit Trigger

Rollback requires the prior binary before dropping `ramp_day`; the migration's
conservative `available_at` delay is intentionally not reversed. Revisit the
step size, cooldown, or calendar model only with deliverability evidence and an
approved migration/config plan.
