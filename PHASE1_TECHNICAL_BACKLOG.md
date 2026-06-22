# Phase 1 Technical Backlog — Restaurant Sales MVP

**Audience:** coding agents and developers.  
**Use with:** `PHASE1_IMPLEMENTATION_GUIDE.md`.  
**Execution style:** implement in priority order unless a human assigns otherwise. Each story includes acceptance criteria that must be satisfied before marking done.

---

## 1. Epic Summary


| Epic   | Name                                     | Priority | Main Owner    |
| ------ | ---------------------------------------- | -------- | ------------- |
| P1-E01 | Backend Foundation and Infrastructure    | Critical | Dev 1         |
| P1-E02 | Auth, Roles, and Tenant Safety           | Critical | Dev 1         |
| P1-E03 | Restaurant, Profile, and Menu Management | Critical | Dev 1         |
| P1-E04 | Demo Website Generator                   | Critical | Dev 1         |
| P1-E05 | Reservation System                       | Critical | Dev 1         |
| P1-E06 | Email Outreach and Tracking              | Critical | Dev 2         |
| P1-E07 | Analytics and Funnel Metrics             | High     | Dev 2         |
| P1-E08 | AI Receptionist MVP                      | High     | Dev 2         |
| P1-E09 | Content Automation MVP                   | Medium   | Dev 2         |
| P1-E10 | Dashboards, QA, Deployment, and Runbook  | High     | Dev 1 + Dev 2 |


---

## 2. Global Acceptance Criteria

Every ticket must satisfy these unless explicitly not applicable:

- Code is committed in the expected package/module.
- API inputs are validated.
- Errors are safe and actionable.
- Tests cover happy path and at least one failure path.
- Database changes include migrations.
- Restaurant-scoped operations enforce tenant/access checks.
- Logs include correlation/request ID where available.
- No secrets are committed.
- Feature works locally.
- Feature is documented through API examples, comments, or this doc when behavior is non-obvious.

---

## P1-E01 — Backend Foundation and Infrastructure

### P1-001 — Initialize repository and Go services

**Story:** As a developer, I need a clean repo layout so that API, worker, frontend, and docs can evolve without confusion.

**Implementation notes:**

- Create `backend/cmd/api` and `backend/cmd/worker`.
- Add `internal/app`, `internal/http`, `internal/platform/config`, `internal/platform/logger`.
- Add `apps/web` placeholder if frontend is in same repo.
- Add Makefile or task runner.

**Acceptance criteria:**

- `make api` or equivalent starts the API locally.
- `make worker` or equivalent starts the worker locally.
- `make test` runs backend tests.
- README contains local setup steps.
- `.env.example` exists without real secrets.

### P1-002 — Add typed configuration and startup validation

**Story:** As a developer, I need typed environment config so that missing secrets or invalid configuration fail fast.

**Acceptance criteria:**

- Config struct includes app, database, Redis, email, LLM, voice, storage, token, and logging settings.
- Required variables fail at startup with clear messages.
- Local defaults are safe.
- Production requires explicit secrets.

### P1-003 — Set up PostgreSQL migrations

**Story:** As a developer, I need repeatable database migrations so the schema can be recreated in local, staging, and production.

**Acceptance criteria:**

- Migration tool is configured.
- Initial migration creates core tables or a minimal migration baseline.
- Migrations can run locally from a documented command.
- Rollback strategy is documented, even if rollback scripts are simple.

### P1-004 — Add database access layer

**Story:** As a developer, I need a reliable database layer so domain repositories can query PostgreSQL safely.

**Acceptance criteria:**   

-  DB connection pool is initialized once.
- Health check verifies database connection.
- Repository pattern or generated SQL pattern is established.
- Tests can run with a test database or mocked repository.

### P1-005 — Add API router, middleware, and health endpoints

**Story:** As a developer, I need a standard HTTP foundation for API handlers.

**Acceptance criteria:**

- `/healthz` returns service health.
- `/readyz` checks database readiness.
- Request ID middleware exists.
- Structured access logging exists.
- Panic recovery middleware returns safe 500 response.
- CORS config is environment-aware.

### P1-006 — Add async job abstraction

**Story:** As a developer, I need a job interface so email, follow-up, content, and summary jobs can run outside request paths.

**Acceptance criteria:**

- Job enqueue interface exists.
- Worker can register handlers.
- At least one sample job can be enqueued and processed.
- Job retries are supported or documented.
- Job payloads include idempotency keys where needed.

