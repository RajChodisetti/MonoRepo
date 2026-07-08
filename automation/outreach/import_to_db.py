#!/usr/bin/env python3
"""
Import outreach leads + scraped restaurant JSON into PostgreSQL.

Reads DATABASE_URL from automation/outreach/.env
Imports all leads/*.json and data/*.json files.
"""

from __future__ import annotations

import argparse
import json
import os
import sys
import uuid
from pathlib import Path

from dotenv import load_dotenv

SCRIPT_DIR = Path(__file__).resolve().parent
MONOREPO_ROOT = SCRIPT_DIR.parent.parent
LEADS_DIR = SCRIPT_DIR / "leads"
DATA_DIR = SCRIPT_DIR / "data"
DEFAULT_RESTAURANTS_FILE = MONOREPO_ROOT / "data" / "restaurants_data.json"
IMPORTED_MENU_NAME = "Imported Menu"


def load_env() -> str:
    load_dotenv(SCRIPT_DIR / ".env")
    url = os.getenv("DATABASE_URL", "").strip()
    if not url:
        raise SystemExit("DATABASE_URL missing in automation/outreach/.env")
    return url


def verify_schema(cur) -> None:
    cur.execute(
        """
        SELECT EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public' AND table_name = 'restaurant_profiles'
        )
        """
    )
    if not cur.fetchone()[0]:
        raise SystemExit(
            "Database schema not ready — restaurant_profiles table missing.\n"
            "Run from MonoRepo root:\n"
            "  make db-up        # start Docker Postgres\n"
            "  make migrate-up   # apply migrations\n"
            "Then retry: python import_to_db.py --restaurants-only"
        )


def verify_ocr_schema(cur) -> None:
    verify_schema(cur)
    cur.execute(
        """
        SELECT EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = 'restaurant_profiles'
              AND column_name = 'ocr_verified'
        )
        """
    )
    if not cur.fetchone()[0]:
        raise SystemExit(
            "Migration 000013 not applied — ocr_verified column missing.\n"
            "Run: make migrate-up"
        )


def get_conn(database_url: str):
    try:
        import psycopg2
        import psycopg2.extras
    except ImportError:
        raise SystemExit("Install psycopg2-binary: pip install psycopg2-binary")

    return psycopg2.connect(database_url)


def jdump(obj) -> str:
    return json.dumps(obj if obj is not None else None)


def first_menu_image_url(images, menu_board_urls: set[str] | None = None) -> str:
    """Pick first food photo URL — never a menu board image."""
    menu_board_urls = menu_board_urls or set()
    if not images:
        return ""
    if isinstance(images, list):
        for img in images:
            if isinstance(img, dict):
                url = str(img.get("url") or img.get("thumbnail") or "").strip()
                img_type = str(img.get("image_type") or img.get("source") or "").lower()
                if not url or url in menu_board_urls:
                    continue
                if img_type in ("menu_document", "menu_ocr", "menu_list"):
                    continue
                return url
            if isinstance(img, str) and img.strip() and img.strip() not in menu_board_urls:
                return img.strip()
    return ""


def image_record_url(img) -> str:
    if isinstance(img, str):
        return img.strip()
    if isinstance(img, dict):
        return str(img.get("url") or img.get("thumbnail") or "").strip()
    return ""


