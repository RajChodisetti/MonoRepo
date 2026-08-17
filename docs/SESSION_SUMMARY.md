Release `13a0a47` is deployed to QA API and production API/admin/template at schema 54; Elysian (`3`) is now the default and first generated preview.
Canonical and personal-mirror `master` point to the merged PR #3 revision, while the original dirty worktree remains preserved and no committed branch contains unique work outside `master`.
At final verification, production outreach remained operator-enabled with the same queued job; the worker was intentionally unchanged on `4d6ea73`, with the same container fingerprint and zero restarts.
QA/production image, HTTP, template `1`/`2`/`3`, database, control/job-set, rollback, and fatal-log checks passed without triggering an outreach send or provider health action.