---

## P1-E02 — Auth, Roles, and Tenant Safety

### P1-007 — Implement admin authentication

**Story:** As an internal admin, I need to log in securely before managing leads and demos.

**Acceptance criteria:**

- Admin user table/model exists.
- Login endpoint verifies credentials.
- Session or JWT is issued.
- Protected routes reject unauthenticated requests.
- Failed login does not reveal which field is wrong.

### P1-008 — Implement role model

**Story:** As the platform, I need roles so internal admins and restaurant owners have different permissions.

**Acceptance criteria:**

- User roles include `internal_admin` and `restaurant_owner` at minimum.
- Auth middleware exposes user ID and role.
- Role checks are reusable.
- Tests prove restaurant owner cannot access another restaurant.

### P1-009 — Add restaurant-scoped access checks

**Story:** As a system owner, I need restaurant data isolated so one restaurant cannot access another restaurant's data.

**Acceptance criteria:**

- All restaurant-scoped APIs check access.
- Internal admin can access all restaurants.
- Restaurant owner can access assigned restaurants only.
- Public demo routes expose only public demo payload data.

---

## P1-E03 — Restaurant, Profile, and Menu Management

### P1-010 — Create restaurant schema and CRUD API

**Story:** As an admin, I need to create restaurant leads so demos can be generated.

**Acceptance criteria:**

- `restaurants` table exists.
- API supports create/list/get/update/archive.
- Required fields are validated.
- Status field supports lead lifecycle.
- List endpoint supports search and status filter.

### P1-011 — Create restaurant profile schema and API

**Story:** As an admin, I need to store detailed restaurant info for demos, AI answers, and content generation.

**Acceptance criteria:**

- `restaurant_profiles` table exists.
- Profile API supports get/upsert.
- Opening hours can be stored as JSON.
- Raw public data can be stored separately from approved public payload.
- Profile has `review_status`.

### P1-012 — Add profile human review workflow

**Story:** As an admin, I need to approve restaurant data before emailing a demo link.

**Acceptance criteria:**

- Admin can mark profile as reviewed.
- Demo/email send is blocked or warned if profile is unreviewed.
- Review timestamp and reviewer ID are stored.

### P1-013 — Create menu and menu item APIs

**Story:** As an admin, I need to add menu items so personalized demos look real.

**Acceptance criteria:**

- `menus` and `menu_items` tables exist.
- API supports create/list/update/delete menu items.
- Items support name, description, price, category, image URL, availability.
- Menu items can be returned grouped by category.

### P1-014 — Create admin UI for restaurant/profile/menu

**Story:** As an admin, I need a dashboard form to manage restaurant details without editing the database.

**Acceptance criteria:**

- Admin can create a restaurant.
- Admin can edit profile fields.
- Admin can add/edit/delete menu items.
- Form validation errors are visible.
- Saved data appears in detail page without refresh issues.

---

## P1-E04 — Demo Website Generator

### P1-015 — Build demo payload builder

**Story:** As the system, I need a stable demo payload so templates can render restaurant-specific sites.

**Acceptance criteria:**

- Builder reads restaurant, profile, menu, and template config.
- Payload includes hero copy, cuisine, menu, hours, address, phone, reservation CTA, AI receptionist CTA, content automation CTA.
- Payload excludes internal notes and raw private metadata.
- Payload is stored as JSON snapshot in `demo_sites.payload`.
- Unit tests cover missing menu/hours/images.

### P1-016 — Implement slug and signed token generation

**Story:** As a system owner, I need safe demo links that do not leak full payloads in URLs.

**Acceptance criteria:**

- Demo slug is unique and human-readable when possible.
- Signed token validates demo site access.
- Token secret comes from environment config.
- Token rotation endpoint exists.
- Expired/invalid token returns safe public error.

### P1-017 — Create demo site schema and API

**Story:** As an admin, I need to generate and manage demo website records.

**Acceptance criteria:**

- `demo_sites` table exists.
- API can create demo site for a restaurant.
- API can regenerate payload.
- API can publish/unpublish demo.
- API returns preview URL.

### P1-018 — Build first restaurant website template

**Story:** As a restaurant owner, I need to see a polished custom website demo.

**Acceptance criteria:**

- Template has hero, menu, about, hours, location, reservation form, CTA sections.
- Template is mobile responsive.
- Template renders dynamic payload.
- Template has clear “Make this live” or “Book a call” CTA.
- Empty optional sections degrade gracefully.

### P1-019 — Build public demo route

