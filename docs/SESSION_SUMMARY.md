Release `6305910` is deployed to QA API and production API/admin at schema 54; the active production worker intentionally remains on `4d6ea73` until its operator-enabled outreach run is idle.
The outreach Operations UI now shows durable provider-confirmed delivery counts for total, Phase 1, Phase 2, and Phase 3, excluding tests, replies, health checks, and failed attempts.
At the final production verification the ledger held 9 confirmed deliveries: Phase 1 = 8, Phase 2 = 1, Phase 3 = 0, Other = 0; the live enabled job may increase these values.
QA/production HTTP, database, UI, zero-restart, worker-preservation, and fatal-log invariants passed, and no send or provider health action was triggered by the deployment.
