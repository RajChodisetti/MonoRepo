# Restaurant Website Templates

Premium Next.js templates sharing the same restaurant payload, switched via `TEMPLATE` env variable or `?template=`.

| `TEMPLATE` | Name | Style |
|------------|------|-------|
| `1` (default) | **Cinematic** | Warm charcoal/brass, scroll-video storytelling |
| `2` | **Aurora** | Futuristic navy/purple, glassmorphism, SaaS-tech motion |
| `3` | **Elysian** | Premium black/gold reservation-led dining |

## Quick start

```bash
cd MonoRepo/template
npm install

# Template 1 — Cinematic
npm run dev
# http://localhost:3000/?id=0

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
TEMPLATE=1
```

Copy from [`template/.env.example`](.env.example). MonoRepo [`.env.example`](../.env.example) also documents `TEMPLATE`.

**Note:** Next.js reads `template/.env.local` automatically. Mirror the MonoRepo root value there for local dev.

## Restaurant switching

Templates use `?id=N` (0-based index into `../data/restaurants_data.json`) or `restaurant_id=<uuid>` for API-backed generated sites:

| URL | Restaurant |
|-----|------------|
| `/?id=0` | Bistro Moncur |
| `/?id=1` | Second restaurant |
| `/restaurant/0` | Redirects to `/?id=0` |

## Stack

- Next.js 15 (App Router) + TypeScript
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
      adapters/             # Shared JSON adapter
    templates/
      cinematic/            # Template 1
      aurora/               # Template 2
      elysian/              # Template 3
  legacy/                   # Original static HTML template
```

## Build

```bash
npm run build              # uses TEMPLATE from env (default 1)
npm run build:cinematic    # TEMPLATE=1
npm run build:aurora       # TEMPLATE=2
npm run build:elysian      # TEMPLATE=3
```

## Legacy static template

```bash
cd MonoRepo && python3 -m http.server 8080
# http://localhost:8080/template/legacy/index.html?id=0
```

## Data sources

- `../data/restaurants_data.json`
- `../data/image_classifications.json`
