#!/usr/bin/env python3
"""Nightly lead OCR verification with explicit, durable state transitions."""

from __future__ import annotations

import argparse
import json
import logging
import os
import sys
import uuid

from env_loader import load_project_env

load_project_env()

from import_to_db import (  # noqa: E402
    apply_verified_record,
    enqueue_post_ocr_preparation,
    fetch_unverified_leads,
    get_conn,
    load_env,
    mark_ocr_status,
    StaleOCRClaim,
    verify_ocr_schema,
)
from google_places_photo import (  # noqa: E402
    fetch_fresh_google_photo_resources,
    resolve_google_photo_uri,
)
from menu_image_ocr import (  # noqa: E402
    MenuOCRConfig,
    collect_candidate_image_urls,
    enrich_restaurant_with_menu_ocr,
    is_trusted_automated_image_url,
)
from tuvi_outreach_agent import Config  # noqa: E402

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


def ocr_max_attempts() -> int:
    try:
        return max(1, int(os.getenv("LEAD_OCR_MAX_ATTEMPTS", "3").strip()))
    except ValueError:
        return 3


def ocr_retry_after_hours() -> int:
    try:
        return max(1, int(os.getenv("LEAD_OCR_RETRY_AFTER_HOURS", "24").strip()))
    except ValueError:
        return 24


