# Session Summary

Branch `codex/reconcile-outbound-inboxes-ramp-20260813` deploys credential overrides at commit `7e72df6` from `/opt/tuvi/releases/monorepo-7e72df6-db-precedence-20260814T200818Z`.
QA API and production API/worker/admin use the new release; all other QA and production services retained their prior start times and zero restart counts.
Every existing account can replace credentials from the UI, and exact database identities take precedence over environment values across sending, health, replies, and inbox polling.
QA and production remain at schema 52 with zero database credentials; both email jobs remain disabled and the deployment created no delivery attempts, outbound snapshots, health checks, or provider email.
Verified database/config backups and explicit `6f1a4da` rollback images are retained; authenticated production browser QA showed all three replace actions with no console errors.
