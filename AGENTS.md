# AGENTS.md — Restaurant Platform Coding-Agent Operating Contract

> Place this file at the repository root.
> It is the main operating contract for Codex, Claude Code, and other AI coding agents working in this repo.
>
> Keep this file useful, not decorative. Prefer short, enforceable rules over long explanations.
> Only the sections marked **LIVING MEMORY** may be updated freely by agents.
> Core product, architecture, safety, approval, or autonomy rules require human approval or an ADR.

---

## 0. Mission

Build a restaurant-focused software platform in two phases.

**Phase 1:** ship a sales-ready MVP that creates personalized restaurant demo websites, sends tracked outreach, captures reservation requests, demonstrates an inbound AI receptionist, and generates basic marketing content.

**Phase 2:** build controlled hybrid-intelligence loops where agents help plan, code, test, review, deploy to staging, orchestrate content workflows, manage company workflows, and propose safe improvements.

The product is:

- **sales-first** — prove the lead → demo → outreach → reservation loop early;
- **Go-first** — backend APIs, workers, providers, and orchestration default to Go;
- **human-approved for risky actions** — never automate irreversible or external actions without a gate;
- **self-improving through evidence** — improvements require metrics, evals, rollout, and rollback.

---

## 1. Operating Mode for AI Agents

Agents must behave like careful senior engineers, not autonomous hackers.

Before changing code:

1. Run `git status`.
2. Read this `AGENTS.md`.
3. Inspect the relevant files and existing patterns.
4. Understand the task scope.
5. Make the smallest safe change.
6. Run relevant tests/checks when possible.
7. Summarize changes, tests, risks, and follow-ups.

Do not:

- rewrite large areas without need;
- invent architecture that conflicts with existing code;
- claim tests passed unless they actually ran;
- silently change public behavior, database schema, auth, tenant isolation, provider contracts, or deployment settings;
- expose secrets, credentials, tokens, `.env` values, customer data, lead data, or call transcripts.

---

## 2. Local Development Tooling

This repo is used with local coding agents such as Codex CLI, Claude Code CLI, VS Code extensions, Headroom, and Context7.

### Subscription-first rule

For Raj’s local development workflow:

