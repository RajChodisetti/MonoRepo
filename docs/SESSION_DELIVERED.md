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

## 2026-07-08 — VM Deployment Plan Draft

**Role:** DevOps / Documentation Agent

**Delivered:** Added `docs/runbooks/vm-deployment-plan.md`, a VM deployment
plan based on the repo's current runnable services and Compose setup. The plan
covers the main Go API, worker, migrations, PostgreSQL, voice agent, Redis,
restaurant services catalog, optional Tuvi Next site, restaurant demo template,
automation jobs, reverse proxy/TLS, env files, backups, rollback, smoke checks,
and open decisions.

**Why:** Raj asked for a complete plan to deploy all services on a VM and asked
that the existing setup be considered before planning.

**Business Value:** Creates a concrete deployment path from local Phase 1
services to one VM, while making clear which pieces are already containerized
and which pieces still need VM Compose/proxy work.

**Plan Fit:** Supports Phase 1 deployment/runbook work by turning the existing
Docker stack into a fuller VM production plan with no intentional service gaps.

**Tests / Checks Run:** Inspection only. Reviewed `AGENTS.md`, `git status`,
`README.md`, `Makefile`, `docs/SERVICES.md`, Phase 1 docs, Docker Compose,
backend Dockerfile/config, stack env example, startup scripts, Tuvi website
runtime docs, restaurant catalog docs, template env, and voice agent Docker/env
docs. Checked local SSH config for a VM alias; none was configured.

**Risks / Follow-ups:** Live VM audit is still pending because this workstation
does not have a usable VM SSH alias in `~/.ssh/config`. The plan includes the
exact audit commands to run once the VM host/deploy user is known.

## 2026-07-08 — Local Catalog Work Rebased on Phase Branch

**Role:** Frontend / Documentation Agent

**Delivered:** Preserved Raj's local `apps/restaurant-services-catalog` work by
committing it locally, creating a safety branch at `local/pre-pull-catalog-work`,
and rebasing the local `phase1_03/backend` branch on top of the latest fetched
`origin/phase1_03/backend`. Resolved the pull-blocking conflicts in the catalog
package files and session summary without replacing the video-enabled restaurant
services catalog.

**Why:** A direct `git pull` was blocked because local session docs and
untracked catalog files would have been overwritten by the incoming branch.

**Business Value:** The active phase branch now has the latest backend work plus
the restored restaurant services catalog assets, including the FAL-generated QR
ordering and rewards reception videos needed for the sales presentation flow.

**Plan Fit:** Keeps the Phase 1 restaurant services marketing surface aligned
with the backend branch while preserving local catalog iteration work for
outreach and demo conversations.

**Tests / Checks Run:**
- `rtk git fetch origin phase1_03/backend` — fetched latest remote branch state
- `rtk git rebase origin/phase1_03/backend` — completed after conflict
  resolution
- `rtk make restaurant-services-catalog-build` — Vite production build passed
- `rtk make test` — backend Go test suite passed
- `rtk rg -n 'qr-ordering-kitchen-v2|rewards-reception-v3-pro' apps/restaurant-services-catalog/src apps/restaurant-services-catalog/dist/index.html`
  — built catalog references both approved video files
- `rtk ls -lh apps/restaurant-services-catalog/public/media/*.mp4` — confirmed
  the current and fallback MP4 assets are present
- `rtk curl -I http://127.0.0.1:5174` — catalog dev server returned `200 OK`
- `rtk curl -s http://127.0.0.1:5174 | rtk rg -n 'qr-ordering-kitchen-v2.mp4|rewards-reception-v3-pro.mp4'`
  — served page references both approved video files
- `rtk curl -I http://127.0.0.1:5174/media/qr-ordering-kitchen-v2.mp4`
  — returned `200 OK` with `video/mp4`
- `rtk curl -I http://127.0.0.1:5174/media/rewards-reception-v3-pro.mp4`
  — returned `200 OK` with `video/mp4`