def sync_classified_images(cur, restaurant_id: uuid.UUID, record: dict) -> tuple[int, int]:
    """Replace menu_images and gallery_images from OCR-classified scrape JSON."""
    images = record.get("images") or {}
    menu_photos = images.get("menu_photos") or []
    gallery = images.get("gallery") or []

    cur.execute("DELETE FROM menu_images WHERE restaurant_id = %s", (restaurant_id,))
    cur.execute("DELETE FROM gallery_images WHERE restaurant_id = %s", (restaurant_id,))

    menu_count = 0
    for i, img in enumerate(menu_photos):
        url = image_record_url(img)
        if not url:
            continue
        thumb = url
        image_type = "menu_document"
        confidence = None
        title = ""
        metadata = {}
        if isinstance(img, dict):
            thumb = str(img.get("thumbnail") or url).strip()
            image_type = str(img.get("image_type") or "menu_document").strip()
            confidence = img.get("confidence")
            title = str(img.get("title") or "").strip()
            metadata = {
                k: v for k, v in img.items()
                if k not in ("url", "thumbnail", "image_type", "confidence", "title")
            }
        cur.execute(
            """
            INSERT INTO menu_images (
                restaurant_id, url, thumbnail_url, image_type, confidence,
                title, source, sort_order, metadata
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s::jsonb)
            ON CONFLICT (restaurant_id, url) DO UPDATE SET
                thumbnail_url = EXCLUDED.thumbnail_url,
                image_type = EXCLUDED.image_type,
                confidence = EXCLUDED.confidence,
                title = EXCLUDED.title,
                source = EXCLUDED.source,
                sort_order = EXCLUDED.sort_order,
                metadata = EXCLUDED.metadata,
                updated_at = now()
            """,
            (
                restaurant_id,
                url,
                thumb,
                image_type,
                confidence,
                title,
                "menu_ocr",
                i,
                jdump(metadata),
            ),
        )
        menu_count += 1

    gallery_count = 0
    for i, img in enumerate(gallery):
        url = image_record_url(img)
        if not url:
            continue
        thumb = url
        image_type = "other"
        confidence = None
        title = ""
        metadata = {}
        if isinstance(img, dict):
            thumb = str(img.get("thumbnail") or url).strip()
            image_type = str(img.get("image_type") or img.get("title") or "other").strip()
            if image_type == "":
                image_type = "other"
            confidence = img.get("confidence")
            title = str(img.get("title") or "").strip()
            metadata = {
                k: v for k, v in img.items()
                if k not in ("url", "thumbnail", "image_type", "confidence", "title")
            }
        cur.execute(
            """
            INSERT INTO gallery_images (
                restaurant_id, url, thumbnail_url, image_type, confidence,
                title, source, sort_order, metadata
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s::jsonb)
            ON CONFLICT (restaurant_id, url) DO UPDATE SET
                thumbnail_url = EXCLUDED.thumbnail_url,
                image_type = EXCLUDED.image_type,
                confidence = EXCLUDED.confidence,
                title = EXCLUDED.title,
                source = EXCLUDED.source,
                sort_order = EXCLUDED.sort_order,
                metadata = EXCLUDED.metadata,
                updated_at = now()
            """,
            (
                restaurant_id,
                url,
                thumb,
                image_type,
                confidence,
                title,
                "menu_ocr",
                i,
                jdump(metadata),
            ),
        )
        gallery_count += 1

    return menu_count, gallery_count


def upsert_lead(cur, lead: dict) -> uuid.UUID | None:
    company = lead.get("company") or {}
    contact = lead.get("contact") or {}
    location = lead.get("location") or {}
    source = lead.get("source") or {}

    name = (company.get("name") or "").strip()
    if not name:
        return None

    email = (contact.get("email") or "").strip()
    person_id = (source.get("person_id") or "").strip()

    if person_id:
        cur.execute(
            """
            SELECT r.id FROM restaurants r
            JOIN restaurant_profiles p ON p.restaurant_id = r.id
            WHERE p.apollo_lead->'source'->>'person_id' = %s
            LIMIT 1
            """,
            (person_id,),
        )
        row = cur.fetchone()
        if row:
            cur.execute(
                "UPDATE restaurants SET name = %s, email = COALESCE(NULLIF(%s, ''), email), updated_at = now() WHERE id = %s",
                (name, email, row[0]),
            )
            cur.execute(
                """
                INSERT INTO restaurant_profiles (restaurant_id, city, country, apollo_lead, scrape_status)
                VALUES (%s, %s, %s, %s::jsonb, 'lead_only')
                ON CONFLICT (restaurant_id) DO UPDATE SET
                    apollo_lead = EXCLUDED.apollo_lead,
                    city = COALESCE(NULLIF(EXCLUDED.city, ''), restaurant_profiles.city),
                    updated_at = now()
                """,
                (
                    row[0],
                    (location.get("city") or "").strip(),
                    (location.get("country") or "Australia").strip(),
                    jdump(lead),
                ),
            )
            return row[0]

    cur.execute(
        """
        SELECT id FROM restaurants
        WHERE lower(name) = lower(%s) AND (%s = '' OR email = %s)
        LIMIT 1
        """,
        (name, email, email),
    )
    row = cur.fetchone()
    if row:
        restaurant_id = row[0]
    else:
        cur.execute(
            "INSERT INTO restaurants (name, email, status) VALUES (%s, %s, 'lead') RETURNING id",
            (name, email),
        )
        restaurant_id = cur.fetchone()[0]

    cur.execute(
        """
        INSERT INTO restaurant_profiles (restaurant_id, city, country, apollo_lead, scrape_status)
        VALUES (%s, %s, %s, %s::jsonb, 'lead_only')
        ON CONFLICT (restaurant_id) DO UPDATE SET
            apollo_lead = EXCLUDED.apollo_lead,
            city = COALESCE(NULLIF(EXCLUDED.city, ''), restaurant_profiles.city),
            updated_at = now()
        """,
        (
            restaurant_id,
            (location.get("city") or "").strip(),
            (location.get("country") or "Australia").strip(),
            jdump(lead),
        ),
    )
    return restaurant_id


