# Session Summary

**Latest (2026-07-14):** Release `ffdf1e6` is deployed on the Tuvi VM against an isolated `monorepo` PostgreSQL role/database with all 23 migrations applied; local and VM credentials are stored only in ignored mode-`0600` env files.
**Workflow:** The deployed stack includes durable 500-call city jobs, Places-first/Apollo enrichment, explicit OCR states, provenance-bound human-reviewed drafts, and PostgreSQL-managed 40-email account rotation.
**Safety:** Email and OCR remain disabled, provider keys are deliberately absent, and no Melbourne scrape, provider call, OCR claim, or email attempt was triggered during deployment.
**Verification / rollback:** All changed images built, public services return HTTP 200, the DB tunnel and role isolation were verified, and fresh pre-cutover/post-migration backups plus the previous source tree are retained on the VM.
