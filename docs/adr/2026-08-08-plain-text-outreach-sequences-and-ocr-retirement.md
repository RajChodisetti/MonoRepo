# ADR: Plain-text outreach sequences and OCR retirement

Date: 2026-08-08
Status: Accepted
Supersedes: `2026-07-14-ocr-review-gated-lead-preparation.md` for outreach
eligibility and `2026-07-19-hybrid-restaurant-media.md` for automated media
classification

## Context

Outreach was coupled to OCR verification, reviewed profiles, published demos,
and one approved HTML campaign. That made a valid imported business contact wait
on unrelated image work, limited outreach to one fixed draft, and introduced
tracking links and markup that are not needed for the approved sales sequence.
OCR also required a separate runtime and provider credentials even though it is
no longer part of lead qualification.

Tuvi Solutions needs a short, editable plain-text sequence. Follow-ups must be
durable, occur three days apart by default, and be exhausted before new
restaurants enter the sending queue. A restaurant expressing interest must leave
automation for human follow-up.

## Decision

- Retire the OCR worker, schedule, image-classification code, configuration, and
  provider-specific OCR credentials. Historical columns and migrations remain
  temporarily so rollback and audit history are not destroyed.
- Treat an imported restaurant as an `inferred_business` outreach lead when it
  has a non-empty name, a valid business email, and recorded source evidence.
  OCR state, profile approval, demo publication, and legacy campaign readiness
  are not eligibility gates.
- Automated selection accepts lifecycle states `lead` and `emailed`. Expressed
  interest pauses automation. Lost, archived, client-onboarding, and active-client
  restaurants are excluded, and a suppression always wins.
- Store versioned outreach sequences and an arbitrary ordered list of enabled
  steps. A draft can be edited, added to, removed from, reordered, previewed, and
  explicitly approved. Existing recipients finish the approved version in which
  they were enrolled; only new recipients use a newly approved active version.
- Seed the three approved Tuvi Solutions emails. Step one is immediately due;
  steps two and three default to 72 hours after the preceding confirmed send.
- Deliver sequence mail as `text/plain` only. Every enabled template must contain
  exactly one `https://tuvisolutions.com` placeholder and one personalized
  unsubscribe placeholder, with no other raw URL or HTML. Sequence sends do not
  add click redirects, demo links, or open pixels.
- Address the owner by first name when available. Otherwise use
  `Hi {restaurant name} team,`.
- Store the current/next step as integers together with last-confirmed and
  next-due timestamps. Only a confirmed provider acceptance advances progress.
  Failed and unknown results do not advance; unknown delivery remains fail-closed.
- Order due recipients with follow-ups before never-contacted restaurants. The
  durable mailbox quota and pacing ledger remains the final provider gate.
- Use a scanner-safe unsubscribe flow: `GET` shows a confirmation page and does
  not mutate state; `POST` records permanent suppression and stops the sequence.
  Both confirmation states provide only the action and a direct Tuvi website link.
- Keep Google Places images live, attributed, and non-persistent. Do not persist
  scraped third-party gallery or menu images. Owner-provided or separately
  licensed assets may be stored, but require an explicit admin approval or
  rejection before public use; no OCR classification is required.
- Make the production restaurant renderer API-only. API failures return an
  unavailable state instead of falling back to bundled scrape/OCR fixtures, and
  public adapters ignore historical thumbnail/gallery/menu-image fields.

## Options Considered

- Keep OCR as a readiness gate: rejected because image analysis does not
  establish whether a public business email is usable for the approved outreach
  policy.
- Store one mutable campaign per restaurant: rejected because it cannot safely
  support more steps, version approval, or reproducible recipient progress.
- Advance on provider attempt or redirect: rejected because it could skip an
  email that was not confirmed delivered.
- Auto-unsubscribe on link navigation: rejected because security scanners and
  previews can issue `GET` requests without the recipient's intent.
- Persist third-party Google media: rejected because the live provider path is
  the compliant source and already supplies attribution.

## Consequences

- Migration `000042` creates sequence, consent-evidence, and recipient-progress
  contracts; migration `000043` adds the manual media-review audit seam.
- Ingestion calls an idempotent enrollment function after each restaurant upsert.
- Admins can inspect and approve sequence versions and see each recipient's exact
  progress before enabling the email job.
- OCR deployment artifacts and keys can be removed without losing historical
  database data. A later cleanup migration may remove obsolete OCR columns after
  the rollback window closes.
- Historical scraped-media fields remain for rollback/audit but are fail-closed
  at public read boundaries and are not copied into the production template image.
- Production deployment does not authorize sending. The database email-job switch
  remains disabled until a separate operator action explicitly enables it.

## Rollback / Revisit Trigger

Disable the email job before rollback. Restore the prior immutable release and
apply migrations `000043` and `000042` down only after confirming no sequence
delivery is in flight. Revisit the inferred-business policy if legal guidance,
provider requirements, or recipient complaint rates require a stricter gate.