**Story:** As a restaurant owner, I need to open a unique link and view the personalized website without logging in.

**Acceptance criteria:**

- Public route loads by slug and token.
- Correct payload renders.
- Invalid/expired link shows safe 404/expired page.
- Page view event is tracked.
- Public route does not expose internal API responses.

### P1-020 — Add admin preview and copy-link UI

**Story:** As an admin, I need to preview and copy demo links before outreach.

**Acceptance criteria:**

- Admin detail page shows latest demo status.
- Admin can generate demo.
- Admin can preview demo.
- Admin can copy link.
- Admin sees warning if profile is not reviewed.

---

## P1-E05 — Reservation System

### P1-021 — Create reservation schema and API

**Story:** As a customer, I need to submit a reservation request from a demo or live site.

**Acceptance criteria:**

- `reservations` table exists.
- API validates name, phone, date, time, party size.
- Default status is `pending`.
- Source field is stored.
- Reservation is linked to restaurant.

### P1-022 — Add reservation form to demo template

**Story:** As a customer, I need a simple form to request a reservation.

**Acceptance criteria:**

- Form appears on demo website.
- Form handles loading, success, and error states.
- Successful request stores reservation.
- UI copy says request is pending/not confirmed.
- Reservation event is tracked.

### P1-023 — Build reservation status update API

**Story:** As a restaurant owner/admin, I need to confirm or reject reservation requests.

**Acceptance criteria:**

- API supports status transition to confirmed/rejected/cancelled.
- Invalid transitions are rejected.
- Updated timestamp is stored.
- Event `reservation.status_changed` is emitted.

### P1-024 — Build reservation dashboard

**Story:** As an admin or restaurant owner, I need to view and act on pending reservations.

**Acceptance criteria:**

- Dashboard lists reservations by restaurant.
- Filters include status and date.
- Detail view shows customer details and source.
- Confirm/reject actions update UI.
- Tenant access checks are enforced.

---

## P1-E06 — Email Outreach and Tracking

### P1-025 — Create campaign and email event schema

**Story:** As the system, I need to store outreach campaigns and events.

**Acceptance criteria:**

- `email_campaigns` and `email_events` tables exist.
- Campaign references restaurant and demo site.
- Events support sent, opened, clicked, bounced, unsubscribed, replied.
- Campaign has status and current step.

### P1-026 — Build email provider adapter

**Story:** As a developer, I need provider-independent email sending.

**Acceptance criteria:**

- `EmailProvider` interface exists.
- At least one provider implementation or mock provider exists.
- Provider errors are wrapped with safe error messages.
- Emails can be disabled or redirected in local/staging.

### P1-027 — Generate personalized email draft

**Story:** As an admin, I need a personalized outreach draft for each restaurant.

**Acceptance criteria:**

- Draft includes restaurant name.
- Draft includes demo link.
- Draft includes one clear CTA.
- Draft includes opt-out text.
- Draft can be reviewed before sending.

### P1-028 — Send first outreach email

**Story:** As an admin, I need to send the approved first email.

**Acceptance criteria:**

- Admin can trigger send from dashboard.
- Email job is queued.
- Sent event is recorded.
- Campaign status updates.
- Sending failure is visible in admin dashboard/logs.

### P1-029 — Implement tracking links

**Story:** As an admin, I need to know if restaurant owners click demo links.

**Acceptance criteria:**

- Tracking token maps to campaign and demo URL.
- Click route records `email.clicked`.
- User is redirected to demo.
- Multiple clicks are recorded.
- Invalid tracking token returns safe error.

### P1-030 — Implement open tracking pixel if provider supports it

**Story:** As an admin, I want optional open tracking, while recognizing it may be unreliable.

**Acceptance criteria:**

- Open pixel endpoint records event.
- Feature can be disabled by config.
- Dashboard labels opens as approximate.
- Open tracking does not block sending.

### P1-031 — Implement follow-up sequence

**Story:** As an admin, I need a simple three-email sequence when there is no response.

**Acceptance criteria:**

- Campaign supports step 1, step 2, step 3.
- Follow-up schedule is configurable.
- Follow-up stops if lead unsubscribes/replies/is closed.
- Worker can enqueue due follow-ups.
- Each sent step is logged.

### P1-032 — Implement suppression/unsubscribe list

**Story:** As a system owner, I need to avoid emailing unsubscribed contacts.

**Acceptance criteria:**

