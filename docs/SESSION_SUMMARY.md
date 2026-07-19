# Session Summary

Release `cbc2eb8` is pushed to `master` and deployed on the VM at `/opt/tuvi/releases/monorepo-cbc2eb8`; `/opt/tuvi/MonoRepo` points there and rollback points to `6c21c15`.
Migration `000032` is applied on the active `monorepo` database: schema is 32, `demo_ready=22`, `lead=922`, and no `lead` rows still have demo records.
UI/UX Pro Max for Codex was installed with `uipro init --ai codex` under `.codex/skills/ui-ux-pro-max`; local and VM smoke checks passed through `python3`.
All Tuvi containers are running with zero restarts; API/admin/website/demo/voice HTTPS checks return 200, outreach remains disabled, and active bulk-send jobs are zero.
Production OCR profile counts are currently `pending=940`, `failed=4`, `verified=0`, so OCR verified-only will stay empty until OCR produces new verified rows.