**Risks / Follow-ups:** The local commit is intentionally still ahead of
`origin/phase1_03/backend` by one commit and has not been pushed. Permanent
Cloudflare Pages deployment still requires Wrangler authentication.

## 2026-07-07 — Phase Branch Catalog Videos Preserved

**Role:** Frontend / Documentation Agent

**Delivered:** Pulled `origin/phase1_03/backend` into a local
`phase1_03/backend` branch and applied the safe local catalog changes on top:
root Makefile shortcuts for the restaurant services catalog, catalog README, and
public `.env.example`. Preserved the branch's FAL-generated restaurant feature
videos (`qr-ordering-kitchen-v2.mp4` and `rewards-reception-v3-pro.mp4`) and the
video-enabled catalog page.

**Why:** Raj asked to apply local changes on top of `phase1_03/backend` and
clarified that the target website is the catalog version with two FAL-generated
videos.

**Business Value:** The current branch now has the latest backend work and the
video-enabled restaurant services site, with simple root commands for local
startup and build verification.

**Plan Fit:** Keeps the Phase 1 restaurant services marketing surface aligned
with the active backend branch.

**Tests / Checks Run:**
- `rtk npm --prefix apps/restaurant-services-catalog install` — dependencies current; no vulnerabilities reported
- `rtk make restaurant-services-catalog-build` — initially failed on broken `lucide@0.547.1`; passed after pinning `lucide` to `0.547.0`
- `rtk make test` — backend tests passed
- `rtk curl -I http://127.0.0.1:5173` — catalog dev server returned `200 OK`
- `rtk curl -s http://127.0.0.1:5173 | rtk rg -n "qr-ordering-kitchen-v2.mp4|rewards-reception-v3-pro.mp4"` — served page references both FAL videos
- `rtk curl -I http://127.0.0.1:5173/media/qr-ordering-kitchen-v2.mp4` — returned `200 OK` with `video/mp4`
- `rtk curl -I http://127.0.0.1:5173/media/rewards-reception-v3-pro.mp4` — returned `200 OK` with `video/mp4`

**Risks / Follow-ups:** The local changes from the older restored catalog app
were not applied wholesale because they would replace the newer video-enabled
catalog implementation. `lucide` is pinned to `0.547.0` because `0.547.1`
installed locally without matching icon module files and broke the Vite build.

## 2026-07-06 — Tuvi Scheduler and Voice Agent Unified API Alignment

**Role:** Backend / Frontend / AI Workflow Agent

**Delivered:** Merged the remote unified Tuvi company consultation backend and
rewired the Tuvi corporate meeting scheduler and corporate voice assistant to use
the same main MonoRepo API:
`GET /api/v1/company/consultations/availability`,
`GET /api/v1/company/consultations/availability/check`, and
`POST /api/v1/company/consultations`. The scheduler now collects phone and sends
`web` source through a server-side Next proxy that keeps `TUVI_API_TOKEN` out of
the browser. The corporate voice assistant now asks for phone before booking and
sends `voice` source directly to the same unified API. The company consultation
success response now includes `prospect_phone`.

**Why:** Raj merged the Tuvi app into the main app and asked that all relevant
apps, including meeting scheduler and voice agent, call the new main endpoint.

**Business Value:** Website and voice-assisted consultation bookings now land in
one Phase 1 company consultation pipeline, reducing split-brain booking state and
making future dashboards, analytics, and follow-up workflows simpler.

**Plan Fit:** Advances P1-E05 reservation capture and P1-E08 voice assistant
integration by routing Tuvi web and voice consultation requests through one main
backend contract.

**Tests / Checks Run:**
- `rtk make test` — backend tests passed
- `rtk python3 -m py_compile voice-sales-agent/tuvi_api_client.py voice-sales-agent/bot.py` — passed
- `rtk npx tsc --noEmit` from `tuvi-website/app` — passed
- `rtk npx -p node@22 node ./node_modules/next/dist/bin/next build` from
  `tuvi-website/app` — passed after clearing stale `.next` output
- `rtk npm run build` from `tuvi-website/app` — passed

