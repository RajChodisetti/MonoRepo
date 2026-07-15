#!/usr/bin/env python3
"""
TripAdvisor restaurant scraper (SerpAPI) — cron-friendly.

Same geo locations as Google / Places (AU major cities by default).
Menu photos are sourced ONLY from TripAdvisor (never Google / Yelp).

Usage (from automation/outreach/):
  python scrape_tripadvisor.py --city Sydney
  python scrape_tripadvisor.py --cities Sydney Melbourne --limit 20
  python scrape_tripadvisor.py --all-cities          # all resolve_cities() defaults
  python scrape_tripadvisor.py --all-cities --merge  # attach menu_photos into Google scrape JSONs
  python scrape_tripadvisor.py --schedule daily      # long-running daily cron loop
"""

from __future__ import annotations

import argparse
import json
import logging
import re
import sys
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

import requests

from env_loader import load_project_env

load_project_env()

from tuvi_outreach_agent import (  # noqa: E402
    AUSTRALIAN_RESTAURANT_CITIES,
    Config,
    city_slug,
    normalize_city_name,
    resolve_cities,
    with_retry,
)

log = logging.getLogger("tripadvisor_scraper")

SERPAPI_URL = "https://serpapi.com/search.json"
TA_DOMAIN_AU = "www.tripadvisor.com.au"

# SerpAPI tripadvisor ssrc: r = Restaurants
TA_SSRC_RESTAURANTS = "r"

MENU_PHOTO_KEYWORDS = (
    "menu",
    "carte",
    "food menu",
    "dinner menu",
    "lunch menu",
    "wine list",
    "beverage",
    "dish",
    "cuisine",
)


# ─────────────────────────────────────────────────────────────
# Paths
# ─────────────────────────────────────────────────────────────

def default_tripadvisor_output_path(city: str | None, cfg: Config | None = None) -> Path:
    """data/tripadvisor_restaurants_<city>.json"""
    cfg = cfg or Config()
    if city:
        return Path("data") / f"tripadvisor_restaurants_{city_slug(normalize_city_name(city))}.json"
    return Path("data") / "tripadvisor_restaurants.json"


def default_google_scrape_path(city: str | None) -> Path:
    """Existing Google/SerpAPI scrape JSON to merge menu_photos into."""
    if city:
        return Path("data") / f"restaurants_data_{city_slug(normalize_city_name(city))}.json"
    return Path("data") / "restaurants_data.json"


# ─────────────────────────────────────────────────────────────
# SerpAPI
# ─────────────────────────────────────────────────────────────

def _serpapi_get(params: dict, cfg: Config, label: str = "serpapi") -> dict:
    params = {**params, "api_key": cfg.SERPAPI_KEY}

    def _call():
        resp = requests.get(SERPAPI_URL, params=params, timeout=max(cfg.SCRAPE_TIMEOUT, 30))
        if resp.status_code == 429:
            raise RuntimeError("HTTP 429 rate limited — will retry")
        if resp.status_code != 200:
            raise RuntimeError(f"HTTP {resp.status_code}: {resp.text[:300]}")
        data = resp.json()
        if data.get("error"):
            raise RuntimeError(data["error"])
        return data

    return with_retry(_call, cfg.RETRY_ATTEMPTS, cfg.RETRY_BACKOFF, label=label)


def _is_menu_photo(title: str, url: str = "") -> bool:
    blob = f"{title} {url}".lower()
    return any(kw in blob for kw in MENU_PHOTO_KEYWORDS)


def _photo_entry(url: str, *, title: str = "", thumbnail: str = "") -> dict | None:
    url = (url or "").strip()
    if not url or not url.startswith("http"):
        return None
    return {
        "url": url,
        "thumbnail": thumbnail or url,
        "title": title or "",
        "source": "tripadvisor",
    }


