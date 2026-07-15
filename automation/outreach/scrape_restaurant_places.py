#!/usr/bin/env python3
"""
Restaurant Data Scrape Pipeline — Google Places API (Google Maps)
==================================================================
Scrapes restaurant data directly from Google Maps via Places API.
No SerpAPI required.

Usage:
  python scrape_restaurant_places.py --city Sydney --total 100
  python scrape_restaurant_places.py --city Perth --total 100
  python scrape_restaurant_places.py --cities Sydney Melbourne Perth --total 100
"""

from __future__ import annotations

import argparse
import json
import logging
import sys
import time
from datetime import datetime, timezone
from pathlib import Path

from env_loader import load_project_env

load_project_env()

from google_places_scraper import (  # noqa: E402
    get_places_api_key,
    scrape_single_restaurant_places,
)
from tuvi_outreach_agent import (  # noqa: E402
    Config,
    default_leads_output_path,
    default_scrape_output_path,
    normalize_city_name,
    resolve_cities,
)

log = logging.getLogger("scrape_places")


def _json_size_bytes(obj: dict) -> int:
    return len(json.dumps(obj, ensure_ascii=False).encode("utf-8"))


def _build_document(restaurants: list[dict], meta_extra: dict) -> dict:
    return {
        "meta": {
            "version": "1.0",
            "scraped_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "source": "google_places_api_new",
            "data_fields": [
                "name", "cuisines", "reviews", "rating", "contact",
                "owners", "images", "hours", "location",
            ],
            **meta_extra,
        },
        "restaurants": restaurants,
    }


