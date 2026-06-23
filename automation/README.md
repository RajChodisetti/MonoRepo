# Melbourne Food Image Scraper

Scrapes **Google Images** for popular Melbourne restaurant dishes, downloads the first usable photo per dish, and stores:

- image files in `automation/images/` (named by SHA-256 hash)
- metadata in `automation/data/catalog.json`

## What it does

1. Reads 50 curated dishes from `dishes_melbourne.json` (Australian, Thai, Indian, Italian, and other cuisines common in Melbourne).
2. For each dish, opens **Google Images** in a headless browser with a Melbourne-focused query, e.g. `melbourne thai restaurant pad thai plated`.
3. Downloads the first valid image result.
4. Saves the file as `images/<Dish Name>.jpg` (or `.png` / `.webp`).
5. Appends dish metadata + `image_hash` to `data/catalog.json`.
6. Skips dishes already in `catalog.json` so you can resume if interrupted.

If Google blocks the browser search, the script automatically falls back to DuckDuckGo image search (unless you pass `--google-only`).

## Setup (one time)

From the repo root:

```bash
cd automation
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

**Recommended:** have Google Chrome installed (the script uses it via Playwright).

If you do not have Chrome:

```bash
playwright install chromium
```

## Run

Scrape all dishes not yet in the catalog:

```bash
cd automation
source .venv/bin/activate
python scrape.py
```

### Useful options

Test with only 3 dishes:

```bash
python scrape.py --limit 3
```

Slow down requests (helps avoid Google blocks):

```bash
python scrape.py --delay 5
```

Force Google only (no fallback):

```bash
python scrape.py --google-only
```

Skip Google and use Bing directly (much faster if Google keeps timing out):

```bash
python scrape.py --skip-google
```

Re-scrape everything from scratch:

```bash
python scrape.py --reset
```

## Output

Images land in:

```text
automation/images/Chicken Parma.jpg
automation/images/Pad Thai.webp
```

Metadata lands in:

```text
automation/data/catalog.json
```

Example entry:

```json
{
  "name": "Pad Thai",
  "cuisines": ["Thai", "Noodles"],
  "location": "Melbourne, Australia",
  "search_query": "melbourne thai restaurant pad thai plated",
  "image_hash": "abc123...",
  "image_file": "images/Chicken Parma.jpg",
  "source_url": "https://...",
  "image_source": "google"
}
```

## Notes

- Google requires JavaScript, so the scraper uses a real browser (Chrome or Chromium).
- Scraped images are for **internal demo/seed data**. Review licensing before customer-facing production use.
- If a dish fails, re-run the script — already-scraped dishes are skipped automatically.
