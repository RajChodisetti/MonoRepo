#!/usr/bin/env python3
"""Daily Places discovery followed by targeted Apollo contact enrichment."""

from __future__ import annotations

import argparse
import logging
import os
import sys
import time

from env_loader import load_project_env

load_project_env()

from apollo_enrichment import (  # noqa: E402
    enrich_missing_contact_with_apollo,
    get_apollo_api_key,
    needs_apollo_enrichment,
)
from google_places_scraper import (  # noqa: E402
    discover_places_for_city,
    get_places_api_key,
    place_to_lead_dict,
    scrape_single_restaurant_places,
)
from import_to_db import run_import  # noqa: E402
from ingestion_merge import merge_scrape_file  # noqa: E402
from ingestion_state import load_state, record_run_summary, save_state  # noqa: E402
from known_leads import KnownLeadsRegistry  # noqa: E402
from niche_config import get_niche, list_niche_types  # noqa: E402
from request_budget import RequestBudget  # noqa: E402
from tuvi_outreach_agent import (  # noqa: E402
    AUSTRALIAN_RESTAURANT_CITIES,
    Config,
    default_scrape_output_path,
    resolve_cities,
)

log = logging.getLogger("daily_ingestion")


def _refuse_when_durable_city_pipeline_is_installed() -> None:
    """Prevent the retired in-memory budget from bypassing scrape_jobs."""
    database_url = os.getenv("DATABASE_URL", "").strip()
    if not database_url:
        raise RuntimeError(
            "The legacy daily ingestion entrypoint requires DATABASE_URL for its safety check; "
            "use POST /api/v1/scrape-jobs and city_scrape_worker.py instead."
        )
    try:
        import psycopg2

        conn = psycopg2.connect(database_url, connect_timeout=10)
        try:
            with conn.cursor() as cur:
                cur.execute("SELECT to_regclass('public.scrape_jobs') IS NOT NULL")
                durable_installed = bool(cur.fetchone()[0])
        finally:
            conn.close()
    except Exception as exc:
        raise RuntimeError(
            "Could not prove that the durable city pipeline is absent; refusing legacy provider calls."
        ) from exc
    if durable_installed:
        raise RuntimeError(
            "daily_ingestion.py is retired after migration 000015; trigger the durable city-scrape API instead."
        )


def _resolve_cities_from_env(args) -> list[str]:
    if args.city:
        return resolve_cities(city=args.city)
    if args.cities:
        return resolve_cities(cities=args.cities)
    raw = os.getenv("INGESTION_CITIES", "").strip()
    if raw:
        return resolve_cities(cities=[city.strip() for city in raw.split(",") if city.strip()])
    return list(AUSTRALIAN_RESTAURANT_CITIES)


