# Session Summary

**Production state (2026-07-19):** Release `caffcfb` and migrations through `000030` remain live; restaurant outreach is off and the email-only OCR worker retains its global 200-request UTC-day ceiling.
**Local unreleased delivery:** Migration `000031` and the application changes implement live attributed Google photo resolution plus durable S3-compatible owner/licensed media, OCR-aware template placement, admin upload/status controls, and non-destructive image imports.
**Public safety:** Menu documents and unclassified photos fail closed across APIs, demo snapshots, templates, legacy fallbacks, uploads, and SEO; Google URLs, resource names, and bytes are never persisted or passed through the Next.js optimizer.
**Mobile demos:** The template switcher now clearly shows the current and next design, says restaurant content stays unchanged, removes Elysian's duplicate mobile control, and preserves all signed-demo/restaurant query parameters.
**Verification / release gate:** 165 Go tests, vet/build, 13 OCR/import tests, Python compilation, admin lint, both TypeScript checks, both Node 22 production builds, OpenAPI/Compose validation, local three-template HTTP smoke checks, and diff checks pass. Deployment still requires explicit approval, migration `000031`, and configured `STORAGE_*`/public CDN variables.