def _menu_board_urls(record: dict) -> set[str]:
    urls: set[str] = set()
    for img in (record.get("images") or {}).get("menu_photos") or []:
        u = image_record_url(img)
        if u:
            urls.add(u)
    return urls


def ensure_menu_id(cur, restaurant_id: uuid.UUID) -> uuid.UUID:
    cur.execute(
        "SELECT id FROM menus WHERE restaurant_id = %s AND name = %s LIMIT 1",
        (restaurant_id, IMPORTED_MENU_NAME),
    )
    row = cur.fetchone()
    if row:
        return row[0]
    cur.execute(
        "INSERT INTO menus (restaurant_id, name, status) VALUES (%s, %s, 'active') RETURNING id",
        (restaurant_id, IMPORTED_MENU_NAME),
    )
    return cur.fetchone()[0]


def sync_menu_items_and_reviews(
    cur,
    restaurant_id: uuid.UUID,
    record: dict,
    *,
    sync_reviews: bool = True,
) -> tuple[int, int]:
    """Sync menu items, optional reviews, and classified image tables."""
    board_urls = _menu_board_urls(record)
    menu_id = ensure_menu_id(cur, restaurant_id)

    cur.execute("DELETE FROM menu_items WHERE menu_id = %s", (menu_id,))
    for i, item in enumerate(record.get("menu_items") or []):
        name = (item.get("name") or "").strip()
        category = (item.get("category") or "").strip()
        if not name and not category:
            continue
        if not name:
            name = category
        cur.execute(
            """
            INSERT INTO menu_items (
                menu_id, name, description, price, price_text, category,
                image_url, images, sort_order
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s::jsonb, %s)
            """,
            (
                menu_id,
                name,
                (item.get("description") or "").strip(),
                item.get("price_numeric"),
                (item.get("price") or "").strip(),
                category,
                first_menu_image_url(item.get("images"), board_urls),
                jdump(item.get("images") or []),
                i,
            ),
        )

    if sync_reviews:
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

    return sync_classified_images(cur, restaurant_id, record)


def apply_verified_record(cur, restaurant_id: uuid.UUID, record: dict) -> tuple[int, int]:
    """
    Write OCR-verified scrape JSON back to an existing restaurant row.
    Updates profile images/raw_public_data, menu items, and classified image tables.
    """
    raw_record = json.dumps(record)
    cur.execute(
        """
        UPDATE restaurant_profiles
        SET
            images = %s::jsonb,
            raw_public_data = %s::jsonb,
            scrape_errors = %s::jsonb,
            updated_at = now()
        WHERE restaurant_id = %s
        """,
        (
            jdump(record.get("images") or {}),
            raw_record,
            jdump(record.get("errors") or []),
            restaurant_id,
        ),
    )
    return sync_menu_items_and_reviews(cur, restaurant_id, record, sync_reviews=False)


