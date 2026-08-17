# Backend Agent Context

## Scope

These instructions apply to `backend/**` and specialize the repository-root
contract for this subtree. They must not weaken its global safety or approval
rules. Run Go and Make commands from the repo root because `go.mod` and
`Makefile` live there.

## Architecture and Entry Points

- `cmd/api/main.go` starts `internal/app.NewAPI`; do not run an
  `internal/app` file directly.
- `cmd/worker/main.go` starts the PostgreSQL-backed worker.
- `cmd/migrate/main.go` applies paired SQL files from `migrations/`.
- `internal/app/` owns dependency wiring and process lifecycle.
- `internal/http/router.go` is the route and middleware registry;
  `internal/http/handlers/` translates HTTP only.
- Domain services own validation, authorization, and state transitions.
- Repository interfaces and `postgres.go` implementations own persistence.
- `internal/store/store.go` composes repositories and verifies the required
  startup schema.
- `internal/jobs/` owns durable job contracts and handlers.
- External systems belong behind interfaces under `internal/providers/`.

Typical flow:

```text
cmd -> app wiring -> router/middleware -> handler -> domain service
    -> repository/provider -> PostgreSQL or external service
```

## Implementation Patterns

- Inspect the existing domain package before adding a new abstraction.
- Keep handlers thin and return the existing safe JSON error shape.
- Pass `context.Context` through request, database, job, and provider calls.
- Use UUIDs for public identities and transactions for multi-step writes.
- Preserve repository interfaces, mocks, and adjacent service tests.
- Reuse `platform/errors` for persistence-level not-found behavior.
- Enforce authenticated/private restaurant scope through both role checks and
  membership/access checks. Never trust a restaurant ID supplied by a client by
  itself; public routes must use their explicit public-data and write policies.
- Add a new sequential `.up.sql` and `.down.sql` pair for every schema change.
  Discover the latest migration first; never renumber an applied migration.
- Update startup schema verification when runtime correctness depends on a new
  table, column, function, or constraint.
- Keep email, Places, storage, calendar, and future AI SDK details inside
  provider adapters.

## Dependency Impact

Before finishing, check these cross-subsystem contracts:

- Route or response changes: update HTTP tests and `docs/openapi/openapi.yaml`;
  update consuming TypeScript types/adapters when applicable.
- Public demo/site payload changes: inspect `internal/demos`,
  `internal/profiles`, `internal/media`, `template/src/data/types/restaurant.ts`,
  and `template/src/lib/adapters/restaurantSiteApi.ts`.
- Scrape, restaurant/profile, consent, or job schema changes: inspect direct
  SQL in `automation/outreach/`, especially `scrape_job_store.py` and
  `import_to_db.py`.
- Admin API changes: inspect `apps/web/src/lib/types.ts` and its BFF consumers.
- Consultation or reservation changes: inspect `web/` and
  `voice-sales-agent/` callers.
- Worker changes: verify job registration in `internal/app/worker.go`, job
  idempotency, leases, retry behavior, and provider ambiguity handling.

## Safety Invariants

- Never expose secrets, bearer tokens, private lead notes, raw enrichment, or
  unapproved media through public routes or logs.
- Public demo payloads remain server-side and token gated.
- Reservations remain `pending` unless explicit confirmation rules exist.
- Real sequence outreach requires recorded inferred-business consent, an
  eligible restaurant, an administrator-approved sequence version, and the
  persisted send control, schedule, quota, and delivery ledger. Do not bypass
  those gates. OCR, media review, and published demos are not outreach
  eligibility requirements in the current workflow.
- The approved database sequence is the sole owner of unsubscribe content.
  Application code must not inject or require unsubscribe copy/tags, expose an
  unsubscribe route, or consult the legacy suppression table. Reintroducing
  that behavior requires an explicit decision and a superseding ADR.
- Webhooks and external callbacks require authentication/signature validation
  where supported.
- Auth, tenant isolation, production migrations, provider routes, real sends,
  and deployments retain the root approval requirements.

## Verification

Use the smallest targeted check first, then the relevant full checks:

```bash
rtk gofmt -w <changed-go-files>
rtk go test ./backend/internal/<affected-package>/...
rtk go test ./backend/...
rtk go vet ./backend/...
rtk go build ./backend/cmd/...
rtk make openapi
```

Run migration checks for migration work. `backend/tests/` is currently only an
integration-test slot, so do not claim database integration coverage unless a
real PostgreSQL-backed check was run. Report every skipped or unavailable check.
