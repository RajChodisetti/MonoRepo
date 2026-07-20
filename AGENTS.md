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

_Last updated: 2026-07-19_

```text
Root Go module with backend under backend/.
Domain packages include internal/restaurants, demos, auth, campaigns, outreach, scrapejobs, leadprep, leadreview, and media; providers include Places and S3-compatible durable object storage.
Automation includes durable PostgreSQL-backed Python city scrape and background email-only OCR workers, plus automation/outreach/scrape_ledger.py and daily_pipeline.py/identity.py; migration 000027 backs the scrape ledger and migration 000030 backs the OCR daily request budget.
HTTP layer: internal/http with handlers and middleware; backend/internal/developer backs the protected read-only SQL/schema console.
Platform: internal/platform/{config,db,logger,errors,metadata,migrations,telemetry}.
SQL migrations: backend/migrations, including durable scrape/OCR/outreach workflow, restaurant-media, demo-ready backfill, template-removal, and OCR-reset migrations through deployed 000035. Integration test slot: backend/tests.
Frontend apps: tuvi-website/app canonical Next.js corporate site, apps/web internal admin portal (Next.js, port 3002, BFF-proxied to the main API — no longer a placeholder), apps/restaurant-services-catalog Vite restaurant-services catalog, and template/ personalized demo sites with three active templates.
Codex project skills: .codex/skills/ui-ux-pro-max from `uipro init --ai codex`.
Phase 1 docs: docs/phase1/PHASE1_IMPLEMENTATION_GUIDE.md and PHASE1_TECHNICAL_BACKLOG.md.
Phase 2 docs: docs/phase2/ (placeholders). ADRs: docs/adr/.
Session docs: docs/SESSION_DELIVERED.md and docs/SESSION_SUMMARY.md.
```

## Current Implementation State

_Last updated: 2026-07-19_

```text
P1-E01 foundation, P1-008 auth, P1-009 restaurant access, and P1-010 restaurant CRUD are implemented.
Repository layout now matches docs/phase1/PHASE1_IMPLEMENTATION_GUIDE.md section 5 (domain packages).
Production app release 778c0fe and migrations 000001-000035 are deployed. Gmail-only env-driven account health, the persisted Outreach UI job toggle, paced 40/account bulk delivery, manual selective outreach sends that bypass the bulk job/global generic-email flag, two-link outreach drafts (`Personalized demo websites` plus `Services catalog`), UUID previews with live Google preview fallback, engagement/transcript evidence, contact/Apollo visibility, OCR filtering, secured voice reads, hybrid public media, the clarified mobile template switcher, demo-ready lifecycle backfill, and the internal-admin Developer SQL/schema console are live.
Three configured Gmail OAuth mailboxes passed their first real-message provider health check. The persisted restaurant email job remains disabled; no bulk outreach job or restaurant delivery attempt was created during deployment.
The background OCR worker is enabled for canonical-email rows only. PostgreSQL enforces a global 200 vision-request UTC-day ceiling; provider SDK retries are disabled, and timeout/429 claims return to pending without consuming an attempt. Migration 000031 returned 21 legacy fingerprint-less verified profiles to pending; migration 000035 reset fingerprint-less Google OCR classifications for a safe refresh. At deployment verification for 88b58eb, OCR budget was 200/200 and the worker was waiting for the UTC reset.
The deployed hybrid media pipeline resolves fresh attributed Google Places photos live and requires an exact one-way OCR resource-fingerprint match; owner/licensed media can use S3-compatible storage, OCR enriches placement metadata, and menu documents/unmatched images fail closed on personalized websites. Admin-opened generated UUID previews may request a no-store live Google fallback when reviewed media is empty. The production storage provider remains disabled until a bucket/CDN is provisioned, which affects owner uploads but not the live Google resolver.
Migration 000032 moved 22 existing `lead` restaurants with draft/published demo records to `demo_ready`; production now has demo_ready=22 and lead=922. Migration 000034 removed Italian Villa `template_id=4`, mapped any existing template 4 demo sessions to template 3, and restored the database constraint to templates 1/2/3.
```

## Recent Agent Updates

_Last updated: 2026-07-19_