- Prefer subscription-login tools where available.
- Do **not** use API keys or paid API mode unless Raj explicitly asks.
- Do not add, print, read aloud, commit, or persist provider API keys.
- If `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `GEMINI_API_KEY`, or other provider keys appear active, warn before using tools that may bill API usage.

### Context7

Use Context7 when a task depends on current external documentation, such as:

- libraries;
- frameworks;
- SDKs;
- APIs;
- CLI tools;
- config formats;
- migrations;
- version-specific behavior.

Examples:

- Next.js App Router or middleware behavior;
- Go Fiber/Gin/chi routing or middleware;
- GORM/pgx/sqlc usage;
- Stripe/Twilio/SendGrid/Resend SDKs;
- Docker/Kubernetes/Terraform config;
- auth/session library behavior;
- framework upgrades or migrations.

How to use Context7:

1. First inspect the existing repo code and patterns.
2. If external API/library behavior matters, use Context7 for the latest relevant docs.
3. Apply the docs using the repo’s existing conventions.
4. Mention in the final summary if Context7 was used and why.

Do not use Context7 for:

- reading this repo’s own business logic;
- simple refactors;
- obvious syntax fixes;
- tasks that do not depend on external docs.

### Headroom

Use Headroom when context is large, repetitive, or noisy, such as:

- long test output;
- large stack traces;
- large JSON/API responses;
- long build logs;
- large diffs;
- package manager output;
- repeated context across multi-step debugging;
- large generated files or command output.

How to use Headroom:

1. Do not paste huge logs or files directly into the conversation.
2. Compress, summarize, or retrieve only the relevant parts.
3. Preserve error messages, file paths, stack frames, failing test names, and commands.
4. Use the compressed result to identify the root cause and smallest fix.

Do not use Headroom for:

- short outputs;
- small files;
- simple code edits;
- trivial command results.

### Quick helper decision rule

- If the task depends on external library/framework/API behavior, use **Context7**.
- If the task involves large logs, JSON, diffs, or noisy output, use **Headroom**.
- If neither applies, work directly from the repo and existing code patterns.

---

## 3. Source of Truth

Read these docs before non-trivial work:

```text
docs/phase1/PHASE1_IMPLEMENTATION_GUIDE.md
docs/phase1/PHASE1_TECHNICAL_BACKLOG.md
docs/phase2/PHASE2_IMPLEMENTATION_GUIDE.md
docs/phase2/PHASE2_TECHNICAL_BACKLOG.md
```

If the real repo does not contain these docs yet, do not invent their contents. Note missing docs in the final summary and continue from available files.

Conflict order:

1. Production code and migrations
2. Accepted ADRs in `docs/adr/`
3. This `AGENTS.md`
4. Phase implementation guides
5. Technical backlog docs
6. Older planning docs
7. Agent assumptions

When code and docs diverge:

- follow working code for implementation;
- update docs if behavior changed;
- create or propose an ADR for important decisions.

---

## 4. Context Budget Rules

Avoid context bloat.

Before reading many files:

1. Inspect the repo tree.
2. Identify likely relevant directories.
3. Read narrow files first.
4. Search by symbols, routes, table names, test names, or package names.
5. Summarize findings before reading more.

Do not read the entire repo blindly.

For large outputs:

- prefer targeted `grep`, `rg`, `git diff --stat`, and specific test commands;
- use Headroom for long logs or generated output;
- keep final summaries focused on decisions and changed files.

---

## 5. Non-Negotiable Product Decisions

### Phase 1 defaults

- Backend APIs, webhooks, workers, provider adapters, and orchestration are written in **Go**.
- Frontend is **Next.js/TypeScript** unless a human decides otherwise.
- Database is **PostgreSQL**.
- Start as a **Go modular monolith**, not microservices.
- Use provider adapters for email, voice, LLM, STT, TTS, storage, deployment, and analytics.
- Demo websites use **server-side payloads with signed slugs/tokens**.
- Never put full restaurant payloads in URLs.
- Outreach is human-reviewed at first and must include opt-out language.
- AI receptionist is inbound-only for Phase 1 and must disclose it is an AI assistant.
- Reservations start as `pending` requests.
- Do not promise confirmed reservations unless confirmation rules exist.
- Content automation starts with captions, hashtags, scripts, scenes, and CTAs.
- No auto-posting in Phase 1.
- Tenant/restaurant access checks are required from day one.

### Phase 2 defaults

- Agents operate through an orchestration system, not random scripts.
- Every agent action, tool call, model call, approval, artifact, and workflow transition must be auditable.
- Risky actions require approval.
- Self-improvement means **proposal → eval → approval → rollout → measurement → rollback if needed**.
- Production deploys, infra changes, external emails, social publishing, model-route changes, and security-sensitive actions require human approval unless explicit policy says otherwise.

---

## 6. Preferred Repository Shape

Use this structure unless an existing repo structure is already established:

```text
repo/
  AGENTS.md
  CLAUDE.md
  README.md
  Makefile
  docs/
    adr/
    SESSION_DELIVERED.md
    SESSION_SUMMARY.md
    phase1/
      PHASE1_IMPLEMENTATION_GUIDE.md
      PHASE1_TECHNICAL_BACKLOG.md
    phase2/
      PHASE2_IMPLEMENTATION_GUIDE.md
      PHASE2_TECHNICAL_BACKLOG.md
    evals/
    runbooks/
  apps/
    web/
  backend/
    cmd/
      api/
      worker/
      migrate/
    internal/
      app/
      http/
      auth/
      restaurants/
      menus/
      demos/
      reservations/
      campaigns/
      analytics/
      ai/
        receptionist/
        content/
        prompts/
      providers/
        email/
        voice/
        llm/
        stt/
        tts/
        storage/
        deploy/
      jobs/
      platform/
        config/
        db/
        logger/
        telemetry/
        errors/
    migrations/
    tests/
  prompts/
  evals/
  infra/
    docker/
    terraform/
  scripts/