def verify_one(
    cur,
    restaurant_id: uuid.UUID,
    record: dict,
    name: str,
    cfg: MenuOCRConfig,
    places_cfg: Config,
    *,
    claim_id: uuid.UUID | None,
    claim_fingerprint: str,
    dry_run: bool,
) -> str:
    if not dry_run and claim_id is None:
        raise StaleOCRClaim("OCR worker did not receive a durable claim token")

    collected_direct_urls = collect_candidate_image_urls(record)
    direct_urls = [
        url for url in collected_direct_urls if is_trusted_automated_image_url(url)
    ]
    ignored_direct_count = len(collected_direct_urls) - len(direct_urls)
    if ignored_direct_count:
        log.warning(
            "  %s — ignored %d untrusted direct image URL(s); unattended OCR uses Places/Google-hosted images only",
            name,
            ignored_direct_count,
        )
    place_id = str(((record.get("google") or {}).get("place_id") or "")).strip()
    images_block = record.setdefault("images", {})
    if isinstance(images_block, dict):
        # Older lead JSON may contain cached, expirable resource names. Never
        # carry them into newly persisted OCR output.
        images_block["google_photos"] = []
    if dry_run:
        if direct_urls or place_id:
            log.info(
                "  %s — would refresh Places photos and OCR up to %d image(s)",
                name,
                cfg.max_images,
            )
            return "dry_run"

    google_photos: list[dict] = []
    photo_refresh_error = ""
    if place_id:
        try:
            google_photos = fetch_fresh_google_photo_resources(place_id, places_cfg)
            if isinstance(images_block, dict):
                images_block["google_photo_count"] = len(google_photos)
        except Exception as exc:
            photo_refresh_error = str(exc)
            if not direct_urls:
                # A provider/configuration failure is not evidence that the
                # restaurant has no images. Leave it retryable as failed.
                raise RuntimeError(
                    "Google Places photos could not be refreshed for OCR"
                ) from exc
            log.warning(
                "  %s — Places photo refresh failed; using trusted direct image fallback: %s",
                name,
                exc,
            )

    if not direct_urls and not google_photos:
        log.warning("  %s — no image candidates; marking no_images (not verified)", name)
        if not dry_run:
            mark_ocr_status(
                cur,
                restaurant_id,
                "no_images",
                claim_id=claim_id,
                claim_fingerprint=claim_fingerprint,
                errors=["no_images"],
            )
        return "no_images"

    if not cfg.enabled:
        raise RuntimeError("Menu OCR is disabled — set HUGGING_FACE_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY")

    analysis_candidates: list[dict] = []
    resolution_errors: list[str] = []
    if photo_refresh_error:
        resolution_errors.append(f"photo resource refresh: {photo_refresh_error}")
    transient_photo_uris: list[str] = []
    transient_photo_names: list[str] = []
    # Reserve one analysis slot for a trusted direct URL when available. This
    # gives OCR a real fallback even when every successfully resolved Places
    # photo later fails during download or model analysis.
    places_limit = cfg.max_images
    if direct_urls and cfg.max_images > 1:
        places_limit -= 1
    for photo in google_photos[:places_limit]:
        try:
            photo_uri = resolve_google_photo_uri(photo["name"], places_cfg)
        except Exception as exc:
            resolution_errors.append(f"{photo['name']}: {exc}")
            continue
        analysis_candidates.append({
            "analysis_url": photo_uri,
            # Deliberately empty: Photo Media URIs are short-lived and must not
            # be persisted to raw_public_data, menu_items, or image tables.
            "persistent_url": "",
            "source": "google_places_photo",
            # The resource name can expire and is used only in memory for this
            # analysis attempt. Persist the stable Place ID as provenance.
            "source_ref": "",
            "google_place_id": place_id,
        })
        transient_photo_uris.append(photo_uri)
        transient_photo_names.append(photo["name"])

    remaining = max(0, cfg.max_images - len(analysis_candidates))
    analysis_candidates.extend(
        {
            "analysis_url": url,
            "persistent_url": url,
            "source": "public_url",
        }
        for url in direct_urls[:remaining]
    )

    if not analysis_candidates:
        raise RuntimeError("Google Places photo resources could not be resolved for OCR")

    log.info("  %s — OCR on %d image(s)", name, len(analysis_candidates))

    enrich_restaurant_with_menu_ocr(
        record,
        cfg,
        analysis_candidates=analysis_candidates,
    )
    ocr_summary = record.get("menu_ocr") or {}
    if (
        int(ocr_summary.get("images_succeeded") or 0) < 1
        and cfg.max_images == 1
        and google_photos
        and direct_urls
        and analysis_candidates
        and analysis_candidates[0].get("source") == "google_places_photo"
    ):
        # With a one-image limit, keep Places first and spend a fallback
        # attempt only when that sole resolved Places image could not be
        # analyzed. The direct URL has already passed the unattended
        # Google-host allowlist above.
        enrich_restaurant_with_menu_ocr(
            record,
            cfg,
            analysis_candidates=[{
                "analysis_url": direct_urls[0],
                "persistent_url": direct_urls[0],
                "source": "public_url",
            }],
        )
        ocr_summary = record.get("menu_ocr") or {}
    if resolution_errors:
        ocr_summary["photo_resolution_errors"] = resolution_errors
        record["menu_ocr"] = ocr_summary
    if int(ocr_summary.get("images_succeeded") or 0) < 1:
        raise RuntimeError("No image could be analyzed successfully")
    serialized_record = json.dumps(record, ensure_ascii=False)
    if any(uri in serialized_record for uri in transient_photo_uris):
        raise RuntimeError("Transient Google Places photo URI reached persistent OCR output")
    if any(name in serialized_record for name in transient_photo_names):
        raise RuntimeError("Expirable Google Places photo name reached persistent OCR output")

    apply_verified_record(
        cur,
        restaurant_id,
        record,
        claim_id=claim_id,
        claim_fingerprint=claim_fingerprint,
    )
    mark_ocr_status(
        cur,
        restaurant_id,
        "verified",
        claim_id=claim_id,
        claim_fingerprint=claim_fingerprint,
    )
    enqueue_post_ocr_preparation(cur, restaurant_id, claim_fingerprint)
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
    places_cfg = Config()

    if not args.dry_run and not cfg.enabled:
        log.error("Menu OCR is disabled — configure a supported vision provider before claiming leads")
        return 1

    log.info("Connecting to database (limit=%d, dry_run=%s)", limit, args.dry_run)
    conn = get_conn(database_url)

    verified = 0
    skipped = 0
    failed = 0

    try:
        with conn.cursor() as cur:
            verify_ocr_schema(cur)
            leads = fetch_unverified_leads(
                cur,
                limit,
                claim=not args.dry_run,
                max_attempts=ocr_max_attempts(),
                retry_after_hours=ocr_retry_after_hours(),
            )
            if not args.dry_run:
                # Persist running claims before any remote OCR/photo requests.
                conn.commit()
            if not leads:
                log.info("No pending leads with raw_public_data")
                return 0

            log.info("Found %d unverified lead(s)", len(leads))

        for restaurant_id, record, name, claim_id, claim_fingerprint in leads:
            label = name or str(restaurant_id)
            try:
                with conn.cursor() as cur:
                    status = verify_one(
                        cur,
                        restaurant_id,
                        record,
                        label,
                        cfg,
                        places_cfg,
                        claim_id=claim_id,
                        claim_fingerprint=claim_fingerprint,
                        dry_run=args.dry_run,
                    )
                    if not args.dry_run:
                        conn.commit()
                    else:
                        conn.rollback()
                if status in ("verified", "dry_run"):
                    verified += 1
                    log.info("  ✓ %s → %s", label, status)
                elif status == "no_images":
                    skipped += 1
                    log.info("  - %s → no_images (not email eligible)", label)
                else:
                    skipped += 1
            except StaleOCRClaim as exc:
                skipped += 1
                conn.rollback()
                log.warning("  - %s — stale OCR result discarded: %s", label, exc)
            except Exception as exc:
                failed += 1
                log.error("  ✗ %s — %s", label, exc)
                if not args.dry_run:
                    conn.rollback()
                    try:
                        with conn.cursor() as cur:
                            mark_ocr_status(
                                cur,
                                restaurant_id,
                                "failed",
                                claim_id=claim_id,
                                claim_fingerprint=claim_fingerprint,
                                errors=[str(exc)[:500]],
                            )
                        conn.commit()
                    except Exception:
                        conn.rollback()

        log.info("Done: verified=%d failed=%d", verified, failed)
        return 1 if failed else 0
    finally:
        conn.close()


if __name__ == "__main__":
    raise SystemExit(main())
