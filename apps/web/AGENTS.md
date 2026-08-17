<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

# Admin Portal Agent Instructions

These instructions specialize the repository-root `AGENTS.md` for everything
under `apps/web/` and must not weaken its global safety or approval rules.

## Scope

This is the Next.js 16 internal-admin portal on port `3002`. It supports lead
review, scrape jobs, demos, campaigns, media, and outreach operations. It is not
a public restaurant site or the Tuvi corporate website.

## Existing Patterns

- Keep browser requests same-origin through `adminFetch` and
  `/api/admin/proxy/*`. Never expose the backend bearer token to client code.
- Keep the session in the `tuvi_admin_token` httpOnly cookie. Direct backend
  access belongs in server routes through `apiFetch`.
- Put shared API shapes in `src/lib/types.ts`, constants in
  `src/lib/constants.ts`, and reusable UI in `src/components/`.
- Reuse `PageHeader`, `StatusBadge`, `EmptyState`, `ErrorBanner`, modal, and
  existing CSS patterns before adding a new component system.
- Every data screen must preserve loading, empty, error, disabled, and success
  states. Destructive or external actions need explicit user confirmation.

## Dependency Impact

- When a backend route, method, error, or response field changes, inspect
  `backend/internal/http/router.go`, its handler/service, the OpenAPI document,
  `src/lib/types.ts`, and every page using that shape.
- Changes to demo template IDs, labels, or order also affect `template/`,
  campaign rendering, generated-site links, analytics constraints, and database
  migrations. Search the whole repo before changing them.
- Keep publish, retry, scrape, health-check, and email-send operations as
  deliberate admin actions. Never trigger one during render, page load, or a
  read-only refresh.
- Do not hide an API contract mismatch with a client-side fallback. Fix or
  explicitly version the producer and consumer together.

## Safety

- This portal is `internal_admin` only. Preserve the BFF boundary and role
  checks; do not add private tokens, lead notes, transcripts, or credentials to
  browser-visible environment variables or logs.
- Never run a real scrape, health message, ad-hoc send, bulk send, publish, or
  production mutation while testing UI changes without explicit approval.
- Do not edit generated `.next/` output or TypeScript build-info files.

## Checks

Run from the repository root:

```bash
rtk npm --prefix apps/web run lint
rtk npm exec --prefix apps/web -- tsc --noEmit --incremental false --pretty false -p apps/web/tsconfig.json
rtk npm --prefix apps/web run build
```

For an affected screen, also verify login expiry/401 handling and the loading,
empty, error, and success paths against a local or explicitly approved backend.
