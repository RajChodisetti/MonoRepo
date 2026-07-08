#!/usr/bin/env python3
"""
Daily lead ingestion — fetch (Apollo) → scrape (Places) → import (Postgres).

Respects combined API request budget, niche type, and dedup against DB/files.
"""

from __future__ import annotations

import argparse
import logging
import os
import sys

from dotenv import load_dotenv

load_dotenv()

from env_loader import load_project_env

load_project_env()

from import_to_db import run_import  # noqa: E402
from ingestion_merge import merge_leads_file, merge_scrape_file  # noqa: E402
from ingestion_state import (  # noqa: E402
    get_apollo_page,
    load_state,
    record_run_summary,
    save_state,
    set_apollo_page,
)
from known_leads import KnownLeadsRegistry  # noqa: E402
from lead_fetch import fetch_leads_for_city  # noqa: E402
from niche_config import get_niche, list_niche_types  # noqa: E402
from request_budget import RequestBudget  # noqa: E402
from scrape_restaurant_places import run_places_scrape_pipeline  # noqa: E402
from tuvi_outreach_agent import (  # noqa: E402
    Config,
    AUSTRALIAN_RESTAURANT_CITIES,
    default_leads_output_path,
    default_scrape_output_path,
    lead_to_dict,
    normalize_city_name,
    resolve_cities,
)

log = logging.getLogger("daily_ingestion")


def _resolve_cities_from_env(args) -> list[str]:
    if args.city:
        return resolve_cities(city=args.city)
    if args.cities:
        return resolve_cities(cities=args.cities)
    raw = os.getenv("INGESTION_CITIES", "").strip()
    if raw:
        return resolve_cities(cities=[c.strip() for c in raw.split(",") if c.strip()])
    return list(AUSTRALIAN_RESTAURANT_CITIES)


def run_daily_ingestion(
    *,
    niche_type: str = "restaurant",
    cities: list[str] | None = None,
    max_requests: int = 500,
    per_page: int = 100,
    import_to_db: bool = True,
    cfg: Config | None = None,
) -> dict:
    cfg = cfg or Config()
    if not cfg.APOLLO_API_KEY:
        raise ValueError("APOLLO_API_KEY is missing from .env")

    niche = get_niche(niche_type)
    cities = cities or list(AUSTRALIAN_RESTAURANT_CITIES)
    budget = RequestBudget(max_requests)
    known = KnownLeadsRegistry.load_combined(niche.slug)
    state = load_state()

    summary = {
        "niche": niche.slug,
        "cities": [c.split(",")[0] for c in cities],
        "max_requests": max_requests,
        "requests_used": 0,
        "leads_fetched": 0,
        "leads_merged": 0,
        "scraped": 0,
        "scrape_merged": 0,
        "skipped_duplicate": 0,
        "imported": 0,
        "import_skipped": 0,
    }

    for city in cities:
        if budget.exhausted:
            break

        city_label = city.split(",")[0].strip()
        log.info("═══ City: %s | niche: %s ═══", city_label, niche.slug)

        leads_path = default_leads_output_path([city], cfg, niche=niche.slug)
        scrape_path = default_scrape_output_path(city, cfg, niche=niche.slug)

        page = get_apollo_page(state, niche.slug, city)
        pending: list[dict] = []
        city_new_leads = []
        city_scraped: list[dict] = []

        while not budget.exhausted:
            if not pending:
                if not budget.can_consume(1):
                    break
                new_leads, next_page, fetch_stats = fetch_leads_for_city(
                    cfg.APOLLO_API_KEY,
                    city,
                    niche,
                    per_page=per_page,
                    max_pages=1,
                    target_leads=per_page,
                    start_page=page,
                    budget=budget,
                    known=known,
                    cfg=cfg,
                )
                set_apollo_page(state, niche.slug, city, next_page)
                page = next_page

                if not new_leads:
                    log.info("  No new leads on Apollo page — moving on")
                    break

                for lead in new_leads:
                    lead_dict = lead_to_dict(lead)
                    known.register_lead_dict(lead_dict)
                    pending.append(lead_dict)
                    city_new_leads.append(lead)

                summary["leads_fetched"] += len(new_leads)
                continue

            if not budget.can_consume(2):
                log.info("  Budget too low for Places scrape (need 2 requests)")
                break

            lead_dict = pending.pop(0)
            skip, reason = known.should_skip_scrape(lead_dict)
            if skip:
                summary["skipped_duplicate"] += 1
                log.info("  Skip scrape: %s (%s)", (lead_dict.get("company") or {}).get("name"), reason)
                continue

            records, _ = run_places_scrape_pipeline(
                cfg=cfg,
                city=city,
                niche=niche.slug,
                budget=budget,
                known=known,
                leads_data=[lead_dict],
                output_path=scrape_path,
            )
            city_scraped.extend(records)
            summary["scraped"] += len(records)

        if city_new_leads:
            merged = merge_leads_file(
                leads_path,
                city_new_leads,
                cities=[city],
                niche_type=niche.slug,
                per_page=per_page,
                max_pages=1,
                target_per_city=None,
            )
            summary["leads_merged"] += merged

        if city_scraped:
            added = merge_scrape_file(scrape_path, city_scraped)
            summary["scrape_merged"] += added

        record_run_summary(state, niche.slug, city, {
            "requests_used": budget.total_used,
            "leads_fetched": len(city_new_leads),
            "scraped": len(city_scraped),
        })

    summary["requests_used"] = budget.total_used
    save_state(state)

    if import_to_db:
        data_file = default_scrape_output_path(cities[0] if len(cities) == 1 else None, cfg, niche=niche.slug)
        data_files = []
        for city in cities:
            path = default_scrape_output_path(city, cfg, niche=niche.slug)
            if path.is_file():
                data_files.append(path)
        if not data_files and data_file.is_file():
            data_files = [data_file]
        imported, skipped = run_import(restaurants_only=True, data_files=data_files, do_import_leads=False)
        summary["imported"] = imported
        summary["import_skipped"] = skipped

    return summary


def main() -> int:
    parser = argparse.ArgumentParser(description="Daily lead ingestion with budget and dedup")
    city_group = parser.add_mutually_exclusive_group()
    city_group.add_argument("--city", metavar="CITY")
    city_group.add_argument("--cities", nargs="+", metavar="CITY")
    parser.add_argument(
        "--type",
        default=os.getenv("INGESTION_TYPE", "restaurant"),
        choices=list_niche_types(),
        help="Business niche (default: restaurant)",
    )
    parser.add_argument(
        "--max-requests",
        type=int,
        default=int(os.getenv("LEAD_INGESTION_MAX_REQUESTS", "500")),
        help="Combined Apollo + Places request cap (default: 500)",
    )
    parser.add_argument("--per-page", type=int, default=100)
    parser.add_argument("--no-import", action="store_true", help="Skip import_to_db step")
    parser.add_argument("-v", "--verbose", action="store_true")
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s  [%(levelname)s]  %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    try:
        cities = _resolve_cities_from_env(args)
        summary = run_daily_ingestion(
            niche_type=args.type,
            cities=cities,
            max_requests=args.max_requests,
            per_page=args.per_page,
            import_to_db=not args.no_import,
        )
    except ValueError as exc:
        log.error("%s", exc)
        return 1

    log.info("Daily ingestion complete: %s", summary)
    print("\nSummary:")
    for key, value in summary.items():
        print(f"  {key}: {value}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
