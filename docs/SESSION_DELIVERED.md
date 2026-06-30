# Session Delivered Log

This is the detailed running delivery log for coding-agent sessions. Update this
before the final response of every implementation or planning session.

Each entry should explain:

- what was delivered;
- why it was delivered;
- tests or checks run;
- business value;
- how the work fits with the rest of the Phase 1 or Phase 2 plan;
- risks, gaps, or follow-ups.

## 2026-06-24 — User Profile API (`GET /api/v1/user/me`)

**Role:** Backend Agent

**Delivered:** `GET /api/v1/user/me` for `restaurant_owner` and `developer` roles — returns full user profile (id, email, full_name, role, is_active, created_at, updated_at) plus linked restaurants with member_role. Parallel to existing `GET /api/v1/admin/me`. OpenAPI updated.

**Why:** Normal users need a dedicated profile endpoint with complete account details, not just JWT claims from `/api/v1/auth/me`.

**Tests / Checks Run:** `go test ./backend/...` — all passing

**Follow-ups:** Optional PATCH `/api/v1/user/me` for profile updates

## 2026-06-24 — Scraped Restaurant Data Schema and Import

**Role:** Backend Agent

**Delivered:** Migration `000008_restaurant_profiles_menus` with `restaurant_profiles`, `menus`, `menu_items`, `restaurant_reviews`, and `restaurant_data_imports` tables; `seed-restaurants-data` command and `make seed-restaurants-data` target; imported `data/restaurants_data.json` into PostgreSQL (8 unique restaurants, 3 duplicate `google_place_id` rows skipped).

**Why:** Scraped SerpAPI restaurant payloads (menus, reviews, contact, location, images) need durable storage before P1-011 profile APIs and demo payload builder can consume them.

**Business Value:** Sales and demo workflows can now query real scraped restaurant profiles, menus, and reviews locally instead of relying on the raw JSON file.

**Plan Fit:** Implements the P1-011 database shape (profiles, menus, menu items) plus review storage; unblocks profile CRUD APIs and demo payload builder polish.

**Tests / Checks Run:**
- `make migrate-up` — success
- `make seed-restaurants-data` — 8 imported, 3 skipped
- SQL verification — menu item and review counts per restaurant confirmed

**Risks / Follow-ups:**
- No profile/menu HTTP APIs yet; data is import-only
- Re-import is idempotent by `google_place_id` but replaces menu items and reviews
- Add profile API routes and link demo builder to imported data

## 2026-06-17 — P1-E01 Backend Foundation

**Role:** Backend Agent

**Delivered:** Initialized the repository as a Go module and added the Phase 1
backend foundation: API, worker, and migration commands; typed environment
configuration; structured logging; PostgreSQL pool wiring with `pgxpool`;
health/readiness endpoints; request ID, access logging, recovery, and CORS
middleware; initial SQL migration; internal migration runner; in-memory job
queue; tests; README; Makefile; `.env.example`; and a foundation-stack ADR.

**Why:** P1-E01 is the required platform base for all Phase 1 product work. The
project needed a runnable backend, local command surface, safe config behavior,
database readiness path, and worker abstraction before auth, restaurant CRUD,
demo generation, reservations, outreach, analytics, AI receptionist, or content
automation can be implemented.

**Business Value:** This turns the repo from planning-only into a runnable
engineering foundation. It reduces delivery risk by giving future work a
standard way to start services, validate configuration, run migrations, expose
health checks, and add background jobs. That directly supports the sales MVP by
making the lead-to-demo-to-reservation loop buildable in small, testable steps.

**Plan Fit:** This completes the first Phase 1 build-order item: Go backend
foundation, config, logging, migrations, and health check. It unblocks P1-E02
auth/roles/tenant safety and P1-E03 restaurant/profile/menu CRUD.

**Checks Run:** `go test ./backend/...`, `make test`, API smoke check for
`/healthz` and `/readyz`, and worker smoke check for sample job processing.

**Risks / Follow-ups:** `make migrate-up` was not run against a real PostgreSQL
database in-session. Durable queue, auth/session provider, and DB repository
pattern choices remain to be finalized as later epics need them.

## 2026-06-17 — Session Delivery Documentation Contract

**Role:** Documentation Agent

**Delivered:** Updated `AGENTS.md` so future sessions read the detailed delivery
log during orientation, overwrite the concise summary doc with 3-5 lines, and
update both session docs before the final response. Added this detailed
delivery log and the concise summary doc.

**Why:** Raj requested a persistent handoff mechanism so each session records
exactly what was delivered, why it matters, its business value, and how it fits
with the larger plan.

