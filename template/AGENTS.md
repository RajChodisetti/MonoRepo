# Restaurant Template Agent Instructions

These instructions specialize the repository-root `AGENTS.md` for everything
under `template/` and must not weaken its global safety or approval rules.

## Scope and Active Templates

This Next.js 16 app renders personalized restaurant demos on port `3000` from
one shared `RestaurantContent` model.

- `3` — Elysian, current default and first in the switch cycle
- `2` — Aurora
- `1` — Cinematic

`src/templates/foodie/` is inactive legacy material and is not imported by the
page renderer. Do not expose or assign template ID `4` to it unless the user
explicitly requests a product decision and all backend/database consumers are
updated with approval.

## Existing Patterns

- Resolve IDs only through `src/lib/templateConfig.ts`; keep the active union
  exactly aligned with backend constraints and analytics.
- Normalize API, signed-demo, and JSON data in `src/lib/adapters/`. Templates
  consume `RestaurantContent`; do not add template-specific backend fetches.
- Keep each template scoped by its immediate server-rendered `data-template`
  wrapper and its own `theme.css`. Shared behavior belongs in `src/components/`.
- Signed demos use `slug` plus opaque `token` and server-side payloads. Never put
  a restaurant payload or private metadata in a URL.
- Preserve media attribution, `unoptimized` handling for live Google media, and
  menu-document exclusion through `SourceAwareImage`/`PhotoAttribution` and the
  adapter boundary.
- Signed demos intentionally omit the template switcher/walkthrough and browser
  voice controls. Do not weaken that boundary accidentally.

## Dependency Impact

- Template ID/default/order changes require a repo-wide search across campaign
  links, generated-site handlers, admin UI, engagement events, constraints,
  migrations, `.env.example`, and documentation.
- Public payload or media-shape changes require coordinated backend handler,
  demo snapshot, adapter/type, attribution, and all-three-template checks.
- Reservation or voice protocol changes affect the Go public APIs and/or
  `voice-sales-agent/`; update both producers and consumers in the same task.
- A shared component or CSS change must be checked in all three templates at
  mobile, tablet, and desktop widths, including reduced motion and keyboard use.

## Safety

- Public pages must expose only approved public data. Preserve pending-only
  reservation wording and never imply a table is confirmed.
- Do not enable or place outbound AI calls. Customer-facing voice-flow or prompt
  changes require explicit approval under the root contract.
- Do not copy or cache Google Places media; preserve attribution and live-media
  handling. Do not edit generated `.next/` or TypeScript build-info files.

## Checks

Run from the repository root:

```bash
rtk npm --prefix template run test:unit
rtk npm --prefix template run lint
rtk npm exec --prefix template -- tsc --noEmit --incremental false --pretty false -p template/tsconfig.json
rtk npm --prefix template run build
```

For template routing, adapter, or shared-style changes, also run
`build:elysian`, `build:aurora`, and `build:cinematic`, then smoke
`?id=0&template=3`, `?id=0&template=2`, and `?id=0&template=1` locally.

<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
