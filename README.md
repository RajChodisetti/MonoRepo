# Restaurant Platform

Sales-first restaurant platform for personalized demo websites, versioned
plain-text outreach, reservation requests, inbound AI receptionist prototypes,
and a public digital-footprint review.

## Prerequisites

- Go 1.26 or newer
- Docker Desktop (for local PostgreSQL)

## Local Setup

Run these from the **repo root** (`MonoRepo/`).

### 1. Environment

Copy the example env file and adjust if needed:

```bash
cp .env.example .env
```

Or use `backend/.env` — the app loads `.env`, `backend/.env`, and `../.env` in that order.

Default database URL (matches Docker setup):

```text
DATABASE_URL=postgres://postgres:postgres@localhost:5432/restaurant_platform?sslmode=disable
```

### 2. Quick start (recommended)

**One command — everything (Postgres + migrations + API + worker):**

```bash
make start
```

Press `Ctrl+C` to stop API and worker. Postgres keeps running until `make stop-all`.

**VM / Docker deployment (all services in containers):**

```bash
make up      # build + start postgres, redis, migrate, api, worker, scrape-worker
make logs    # follow logs
make down    # stop stack
```

Copy `backend/.env` (Resend `EMAIL_API_KEY` and related settings) before `make up` on a VM. Set `PUBLIC_BASE_URL` to your server URL.

**Voice sales agent (Docker profile `voice`):**

```bash
# Requires voice-sales-agent/.env (API keys — copy from .env.example)
make voice-up     # build + start agent + Redis on :8000
make voice-logs   # follow agent logs
make voice-down   # stop agent + Redis
```

UI: http://localhost:8000/?agent=corporate

Agent reaches host APIs via `host.docker.internal` (`MONOREPO_API_URL` :8080).
Set `TUVI_API_TOKEN` to the same token used by the main API consultation endpoints.

**Legacy — API only (no email worker):**

```bash
make dev
```

Or step by step:

```bash
make setup   # db-up + migrate-up
make api     # start API (requires database + migrations)
make worker  # second terminal — required for email sends
```

## Commands

All commands below are run from the **repo root**.

### Database (Docker)

```bash
make db-up      # start PostgreSQL and wait until healthy
make db-down    # stop database container (data is kept)
make db-reset   # stop database and delete all data (fresh start)
```

PostgreSQL runs on `localhost:5432` via `infra/docker/docker-compose.yml`.

| Setting  | Value                 |
| -------- | --------------------- |
| Host     | `localhost`           |
| Port     | `5432`                |
| User     | `postgres`            |
| Password | `postgres`            |
| Database | `restaurant_platform` |

### Migrations

```bash
make migrate-up    # apply pending SQL migrations
make migrate-down  # roll back the most recently applied migration
```

Direct Go equivalent (no Make):

```bash
go run ./backend/cmd/migrate up
go run ./backend/cmd/migrate down
```

Migration files live in `backend/migrations/`.

Current migrations:

- `000001_foundation` — `app_metadata`, `job_runs`
- `000002_auth_users` — `users` table
- `000003_drop_users_role_check` — prepares the user-role model migration
- `000004_role_model_restaurants` — `restaurants`, `restaurant_members`, role renames
- `000005_demo_sites` — public demo sites with token-gated access
- `000006_restaurant_lead_fields` — `email`, `is_contacted`, `shown_interest` on restaurants
- `000007_restaurant_status` — lead lifecycle `status` column
- `000008_restaurant_profiles_menus` — restaurant profiles, menus, and menu items
- `000009_email_campaigns` — outreach campaigns, events, tracking tokens, and suppressions
- `000010_restaurant_image_tables` — restaurant and menu-item image records
- `000011_reservations` — pending restaurant reservation requests
- `000012_company_consultations` — Tuvi consultation requests
- `000013_lead_ocr_verified` — legacy OCR verification projection
- `000014_restaurant_email_sent` — `email_sent` flag on restaurants for bulk outreach tracking
- `000015_city_scrape_jobs` — durable city jobs, grid cells, candidates, and request checkpoints
- `000016_ocr_status_and_post_ocr` — claimed OCR states and post-OCR preparation state
- `000017_outreach_email_quota` — durable account quotas, delivery attempts, and send counters
- `000018_lead_review_audit` — attributable profile, demo, and campaign review gates
- `000019_auto_lead_draft_provenance` — OCR fingerprint provenance for generated drafts
- `000020_job_and_delivery_leases` — fenced job and provider-delivery leases
- `000021_immutable_outreach_recipient` — immutable recipient attribution on delivery and tracking rows
- `000022_auto_artifact_profile_provenance` — identity/public-payload provenance for automatic drafts
- `000023_scrape_coverage_incomplete_state` — visible retry state for provider-capped grid leaves
- `000024` through `000041` — durable pacing, consultations, media history,
  engagement, templates, and calendar revisions (see the migration files for
  exact rollback contracts)
