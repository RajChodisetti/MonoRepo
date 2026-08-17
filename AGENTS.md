# AGENTS.md — Tuvi Repository Operating Contract

This file contains only rules that apply to the whole repository. More-specific
instructions live in nested `AGENTS.md` files and take precedence inside their
directories. Keep this file below 16 KiB: Codex has a 32 KiB default budget for
all project instructions combined, and an oversized root file can hide the rules
an agent needs most.

## 1. Mandatory task start

Before changing files:

1. Run `rtk git status --short` and preserve all pre-existing work.
2. Run `rtk ./scripts/agent-context.sh <planned paths>`; use `--changed` when
   continuing existing work.
3. Read `docs/SESSION_SUMMARY.md`, `docs/ai/DEPENDENCY_MAP.md`, every nested
   `AGENTS.md` that applies to the planned files, and
   `docs/ai/CURRENT_STATE.md` when that optional snapshot exists.
4. Inspect the implementation, tests, callers, and contracts before deciding on
   a change. Do not infer the architecture from filenames or planning docs.
5. State the intended scope, dependency impact, checks, and approval gates.

Do not read the entire 100+ KiB `docs/SESSION_DELIVERED.md` during normal
orientation. It is a historical log; search it only when older evidence matters.

## 2. Instruction and source-of-truth order

Direct system, developer, and user instructions override repository docs. Within
the repo, use this order:

1. Working production code and applied migrations.
2. Accepted ADRs under `docs/adr/`.
3. The nearest nested `AGENTS.md`, then this file.
4. `docs/ai/DEPENDENCY_MAP.md` and the optional
   `docs/ai/CURRENT_STATE.md` snapshot when present.
5. Phase implementation guides and backlogs.
6. READMEs, historical notes, and agent assumptions.

When code and documentation disagree, implement against working code, flag the
disagreement, and update the stale document when it is in scope. Never invent
the contents of missing docs. The Phase 2 guide/backlog are currently absent.

## 3. Context router

| Path | What it owns | Scoped instructions |
| --- | --- | --- |
| `backend/` | Go API, worker, domain logic, providers, PostgreSQL migrations | `backend/AGENTS.md` |
| `automation/outreach/` | Durable city ingestion/import and Python scrape worker | `automation/outreach/AGENTS.md` |
| `apps/web/` | Internal admin Next.js portal and same-origin BFF | `apps/web/AGENTS.md` |
| `apps/andre-admin/` | Local-only Andre prototype operations UI | `apps/andre-admin/AGENTS.md` |
| `template/` | Public personalized restaurant renderer and three templates | `template/AGENTS.md` |
| `web/` | Canonical Tuvi corporate Next.js site | `web/AGENTS.md` |
| `apps/restaurant-services-catalog/` | Standalone static Vite catalog | `apps/restaurant-services-catalog/AGENTS.md` |
| `voice-sales-agent/` | Browser/voice runtime for corporate and restaurant modes | `voice-sales-agent/AGENTS.md` |
| `andre-voice-agent/` | Legacy real-estate voice prototype with active outbound-call code | `andre-voice-agent/AGENTS.md` |
| `ocr-electrical-poc/` | Standalone electrical-image OCR proof of concept | `ocr-electrical-poc/AGENTS.md` |
| `infra/` | Docker/VM topology and deployment configuration | `infra/AGENTS.md` |
| `docs/` | ADRs, runbooks, API docs, current state, and delivery history | `docs/AGENTS.md` |

For a task that crosses paths, read every applicable file. Use
`docs/ai/DEPENDENCY_MAP.md` for producer/consumer seams and
`docs/ai/README.md` for the complete workflow.

## 4. Product mission and non-negotiable behavior

Tuvi is a sales-first restaurant platform. Phase 1 proves the lead → reviewed
profile → personalized demo → approved outreach → engagement/reservation loop.
Phase 2 adds auditable, approval-gated orchestration only after that loop works.

Global product rules:

- Go is the default backend language; PostgreSQL is the source of truth.
- The backend remains a modular monolith unless an accepted ADR says otherwise.
- Provider SDKs stay behind adapters; domain services own business policy.
- Public demo data is server-side and accessed with opaque, expiring tokens.
  Never place a full restaurant payload or bearer secret in a URL.
- Public surfaces must never expose lead notes, raw enrichment, private metadata,
  credentials, or unreviewed data.
- Authenticated/private restaurant-scoped reads and writes require role and
  membership checks; deliberately public site, media, availability, and pending
  reservation endpoints retain their separate fail-closed contracts.
- Reservation submissions are `pending` requests until explicitly confirmed.
- Restaurant AI must disclose that it is an AI assistant, use approved knowledge,
  and escalate uncertainty. It must not pretend to be human.
- Phase 1 voice behavior is inbound-only. Do not place or enable outbound AI
  calls even if experimental outbound code exists.
- Outreach sequence versions remain drafts until an administrator approves
  them. Eligible inferred-business leads may then enroll in the active approved
  sequence, but delivery still requires the persisted operator send control,
  schedule, quota, and provider safeguards. Never auto-post generated content.
- Google media may be resolved live with attribution and must not be persisted.
  Only explicitly approved owner-granted/licensed assets may be stored durably,
  and those assets require approval before public use. Menu documents and
  unmatched/unreviewed images fail closed on public sites.

