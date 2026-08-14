# Restaurant Lead & Outreach Pipelines

Python pipelines for discovering businesses with Google Places API (New), enriching missing owner/work-email details with Apollo, filtering no-website restaurants, and drafting outreach emails. Broad Apollo discovery and SerpAPI scripts remain available as legacy/manual tools.

Run all commands from **`MonoRepo/automation/outreach/`**.

## Pipelines

| Script | Purpose |
|--------|---------|
| `city_scrape_worker.py` | Primary durable PostgreSQL city worker: grid Places discovery → targeted Apollo → direct import |
| `daily_ingestion.py` | Legacy one-shot Places/Apollo import; do not schedule beside the durable worker |
| `cron_lead_ingestion.sh` | Legacy disabled-by-default wrapper; keep removed from production crontab |
| `fetch_restaurant_leads.py` | Legacy/manual Apollo decision-maker fetch |
| `scrape_restaurant_places.py` | Enrich existing lead files via Google Places API (New) |
| `scrape_restaurant_data.py` | Scrape restaurants via SerpAPI (legacy) |
| `city_pipeline.py` | Fetch leads + scrape in one command per city |
| `fetch_restaurants_no_website.py` | Filter scraped JSON for restaurants without a website |
| `tuvi_outreach_agent.py` | Legacy/manual draft tooling; production sends use the Go quota workflow |

## Setup

```bash
cd automation/outreach
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
# Edit .env with your API keys
```

## Legacy/manual city pipeline (Apollo + Places)

```bash
# Full pipeline: fetch 100 leads + scrape for Sydney
python city_pipeline.py --city Sydney --total 100

# Multiple cities
python city_pipeline.py --cities Sydney Melbourne Perth --total 100

# Fetch leads only
python city_pipeline.py --city Melbourne --total 100 --fetch-only

# Scrape only (leads already in leads/lead_<city>.json)
python city_pipeline.py --city Perth --total 100 --scrape-only

# Budget-aware mode (500 combined Apollo + Places requests)
python city_pipeline.py --city Sydney --type restaurant --max-requests 500
```

## Primary durable city ingestion

Production uses the private `POST /api/v1/scrape-jobs` API and the long-running
`scrape-worker` Compose service. It persists grid cells, page tokens, Place-ID
candidates, a combined 500-call window, and a 24-hour `resume_at`; completed
coverage cycles revisit the city for newly added Place IDs. See
`docs/runbooks/lead-scrape-outreach.md` from the repository root.

Apollo is optional enrichment in this durable path. When it is disabled,
misconfigured, unavailable, or rejects a request, the worker records/skips the
enrichment and continues importing verified Google Places data. The shared
request ceiling still pauses all provider work because no further Places calls
are allowed after that ceiling. Failed jobs can be deliberately resumed through
`POST /api/v1/scrape-jobs/{id}/resume`; completed cells and imported candidates
remain intact.

## Retired one-shot ingestion

`daily_ingestion.py`, `cron_lead_ingestion.sh`, and `make ingest-daily` are
retained only to fail closed and direct operators to the durable API. Once
migration `000015` is present, `daily_ingestion.py` refuses provider calls so
it cannot bypass the PostgreSQL request ledger. Do not install its cron entry.
Trigger all production city work through `POST /api/v1/scrape-jobs`.

| Env var | Default | Purpose |
|---------|---------|---------|
| `LEAD_INGESTION_ENABLED` | `false` | Gate for cron script |
| `LEAD_INGESTION_MAX_REQUESTS` | `500` | Combined Places + Apollo request cap |
| `INGESTION_TYPE` | `restaurant` | Business niche |
| `INGESTION_CITIES` | all 5 AU cities | Comma-separated cities |
| `INGESTION_ENV_FILE` | unset | Protected host-side env file loaded after local defaults |
| `DATABASE_URL` | required | PostgreSQL source of truth and fail-closed dedup store |
| `PLACES_API` / `GOOGLE_PLACES_API_KEY` | required | Places API (New) credential |
| `APOLLO_API_KEY` | optional | Targeted owner/work-email enrichment credential; a missing key falls back to Places-only import |
| `APOLLO_ENRICHMENT_ENABLED` | `true` | Run Apollo after Places for missing contact fields |

Run summaries: `state/ingestion_state.json` (gitignored)

Logs: `logs/lead_ingestion_YYYYMMDD.log`

## Legacy Apollo fetch

```bash
python fetch_restaurant_leads.py --city Sydney
python fetch_restaurant_leads.py --cities Sydney Melbourne
python fetch_restaurant_leads.py   # all 5 Australian cities → leads/lead.json
```

## Scrape restaurant data

```bash
# Google Places (recommended)
python scrape_restaurant_places.py --city Sydney --total 100

# SerpAPI (legacy)
python scrape_restaurant_data.py --city Sydney
```

## Filter no-website restaurants

```bash
python fetch_restaurants_no_website.py --all-cities
python fetch_restaurants_no_website.py --input data/restaurants_data_sydney.json
```

## Outreach (email drafts)

```bash
# Dry-run — local drafts only
python tuvi_outreach_agent.py --csv sample_leads.csv --no-zoho --no-slack

# Fetch leads + draft emails
python tuvi_outreach_agent.py --fetch-leads --run-outreach --no-zoho --no-slack
```

## Output directories

Generated at runtime (gitignored):

- `leads/` — legacy/manual Apollo lead JSON
- `data/` — scraped restaurant profiles
- `output/` — filtered subsets (e.g. no-website)
- `drafts/` — email draft backups
- `reports/` — run summaries

Seed data for the backend lives separately at `MonoRepo/data/restaurants_data.json`.

## Outreach enrollment after import

Every imported restaurant with a non-empty name and valid business email is
recorded as an `inferred_business` lead with its source evidence and
idempotently enrolled in the currently approved plain-text outreach sequence.
This path does not require image analysis, a generated demo, or a reviewed
profile. Delivery remains disabled until an administrator explicitly enables
the persisted email job.

Google listing media stays live and attributed; scraped Google, review, and
menu images are removed before raw or structured data is persisted. The former
TripAdvisor menu-photo scraper and cron wrapper are retired. Owner/licensed
uploads require explicit admin approval before public use.