def mark_ocr_verified(cur, restaurant_id: uuid.UUID, *, note: str = "") -> None:
    errors = jdump([note]) if note else "[]"
    cur.execute(
        """
        UPDATE restaurant_profiles
        SET
            ocr_verified = true,
            ocr_verified_at = now(),
            ocr_verification_errors = %s::jsonb,
            updated_at = now()
        WHERE restaurant_id = %s
        """,
        (errors if note else "[]", restaurant_id),
    )


def append_ocr_verification_error(cur, restaurant_id: uuid.UUID, message: str) -> None:
    cur.execute(
        """
        UPDATE restaurant_profiles
        SET
            ocr_verification_errors = COALESCE(ocr_verification_errors, '[]'::jsonb) || %s::jsonb,
            updated_at = now()
        WHERE restaurant_id = %s
        """,
        (jdump([message]), restaurant_id),
    )


def fetch_unverified_leads(cur, limit: int) -> list[tuple[uuid.UUID, dict, str]]:
    cur.execute(
        """
        SELECT rp.restaurant_id, rp.raw_public_data, r.name
        FROM restaurant_profiles rp
        JOIN restaurants r ON r.id = rp.restaurant_id
        WHERE rp.ocr_verified = false
          AND rp.raw_public_data IS NOT NULL
          AND rp.raw_public_data <> '{}'::jsonb
        ORDER BY rp.created_at ASC
        LIMIT %s
        FOR UPDATE OF rp SKIP LOCKED
        """,
        (limit,),
    )
    rows = []
    for restaurant_id, raw_public_data, name in cur.fetchall():
        if isinstance(raw_public_data, dict):
            record = raw_public_data
        elif isinstance(raw_public_data, str):
            record = json.loads(raw_public_data)
        else:
            record = json.loads(json.dumps(raw_public_data))
        rows.append((restaurant_id, record, name or ""))
    return rows


