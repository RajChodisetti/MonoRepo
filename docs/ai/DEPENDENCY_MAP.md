# Dependency and Change-Impact Map

Use this map before editing a contract. It lists the seams that have produced or
can produce regressions; it is intentionally narrower than a full code map.

## Runtime flow

```text
admin browser
  -> apps/web same-origin BFF
  -> Go HTTP router/handlers
  -> domain services/repositories
  -> PostgreSQL

durable city scrape (Python)
  -> PostgreSQL restaurant/profile rows + inferred-business consent evidence
  -> active administrator-approved sequence enrollment
  -> persisted operator send control + Sydney schedule + mailbox quota
  -> Go outreach worker + Gmail adapter + delivery-attempt ledger

public demo/site
  <- Go demo/profile/media/reservation/engagement handlers
  <- template adapters and shared RestaurantContent types
  <- Cinematic (1), Aurora (2), Elysian (3)

corporate booking
  -> web server routes
  -> Go consultation API
  -> PostgreSQL consultation slots

browser voice (corporate or restaurant)
  -> voice-sales-agent
  -> Go consultation or pending-reservation APIs

standalone prototypes (not production restaurant dependencies)
  -> apps/andre-admin -> andre-voice-agent: legacy real-estate UI/agent with outbound code
  -> ocr-electrical-poc: electrical-image Gemini proof of concept
```

## High-risk seams

| Change trigger | Required producers/consumers to inspect | Minimum evidence |
| --- | --- | --- |
| SQL column/table/lifecycle state | Paired `backend/migrations` files; Go repositories/services; `backend/internal/store`; Python direct SQL/import; startup migration tests; admin/public response types | Migration discovery test, affected Go tests, affected Python tests, disposable-DB apply/rollback when behavior is nontrivial |
| Private API route or JSON field | `backend/internal/http/router.go`; handler/service/repository; auth/access middleware; `docs/openapi/openapi.yaml`; `apps/web/src/lib/{api,client-api,types}.ts`; relevant BFF route/UI | Go handler/service tests, OpenAPI validation, admin lint/type-check/build |
| Public restaurant/demo payload | `demos/payload.go`; `profiles/site*.go`; media service; public handlers; template `RestaurantContent` types; adapters; all three renderers | Go payload/handler tests, template type-check/build, smoke IDs 1/2/3, verify no private fields |
| Template ID/default/order | Template config/switcher; three template wrappers; backend generated links; campaign render/order; admin types/UI; Docker `TEMPLATE` arg; env example/readme | Targeted Go tests, admin type-check, template type-check/build, server-rendered smoke for 1/2/3 |
| Acquisition/import/consent | Python city worker, request ledger, Places/Apollo merge, `import_to_db.py`; restaurant lifecycle and inferred-business evidence; enrollment function; admin scrape UI | Python scrape/import tests, Go restaurant/outreach tests, durable resume and no-Apollo fail-open cases |
| Campaign/outreach behavior | Approved sequence versions and steps; enrollment; runtime control and Sydney schedule; mailbox ramp/quota/failover; tracking/inbox; admin outreach UI | Go campaign/job/outreach/provider tests, admin checks, and evidence that deployment did not enable or trigger sending |
| Auth/tenant access | Auth service/JWT; router middleware; restaurant membership; every affected handler and BFF/session boundary | Unauthorized, wrong-role, and cross-restaurant tests plus happy path |
| Media source/visibility | Places resolver; storage adapter; media/profile repositories; approval metadata; admin controls; public payload; source-aware template components | Go media/public boundary tests, template build/smoke, attribution and menu-document fail-closed checks |
| Consultation contract | Go consultations handler/service/repository; corporate server routes/forms; voice corporate tools; typed config/env | Go consultation tests, corporate build, safe mocked voice/tool test |
| Voice prompt/tool change | Restaurant and corporate prompts/tools; `bot.py`; API clients; reservation/consultation contracts; disclosure/escalation policy | Inbound-only policy suite, static/import check, and safe browser/health smoke; never place a real call without approval |
| Andre prototype change | `andre-voice-agent` browser/Twilio routes, prompts, property store, call logs, opt-out controls, and dialing guardrails | Static compile and isolated local checks only; never invoke `/call`, `dial.py`, callback tools, or a provider without separate explicit approval |
| Andre admin change | `apps/andre-admin` session/BFF routes, property UI, browser voice, outbound dial UI; `andre-voice-agent` API/auth contracts | Lint, type-check, and build with local-only configuration; never smoke the call route or expose server credentials to the browser |
| Electrical OCR prototype change | `ocr-electrical-poc` schema, prompts, Gemini adapter, fixtures, and dataset tooling | Static compile and fixture-only validation; never upload real images or call the provider without explicit approval |
| Config/env variable | `.env.example`; typed Go/Python/Next config; Compose/Docker build args and runtime env; deployment runbook | Config tests, Compose render, relevant production-equivalent build; never print values |
| Compose/deployment topology | `infra/docker`; root scripts/Make targets; service ports/health checks; `docs/SERVICES.md`; rollback runbook | `docker compose ... config`, targeted image build/smoke; production action still requires approval |

