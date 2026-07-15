# ADR: Durable, resumable city scraping

Date: 2026-07-14
Status: Accepted

## Context

City scraping previously ran as a one-shot Python process. Its request counter,
place pagination, and deduplication progress were not durable, so a restart
could repeat provider calls or lose the exact point at which a daily allowance
was exhausted. A broad city query can also saturate the Places result window
and omit businesses.

## Decision

- The Go API creates an internal-admin `scrape_jobs` record per city and niche.
- A PostgreSQL-backed Python worker owns provider execution and checkpoints each
  grid cell, page token, Place ID candidate, request reservation, lease, and
  import state.
- Every state mutation is fenced to the current unexpired job lease, preventing
  a stale worker from overwriting a replacement worker after takeover.
- Each job starts with a 4x4 grid. A cell that fills three 20-result pages is
  split into four children, up to a configurable depth.
- Google Places runs first. Apollo runs only for the missing owner/work-email
  fields of the same discovered business.
- A combined Places/Apollo allowance is atomically capped at 500 provider calls.
  At the cap, work waits until `resume_at = now() + 24 hours` and resumes from
  persisted state.
- A fully covered city enters a 24-hour `revisit` wait rather than a terminal
  state. The next cycle scans existing leaf cells again and imports only new
  Google Place IDs.
- A cell that is still saturated at the configured maximum depth is recorded as
  `coverage_incomplete`, allowing later sibling cells to finish without falsely
  completing the city pass. Capped leaves retry after a dedicated 24-hour wait
  and are surfaced by the `cells_saturated` progress count.

## Options Considered

- Keep the daily one-shot script and JSON checkpoint files: rejected because
  process restarts and multiple workers cannot coordinate safely.
- Use one city-wide Places query: rejected because provider result windows can
  hide dense-area businesses.
- Use a new external workflow engine: deferred; PostgreSQL already provides the
  leases, uniqueness, and checkpoints needed for Phase 1.

## Consequences

- The API and scrape worker must share the same PostgreSQL database migrated
  through `000023`.
- `waiting` is a healthy continuous-job state; `last_cycle_completed_at`
  distinguishes completed coverage from rate-limit waits.
- The 500-call allowance covers Places and Apollo calls/retries. Website HTTP
  fetches are not provider-API calls and are not included.
- Google cannot guarantee exhaustive results, but subdivision and recurring
  Place-ID-deduplicated passes materially improve coverage.

## Rollback / Revisit Trigger

Disable the `scrape-worker` service and stop creating jobs. Roll back `000023`
before `000015`, and only remove `000015` after confirming that its checkpoints
are no longer required.
Revisit the design if more cities require dynamic geocoding or PostgreSQL
polling becomes a throughput bottleneck.
