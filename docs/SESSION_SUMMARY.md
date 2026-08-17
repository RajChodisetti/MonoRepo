Release `8c34503` is deployed to production API/admin/worker and PR head `ff5f634` to the existing QA API at schema 54; canonical and personal-mirror `master` are synchronized.
The `contact` Gmail tombstone loop is fixed: all six mailboxes have successfully polled since rollout, their stored errors are clear, and non-message-404 failures remain retryable.
The admin now defaults to named restaurant-linked inbox threads; the final 10-day snapshot shows 6 restaurant threads and hides 358 unmatched/unnamed threads unless the filter is disabled.
Production outreach control and the future queued job are unchanged, with 21 sent, 2 failed, and zero in-flight attempts; deployment triggered no send, migration, or provider health action.
QA/prod rollback containers, immutable image tags, mode-0600 database/config backups, and the original dirty worktree remain preserved.
