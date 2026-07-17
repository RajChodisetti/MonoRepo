# Session Summary

**Latest (2026-07-16):** Merged `origin/master`'s admin_portal PR into this branch — `apps/web` is now the real Next.js internal admin console (port 3002), not a placeholder.
**Admin portal:** Dashboard, scrape-jobs, restaurant/profile review, demo/campaign approval, and outreach bulk-send screens, proxied to the Go API via `/api/admin/*` with an httpOnly session cookie.
**Docs:** `docs/SERVICES.md` and `AGENTS.md` LIVING MEMORY updated to reflect the new repo shape; `RTK.md` reviewed and left unchanged (still accurate).
**Gap:** The admin portal is not yet in `infra/docker/docker-compose.vm.yml` or the VM Caddyfile — no public domain serves it yet.
**Previous (2026-07-15):** PostgreSQL is the sole authority for consultation availability/bookings; migration 25 and commit `000cdd8` are deployed and healthy.
