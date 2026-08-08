# Tuvi corporate website (Next.js)

Canonical Tuvi marketing site with restaurant search → AI SEO report, stable
legal pages, consultation booking, and the corporate browser voice experience.

## Getting Started

```bash
cd web
cp .env.example .env.local
# Point MONOREPO_API_URL and CONSULTATION_API_URL at the main Go API.
npm install
npm run dev -- -p 3001
```

Open [http://localhost:3001](http://localhost:3001).

## Restaurant search → AI report

1. On the hero, type a restaurant name (AU by default).
2. Next BFF proxies to MonoRepo:
   - `GET {MONOREPO_API_URL}/api/public/v1/seo/search?q=`
   - `GET {MONOREPO_API_URL}/api/public/v1/seo/report/{placeId}`
3. Select a result or submit → `/report/[placeId]`
4. Report page loads `/api/restaurants/[placeId]` (proxy), shows the 100-point SEO breakdown + AI summary, and offers email capture via `POST /api/leads`.

### Env (server-only)

| Variable | Purpose |
| --- | --- |
| `MONOREPO_API_URL` | MonoRepo Go API base (default `http://localhost:8080`) |
| `CONSULTATION_API_URL` | Optional consultation-specific API base; defaults to `MONOREPO_API_URL` |
| `TUVI_API_TOKEN` | Server-only bearer token for company consultation endpoints |
| `NEXT_PUBLIC_CONTACT_EMAIL` | Public support/contact address |
| `NEXT_PUBLIC_VOICE_AGENT_URL` | Public browser voice endpoint |
| `VOICE_AGENT_URL` | Optional server-only voice endpoint override |

Places credentials (`GOOGLE_PLACES_API_KEY` / `PLACES_API`) must be configured on the **Go API**, not this app.

See `.env.example`.

## Consultation booking

`/book` loads available times through the same-origin
`/api/consultations/availability` handler and saves confirmed bookings through
`/api/consultations`. Both handlers attach `TUVI_API_TOKEN` only on the server.
The Go API and PostgreSQL slot ledger are authoritative: confirmed rows are
removed from later availability responses, and no Google Calendar is queried or
updated. `/demo` redirects to `/book` for compatibility with older campaign and
navigation links.

Stable public contract routes are `/privacy`, `/terms`, and
`/google-workspace`.

## Smoke check

With the MonoRepo API and `npm run dev -- -p 3001` running:

```bash
# Search
curl -s "http://localhost:3001/api/restaurants/search?q=sydney" | head -c 400

# Pick a placeId from results, then:
curl -s "http://localhost:3001/api/restaurants/<placeId>" | head -c 400

# Booking route and stable public pages
curl -I http://localhost:3001/book
curl -I http://localhost:3001/privacy
curl -I http://localhost:3001/terms
curl -I http://localhost:3001/google-workspace
```

Homepage `ReportMockup` remains a static marketing demo; live reports use `/report/[placeId]`.
