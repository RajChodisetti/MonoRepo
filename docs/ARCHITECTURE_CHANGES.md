# Crucial Architecture Changes

Date: 2026-07-14

This is the short version of the changes made to the restaurant lead workflow.

| Area | Before / missing | Current state |
|---|---|---|
| City scraping | Scraping was effectively a one-shot process with no durable city job, checkpoint, or reliable rate-limit recovery. | A private HTTP API creates durable city jobs. PostgreSQL stores grid cells, page/candidate progress, provider-call counts, status, and `resume_at`. Work stops at 500 combined Places/Apollo calls, resumes after 24 hours, dynamically subdivides dense cells, deduplicates Place IDs, and revisits cities for new listings. |
| Places and Apollo | Provider order and fallback behavior did not reliably preserve valid restaurant leads or control Apollo usage. | Google Places is the discovery source. Apollo runs afterward only when owner or work-email details are missing. An Apollo no-match does not discard the Places lead. |
| OCR and lead readiness | OCR used a boolean-style completion signal, so failures and missing images could look complete. | OCR is a durable state machine: `pending`, `running`, `verified`, `no_images`, and `failed`. Only `verified` leads proceed; missing images remain visibly unresolved, and retries/claims survive worker restarts. |
| Demo and campaign creation | OCR completion was not safely connected to reusable sales artifacts or proof of what was reviewed. | Verified OCR queues idempotent creation of a demo draft and campaign draft. Artifacts retain OCR/profile provenance, and real outreach still requires human profile approval, demo publication, campaign approval, and an administrator starting the send. |
| Demo access | Preview access did not have a fully defined expiry, revocation, rotation, and safe-payload contract. | Each demo uses a random opaque token whose bcrypt hash is stored in PostgreSQL. Payloads stay server-side; access is expiring, revocable, rotatable, and limited to public-safe data. |
| Email delivery | Send limits and account rotation were not durable enough for restarts or multiple workers, and delivery sequencing was incomplete. | Google Workspace Gmail API and Zoho are supported through HTTP(S) only. PostgreSQL enforces 40 attempts per account, account rotation, a 24-hour cooldown, leases, immutable recipient attribution, and global/per-account send sequences. Ambiguous sends fail closed instead of being retried blindly. |
| Production data and recovery | The application lacked a dedicated least-privilege database boundary and a fully reproducible migrated deployment for this workflow. | The VM now runs the application against an isolated `monorepo` PostgreSQL role/database with all 23 migrations applied. Credentials are kept in ignored mode-`0600` env files, database access is loopback/tunnel-only, and pre/post migration backups plus the previous release are retained. |

## Current production position

The architecture is deployed and the Places, Apollo, optional SerpAPI, and Hugging Face credentials are stored in protected env files. External automation remains intentionally safe: email and OCR are disabled, Google Workspace mailbox OAuth credentials are not configured, and no Melbourne scrape, OCR run, or real email send was triggered during this update. The operator procedure is in [lead-scrape-ocr-outreach.md](runbooks/lead-scrape-ocr-outreach.md).
