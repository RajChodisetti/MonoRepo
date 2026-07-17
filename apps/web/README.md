# Tuvi Admin Portal (`apps/web`)

Next.js internal_admin console for the MonoRepo lead workflow: scrape jobs, restaurant review, demos/campaigns, and bulk outreach.

## Stack

- Next.js 15 (App Router) + React 19 + TypeScript + Tailwind 4
- Runs on **port 3002**
- Talks to the Go API via same-origin BFF routes (`/api/admin/*`) with an httpOnly session cookie

## Setup

```bash
cd apps/web
cp .env.example .env.local
# edit API_BASE_URL if needed
npm install
npm run dev
```

Open http://localhost:3002/login

### Env

| Variable | Default | Notes |
|----------|---------|--------|
| `API_BASE_URL` | `https://api.tuvisolutions.com` | No trailing slash. Local API: `http://localhost:8080` |

Login requires an `internal_admin` user (seeded with `make seed-admin` / production admin).

## CORS note

The Go API default CORS allowlist includes `localhost:3000` and `:3001`. This portal uses **3002** and proxies through Next, so browser CORS to the Go API is usually not required for admin calls.

If you call the Go API directly from the browser later, add:

```bash
CORS_ALLOWED_ORIGINS=...,http://localhost:3002,http://127.0.0.1:3002
```

to local `.env` / VM `stack.env`.

## Screens

| Route | Purpose |
|-------|---------|
| `/login` | Admin sign-in |
| `/dashboard` | Scrape + outreach overview |
| `/scrape-jobs` | Trigger / list / poll / retry |
| `/scrape-jobs/[id]` | Live progress |
| `/restaurants` | Filterable lead table |
| `/restaurants/[id]` | Overview, profile OCR review, demo, campaign, members |
| `/outreach` | Bulk-send trigger + status |

### Lead workflow order

1. Scrape city → restaurants appear  
2. OCR verified (worker/job)  
3. Profile approve  
4. Create + publish demo  
5. Create + approve campaign  
6. Outreach → Start bulk send  

## Scripts

```bash
npm run dev    # http://localhost:3002
npm run build
npm run start
npm run lint
```
