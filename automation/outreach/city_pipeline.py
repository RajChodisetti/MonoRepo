#!/usr/bin/env python3
"""
City Pipeline — Fetch leads (Apollo) + Scrape restaurants (SerpAPI)
===================================================================
One command per city, or batch all three.

Usage:
  # Full pipeline for one city (100 leads → scrape all 100)
  python city_pipeline.py --city Sydney --total 100

  # All three cities
  python city_pipeline.py --cities Sydney Melbourne Perth --total 100

  # Only scrape (leads already in leads/lead_<city>.json)
  python city_pipeline.py --city Perth --total 100 --scrape-only

  # Only fetch leads
  python city_pipeline.py --city Melbourne --total 100 --fetch-only
"""

from __future__ import annotations

import argparse
import logging
import sys

from dotenv import load_dotenv

load_dotenv()

from scrape_restaurant_places import run_places_scrape_pipeline  # noqa: E402
from scrape_restaurant_data import run_restaurant_scrape_pipeline  # noqa: E402
from tuvi_outreach_agent import (  # noqa: E402
    Config,
    default_leads_output_path,
    default_scrape_output_path,
    resolve_cities,
    run_lead_fetch_pipeline,
)

SUPPORTED_CITIES = ("Sydney", "Melbourne", "Perth", "Adelaide", "Brisbane")


def run_city_pipeline(
    city: str,
    total: int = 100,
    *,
    fetch: bool = True,
    scrape: bool = True,
    per_page: int = 100,
    max_pages: int = 20,
    max_reviews: int = 40,
    max_size_mb: float = 50.0,
    engine: str = "places",
    cfg: Config | None = None,
) -> dict:
    """Run fetch + scrape for one city. Returns paths summary."""
    cfg = cfg or Config()
    cities = resolve_cities(city=city)
    city_label = cities[0].split(",")[0]

    leads_path = default_leads_output_path(cities, cfg)
    scrape_path = default_scrape_output_path(city, cfg)
    result = {
        "city": city_label,
        "total": total,
        "leads_file": str(leads_path),
        "scrape_file": str(scrape_path),
        "leads_count": 0,
    }

    if fetch:
        logging.info(f"═══ [{city_label}] STEP 1/2 — Fetch {total} leads from Apollo ═══")
        leads, leads_path = run_lead_fetch_pipeline(
            cfg=cfg,
            cities=cities,
            per_page=per_page,
            max_pages=max_pages,
            target_per_city=total,
            output_path=leads_path,
        )
        result["leads_count"] = len(leads)
        result["leads_file"] = str(leads_path)
        logging.info(f"  ✅ {len(leads)} leads → {leads_path}")

    if scrape:
        engine_label = "Google Places API" if engine == "places" else "SerpAPI"
        logging.info(f"═══ [{city_label}] STEP 2/2 — Scrape {total} restaurants via {engine_label} ═══")
        scrape_kwargs = dict(
            cfg=cfg,
            input_path=leads_path,
            output_path=scrape_path,
            city=city,
            max_size_mb=max_size_mb,
            limit=total,
        )
        if engine == "places":
            scrape_kwargs["max_reviews"] = min(max_reviews, 5)
            restaurants, scrape_path = run_places_scrape_pipeline(**scrape_kwargs)
        else:
            scrape_kwargs["max_reviews"] = max_reviews
            restaurants, scrape_path = run_restaurant_scrape_pipeline(**scrape_kwargs)
        ok = sum(1 for r in restaurants if r.get("scrape_status") == "success")
        result["scrape_file"] = str(scrape_path)
        result["scraped_count"] = len(restaurants)
        result["scrape_success"] = ok
        logging.info(f"  ✅ {ok}/{len(restaurants)} scraped → {scrape_path}")

    return result


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Fetch restaurant leads + scrape full data per city",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=(
            "Examples:\n"
            "  python city_pipeline.py --city Sydney --total 100\n"
            "  python city_pipeline.py --city Perth --total 100 --scrape-only\n"
            "  python city_pipeline.py --cities Sydney Melbourne Perth --total 100\n"
        ),
    )
    city_group = parser.add_mutually_exclusive_group(required=True)
    city_group.add_argument("--city", metavar="CITY", help="One city: Sydney, Melbourne, Perth, ...")
    city_group.add_argument(
        "--cities", nargs="+", metavar="CITY",
        help="Multiple cities (e.g. --cities Sydney Melbourne Perth)",
    )
    parser.add_argument("--total", type=int, default=100, metavar="N",
                        help="Leads to fetch & scrape per city (default: 100)")
    parser.add_argument("--per-page", type=int, default=100)
    parser.add_argument("--max-pages", type=int, default=20)
    parser.add_argument("--max-reviews", type=int, default=40)
    parser.add_argument("--max-size-mb", type=float, default=50.0)
    parser.add_argument("--engine", choices=["places", "serpapi"], default="places",
                        help="Scrape engine: places (Google Maps API, default) or serpapi")
    parser.add_argument("--fetch-only", action="store_true", help="Only Apollo fetch, skip scrape")
    parser.add_argument("--scrape-only", action="store_true", help="Only scrape, skip Apollo fetch")
    parser.add_argument("-v", "--verbose", action="store_true")
    args = parser.parse_args()

    if args.fetch_only and args.scrape_only:
        print("ERROR: Use only one of --fetch-only or --scrape-only", file=sys.stderr)
        return 1

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s  [%(levelname)s]  %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    cfg = Config()
    if not args.scrape_only and not cfg.APOLLO_API_KEY:
        print("ERROR: APOLLO_API_KEY missing in .env", file=sys.stderr)
        return 1
    if not args.fetch_only and args.engine == "serpapi" and not cfg.SERPAPI_KEY:
        print("ERROR: SERPAPI_KEY missing in .env (or use --engine places)", file=sys.stderr)
        return 1
    if not args.fetch_only and args.engine == "places" and not cfg.PLACES_API_KEY:
        print("ERROR: PLACES_API missing in .env — add Google Places API key", file=sys.stderr)
        return 1

    if args.city:
        city_list = [args.city]
    else:
        city_list = args.cities

    do_fetch = not args.scrape_only
    do_scrape = not args.fetch_only

    print(f"\nPipeline: {', '.join(city_list)} | {args.total} per city | engine={args.engine}")
    print(f"Steps: {'fetch ' if do_fetch else ''}{'scrape' if do_scrape else ''}\n")

    summaries = []
    for city in city_list:
        try:
            summary = run_city_pipeline(
                city=city,
                total=args.total,
                fetch=do_fetch,
                scrape=do_scrape,
                per_page=args.per_page,
                max_pages=args.max_pages,
                max_reviews=args.max_reviews,
                max_size_mb=args.max_size_mb,
                engine=args.engine,
                cfg=cfg,
            )
            summaries.append(summary)
        except ValueError as e:
            logging.error(f"  ❌ {city}: {e}")
            return 1

    print("\n" + "═" * 60)
    print("  PIPELINE COMPLETE")
    print("═" * 60)
    for s in summaries:
        print(f"\n  {s['city']}:")
        if do_fetch:
            print(f"    Leads   → {s['leads_file']} ({s.get('leads_count', '?')} leads)")
        if do_scrape:
            print(f"    Scrape  → {s['scrape_file']} ({s.get('scrape_success', '?')}/{s.get('scraped_count', '?')} ok)")
    print()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
