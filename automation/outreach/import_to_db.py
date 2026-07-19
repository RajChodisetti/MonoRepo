#!/usr/bin/env python3
"""
Import outreach leads + scraped restaurant JSON into PostgreSQL.

Reads DATABASE_URL from automation/outreach/.env
Imports all leads/*.json and data/*.json files.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import sys
import uuid
from pathlib import Path

from dotenv import load_dotenv

from scrape_safety import sanitize_sensitive_urls

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
        raise SystemExit("DATABASE_URL missing from the configured environment")
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
        SELECT COUNT(*) = 7
        FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = 'restaurant_profiles'
              AND column_name IN (
                  'ocr_status',
                  'ocr_started_at',
                  'ocr_completed_at',
                  'ocr_attempts',
                  'ocr_input_fingerprint',
                  'ocr_claim_id',
                  'ocr_claim_fingerprint'
              )
        """
    )
    if not cur.fetchone()[0]:
        raise SystemExit(
            "Migration 000016 not applied — ocr_status column missing.\n"
            "Run: make migrate-up"
        )
    cur.execute(
        """
        SELECT
            EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_schema = 'public'
                  AND table_name = 'demo_sites'
                  AND column_name = 'source_profile_fingerprint'
            )
            AND EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_schema = 'public'
                  AND table_name = 'email_campaigns'
                  AND column_name = 'source_profile_fingerprint'
            )
            AND to_regprocedure('lead_artifact_current_profile_fingerprint(uuid)') IS NOT NULL
        """
    )
    if not cur.fetchone()[0]:
        raise SystemExit(
            "Migration 000022 not applied — automatic artifact profile provenance is missing.\n"
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


class StaleOCRClaim(RuntimeError):
    """The OCR input or worker claim changed before a result was persisted."""


def lock_restaurant_workflow(cur, restaurant_id: uuid.UUID) -> None:
    """Acquire the shared cross-language transaction lock before row locks."""
    cur.execute(
        "SELECT pg_advisory_xact_lock(hashtextextended(%s, 0))",
        (f"lead-workflow:{restaurant_id}",),
    )


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


def ocr_input_fingerprint(record: dict) -> str:
    """Hash only stable image inputs that can materially change OCR output."""
    refs: set[str] = set()

    def add(value: str) -> None:
        value = str(value or "").strip()
        if value:
            refs.add(value)

    images = record.get("images") or {}
    add(images.get("thumbnail") or "")
    for collection in ("gallery", "menu_photos"):
        for image in images.get(collection) or []:
            add(image_record_url(image))
            if isinstance(image, dict):
                add(image.get("source_ref") or "")
    google_place_id = str(((record.get("google") or {}).get("place_id") or "")).strip()
    if google_place_id:
        add(f"google-place:{google_place_id}")
        add(f"google-photo-count:{images.get('google_photo_count') or 0}")
    for item in record.get("menu_items") or []:
        for image in item.get("images") or []:
            add(image_record_url(image))

    encoded = json.dumps(sorted(refs), separators=(",", ":")).encode("utf-8")
    return hashlib.sha256(encoded).hexdigest()


def sync_classified_images(cur, restaurant_id: uuid.UUID, record: dict) -> tuple[int, int]:
    """Replace menu_images and gallery_images from OCR-classified scrape JSON."""
    images = record.get("images") or {}
    menu_photos = images.get("menu_photos") or []
    gallery = images.get("gallery") or []

    # Replace only OCR-managed rows and only after a replacement set actually
    # contains stable URLs. Google Places media URLs are intentionally absent;
    # an empty transient-only result must never erase durable/manual media.
    menu_has_persistent_urls = any(image_record_url(img) for img in menu_photos)
    gallery_has_persistent_urls = any(image_record_url(img) for img in gallery)
    if menu_has_persistent_urls:
        cur.execute(
            "DELETE FROM menu_images WHERE restaurant_id = %s AND source = 'menu_ocr'",
            (restaurant_id,),
        )
    if gallery_has_persistent_urls:
        cur.execute(
            "DELETE FROM gallery_images WHERE restaurant_id = %s AND source = 'menu_ocr'",
            (restaurant_id,),
        )

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


def _invalidate_review_for_artifact_source_change(
    cur,
    restaurant_id: uuid.UUID,
) -> None:
    """Require fresh gates and preparation after identity/public-payload changes."""
    cur.execute(
        """
        UPDATE restaurant_profiles
        SET review_status = 'draft',
            reviewed_at = NULL,
            reviewed_by = NULL,
            updated_at = now()
        WHERE restaurant_id = %s
        """,
        (restaurant_id,),
    )
    _enqueue_current_verified_preparation(cur, restaurant_id)
    cur.execute(
        """
        UPDATE demo_sites
        SET status = 'draft',
            published_at = NULL,
            published_by = NULL,
            updated_at = now()
        WHERE restaurant_id = %s AND status = 'published'
        """,
        (restaurant_id,),
    )
    cur.execute(
        """
        UPDATE email_campaigns
        SET status = 'draft',
            approved_at = NULL,
            approved_by = NULL,
            updated_at = now()
        WHERE restaurant_id = %s AND status = 'approved'
        """,
        (restaurant_id,),
    )


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
            SELECT r.id, r.name, r.email FROM restaurants r
            JOIN restaurant_profiles p ON p.restaurant_id = r.id
            WHERE p.apollo_lead->'source'->>'person_id' = %s
            LIMIT 1
            """,
            (person_id,),
        )
        row = cur.fetchone()
        if row:
            restaurant_id, existing_name, existing_email = row
            lock_restaurant_workflow(cur, restaurant_id)
            resolved_email = email or str(existing_email or "")
            identity_changed = (
                str(existing_name or "") != name
                or str(existing_email or "") != resolved_email
            )
            if identity_changed:
                cur.execute(
                    "UPDATE restaurants SET name = %s, email = %s, updated_at = now() WHERE id = %s",
                    (name, resolved_email, restaurant_id),
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
                    restaurant_id,
                    (location.get("city") or "").strip(),
                    (location.get("country") or "Australia").strip(),
                    jdump(lead),
                ),
            )
            if identity_changed:
                _invalidate_review_for_artifact_source_change(cur, restaurant_id)
            return restaurant_id

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

    lock_restaurant_workflow(cur, restaurant_id)

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


def _lock_active_ocr_claim(
    cur,
    restaurant_id: uuid.UUID,
    claim_id: uuid.UUID,
    claim_fingerprint: str,
) -> None:
    cur.execute(
        """
        SELECT 1
        FROM restaurant_profiles
        WHERE restaurant_id = %s
          AND ocr_status = 'running'
          AND ocr_claim_id = %s
          AND ocr_claim_fingerprint = %s
          AND ocr_input_fingerprint = %s
        FOR UPDATE
        """,
        (restaurant_id, claim_id, claim_fingerprint, claim_fingerprint),
    )
    if cur.fetchone() is None:
        raise StaleOCRClaim("OCR claim or input fingerprint is no longer current")


def apply_verified_record(
    cur,
    restaurant_id: uuid.UUID,
    record: dict,
    *,
    claim_id: uuid.UUID,
    claim_fingerprint: str,
) -> tuple[int, int]:
    """
    Write OCR-verified scrape JSON back to an existing restaurant row.
    Updates profile images/raw_public_data, menu items, and classified image tables.
    """
    lock_restaurant_workflow(cur, restaurant_id)
    _lock_active_ocr_claim(cur, restaurant_id, claim_id, claim_fingerprint)
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


def mark_ocr_status(
    cur,
    restaurant_id: uuid.UUID,
    status: str,
    *,
    claim_id: uuid.UUID,
    claim_fingerprint: str,
    errors: list[str] | None = None,
    summary: dict | None = None,
) -> None:
    if status not in ("verified", "no_images", "failed"):
        raise ValueError(f"Unsupported OCR status: {status}")
    summary_json = jdump(summary) if summary is not None else None
    lock_restaurant_workflow(cur, restaurant_id)
    cur.execute(
        """
        UPDATE restaurant_profiles
        SET
            ocr_status = %s,
            ocr_verified = (%s = 'verified'),
            ocr_verified_at = CASE WHEN %s = 'verified' THEN now() ELSE NULL END,
            ocr_started_at = CASE
                WHEN %s = 'running' THEN now()
                ELSE ocr_started_at
            END,
            ocr_completed_at = CASE
                WHEN %s IN ('verified', 'no_images', 'failed') THEN now()
                ELSE NULL
            END,
            ocr_claim_id = NULL,
            ocr_claim_fingerprint = NULL,
            ocr_verification_errors = %s::jsonb,
            raw_public_data = CASE
                WHEN %s::jsonb IS NULL THEN raw_public_data
                ELSE jsonb_set(
                    COALESCE(raw_public_data, '{}'::jsonb),
                    '{menu_ocr}',
                    %s::jsonb,
                    true
                )
            END,
            updated_at = now()
        WHERE restaurant_id = %s
          AND ocr_status = 'running'
          AND ocr_claim_id = %s
          AND ocr_claim_fingerprint = %s
          AND ocr_input_fingerprint = %s
        """,
        (
            status,
            status,
            status,
            status,
            status,
            jdump(errors or []),
            summary_json,
            summary_json,
            restaurant_id,
            claim_id,
            claim_fingerprint,
            claim_fingerprint,
        ),
    )
    if cur.rowcount != 1:
        raise StaleOCRClaim("OCR claim or input fingerprint is no longer current")
    sync_demo_ready_status(cur, restaurant_id)


def mark_ocr_verified(
    cur,
    restaurant_id: uuid.UUID,
    *,
    claim_id: uuid.UUID,
    claim_fingerprint: str,
    note: str = "",
) -> None:
    """Compatibility wrapper for older callers."""
    if note == "no_images":
        mark_ocr_status(
            cur,
            restaurant_id,
            "no_images",
            claim_id=claim_id,
            claim_fingerprint=claim_fingerprint,
            errors=[note],
        )
        return
    mark_ocr_status(
        cur,
        restaurant_id,
        "verified",
        claim_id=claim_id,
        claim_fingerprint=claim_fingerprint,
        errors=[note] if note else [],
    )


def release_ocr_claim(
    cur,
    restaurant_id: uuid.UUID,
    *,
    claim_id: uuid.UUID,
    claim_fingerprint: str,
    reason: str,
) -> None:
    """Return quota/timeout work to pending without consuming an attempt."""
    lock_restaurant_workflow(cur, restaurant_id)
    cur.execute(
        """
        UPDATE restaurant_profiles
        SET ocr_status = 'pending',
            ocr_verified = false,
            ocr_verified_at = NULL,
            ocr_started_at = NULL,
            ocr_completed_at = NULL,
            ocr_attempts = GREATEST(ocr_attempts - 1, 0),
            ocr_claim_id = NULL,
            ocr_claim_fingerprint = NULL,
            ocr_verification_errors = %s::jsonb,
            updated_at = now()
        WHERE restaurant_id = %s
          AND ocr_status = 'running'
          AND ocr_claim_id = %s
          AND ocr_claim_fingerprint = %s
          AND ocr_input_fingerprint = %s
        """,
        (
            jdump([str(reason or "OCR work deferred")[:500]]),
            restaurant_id,
            claim_id,
            claim_fingerprint,
            claim_fingerprint,
        ),
    )
    if cur.rowcount != 1:
        raise StaleOCRClaim("OCR claim or input fingerprint is no longer current")
    sync_demo_ready_status(cur, restaurant_id)


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


def sync_demo_ready_status(cur, restaurant_id: uuid.UUID) -> None:
    cur.execute(
        """
        WITH eligibility AS (
            SELECT r.id,
                   NULLIF(BTRIM(r.email), '') IS NOT NULL
                   AND EXISTS (
                     SELECT 1
                     FROM restaurant_profiles rp
                     WHERE rp.restaurant_id = r.id
                       AND rp.ocr_status = 'verified'
                   ) AS eligible
            FROM restaurants r
            WHERE r.id = %s
        )
        UPDATE restaurants r
        SET status = CASE WHEN eligibility.eligible THEN 'demo_ready' ELSE 'lead' END,
            updated_at = now()
        FROM eligibility
        WHERE r.id = eligibility.id
          AND r.status IN ('lead', 'demo_ready')
          AND r.status <> CASE WHEN eligibility.eligible THEN 'demo_ready' ELSE 'lead' END
        """,
        (restaurant_id,),
    )


def _enqueue_current_verified_preparation(
    cur,
    restaurant_id: uuid.UUID,
    expected_fingerprint: str | None = None,
) -> bool:
    cur.execute(
        """
        SELECT rp.ocr_input_fingerprint,
               lead_artifact_current_profile_fingerprint(rp.restaurant_id)
        FROM restaurant_profiles rp
        JOIN restaurants r ON r.id = rp.restaurant_id
        WHERE rp.restaurant_id = %s
          AND rp.ocr_status = 'verified'
          AND NULLIF(BTRIM(r.email), '') IS NOT NULL
          AND (%s IS NULL OR rp.ocr_input_fingerprint = %s)
        """,
        (restaurant_id, expected_fingerprint, expected_fingerprint),
    )
    row = cur.fetchone()
    if row is None:
        return False
    fingerprint = str(row[0] or "legacy")
    profile_fingerprint = str(row[1] or "").strip()
    if not profile_fingerprint:
        raise RuntimeError("verified lead has no current artifact profile fingerprint")
    idempotency_key = (
        f"lead.prepare:{restaurant_id}:{fingerprint}:{profile_fingerprint}"
    )
    cur.execute(
        """
        INSERT INTO job_runs (
            job_type, status, payload, idempotency_key, max_attempts
        ) VALUES (
            'lead.prepare', 'queued',
            jsonb_build_object('restaurant_id', %s::text),
            %s,
            3
        )
        ON CONFLICT (idempotency_key) WHERE idempotency_key IS NOT NULL
        DO UPDATE SET
            status = 'queued',
            payload = EXCLUDED.payload,
            attempts = 0,
            max_attempts = EXCLUDED.max_attempts,
            last_error = NULL,
            available_at = now(),
            locked_at = NULL,
            locked_by = NULL,
            lease_expires_at = NULL,
            updated_at = now()
        WHERE job_runs.job_type = 'lead.prepare'
          AND job_runs.status IN ('completed', 'failed')
        """,
        (str(restaurant_id), idempotency_key),
    )
    return True


def enqueue_post_ocr_preparation(
    cur,
    restaurant_id: uuid.UUID,
    expected_fingerprint: str,
) -> None:
    if not _enqueue_current_verified_preparation(
        cur,
        restaurant_id,
        expected_fingerprint,
    ):
        raise StaleOCRClaim("verified OCR input changed before draft preparation was queued")


def fetch_unverified_leads(
    cur,
    limit: int,
    *,
    claim: bool = True,
    max_attempts: int = 3,
    retry_after_hours: int = 24,
) -> list[tuple[uuid.UUID, dict, str, uuid.UUID | None, str]]:
    max_attempts = max(1, int(max_attempts))
    retry_after_hours = max(1, int(retry_after_hours))
    if not claim:
        cur.execute(
            """
            SELECT rp.restaurant_id, rp.raw_public_data, r.name,
                   NULL::uuid, rp.ocr_input_fingerprint
            FROM restaurant_profiles rp
            JOIN restaurants r ON r.id = rp.restaurant_id
            WHERE (
                rp.ocr_status = 'pending'
                OR (
                    rp.ocr_status = 'failed'
                    AND rp.ocr_attempts < %s
                    AND rp.ocr_completed_at <= now() - (%s * interval '1 hour')
                )
            )
              AND rp.raw_public_data IS NOT NULL
              AND rp.raw_public_data <> '{}'::jsonb
              AND NULLIF(BTRIM(r.email), '') IS NOT NULL
            ORDER BY
              CASE
                WHEN EXISTS (
                  SELECT 1 FROM demo_sites demo
                  WHERE demo.restaurant_id = rp.restaurant_id
                    AND demo.status = 'published'
                ) THEN 0
                WHEN EXISTS (
                  SELECT 1 FROM demo_sites demo
                  WHERE demo.restaurant_id = rp.restaurant_id
                ) THEN 1
                ELSE 2
              END,
              rp.created_at ASC
            LIMIT %s
            """,
            (max_attempts, retry_after_hours, limit),
        )
        rows_data = cur.fetchall()
    else:
        # A repeatedly crashed worker must not reclaim the same unchanged input
        # forever. Convert an exhausted stale claim into the explicit failed
        # remediation state before selecting new work.
        cur.execute(
            """
            UPDATE restaurant_profiles
            SET ocr_status = 'failed',
                ocr_verified = false,
                ocr_verified_at = NULL,
                ocr_completed_at = now(),
                ocr_claim_id = NULL,
                ocr_claim_fingerprint = NULL,
                ocr_verification_errors = COALESCE(ocr_verification_errors, '[]'::jsonb)
                  || '["stale OCR claim exhausted automatic attempt limit"]'::jsonb,
                updated_at = now()
            WHERE ocr_status = 'running'
              AND ocr_started_at < now() - interval '2 hours'
              AND ocr_attempts >= %s
            """,
            (max_attempts,),
        )
        cur.execute(
            """
            WITH candidates AS (
                SELECT rp.restaurant_id, rp.raw_public_data, r.name
                FROM restaurant_profiles rp
                JOIN restaurants r ON r.id = rp.restaurant_id
                WHERE (
                    rp.ocr_status = 'pending'
                    OR (
                        rp.ocr_status = 'running'
                        AND rp.ocr_started_at < now() - interval '2 hours'
                        AND rp.ocr_attempts < %s
                    )
                    OR (
                        rp.ocr_status = 'failed'
                        AND rp.ocr_attempts < %s
                        AND rp.ocr_completed_at <= now() - (%s * interval '1 hour')
                    )
                )
                  AND rp.raw_public_data IS NOT NULL
                  AND rp.raw_public_data <> '{}'::jsonb
                  AND NULLIF(BTRIM(r.email), '') IS NOT NULL
                ORDER BY
                  CASE
                    WHEN EXISTS (
                      SELECT 1 FROM demo_sites demo
                      WHERE demo.restaurant_id = rp.restaurant_id
                        AND demo.status = 'published'
                    ) THEN 0
                    WHEN EXISTS (
                      SELECT 1 FROM demo_sites demo
                      WHERE demo.restaurant_id = rp.restaurant_id
                    ) THEN 1
                    ELSE 2
                  END,
                  rp.created_at ASC
                LIMIT %s
                FOR UPDATE OF rp SKIP LOCKED
            )
            UPDATE restaurant_profiles rp
            SET
                ocr_status = 'running',
                ocr_verified = false,
                ocr_verified_at = NULL,
                ocr_started_at = now(),
                ocr_completed_at = NULL,
                ocr_attempts = rp.ocr_attempts + 1,
                ocr_claim_id = gen_random_uuid(),
                ocr_claim_fingerprint = rp.ocr_input_fingerprint,
                ocr_verification_errors = '[]'::jsonb,
                updated_at = now()
            FROM candidates c
            WHERE rp.restaurant_id = c.restaurant_id
            RETURNING rp.restaurant_id, c.raw_public_data, c.name,
                      rp.ocr_claim_id, rp.ocr_claim_fingerprint
            """,
            (max_attempts, max_attempts, retry_after_hours, limit),
        )
        rows_data = cur.fetchall()

    rows = []
    for restaurant_id, raw_public_data, name, claim_id, claim_fingerprint in rows_data:
        if isinstance(raw_public_data, dict):
            record = raw_public_data
        elif isinstance(raw_public_data, str):
            record = json.loads(raw_public_data)
        else:
            record = json.loads(json.dumps(raw_public_data))
        rows.append((restaurant_id, record, name or "", claim_id, claim_fingerprint or ""))
    return rows


def upsert_restaurant_record(cur, record: dict, source_file: str) -> tuple[uuid.UUID, bool]:
    """Returns (restaurant_id, skipped_duplicate)."""
    record = sanitize_sensitive_urls(record)
    try:
        from menu_image_ocr import _sanitize_menu_item_images
        _sanitize_menu_item_images(record)
    except ImportError:
        pass

    google = record.get("google") or {}
    place_id = (google.get("place_id") or "").strip()
    incoming_name = (record.get("name") or "").strip()
    incoming_email = ((record.get("contact") or {}).get("email") or "").strip()
    identity_changed = False

    if place_id:
        cur.execute(
            """
            SELECT r.id, r.name, r.email
            FROM restaurant_profiles rp
            JOIN restaurants r ON r.id = rp.restaurant_id
            WHERE rp.google_place_id = %s
            LIMIT 1
            """,
            (place_id,),
        )
        row = cur.fetchone()
        if row:
            restaurant_id, existing_name, existing_email = row
            lock_restaurant_workflow(cur, restaurant_id)
            resolved_email = incoming_email or str(existing_email or "")
            identity_changed = (
                str(existing_name or "") != incoming_name
                or str(existing_email or "") != resolved_email
            )
            if identity_changed:
                cur.execute(
                    """
                    UPDATE restaurants
                    SET name = %s,
                        email = %s,
                        updated_at = now()
                    WHERE id = %s
                    """,
                    (incoming_name, resolved_email, restaurant_id),
                )
        else:
            cur.execute(
                "INSERT INTO restaurants (name, email, status) VALUES (%s, %s, 'lead') RETURNING id",
                (incoming_name, incoming_email),
            )
            restaurant_id = cur.fetchone()[0]
    else:
        cur.execute(
            "INSERT INTO restaurants (name, email, status) VALUES (%s, %s, 'lead') RETURNING id",
            (incoming_name, incoming_email),
        )
        restaurant_id = cur.fetchone()[0]

    lock_restaurant_workflow(cur, restaurant_id)

    contact = record.get("contact") or {}
    location = record.get("location") or {}
    coords = location.get("coordinates") or {}
    input_fingerprint = ocr_input_fingerprint(record)

    # Serialize scrape replacement against OCR finalization and remember
    # whether the human-approved image input is being replaced.
    cur.execute(
        """
        SELECT ocr_input_fingerprint,
               ocr_status,
               raw_public_data,
               COALESCE(lead_artifact_current_profile_fingerprint(restaurant_id), '')
        FROM restaurant_profiles
        WHERE restaurant_id = %s
        FOR UPDATE
        """,
        (restaurant_id,),
    )
    existing_profile = cur.fetchone()
    existing_profile_fingerprint = (
        str(existing_profile[3] or "") if existing_profile is not None else ""
    )
    input_changed = (
        existing_profile is not None
        and str(existing_profile[0] or "") != input_fingerprint
    )
    preserve_verified_ocr = (
        existing_profile is not None
        and str(existing_profile[1] or "") == "verified"
        and not input_changed
    )
    if preserve_verified_ocr:
        existing_raw = existing_profile[2]
        if isinstance(existing_raw, str):
            try:
                existing_raw = json.loads(existing_raw)
            except json.JSONDecodeError:
                existing_raw = {}
        if isinstance(existing_raw, dict):
            # A new Places/Apollo pass may refresh contact and review fields,
            # but it must not erase the OCR-derived menu/classification output
            # while retaining verified eligibility for the same image input.
            for key in ("images", "menu_items", "menu_ocr"):
                if key in existing_raw:
                    record[key] = existing_raw[key]

    raw_record = json.dumps(record)

    cur.execute(
        """
        INSERT INTO restaurant_profiles (
            restaurant_id, opening_hours, phone, website, address, city, state, country,
            latitude, longitude, google_place_id, google_data_id, rating, reviews_count,
            price_level, cuisines, owners, images, apollo_lead, scrape_status, scrape_errors,
            dietary_options, raw_public_data, ocr_input_fingerprint
        ) VALUES (
            %s, %s::jsonb, %s, %s, %s, %s, %s, %s,
            %s, %s, %s, %s, %s, %s,
            %s, %s::jsonb, %s::jsonb, %s::jsonb, %s::jsonb, %s, %s::jsonb,
            %s::jsonb, %s::jsonb, %s
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
            owners = CASE
                WHEN EXCLUDED.owners <> '[]'::jsonb THEN EXCLUDED.owners
                ELSE restaurant_profiles.owners
            END,
            images = EXCLUDED.images,
            apollo_lead = CASE
                WHEN EXCLUDED.apollo_lead <> '{}'::jsonb THEN EXCLUDED.apollo_lead
                ELSE restaurant_profiles.apollo_lead
            END,
            scrape_status = EXCLUDED.scrape_status,
            scrape_errors = EXCLUDED.scrape_errors,
            dietary_options = EXCLUDED.dietary_options,
            raw_public_data = EXCLUDED.raw_public_data,
            ocr_status = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN 'pending'
                WHEN restaurant_profiles.ocr_status = 'running'
                    AND EXCLUDED.raw_public_data IS DISTINCT FROM restaurant_profiles.raw_public_data
                    THEN 'pending'
                ELSE restaurant_profiles.ocr_status
            END,
            ocr_verified = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN false
                WHEN restaurant_profiles.ocr_status = 'running'
                    AND EXCLUDED.raw_public_data IS DISTINCT FROM restaurant_profiles.raw_public_data
                    THEN false
                ELSE restaurant_profiles.ocr_verified
            END,
            ocr_verified_at = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN NULL
                WHEN restaurant_profiles.ocr_status = 'running'
                    AND EXCLUDED.raw_public_data IS DISTINCT FROM restaurant_profiles.raw_public_data
                    THEN NULL
                ELSE restaurant_profiles.ocr_verified_at
            END,
            ocr_started_at = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN NULL
                WHEN restaurant_profiles.ocr_status = 'running'
                    AND EXCLUDED.raw_public_data IS DISTINCT FROM restaurant_profiles.raw_public_data
                    THEN NULL
                ELSE restaurant_profiles.ocr_started_at
            END,
            ocr_completed_at = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN NULL
                WHEN restaurant_profiles.ocr_status = 'running'
                    AND EXCLUDED.raw_public_data IS DISTINCT FROM restaurant_profiles.raw_public_data
                    THEN NULL
                ELSE restaurant_profiles.ocr_completed_at
            END,
            ocr_attempts = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN 0
                WHEN restaurant_profiles.ocr_status = 'running'
                    AND EXCLUDED.raw_public_data IS DISTINCT FROM restaurant_profiles.raw_public_data
                    THEN 0
                ELSE restaurant_profiles.ocr_attempts
            END,
            ocr_verification_errors = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN '[]'::jsonb
                WHEN restaurant_profiles.ocr_status = 'running'
                    AND EXCLUDED.raw_public_data IS DISTINCT FROM restaurant_profiles.raw_public_data
                    THEN '[]'::jsonb
                ELSE restaurant_profiles.ocr_verification_errors
            END,
            ocr_claim_id = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN NULL
                WHEN restaurant_profiles.ocr_status = 'running'
                    AND EXCLUDED.raw_public_data IS DISTINCT FROM restaurant_profiles.raw_public_data
                    THEN NULL
                ELSE restaurant_profiles.ocr_claim_id
            END,
            ocr_claim_fingerprint = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN NULL
                WHEN restaurant_profiles.ocr_status = 'running'
                    AND EXCLUDED.raw_public_data IS DISTINCT FROM restaurant_profiles.raw_public_data
                    THEN NULL
                ELSE restaurant_profiles.ocr_claim_fingerprint
            END,
            review_status = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN 'draft'
                ELSE restaurant_profiles.review_status
            END,
            reviewed_at = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN NULL
                ELSE restaurant_profiles.reviewed_at
            END,
            reviewed_by = CASE
                WHEN EXCLUDED.ocr_input_fingerprint IS DISTINCT FROM restaurant_profiles.ocr_input_fingerprint
                    THEN NULL
                ELSE restaurant_profiles.reviewed_by
            END,
            ocr_input_fingerprint = EXCLUDED.ocr_input_fingerprint,
            updated_at = CASE
                WHEN EXCLUDED.raw_public_data IS DISTINCT FROM restaurant_profiles.raw_public_data
                    THEN now()
                ELSE restaurant_profiles.updated_at
            END
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
            input_fingerprint,
        ),
    )

    if input_changed:
        cur.execute(
            """
            UPDATE demo_sites
            SET status = 'draft',
                published_at = NULL,
                published_by = NULL,
                updated_at = now()
            WHERE restaurant_id = %s
              AND status = 'published'
            """,
            (restaurant_id,),
        )
        cur.execute(
            """
            UPDATE email_campaigns
            SET status = 'draft',
                approved_at = NULL,
                approved_by = NULL,
                updated_at = now()
            WHERE restaurant_id = %s
              AND status = 'approved'
            """,
            (restaurant_id,),
        )

    sync_menu_items_and_reviews(cur, restaurant_id, record, sync_reviews=True)

    cur.execute(
        "SELECT COALESCE(lead_artifact_current_profile_fingerprint(%s), '')",
        (restaurant_id,),
    )
    current_profile_fingerprint = str(cur.fetchone()[0] or "")
    if existing_profile is not None and (
        identity_changed
        or existing_profile_fingerprint != current_profile_fingerprint
    ):
        _invalidate_review_for_artifact_source_change(cur, restaurant_id)

    sync_demo_ready_status(cur, restaurant_id)

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
            verify_ocr_schema(cur)
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
