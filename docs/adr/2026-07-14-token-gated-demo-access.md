# ADR: Token-gated demo access with opaque bearer tokens

Date: 2026-07-14
Status: Accepted

## Context

The implemented demo route has always issued a cryptographically random
per-demo token, stored a bcrypt hash on `demo_sites`, and validated the supplied
token against that hash. Some planning documents called this a “signed token”
and exposed an unused `DEMO_TOKEN_SECRET`, even though no HMAC or other token
signature was produced or verified.

The outreach workflow also needs immediate revocation through unpublishing,
expiry, and token rotation, so it already performs a database lookup for every
token-gated demo request.

## Decision

- Keep a 256-bit random opaque bearer token per demo.
- Store only its bcrypt hash on `demo_sites`.
- Keep the current token on the reviewable campaign artifact only while it is
  needed to construct that campaign's tracked demo URL.
- Require HTTPS in production, apply demo expiry, support administrator-only
  rotation/regeneration, and return token-gated payloads with `Cache-Control:
  no-store`.
- Do not expose or require a global `DEMO_TOKEN_SECRET`; `DEMO_TOKEN_TTL` is the
  operational expiry control.
- Use “opaque” or “token-gated” in current API and operator documentation.

## Options Considered

1. HMAC-sign a stateless token with a global secret. Rejected for the current
   workflow because publication state, review state, revocation, and rotation
   already require database access, while a global secret adds another
   production secret and a broader rotation blast radius.
2. Keep the current opaque token but continue calling it signed. Rejected
   because that misstates the security mechanism and led operators to expect an
   unused secret.

## Consequences

- Demo access remains revocable and token rotation stays per demo.
- A leaked bearer token grants access until the demo is unpublished, rotated,
  or expires; tokens must therefore stay out of logs and non-HTTPS URLs.
- The core `AGENTS.md` wording still uses “signed” as a product shorthand. Its
  non-living sections require a separate human-approved edit if that wording is
  to be changed; this ADR records the implemented mechanism without silently
  editing the operating contract.

## Rollback / Revisit Trigger

Revisit if demos must become stateless across databases, if a separate public
verification service is introduced, or if campaign-held bearer tokens need
encryption at rest. A future signed-token design must preserve per-demo
revocation and rotation semantics.