def upsert_restaurant_record(cur, record: dict, source_file: str) -> tuple[uuid.UUID, bool]:
    """Returns (restaurant_id, skipped_duplicate)."""
    try:
        from menu_image_ocr import _sanitize_menu_item_images
        _sanitize_menu_item_images(record)
    except ImportError:
        pass

    google = record.get("google") or {}
    place_id = (google.get("place_id") or "").strip()

    if place_id:
        cur.execute(
            "SELECT restaurant_id FROM restaurant_profiles WHERE google_place_id = %s LIMIT 1",
            (place_id,),
        )
        row = cur.fetchone()
        if row:
            restaurant_id = row[0]
            cur.execute(
                "UPDATE restaurants SET name = %s, email = %s, updated_at = now() WHERE id = %s",
                (
                    (record.get("name") or "").strip(),
                    ((record.get("contact") or {}).get("email") or "").strip(),
                    restaurant_id,
                ),
            )
        else:
            cur.execute(
                "INSERT INTO restaurants (name, email, status) VALUES (%s, %s, 'lead') RETURNING id",
                (
                    (record.get("name") or "").strip(),
                    ((record.get("contact") or {}).get("email") or "").strip(),
                ),
            )
            restaurant_id = cur.fetchone()[0]
    else:
        cur.execute(
            "INSERT INTO restaurants (name, email, status) VALUES (%s, %s, 'lead') RETURNING id",
            (
                (record.get("name") or "").strip(),
                ((record.get("contact") or {}).get("email") or "").strip(),
            ),
        )
        restaurant_id = cur.fetchone()[0]

    contact = record.get("contact") or {}
    location = record.get("location") or {}
    coords = location.get("coordinates") or {}
    raw_record = json.dumps(record)

    cur.execute(
        """
        INSERT INTO restaurant_profiles (
            restaurant_id, opening_hours, phone, website, address, city, state, country,
            latitude, longitude, google_place_id, google_data_id, rating, reviews_count,
            price_level, cuisines, owners, images, apollo_lead, scrape_status, scrape_errors,
            dietary_options, raw_public_data
        ) VALUES (
            %s, %s::jsonb, %s, %s, %s, %s, %s, %s,
            %s, %s, %s, %s, %s, %s,
            %s, %s::jsonb, %s::jsonb, %s::jsonb, %s::jsonb, %s, %s::jsonb,
            %s::jsonb, %s::jsonb
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
            google_place_id = EXCLUDED.google_place_id,
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
            ocr_verified = false,
            ocr_verified_at = NULL,
            ocr_verification_errors = '[]'::jsonb,
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
        ),
    )

    sync_menu_items_and_reviews(cur, restaurant_id, record, sync_reviews=True)

    return restaurant_id, False


def import_leads(cur) -> int:
    count = 0
    for path in sorted(LEADS_DIR.glob("*.json")):
        payload = json.loads(path.read_text(encoding="utf-8"))
        for lead in payload.get("leads") or []:
            if upsert_lead(cur, lead):
                count += 1
        print(f"  leads file: {path.name} → {len(payload.get('leads') or [])} leads processed")
    return count


def import_restaurant_data(cur, extra_files: list[Path] | None = None) -> tuple[int, int]:
    imported = 0
    skipped = 0
    seen_place_ids: set[str] = set()

    paths: list[Path] = []
    if extra_files:
        paths.extend(extra_files)
    paths.extend(sorted(DATA_DIR.glob("*.json")))

    seen_paths: set[Path] = set()
    unique_paths = []
    for path in paths:
        resolved = path.resolve()
        if resolved in seen_paths or not path.is_file():
            continue
        seen_paths.add(resolved)
        unique_paths.append(path)

    for path in unique_paths:
        payload = json.loads(path.read_text(encoding="utf-8"))
        file_imported = 0
        for record in payload.get("restaurants") or []:
            place_id = ((record.get("google") or {}).get("place_id") or "").strip()
            if place_id:
                if place_id in seen_place_ids:
                    skipped += 1
                    continue
                seen_place_ids.add(place_id)

            upsert_restaurant_record(cur, record, str(path))
            imported += 1
            file_imported += 1

        cur.execute(
            """
            INSERT INTO restaurant_data_imports (source_file, meta, restaurants_imported, restaurants_skipped)
            VALUES (%s, %s::jsonb, %s, %s)
            """,
            (str(path), jdump(payload.get("meta") or {}), file_imported, 0),
        )
        print(f"  data file: {path.name} → {file_imported} restaurants imported")

    return imported, skipped


def run_import(
    *,
    restaurants_only: bool = False,
    data_files: list[Path] | None = None,
    do_import_leads: bool = True,
    niche_type: str = "restaurant",
) -> tuple[int, int]:
    """Import leads and/or scrape data into PostgreSQL. Returns (imported, skipped)."""
    database_url = load_env()
    conn = get_conn(database_url)
    conn.autocommit = False

    files = list(data_files or [])
    if niche_type == "restaurant" and DEFAULT_RESTAURANTS_FILE.is_file():
        if DEFAULT_RESTAURANTS_FILE not in files:
            files.insert(0, DEFAULT_RESTAURANTS_FILE)

    try:
        with conn.cursor() as cur:
            verify_schema(cur)
            if do_import_leads and not restaurants_only:
                print("\n[1/2] Importing leads…")
                import_leads(cur)
            elif restaurants_only:
                print("\nSkipping leads import (--restaurants-only)")

            step = "[2/2]" if do_import_leads and not restaurants_only else "[1/1]"
            print(f"\n{step} Importing restaurant scrape data…")
            if files:
                for path in files:
                    print(f"  including: {path}")
            imported, skipped = import_restaurant_data(cur, files or None)
        conn.commit()
        return imported, skipped
    except Exception:
        conn.rollback()
        raise
    finally:
        conn.close()


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Import outreach leads and scraped restaurant JSON into PostgreSQL.",
    )
    parser.add_argument(
        "--restaurants-only",
        action="store_true",
        help="Skip leads import; only import restaurant JSON data.",
    )
    parser.add_argument(
        "--data-file",
        action="append",
        default=[],
        type=Path,
        help="Extra restaurant JSON file to import (default includes MonoRepo/data/restaurants_data.json).",
    )
    parser.add_argument(
        "--type",
        default="restaurant",
        help="Niche type for default data file glob (restaurant, dentist, plumber).",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    try:
        imported, skipped = run_import(
            restaurants_only=args.restaurants_only,
            data_files=list(args.data_file),
            do_import_leads=not args.restaurants_only,
            niche_type=args.type,
        )
    except Exception as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 1

    print("\nDone.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
