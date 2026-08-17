# AI Task Context System

This directory is the compact navigation layer for coding agents. It does not
replace code, migrations, ADRs, or scoped `AGENTS.md` files; it tells an agent
which of those sources matter for a specific task.

## Why this exists

The repository previously relied on one 39,931-byte root `AGENTS.md`. Codex's
default total project-document budget is 32 KiB, so the file could be truncated.
It also mixed global safety rules, subsystem details, history, and current state,
causing agents to load too much context while still missing downstream seams.

The replacement has four parts:

1. A small root contract with global product and safety invariants.
2. Nested `AGENTS.md` files that apply only to their subsystem.
3. Thin `CLAUDE.md` imports so Claude Code consumes the same canonical rules.
4. A path router and dependency map that expose affected consumers and checks.

The hierarchy follows the current Codex rule that a nested `AGENTS.md` applies
to its directory tree and takes precedence over a parent file. Claude files use
`@AGENTS.md` imports instead of copying rules. See the accepted ADR for rationale.

## Standard workflow

From the repository root:

```bash
rtk git status --short
rtk ./scripts/agent-context.sh backend/internal/demos template/src
```

If continuing existing dirty work:

```bash
rtk ./scripts/agent-context.sh --changed
```

Then:

1. Read the root and printed scoped instructions.
2. Read the printed ADR/runbook/contracts and the optional `CURRENT_STATE.md`
   snapshot when it exists.
3. Search producers and consumers named in `DEPENDENCY_MAP.md`.
4. Write a small task brief before editing.
5. Run the printed checks and inspect the final diff.

## Small task brief

Use this structure in a plan or working note; do not create a permanent file for
every task.

```text
goal:
user-visible acceptance criteria:
in-scope paths:
out-of-scope paths:
contracts changed (API / SQL / job / config / UI):
known consumers:
security or approval gates:
tests and builds:
rollback or safe failure behavior:
```

If the contract or consumer list is unknown, investigate before editing. A task
is not “frontend only” when it changes an API shape, and it is not “backend only”
when a public payload or database contract has other-language consumers.

## Reading policy

- Read `CURRENT_STATE.md` when present; it is an optional short dated snapshot,
  not a substitute for code, migrations, or live operational evidence.
- Read the nearest scoped `AGENTS.md` for every edited path.
- Read accepted ADRs for the behavior being changed.
- Read `docs/openapi/openapi.yaml` when an HTTP contract changes.
- Read the relevant Phase 1 ticket for acceptance criteria, not the entire
  backlog by default.
- Search `docs/SESSION_DELIVERED.md` only for historical evidence. Do not load it
  wholesale during orientation.
- Treat `docs/ARCHITECTURE_CHANGES.md` and dated work logs as historical unless
  their statements are confirmed by current code/state.

## Context7, RTK, and large output

- RTK reduces shell-output noise; it does not supply business context or replace
  tests. Prefix every agent-run shell command with `rtk`.
- Context7 is for current external documentation after local code is inspected.
  It is not a substitute for understanding this repo's contracts.
- Use targeted searches and compressed test/log output, but retain the exact
  failing assertion, path, and command.

## Maintenance

Run the structural guardrail after changing this system:

```bash
rtk ./scripts/check-agent-context.sh
```

Keep global rules in the root, subsystem patterns in the nearest nested file,
cross-subsystem seams in `DEPENDENCY_MAP.md`, and historical deliveries in the
session log. Avoid duplicating the same rule across all four locations.

## References

- [Codex AGENTS.md implementation and 32 KiB default](https://github.com/openai/codex/blob/main/codex-rs/core/src/agents_md.rs)
- [Claude Code repository](https://github.com/anthropics/claude-code)
- [Scoped-agent-context ADR](../adr/2026-07-20-scoped-agent-context.md)