**Business Value:** This creates continuity between coding sessions. Future
agents can quickly understand what already exists, why it was built, and what
work it unlocks without re-reading every implementation detail or losing the
sales-MVP context.

**Plan Fit:** This strengthens the operating process around Phase 1 and Phase 2
delivery. It does not change product behavior, but it improves execution
discipline and handoff quality for every subsequent epic.

**Checks Run:** Documentation-only change; no code tests were required.

**Risks / Follow-ups:** Future agents must keep `docs/SESSION_SUMMARY.md`
concise and overwritten, while keeping this file as the detailed running log.

## 2026-06-22 — P1-E01 Foundation Completion + Auth Prep

**Role:** Backend Agent

**Delivered:** Completed P1-001 through P1-005 and early P1-E02 prep. Fiber API
with graceful shutdown; strict typed config (`env.go`); Docker Compose for
PostgreSQL 16; migrations 000001 (foundation) and 000002 (users); repository
pattern for metadata and users with Postgres + Mock implementations; `store`
package with startup verification; required DB connect before API/worker start;
`APP_ROLE` gating on health endpoints; aligned `.env.example` and `backend/.env`;
work log at `docs/work-log/2026-06-22-backend-foundation-session.md`; README
command updates.

**Why:** The sales MVP needs a reliable, testable backend base before auth,
restaurant CRUD, demos, reservations, outreach, and AI workflows can be built.
Each P1-E01 ticket unblocks the next layer of product work.

**Business Value:** Developers can run `make dev` and get a working API against
real PostgreSQL with migrations, health checks, structured logging, and a clean
path to add domain APIs. Reduces integration risk for the lead-to-demo loop.

**Plan Fit:** Finishes Phase 1 build-order items 1–5 (foundation through HTTP
layer). Unblocks P1-007 login, P1-008 auth middleware, and P1-E03 restaurant
CRUD.

**Checks Run:** `make test`, `go test ./backend/...`, `make migrate-up`,
`make api` smoke with `/healthz` and `/readyz`.

**Risks / Follow-ups:** No Postgres integration tests for repositories yet
(mocks only). P1-006 durable jobs not wired to `job_runs` table. Restrict
signup roles in staging/production. Health endpoints behind JWT may need
separate probe path for load balancers later.

## 2026-06-22 — JWT Auth + Protected Health Endpoints

**Role:** Backend Agent

**Delivered:** JWT-based signup/login (`POST /api/v1/auth/signup`, `POST /api/v1/auth/login`),
bcrypt password hashing, `TokenManager` with `JWT_ACCESS_TOKEN_TTL`, auth service layer,
`RequireAuth` and `RequireRole` middleware, developer-only `/healthz` and `/readyz`.
Removed `APP_ROLE` env gating and `health_access.go`. Migration `000003` drops
`users_role_check` DB constraint. Updated README, `.env.example`, and `backend/.env`.

**Why:** PM requested roles not be hardcoded in schema; authorization should come from
JWT issued at login with role stored in the database.

**Checks Run:** `go test ./backend/...`

**Follow-ups:** Restrict public signup role selection in production; add restaurant-scoped
RBAC (P1-008/P1-009).

## 2026-06-23 — Melbourne Food Image Scraper (automation/)

**Role:** Backend / Automation Agent

**Delivered:** Python pipeline under `automation/` that scrapes food images for
50 curated Melbourne restaurant dishes (Australian, Thai, Indian, Italian, and
other cuisines). Saves images to `automation/images/` by SHA-256 hash and writes
metadata to `automation/data/catalog.json`.

**Why:** Pre-builds demo menu image assets for P1-013 menus and P1-048 seed data
so sales demos look realistic without manual photo hunting per lead.

**Checks Run:** `python -m py_compile automation/scrape.py`; live scrape test
with `--limit 3` (3 images saved via Bing fallback).

**Business Value:** Reusable food catalog with hash-based image storage for
Melbourne-focused demo restaurants.

**Plan Fit:** Upstream content work for P1-E03 menu items (`image_url`) and
P1-048 seed/demo data.

**Risks / Follow-ups:** Google Images requires a browser and may CAPTCHA; Bing
fallback used when Google times out. Review image licensing before production.
Run full scrape with `python scrape.py --skip-google` for speed or default mode
for Google-first.

## 2026-06-22 — P1-009 Restaurant-Scoped Access Checks

**Role:** Backend Agent

