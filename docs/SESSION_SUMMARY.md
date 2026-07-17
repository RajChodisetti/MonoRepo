# Session Summary

**Latest (2026-07-17):** With explicit approval, production OCR was configured for `Qwen/Qwen2.5-VL-7B-Instruct:hyperbolic` and a one-shot five-restaurant pilot was run after validated database and ingestion-config backups.
**Pilot result:** Exactly five profiles and 50 Google Places images were attempted; all calls returned HTTP 400, producing 0 verified and 5 failed rows. Production now has 8 failed and 471 pending profiles, with no running claims, persisted OCR images, new drafts, campaigns, or outreach actions.
**Root cause:** Hugging Face's live model metadata now maps this model only to `featherless-ai`; Hyperbolic is no longer a live provider for it, and no successful inference/cost/latency sample was produced.
**Safety:** Persistent OCR remains disabled, no OCR cron or container is running, API/admin health checks return 200, and the root-only backup/log audit trail is intact.
**Approval needed:** Any switch to Featherless or another current VLM/provider is a separate production model-route decision; start with a non-restaurant compatibility probe before another database pilot.