- `000042_plain_text_outreach_sequences` — approved sequence versions,
  inferred-business evidence, and integer recipient progress
- `000043_manual_media_review` — explicit owner/licensed media approval audit
- `000044` through `000046` — SEO unlock security and current outreach copy
- `000047_deterministic_restaurant_greeting` — inactive greeting01 sequence draft
- `000048_email_messages` — accepted outbound snapshots and inbound reply capture
- `000049_outreach_email_ramp` — durable 5-to-40 per-mailbox warm-up cycles

**Rollback notes:**

- `migrate-down` only reverses the **last applied** migration.
- Each migration has a matching `.down.sql` file (for example `000001_foundation.down.sql`).
- On production, take a backup before rolling back schema changes.

### App

```bash
make api      # run the Go net/http API on HTTP_ADDR
make worker   # run the Go worker with the PostgreSQL-backed job queue
make test     # run all backend tests
make fmt      # format backend Go files
make openapi  # validate docs/openapi/openapi.yaml
make swagger  # Swagger UI at http://localhost:8081
```

### API documentation (OpenAPI / Swagger)

- Spec: [`docs/openapi/openapi.yaml`](docs/openapi/openapi.yaml)
- Guide: [`docs/openapi/README.md`](docs/openapi/README.md)
- Postman collection: [`postman/`](postman/)

```bash
make swagger   # local Swagger UI (Docker)
make openapi   # validate the spec
```

Or import `openapi.yaml` at [editor.swagger.io](https://editor.swagger.io).

Direct Go equivalents (no Make):

```bash
go run ./backend/cmd/api       # start API
go run ./backend/cmd/worker    # start worker
go test ./backend/...          # run tests
go test ./backend/... -v       # run tests with verbose output
```

**Note:** Always start the API via `cmd/api`, not `internal/app/api.go`:

```bash
# correct
go run ./backend/cmd/api

# wrong — package app is not main
go run ./backend/internal/app/api.go
```

### Shortcuts

```bash
make setup    # make db-up + make migrate-up
make dev      # make setup + make api
make seed-admin  # create first admin user from ADMIN_EMAIL / ADMIN_PASSWORD
make seed-demo-fixture  # restaurant + owner membership + published demo site
```

## Auth, Admin, and Health Checks

### Admin (P1-007)

Admin accounts cannot be created via public signup. Seed the first admin:

```bash
make seed-admin
```

Then login and access protected admin routes:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@local.test","password":"password123"}'

curl http://localhost:8080/api/v1/admin/me \
  -H "Authorization: Bearer <access_token>"
```

### Developer health checks

Health endpoints require a **developer** JWT:

```bash
# Signup (choose role in body — developer for health access)
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@local.test","password":"password123","full_name":"Dev","role":"developer"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@local.test","password":"password123"}'

# Use access_token from the response
curl http://localhost:8080/healthz \
  -H "Authorization: Bearer <access_token>"

curl http://localhost:8080/readyz \
  -H "Authorization: Bearer <access_token>"