**Risks / Follow-ups:** Set the same `TUVI_API_TOKEN` in the main API, Tuvi
website server env, and corporate voice-agent env before trying bookings. The
unified API supports Google Calendar and SMTP through main backend configuration;
local defaults keep those providers disabled.

## 2026-06-30 — RTK Codex Initialization

**Role:** DevOps / Documentation Agent

**Delivered:** Installed and initialized RTK for Codex globally and locally.
Global Codex config now includes `/Users/rajchodisetti/.codex/RTK.md`, and this
repo now includes `RTK.md` with an `@RTK.md` reference from `AGENTS.md`.

**Why:** Raj requested RTK initialization, especially for Codex integrations, so
future Codex sessions get the token-optimized command instruction automatically.

**Business Value:** Reduces command-output context usage during future coding
sessions while preserving normal command behavior.

**Plan Fit:** Supports the local development tooling rules in `AGENTS.md` and
keeps agent command usage consistent across repo-local and global Codex contexts.

**Tests / Checks Run:**
- `rtk init -g --codex` — configured global Codex RTK files
- `rtk init --codex` — configured repo-local Codex RTK files
- `rtk init --codex --show` — verified global and local Codex RTK config
- `rtk read RTK.md` — verified repo-local RTK instructions

**Risks / Follow-ups:** Generic `rtk init --show` reports non-Codex hook status;
use `rtk init --codex --show` to verify Codex integration specifically.

## 2026-06-30 — Service and Workflow Inventory

**Role:** Planner / Documentation Agent

**Delivered:** Inspected the repo command surface, Phase 1/Phase 2 guides, Makefile,
Docker Compose, current backend packages, template frontend, and automation
README files to produce a business-prioritized service inventory and trigger
workflow map.

**Why:** Raj requested a clear view of which services exist or are planned, how
to start them locally, and which business workflows trigger downstream services.

**Business Value:** Clarifies the shortest path to running the sales MVP locally:
database, migrations, API, worker, seed data, demo template, and optional
automation. It also separates currently implemented services from planned Phase
1/Phase 2 services so execution can stay focused.

**Plan Fit:** Supports Phase 1 lead-to-demo-to-reservation sequencing and
identifies the Phase 2 orchestration services as future work after the Phase 1
sales loop is working.

**Tests / Checks Run:** Inspection only. Ran targeted reads of `AGENTS.md`,
`Makefile`, `README.md`, Phase 1/Phase 2 docs, Docker Compose, backend router,
app startup, worker startup, config, template README/package, and automation
README files. No code tests were run because no product code changed.

**Risks / Follow-ups:** `apps/web` is still a placeholder; the runnable demo
frontend currently lives under `template/`. Phase 2 docs are under `phase2/`
rather than `docs/phase2/`.

## 2026-06-24 — User Profile API (`GET /api/v1/user/me`)

**Role:** Backend Agent

**Delivered:** `GET /api/v1/user/me` for `restaurant_owner` and `developer` roles — returns full user profile (id, email, full_name, role, is_active, created_at, updated_at) plus linked restaurants with member_role. Parallel to existing `GET /api/v1/admin/me`. OpenAPI updated.

**Why:** Normal users need a dedicated profile endpoint with complete account details, not just JWT claims from `/api/v1/auth/me`.

**Tests / Checks Run:** `go test ./backend/...` — all passing

**Follow-ups:** Optional PATCH `/api/v1/user/me` for profile updates

## 2026-07-03 — Restaurant Services Catalog App

**Role:** Frontend Agent

**Delivered:** Added a new standalone Vite JavaScript app at
`apps/restaurant-services-catalog` for an interactive Tuvi restaurant services
catalog. The app uses the existing `presentation/index.html` service story as
content grounding, adds a premium visual system, scroll-driven motion sections,
interactive service selection, QR/rewards/voice/service modules, an ROI
estimator, Cloudflare Pages config, and locally stored fal-generated media
assets.