def _extract_photos_from_result(row: dict) -> tuple[list[dict], list[dict]]:
    """Return (gallery, menu_photos) — menu_photos ONLY TripAdvisor images tagged as menu."""
    gallery: list[dict] = []
    menu_photos: list[dict] = []
    seen: set[str] = set()

    def add(url: str, title: str = "", thumbnail: str = ""):
        entry = _photo_entry(url, title=title, thumbnail=thumbnail)
        if not entry:
            return
        key = entry["url"]
        if key in seen:
            return
        seen.add(key)
        if _is_menu_photo(title, url):
            menu_photos.append(entry)
        else:
            gallery.append(entry)

    thumb = row.get("thumbnail") or row.get("thumbnail_url") or ""
    if thumb:
        add(thumb, title=row.get("title") or "")

    for key in ("photos", "images", "photo"):
        block = row.get(key)
        if not block:
            continue
        if isinstance(block, dict):
            block = [block]
        if not isinstance(block, list):
            continue
        for img in block:
            if isinstance(img, str):
                add(img)
            elif isinstance(img, dict):
                add(
                    img.get("url") or img.get("image") or img.get("original") or "",
                    title=img.get("title") or img.get("description") or img.get("caption") or "",
                    thumbnail=img.get("thumbnail") or "",
                )

    # Reviews sometimes carry diner/menu photos
    for rev in row.get("reviews") or row.get("user_reviews") or []:
        if not isinstance(rev, dict):
            continue
        for img in rev.get("images") or rev.get("photos") or []:
            if isinstance(img, str):
                add(img, title=rev.get("title") or "review")
            elif isinstance(img, dict):
                add(
                    img.get("url") or img.get("image") or "",
                    title=img.get("title") or rev.get("title") or "review",
                    thumbnail=img.get("thumbnail") or "",
                )

    return gallery, menu_photos


def _location_id_from_row(row: dict) -> str:
    for key in ("location_id", "locationId", "id"):
        val = row.get(key)
        if val is not None and str(val).strip():
            return str(val).strip()
    link = row.get("link") or row.get("url") or ""
    m = re.search(r"-d(\d+)", link)
    return m.group(1) if m else ""


