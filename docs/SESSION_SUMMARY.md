Local branch `codex/sydney-send-window-inbox-reply-20260815` adds migration 54 and an internal-admin UI for a persisted Australia/Sydney scheduled-outreach window, defaulting to 07:00-12:00.
Workers read the saved window transactionally without restart and reject a window that cannot fit configured daily mailbox quota; direct/test/reply/health email remains exempt.
Inbox Reply remains available after prior admin responses and sends through the captured receiving mailbox/address, including safe plus-address normalization.
Backend, admin, migration-discovery, and OpenAPI checks pass; QA-first then production deployment is explicitly approved, while real outreach remains disabled and unauthorized.