**Why:** Tuvi needs a polished services catalog that can be shared with
restaurant owners after demo outreach, showing the full platform package beyond
one personalized demo website.

**Business Value:** The catalog turns the service offering into a sales asset:
owners can see demo websites, QR ordering, rewards, AI receptionist,
reservations, outreach, content automation, and token-gated demos in one guided
experience.

**Plan Fit:** Supports Phase 1 sales positioning and demo/outreach workflows by
giving campaigns a richer Tuvi services destination after the restaurant-specific
demo link.

**Checks Run:**
- `npm install` in `apps/restaurant-services-catalog` — success, 0 vulnerabilities.
- `npm run build` — success.
- `curl` smoke checks for local and Cloudflare tunnel page/media assets — 200.
- Playwright screenshots for desktop/mobile hero and services sections.

**Risks / Follow-ups:** Cloudflare Pages direct deploy was blocked because
Wrangler is not authenticated in this shell. The app is currently hosted via a
Cloudflare quick tunnel backed by a local static server. Run
`npm run deploy` after `wrangler login` or with a Cloudflare API token to publish
to Pages.

## 2026-07-03 — Restaurant Services Catalog Polish

**Role:** Frontend Agent

**Delivered:** Polished the hosted restaurant services catalog by removing
visible AI-generation/provider labels from the UI, replacing the hero background
video with pointer/touch-reactive canvas motion graphics, and tightening the
desktop/mobile visual presentation after screenshot review.

**Why:** The catalog should read as a premium Tuvi-owned sales experience, not
as an asset-generation demo. The hero needed a lighter, interactive motion layer
without relying on a background video.

**Business Value:** Restaurant owners see a cleaner services story with fewer
technical labels and a more polished first impression, improving the catalog's
fit as an outreach and sales destination.

**Plan Fit:** Supports Phase 1 demo/outreach by improving the public-facing
services destination already hosted through the Cloudflare quick tunnel.

**Checks Run:**
- `npm run build` in `apps/restaurant-services-catalog` — success.
- Source/dist scan confirmed the removed AI-generation/provider phrases are no
  longer visible in the app markup.
- Cloudflare tunnel smoke check for the hosted catalog — 200.
- Playwright screenshot review for desktop hero, mobile hero, and media section.

**Risks / Follow-ups:** The Cloudflare Pages permanent deploy remains blocked
until Wrangler is authenticated. The current hosted preview depends on the local
static server and Cloudflare quick tunnel staying up.

## 2026-07-03 — Restaurant Services Catalog Offer Revision

**Role:** Frontend Agent

**Delivered:** Updated the hosted catalog to focus only on restaurant-facing
Tuvi services. Renamed "Premium Demo Websites" to "Custom Premium Websites",
removed the scroll-driven "invisible lead to booked sales call" transformation
block, removed internal sales-pipeline/token-gated demo language from the
service catalog, and replaced the obsolete single generated loop with two new
fal-generated cinematic service-flow videos: QR table ordering routed to the
kitchen, and reception ordering with QR rewards. Follow-up cleanup removed
client-facing production/style language such as "cinematic product moments",
changed the section to "Guest Experience", changed the nav label from "Media" to
"Experience", neutralized public media filenames, and removed the public media
manifest so provider/generation details are not exposed in the hosted page. A
later client-facing cleanup removed the value estimator, removed the two extra
image tiles from the QR/rewards section, simplified the top-left Tuvi wordmark,
removed vague catalog/guest-experience phrasing, tightened hero/section spacing,
and upgraded service cards with a glass treatment plus color highlight on hover,
focus, touch, and selected state.

**Why:** Raj clarified that the catalog should sell what Tuvi can do for
restaurants, not describe Tuvi's internal lead-acquisition and sales process.

**Business Value:** Restaurant owners now see concrete guest-facing outcomes:
custom websites, QR ordering, rewards, voice agents, reservations, promotions,
menu/photo automation, and owner dashboards. The new videos make the QR ordering
and rewards flows easier to understand quickly.

