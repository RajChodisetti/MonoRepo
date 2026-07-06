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
```

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

Merged menu photos look like:

```json
"images": {
  "menu_photos": [{ "url": "...", "source": "tripadvisor" }],
  "menu_photos_source": "tripadvisor"
}
```