```text
2026-07-19 — Backend/Frontend-adjacent/Security/Test/DevOps/Documentation — Pushed and deployed 778c0fe to let internal-admin manual/selective sends from restaurant list/detail bypass the bulk email-job flag and generic `EMAIL_DISABLE_SENDING` while retaining sender, contact email, campaign draft, suppression, and admin checks. New outreach drafts now render only two promotional links: `Personalized demo websites` and `Services catalog`; production smokes returned 200/307/401 as expected with zero API/admin-web restarts, and no real email was sent during verification.
2026-07-19 — Backend/Frontend/Security/Test/DevOps/Documentation — Pushed and deployed e4d6801 from a clean origin/master release to add `/admin/developer`, a protected internal-admin read-only SQL console and schema browser with menu-item popularity shortcuts. Production smokes returned 307-to-login for `/admin/developer`, 401 for unauthenticated `/api/v1/developer/schema`, 200 for `/admin/login` and public restaurants, zero restarts on API/admin-web, and live DB counts showed 944 restaurants, 944 menus, and 47 menu item rows.
2026-07-19 — Frontend/Backend/Test/DevOps/Documentation — Pushed and deployed 88b58eb/migrations 000034-000035 to remove the rejected Italian Villa template and repair generated-preview photos. Template 4 is gone from app/admin/engagement constraints, generated UUID previews can request attributed live Google photos when reviewed media is empty, a live preview returned 10 Google media objects, production smokes returned 200, all Tuvi containers have zero restarts, and rollback points to `/opt/tuvi/releases/monorepo-b2a2f83`.
2026-07-19 — Frontend/Backend/Test/DevOps/Documentation — Pushed and deployed 5ffddf2/migration 000033 with the Italian Villa experimental personalized-site template; superseded later the same day by 88b58eb removal. The admin Demo tab listed template 4 under Experimental templates, engagement accepted `template_id=4`, production demo/admin/API/voice smokes returned 200, all Tuvi containers had zero restarts, and rollback pointed to `/opt/tuvi/releases/monorepo-3b0c246`.
2026-07-19 — Backend/DevOps/Test/Documentation — Pushed and deployed cbc2eb8/migration 000032 and installed `.codex/skills/ui-ux-pro-max` via `uipro init --ai codex`. The demo-ready filter is backed by lifecycle updates plus a 22-row production backfill; API/admin/website/demo/voice smokes return 200, all Tuvi containers have zero restarts, outreach remains off with zero active bulk-send jobs, and OCR remains at 200/200 with 0 verified profiles.
2026-07-19 — Full-Stack/Backend/Frontend/AI-Workflow/Security/Test/DevOps/Documentation — Pushed b5f7299 and packaging fix 6c21c15 directly to master and deployed 6c21c15/migration 000031. Live attributed Google resolution, OCR-aware non-menu website placement, admin media controls, S3-compatible owned media support, and the clear mobile template preview are live. All production surfaces return 200 with zero restarts; OCR is stable and waiting at 200/200 for the UTC reset, outreach remains off, and 3/3 Gmail accounts remain healthy. Durable owner uploads remain disabled pending bucket/CDN configuration.
2026-07-19 — Full-Stack/Backend/Frontend/AI-Workflow/Security/Test/Documentation — Locally implemented unreleased migration 000031 and hybrid website media: live attributed/non-cached Google resolution, durable S3-compatible owner/licensed assets, non-destructive imports, OCR metadata/template placement, admin upload/status controls, and menu-document exclusion at every public boundary. 165 Go tests and both Node 22 production builds pass; production remains caffcfb/000030 pending storage configuration and approval.
2026-07-19 — Full-Stack/Backend/Frontend/AI-Workflow/Security/Test/DevOps/Documentation — Pushed and deployed caffcfb with Gmail health/outreach/admin engagement plus a background email-only OCR worker, durable 200/day global vision budget, timeout/429 claim release, and the `/admin` scrape-detail redirect fix. Migrations 000015-000030 were applied; 3/3 Gmail checks are healthy/provider-accepted, restaurant outreach remains off with zero recent attempts, and OCR was running at 17/200 with no email-less claims.
2026-07-18 — Full-Stack/Backend/Frontend/Security/Test/Documentation — Reconciled the latest merged GitHub state and locally added Gmail-only env-driven credentials/health, a persisted UI outreach switch, presentation plus personalized tracked links, admin and outreach template/foreground-time/transcript evidence, address/phone/Apollo diagnostics, automatic click interest, OCR filtering, and authentication for existing voice reads. Production data shows 708 restaurants, 589 missing email, 511 Apollo `no_candidate`, 78 `skipped_no_domain`, and 119 successful enrichments; no deployment or real email occurred because production has zero Gmail OAuth account records.
2026-07-17 — AI Workflow/Backend/Frontend/Security/Test/Cost-Eval/DevOps — Benchmarked four low-cost VLM routes, selected Gemma 3 4B via DeepInfra, deployed fd2cc94/migration 000028, and changed OCR verification to require every scraped photo. A one-row production pilot completed 10/10 photos in 51 seconds for about $0.00041; OCR remains disabled and unscheduled.
2026-07-17 — AI Workflow/Security/Cost-Eval/DevOps — With explicit approval, changed the production OCR model variables to Qwen2.5-VL-7B via Hyperbolic and ran an exactly five-profile pilot after database/config backups. All 50 calls returned HTTP 400 because the live model registry now maps this model only to Featherless; 0 verified / 5 failed, no image rows or drafts, and OCR remains disabled and unscheduled.
2026-07-17 — Full-Stack/Backend/Security/DevOps — Deployed release 199241c with on-demand attributed Google Places photo URLs, explicit OCR checked state, real restaurant demo snapshots, and UUID-linked generator templates. A three-row OCR preflight resolved images but failed because the configured Qwen2-VL route is no longer served; OCR remains disabled pending explicit approval for a current model route.
2026-07-16 — Documentation Agent — Merged origin/master (PR #8, admin_portal) into agent/tuvi-oauth-homepage-verification, bringing in the real apps/web admin portal (dashboard, scrape-jobs, restaurants, outreach screens) and automation/outreach scrape-ledger changes; updated docs/SERVICES.md and this Repo Shape entry to match. No production deployment performed; apps/web is not yet wired into infra/docker/docker-compose.vm.yml or the VM Caddyfile.
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

_Last updated: 2026-07-19_

```text
- Go is primary backend language.
- PostgreSQL is source of truth.
- Next.js/TypeScript is default frontend.
- Phase 1 starts as modular monolith.
- P1 foundation uses standard net/http routing, slog logging, pgxpool for PostgreSQL, SQL files plus internal migration runner, and an in-memory worker queue. See docs/adr/2026-06-17-p1-foundation-stack.md.
- Implemented demo links use per-demo random opaque tokens, bcrypt hashes, expiry, and server-side payloads; see ADR `2026-07-14-token-gated-demo-access.md` (core shorthand above remains unchanged).
- City acquisition is Google Places first with Apollo used only for missing owner/work-email enrichment; one persisted 500-call window resumes after 24 hours.
- Durable OCR uses `google/gemma-3-4b-it:deepinfra`; only canonical-email rows are claimed, the global vision budget is capped at 200 requests per UTC day in PostgreSQL, and timeout/429 claims return to pending without consuming an attempt. `verified` requires every discovered scraped photo to resolve and return a successful structured result. Verification creates drafts only; profile approval, demo publication, campaign approval, and bulk-start remain separate administrator gates.
- Personalized-site media uses a hybrid policy: Google Places photos are resolved live with attribution, no application/optimizer caching, and an exact one-way OCR resource-fingerprint match, while only owner-granted or separately licensed assets may be copied into S3-compatible durable storage. Admin-opened generated UUID previews may request a no-store live Google fallback when reviewed media is empty; published/token-gated demos stay on the reviewed-media path. Menu documents and unmatched photos are never public; OCR metadata controls hero/gallery placement. See ADR `2026-07-19-hybrid-restaurant-media.md`.
- Restaurant outreach uses Google Workspace Gmail only, loaded from an env JSON account list with one OAuth refresh token per mailbox. PostgreSQL backs the admin UI job toggle, 40/account cycles, durable eight-hour pacing, 24-hour cooldowns, leases, at-most-once ambiguity handling, and daily real-message health evidence; SMTP and Google API-key mailbox auth are rejected. See ADR `2026-07-18-gmail-outreach-and-health.md`.
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
