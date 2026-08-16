Release `4d6ea73` is deployed to QA API and production API/worker/admin at schema 54 with matching immutable digests, revision labels, and runtime version markers.
Saved signatures now rehydrate authoritatively, test-send skips definitively rejected sender credentials, scheduled rotation quarantines/defer-recovers safely, and Template 1 shows only canonical `greeting01`.
QA/production HTTP, database, disabled-outreach, no-delivery, zero-restart, and fatal-log invariants passed; stale release pointers were reconciled to the new release.
Outreach remains disabled and no provider email or health probe was sent; a separately approved test-recipient smoke and live PostgreSQL CTE integration coverage remain outstanding.
