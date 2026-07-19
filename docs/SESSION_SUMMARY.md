# Session Summary

**Production state (2026-07-19):** Release `6c21c15` and migrations through `000031` are live with fresh attributed Google photo resolution, OCR-aware non-menu placement, admin media controls, S3-compatible owned-media support, and the clarified mobile template preview.
**Public safety:** Google URLs, resource names, and bytes are not persisted; only exact OCR fingerprint matches render, menu/unclassified photos fail closed, and restaurant-detail responses are `no-store`.
**OCR / outreach:** Migration reset 21 legacy fingerprint-less profiles; 940 are pending, 4 failed, and 154 email-equipped profiles can resume after the current 200/200 UTC-day budget resets. Outreach remains off and 3/3 Gmail health records remain healthy.
**Verification:** All eight production containers report running with zero restarts; public website, API, admin, voice, and templates 1/2/3 return HTTP 200. Pre-release Go/Python/TypeScript/build/OpenAPI/Compose checks and post-release schema/log/cache/control smokes pass.
**Follow-up:** Production durable owner uploads remain disabled until an S3-compatible bucket/CDN is configured; this does not block the live Google resolver.