**Delivered:** Hardened restaurant access layer with `RestaurantID` in request
context, extended `access.Service` (list/create/members), reusable
`protectRestaurantScoped` / `protectRestaurantAdmin` middleware, restaurant list
and member APIs, migration `000005_demo_sites`, public demo route
`GET /api/public/v1/demo/{slug}?token=...` with allowlisted `PublicDemoPayload`
DTO, admin `POST /api/v1/restaurants/{id}/demo-sites`, `make seed-demo-fixture`,
Postman collection updates, and table-driven isolation tests.

**Why:** P1-009 requires tenant isolation before restaurant CRUD, profiles, and
demo generator work can safely scale.

**Checks Run:** `go test ./backend/...`

**Business Value:** Restaurant owners cannot read another tenant's data; public
demo links expose only sales-safe payload fields.

**Plan Fit:** Completes P1-009 acceptance criteria; unblocks P1-010 full
restaurant CRUD and P1-015+ demo payload builder.

**Risks / Follow-ups:** Demo token in query string may appear in proxy logs;
P1-016 can move to header/path. Full restaurant lead fields deferred to P1-010.

## 2026-06-22 — Restaurant Lead Fields (sales scope)

**Role:** Backend Agent

**Delivered:** Migration `000006_restaurant_lead_fields` adding `email`,
`is_contacted`, and `shown_interest` to `restaurants`. API create/list/get now
return lead fields plus `created_at`/`updated_at`. Create requires `name` and
`email`. Repository `MarkShownInterest` added for future email click tracking.
No image or profile fields (deferred per scope).

**Why:** Aligns `restaurants` table with sales MVP lead tracking described in
the implementation guide and product scope: outreach email → link click →
`shown_interest = true`.

**Checks Run:** `go test ./backend/...`

**Follow-ups:** Wire `MarkShownInterest` from campaign/email click tracking;
admin PATCH to set `is_contacted` when outreach is sent.

## 2026-06-22 — P1-010 Restaurant CRUD + Query Filters

**Role:** Backend Agent

**Delivered:** Completed P1-010 with migration `000007_restaurant_status`, full
restaurant lifecycle `status`, `PATCH /api/v1/restaurants/{id}`,
`PATCH /api/v1/restaurants/{id}/status`, soft archive via
`DELETE /api/v1/restaurants/{id}`, and list filters via query params
(`?restaurant=`, `?status=`, `?is_contacted=`, `?shown_interest=`,
`?include_archived=`). Auto-status rules: `is_contacted` moves lead to
`emailed`; `shown_interest` moves to `interested`.

**Checks Run:** `go test ./backend/...`

**Plan Fit:** P1-010 acceptance criteria met; unblocks P1-011 profiles and demo
payload builder.

## 2026-06-22 — Repository Structure Alignment (Phase 1 Guide §5)

**Role:** Backend Agent

**Delivered:** Restructured `backend/internal` from layered `repositories/` +
`services/` into domain packages aligned with
`docs/phase1/PHASE1_IMPLEMENTATION_GUIDE.md`: `restaurants/` (repo + access
service), `demos/` (repo + demo service + token helpers), `auth/` (JWT, user
repo, signup/login service). Moved shared errors to `platform/errors/` and
metadata repo to `platform/metadata/`. Added placeholder `doc.go` packages for
future domains (`menus`, `reservations`, `campaigns`, `analytics`, `ai/*`,
`providers/*`, `platform/telemetry`, `backend/tests`). Moved
`PHASE1_IMPLEMENTATION_GUIDE.md` and `PHASE1_TECHNICAL_BACKLOG.md` to
`docs/phase1/`.

**Why:** The implementation guide defines domain-owned packages so Phase 1
modules can grow without cross-layer import sprawl and match the documented
monolith boundaries.

**Checks Run:** `go test ./backend/...` — pass

**Risks / Follow-ups:** No API behavior change intended; imports only. Add ADR
if we later split `auth` user persistence from JWT helpers.

## 2026-06-23 — Postman Full Flow Update

**Role:** Documentation Agent

**Delivered:** Rebuilt `postman/Restaurant-Platform.postman_collection.json`
with ordered **01 — PM Demo Flow** folder (15 steps), fixed role names
(`internal_admin`, `restaurant_owner`), added Auth Me, List Members, filter
examples, owner/admin token switching, and updated environment + README.

**Checks Run:** JSON validation on collection and environment files.

**Plan Fit:** Supports PM demo of P1-008 through P1-010 without manual curl.

## 2026-06-23 — OpenAPI / Swagger Documentation

**Role:** Documentation Agent

**Delivered:** OpenAPI 3.0 spec at `docs/openapi/openapi.yaml` covering all 17
endpoints (auth, admin, restaurants, members, demo sites, public demo, health).
Added `docs/openapi/README.md`, `make openapi` validation target, and README/Postman links.

**Checks Run:** `make openapi` — pass