```

| Endpoint | Auth | Purpose |
| -------- | ---- | ------- |
| `GET /api/v1/auth/me` | Bearer (any role) | Exposes `user_id`, `email`, `role` from JWT |
| `POST /api/v1/auth/signup` | Public | Create `restaurant_owner` or `developer` (local/test) |
| `POST /api/v1/auth/login` | Public | Verify credentials and receive JWT |
| `GET /api/v1/admin/me` | Bearer + `internal_admin` | Current internal admin profile |
| `GET /api/v1/restaurants` | Bearer + `internal_admin` or `restaurant_owner` | List restaurants (supports query filters) |
| `GET /api/v1/restaurants/shared-emails` | Bearer + `internal_admin` | Group shared emails, counts, and restaurant records |
| `POST /api/v1/restaurants` | Bearer + `internal_admin` | Create restaurant lead |
| `GET /api/v1/restaurants/{id}` | Bearer + membership | Get restaurant (owner or internal admin) |
| `PATCH /api/v1/restaurants/{id}` | Bearer + `internal_admin` | Update name, email, `is_contacted`, `shown_interest` |
| `PATCH /api/v1/restaurants/{id}/status` | Bearer + `internal_admin` | Update lifecycle status |
| `DELETE /api/v1/restaurants/{id}` | Bearer + `internal_admin` | Soft archive (`status = archived`) |
| `GET /api/v1/restaurants/{id}/members` | Bearer + `internal_admin` | List restaurant members |
| `POST /api/v1/restaurants/{id}/members` | Bearer + `internal_admin` | Assign owner to restaurant |
| `GET /api/v1/restaurants/{id}/messages` | Bearer + `internal_admin` | List captured outbound snapshots and inbound replies |
| `GET /api/v1/restaurants/{id}/outreach-greeting` | Bearer + `internal_admin` | Preview restaurant-specific greeting placeholders |
| `POST /api/v1/restaurants/{id}/demo-sites` | Bearer + `internal_admin` | Create a draft demo site (returns one-time token) |
| `GET /api/v1/restaurants/{id}/profile/review-preview` | Bearer + `internal_admin` | Inspect profile/contact and capture review versions without bearer secrets |
| `PATCH /api/v1/restaurants/{id}/profile/review` | Bearer + `internal_admin` | Audit profile approval/rejection for demo/public content |
| `GET /api/v1/demo-sites/{id}/review-preview` | Bearer + `internal_admin` | Inspect exact allowlisted demo payload without exposing its token |
| `PATCH /api/v1/demo-sites/{id}/status` | Bearer + `internal_admin` | Audit demo publish/unpublish after profile approval |
| `POST /api/v1/scrape-jobs` | Bearer + `internal_admin` | Create/return a durable Places-first city scrape job |
| `GET /api/v1/scrape-jobs` | Bearer + `internal_admin` | List recent durable city scrape jobs |
| `GET /api/v1/scrape-jobs/{id}` | Bearer + `internal_admin` | Monitor request window, resume time, grid, and candidate progress |
| `POST /api/v1/scrape-jobs/{id}/resume` | Bearer + `internal_admin` | Resume a failed job from its durable cell/candidate checkpoints |
| `PATCH /api/v1/outreach/email-job` | Bearer + `internal_admin` | Explicitly enable/disable the persisted email job |
| `POST /api/v1/outreach/bulk-send` | Bearer + `internal_admin` | Queue the enabled, quota-managed sequence workflow |
| `GET /api/v1/outreach/bulk-send/status` | Bearer + `internal_admin` | Follow-up/new counts, pacing, and active/last job status |
| `GET/POST /api/v1/outreach/sequences...` | Bearer + `internal_admin` | List, draft, edit, preview, and approve sequence versions |
| `GET /api/v1/outreach/recipients` | Bearer + `internal_admin` | Inspect integer sequence progress and next-due state |
| `GET /api/v1/outreach/inbox` | Bearer + `internal_admin` | List restaurant and unmatched reply threads |
| `POST /api/v1/outreach/messages/{id}/read` | Bearer + `internal_admin` | Mark one captured message read |
| `POST /api/v1/outreach/messages/{id}/reply` | Bearer + `internal_admin` | Send and snapshot a custom same-mailbox reply |
| `GET /api/public/v1/demo/{slug}?token=...` | Public | Public demo payload only (no internal fields) |
| `GET /api/public/v1/restaurants/{id}/table-availability` | Public | Return available RFC3339 reservation-request slots |
| `PUT /api/public/v1/restaurants/{id}/reservations` | Public | Submit an idempotent reservation request in `pending` status |
| `GET /t/click/{token}` | Public | Legacy tracked-campaign redirect (not used by plain-text sequences) |
| `GET /t/open/{token}` | Public | Legacy tracked-campaign pixel (not used by plain-text sequences) |
| `GET /healthz` | Bearer + `developer` role | Process is running |
| `GET /readyz` | Bearer + `developer` role | PostgreSQL is connected and ready |

**Roles:** `internal_admin`, `restaurant_owner`, `developer` (ops/health only in local dev).

### Restaurant access (P1-009)

Restaurant owners only see restaurants they are assigned to via `restaurant_members`.
Internal admins can access all restaurants. Every route under `/api/v1/restaurants/{id}/...`
runs through access middleware.

### Restaurant list filters (P1-010)

Query params on `GET /api/v1/restaurants` (no request body):

| Param | Example | Purpose |
| ----- | ------- | ------- |
| `restaurant` | `?restaurant=thai` | Case-insensitive name search |
| `status` | `?status=lead` | Filter by lifecycle status |
| `is_contacted` | `?is_contacted=true` | Filter by contact flag |
| `shown_interest` | `?shown_interest=true` | Filter by interest flag |
| `include_archived` | `?include_archived=true` | Include archived leads (hidden by default) |

Status values: `lead`, `emailed`, `interested`, `client_onboarding`,
`active_client`, `lost`, and `archived`. `demo_ready` remains accepted only as a
legacy rollback-compatible value; migration `000042` normalizes existing rows to
`lead`, and sequence eligibility never depends on it.

Seed a full fixture for manual testing:

```bash
make migrate-up
make seed-demo-fixture
```

Default fixture credentials:

- Owner: `owner@local.test` / `password123`
- Public demo: `GET /api/public/v1/demo/demo-fixture-cafe?token=demo-fixture-token-value-32chars`

The API starts only after:

1. Database connection succeeds (with retry)
2. Every discovered migration through `000043` is verified (`VerifyStartup`)

On startup you should see logs like `database_connected_successfully`.

## Configuration Notes

Local defaults are safe for development. Production and staging require explicit
`DATABASE_URL`, `REDIS_URL`, and `TOKEN_SECRET` settings. Provider credentials
are optional while providers are set to `disabled`.

Bulk outreach enrolls eligible inferred-business leads only into the currently
approved plain-text sequence. Sending is controlled by the persisted email-job
switch in the admin outreach portal; keep that switch disabled until sequence
previews and recipient counts are reviewed. Redirected/dry-run deliveries remain
at the same sequence step and are recorded as skipped instead of being marked
emailed. The saved database template is the sole source of any unsubscribe copy
or link; the application does not inject or validate an unsubscribe merge tag,
serve an unsubscribe route, or consult the legacy suppression table.

The optional first-email `{{greeting01}}` merge field is deterministic and uses
no AI. It may appear exactly once in the first enabled body, never in a subject
or follow-up. City, cuisine, rating, and review count are used only for profiles
with a Google place id and a successful scrape; unsafe or missing facts select
generic wording without blocking delivery. Existing `{{greeting}}` templates
remain compatible. Migration `000047` creates only an inactive draft for human
review and does not activate a sequence or enable the email job.

Key environment variables:

| Variable               | Purpose                                      |
| ---------------------- | -------------------------------------------- |
| `APP_ENV`              | `local`, `test`, `staging`, or `production`  |
| `HTTP_ADDR`            | API listen address (default `:8080`)         |
| `CORS_ALLOWED_ORIGINS` | Comma-separated allowed origins              |
| `DATABASE_URL`         | PostgreSQL connection string                 |
| `TOKEN_SECRET`         | JWT signing secret (min 32 chars in prod/staging) |
| `JWT_ACCESS_TOKEN_TTL` | Access token lifetime (default `24h`)        |
| `DEMO_TOKEN_TTL`       | Opaque demo-link token expiry (default `720h`) |
| `TUVI_API_TOKEN`        | Bearer token for Tuvi company consultation endpoints |

See `.env.example` for the full list.

## Tuvi corporate website

The canonical marketing, SEO-report, legal, consultation-booking, and corporate
voice site lives in root `web/` (it is not the personalized restaurant demo
renderer):

```bash
npm --prefix web install
npm --prefix web run dev -- -p 3001
```

It runs at **http://localhost:3001** so it can run alongside the restaurant
template on **http://localhost:3000**. See [web/README.md](web/README.md).

The corporate website proxies consultation requests through same-origin
`/api/consultations/*` handlers to the main API on **http://localhost:8080**.
PostgreSQL is the booking and availability authority; this flow does not query
or update Google Calendar.

## Documentation

- [Service inventory](docs/SERVICES.md) — ports, start commands, one-shot jobs, and service interlinks
- [Today's work log (2026-06-22)](docs/work-log/2026-06-22-backend-foundation-session.md) — detailed session notes
- [Phase 1 technical backlog](PHASE1_TECHNICAL_BACKLOG.md) — tickets and acceptance criteria
- [AGENTS.md](AGENTS.md) — coding-agent operating contract
