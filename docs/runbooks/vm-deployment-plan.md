# VM deployment plan

Status: production uses release directories under `/opt/tuvi/releases`,
protected configuration under `/opt/tuvi/env`, and
`infra/docker/docker-compose.vm.yml`.

## Services

- PostgreSQL and migration job
- Go API and durable worker
- Places/Apollo scrape worker
- internal admin portal
- Tuvi corporate website
- restaurant template
- restaurant services catalog
- voice runtime and Redis where enabled

Image OCR is retired. Do not restore an OCR container, cron entry, provider key,
or one-shot job.

## Environments

- Production website: `https://tuvisolutions.com`
- Production admin/API: `https://api.tuvisolutions.com/admin`
- Production demo: `https://demo.tuvisolutions.com`
- QA website: `https://qa-tuvi.170.64.154.143.sslip.io`
- QA API: `https://qa-api.170.64.154.143.sslip.io`

The current QA footprint contains isolated PostgreSQL, Redis, API, and corporate
website services. It does not currently include separate QA admin, durable
worker, or scrape-worker services. Deploy and verify only the services that are
already configured there; adding a new QA service is an infrastructure change
that requires a separate review.

## Protected configuration

Keep these files mode `0600` and never print their values:

- `/opt/tuvi/env/stack.env`: database, auth, email, and application config
- `/opt/tuvi/env/ingestion.env`: Places/Apollo ingestion credentials and limits
- `/opt/tuvi/env/places-api.env`: API-side Places photo resolver only
- `/opt/tuvi/env/llm.env`: API-only public-review provider, model, and key
- voice/provider-specific files already used by the voice deployment

The public digital-footprint report reads `LLM_PROVIDER`, `LLM_API_KEY`, and
`LLM_MODEL` from the API-only `llm.env`. The production model should be
vision-capable and low latency. Reusing another protected OpenAI secret requires
an explicit, in-place host update; never copy the secret into source, shell
history, or logs.

## Release procedure

1. For a full-stack rollout, keep the persisted outreach email job disabled.
   If an operator-approved job is already enabled, an unrelated service-scoped
   release must preserve its control tuple, immutable job identity/count and
   newest-created tuple, and running worker container/image/start/restart
   fingerprint exactly; deploy only unaffected services with `--no-deps`.
   Existing-job status, attempts, leases, and delivery progress may advance.
2. Record the currently running image IDs, commit, migration version, outreach
   control/job aggregates, and aggregate health without exposing customer data.
3. Back up PostgreSQL and protected environment files.
4. Create a new immutable release directory from the reviewed commit.
5. Validate both Compose files and build only the affected service images. Pin
   promoted services to immutable revision-labelled image tags.
6. Run the migration container only when reviewed migration files changed;
   otherwise verify the exact expected migration lineage and schema version.
7. Recreate only the reviewed services from the same release. Do not use a
   generic command that could start dependencies or services outside the
   release scope.
8. Use only the read-only production smoke routes listed below. Keep workflow
   mutations, provider simulations, and report-generation checks in an isolated
   QA/canary/test environment.
9. Confirm there is no OCR service/process/cron/config and no running legacy
   image-analysis claim.
10. Confirm the production email control tuple, immutable job
    identity/count/newest-created tuple, and worker fingerprint match the
    pre-release record. Existing-job lifecycle and delivery progress may
    advance. Enabling or disabling outreach is a separate deliberate operator
    action, never a side effect of deployment.

## Required production read-only smoke checks

- API health/readiness, public restaurant-list read, authentication boundary,
  exact image revision, and exact expected migration lineage
- admin login and unauthenticated redirect boundary; use an existing protected
  session only for a pure identity/BFF read
- unsigned template renders for the default, explicit `1`/`2`/`3`, and invalid
  fallback without creating an engagement session
- unchanged outreach control/job identity and worker fingerprints, allowing
  normal progress by the already-running job
- no fatal logs and no restart-count increase outside the promoted services
- no real provider send or health action during deployment smoke tests

Do not call authenticated `/api/v1/outreach/bulk-send/status` or load
`/admin/outreach` in a browser during a preservation deployment. That status
read reconciles stale delivery attempts and can write attempts, events, and
campaign state.

## Isolated QA/canary behavior checks

- public restaurant search and report with a non-provider fixture or cached hit
- report completion near the configured deadline; blocked sites return a
  labeled partial result rather than holding the request
- mobile report identity/map/photos visible above the fold
- sequence draft/edit/reorder/approve workflow
- plain-text preview matches the saved database template; no unsubscribe copy
  or URL is injected by application code
- the retired application unsubscribe route returns 404 and the sender does not
  consult or write the legacy suppression table
- follow-up claim order precedes new-recipient claims
- failed/unknown fake-provider attempt does not advance the step
- confirmed fake-provider attempt advances once and schedules the configured
  delay
- no real provider send; all provider outcomes are fixture-backed

## Rollback

Recreate the previous release images first. Apply a migration down only when
the release-specific rollback notes allow it and no new sequence progress would
be lost. Restore protected configuration from the backup without printing it.
Historical OCR columns are intentionally retained during stabilization, but
retired credentials and services are not restored.
