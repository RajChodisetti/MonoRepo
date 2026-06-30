#!/usr/bin/env python3
"""
Restaurant Data Scraping Pipeline (SerpAPI)
============================================
Reads leads/lead.json and enriches each restaurant with:
  • name, cuisines, rating, reviews count
  • menu items (with images)
  • reviews (reviewer, text, stars)
  • phone, email, website
  • owner names (from Apollo lead + Google data)

Output: data/restaurants_data.json  (up to 50 MB by default)

Usage:
  python scrape_restaurant_data.py
  python scrape_restaurant_data.py --input leads/lead.json --max-size-mb 50
  python scrape_restaurant_data.py --limit 5          # test first 5 only
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

import requests
from dotenv import load_dotenv

load_dotenv()

from tuvi_outreach_agent import (  # noqa: E402
    Config,
    _lead_from_dict,
    city_slug,
    default_leads_output_path,
    default_scrape_output_path,
    normalize_city_name,
    resolve_cities,
    with_retry,
)


def filter_leads_by_city(leads_data: list[dict], city: str) -> list[dict]:
    """Keep only leads matching the given city (case-insensitive)."""
    target = city_slug(normalize_city_name(city))
    filtered = [
        lead for lead in leads_data
        if city_slug((lead.get("location") or {}).get("city", "")) == target
        or city_slug((lead.get("location") or {}).get("city", "") + ", Australia") == target
    ]
    return filtered

log = logging.getLogger("restaurant_scraper")

SERPAPI_URL = "https://serpapi.com/search.json"
EMAIL_RE = re.compile(r"[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}")
OWNER_TITLE_KEYWORDS = ("owner", "founder", "co-founder", "proprietor", "managing director")


# ─────────────────────────────────────────────────────────────
# SerpAPI helpers
# ─────────────────────────────────────────────────────────────

def _serpapi_get(params: dict, cfg: Config, label: str = "serpapi") -> dict:
    params = {**params, "api_key": cfg.SERPAPI_KEY}

    def _call():
        resp = requests.get(SERPAPI_URL, params=params, timeout=cfg.SCRAPE_TIMEOUT)
        if resp.status_code != 200:
            raise RuntimeError(f"HTTP {resp.status_code}: {resp.text[:300]}")
        data = resp.json()
        if data.get("error"):
            raise RuntimeError(data["error"])
        return data

    return with_retry(_call, cfg.RETRY_ATTEMPTS, cfg.RETRY_BACKOFF, label=label)


def _find_place(company_name: str, city: str, cfg: Config) -> dict:
    """Search Google Maps and return full place_results."""
    query = f"{company_name} restaurant {city} Australia"
    data = _serpapi_get(
        {"engine": "google_maps", "q": query, "hl": "en", "gl": "au"},
        cfg,
        label=f"maps-search:{company_name}",
    )

    place = data.get("place_results") or {}
    if not place and data.get("local_results"):
        place = data["local_results"][0]

    place_id = place.get("place_id")
    data_id = place.get("data_id")
    if not place_id and not data_id:
        return place

    # Always fetch full place details for menu, reviews, images
    params: dict[str, Any] = {"engine": "google_maps", "type": "place", "hl": "en", "gl": "au"}
    if place_id:
        params["place_id"] = place_id
    else:
        params["data_id"] = data_id

    detail = _serpapi_get(params, cfg, label=f"maps-place:{company_name}")
    return detail.get("place_results") or place


def _fetch_reviews(
    place: dict,
    cfg: Config,
    max_reviews: int,
) -> list[dict]:
    """Paginate google_maps_reviews until max_reviews reached."""
    place_id = (place.get("place_id") or "").strip() or None
    data_id = (place.get("data_id") or "").strip() or None
    if not place_id and not data_id:
        return []

    collected: list[dict] = []
    next_token: str | None = None
    pages = 0
    max_pages = max(1, (max_reviews // 8) + 2)

    while len(collected) < max_reviews and pages < max_pages:
        params: dict[str, Any] = {
            "engine": "google_maps_reviews",
            "hl": "en",
            "sort_by": "qualityScore",
        }
        # SerpAPI requires data_id/place_id on every page, including paginated requests
        if data_id:
            params["data_id"] = data_id
        elif place_id:
            params["place_id"] = place_id
        else:
            break

        if next_token:
            params["next_page_token"] = next_token
            params["num"] = min(20, max_reviews - len(collected))

        try:
            data = _serpapi_get(params, cfg, label=f"reviews:p{pages}")
        except Exception as exc:
            log.warning(f"    Reviews page {pages} failed: {exc}")
            break

        for rev in data.get("reviews") or []:
            user = rev.get("user") or {}
            images = []
            for img in rev.get("images") or []:
                if isinstance(img, str):
                    images.append(img)
                else:
                    url = img.get("image") or img.get("thumbnail")
                    if url:
                        images.append(url)

            collected.append({
                "reviewer": user.get("name", ""),
                "review": rev.get("snippet") or (rev.get("extracted_snippet") or {}).get("original", ""),
                "stars": rev.get("rating"),
                "date": rev.get("date") or rev.get("iso_date", ""),
                "images": images,
                "source": rev.get("source", "google_maps"),
            })
            if len(collected) >= max_reviews:
                break

        next_token = (data.get("serpapi_pagination") or {}).get("next_page_token")
        pages += 1
        if not next_token:
            break
        time.sleep(cfg.SCRAPE_DELAY)

    return collected


def _parse_place_reviews(place: dict, max_reviews: int) -> list[dict]:
    """Fallback reviews from place_results.user_reviews.most_relevant."""
    reviews = []
    for rev in (place.get("user_reviews") or {}).get("most_relevant") or []:
        images = []
        for img in rev.get("images") or []:
            if isinstance(img, str):
                images.append(img)
            else:
                url = img.get("image") or img.get("thumbnail")
                if url:
                    images.append(url)
        reviews.append({
            "reviewer": rev.get("username", ""),
            "review": rev.get("description", ""),
            "stars": rev.get("rating"),
            "date": rev.get("date") or rev.get("date_iso8601", ""),
            "images": images,
            "source": "google_maps_place",
        })
        if len(reviews) >= max_reviews:
            break
    return reviews


def _extract_emails_from_text(text: str) -> list[str]:
    found = []
    for raw in EMAIL_RE.findall(text):
        email = raw.rstrip(".")
        if any(x in email.lower() for x in ("example.com", "sentry", "wixpress", "schema.org")):
            continue
        found.append(email)
    return list(dict.fromkeys(found))


def _find_restaurant_email(
    company_name: str,
    city: str,
    website: str,
    cfg: Config,
) -> str:
    """Search Google for restaurant contact email."""
    domain = ""
    if website:
        domain = re.sub(r"^https?://(www\.)?", "", website).split("/")[0]

    queries = [
        f'"{company_name}" {city} email',
        f'"{company_name}" contact email Australia',
    ]
    if domain:
        queries.insert(0, f"site:{domain} email OR contact")

    all_emails: list[str] = []
    for q in queries:
        try:
            data = _serpapi_get(
                {"engine": "google", "q": q, "hl": "en", "gl": "au", "num": 10},
                cfg,
                label="email-search",
            )
            blob = json.dumps(data)
            all_emails.extend(_extract_emails_from_text(blob))
            for res in data.get("organic_results") or []:
                all_emails.extend(_extract_emails_from_text(
                    (res.get("title") or "") + " " + (res.get("snippet") or "")
                ))
        except Exception as exc:
            log.warning(f"    Email search failed: {exc}")
        if all_emails:
            break
        time.sleep(cfg.SCRAPE_DELAY)

    if domain:
        for email in all_emails:
            if domain.split(".")[0] in email.lower():
                return email

    return all_emails[0] if all_emails else ""


def _build_menu_items(place: dict) -> list[dict]:
    """Parse menu categories + highlights into items with images."""
    menu = place.get("menu") or {}
    highlight_map: dict[str, dict] = {}
    for h in menu.get("highlights") or []:
        title = (h.get("title") or "").strip().lower()
        if title:
            highlight_map[title] = h

    menu_images = []
    for img in menu.get("images") or []:
        url = img.get("image") or img.get("thumbnail")
        if url:
            menu_images.append({
                "url": url,
                "thumbnail": img.get("thumbnail", url),
                "date": img.get("date", ""),
            })

    items: list[dict] = []
    for category in menu.get("categories") or []:
        cat_name = category.get("title", "")
        for raw in category.get("items") or []:
            name = raw.get("title", "")
            images: list[dict] = []

            hl = highlight_map.get(name.strip().lower())
            if hl:
                images.append({
                    "url": hl.get("image") or hl.get("thumbnail", ""),
                    "thumbnail": hl.get("thumbnail", ""),
                    "popular": hl.get("popular", False),
                })

            items.append({
                "name": name,
                "category": cat_name,
                "description": raw.get("description", ""),
                "price": raw.get("price", ""),
                "price_numeric": raw.get("extracted_price"),
                "images": images,
            })

    # Highlights not already in categories
    existing = {i["name"].strip().lower() for i in items}
    for h in menu.get("highlights") or []:
        title = h.get("title", "")
        if title.strip().lower() in existing:
            continue
        items.append({
            "name": title,
            "category": "Popular Highlights",
            "description": "",
            "price": "",
            "price_numeric": None,
            "images": [{
                "url": h.get("image") or h.get("thumbnail", ""),
                "thumbnail": h.get("thumbnail", ""),
                "popular": h.get("popular", False),
            }],
        })

    # Attach shared menu photos to items missing images (round-robin)
    if menu_images:
        idx = 0
        for item in items:
            if not item["images"] and menu_images:
                img = menu_images[idx % len(menu_images)]
                item["images"].append(img)
                idx += 1

    return items


def _extract_cuisines(place: dict) -> list[str]:
    cuisines: list[str] = []
    for key in ("types", "type"):
        val = place.get(key)
        if isinstance(val, list):
            cuisines.extend(val)
        elif isinstance(val, str) and val:
            cuisines.append(val)

    for ext in place.get("extensions") or []:
        if "cuisine" in str(ext).lower() or "offerings" in str(ext).lower():
            for k, v in ext.items():
                if isinstance(v, list):
                    cuisines.extend(v)

    # Deduplicate, filter generic
    skip = {"restaurant", "food", "point of interest", "establishment"}
    seen = set()
    result = []
    for c in cuisines:
        cl = c.strip()
        if cl.lower() in skip or cl.lower() in seen:
            continue
        seen.add(cl.lower())
        result.append(cl)
    return result


def _extract_owners(lead_contact: dict, place: dict) -> list[str]:
    owners: list[str] = []

    name = (lead_contact.get("name") or "").strip()
    title = (lead_contact.get("job_title") or "").lower()
    if name and any(kw in title for kw in OWNER_TITLE_KEYWORDS):
        owners.append(name)

    # Sometimes mentioned in description
    desc = place.get("description") or ""
    for match in re.findall(r"(?:owned by|founder|chef)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)?)", desc, re.I):
        owners.append(match.strip())

    return list(dict.fromkeys(owners))


def scrape_single_restaurant(lead_dict: dict, cfg: Config, max_reviews: int) -> dict:
    """Scrape full restaurant profile for one lead entry."""
    lead = _lead_from_dict(lead_dict)
    company = lead_dict.get("company") or {}
    contact = lead_dict.get("contact") or {}
    location = lead_dict.get("location") or {}
    city = location.get("city") or lead.extra.get("city", "")

    result: dict[str, Any] = {
        "name": company.get("name") or lead.company_name,
        "cuisines": [],
        "rating": None,
        "reviews_count": None,
        "contact": {
            "phone": "",
            "email": contact.get("email") or "",
            "website": company.get("domain") or "",
        },
        "owners": _extract_owners(contact, {}),
        "location": {
            "address": "",
            "city": city,
            "state": location.get("state", ""),
            "country": location.get("country", "Australia"),
            "coordinates": {},
        },
        "menu_items": [],
        "reviews": [],
        "images": {
            "thumbnail": "",
            "gallery": [],
            "menu_photos": [],
        },
        "hours": {},
        "price_level": "",
        "apollo_lead": lead_dict,
        "google": {
            "place_id": "",
            "data_id": "",
        },
        "scrape_status": "pending",
        "errors": [],
    }

    try:
        place = _find_place(result["name"], city, cfg)
        if not place:
            result["scrape_status"] = "not_found"
            result["errors"].append("No Google Maps listing found")
            return result

        time.sleep(cfg.SCRAPE_DELAY)

        result["name"] = place.get("title") or result["name"]
        result["cuisines"] = _extract_cuisines(place)
        result["rating"] = place.get("rating")
        result["reviews_count"] = place.get("reviews")
        result["contact"]["phone"] = place.get("phone") or ""
        result["contact"]["website"] = place.get("website") or result["contact"]["website"]
        result["price_level"] = place.get("price") or ""
        result["hours"] = place.get("operating_hours") or {}
        if not result["hours"] and isinstance(place.get("hours"), dict):
            result["hours"] = place["hours"]
        elif not result["hours"] and isinstance(place.get("hours"), list):
            result["hours"] = {k: v for block in place["hours"] for k, v in block.items()}
        elif isinstance(place.get("hours"), str):
            result["hours"] = {"summary": place["hours"]}
        result["location"]["address"] = place.get("address") or ""
        coords = place.get("gps_coordinates") or {}
        result["location"]["coordinates"] = {
            "latitude": coords.get("latitude"),
            "longitude": coords.get("longitude"),
        }
        result["google"]["place_id"] = place.get("place_id") or ""
        result["google"]["data_id"] = place.get("data_id") or ""
        result["images"]["thumbnail"] = place.get("thumbnail") or ""

        for img in place.get("images") or []:
            if isinstance(img, str):
                result["images"]["gallery"].append({"title": "", "url": img})
            else:
                url = img.get("thumbnail") or img.get("image")
                if url:
                    result["images"]["gallery"].append({
                        "title": img.get("title", ""),
                        "url": url,
                    })

        menu = place.get("menu") or {}
        for img in menu.get("images") or []:
            url = img.get("image") or img.get("thumbnail")
            if url:
                result["images"]["menu_photos"].append({
                    "url": url,
                    "thumbnail": img.get("thumbnail", url),
                    "date": img.get("date", ""),
                })

        result["menu_items"] = _build_menu_items(place)
        result["owners"] = _extract_owners(contact, place)

        # Reviews — API first, merge with place embed, dedupe by reviewer+snippet
        api_reviews = _fetch_reviews(place, cfg, max_reviews)
        embedded = _parse_place_reviews(place, max_reviews)
        seen_rev: set[str] = set()
        reviews = []
        for rev in api_reviews + embedded:
            key = f"{rev.get('reviewer','')}|{(rev.get('review') or '')[:60]}"
            if key in seen_rev:
                continue
            seen_rev.add(key)
            reviews.append(rev)
            if len(reviews) >= max_reviews:
                break
        result["reviews"] = reviews

        # Email — Apollo contact first, then Google search
        if not result["contact"]["email"]:
            result["contact"]["email"] = _find_restaurant_email(
                result["name"], city, result["contact"]["website"], cfg
            )

        result["scrape_status"] = "success"

    except Exception as exc:
        result["scrape_status"] = "error"
        result["errors"].append(str(exc))
        log.error(f"  Scrape failed for {result['name']}: {exc}")

    return result


# ─────────────────────────────────────────────────────────────
# Pipeline orchestration
# ─────────────────────────────────────────────────────────────

def _json_size_bytes(obj: dict) -> int:
    return len(json.dumps(obj, ensure_ascii=False).encode("utf-8"))


def _build_document(
    restaurants: list[dict],
    meta_extra: dict,
) -> dict:
    return {
        "meta": {
            "version": "1.0",
            "scraped_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "source": "serpapi",
            "data_fields": [
                "name", "cuisines", "menu_items", "reviews",
                "rating", "contact", "owners", "images",
            ],
            **meta_extra,
        },
        "restaurants": restaurants,
    }


def run_restaurant_scrape_pipeline(
    cfg: Config | None = None,
    input_path: str | Path | None = None,
    output_path: str | Path | None = None,
    city: str | None = None,
    max_size_mb: float = 50.0,
    max_reviews: int = 40,
    limit: int | None = None,
) -> tuple[list[dict], Path]:
    """
    Scrape restaurant data for all leads in lead.json.
    Stops adding restaurants when output approaches max_size_mb.
    Saves incrementally after each restaurant.
    """
    cfg = cfg or Config()
    if not cfg.SERPAPI_KEY:
        raise ValueError("SERPAPI_KEY is missing from .env")

    cfg.SCRAPE_TIMEOUT = max(cfg.SCRAPE_TIMEOUT, 30)

    if city and input_path is None:
        input_path = default_leads_output_path(resolve_cities(city=city), cfg)
    elif input_path is None:
        input_path = Path(cfg.LEADS_DIR) / cfg.LEADS_FILE
    else:
        input_path = Path(input_path)

    if output_path is None:
        output_path = default_scrape_output_path(city, cfg)
    else:
        output_path = Path(output_path)

    output_path.parent.mkdir(parents=True, exist_ok=True)

    leads_data = json.loads(input_path.read_text(encoding="utf-8")).get("leads", [])
    if not leads_data:
        raise ValueError(f"No leads found in {input_path}")

    if city:
        leads_data = filter_leads_by_city(leads_data, city)
        if not leads_data:
            raise ValueError(
                f"No leads for city '{city}' in {input_path}. "
                f"Run: python fetch_restaurant_leads.py --city {city}"
            )
        meta_extra_city = normalize_city_name(city).split(",")[0]
    else:
        meta_extra_city = None

    if limit:
        leads_data = leads_data[:limit]

    max_bytes = int(max_size_mb * 1024 * 1024)
    buffer_bytes = int(0.5 * 1024 * 1024)  # stop 0.5 MB before limit

    restaurants: list[dict] = []
    meta_extra = {
        "input_file": str(input_path),
        "city_filter": meta_extra_city,
        "max_size_mb": max_size_mb,
        "max_reviews_per_restaurant": max_reviews,
        "total_requested": len(leads_data),
        "total_scraped": 0,
        "stopped_reason": None,
        "file_size_bytes": 0,
        "file_size_mb": 0.0,
    }

    log.info(
        f"\n{'═'*60}\n"
        f"  RESTAURANT SCRAPE  —  {len(leads_data)} lead(s)  |  "
        f"max {max_size_mb} MB\n"
        f"{'═'*60}"
    )

    for idx, lead_dict in enumerate(leads_data, start=1):
        name = (lead_dict.get("company") or {}).get("name", "Unknown")
        log.info(f"[{idx}/{len(leads_data)}] Scraping: {name}")

        record = scrape_single_restaurant(lead_dict, cfg, max_reviews)
        restaurants.append(record)

        doc = _build_document(restaurants, {**meta_extra, "total_scraped": len(restaurants)})
        size = _json_size_bytes(doc)

        if size >= max_bytes - buffer_bytes:
            meta_extra["stopped_reason"] = f"size_limit_{max_size_mb}mb"
            log.warning(
                f"  Size limit reached ({size / 1024 / 1024:.1f} MB) — "
                f"stopping after {len(restaurants)} restaurants"
            )
            doc["meta"]["stopped_reason"] = meta_extra["stopped_reason"]
            output_path.write_text(
                json.dumps(doc, indent=2, ensure_ascii=False) + "\n",
                encoding="utf-8",
            )
            break

        # Incremental save
        doc["meta"]["file_size_bytes"] = size
        doc["meta"]["file_size_mb"] = round(size / 1024 / 1024, 2)
        output_path.write_text(
            json.dumps(doc, indent=2, ensure_ascii=False) + "\n",
            encoding="utf-8",
        )

        status = record.get("scrape_status", "?")
        menu_n = len(record.get("menu_items") or [])
        rev_n = len(record.get("reviews") or [])
        log.info(f"  ✅ {status} — {menu_n} menu items, {rev_n} reviews")

        time.sleep(cfg.SCRAPE_DELAY)

    else:
        meta_extra["stopped_reason"] = "completed_all_leads"

    final_doc = _build_document(restaurants, {
        **meta_extra,
        "total_scraped": len(restaurants),
    })
    final_size = _json_size_bytes(final_doc)
    final_doc["meta"]["file_size_bytes"] = final_size
    final_doc["meta"]["file_size_mb"] = round(final_size / 1024 / 1024, 2)
    final_doc["meta"]["stopped_reason"] = meta_extra["stopped_reason"]

    output_path.write_text(
        json.dumps(final_doc, indent=2, ensure_ascii=False) + "\n",
        encoding="utf-8",
    )

    ok = sum(1 for r in restaurants if r.get("scrape_status") == "success")
    log.info(
        f"\n{'═'*60}\n"
        f"  COMPLETE — {ok}/{len(restaurants)} succeeded  |  "
        f"{final_doc['meta']['file_size_mb']} MB  →  {output_path}\n"
        f"{'═'*60}\n"
    )

    return restaurants, output_path


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Scrape restaurant data from Google via SerpAPI",
    )
    parser.add_argument("--input", default=None,
                        help="Input leads JSON (auto: leads/lead_<city>.json when --city is set)")
    parser.add_argument("--city", metavar="CITY",
                        help="Scrape only this city (e.g. --city Sydney). Auto-picks lead file & output.")
    parser.add_argument("--output", default=None,
                        help="Output JSON path (auto: data/restaurants_data_<city>.json)")
    parser.add_argument("--max-size-mb", type=float, default=50.0,
                        help="Max output file size in MB (default: 50)")
    parser.add_argument("--max-reviews", type=int, default=40,
                        help="Max reviews per restaurant (default: 40)")
    parser.add_argument("--total", type=int, default=None, metavar="N",
                        help="Scrape first N leads (default: 100 when --city is set)")
    parser.add_argument("--limit", type=int, default=None,
                        help="Alias for --total (scrape first N leads only)")
    parser.add_argument("-v", "--verbose", action="store_true")
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s  [%(levelname)s]  %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    try:
        total = args.total if args.total is not None else args.limit
        if total is None and args.city:
            total = 100

        _, out = run_restaurant_scrape_pipeline(
            input_path=args.input,
            output_path=args.output,
            city=args.city,
            max_size_mb=args.max_size_mb,
            max_reviews=args.max_reviews,
            limit=total,
        )
    except ValueError as e:
        log.error(str(e))
        return 1

    print(f"\nSaved to: {out}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