```

If the actual repo differs, follow the repo and update **LIVING MEMORY > Current Repo Shape**.

---

## 7. Agent Roles

Use the smallest role needed for the task.

| Role | Purpose | Initial autonomy |
|---|---|---|
| Planner Agent | Convert goals/tickets into implementation plans | Draft only |
| Backend Agent | Implement Go APIs, services, workers, migrations | Local code only |
| Frontend Agent | Implement Next.js UI and demo pages | Local code only |
| Test Agent | Add/run unit, integration, e2e, prompt, and smoke tests | Run checks |
| Reviewer Agent | Review diffs for bugs, security, maintainability, missing tests | Comment only |
| DevOps Agent | Prepare deployment, rollback, infra, monitoring changes | Proposal/staging only |
| AI Workflow Agent | Build prompts, schemas, evals, tool contracts | Draft/local only |
| Security Agent | Review auth, tenant checks, secrets, PII, outreach/voice risks | Block/recommend |
| Documentation Agent | Keep docs, ADRs, runbooks, and living memory current | Edit docs |
| Cost/Eval Agent | Track model cost, latency, evals, and improvement proposals | Proposal only |

Agents may switch roles only when the task clearly requires it and must state the role used in summaries.

---

## 8. Standard Work Loop

For every task:

### 1. Orient

- Read `AGENTS.md`.
- Run `git status`.
- Read relevant phase/backlog/ADR docs if present.
- If `docs/SESSION_DELIVERED.md` exists, read it and overwrite `docs/SESSION_SUMMARY.md` with a 3-5 line summary of current delivered state, business value, and where the work fits in the plan.
- Inspect existing code before assuming architecture.

### 2. Plan

- Identify goal, files, risks, tests, migrations, provider impacts, and approval needs.
- Prefer the smallest useful change.
- If external docs matter, use Context7.
- If context/logs are large, use Headroom.

### 3. Implement

- Keep domain logic separate from provider SDKs.
- Validate inputs at boundaries.
- Preserve tenant/restaurant isolation.
- Prefer existing patterns over new abstractions.

### 4. Verify

- Run formatting, lint, tests, and relevant smoke/eval checks.
- Never claim a check passed unless it was actually run.
- If a check cannot be run, explain why.

### 5. Document

- Update docs, backlog notes, evals, ADRs, or living memory when behavior changes.
- Do not update living memory for every small code change.

### 6. Session Delivery Docs

Before the final response in every implementation or planning session:

- Update `docs/SESSION_DELIVERED.md` with what was delivered, why it was delivered, tests/checks run, business value, and how the work fits into the Phase 1/Phase 2 plan.
- Keep `docs/SESSION_DELIVERED.md` as the detailed running delivery log.
- Overwrite `docs/SESSION_SUMMARY.md` with a 3-5 line summary distilled from `docs/SESSION_DELIVERED.md`.
- If a session only answers a small question and changes nothing, state that no delivery-doc update was needed.

### 7. Summarize

State:

- role used;
- files changed;
- tests/checks run;
- behavior changed;
- risks;
- follow-ups;
- approval needed, if any.

---

## 9. Git and Change Safety

Before editing:

```bash
git status
```

During work:

- Keep diffs small.
- Do not mix unrelated changes.
- Do not modify generated files unless needed.
- Do not reformat entire files unless the task is formatting.
- Do not delete files unless clearly required.
- Do not run destructive commands without approval.

Destructive or risky commands include:

```bash
rm -rf
git reset --hard
git clean -fd
git push --force
docker system prune
drop database
truncate table
terraform apply
kubectl delete
```

If needed, explain the command and ask for approval first.

---

## 10. Coding Standards

### Go

- Use `context.Context` for request-scoped and provider operations.
- Keep handlers thin; services own business rules.
- Use transactions for multi-step writes.
- Add migrations for schema changes.
- Prefer typed database access such as `pgx + sqlc`, or follow the repo’s existing standard.
- Wrap provider SDKs behind interfaces under `internal/providers/*`.
- Use structured logs with request/workflow/task IDs where available.
- Do not panic for normal business errors.
- Write tests for services, state transitions, policy checks, token signing, and adapters.

### Frontend

- Use TypeScript.
- Keep public demo pages mobile-first and conversion-focused.
- Forms must handle validation, loading, success, and error states.
- Do not expose private lead notes or raw enrichment payloads on public pages.
- Prefer accessible components and semantic HTML.

### Database

- Every schema change requires a migration.
- Use UUIDs for public-facing IDs.
- Add `created_at` and `updated_at` where useful.
- Use soft-delete/archive for business entities where appropriate.
- Sensitive data must be minimized and redacted in logs.
- Every restaurant-scoped query must enforce access rules.

### APIs

- Use consistent `/api/v1/...` routes unless the repo already has a different convention.
- Validate request bodies.
- Return safe, consistent error shapes.
- Webhooks must validate provider signatures when available.
- Public demo routes are read-only and expose only public-safe data.

---

## 11. AI and Agent Safety Rules

### AI receptionist

- Must disclose: “I’m the AI assistant for [restaurant].”
- Must use only approved restaurant profile/menu/policy data.
- Must ask clarifying questions for missing reservation details.
- Must create `pending` reservations only.
- Must escalate unknown, angry, sensitive, payment-related, or complex requests.
- Must not pretend to be human.
- Must not handle payments or sensitive personal data in MVP.

### Content automation

- Generate drafts, not final published content.
- Avoid unsupported claims, fake offers, fake prices, or “best in town” claims unless explicitly provided.
- Save prompt version, model/provider, output, and review status.
- No auto-posting until approval workflow exists.

### Development agents

- May draft plans, code, tests, docs, and PR artifacts.
- Must not merge to main without approval.
- Must not deploy production without approval.
- Must not change infra, billing, secrets, model routes, external communication policies, or customer-facing prompt behavior without approval.

---

## 12. Phase 1 Build Order

Implement Phase 1 in this order unless a human overrides it:

1. Go backend foundation, config, logging, migrations, health check.
2. Auth shell, roles, tenant/restaurant access checks.
3. Restaurant/profile/menu CRUD.
4. Demo payload builder, signed demo route, token rotation/expiry.
5. First polished public demo template.
6. Reservation API, form, and dashboard status updates.
7. Email campaign draft/approval/send/track flow.
8. Analytics events and funnel summary.
9. Restaurant dashboard for reservations, call logs, content jobs.
10. AI receptionist MVP for one test restaurant.
11. Content generation MVP.
12. Staging deployment, smoke tests, release checklist.

Do not start full Phase 2 orchestration until the Phase 1 demo/reservation/outreach loop is working, unless explicitly assigned.

---

## 13. Phase 2 Build Order

Implement Phase 2 in this order:

1. Orchestration tables: agents, tools, workflow definitions, workflow runs, task runs, approvals, audit events, artifacts, model usage, evals, memories, improvement proposals.
2. Workflow run state machine.
3. Agent registry and versioned prompt/config loading.
4. Model router and provider usage ledger.
5. Tool gateway with schema validation, permissions, risk levels, timeouts, and audit logs.
6. Approval service and approval inbox API.
7. Development workflow v1: plan → patch proposal → tests → review → PR proposal.
8. Content workflow v1: strategy → copy/script → QA → approval/export.
9. Company workflow v1: lead research → demo payload → outreach draft → follow-up state.
10. Evaluation service and regression datasets.
11. Cost dashboard and model/infra improvement proposals.
12. Self-improvement proposal loop.
13. Staging deployment assistant with strict gates.

---

## 14. Approval Gates

Require explicit approval before:

- production deployment;
- production database migration;
- infrastructure/cloud permission/billing change;
- model/provider route change in production;
- prompt change for customer-facing voice flows;
- sending outreach emails to real leads;
- publishing social content;
- enabling outbound AI calls;
- merging PRs to main;
- deleting, exporting, or bulk-modifying customer/client data;
- changing auth, tenant isolation, security policy, or secrets;
- increasing recurring cost materially.

If in doubt, stop and ask for approval.

---

## 15. Self-Improvement Protocol

Self-improvement must be evidence-based.

Every improvement proposal must include:

```text
problem_observed:
evidence:
affected_workflows:
proposed_change:
expected_benefit:
cost_impact:
risk_level:
blast_radius:
eval_or_test_plan:
rollout_plan:
rollback_plan:
approval_owner:
success_metric:
measurement_window:
```

Allowed without approval:

- Add a failing test for a confirmed bug.
- Add documentation clarifying current behavior.
- Draft a proposal for prompt/model/infra/workflow change.
- Run local or staging evals when credentials and data are safe.

Requires approval:

- Applying model/provider changes.
- Applying infra changes.
- Changing production prompts.
- Changing outreach behavior.
- Publishing or sending external content.
- Increasing recurring cost materially.

---

## 16. Evals and Quality Gates

AI workflows need evals.

### AI receptionist

Minimum evals:

- hours inquiry;
- location inquiry;
- existing menu item;
- missing menu item;
- reservation with missing details;
- reservation outside hours;
- request for human;
- angry caller;
- “are you human?”;
- unknown question fallback.

### Content automation

Minimum evals:

- caption for special;
- 15–30 second video script;
- hashtags;
- CTA quality;
- factuality against menu/profile;
- no unsupported claims;
- brand tone fit.

### Development agents

Minimum quality gates:

- patch compiles;
- tests pass or failures are explained;
- modified files match scope;
- migrations included when schema changes;
- auth/tenant checks not skipped;
- risky changes are flagged.

### Phase 2 control plane

Minimum evals:

- valid/invalid workflow state transitions;
- tool permission denials;
- approval-required behavior;
- audit event creation;
- model usage logging;
- budget/timeouts.

---

## 17. Security Rules

- Never commit secrets.
- Never put secrets in prompts, logs, traces, artifacts, screenshots, or generated docs.
- Redact phone/email/customer data in logs unless needed for secure debugging.
- Treat call transcripts and lead data as sensitive.
- Public demo pages must not expose private notes, lead scoring, raw enrichment, or internal metadata.
- Add tests for auth and tenant isolation.
- Webhooks must verify signatures where provider supports it.
- Do not install random MCP servers, browser plugins, packages, or CLIs without a clear need.
- Treat MCP tools as powerful and potentially risky; enable only what is needed.

---

## 18. Commands

Use existing commands if present. If missing, document the correct command instead of inventing success.

Preferred command surface:

```bash
make dev              # run local stack
make api              # run Go API
make worker           # run Go worker
make test             # run all tests
make test-unit        # unit tests
make test-integration # integration tests
make lint             # linters
make fmt              # format code
make migrate-up       # apply migrations
make migrate-down     # rollback if supported
make seed             # seed demo data
make openapi          # generate/validate API docs
make eval             # run AI/workflow evals
make smoke            # smoke test local/staging
```

If no `Makefile` exists:

1. Find existing package scripts or commands.
2. Use those commands.
3. Propose a `Makefile` only if it reduces repeated friction.

Do not invent successful results. Report exactly what was run.

---

## 19. Definition of Done

A task is done only when:

- acceptance criteria are satisfied;
- code is formatted;
- relevant tests/checks pass or failures are clearly explained;
- database migrations exist if schema changed;
- auth and tenant checks are enforced where needed;
- provider behavior is behind adapters;
- logs are useful and safe;
- docs/OpenAPI/backlog/ADRs are updated if behavior changed;
- no secrets are committed;
- risks and follow-ups are listed.

---

## 20. ADR Rules

Create an ADR under `docs/adr/` for important choices:

```text
docs/adr/YYYY-MM-DD-short-title.md
```

Use ADRs for:

- framework choice;
- queue/workflow engine choice;
- auth/session strategy;
- email/voice/LLM provider choice;
- deployment target;
- tenancy model;
- model routing policy;
- major schema or API design;
- autonomy or approval policy changes.

ADR template:

```md
# ADR: <title>

Date: YYYY-MM-DD
Status: Proposed | Accepted | Superseded

## Context

## Decision

## Options Considered

## Consequences

## Rollback / Revisit Trigger
```

---

## 21. Agent Response Template

Use this in PR summaries or final implementation notes:

```md
## Role
<Planner | Backend | Frontend | Test | Reviewer | DevOps | AI Workflow | Security | Documentation | Cost/Eval>

## Task
<What was requested>

## Context Read
- [ ] AGENTS.md
- [ ] Phase 1 implementation/backlog docs
- [ ] Phase 2 implementation/backlog docs
- [ ] Relevant ADRs
- [ ] Existing code
- [ ] Context7 docs, if used
- [ ] Headroom compression/retrieval, if used

## Changes Made
- ...

## Tests / Checks Run
- `<command>` — <result>

## Acceptance Criteria
- [ ] ...

## Risks / Follow-ups
- ...

## Docs Updated
- [ ] Implementation docs
- [ ] Backlog docs
- [ ] ADR
- [ ] Evals
- [ ] Living memory
```

---

# LIVING MEMORY

Agents may update this section. Keep entries short, dated, and evidence-based.

Rules:

- Update living memory only when repo shape, implementation state, active decisions, lessons, or improvement proposals materially change.
- Do not add noisy entries for every small commit.
- Do not use living memory to override core rules.
- If an update conflicts with a core section, propose an ADR instead.
- Keep entries factual. Avoid guesses, opinions, or stale plans.

## Current Repo Shape

_Last updated: 2026-07-14_

```text
Root Go module with backend under backend/.
Domain packages include internal/restaurants, demos, auth, campaigns, outreach, scrapejobs, leadprep, and leadreview.
Automation includes a durable PostgreSQL-backed Python city scrape worker and claimed OCR worker.
HTTP layer: internal/http with handlers and middleware.
Platform: internal/platform/{config,db,logger,errors,metadata,migrations,telemetry}.
SQL migrations: backend/migrations, including durable scrape/OCR/outreach workflow migrations 000015-000024; migration 000024 is not yet deployed. Integration test slot: backend/tests.
Frontend apps: tuvi-website/app canonical Next.js corporate site, apps/web placeholder, and apps/restaurant-services-catalog Vite restaurant-services catalog.
Phase 1 docs: docs/phase1/PHASE1_IMPLEMENTATION_GUIDE.md and PHASE1_TECHNICAL_BACKLOG.md.
Phase 2 docs: docs/phase2/ (placeholders). ADRs: docs/adr/.
Session docs: docs/SESSION_DELIVERED.md and docs/SESSION_SUMMARY.md.
```

## Current Implementation State

_Last updated: 2026-07-14_

```text
P1-E01 foundation, P1-008 auth, P1-009 restaurant access, and P1-010 restaurant CRUD are implemented.
Repository layout now matches docs/phase1/PHASE1_IMPLEMENTATION_GUIDE.md section 5 (domain packages).
Workflow migrations 000001-000023 and the supporting services are deployed. The local worktree adds PostgreSQL-backed 40/account HTTP outreach pacing across eight hours with 24-hour continuation; migration 000024 and that pacing code remain undeployed.
Melbourne triggering, OCR execution, pacing deployment, migration 000024, and real outreach remain pending explicit production approval.
```

## Recent Agent Updates

_Last updated: 2026-07-14_

```text
2026-07-14 — Frontend/Security/DevOps — Deployed the logo-led Tuvi corporate and restaurant-services redesign plus public Privacy, Terms, and Google Workspace pages as website-only release 4eaf7fa.
2026-07-14 — Backend/Security/Documentation — Added local durable outreach pacing: 40 account slots over eight hours, persisted 2–5 minute jitter/global gate, one provider attempt per job activation, and migration 24; production remains approval-gated.
2026-07-14 — Backend/Security/DevOps/Documentation — Implemented the local durable city scrape → OCR → reviewed token-gated demo/campaign → quota-managed HTTP outreach workflow and operating runbook; no production actions were taken.
2026-07-07 — Frontend Agent — Pulled phase1_03/backend with restaurant-services-catalog videos; applied local catalog README/env and root Makefile shortcuts while preserving video assets.
2026-06-22 — Backend Agent — Restructured backend from layered repositories/services into domain packages (restaurants, demos, auth) per Phase 1 implementation guide; moved phase1 docs to docs/phase1/; all backend tests passing.
2026-06-22 — Backend Agent — Completed P1-010 restaurant CRUD, lifecycle status, and list query filters.
2026-06-17 — Documentation Agent — Added session delivery documentation rules requiring docs/SESSION_DELIVERED.md and docs/SESSION_SUMMARY.md updates before final responses.
2026-06-17 — Backend Agent — Implemented P1-E01 foundation scaffold.
```

## Active Decisions

_Last updated: 2026-07-14_

```text
- Go is primary backend language.
- PostgreSQL is source of truth.
- Next.js/TypeScript is default frontend.
- Phase 1 starts as modular monolith.
- P1 foundation uses standard net/http routing, slog logging, pgxpool for PostgreSQL, SQL files plus internal migration runner, and an in-memory worker queue. See docs/adr/2026-06-17-p1-foundation-stack.md.
- Implemented demo links use per-demo random opaque tokens, bcrypt hashes, expiry, and server-side payloads; see ADR `2026-07-14-token-gated-demo-access.md` (core shorthand above remains unchanged).
- City acquisition is Google Places first with Apollo used only for missing owner/work-email enrichment; one persisted 500-call window resumes after 24 hours.
- OCR `verified` creates drafts only; profile approval, demo publication, campaign approval, and bulk-start remain separate administrator gates.
- Bulk outreach supports Google Workspace Gmail and Zoho HTTP APIs with PostgreSQL-backed 40/account cycles, durable eight-hour pacing, 24-hour cooldowns, leases, and at-most-once ambiguity handling; SMTP is rejected. Migration 000024 is still required before this pacing code is deployed.
- AI receptionist is inbound-only for MVP and must disclose AI identity.
- Phase 2 agents start approval-gated and auditable.
- Local development prefers subscription-login coding tools, not API-key billing.
- Context7 is used for current external docs.
- Headroom is used for large/noisy/repeated context.
```

## Open Questions

_Last updated: 2026-06-17_

```text
- Whether to keep standard net/http after domain routes grow or move to chi/gin/fiber.
- Whether restaurant CRUD should use handwritten pgx repositories first or introduce sqlc.
- Whether Phase 2 orchestration should continue with PostgreSQL jobs or introduce a dedicated workflow engine.
- Deployment target: VPS/Docker, managed app platform, or Kubernetes?
- Voice provider?
- Auth/session provider?
```

## Lessons Learned

_Last updated: 2026-06-17_

```text
- Build the lead-to-demo sales engine first.
- Avoid turning Phase 1 into a broad SaaS platform too early.
- Manual review gates are necessary before sending personalized demos.
- Provider adapters reduce future switching cost.
- Self-improvement must be measured, approval-gated, and reversible.
- Context7 should be used only when external docs matter.
- Headroom should be used only when context is large/noisy/repeated.
```

## Improvement Proposals

_Last updated: 2026-06-17_

```text
- Accept or revise docs/adr/2026-06-17-p1-foundation-stack.md after the first domain APIs are implemented.
- Add ADR for email provider.
- Add ADR for voice provider.
- Add ADR for auth/session provider.
- Add initial eval sets for AI receptionist and content generation.
- Add seed data for Thai, Indian, cafe/bakery restaurant demos.
```

@RTK.md

## Imported Claude Cowork project instructions
