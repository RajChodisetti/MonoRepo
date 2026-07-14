# ADR: OCR states and review-gated lead preparation

Date: 2026-07-14
Status: Accepted

## Context

The legacy `ocr_verified` boolean could not distinguish unprocessed work,
active work, a verified result, a lead with no usable images, or a failed OCR
attempt. Google Places photo URLs are also short-lived media references, and a
verified lead still needs reviewable demo and campaign artifacts before any
outreach can be approved.

## Decision

- `restaurant_profiles.ocr_status` is the source of truth with states
  `pending`, `running`, `verified`, `no_images`, and `failed`.
- The old boolean remains a compatibility projection and is true only for
  `verified`.
- Imports fingerprint stable image inputs. A changed fingerprint resets OCR to
  `pending`; unchanged imports preserve the existing outcome.
- The OCR job claims pending/stale work in PostgreSQL. It refreshes current
  Places photo resource names immediately before analysis and persists neither
  those expirable names nor the short-lived `photoUri`.
- Each OCR claim carries a UUID and the exact input fingerprint. Finalization is
  conditional on both values so a reclaimed worker cannot overwrite newer
  scrape input.
- At least one image analysis must succeed before `verified`. `no_images` is a
  non-eligible remediation state. `failed` is non-eligible and retries after a
  configurable cooldown up to a configurable attempt limit.
- Verification enqueues an idempotent `lead.prepare` job. The Go worker creates
  a token-gated demo in `draft` and an outreach campaign in `draft`.
- Automatically generated drafts record independent OCR provenance and a
  deterministic SHA-256 fingerprint of restaurant identity plus the exact
  generated public payload. A change to either fingerprint refreshes only an
  automatic draft; manual or non-draft artifacts remain operator-owned.
- Demo publication, campaign approval, bulk eligibility, and the final quota
  claim recompute and compare both provenances. A stale automatic artifact
  therefore cannot send after a newer OCR/profile/menu/recipient input exists.
- A reviewed restaurant name or recipient change clears downstream gates and
  queues a profile-fingerprint-specific `lead.prepare` job even when the image
  fingerprint did not change.
- An image-fingerprint change clears profile-review and demo-publication audit
  state and returns only an approved (not in-flight/terminal) campaign to draft.
- Real outreach still requires three explicit internal-admin gates: approve the
  profile, publish the generated demo, and approve the campaign. These actions
  are separate from sending and are audited.
- Tracked outreach resolves the published demo through its slug/token payload,
  not the ungated public restaurant index. Token mismatch or expiry blocks
  approval/send. An admin-only regeneration action requires a draft demo,
  rotates token/expiry, rebuilds current email content, clears campaign
  approval, and records the actor before re-review.

## Options Considered

- Keep the boolean plus an error array: rejected because queries and retries
  remain ambiguous.
- Persist Places resource names or resolved media URLs: rejected because Google
  documents resource names as expirable and returns short-lived media URLs.
  Stable Place ID plus observed photo count forms durable input provenance.
- Auto-publish and auto-approve after OCR: rejected because OCR is not a human
  review and external outreach requires an approval gate.

## Consequences

- Migrations `000016` and `000022` must precede ingestion/OCR, and the Go worker that knows
  `lead.prepare` must be deployed before OCR is enabled.
- `no_images` and exhausted `failed` leads are visible for operator remediation
  but never pass bulk eligibility.
- Manual, in-flight, sent, stopped, and ambiguous campaigns are operator-owned
  and are not silently overwritten when a later OCR fingerprint changes.

## Rollback / Revisit Trigger

Disable the OCR schedule before rolling back. Migration `000016` can restore
the compatibility index, but the richer outcomes will be lost. Revisit retry
policy when a human remediation queue is added.
