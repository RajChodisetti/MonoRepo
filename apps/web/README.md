# Tuvi Admin Portal (`apps/web`)

Next.js `internal_admin` console for the MonoRepo lead workflow: scrape jobs,
restaurant review, demos, and plain-text outreach sequences.

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
| `/restaurants/[id]` | Overview, contact details, approved media, demo, and members |
| `/outreach` | Sequence editor, recipient progress, email job, and sender health |

### Lead workflow order

1. Scrape city → restaurants with a name and valid business email are enrolled.
2. Confirm the imported lead has recorded `inferred_business` source evidence.
3. Review or edit the plain-text outreach sequence, then approve its version.
4. Review recipient progress and sender health.
5. Explicitly enable the email job when real sending is authorized.

Due follow-ups are selected before new restaurants. Confirmed provider delivery
advances the recipient's integer sequence step; failed or unknown outcomes do
not. Interest pauses automation, while lost, archived, onboarding, and active
client restaurants are excluded.

## Scripts

```bash
npm run dev    # http://localhost:3002
npm run build
npm run start
npm run lint
```
