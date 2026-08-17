# Restaurant Website Templates

Premium Next.js templates sharing the same restaurant payload, switched via `TEMPLATE` env variable or `?template=`.

| `TEMPLATE` | Name | Style |
|------------|------|-------|
| `1` | **Cinematic** | Warm charcoal/brass, scroll-video storytelling |
| `2` | **Aurora** | Futuristic navy/purple, glassmorphism, SaaS-tech motion |
| `3` (default) | **Elysian** | Premium black/gold reservation-led dining |

## Quick start

```bash
cd MonoRepo/template
npm install

# Template 3 — Elysian (default)
npm run dev
# http://localhost:3000/?id=0

# Template 1 — Cinematic
npm run dev:cinematic
# or: TEMPLATE=1 npm run dev
# http://localhost:3000/?id=0&template=1

# Template 2 — Aurora
npm run dev:aurora
# or: TEMPLATE=2 npm run dev
# http://localhost:3000/?id=0

# Template 3 — Elysian
npm run dev:elysian
# or: TEMPLATE=3 npm run dev
# http://localhost:3000/?id=0&template=3
```

## Environment

Set in **MonoRepo root** `.env` and/or **`template/.env.local`**:

```bash
# 1 = Cinematic, 2 = Aurora, 3 = Elysian
TEMPLATE=3
```

Copy from [`template/.env.example`](.env.example). MonoRepo [`.env.example`](../.env.example) also documents `TEMPLATE`.

**Note:** Next.js reads `template/.env.local` automatically. Mirror the MonoRepo root value there for local dev.

## Restaurant switching

Templates are API-only. Use `?id=N` for the public API index,
`restaurant_id=<uuid>` for an admin preview, or the signed `slug` + `token`
pair for a published demo. An API miss fails closed and never falls back to a
bundled scrape fixture.

The Elysian-first order applies to public and admin demo previews. Approved
outreach sequences own their links separately; changing those customer-facing
destinations requires its own reviewed sequence decision.

| URL | Restaurant |
|-----|------------|
| `/?id=0` | First public API restaurant |
| `/?restaurant_id=<uuid>` | Admin/API-backed restaurant preview |
| `/restaurant/0` | Redirects to `/?id=0` |

## Stack

- Next.js 16 (App Router) + TypeScript
- Tailwind CSS v4
- GSAP + ScrollTrigger
- Lenis smooth scrolling
- Framer Motion

## Project structure

```
template/
  src/
    app/                    # Router + layout
    lib/
      templateConfig.ts     # TEMPLATE env reader
      adapters/             # Public/signed API adapters
    templates/
      cinematic/            # Template 1
      aurora/               # Template 2
      elysian/              # Template 3
  legacy/                   # Original static HTML template
```

## Build

```bash
npm run build              # uses TEMPLATE from env (default 3)
npm run build:cinematic    # TEMPLATE=1
npm run build:aurora       # TEMPLATE=2
npm run build:elysian      # TEMPLATE=3
```

## Data sources

- Main Go API public restaurant and signed-demo payloads
- Live attributed Google Places media returned by the API
- Explicitly approved owner/licensed media returned by the API

The production template image does not contain root scrape or OCR fixture JSON.
Legacy static files under `template/legacy` are historical reference only and
must not be deployed as a customer-facing renderer.
