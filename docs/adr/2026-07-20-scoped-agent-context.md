# ADR: Scoped Agent Context and Change-Impact Guardrails

Date: 2026-07-20
Status: Accepted

## Context

The repository used one 921-line, 39,931-byte root `AGENTS.md` for product
policy, subsystem conventions, current state, and delivery history. Codex's
default aggregate project-document limit is 32 KiB, so the file could be
truncated. Claude Code had no repository `CLAUDE.md`. Agents therefore received
large generic context but no deterministic path-specific dependency or test
guidance.

Regressions cluster at cross-system seams: Python ingestion/import and Go
sharing SQL contracts; Go public payloads feeding three templates; admin BFF types tracking
private API responses; and voice/corporate tools sharing reservation and
consultation APIs. Root `make test` covers only Go and cannot guard these seams.

## Decision

- Keep the root `AGENTS.md` below 16 KiB and reserve it for global invariants,
  safety, routing, and completion rules.
- Add concise nested `AGENTS.md` files at real subsystem boundaries. Nested rules
  own local architecture, patterns, dependency seams, and verification.
- Add thin `CLAUDE.md` files that import the canonical sibling `AGENTS.md`
  instead of maintaining a duplicate policy set.
- Maintain a cross-system dependency map and, when operational state warrants
  one, an optional short current-state snapshot under `docs/ai/`; keep detailed
  delivery history outside normal startup context.
- Add `scripts/agent-context.sh` for path-based context/check routing and
  `scripts/check-agent-context.sh` plus CI for hierarchy/import/size validation.
- Keep enforcement repository-native and dependency-free. Do not introduce a
  hosted context service or opaque generated knowledge base while the codebase
  can be routed accurately with versioned files and scripts.

## Options Considered

1. Keep expanding the root file. Rejected because it already exceeds the default
   instruction budget and mixes unrelated subsystem context.
2. Raise every developer's Codex byte limit. Rejected because it is local-tool
   configuration, does not help Claude/other agents, and encourages context
   growth rather than precision.
3. Adopt an external context-indexing package. Deferred because it adds runtime,
   freshness, privacy, and onboarding dependencies without enforcing change
   impact or tests.
4. Use scoped, versioned repository instructions with a deterministic router.
   Accepted because scope/precedence are supported by Codex and thin imports keep
   Claude guidance aligned.

## Consequences

- Agents load less global text and more relevant local constraints.
- Cross-stack work becomes explicit before edits and has a path-based check list.
- Adding a new deployable subsystem requires a scoped instruction pair and a
  router/check update.
- Static maps can drift; CI validates structure and maintainers must update the
  dependency map when contracts/topology materially change.
- This system improves reasoning and enforcement but does not replace regression
  tests. Frontend and voice browser/integration coverage remains limited even
  though the inbound-only voice policy suite exists.

## Rollback / Revisit Trigger

Revisit if the repo adopts a supported monorepo build graph that can generate
the same dependency/test routing, if agent tools standardize on another scoped
instruction format, or if measured regressions show the hierarchy is missing
critical seams. Roll back by restoring one canonical root contract only after
proving it fits every target agent's instruction budget.
