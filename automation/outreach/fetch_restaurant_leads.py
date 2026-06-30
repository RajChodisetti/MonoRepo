#!/usr/bin/env python3
"""
Restaurant Lead Fetch Pipeline
==============================
Fetches restaurant decision-maker leads from Apollo.io.

Usage:
  python fetch_restaurant_leads.py --city Sydney
  python fetch_restaurant_leads.py --city Perth
  python fetch_restaurant_leads.py --cities Sydney Melbourne
  python fetch_restaurant_leads.py                    # all 5 cities
"""

import argparse
import logging
import sys

from tuvi_outreach_agent import (
    AUSTRALIAN_RESTAURANT_CITIES,
    Config,
    default_leads_output_path,
    resolve_cities,
    run_lead_fetch_pipeline,
)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Fetch restaurant leads from Australian cities via Apollo.io",
        epilog=(
            "Examples:\n"
            "  python fetch_restaurant_leads.py --city Sydney\n"
            "  python fetch_restaurant_leads.py --city Perth --max-pages 10\n"
            "  python fetch_restaurant_leads.py --cities Sydney Melbourne\n"
        ),
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    city_group = parser.add_mutually_exclusive_group()
    city_group.add_argument(
        "--city", metavar="CITY",
        help="One city only: Sydney, Melbourne, Perth, Adelaide, Brisbane",
    )
    city_group.add_argument(
        "--cities", nargs="+", metavar="CITY",
        help="Multiple cities (e.g. --cities Sydney Perth)",
    )
    parser.add_argument("--per-page", type=int, default=100,
                        help="Apollo results per page (default: 100, max Apollo allows)")
    parser.add_argument("--max-pages", type=int, default=20,
                        help="Max pages per city as safety cap (default: 20)")
    parser.add_argument("--total", type=int, default=None, metavar="N",
                        help="Target leads per city (default: 100 when using --city)")
    parser.add_argument("--output", metavar="FILE",
                        help="Output JSON path (auto: leads/lead_sydney.json for --city Sydney)")
    parser.add_argument("-v", "--verbose", action="store_true", help="Debug logging")
    args = parser.parse_args()

    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)

    cfg = Config()
    if not cfg.APOLLO_API_KEY:
        print("ERROR: Set APOLLO_API_KEY in your .env file.", file=sys.stderr)
        return 1

    try:
        cities = resolve_cities(city=args.city, cities=args.cities)
    except ValueError as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return 1

    output = args.output or str(default_leads_output_path(cities, cfg))
    target = args.total if args.total is not None else (100 if len(cities) == 1 else None)

    print("Cities:", ", ".join(c.split(",")[0] for c in cities))
    print(f"Output: {output}")
    print(f"Target: {target or 'all available'} leads per city")
    print(f"Settings: {args.per_page} per page, up to {args.max_pages} pages per city\n")

    try:
        leads, json_path = run_lead_fetch_pipeline(
            cfg=cfg,
            cities=cities,
            per_page=args.per_page,
            max_pages=args.max_pages,
            target_per_city=target,
            output_path=output,
        )
    except ValueError as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return 1

    print(f"\nDone — {len(leads)} restaurant leads saved to:")
    print(f"  {json_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
