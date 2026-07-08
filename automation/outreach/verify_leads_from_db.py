#!/usr/bin/env python3
"""
Nightly lead OCR verification — process restaurant_profiles where ocr_verified=false.

Requires migration 000013_lead_ocr_verified and MENU OCR API keys in backend/.env.
"""

from __future__ import annotations

import argparse
import logging
import os
import sys
import uuid

from env_loader import load_project_env

load_project_env()

from import_to_db import (  # noqa: E402
    append_ocr_verification_error,
    apply_verified_record,
    fetch_unverified_leads,
    get_conn,
    load_env,
    mark_ocr_verified,
    verify_ocr_schema,
)
from menu_image_ocr import (  # noqa: E402
    MenuOCRConfig,
    collect_candidate_image_urls,
    enrich_restaurant_with_menu_ocr,
)

log = logging.getLogger("verify_leads_from_db")


def env_enabled() -> bool:
    return os.getenv("LEAD_OCR_VERIFICATION_ENABLED", "false").lower() in ("1", "true", "yes", "on")


def default_batch_size() -> int:
    raw = os.getenv("LEAD_OCR_BATCH_SIZE", "50").strip()
    try:
        value = int(raw)
    except ValueError:
        value = 50
    return max(1, value)


def verify_one(
    cur,
    restaurant_id: uuid.UUID,
    record: dict,
    name: str,
    cfg: MenuOCRConfig,
    *,
    dry_run: bool,
) -> str:
    candidates = collect_candidate_image_urls(record)
    if not candidates:
        log.warning("  %s — no image candidates; marking verified (no_images)", name)
        if not dry_run:
            mark_ocr_verified(cur, restaurant_id, note="no_images")
        return "no_images"

    if not cfg.enabled:
        raise RuntimeError("Menu OCR is disabled — set HUGGING_FACE_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY")

    log.info("  %s — OCR on %d image(s)", name, len(candidates))
    if dry_run:
        return "dry_run"

    enrich_restaurant_with_menu_ocr(record, cfg)
    apply_verified_record(cur, restaurant_id, record)
    mark_ocr_verified(cur, restaurant_id)
    return "verified"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Verify unverified restaurant leads via menu OCR")
    parser.add_argument("--limit", type=int, default=0, help="Max restaurants to process (default: LEAD_OCR_BATCH_SIZE)")
    parser.add_argument("--dry-run", action="store_true", help="Log actions without writing to DB")
    parser.add_argument("--force", action="store_true", help="Run even when LEAD_OCR_VERIFICATION_ENABLED is false")
    parser.add_argument("-v", "--verbose", action="store_true")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s  [%(levelname)s]  %(message)s",
    )

    if not args.force and not env_enabled():
        log.info("LEAD_OCR_VERIFICATION_ENABLED is not true — skipping")
        return 0

    limit = args.limit if args.limit > 0 else default_batch_size()
    database_url = load_env()
    cfg = MenuOCRConfig()

    log.info("Connecting to database (limit=%d, dry_run=%s)", limit, args.dry_run)
    conn = get_conn(database_url)

    verified = 0
    skipped = 0
    failed = 0

    try:
        with conn.cursor() as cur:
            verify_ocr_schema(cur)
            leads = fetch_unverified_leads(cur, limit)
            if not leads:
                log.info("No unverified leads with raw_public_data")
                return 0

            log.info("Found %d unverified lead(s)", len(leads))

        for restaurant_id, record, name in leads:
            label = name or str(restaurant_id)
            try:
                with conn.cursor() as cur:
                    status = verify_one(cur, restaurant_id, record, label, cfg, dry_run=args.dry_run)
                    if not args.dry_run:
                        conn.commit()
                    else:
                        conn.rollback()
                if status in ("verified", "no_images", "dry_run"):
                    verified += 1
                    log.info("  ✓ %s → %s", label, status)
                else:
                    skipped += 1
            except Exception as exc:
                failed += 1
                log.error("  ✗ %s — %s", label, exc)
                if not args.dry_run:
                    conn.rollback()
                    try:
                        with conn.cursor() as cur:
                            append_ocr_verification_error(cur, restaurant_id, str(exc))
                        conn.commit()
                    except Exception:
                        conn.rollback()

        log.info("Done: verified=%d failed=%d", verified, failed)
        return 1 if failed else 0
    finally:
        conn.close()


if __name__ == "__main__":
    raise SystemExit(main())