- Unsubscribe token/link exists.
- Unsubscribed email is suppressed from future sends.
- Campaign stops with reason `unsubscribed`.
- Admin can see unsubscribe status.

---

## P1-E07 — Analytics and Funnel Metrics

### P1-033 — Create analytics event schema and service

**Story:** As the business, I need funnel analytics to understand outreach performance.

**Acceptance criteria:**

- `analytics_events` table exists.
- Service can record event type, restaurant ID, demo site ID, metadata, timestamp.
- Event creation is non-blocking or low-risk.
- Tests cover event creation.

### P1-034 — Track demo views and CTA clicks

**Story:** As an admin, I need to see whether demos are engaging owners.

**Acceptance criteria:**

- Demo page view is tracked.
- Reservation CTA click is tracked.
- AI receptionist CTA click is tracked.
- Content automation CTA click is tracked.
- Events include demo site ID.

### P1-035 — Build analytics summary endpoint

**Story:** As an admin, I need a summary of funnel metrics.

**Acceptance criteria:**

- Endpoint returns lead count, demos generated, emails sent, clicks, demo views, reservations, CTA clicks.
- Summary can filter by date range.
- Endpoint is protected.
- Dashboard consumes endpoint or endpoint is ready for UI.

---

## P1-E08 — AI Receptionist MVP

### P1-036 — Define AI receptionist prompt and intent schema

**Story:** As a developer, I need a structured AI flow so the receptionist behaves predictably.

**Acceptance criteria:**

- System prompt discloses AI identity.
- Intent schema includes hours, location, menu, reservation, callback, unknown.
- Fallback response exists.
- Prompt forbids unsupported claims and human impersonation.
- Prompt tests cover required behaviors.

### P1-037 — Build restaurant knowledge context loader

**Story:** As the AI receptionist, I need approved restaurant knowledge to answer questions.

**Acceptance criteria:**

- Loader retrieves restaurant, profile, menu, reservation policy.
- Loader excludes unreviewed/private fields.
- Missing data produces explicit unknown/fallback behavior.
- Context payload is logged only in safe/redacted form.

### P1-038 — Implement LLM provider adapter

**Story:** As a developer, I need model-provider-independent AI generation.

**Acceptance criteria:**

- `LLMProvider` interface exists.
- Text generation and structured JSON generation are supported.
- Timeouts and retries are configured.
- Provider usage is logged for cost tracking later.
- Tests can use a fake provider.

### P1-039 — Integrate inbound voice webhook

**Story:** As a caller, I need to call a test number and reach the AI receptionist.

**Acceptance criteria:**

- Voice webhook validates provider signature/secret.
- Inbound call maps to a restaurant.
- AI greeting is returned/played.
- Session/call ID is stored.
- Unsupported restaurant mapping returns safe fallback.

### P1-040 — Handle general inquiry intents

**Story:** As a caller, I need answers for basic restaurant questions.

**Acceptance criteria:**

- AI answers hours questions from profile.
- AI answers location questions from profile.
- AI answers menu questions from menu items.
- Unknown questions trigger fallback/callback.
- Transcript snippets are captured when provider supports it.

### P1-041 — Create reservation from AI call

**Story:** As a caller, I need the AI to take my reservation request.

**Acceptance criteria:**

- AI collects name, phone, date, time, party size.
- AI confirms details with caller before creation.
- Reservation is created with source `ai_call` and status `pending`.
- AI states restaurant will confirm.
- Call log links reservation ID.

### P1-042 — Store call logs and summaries

**Story:** As a restaurant owner/admin, I need call summaries for follow-up.

**Acceptance criteria:**

- `ai_call_logs` table exists.
- Transcript and summary are stored when available.
- Intent and escalation flag are stored.
- Call-ended webhook triggers summary job if needed.
- Dashboard/API can fetch call logs by restaurant.

---

## P1-E09 — Content Automation MVP

### P1-043 — Create content job schema and API

**Story:** As a restaurant owner/admin, I need to submit content prompts.

**Acceptance criteria:**

- `content_jobs` table exists.
- API accepts restaurant ID and prompt.
- Job starts as `pending`.
- Job can be fetched by ID.
- Restaurant access is enforced.

### P1-044 — Implement content generation worker

**Story:** As the system, I need to generate content asynchronously from restaurant context.

**Acceptance criteria:**

- Worker loads restaurant/menu/brand context.
- Worker calls LLM provider.
- Output includes caption, hashtags, script, scene ideas, CTA.
- Job status changes to `completed` or `failed`.
- Errors are stored safely.

