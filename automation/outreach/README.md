# Restaurant Lead & Outreach Pipelines

Python pipelines for fetching restaurant leads (Apollo.io), scraping restaurant profiles (Google Places / SerpAPI), filtering no-website restaurants, and drafting outreach emails.

Run all commands from **`MonoRepo/automation/outreach/`**.

## Pipelines

| Script | Purpose |
|--------|---------|
| `fetch_restaurant_leads.py` | Fetch restaurant decision-maker leads from Apollo.io |
| `scrape_restaurant_places.py` | Scrape restaurants via Google Places API |
| `scrape_restaurant_data.py` | Scrape restaurants via SerpAPI (legacy) |
| `scrape_tripadvisor.py` | TripAdvisor scrape (SerpAPI) — menu photos **TripAdvisor-only** |
| `cron_tripadvisor.sh` | Daily cron wrapper for TripAdvisor + merge |
| `verify_leads_from_db.py` | Nightly OCR verification for unverified DB leads |
| `cron_lead_ocr_verify.sh` | Daily cron wrapper for lead OCR verification |
| `city_pipeline.py` | Fetch leads + scrape in one command per city |
| `fetch_restaurants_no_website.py` | Filter scraped JSON for restaurants without a website |
| `tuvi_outreach_agent.py` | Full outreach: scrape sites, draft emails, Zoho/Slack |

## Setup

```bash
cd automation/outreach
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
# Edit .env with your API keys
```

## Quick start — city pipeline

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

## Daily lead ingestion (cron)

Fetches **new** leads only (dedup against DB + local JSON), respects a **500 combined API request** budget (Apollo + Google Places), then imports to Postgres.

```bash
# Manual run (same as cron)
make ingest-daily

# Or directly
LEAD_INGESTION_ENABLED=true python daily_ingestion.py --type restaurant

# Other niches (beta filters)
python daily_ingestion.py --city Sydney --type dentist --max-requests 500
python daily_ingestion.py --city Sydney --type plumber --max-requests 500
```

Cron (daily 02:00):

```bash
chmod +x cron_lead_ingestion.sh
crontab -e
# 0 2 * * * LEAD_INGESTION_ENABLED=true /ABS/PATH/MonoRepo/automation/outreach/cron_lead_ingestion.sh
```

| Env var | Default | Purpose |
|---------|---------|---------|
| `LEAD_INGESTION_ENABLED` | `false` | Gate for cron script |
| `LEAD_INGESTION_MAX_REQUESTS` | `500` | Combined Apollo + Places cap |
| `INGESTION_TYPE` | `restaurant` | Business niche |
| `INGESTION_CITIES` | all 5 AU cities | Comma-separated cities |

State file (Apollo page cursor): `state/ingestion_state.json` (gitignored)

Logs: `logs/lead_ingestion_YYYYMMDD.log`

## Fetch leads only

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

- `leads/` — Apollo lead JSON
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

## Lead OCR verification (cron)

After restaurants are imported into PostgreSQL (`import_to_db.py`), a nightly job can OCR/classify photos, clean dish-card images, sync `menu_images` / `gallery_images`, and mark each profile `ocr_verified=true`.

Requires migration `000013_lead_ocr_verified` and menu OCR API keys (`HUGGING_FACE_API_KEY`, etc. in `backend/.env`).

```bash
# Manual run (from MonoRepo root)
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
| `MENU_OCR_ENABLED` | Reuses existing menu OCR pipeline flags |

Re-importing scrape JSON resets `ocr_verified=false` on that profile.

### Crontab

```bash
chmod +x cron_lead_ocr_verify.sh
crontab -e
# Daily 03:00 — edit path to your checkout:
# 0 3 * * * LEAD_OCR_VERIFICATION_ENABLED=true /ABS/PATH/MonoRepo/automation/outreach/cron_lead_ocr_verify.sh
```

Logs: `logs/lead_ocr_verify_YYYYMMDD.log`

Merged menu photos look like:

```json
"images": {
  "menu_photos": [{ "url": "...", "source": "tripadvisor" }],
  "menu_photos_source": "tripadvisor"
}
```

