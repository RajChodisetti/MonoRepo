# Service Inventory

This repo contains long-running services, one-shot jobs, static sites, and automation pipelines. Run commands from the repo root unless a service row says otherwise.

## Port Plan

| Port | Service | Notes |
| --- | --- | --- |
| `3000` | Restaurant template Next app | Customer/demo website renderer. Kept on `3000` because main API demo URLs and defaults point here. |
| `3001` | Tuvi corporate website Next app | Moved off `3000` so both websites can run at the same time. |
| `5432` | PostgreSQL | Shared local Docker Postgres. Main app, worker, restaurant data, and company consultations use `restaurant_platform`. |
| `8080` | Main restaurant platform API | Go `net/http` API. |
| `8081` | Swagger UI | Local OpenAPI viewer. |
| `8000` | Voice sales agent | Expected in-repo service at `voice-sales-agent/`; FastAPI/WebSocket voice runtime used by both websites. |
| `5173` | Restaurant services catalog | Vite default dev port. |
| `5500` | Presentation site | Static Python HTTP server. |

## Long-Running Services

| Service | Start | Dependencies | Links |
| --- | --- | --- | --- |
| PostgreSQL | `make db-up` | Docker | Required by main API, worker, migrations, imports, seeders, and company consultations. |
| Main API | `make api` | PostgreSQL + `make migrate-up` | Serves private admin/restaurant APIs, public demo/site/reservation APIs, tracking routes, and campaign controls. |
| Worker | `make worker` | PostgreSQL + `make migrate-up` | Polls leased `job_runs`, creates post-OCR demo/campaign drafts, and runs quota-managed outreach through HTTP API providers. |
| City scrape worker | Compose `scrape-worker` | PostgreSQL, migrations, Places key, Apollo key by default | Long-running durable grid worker; Places first, targeted Apollo second, 500 combined provider calls per 24-hour window. |
| Restaurant template | `cd template && npm run dev` | Node deps; optionally main API and voice agent | Reads local JSON by default; uses `NEXT_PUBLIC_API_URL=http://localhost:8080` for DB-backed site data and `NEXT_PUBLIC_VOICE_AGENT_URL=http://localhost:8000` for voice. |
| Tuvi corporate website | `cd tuvi-website/app && npm run dev` | Node deps; main API for booking | Runs on `3001`, proxies bookings to `CONSULTATION_API_URL=http://localhost:8080`, uses voice agent with `agent=corporate`. |
| Voice sales agent | `cd voice-sales-agent && make dev` or Docker compose | Provider keys and service env | Expected to expose `GET /readyz/browser` and `WS /browser-stream`. Restaurant template passes `restaurant_index`; corporate site passes `agent=corporate`; corporate bookings call main API company consultation endpoints. |
| Restaurant services catalog | `cd apps/restaurant-services-catalog && npm run dev` | Node deps | Standalone Vite site; deploys with Wrangler. |
| Presentation site | `cd presentation && python3 -m http.server 5500` | Python | Standalone static presentation. |

## Stack Commands

| Command | Starts |
| --- | --- |
| `make setup` | PostgreSQL + main migrations |
| `make dev` | PostgreSQL + main migrations + main API |
| `make start` | PostgreSQL + main migrations + main API + worker |
| `make up` | Docker stack profile: PostgreSQL + Redis + migrate + API + worker + durable scrape worker (OCR remains one-shot) |
| `make swagger` | Swagger UI for OpenAPI docs |

## One-Shot Jobs And Pipelines

| Job | Start | Purpose |
| --- | --- | --- |
| Main migrations | `make migrate-up`, `make migrate-down` | Apply or roll back SQL migrations for `restaurant_platform`. |
| Seed admin | `make seed-admin` | Create first internal admin from `ADMIN_*` env. |
| Seed demo fixture | `make seed-demo-fixture` | Create restaurant, owner membership, and published demo site. |
| Seed restaurant data | `make seed-restaurants-data` | Import `data/restaurants_data.json` through Go. |
| Import outreach data | `make import-outreach` | Import automation/outreach restaurant JSON into local DB. |
| Sanitize import | `make sanitize-import` | Strip bad menu-board image matches and import. |
| OCR all | `make ocr-all` | Run menu image OCR/classification and import. |
| Verify leads OCR | `make verify-leads-ocr` | OCR-verify unverified `restaurant_profiles` rows in DB (batch/cron). |
| Database-networked OCR job | `make ocr-job` | One-shot claimed OCR batch; resolves Places photo resources, writes explicit OCR state, and enqueues `lead.prepare`. |
| Retired daily lead ingestion | `make ingest-daily` | Legacy Places/Apollo entrypoint retained for reference; do not run after durable scrape migrations are installed. |
| Legacy outreach city pipeline | `cd automation/outreach && python city_pipeline.py --city Sydney --total 100` | Manual Apollo + Places workflow. |
| Legacy fetch leads | `cd automation/outreach && python fetch_restaurant_leads.py --city Sydney` | Manual Apollo restaurant decision-maker leads. |
| Scrape places | `cd automation/outreach && python scrape_restaurant_places.py --city Sydney --total 100` | Google Places API (New) enrichment for existing lead files. |
| Legacy SerpAPI scrape | `cd automation/outreach && python scrape_restaurant_data.py --city Sydney` | Older SerpAPI scrape path. |
| Filter no-website leads | `cd automation/outreach && python fetch_restaurants_no_website.py --all-cities` | Filter scraped output. |
| Outreach drafts | `cd automation/outreach && python tuvi_outreach_agent.py --csv sample_leads.csv --no-zoho --no-slack` | Draft outreach without external writes. |
| Melbourne image scraper | `cd automation && python scrape.py` | Browser-based food image collection. |
| OpenAPI validation | `make openapi` | Validate `docs/openapi/openapi.yaml`. |

## Interlinks

```text
main API :8080
  -> PostgreSQL restaurant_platform
  -> job_runs table for worker jobs
  -> public restaurant/site/demo data consumed by template :3000

worker
  -> PostgreSQL job_runs
  -> draft demo/campaign preparation after verified OCR
  -> Gmail API or Zoho quota-managed outreach adapter (disabled by default)

scrape-worker
  -> PostgreSQL scrape_jobs/cells/candidates
  -> Google Places API (New)
  -> targeted Apollo owner/work-email enrichment

scheduled ocr-job
  -> PostgreSQL pending restaurant_profiles
  -> short-lived Places photo media + configured vision provider
  -> lead.prepare job_runs -> worker

template :3000
  -> optional main API :8080
  -> voice-sales-agent :8000
       /readyz/browser
       /browser-stream?restaurant_index=N

tuvi corporate website :3001
  -> main API :8080 /api/v1/company/consultations/*
  -> voice-sales-agent :8000
       /readyz/browser
       /browser-stream?agent=corporate

main API company consultations
  -> PostgreSQL restaurant_platform.company_consultations
  -> optional Google Calendar
  -> configured Resend or Zoho generic HTTP API provider
  <- voice-sales-agent corporate flow via MONOREPO_API_URL=http://localhost:8080
```

## Notes

- `apps/web` is currently only a placeholder for the future Phase 1 dashboard.
- The voice agent source lives at `voice-sales-agent/`; use `make voice-up` from the MonoRepo root for the Docker profile.
- `tuvi-website/backend` is legacy reference code. Normal runtime uses the main API for consultation scheduling.