### P1-045 — Build content generation UI

**Story:** As a user, I need a simple dashboard UI to generate and copy content.

**Acceptance criteria:**

- User can enter prompt.
- UI shows loading state.
- Generated content appears in UI.
- User can copy caption/script.
- Content history is visible.

---

## P1-E10 — Dashboards, QA, Deployment, and Runbook

### P1-046 — Build admin dashboard shell

**Story:** As an internal admin, I need a dashboard to navigate leads, demos, campaigns, reservations, calls, and content jobs.

**Acceptance criteria:**

- Authenticated admin layout exists.
- Navigation includes restaurants, campaigns, reservations, analytics.
- Restaurant detail page links to profile/menu/demo/campaign sections.
- Responsive enough for laptop and tablet usage.

### P1-047 — Build restaurant owner dashboard shell

**Story:** As a restaurant owner, I need a simple view of reservations, calls, and generated content.

**Acceptance criteria:**

- Owner dashboard is protected by auth.
- Owner only sees assigned restaurant data.
- Sections include reservations, call logs, content jobs.
- Empty states are clear.

### P1-048 — Add seed/demo data

**Story:** As a developer/salesperson, I need sample restaurant data to test and demo the system quickly.

**Acceptance criteria:**

- Seed script creates at least two restaurants.
- Seed includes menus, profiles, demo sites, reservations, campaigns.
- Seed data contains no real private data.
- README explains how to load/reset seed data.

### P1-049 — Add deployment config

**Story:** As a team, we need a deployed staging MVP for demos.

**Acceptance criteria:**

- Dockerfile exists for API and worker.
- Deployment config exists for chosen target.
- Environment variables are documented.
- Staging uses separate DB/secrets.
- Health checks work after deploy.

### P1-050 — Add logging and error tracking

**Story:** As a developer, I need production-friendly logs to debug issues.

**Acceptance criteria:**

- API and worker use structured logs.
- Request ID is included in logs.
- Provider errors are logged with safe metadata.
- Sensitive fields are redacted.
- Critical worker failures are visible.

### P1-051 — Add API documentation

**Story:** As a coding agent/developer, I need API examples to integrate frontend and backend quickly.

**Acceptance criteria:**

- Endpoint list exists in docs or OpenAPI.
- Request/response examples exist for core endpoints.
- Auth requirements are documented.
- Error format is documented.

### P1-052 — Add release checklist

**Story:** As a team, we need a repeatable release gate.

**Acceptance criteria:**

- Checklist includes migrations, tests, smoke test, rollback notes.
- Smoke test covers lead → demo → reservation → email tracking.
- AI receptionist test call is included when voice is enabled.
- Release owner signs off before production demo usage.

---

## 3. Suggested Sprint Breakdown

### Sprint 1 — Foundation

- P1-001 to P1-009.
- P1-010 basic restaurant CRUD.

**Sprint exit:** API, DB, auth shell, restaurant CRUD working locally.

### Sprint 2 — Restaurant data and demo generator

- P1-011 to P1-020.

**Sprint exit:** Admin can create restaurant/profile/menu and generate public demo link.

### Sprint 3 — Reservations and outreach

- P1-021 to P1-032.

**Sprint exit:** Demo captures reservation; first email campaign sends and tracks clicks.

### Sprint 4 — Analytics and dashboards

- P1-033 to P1-035.
- P1-046 to P1-048.

**Sprint exit:** Internal dashboard shows lead/demo/campaign/reservation funnel.

### Sprint 5 — AI receptionist and content MVP

- P1-036 to P1-045.

**Sprint exit:** Test restaurant can receive inbound AI call; content prompt generates output.

### Sprint 6 — Deployment, QA, release

- P1-049 to P1-052.
- Stabilization and bug fixes.

**Sprint exit:** Staging MVP can be shown to real restaurant owners.

---

## 4. Agent Execution Guidance

When a coding agent implements this backlog:

1. Read `PHASE1_IMPLEMENTATION_GUIDE.md` first.
2. Implement tickets in dependency order.
3. Prefer small PRs/commits per ticket.
4. Do not introduce extra services without clear reason.
5. Keep provider SDKs isolated behind adapter interfaces.
6. Add tests with every service-level change.
7. Do not skip security/tenant checks.
8. Keep AI features bounded and auditable.
9. Mark unclear assumptions in code comments or a `docs/assumptions.md` file.
10. Stop before destructive production actions unless explicitly instructed by a human.

