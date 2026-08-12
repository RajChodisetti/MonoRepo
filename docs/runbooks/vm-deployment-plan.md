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

1. Keep the persisted outreach email job disabled.
2. Record the currently running image IDs, commit, migration version, and
   aggregate health without exposing customer data.
3. Back up PostgreSQL and protected environment files.
4. Create a new immutable release directory from the reviewed commit.
5. Validate both Compose files and build the API, worker, scrape worker,
   admin, corporate website, and template images.
6. Run the migration container and verify the expected schema version.
7. Recreate the services from the same release. Do not use a generic command
   that could start services not present in the reviewed Compose file.
8. Verify health/readiness, public report partial-fallback behavior, admin
   sequence endpoints, the absence of application unsubscribe routes, and
   aggregate enrollment counts.
9. Confirm there is no OCR service/process/cron/config and no running legacy
   image-analysis claim.
10. Confirm the production email job is still disabled. Enabling it is a
    separate deliberate admin action after previews and counts are reviewed.

## Required smoke checks

- API health and migration version
- admin authentication and BFF proxy
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
- no real provider send during deployment smoke tests

## Rollback

Recreate the previous release images first. Apply a migration down only when
the release-specific rollback notes allow it and no new sequence progress would
be lost. Restore protected configuration from the backup without printing it.
Historical OCR columns are intentionally retained during stabilization, but
retired credentials and services are not restored.
