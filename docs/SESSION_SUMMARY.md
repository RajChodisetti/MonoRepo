# Session Summary

**Latest (2026-07-08):** Added `docs/runbooks/vm-deployment-plan.md`, a complete VM deployment draft for the current MonoRepo services.

**Deployment state:** The repo already containerizes the Go API, worker, migrations, PostgreSQL, voice agent, and Redis; the catalog, Tuvi Next site, demo template, reverse proxy, TLS, backups, and production env wiring still need VM Compose/proxy work.

**VM audit:** Local SSH config has no VM alias, so live VM inspection is pending; the runbook includes exact audit commands to run once the VM host/deploy user is available.

**Branch state:** `phase1_03/backend` remains clean and ahead of `origin/phase1_03/backend` by the local catalog preservation commit plus this planning doc update.

**Next:** Confirm the VM SSH target and preferred domain/subdomain layout, then convert the plan into Compose/proxy files.