**Plan Fit:** Improves the Phase 1 services destination that can be linked from
demo-site conversations and used as a Tuvi walkthrough for restaurant owners.

**Checks Run:**
- `npm run build` in `apps/restaurant-services-catalog` — success.
- Source/dist phrase scan confirmed removed internal labels and old video
  references are gone.
- Hosted page-source scan confirmed provider/generation/prompt/style wording is
  not present in client-facing HTML.
- Source/dist scan confirmed estimator code, estimator markup, old vague labels,
  and unused image references are gone.
- Cloudflare tunnel smoke checks for the hosted page and both new MP4 assets —
  200; old generated/provider-named asset paths — 404.
- Playwright screenshots for desktop hero, mobile hero, services section, and
  QR/rewards media section.

**Risks / Follow-ups:** Permanent Cloudflare Pages deploy still needs Wrangler
authentication. The current hosted URL depends on the local static server and
Cloudflare quick tunnel staying online.

## 2026-07-03 — Restaurant Services Catalog Tuvi Solutions Update

**Role:** Frontend Agent

**Delivered:** Updated the hosted catalog to use "Tuvi Solutions" consistently
instead of standalone "Tuvi", replaced the services grid with a sticky parallax
service-card stack, and added animated icon/symbol treatments to every service
card using the existing Lucide icon system plus CSS motion. Cleaned the services
section copy so it sells restaurant outcomes instead of explaining the scroll
interaction. Tightened the landing hero first fold from mobile screenshot
feedback by making the hero height content-driven, brightening the restaurant
background layer, softening the dark overlay, adding topbar separation, hiding
the mobile scroll cue, and trimming the lower hero gap.
Added viewport-managed video loading/playback so large MP4 walkthroughs use
poster frames initially, load only when near view, pause when offscreen, and
resume only when visible to reduce mobile buffering and stutter through the
Cloudflare quick tunnel.

**Video Work:** After fal billing was recharged, generated and downloaded two
new service videos without deleting the previous files. The approved QR ordering
video remains `qr-ordering-kitchen-v2.mp4`. The rewards video was regenerated
again with Kling V3 Pro using a tighter script focused on scan, order
confirmation, points added, and redeem options, then wired as
`rewards-reception-v3-pro.mp4` with `rewards-reception-v3-pro-poster.png`. Older
rewards files remain in `public/media` for fallback/reference.

**Why:** Raj asked the catalog to present the company as Tuvi Solutions, make the
services area feel more premium and progressive, and produce clearer service-flow
videos without removing the current assets.

**Business Value:** Restaurant owners now see one premium service module at a
time with clearer visual emphasis, plus clearer service-flow videos for table QR
ordering to kitchen operations and counter QR rewards.

**Plan Fit:** Keeps the hosted Tuvi Solutions sales destination aligned with the
demo/outreach path while preserving reusable media assets for future iteration.

**Checks Run:**
- `npm run build` in `apps/restaurant-services-catalog` — success.
- Rebuilt the live-served catalog `dist` and confirmed the Cloudflare tunnel
  serves the updated CSS bundle with the brighter background, mobile topbar, and
  tighter hero spacing rules.
- Confirmed the live tunnel serves deferred video tags (`data-src`,
  `preload="none"`) and the JS bundle contains the managed video playback logic.
- Source scan confirmed no standalone "Tuvi" remains in the catalog source.
- Source scan confirmed removed internal labels such as AI-generated wording,
  prompts, estimator text, and scroll-driven/catalog labels remain absent.
- Local and Cloudflare tunnel smoke checks for page, `qr-ordering-kitchen-v2.mp4`,
  `rewards-reception-v3-pro.mp4`, and `rewards-reception-v3-pro-poster.png`
  returned 200.
- Playwright screenshot confirmed the latest walkthrough labels, V2 posters, and
  captions render cleanly in the media section.

**Risks / Follow-ups:** The first pre-recharge ordering request stayed queued at
position 0, so a fresh ordering retry was used for the live V2 file. Permanent
Cloudflare Pages deploy still requires Wrangler authentication.

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
