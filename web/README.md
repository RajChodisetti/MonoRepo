# Tuvi web (Next.js)

Marketing site with restaurant search → AI SEO report.

## Getting Started

```bash
cd web
cp .env.example .env.local
# Point MONOREPO_API_URL at the running Go API (Places key lives there)
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

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

Places credentials (`GOOGLE_PLACES_API_KEY` / `PLACES_API`) must be configured on the **Go API**, not this app.

See `.env.example`.

### Smoke check

With MonoRepo API + `npm run dev` running:

```bash
# Search
curl -s "http://localhost:3000/api/restaurants/search?q=sydney" | head -c 400

# Pick a placeId from results, then:
curl -s "http://localhost:3000/api/restaurants/<placeId>" | head -c 400
```

Homepage `ReportMockup` remains a static marketing demo; live reports use `/report/[placeId]`.