def run_daily_ingestion(
    *,
    niche_type: str = "restaurant",
    cities: list[str] | None = None,
    max_requests: int = 500,
    target_per_city: int = 100,
    import_to_db: bool = True,
    cfg: Config | None = None,
) -> dict:
    _refuse_when_durable_city_pipeline_is_installed()
    cfg = cfg or Config()
    get_places_api_key(cfg)
    apollo_enabled = bool(getattr(cfg, "APOLLO_ENRICHMENT_ENABLED", True))
    if apollo_enabled:
        get_apollo_api_key(cfg)
    if import_to_db and not os.getenv("DATABASE_URL", "").strip():
        raise ValueError(
            "DATABASE_URL is required when database import is enabled; "
            "set it in the process environment or INGESTION_ENV_FILE"
        )
    if target_per_city < 1:
        raise ValueError("target_per_city must be at least 1")

    niche = get_niche(niche_type)
    cities = cities or list(AUSTRALIAN_RESTAURANT_CITIES)
    budget = RequestBudget(max_requests)
    known = KnownLeadsRegistry.load_combined(
        niche.slug,
        require_database=import_to_db,
    )
    state = load_state()

    summary = {
        "source": "google_places_api_new+apollo_contact_enrichment",
        "niche": niche.slug,
        "cities": [city.split(",")[0] for city in cities],
        "max_requests": max_requests,
        "requests_used": 0,
        "places_discovered": 0,
        "scraped": 0,
        "scrape_success": 0,
        "scrape_merged": 0,
        "skipped_duplicate": 0,
        "apollo_candidates": 0,
        "apollo_enriched": 0,
        "apollo_owner_filled": 0,
        "apollo_email_filled": 0,
        "apollo_skipped": 0,
        "apollo_requests": 0,
        "imported": 0,
        "import_skipped": 0,
    }

    for city in cities:
        if budget.exhausted:
            break

        city_label = city.split(",")[0].strip()
        scrape_path = default_scrape_output_path(city, cfg, niche=niche.slug)
        log.info("═══ City: %s | niche: %s | source: Google Places ═══", city_label, niche.slug)

        discovered = discover_places_for_city(
            city=city,
            niche=niche.slug,
            limit=target_per_city,
            cfg=cfg,
            budget=budget,
            known=known,
        )
        summary["places_discovered"] += len(discovered)

        city_scraped = 0
        city_success = 0
        city_merged = 0
        city_skipped = 0
        city_apollo_candidates = 0
        city_apollo_enriched = 0
        city_apollo_requests = 0

        for place in discovered:
            if not budget.can_consume(1):
                log.info("  Places detail budget exhausted")
                break

            lead_dict = place_to_lead_dict(place, city, niche.slug)
            skip, reason = known.should_skip_scrape(lead_dict)
            if skip:
                city_skipped += 1
                summary["skipped_duplicate"] += 1
                log.info("  Skip %s (%s)", (lead_dict.get("company") or {}).get("name"), reason)
                continue

            record = scrape_single_restaurant_places(
                lead_dict,
                cfg,
                max_reviews=5,
                query_suffix=niche.places_query_suffix,
                budget=budget,
            )
            city_scraped += 1
            summary["scraped"] += 1
            if record.get("scrape_status") == "success":
                city_success += 1
                summary["scrape_success"] += 1

            if (
                apollo_enabled
                and record.get("scrape_status") == "success"
                and needs_apollo_enrichment(record)
            ):
                city_apollo_candidates += 1
                summary["apollo_candidates"] += 1
                before_apollo = budget.total_used
                record, apollo_stats = enrich_missing_contact_with_apollo(
                    record,
                    cfg,
                    niche,
                    budget=budget,
                )
                apollo_requests = budget.total_used - before_apollo
                city_apollo_requests += apollo_requests
                summary["apollo_requests"] += apollo_requests
                record["apollo_enrichment"] = apollo_stats
                if apollo_stats.get("status") == "enriched":
                    city_apollo_enriched += 1
                    summary["apollo_enriched"] += 1
                else:
                    summary["apollo_skipped"] += 1
                if apollo_stats.get("owner_added"):
                    summary["apollo_owner_filled"] += 1
                if apollo_stats.get("email_added"):
                    summary["apollo_email_filled"] += 1
                if apollo_stats.get("status") == "error":
                    raise RuntimeError(
                        "Apollo contact enrichment failed: "
                        f"{apollo_stats.get('error') or 'unknown provider error'}"
                    )
                if apollo_stats.get("status") == "budget_exhausted":
                    log.info(
                        "  Apollo budget exhausted before enriching %s",
                        record.get("name") or "business",
                    )
                    break

            known.register_scrape_record(record)
            added = merge_scrape_file(scrape_path, [record])
            city_merged += added
            summary["scrape_merged"] += added
            time.sleep(cfg.SCRAPE_DELAY)

        record_run_summary(
            state,
            niche.slug,
            city,
            {
                "source": "google_places_api_new+apollo_contact_enrichment",
                "requests_used": budget.total_used,
                "places_discovered": len(discovered),
                "scraped": city_scraped,
                "scrape_success": city_success,
                "scrape_merged": city_merged,
                "skipped_duplicate": city_skipped,
                "apollo_candidates": city_apollo_candidates,
                "apollo_enriched": city_apollo_enriched,
                "apollo_requests": city_apollo_requests,
            },
        )

    summary["requests_used"] = budget.total_used
    save_state(state)

    if import_to_db:
        data_files = [
            path
            for city in cities
            if (path := default_scrape_output_path(city, cfg, niche=niche.slug)).is_file()
        ]
        imported, skipped = run_import(
            restaurants_only=True,
            data_files=data_files,
            do_import_leads=False,
        )
        summary["imported"] = imported
        summary["import_skipped"] = skipped

    return summary


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Daily Google Places discovery plus Apollo contact enrichment "
            "and database import"
        ),
    )
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
        help="Combined Google Places + Apollo request cap (default: 500)",
    )
    parser.add_argument(
        "--target-per-city",
        "--per-page",
        dest="target_per_city",
        type=int,
        default=100,
        help="Maximum new businesses to enrich per city (default: 100)",
    )
    parser.add_argument("--no-import", action="store_true", help="Skip PostgreSQL import")
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
            target_per_city=args.target_per_city,
            import_to_db=not args.no_import,
        )
    except (ValueError, RuntimeError) as exc:
        log.error("%s", exc)
        return 1

    log.info("Daily ingestion complete: %s", summary)
    print("\nSummary:")
    for key, value in summary.items():
        print(f"  {key}: {value}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
