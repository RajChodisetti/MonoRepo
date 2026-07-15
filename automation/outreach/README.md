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
| `scrape_tripadvisor.py` | TripAdvisor scrape (SerpAPI) — menu photos **TripAdvisor-only** |
| `cron_tripadvisor.sh` | Daily cron wrapper for TripAdvisor + merge |
| `verify_leads_from_db.py` | Claimed OCR state machine for pending DB leads |
| `cron_lead_ocr_verify.sh` | Legacy/local OCR wrapper; production uses the Compose `ocr-job` |
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
`docs/runbooks/lead-scrape-ocr-outreach.md` from the repository root.

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
| `APOLLO_API_KEY` | required by default | Targeted owner/work-email enrichment credential |
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

## TripAdvisor menu photos (cron)

Menu photos used for outreach cards must come from **TripAdvisor only** (not Google / Yelp).

Cities match the default Google / Places geo: **Sydney, Melbourne, Perth, Adelaide, Brisbane**.

```bash
# One city (write data/restaurants_data_<city>_tripadvisor.json)
python scrape_tripadvisor.py --city Sydney --limit 100

# All default cities
python scrape_tripadvisor.py --all-cities --limit 100

# Merge TripAdvisor menu_photos into Google scrape files
# (sets images.menu_photos_source = "tripadvisor")
python scrape_tripadvisor.py --all-cities --limit 100 --merge

# Long-running process (no crontab)
python scrape_tripadvisor.py --all-cities --merge --schedule daily
```

### Crontab

```bash
chmod +x cron_tripadvisor.sh
crontab -e
# Daily 02:15 — edit path to your checkout:
# 15 2 * * * /ABS/PATH/MonoRepo/automation/outreach/cron_tripadvisor.sh
```

Logs: `logs/tripadvisor_cron_YYYYMMDD.log`

## Lead OCR verification

After restaurants are imported into PostgreSQL, a scheduled job claims
`pending` profiles, OCR/classifies trusted images, syncs menu/gallery data, and
finishes as `verified`, `no_images`, or `failed`. Only `verified` queues the
idempotent `lead.prepare` job that creates reviewable demo and campaign drafts.

Requires migrations `000016_ocr_status_and_post_ocr` and
`000022_auto_artifact_profile_provenance`, plus a vision-provider key
(`HUGGING_FACE_API_KEY`, `OPENAI_API_KEY`, or `GEMINI_API_KEY`).

The direct commands below are for controlled local/manual operation only:

```bash
# Manual/local run (from MonoRepo root)
make verify-leads-ocr

# Or directly
cd automation/outreach
LEAD_OCR_VERIFICATION_ENABLED=true python verify_leads_from_db.py --force --dry-run
LEAD_OCR_VERIFICATION_ENABLED=true python verify_leads_from_db.py --force --limit 5
```

| Env | Description |
|-----|-------------|
| `LEAD_OCR_VERIFICATION_ENABLED` | `true` enables cron + script (default `false`) |
| `LEAD_OCR_BATCH_SIZE` | Max restaurants per run (default `50`) |
| `LEAD_OCR_MAX_ATTEMPTS` | Maximum automatic attempts for one unchanged OCR input (default `3`) |
| `LEAD_OCR_RETRY_AFTER_HOURS` | Cooldown before retrying a failed OCR attempt (default `24`) |
| `MENU_OCR_ENABLED` | Reuses existing menu OCR pipeline flags |

OCR refreshes current photo resource names from Place Details immediately
before analysis; expirable resource names and short-lived Photo Media URLs are
never retained in durable lead data. A failed unchanged input retries after the
configured cooldown until the maximum attempt count. After that, inspect the
record and explicitly requeue it as described in the production runbook.

When a later import changes the OCR image fingerprint, the profile returns to
`pending`, stale human approvals are cleared, and the automatic draft may be
refreshed only after OCR verifies the new inputs.

An unchanged OCR fingerprint does not hide a changed restaurant identity or
generated profile/menu payload. That source change receives its own durable
`lead.prepare` key, returns automatic artifacts to draft, and refreshes their
profile provenance before they can be reviewed again.

Do not install `cron_lead_ocr_verify.sh` in production. Production runs the
one-shot Compose `ocr-job` under a non-overlapping `flock` schedule so each run
has the intended image resolver and isolated dependencies. Use the staged OCR
rollout and exact locked command in
`docs/runbooks/lead-scrape-ocr-outreach.md`.

Merged menu photos look like:

```json
"images": {
  "menu_photos": [{ "url": "...", "source": "tripadvisor" }],
  "menu_photos_source": "tripadvisor"
}
```
