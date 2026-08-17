# Documentation Scope Instructions

## Scope

Applies to `docs/**`: current state, ADRs, runbooks, OpenAPI, phase plans, business
plans, work logs, and session delivery records.

## Source roles

- `docs/ai/CURRENT_STATE.md`: optional short dated navigation snapshot when the
  repository maintains one.
- `docs/ai/DEPENDENCY_MAP.md`: stable cross-system seams and check routing.
- `docs/SESSION_SUMMARY.md`: 3–5 line most-recent session handoff.
- `docs/SESSION_DELIVERED.md`: append-only detailed delivery history; do not read
  or rewrite wholesale for normal orientation.
- `docs/adr/*`: accepted/proposed architectural decisions.
- `docs/openapi/openapi.yaml`: implemented HTTP contract, reconciled with router.
- `docs/phase1/*`: target behavior/backlog; code may have advanced or diverged.
- Dated work logs and `ARCHITECTURE_CHANGES.md`: historical evidence, not current
  operational truth unless reconfirmed.

## Editing rules

- State whether a fact is current, deployed, local/unreleased, proposed, or
  historical. Never collapse those states.
- Ground implemented claims in code, migrations, tests, or read-only evidence.
- Do not record secrets, personal data, raw transcripts, or environment values.
- Keep links relative and verify referenced paths exist.
- Update OpenAPI with API contract changes; update an ADR for material decisions.
- Only the coordinating agent edits session/state docs in multi-agent work.
- Do not copy large code blocks or logs when a precise path/command/result is
  enough.

## Verification

```bash
rtk ./scripts/check-agent-context.sh
rtk make openapi                    # only when OpenAPI changes
rtk git diff --check
```

If `make openapi` needs a package download or network access, report that rather
than claiming validation succeeded.
