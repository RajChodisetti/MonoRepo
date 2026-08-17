# Restaurant Services Catalog Agent Instructions

These instructions specialize the repository-root `AGENTS.md` for everything
under `apps/restaurant-services-catalog/` and must not weaken its global safety
or approval rules.

## Scope and Boundary

This is a standalone Vite marketing/catalog surface on port `5173`. It is not
the internal admin portal, a generated restaurant template, or the main Go
application. The repo treats `web/` as the canonical
corporate site; confirm the intended public surface before moving or duplicating
content between the two.

## Existing Patterns

- Keep the app dependency-light: structure/content/interaction live in
  `src/main.js`, visual tokens and responsive behavior in `src/styles.css`, and
  static media in `public/media/`.
- Reuse the existing CSS custom properties, service-card interaction,
  accessibility attributes, and Lucide icon setup.
- This is a static client app. Only `VITE_*` values are public configuration;
  never add backend bearer tokens, provider credentials, or privileged APIs.
- Contact actions must remain safe client-side navigation or `mailto:` behavior
  unless a reviewed server-side boundary is introduced elsewhere.

## Dependency Impact

- Product/service copy is marketing positioning, not evidence that a feature is
  shipped. Cross-check the current backend, admin, template, and business-gap
  docs before strengthening a claim.
- If a shared Tuvi claim, URL, or asset changes, inspect `web/` and
  other public references, but do not copy implementations automatically.
- Media changes must preserve licensing, poster/video pairing, responsive
  sizing, and reasonable bundle/page weight.
- Keep keyboard behavior, visible focus, `aria-pressed`, reduced motion, and
  mobile layout intact when changing interactive cards or navigation.

## Safety

- `npm run deploy` publishes to Cloudflare Pages and is approval-gated. Never
  run it as part of implementation or verification without explicit approval.
- Do not claim live ordering, payments, loyalty, SMS, owner dashboards, or other
  roadmap features unless the product behavior actually exists and the user
  approves the public claim.
- Do not edit generated `dist/` output.

## Checks

Run from the repository root:

```bash
rtk npm --prefix apps/restaurant-services-catalog run build
```

There is no automated browser test suite here. For UI changes, also use local
`npm run dev` or `npm run preview` to check keyboard operation and phone,
tablet, and desktop layouts. Do not deploy during verification.
