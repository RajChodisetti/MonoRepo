# Deterministic Restaurant Greeting

Date: 2026-08-14
Status: Accepted (local/unreleased)

## Context

The first outreach email needs a short restaurant-specific opening without
introducing generated claims, model/provider dependencies, or a separate
approval workflow for each recipient. Existing live, preview, and template-test
paths must not drift, and legacy approved sequences using `{{greeting}}` must
continue to render unchanged.

## Decision

- Add `{{greeting01}}` as an optional merge field allowed exactly once in the
  first enabled email body. It is forbidden in subjects and later emails.
- Render it through one pure `GreetingFacts` function shared by live delivery,
  saved-sequence preview, and template test send.
- Produce one safe salutation, a blank separator, and exactly two greeting
  lines. Use owner first name when safely cleaned; otherwise address the
  restaurant team.
- Use city, cuisine, rating, and review count only when `google_place_id` is
  present and `scrape_status = 'success'`. Cuisine must be the first safe array
  value ending in `Restaurant`; the suffix is removed for natural wording.
- Mention rating only from 4.0 through 5.0 with at least 10 reviews. Missing,
  malformed, multiline, oversized, placeholder-like, or otherwise unsafe facts
  select deterministic fallback wording and never block delivery.
- Let internal-admin preview and test-send requests optionally identify a
  restaurant. Server-side facts then override synthetic inputs, and responses
  expose only the rendered greeting plus non-sensitive fact-category names.
- Migration `000047` clones the current active sequence into a separate inactive
  draft, opts its first enabled step into `{{greeting01}}`, and removes the
  repetitive opening sentence. It does not approve, activate, or enable sending.

## Options Considered

- Generate greetings with an AI model: rejected because listing facts already
  support useful copy and generated claims add cost, nondeterminism, and review
  risk.
- Persist a generated greeting per restaurant: rejected because the pure
  renderer and existing campaign body snapshot provide the required audit trail
  without another mutable source of truth.
- Replace `{{greeting}}` globally: rejected because approved follow-ups and
  historical sequences must remain compatible.

## Consequences

- Identical facts produce identical unsigned sequence content in live, preview,
  and test-send paths; provider signatures remain a separate shared boundary.
- The delivery query loads a small set of listing fields but never review text,
  menu items, descriptions, or inferred claims.
- Administrators can search the existing restaurant endpoint to verify the
  exact greeting and fact categories before approval or a deliberate test send.
- Applying the migration alone cannot alter active outreach. Activation and the
  persisted email-job switch remain explicit administrator actions.

## Rollback / Revisit Trigger

The down migration deletes only the fixed migration-created draft while it is
still inactive and untouched; it fails closed after edits or activation. Revisit
the fact thresholds or allowed categories only with evidence that the current
fallbacks are misleading or materially reduce useful personalization.
