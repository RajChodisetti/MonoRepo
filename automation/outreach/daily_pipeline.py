#!/usr/bin/env python3
"""
Daily incremental scrape pipeline — cron entrypoint.

Fetches Apollo leads + scrapes only new/stale restaurants (skips fresh DB records).

Usage:
  python daily_pipeline.py
  python daily_pipeline.py --cities Sydney Melbourne Perth --total 500
  python daily_pipeline.py --refresh-days 7 --scrape-only
"""

from __future__ import annotations

import argparse
import logging
import os
import sys
from datetime import datetime, timezone

from dotenv import load_dotenv

load_dotenv()

from city_pipeline import run_city_pipeline, SUPPORTED_CITIES  # noqa: E402
from tuvi_outreach_agent import Config  # noqa: E402

log = logging.getLogger("daily_pipeline")


def _default_cities() -> list[str]:
    raw = os.getenv("SCRAPE_CITIES", "Sydney,Melbourne,Perth")
    return [c.strip() for c in raw.split(",") if c.strip()]


def _default_total() -> int:
    return int(os.getenv("SCRAPE_DAILY_LIMIT", "500"))


def _default_refresh_days() -> int:
    return int(os.getenv("SCRAPE_REFRESH_DAYS", "7"))


def main() -> int:
    parser = argparse.ArgumentParser(description="Daily incremental restaurant scrape (cron entrypoint)")
    parser.add_argument("--cities", nargs="+", default=None, metavar="CITY",
                        help=f"Cities to process (default: SCRAPE_CITIES env or Sydney Melbourne Perth)")
    parser.add_argument("--total", type=int, default=None,
                        help="Leads per city (default: SCRAPE_DAILY_LIMIT env or 500)")
    parser.add_argument("--refresh-days", type=int, default=None,
                        help="Re-scrape after N days (default: SCRAPE_REFRESH_DAYS env or 7)")
    parser.add_argument("--fetch-only", action="store_true")
    parser.add_argument("--scrape-only", action="store_true")
    parser.add_argument("--no-incremental", dest="incremental", action="store_false", default=True)
    parser.add_argument("-v", "--verbose", action="store_true")
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s  [%(levelname)s]  %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    cities = args.cities or _default_cities()
    total = args.total if args.total is not None else _default_total()
    refresh_days = args.refresh_days if args.refresh_days is not None else _default_refresh_days()

    if args.fetch_only and args.scrape_only:
        print("ERROR: Use only one of --fetch-only or --scrape-only", file=sys.stderr)
        return 1

    cfg = Config()
    do_fetch = not args.scrape_only
    do_scrape = not args.fetch_only

    if do_fetch and not cfg.APOLLO_API_KEY:
        log.error("APOLLO_API_KEY missing — required for lead fetch")
        return 1
    if do_scrape and not cfg.PLACES_API_KEY:
        log.error("PLACES_API missing — required for Places scrape")
        return 1

    started = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    log.info(f"Daily pipeline started at {started}")
    log.info(f"Cities: {cities} | total/city: {total} | refresh_days: {refresh_days} | incremental: {args.incremental}")

    totals = {"skipped": 0, "new": 0, "refreshed": 0, "failed": 0, "scraped": 0}
    failed_cities: list[str] = []

    for city in cities:
        if city not in SUPPORTED_CITIES and city.split(",")[0] not in SUPPORTED_CITIES:
            log.warning(f"City '{city}' not in supported list — proceeding anyway")
        try:
            summary = run_city_pipeline(
                city=city,
                total=total,
                fetch=do_fetch,
                scrape=do_scrape,
                incremental=args.incremental,
                refresh_days=refresh_days,
                cfg=cfg,
            )
            totals["skipped"] += summary.get("skipped_existing", 0)
            totals["new"] += summary.get("scraped_new", 0)
            totals["refreshed"] += summary.get("refreshed_stale", 0)
            totals["failed"] += summary.get("scrape_failed", 0)
            totals["scraped"] += summary.get("scrape_success", 0)
        except Exception as exc:
            log.error(f"City {city} failed: {exc}")
            failed_cities.append(city)

    log.info(
        f"Daily pipeline done — scraped={totals['scraped']} skipped={totals['skipped']} "
        f"new={totals['new']} refreshed={totals['refreshed']} failed={totals['failed']}"
    )
    if failed_cities:
        log.error(f"Failed cities: {', '.join(failed_cities)}")
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
