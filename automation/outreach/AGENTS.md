# Outreach Automation Agent Context

## Scope

These instructions apply to `automation/outreach/**` and specialize the
repository-root contract for this subtree. They must not weaken its global
safety or approval rules. This directory writes the same production PostgreSQL
contracts consumed by the Go backend, so Python changes are not isolated.

## Supported Production Path

- Durable city work begins through `POST /api/v1/scrape-jobs` and is consumed
  by `city_scrape_worker.py`.
- `scrape_job_store.py` persists grid cells, candidates, leases, checkpoints,
  request windows, and resume state.
- Google Places is the discovery source. Apollo enrichment is optional and may
  add missing owner/work-email details; a missing credential, no-match, or
  provider failure must not discard valid Places data or fail the city job.
- `import_to_db.py` upserts restaurants/profiles with recorded
  `inferred_business` consent evidence. The database enrollment function then
  admits eligible restaurants to the active administrator-approved sequence.
- Actual scheduled delivery, Gmail rotation, inbox sync, and delivery evidence
  are implemented in Go under `backend/internal/outreach`.
- Image OCR is retired from production and is not a lead or outreach
  prerequisite. Do not restore an OCR container, cron job, budget, or provider
  key as a parallel path.
- `tuvi_outreach_agent.py`, broad fetch scripts, `city_pipeline.py`,
  `daily_ingestion.py`, `daily_pipeline.py`, and cron wrappers are legacy or
  controlled manual tooling. Do not schedule them beside the durable worker.

Follow `docs/runbooks/lead-scrape-outreach.md` for production operation.

## Implementation Patterns

- Modules are intentionally flat and use local imports. Follow the existing
  import/test pattern instead of introducing a second package layout.
- Load configuration through `env_loader.py`; never hard-code credentials.
- Keep Places/Apollo accounting in `request_budget.py` and durable PostgreSQL
  ledgers, not process memory.
- Preserve Place-ID deduplication, canonical identities, fenced claims,
  checkpoint/resume behavior, retry cooldowns, and idempotent writes.
- Provider calls need explicit timeouts and safe handling for rate limits and
  partial coverage. A restart must not reset or duplicate request accounting.
- Generated data belongs only in the existing ignored `leads/`, `data/`,
  `output/`, `drafts/`, `reports/`, `logs/`, or `state/` locations.
- Add `unittest` coverage beside the module using the existing `*_test.py`
  naming convention.

## Dependency Impact

For every direct SQL, job-payload, consent, or lifecycle change, inspect:

- `backend/migrations/` for tables, columns, constraints, functions, and
  persisted check values.
- `backend/internal/store/store.go` for startup schema requirements.
- `backend/internal/scrapejobs/` for city-job API and resume behavior.
- `backend/internal/restaurants/` for lifecycle and consent contracts.
- `backend/internal/outreach/` and `backend/internal/campaigns/` for enrollment,
  sequence eligibility, schedules, quota, and delivery state.
- `apps/web/` for admin-visible statuses, reasons, counters, and resume actions.

Do not rename a persisted state, alter an identity rule, or change a job payload
without updating every producer, consumer, migration, and recovery path.

## Safety Invariants

- Never print, commit, or copy `.env` contents, provider keys, OAuth tokens,
  lead personal data, or temporary signed URLs into logs or fixtures.
- Do not make real provider calls merely to test code. Mock network calls in
  unit tests.
- Preserve the durable Places/Apollo request window; a restart must not reset
  the budget.
- Missing optional Apollo data must not make a valid Places lead disappear.
- Importing a record must not make scraped media public; Google media still
  requires attribution and owner/licensed assets require explicit approval.
- Never send real outreach, enable the production email job, resume a
  production scrape, mutate production data, or deploy from this directory
  without the root approval gate.

## Verification

Run from the repository root with Python 3.12:

```bash
rtk automation/outreach/.venv/bin/python -m py_compile <changed-python-files>
rtk automation/outreach/.venv/bin/python -m unittest discover -s automation/outreach -p '*_test.py'
```

For schema, enrollment, or Go-consumed contract changes, also run the affected
Go tests and normally `rtk go test ./backend/...`. Build
`infra/docker/Dockerfile.outreach` when its COPY allowlist or requirements are
affected. Report mocked coverage separately from any explicitly approved
database or external-provider verification.