def run_places_scrape_pipeline(
    cfg: Config | None = None,
    input_path: str | Path | None = None,
    output_path: str | Path | None = None,
    city: str | None = None,
    max_size_mb: float = 50.0,
    max_reviews: int = 5,
    limit: int | None = None,
    niche: str = "restaurant",
    budget=None,
    known=None,
    leads_data: list[dict] | None = None,
) -> tuple[list[dict], Path]:
    """Scrape leads using Google Places API."""
    from known_leads import KnownLeadsRegistry
    from niche_config import get_niche
    from request_budget import BudgetExhausted

    cfg = cfg or Config()
    niche_cfg = get_niche(niche)
    get_places_api_key(cfg)  # validate key early
    cfg.SCRAPE_TIMEOUT = max(cfg.SCRAPE_TIMEOUT, 20)

    if city and input_path is None:
        input_path = default_leads_output_path(resolve_cities(city=city), cfg, niche=niche)
    elif input_path is None:
        input_path = Path(cfg.LEADS_DIR) / cfg.LEADS_FILE
    else:
        input_path = Path(input_path)

    if output_path is None:
        output_path = default_scrape_output_path(city, cfg, niche=niche)
    else:
        output_path = Path(output_path)

    output_path.parent.mkdir(parents=True, exist_ok=True)

    if leads_data is None:
        raw = json.loads(input_path.read_text(encoding="utf-8"))
        leads_data = raw.get("leads", [])
    if not leads_data:
        raise ValueError(f"No leads in {input_path}")

    if city:
        target_slug = city.lower().split(",")[0].strip()
        leads_data = [
            ld for ld in leads_data
            if (ld.get("location") or {}).get("city", "").lower() == target_slug
        ]
        if not leads_data:
            raise ValueError(f"No leads for {city} in {input_path}")

    if limit:
        leads_data = leads_data[:limit]

    meta_extra = {
        "input_file": str(input_path),
        "city_filter": normalize_city_name(city).split(",")[0] if city else None,
        "lead_type": niche_cfg.lead_type,
        "max_size_mb": max_size_mb,
        "max_reviews_per_restaurant": max_reviews,
        "total_requested": len(leads_data),
        "total_scraped": 0,
        "skipped_duplicate": 0,
        "stopped_reason": None,
        "file_size_bytes": 0,
        "file_size_mb": 0.0,
    }

    registry = known if known is not None else KnownLeadsRegistry()

    max_bytes = int(max_size_mb * 1024 * 1024)
    buffer_bytes = int(0.5 * 1024 * 1024)
    restaurants: list[dict] = []

    log.info(
        f"\n{'═'*60}\n"
        f"  GOOGLE PLACES SCRAPE — {len(leads_data)} restaurant(s)\n"
        f"{'═'*60}"
    )

    for idx, lead_dict in enumerate(leads_data, start=1):
        name = (lead_dict.get("company") or {}).get("name", "Unknown")
        skip, reason = registry.should_skip_scrape(lead_dict)
        if skip:
            meta_extra["skipped_duplicate"] += 1
            log.info(f"[{idx}/{len(leads_data)}] SKIP {name} ({reason})")
            continue

        existing_place_id = str((lead_dict.get("google") or {}).get("place_id") or "").strip()
        requests_needed = 1 if existing_place_id else 2
        if budget is not None and not budget.can_consume(requests_needed):
            meta_extra["stopped_reason"] = "request_budget_exhausted"
            log.info("Places scrape stopped — request budget exhausted")
            break

        log.info(f"[{idx}/{len(leads_data)}] {name}")

        try:
            record = scrape_single_restaurant_places(
                lead_dict,
                cfg,
                max_reviews,
                query_suffix=niche_cfg.places_query_suffix,
                budget=budget,
            )
        except BudgetExhausted:
            meta_extra["stopped_reason"] = "request_budget_exhausted"
            break

        restaurants.append(record)
        registry.register_scrape_record(record)

        doc = _build_document(restaurants, {**meta_extra, "total_scraped": len(restaurants)})
        size = _json_size_bytes(doc)

        if size >= max_bytes - buffer_bytes:
            meta_extra["stopped_reason"] = f"size_limit_{max_size_mb}mb"
            doc["meta"]["stopped_reason"] = meta_extra["stopped_reason"]
            output_path.write_text(json.dumps(doc, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
            break

        doc["meta"]["file_size_bytes"] = size
        doc["meta"]["file_size_mb"] = round(size / 1024 / 1024, 2)
        output_path.write_text(json.dumps(doc, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

        log.info(
            f"  ✅ {record.get('scrape_status')} — "
            f"rating {record.get('rating')} | {len(record.get('reviews') or [])} reviews | "
            f"{len(record.get('images', {}).get('gallery') or [])} photos"
        )
        time.sleep(cfg.SCRAPE_DELAY)

    else:
        meta_extra["stopped_reason"] = "completed_all_leads"

    final_doc = _build_document(restaurants, {**meta_extra, "total_scraped": len(restaurants)})
    final_size = _json_size_bytes(final_doc)
    final_doc["meta"]["file_size_bytes"] = final_size
    final_doc["meta"]["file_size_mb"] = round(final_size / 1024 / 1024, 2)
    final_doc["meta"]["stopped_reason"] = meta_extra["stopped_reason"]
    output_path.write_text(json.dumps(final_doc, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

    ok = sum(1 for r in restaurants if r.get("scrape_status") == "success")
    log.info(f"\nCOMPLETE — {ok}/{len(restaurants)} succeeded → {output_path}\n")
    return restaurants, output_path


def main() -> int:
    parser = argparse.ArgumentParser(description="Scrape restaurants via Google Places API")
    parser.add_argument("--city", metavar="CITY", help="e.g. Sydney, Perth, Melbourne")
    parser.add_argument("--cities", nargs="+", metavar="CITY", help="Multiple cities")
    parser.add_argument("--input", default=None, help="Leads JSON (auto: leads/lead_<city>.json)")
    parser.add_argument("--output", default=None, help="Output JSON path")
    parser.add_argument("--total", type=int, default=None, help="Scrape N leads (default: 100 with --city)")
    parser.add_argument("--max-reviews", type=int, default=5, help="Max reviews (Places API max ~5)")
    parser.add_argument("--max-size-mb", type=float, default=50.0)
    parser.add_argument("-v", "--verbose", action="store_true")
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s  [%(levelname)s]  %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    if not args.city and not args.cities:
        print("ERROR: Pass --city Sydney or --cities Sydney Melbourne Perth", file=sys.stderr)
        return 1

    city_list = [args.city] if args.city else args.cities
    total = args.total if args.total is not None else 100

    for city in city_list:
        try:
            _, out = run_places_scrape_pipeline(
                input_path=args.input,
                output_path=args.output,
                city=city,
                max_size_mb=args.max_size_mb,
                max_reviews=args.max_reviews,
                limit=total,
            )
            print(f"Saved: {out}")
        except ValueError as e:
            log.error(str(e))
            return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
