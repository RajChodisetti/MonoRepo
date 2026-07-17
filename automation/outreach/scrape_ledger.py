"""
PostgreSQL scrape ledger — dedup lookup, skip decisions, upsert, audit runs.
"""

from __future__ import annotations

import json
import os
import uuid
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Any, Literal

from dotenv import load_dotenv

from identity import content_hash, identity_hash_from_record

SCRIPT_DIR = Path(__file__).resolve().parent
IMPORTED_MENU_NAME = "Imported Menu"

ScrapeAction = Literal["skip", "scrape_new", "refresh_stale"]


@dataclass
class LedgerRow:
    restaurant_id: uuid.UUID
    google_place_id: str | None
    identity_hash: str | None
    last_scraped_at: datetime | None
    content_hash: str | None
    scrape_status: str | None


def load_database_url() -> str:
    load_dotenv(SCRIPT_DIR / ".env")
    url = os.getenv("DATABASE_URL", "").strip()
    if not url:
        raise ValueError("DATABASE_URL missing in automation/outreach/.env")
    return url


def get_connection(database_url: str | None = None):
    try:
        import psycopg2
    except ImportError as exc:
        raise RuntimeError("Install psycopg2-binary: pip install psycopg2-binary") from exc
    return psycopg2.connect(database_url or load_database_url())


def jdump(obj: Any) -> str:
    return json.dumps(obj if obj is not None else None)


def lookup_restaurant(
    cur,
    place_id: str | None,
    identity_hash: str | None,
) -> LedgerRow | None:
    """Find existing restaurant by google_place_id first, then identity_hash."""
    pid = (place_id or "").strip()
    ihash = (identity_hash or "").strip()

    if pid:
        cur.execute(
            """
            SELECT restaurant_id, google_place_id, identity_hash,
                   last_scraped_at, content_hash, scrape_status
            FROM restaurant_profiles
            WHERE google_place_id = %s
            LIMIT 1
            """,
            (pid,),
        )
        row = cur.fetchone()
        if row:
            return LedgerRow(*row)

    if ihash:
        cur.execute(
            """
            SELECT restaurant_id, google_place_id, identity_hash,
                   last_scraped_at, content_hash, scrape_status
            FROM restaurant_profiles
            WHERE identity_hash = %s
            LIMIT 1
            """,
            (ihash,),
        )
        row = cur.fetchone()
        if row:
            return LedgerRow(*row)

    return None


def should_scrape(
    row: LedgerRow | None,
    *,
    incremental: bool = True,
    refresh_days: int = 7,
) -> ScrapeAction:
    if not incremental:
        return "scrape_new" if row is None else "refresh_stale"
    if row is None:
        return "scrape_new"
    if row.last_scraped_at is None:
        return "scrape_new"
    if row.scrape_status in ("error", "failed", "not_found"):
        return "scrape_new"

    cutoff = datetime.now(timezone.utc) - timedelta(days=refresh_days)
    last = row.last_scraped_at
    if last.tzinfo is None:
        last = last.replace(tzinfo=timezone.utc)

    if last >= cutoff:
        return "skip"
    return "refresh_stale"


def first_menu_image_url(images) -> str:
    if not images:
        return ""
    if isinstance(images, list) and images:
        first = images[0]
        if isinstance(first, dict):
            return str(first.get("url") or first.get("thumbnail") or "")
        if isinstance(first, str):
            return first
    return ""