## 5. Architecture and dependency rules

- `backend/cmd/*` are entrypoints. `backend/internal/app` wires dependencies.
  The API currently adapts the standard `net/http` router through Fiber.
- HTTP handlers translate boundaries; services own policy; repositories own
  persistence; `internal/providers/*` owns vendor behavior.
- The Go API/worker and Python city-ingestion/import process share PostgreSQL
  contracts. A schema, consent, or lifecycle change must inspect both languages.
- The admin browser talks to `apps/web` BFF routes. Bearer tokens remain in
  httpOnly server-side sessions; client components do not call the Go API
  directly.
- Public Go payloads feed `template/src/lib/adapters`, shared TypeScript data
  types, and Cinematic/Aurora/Elysian. Contract changes must verify all three.
- Corporate consultation booking in `web/` uses the main Go API through
  same-origin server routes.
- Production ingestion enters through durable scrape jobs. Legacy/manual Python
  scripts must not be scheduled beside the durable worker.

Before changing an interface, search for all readers and writers: routes, JSON
fields, SQL columns, job payloads, env variables, template IDs, and provider
contracts. Update producer, consumers, tests, and docs in the same task or state
why a consumer is intentionally unaffected.

## 6. Change discipline

- Make the smallest complete change. Do not rewrite unrelated areas.
- Preserve dirty work and avoid generated files unless generation is required.
- Never run `git reset --hard`, `git clean -fd`, broad recursive deletion,
  database drop/truncate, Docker prune, or an equivalent destructive command
  without explicit approval for exact targets. Prefer recoverable operations
  and do not delete unrelated files merely to make a check pass.
- Do not silently change public behavior, auth, tenant isolation, schemas,
  provider contracts, deployment settings, prompt policy, or external actions.
- Every schema change needs the next sequential `.up.sql` and `.down.sql` pair,
  relevant repository updates, and migration discovery tests.
- Every API contract change must inspect `backend/internal/http/router.go`,
  handler/service tests, `docs/openapi/openapi.yaml`, and affected frontend
  clients/types.
- Every config change must inspect `.env.example`, typed config, Docker/VM env
  forwarding, and the services that consume it.
- Add or update a regression test for behavior changes and confirmed bugs.
- Do not mark a task done with unexplained failing checks.

## 7. Tooling and context efficiency

- Prefix every shell command with `rtk`. In chains, prefix every command segment.
- Prefer `rg`/`rg --files` for search and narrow reads over whole-repo dumps.
- Use Context7 only when current external library, framework, SDK, API, CLI, or
  configuration behavior matters. Inspect repo patterns first.
- Use Headroom/RTK summaries for large logs, diffs, JSON, and test output while
  retaining exact errors, paths, stack frames, and failing test names.
- Prefer subscription-authenticated developer tools. Do not use billable API-key
  mode unless Raj explicitly asks. Never print or persist secrets.
- Runtime-aligned local versions are Go 1.26, Node 22, and Python 3.12. A check
  that passes only on a different host version is not sufficient evidence.

## 8. Security and approval gates

Never expose secrets, `.env` values, customer/lead data, call transcripts, or
personal information in prompts, logs, screenshots, fixtures, or summaries.
Redact sensitive data and validate webhook signatures where supported.

Explicit human approval is required before:

- production deploys or production migrations;
- infrastructure, permissions, billing, secrets, or production provider/model
  routing changes;
- authentication, tenant-isolation, or security-policy changes;
- customer-facing voice prompt/policy changes;
- sending outreach to real leads, publishing social content, or enabling
  outbound calls;
- merging to `master`, force-pushing, deleting/bulk-modifying customer data, or
  materially increasing recurring cost.

Local code, tests, docs, dry runs, and read-only diagnostics are allowed when
they do not trigger external actions or use unsafe production data.

## 9. Verification and definition of done

Use the path-aware commands printed by `scripts/agent-context.sh`; do not assume
`make test` covers this monorepo. At minimum:

- format/lint/type-check the changed unit;
- run targeted tests plus the nearest integration/build check;
- verify changed contracts at each consumer seam;
- run `rtk git diff --check` and inspect the final diff;
- report commands exactly as run, including failures or skipped checks.

A task is done only when acceptance criteria are satisfied, risky behavior is
still gated, docs/contracts are current, and no unrelated changes or secrets
were introduced.

## 10. ADRs and session delivery

Create or update an ADR for major framework, provider, tenancy, schema/API,
deployment, model-routing, or autonomy-policy decisions. Use
`docs/adr/YYYY-MM-DD-short-title.md` with Context, Decision, Options,
Consequences, and Rollback/Revisit Trigger.

At the end of an implementation or planning session:

1. Append a concise entry to `docs/SESSION_DELIVERED.md` with role, delivery,
   checks, business value, risks, and deployment/approval state.
2. Overwrite `docs/SESSION_SUMMARY.md` with a 3–5 line current summary.
3. When `docs/ai/CURRENT_STATE.md` exists, update it only when deployed state,
   repo topology, a material active decision, or a significant unreleased
   cross-cutting change changed.

Only one coordinating agent should edit session/state docs in a multi-agent
task to avoid conflicts.
