<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

# Tuvi Corporate Website Agent Instructions

These instructions specialize the repository-root `AGENTS.md` for everything
under `web/` and must not weaken its global safety or approval rules.

## Scope

This is the canonical Next.js 16 Tuvi corporate website, normally served on
port `3001`. It includes public marketing and legal pages, restaurant search and
SEO reports, consultation booking, and the corporate browser-voice experience.

## Existing Patterns

- Keep shared navigation and layout in `src/components/layout/`, page sections
  in `src/components/sections/`, and product/resource content under
  `src/content/`.
- Restaurant search and report browser requests use same-origin
  `src/app/api/restaurants/*` routes. Provider credentials and privileged Go API
  access remain server-side; never expose them through `NEXT_PUBLIC_*` values.
- Consultation browser requests use same-origin
  `src/app/api/consultations/*` routes. Only server routes may attach
  `TUVI_API_TOKEN` to the main Go API.
- PostgreSQL through the main Go API is the consultation-slot authority. Do not
  add a second calendar or local availability source.
- Browser voice uses `/browser-stream?agent=corporate`; keep its event contract
  aligned with `voice-sales-agent/` and the shared hook/widget behavior.
- Preserve the existing responsive structure, legal routes, accessible motion
  behavior, and media/cache boundaries.

## Dependency Impact

- Search/report schema changes affect Go SEO handlers, the restaurant BFF
  routes, report presentation/PDF helpers, tests, and public result pages.
- Consultation schema, validation, or status changes affect the Go
  consultations domain, API proxy routes, booking form, and voice tools.
- Voice event, readiness, or tool changes affect this app, `template/`, and
  `voice-sales-agent/`; update all consumers together.
- Service and guarantee claims must match implemented behavior. Do not present
  roadmap capabilities as live.
- Asset replacements require responsive/performance checks and stable public
  paths unless every reference and cache behavior is updated.

## Safety

- The `/api/voice-agent/call` compatibility route does not authorize outbound
  calling. Do not invoke, expose, enable, or expand an outbound call flow
  without explicit human approval. Never use real phone numbers in tests.
- Customer-facing voice prompt or policy changes require explicit approval.
- Do not log search/consultation contact details, API tokens, call secrets, or
  voice transcripts. Do not edit `.next/` or TypeScript build-info output.

## Checks

Run from the repository root with Node 22:

```bash
rtk npm --prefix web run lint
rtk npm exec --prefix web -- tsc --noEmit --incremental false --pretty false -p web/tsconfig.json
rtk npm --prefix web run build
```

For routing or content changes, verify `/`, `/book`, `/privacy`, `/terms`,
`/google-workspace`, and an explicitly safe report path. Test bookings only
against a local or explicitly approved backend; do not place a phone call.
