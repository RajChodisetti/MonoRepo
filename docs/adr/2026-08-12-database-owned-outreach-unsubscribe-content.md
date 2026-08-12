# Database-Owned Outreach Unsubscribe Content

Date: 2026-08-12  
Status: Accepted

## Context

Outreach previously implemented unsubscribe behavior in several application
layers: merge-tag validation and rendering, tracking tokens and public routes,
suppression persistence and eligibility checks, quota-claim predicates, and
admin recipient status. That duplicated responsibility outside the approved
saved email template and repeatedly reappeared when older branches were merged.

## Decision

The approved PostgreSQL outreach sequence template is the only application
source for unsubscribe text or URLs. Administrators own that content as part of
the reviewed template.

Application code must not:

- append or require unsubscribe copy;
- define or render an unsubscribe-specific merge tag;
- create unsubscribe tracking tokens or expose unsubscribe HTTP routes;
- read or write the legacy `email_suppressions` table; or
- use suppression state in eligibility, delivery claims, API responses, or the
  admin UI.

Applied migrations, historical event values, and historical suppression rows
remain in place for audit and rollback compatibility, but have no runtime role.
General plain-text validation, lifecycle/consent gates, delivery idempotency,
quota controls, and human approval remain unchanged.

## Options Considered

- Keep the application-managed confirmation route and suppression checks. This
  was rejected because it duplicates the database-template responsibility.
- Append a fixed footer in the sender. This was rejected because it makes the
  delivered message differ from the approved saved template.
- Keep only a specialized unsubscribe merge tag. This was rejected because it
  still requires application-specific validation, rendering, and token logic.

## Consequences

- Preview, test-send, and live delivery use the same saved sequence body.
- Template reviewers must include any required unsubscribe content directly in
  the database-backed template and verify the rendered test emails.
- The retired `/t/unsubscribe/{token}` route is expected to return 404.
- Reintroducing application unsubscribe or suppression behavior requires a new
  explicit human decision and an ADR that supersedes this one.

## Rollback / Revisit Trigger

Revisit only if a legal/compliance requirement, provider contract, or verified
delivery limitation requires application-managed unsubscribe behavior. Any
rollback must define the source of truth, restore regression coverage, and be
reviewed before production rollout.