def upsert_after_scrape(
    cur,
    record: dict,
    *,
    identity_hash: str,
    discovery_rank: int | None = None,
    set_scraped_at: bool = True,
) -> uuid.UUID:
    """
    Upsert restaurant + profile + menus/reviews after a successful scrape.
    Returns restaurant_id.
    """
    google = record.get("google") or {}
    place_id = (google.get("place_id") or "").strip()
    chash = content_hash(record)
    contact = record.get("contact") or {}
    location = record.get("location") or {}
    coords = location.get("coordinates") or {}
    email = (contact.get("email") or "").strip()
    name = (record.get("name") or "").strip()
    raw_record = json.dumps(record)

    existing = lookup_restaurant(cur, place_id, identity_hash)
    if existing:
        restaurant_id = existing.restaurant_id
        cur.execute(
            "UPDATE restaurants SET name = %s, email = COALESCE(NULLIF(%s, ''), email), updated_at = now() WHERE id = %s",
            (name, email, restaurant_id),
        )
    else:
        cur.execute(
            "INSERT INTO restaurants (name, email, status) VALUES (%s, %s, 'lead') RETURNING id",
            (name, email),
        )
        restaurant_id = cur.fetchone()[0]

    cur.execute(
        f"""
        INSERT INTO restaurant_profiles (
            restaurant_id, opening_hours, phone, website, address, city, state, country,
            latitude, longitude, google_place_id, google_data_id, rating, reviews_count,
            price_level, cuisines, owners, images, apollo_lead, scrape_status, scrape_errors,
            dietary_options, raw_public_data, identity_hash, content_hash, discovery_rank,
            last_scraped_at
        ) VALUES (
            %s, %s::jsonb, %s, %s, %s, %s, %s, %s,
            %s, %s, %s, %s, %s, %s,
            %s, %s::jsonb, %s::jsonb, %s::jsonb, %s::jsonb, %s, %s::jsonb,
            %s::jsonb, %s::jsonb, %s, %s, %s,
            {"now()" if set_scraped_at and record.get("scrape_status") == "success" else "NULL"}
        )
        ON CONFLICT (restaurant_id) DO UPDATE SET
            opening_hours = EXCLUDED.opening_hours,
            phone = EXCLUDED.phone,
            website = EXCLUDED.website,
            address = EXCLUDED.address,
            city = EXCLUDED.city,
            state = EXCLUDED.state,
            country = EXCLUDED.country,
            latitude = EXCLUDED.latitude,
            longitude = EXCLUDED.longitude,
            google_place_id = COALESCE(EXCLUDED.google_place_id, restaurant_profiles.google_place_id),
            google_data_id = EXCLUDED.google_data_id,
            rating = EXCLUDED.rating,
            reviews_count = EXCLUDED.reviews_count,
            price_level = EXCLUDED.price_level,
            cuisines = EXCLUDED.cuisines,
            owners = EXCLUDED.owners,
            images = EXCLUDED.images,
            apollo_lead = EXCLUDED.apollo_lead,
            scrape_status = EXCLUDED.scrape_status,
            scrape_errors = EXCLUDED.scrape_errors,
            dietary_options = EXCLUDED.dietary_options,
            raw_public_data = EXCLUDED.raw_public_data,
            identity_hash = COALESCE(EXCLUDED.identity_hash, restaurant_profiles.identity_hash),
            content_hash = EXCLUDED.content_hash,
            discovery_rank = COALESCE(EXCLUDED.discovery_rank, restaurant_profiles.discovery_rank),
            last_scraped_at = CASE
                WHEN EXCLUDED.scrape_status = 'success' THEN now()
                ELSE restaurant_profiles.last_scraped_at
            END,
            updated_at = now()
        """,
        (
            restaurant_id,
            jdump(record.get("hours") or {}),
            (contact.get("phone") or "").strip(),
            (contact.get("website") or "").strip(),
            (location.get("address") or "").strip(),
            (location.get("city") or "").strip(),
            (location.get("state") or "").strip(),
            (location.get("country") or "").strip(),
            coords.get("latitude"),
            coords.get("longitude"),
            place_id or None,
            (google.get("data_id") or "").strip() or None,
            record.get("rating"),
            record.get("reviews_count"),
            (record.get("price_level") or "").strip(),
            jdump(record.get("cuisines") or []),
            jdump(record.get("owners") or []),
            jdump(record.get("images") or {}),
            jdump(record.get("apollo_lead") or {}),
            (record.get("scrape_status") or "unknown").strip(),
            jdump(record.get("errors") or []),
            jdump(record.get("cuisines") or []),
            raw_record,
            identity_hash or identity_hash_from_record(record) or None,
            chash,
            discovery_rank,
        ),
    )

    # Menus
    cur.execute(
        "SELECT id FROM menus WHERE restaurant_id = %s AND name = %s LIMIT 1",
        (restaurant_id, IMPORTED_MENU_NAME),
    )
    row = cur.fetchone()
    if row:
        menu_id = row[0]
    else:
        cur.execute(
            "INSERT INTO menus (restaurant_id, name, status) VALUES (%s, %s, 'active') RETURNING id",
            (restaurant_id, IMPORTED_MENU_NAME),
        )
        menu_id = cur.fetchone()[0]

    cur.execute("DELETE FROM menu_items WHERE menu_id = %s", (menu_id,))
    for i, item in enumerate(record.get("menu_items") or []):
        item_name = (item.get("name") or "").strip()
        category = (item.get("category") or "").strip()
        if not item_name and not category:
            continue
        if not item_name:
            item_name = category
        cur.execute(
            """
            INSERT INTO menu_items (
                menu_id, name, description, price, price_text, category,
                image_url, images, sort_order
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s::jsonb, %s)
            """,
            (
                menu_id,
                item_name,
                (item.get("description") or "").strip(),
                item.get("price_numeric"),
                (item.get("price") or "").strip(),
                category,
                first_menu_image_url(item.get("images")),
                jdump(item.get("images") or []),
                i,
            ),
        )

    cur.execute("DELETE FROM restaurant_reviews WHERE restaurant_id = %s", (restaurant_id,))
    for i, rev in enumerate(record.get("reviews") or []):
        cur.execute(
            """
            INSERT INTO restaurant_reviews (
                restaurant_id, reviewer, review_text, stars, review_date, images, source, sort_order
            ) VALUES (%s, %s, %s, %s, %s, %s::jsonb, %s, %s)
            """,
            (
                restaurant_id,
                (rev.get("reviewer") or "").strip(),
                (rev.get("review") or "").strip(),
                rev.get("stars"),
                (rev.get("date") or "").strip(),
                jdump(rev.get("images") or []),
                (rev.get("source") or "").strip(),
                i,
            ),
        )

    return restaurant_id


def start_scrape_run(cur, city: str, meta: dict | None = None) -> uuid.UUID:
    run_id = uuid.uuid4()
    cur.execute(
        """
        INSERT INTO scrape_runs (id, city, status, meta)
        VALUES (%s, %s, 'running', %s::jsonb)
        """,
        (run_id, city, jdump(meta or {})),
    )
    return run_id


def finish_scrape_run(
    cur,
    run_id: uuid.UUID,
    *,
    status: str,
    leads_seen: int = 0,
    skipped_existing: int = 0,
    scraped_new: int = 0,
    refreshed_stale: int = 0,
    failed: int = 0,
    meta: dict | None = None,
) -> None:
    cur.execute(
        """
        UPDATE scrape_runs SET
            finished_at = now(),
            status = %s,
            leads_seen = %s,
            skipped_existing = %s,
            scraped_new = %s,
            refreshed_stale = %s,
            failed = %s,
            meta = COALESCE(meta, '{}'::jsonb) || %s::jsonb
        WHERE id = %s
        """,
        (
            status,
            leads_seen,
            skipped_existing,
            scraped_new,
            refreshed_stale,
            failed,
            jdump(meta or {}),
            run_id,
        ),
    )
