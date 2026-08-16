Release `a07fffc` is deployed to QA and production with schema 54 and an internal-admin UI for a persisted Australia/Sydney scheduled-outreach window, currently 07:00-12:00.
Six production mailboxes are interleaved so their aggregate 240-message daily allowance can fit inside that window while each mailbox retains its own cadence; scheduled outreach remains disabled with no active job.
Inbox Reply remains available after prior admin responses and sends through the captured receiving mailbox/address, including safe plus-address normalization; direct/test/reply/health mail remains outside the scheduled-outreach policy.
QA and production checks pass on backend image `6f36bf30b15d`; migration backups and release/image rollback targets are retained.
