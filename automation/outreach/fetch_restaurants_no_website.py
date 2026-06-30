#!/usr/bin/env python3
"""
Fetch restaurants where website is empty from scraped JSON files.

Usage:
  python fetch_restaurants_no_website.py
  python fetch_restaurants_no_website.py --input data/restaurants_data_melbourne.json
  python fetch_restaurants_no_website.py --input data/restaurants_data_sydney.json --output output/no_website_sydney.json
  python fetch_restaurants_no_website.py --all-cities
"""

from __future__ import annotations

import argparse
import json
from datetime import datetime, timezone
from pathlib import Path


def has_empty_website(restaurant: dict) -> bool:
    website = (restaurant.get("contact") or {}).get("website") or ""
    return not str(website).strip()


def load_restaurants(path: Path) -> list[dict]:
    data = json.loads(path.read_text(encoding="utf-8"))
    if isinstance(data, dict):
        return data.get("restaurants") or []
    if isinstance(data, list):
        return data
    return []


def filter_no_website(restaurants: list[dict]) -> list[dict]:
    return [r for r in restaurants if has_empty_website(r)]


def slim_record(restaurant: dict) -> dict:
    contact = restaurant.get("contact") or {}
    location = restaurant.get("location") or {}
    return {
        "name": restaurant.get("name", ""),
        "rating": restaurant.get("rating"),
        "reviews_count": restaurant.get("reviews_count"),
        "phone": contact.get("phone", ""),
        "email": contact.get("email", ""),
        "website": contact.get("website", ""),
        "address": location.get("address", ""),
        "city": location.get("city", ""),
        "place_id": (restaurant.get("google") or {}).get("place_id", ""),
        "scrape_status": restaurant.get("scrape_status", ""),
    }


def run_query(input_path: Path, output_path: Path | None) -> list[dict]:
    restaurants = load_restaurants(input_path)
    matched = filter_no_website(restaurants)

    result = {
        "meta": {
            "queried_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "query": 'contact.website == ""',
            "source_file": str(input_path),
            "total_restaurants": len(restaurants),
            "matched_count": len(matched),
        },
        "restaurants": [slim_record(r) for r in matched],
    }

    if output_path:
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(json.dumps(result, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

    return matched


def main() -> int:
    parser = argparse.ArgumentParser(description='Fetch restaurants with website=""')
    parser.add_argument("--input", "-i", metavar="FILE", help="Scraped JSON file")
    parser.add_argument("--output", "-o", metavar="FILE", help="Save results JSON")
    parser.add_argument(
        "--all-cities", action="store_true",
        help="Scan data/restaurants_data_*.json",
    )
    args = parser.parse_args()

    if args.all_cities:
        files = sorted(Path("data").glob("restaurants_data_*.json"))
        if not files:
            print("No files in data/restaurants_data_*.json")
            return 1
        total = 0
        for path in files:
            city = path.stem.replace("restaurants_data_", "")
            out = Path("output") / f"no_website_{city}.json"
            matched = run_query(path, out)
            total += len(matched)
            print(f"{path.name}: {len(matched)} with empty website → {out}")
        print(f"\nTotal: {total}")
        return 0

    if not args.input:
        print("ERROR: pass --input FILE or --all-cities", flush=True)
        return 1

    input_path = Path(args.input)
    if not input_path.exists():
        print(f"ERROR: file not found: {input_path}", flush=True)
        return 1

    output_path = Path(args.output) if args.output else None
    matched = run_query(input_path, output_path)

    print(f"Source: {input_path}")
    print(f'Matched (website=""): {len(matched)}')
    for r in matched[:20]:
        loc = r.get("location") or {}
        print(f"  - {r.get('name')} | {loc.get('city', '')} | website={repr((r.get('contact') or {}).get('website', ''))}")
    if len(matched) > 20:
        print(f"  ... and {len(matched) - 20} more")
    if output_path:
        print(f"Saved: {output_path}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