def search_tripadvisor_restaurants(
    city: str,
    cfg: Config,
    *,
    limit: int = 100,
) -> list[dict]:
    """Paginate TripAdvisor restaurant listings for a city."""
    city_name = normalize_city_name(city).split(",")[0].strip()
    # SerpAPI example: q=<city>&ssrc=r (Restaurants filter)
    query = city_name
    collected: list[dict] = []
    offset = 0
    page = 0
    max_pages = max(3, (limit // 30) + 2)

    while len(collected) < limit and page < max_pages:
        params: dict[str, Any] = {
            "engine": "tripadvisor",
            "q": query,
            "ssrc": TA_SSRC_RESTAURANTS,
            "tripadvisor_domain": TA_DOMAIN_AU,
            "limit": min(30, limit - len(collected)),
        }
        if offset:
            params["offset"] = offset

        try:
            data = _serpapi_get(params, cfg, label=f"ta-search:{city_name}:p{page}")
        except Exception as exc:
            log.warning("TripAdvisor search failed for %s page %s: %s", city_name, page, exc)
            break

        rows = (
            data.get("locations")
            or data.get("local_results")
            or data.get("organic_results")
            or data.get("restaurants")
            or []
        )
        if not rows:
            log.info("No more TripAdvisor results for %s (page %s)", city_name, page)
            break

        for row in rows:
            if not isinstance(row, dict):
                continue
            title = (row.get("title") or row.get("name") or "").strip()
            if not title:
                continue
            collected.append(row)
            if len(collected) >= limit:
                break

        page += 1
        # SerpAPI tripadvisor often paginates with offset of 30
        offset = len(collected)
        pagination = data.get("serpapi_pagination") or {}
        if pagination.get("next_offset") is not None:
            offset = int(pagination["next_offset"])
        elif not pagination.get("next"):
            # no explicit next page — stop after this pack if short page
            if len(rows) < 10:
                break

        time.sleep(cfg.SCRAPE_DELAY)

    return collected[:limit]


def fetch_tripadvisor_place_photos(location_id: str, cfg: Config) -> list[dict]:
    """TripAdvisor Place API — richer photo set for one restaurant."""
    if not location_id:
        return []
    params = {
        "engine": "tripadvisor_place",
        "place_id": location_id,
        "tripadvisor_domain": TA_DOMAIN_AU,
    }
    try:
        data = _serpapi_get(params, cfg, label=f"ta-place:{location_id}")
    except Exception as exc:
        log.debug("Place photos skip %s: %s", location_id, exc)
        return []

    photos: list[dict] = []
    seen: set[str] = set()

    # Place payload may nest photos under place_results or top-level
    place = data.get("place_results") or data
    candidates: list[Any] = []
    for key in ("photos", "images", "photo"):
        block = place.get(key) if isinstance(place, dict) else None
        if isinstance(block, list):
            candidates.extend(block)
        elif isinstance(block, dict):
            candidates.append(block)

    for album in place.get("photo_albums") or place.get("albums") or []:
        if not isinstance(album, dict):
            continue
        album_title = album.get("title") or album.get("name") or ""
        for img in album.get("photos") or album.get("images") or []:
            if isinstance(img, dict):
                img = {**img, "title": img.get("title") or album_title}
            candidates.append(img)

    for img in candidates:
        if isinstance(img, str):
            url, title, thumb = img, "", img
        elif isinstance(img, dict):
            url = img.get("url") or img.get("image") or img.get("original") or ""
            # dynamic-media CDN URLs often embed {width}/{height}
            url = url.replace("{width}", "1200").replace("{height}", "900")
            title = img.get("title") or img.get("description") or img.get("caption") or ""
            thumb = img.get("thumbnail") or url
        else:
            continue
        if not url or url in seen:
            continue
        seen.add(url)
        entry = _photo_entry(url, title=title, thumbnail=thumb)
        if entry:
            photos.append(entry)

    return photos


def fetch_tripadvisor_reviews_photos(
    location_id: str,
    cfg: Config,
    *,
    max_pages: int = 2,
) -> list[dict]:
    """Pull photos attached to TripAdvisor reviews (includes menu uploads)."""
    if not location_id:
        return []

    photos: list[dict] = []
    seen: set[str] = set()
    offset = 0

    for page in range(max_pages):
        params: dict[str, Any] = {
            "engine": "tripadvisor_reviews",
            "location_id": location_id,
            "tripadvisor_domain": TA_DOMAIN_AU,
        }
        if offset:
            params["offset"] = offset

        try:
            data = _serpapi_get(params, cfg, label=f"ta-reviews:{location_id}:p{page}")
        except Exception as exc:
            log.debug("Reviews photos skip %s: %s", location_id, exc)
            break

        for rev in data.get("reviews") or []:
            if not isinstance(rev, dict):
                continue
            for img in rev.get("images") or rev.get("photos") or []:
                if isinstance(img, str):
                    url, title, thumb = img, "review", img
                else:
                    url = img.get("url") or img.get("image") or ""
                    title = img.get("title") or "review"
                    thumb = img.get("thumbnail") or url
                if not url or url in seen:
                    continue
                seen.add(url)
                entry = _photo_entry(url, title=title, thumbnail=thumb)
                if entry:
                    photos.append(entry)

        pagination = data.get("serpapi_pagination") or {}
        next_off = pagination.get("next_offset")
        if next_off is None:
            break
        offset = int(next_off)
        time.sleep(cfg.SCRAPE_DELAY)

    return photos


def normalize_tripadvisor_restaurant(row: dict, city: str, cfg: Config) -> dict:
    """Map SerpAPI TripAdvisor row → restaurants_data-compatible record."""
    city_full = normalize_city_name(city)
    city_only = city_full.split(",")[0].strip()
    location_id = _location_id_from_row(row)
    gallery, menu_photos = _extract_photos_from_result(row)

    # Supplement from Place albums + review uploads (TripAdvisor only)
    extra_photos = fetch_tripadvisor_place_photos(location_id, cfg)
    extra_photos.extend(fetch_tripadvisor_reviews_photos(location_id, cfg))
    seen_menu = {p["url"] for p in menu_photos}
    seen_gallery = {g["url"] for g in gallery}
    for p in extra_photos:
        url = p.get("url") or ""
        if not url:
            continue
        if _is_menu_photo(p.get("title") or "", url):
            if url not in seen_menu:
                menu_photos.append(p)
                seen_menu.add(url)
        elif url not in seen_gallery:
            gallery.append(p)
            seen_gallery.add(url)

    # TripAdvisor THUMB as gallery if nothing else
    if not gallery and row.get("thumbnail"):
        entry = _photo_entry(row["thumbnail"], title=row.get("title") or "")
        if entry:
            gallery.append(entry)

    cuisines: list[str] = []
    for key in ("cuisine", "cuisines", "type"):
        val = row.get(key)
        if isinstance(val, list):
            cuisines.extend(str(x) for x in val if x)
        elif isinstance(val, str) and val:
            cuisines.extend([c.strip() for c in val.split(",") if c.strip()])

    address = row.get("address") or row.get("location") or ""
    if isinstance(address, dict):
        address = address.get("address") or address.get("display") or ""

    link = row.get("link") or row.get("url") or ""
    if link and not link.startswith("http"):
        link = f"https://{TA_DOMAIN_AU}{link}"

    return {
        "name": (row.get("title") or row.get("name") or "").strip(),
        "cuisines": list(dict.fromkeys(cuisines)),
        "rating": row.get("rating") or row.get("rating_value"),
        "reviews_count": row.get("reviews") or row.get("reviews_count") or row.get("num_reviews"),
        "contact": {
            "phone": row.get("phone") or "",
            "email": "",
            "website": row.get("website") or "",
        },
        "owners": [],
        "location": {
            "address": address if isinstance(address, str) else "",
            "city": city_only,
            "state": "",
            "country": "Australia",
            "coordinates": {},
        },
        "menu_items": [],
        "reviews": [],
        "images": {
            "thumbnail": row.get("thumbnail") or "",
            "gallery": gallery,
            # HARD RULE: menu_photos only from TripAdvisor
            "menu_photos": menu_photos,
            "menu_photos_source": "tripadvisor",
        },
        "hours": {},
        "price_level": row.get("price_level") or row.get("price") or "",
        "tripadvisor": {
            "location_id": location_id,
            "link": link,
            "domain": TA_DOMAIN_AU,
        },
        "scrape_status": "success" if menu_photos or gallery or location_id else "partial",
        "errors": [],
        "source": "tripadvisor",
    }


def _json_size_bytes(obj: dict) -> int:
    return len(json.dumps(obj, ensure_ascii=False).encode("utf-8"))


def _build_document(restaurants: list[dict], meta_extra: dict) -> dict:
    return {
        "meta": {
            "version": "1.0",
            "scraped_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "source": "tripadvisor_serpapi",
            "menu_photos_source": "tripadvisor_only",
            "data_fields": [
                "name", "cuisines", "rating", "contact", "images.menu_photos",
            ],
            **meta_extra,
        },
        "restaurants": restaurants,
    }


# ─────────────────────────────────────────────────────────────
# Pipeline
# ─────────────────────────────────────────────────────────────

def run_tripadvisor_scrape_for_city(
    city: str,
    cfg: Config | None = None,
    *,
    limit: int = 100,
    output_path: Path | None = None,
) -> tuple[list[dict], Path]:
    cfg = cfg or Config()
    if not cfg.SERPAPI_KEY:
        raise ValueError("SERPAPI_KEY is missing from .env")

    city_full = normalize_city_name(city)
    output_path = output_path or default_tripadvisor_output_path(city_full, cfg)
    output_path.parent.mkdir(parents=True, exist_ok=True)

    log.info(
        "\n%s\n  TRIPADVISOR SCRAPE — %s  |  limit %s\n%s",
        "═" * 60,
        city_full,
        limit,
        "═" * 60,
    )

    rows = search_tripadvisor_restaurants(city_full, cfg, limit=limit)
    restaurants: list[dict] = []

    for idx, row in enumerate(rows, start=1):
        name = (row.get("title") or row.get("name") or "Unknown").strip()
        log.info("[%s/%s] TripAdvisor: %s", idx, len(rows), name)
        try:
            record = normalize_tripadvisor_restaurant(row, city_full, cfg)
        except Exception as exc:
            log.error("  Failed %s: %s", name, exc)
            record = {
                "name": name,
                "images": {"thumbnail": "", "gallery": [], "menu_photos": [], "menu_photos_source": "tripadvisor"},
                "location": {"city": city_full.split(",")[0], "country": "Australia"},
                "scrape_status": "error",
                "errors": [str(exc)],
                "source": "tripadvisor",
            }
        restaurants.append(record)
        menu_n = len((record.get("images") or {}).get("menu_photos") or [])
        log.info("  status=%s menu_photos=%s (tripadvisor only)", record.get("scrape_status"), menu_n)

        doc = _build_document(
            restaurants,
            {
                "city": city_full,
                "total_requested": len(rows),
                "total_scraped": len(restaurants),
            },
        )
        output_path.write_text(json.dumps(doc, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        time.sleep(cfg.SCRAPE_DELAY)

    final = _build_document(
        restaurants,
        {
            "city": city_full,
            "total_requested": len(rows),
            "total_scraped": len(restaurants),
            "file_size_bytes": 0,
            "file_size_mb": 0.0,
        },
    )
    size = _json_size_bytes(final)
    final["meta"]["file_size_bytes"] = size
    final["meta"]["file_size_mb"] = round(size / 1024 / 1024, 2)
    output_path.write_text(json.dumps(final, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

    ok = sum(1 for r in restaurants if r.get("scrape_status") in ("success", "partial"))
    log.info(
        "\n%s\n  COMPLETE %s — %s/%s ok → %s\n%s\n",
        "═" * 60,
        city_full,
        ok,
        len(restaurants),
        output_path,
        "═" * 60,
    )
    return restaurants, output_path


def _norm_name(name: str) -> str:
    return re.sub(r"[^a-z0-9]+", "", (name or "").lower())


def merge_menu_photos_into_google_scrape(city: str, ta_records: list[dict]) -> Path | None:
    """
    Attach TripAdvisor menu_photos onto matching Google/Places restaurant records.
    Replaces any prior menu_photos so the field stays TripAdvisor-only.
    """
    google_path = default_google_scrape_path(city)
    if not google_path.exists():
        log.warning("No Google scrape file to merge into: %s", google_path)
        return None

    doc = json.loads(google_path.read_text(encoding="utf-8"))
    restaurants = doc.get("restaurants") or []
    ta_by_name = {_norm_name(r.get("name", "")): r for r in ta_records if r.get("name")}

    merged = 0
    for rec in restaurants:
        key = _norm_name(rec.get("name", ""))
        ta = ta_by_name.get(key)
        if not ta:
            continue
        menu_photos = list((ta.get("images") or {}).get("menu_photos") or [])
        # Strict: only keep entries labelled tripadvisor
        menu_photos = [p for p in menu_photos if (p.get("source") or "tripadvisor") == "tripadvisor"]
        if not menu_photos:
            continue
        images = rec.setdefault("images", {})
        images["menu_photos"] = menu_photos
        images["menu_photos_source"] = "tripadvisor"
        rec["tripadvisor"] = ta.get("tripadvisor") or {}
        merged += 1

    meta = doc.setdefault("meta", {})
    meta["menu_photos_source"] = "tripadvisor_only"
    meta["tripadvisor_menu_photos_merged_at"] = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    meta["tripadvisor_menu_photos_merged_count"] = merged

    google_path.write_text(json.dumps(doc, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    log.info("Merged TripAdvisor menu_photos into %s restaurants → %s", merged, google_path)
    return google_path


def run_all_cities(
    cities: list[str],
    cfg: Config,
    *,
    limit: int,
    merge: bool,
) -> list[Path]:
    outputs: list[Path] = []
    for city in cities:
        records, path = run_tripadvisor_scrape_for_city(city, cfg, limit=limit)
        outputs.append(path)
        if merge:
            merge_menu_photos_into_google_scrape(city, records)
    return outputs


def run_scheduled(cities: list[str], cfg: Config, *, limit: int, merge: bool, interval_hours: int) -> None:
    """Long-running loop (alternative to system crontab)."""
    log.info("Schedule mode: every %sh for cities %s", interval_hours, cities)
    while True:
        started = datetime.now(timezone.utc)
        try:
            run_all_cities(cities, cfg, limit=limit, merge=merge)
        except Exception as exc:
            log.exception("Scheduled run failed: %s", exc)
        elapsed = (datetime.now(timezone.utc) - started).total_seconds()
        sleep_for = max(60.0, interval_hours * 3600 - elapsed)
        log.info("Sleeping %.0fs until next TripAdvisor run…", sleep_for)
        time.sleep(sleep_for)


# ─────────────────────────────────────────────────────────────
# CLI
# ─────────────────────────────────────────────────────────────

def main() -> int:
    parser = argparse.ArgumentParser(
        description="Scrape TripAdvisor restaurants (SerpAPI) — menu photos TripAdvisor-only",
    )
    parser.add_argument("--city", metavar="CITY", help="Single city (e.g. Sydney)")
    parser.add_argument("--cities", nargs="+", metavar="CITY", help="Multiple cities")
    parser.add_argument(
        "--all-cities",
        action="store_true",
        help="Scrape all default AU restaurant cities (same geo as Google/Places)",
    )
    parser.add_argument("--limit", type=int, default=100, help="Max restaurants per city (default 100)")
    parser.add_argument(
        "--merge",
        action="store_true",
        help="Merge TripAdvisor menu_photos into data/restaurants_data_<city>.json",
    )
    parser.add_argument(
        "--schedule",
        choices=("daily", "hourly"),
        default=None,
        help="Long-running loop: daily (24h) or hourly",
    )
    parser.add_argument("-v", "--verbose", action="store_true")
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s  [%(levelname)s]  %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    if args.all_cities:
        cities = list(AUSTRALIAN_RESTAURANT_CITIES)
    else:
        try:
            cities = resolve_cities(city=args.city, cities=args.cities)
        except ValueError as exc:
            log.error("%s", exc)
            return 1

    cfg = Config()
    if not cfg.SERPAPI_KEY:
        log.error("SERPAPI_KEY missing — set it in automation/outreach/.env")
        return 1

    try:
        if args.schedule:
            hours = 24 if args.schedule == "daily" else 1
            run_scheduled(cities, cfg, limit=args.limit, merge=args.merge, interval_hours=hours)
            return 0

        paths = run_all_cities(cities, cfg, limit=args.limit, merge=args.merge)
    except ValueError as exc:
        log.error("%s", exc)
        return 1

    print("\nTripAdvisor outputs:")
    for p in paths:
        print(f"  • {p}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