## Backend dependency direction

```text
cmd -> internal/app -> http/jobs -> domain services -> repositories/interfaces
                                       -> provider interfaces
repositories/provider adapters -> PostgreSQL or external provider
```

Handlers must not become the policy layer. Domain packages must not import HTTP
handlers. Provider-specific logic must not leak into domain services. Wiring
changes normally belong in `backend/internal/app` or `backend/internal/http`.

## Frontend boundaries

- `apps/web`: internal admin only; browser → same-origin BFF → Go API. Never
  expose the bearer token or move server-only env variables into client code.
- `template`: public restaurant/demo experience. Shared adapters/types are the
  contract seam; every template must handle optional/empty data safely.
- `web`: canonical Next.js 16 corporate marketing, restaurant-report, and
  booking site. It calls the main Go API through same-origin server routes.
- `apps/restaurant-services-catalog`: self-contained static Vite site. Only
  `VITE_*` public configuration is allowed.

## Verification matrix

The router prints a tailored subset. Canonical commands are:

```text
backend:        rtk go test ./backend/...
                rtk go vet ./backend/...
                rtk go build ./backend/cmd/...
admin:          rtk npm --prefix apps/web run lint
                rtk npm exec --prefix apps/web -- tsc --noEmit --incremental false --pretty false -p apps/web/tsconfig.json
                rtk npm --prefix apps/web run build
Andre admin:    rtk npm --prefix apps/andre-admin run lint
                rtk npm exec --prefix apps/andre-admin -- tsc --noEmit --incremental false --pretty false -p apps/andre-admin/tsconfig.json
                rtk npm --prefix apps/andre-admin run build
template:       rtk npm --prefix template run test:unit
                rtk npm exec --prefix template -- tsc --noEmit --incremental false --pretty false -p template/tsconfig.json
                rtk npm --prefix template run build
corporate:      rtk npm --prefix web run lint
                rtk npm exec --prefix web -- tsc --noEmit --incremental false --pretty false -p web/tsconfig.json
                rtk npm --prefix web run build
catalog:        rtk npm --prefix apps/restaurant-services-catalog run build
Python:         rtk automation/outreach/.venv/bin/python -m unittest discover -s automation/outreach -p '*_test.py'
voice:          rtk python3 -m unittest discover -s voice-sales-agent/tests -p 'test_*.py'
Andre:          rtk python3 -m compileall -q -x '(^|/)(\.venv|__pycache__)(/|$)' andre-voice-agent
electrical OCR: rtk python3 -m compileall -q -x '(^|/)(\.venv|__pycache__)(/|$)' ocr-electrical-poc
OpenAPI:        rtk make openapi
context docs:   rtk ./scripts/check-agent-context.sh
all changes:    rtk git diff --check
```

The repository currently contains paired migrations through `000054`; discover
the latest pair rather than hard-coding that number into runtime logic. Run
Next.js 16 checks with Node 22 and Python checks with Python 3.12 when host/runtime
behavior can differ. Voice checks must remain non-billable and non-calling unless
explicitly authorized.

For outreach changes, the accepted database-owned unsubscribe ADR is binding:
the saved sequence owns all unsubscribe content, while application tags, routes,
and legacy suppression-table behavior remain retired unless a superseding ADR is
explicitly approved.
