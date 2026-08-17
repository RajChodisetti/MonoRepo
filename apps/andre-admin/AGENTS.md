# Andre Admin Prototype Instructions

These instructions specialize the repository-root `AGENTS.md` for
`apps/andre-admin/` and must not weaken its global safety or approval rules.

## Scope and Status

This Next.js 15 app is the local-only operations UI for the standalone Andre
real-estate prototype on port `8001`. It runs on port `3003` and provides
property CRUD, browser voice, and an outbound-call control. It is not the Tuvi
restaurant admin (`apps/web`) and is not in the production VM stack.

## Boundaries and Safety

- Treat `andre-voice-agent` as the API contract owner. Update its property,
  browser-WebSocket, session, and call routes together with this app's BFF and
  types.
- Never click, invoke, automate, or smoke `/api/voice/call` without separate
  explicit approval for the exact external phone call. A deploy/build request
  alone does not authorize dialing.
- `CALL_API_SECRET` and session signing material are server-only. Never expose
  them through `NEXT_PUBLIC_*`, client components, logs, or screenshots.
- The legacy `NEXT_PUBLIC_PRELOAD_ADMIN_PASSWORD` option is unsafe for shared or
  deployed environments. Do not configure, expand, or rely on it for real
  access control.
- Property leads, phone numbers, voice transcripts, and call identifiers are
  sensitive. Use synthetic fixtures and keep them out of commits and output.

## Existing Patterns

- Server routes and `src/lib/agent.ts` proxy the Andre API; client components do
  not receive the call API secret.
- `src/lib/session.ts` and middleware own the local admin session boundary.
- Browser voice uses the public WebSocket URL, while property and call actions
  stay behind same-origin routes.

## Safe Checks

Run from the repository root without starting the agent or calling providers:

```bash
rtk npm --prefix apps/andre-admin run lint
rtk npm exec --prefix apps/andre-admin -- tsc --noEmit --incremental false --pretty false -p apps/andre-admin/tsconfig.json
rtk npm --prefix apps/andre-admin run build
```

Do not use the outbound-call UI as a smoke test.

<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
