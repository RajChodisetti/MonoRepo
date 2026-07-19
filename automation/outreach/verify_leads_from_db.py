#!/usr/bin/env python3
"""Nightly lead OCR verification with explicit, durable state transitions."""

from __future__ import annotations

import argparse
import json
import logging
import os
import sys
import time
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
    release_ocr_claim,
    StaleOCRClaim,
    verify_ocr_schema,
)
from google_places_photo import (  # noqa: E402
    PhotoRequestTransientError,
    fetch_fresh_google_photo_resources,
    resolve_google_photo_uri,
)
from menu_image_ocr import (  # noqa: E402
    MenuOCRConfig,
    MenuImageAnalyzer,
    OCRTransientError,
    all_scraped_photos_processed,
    collect_candidate_image_urls,
    enrich_restaurant_with_menu_ocr,
    is_trusted_automated_image_url,
)
from ocr_request_budget import (  # noqa: E402
    DurableOCRRequestBudget,
    OCRDailyBudgetExhausted,
)
from tuvi_outreach_agent import Config  # noqa: E402

log = logging.getLogger("verify_leads_from_db")


class OCRIncompleteError(RuntimeError):
    """The attempt ran, but not every discovered photo completed successfully."""

    def __init__(self, message: str, summary: dict):
        super().__init__(message)
        self.summary = summary


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


def ocr_daily_request_limit() -> int:
    try:
        configured = int(os.getenv("LEAD_OCR_DAILY_REQUEST_LIMIT", "200").strip())
    except ValueError:
        configured = 200
    return min(200, max(1, configured))


def ocr_worker_poll_seconds() -> int:
    try:
        configured = int(os.getenv("OCR_WORKER_POLL_SECONDS", "900").strip())
    except ValueError:
        configured = 900
    return max(60, configured)


def verify_one(
    cur,
    restaurant_id: uuid.UUID,
    record: dict,
    name: str,
    cfg: MenuOCRConfig,
    places_cfg: Config,
    analyzer: MenuImageAnalyzer,
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
                "  %s — would refresh Places photos and OCR every discovered image",
                name,
            )
            return "dry_run"

    google_photos: list[dict] = []
    photo_refresh_error = ""
    if place_id:
        try:
            google_photos = fetch_fresh_google_photo_resources(place_id, places_cfg)
            if isinstance(images_block, dict):
                images_block["google_photo_count"] = len(google_photos)
        except PhotoRequestTransientError as exc:
            raise OCRTransientError(
                "Google Places photo refresh timed out or was temporarily unavailable"
            ) from exc
        except Exception as exc:
            photo_refresh_error = str(exc)
            if not direct_urls:
                # A provider/configuration failure is not evidence that the
                # restaurant has no images. Leave it retryable as failed.
                raise OCRIncompleteError(
                    "Google Places photos could not be refreshed for OCR",
                    {
                        "provider": cfg.primary_provider,
                        "model": (
                            cfg.hf_vision_model
                            if cfg.primary_provider == "huggingface"
                            else cfg.openai_model
                            if cfg.primary_provider == "openai"
                            else cfg.gemini_model
                        ),
                        "images_discovered": 0,
                        "images_resolved": 0,
                        "images_analyzed": 0,
                        "images_succeeded": 0,
                        "images_failed": 0,
                        "all_images_processed": False,
                        "photo_resolution_error_count": 1,
                    },
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
                summary={
                    "images_discovered": 0,
                    "images_resolved": 0,
                    "images_analyzed": 0,
                    "images_succeeded": 0,
                    "images_failed": 0,
                    "all_images_processed": False,
                },
            )
        return "no_images"

    if not cfg.enabled:
        raise RuntimeError("Menu OCR is disabled — set HUGGING_FACE_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY")

    analysis_candidates: list[dict] = []
    resolution_errors: list[str] = []
    if photo_refresh_error:
        resolution_errors.append("Google Places photo resource refresh failed")
    transient_photo_uris: list[str] = []
    transient_photo_names: list[str] = []
    for photo_index, photo in enumerate(google_photos):
        try:
            photo_uri = resolve_google_photo_uri(photo["name"], places_cfg)
        except PhotoRequestTransientError as exc:
            raise OCRTransientError(
                "Google Places photo resolution timed out or was temporarily unavailable"
            ) from exc
        except Exception:
            resolution_errors.append(
                f"Google Places photo {photo_index + 1} could not be resolved"
            )
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
            "source_index": photo_index,
        })
        transient_photo_uris.append(photo_uri)
        transient_photo_names.append(photo["name"])

    analysis_candidates.extend(
        {
            "analysis_url": url,
            "persistent_url": url,
            "source": "public_url",
        }
        for url in direct_urls
    )

    images_discovered = len(google_photos) + len(direct_urls)
    if not analysis_candidates:
        raise OCRIncompleteError(
            "No scraped photo could be resolved for OCR",
            {
                "provider": cfg.primary_provider,
                "model": (
                    cfg.hf_vision_model
                    if cfg.primary_provider == "huggingface"
                    else cfg.openai_model
                    if cfg.primary_provider == "openai"
                    else cfg.gemini_model
                ),
                "images_discovered": images_discovered,
                "images_resolved": 0,
                "images_analyzed": 0,
                "images_succeeded": 0,
                "images_failed": 0,
                "all_images_processed": False,
                "photo_resolution_error_count": len(resolution_errors),
            },
        )

    log.info(
        "  %s — OCR on all %d resolved image(s) from %d discovered photo(s)",
        name,
        len(analysis_candidates),
        images_discovered,
    )

    enrich_restaurant_with_menu_ocr(
        record,
        cfg,
        analyzer=analyzer,
        analysis_candidates=analysis_candidates,
        process_all_candidates=True,
    )
    ocr_summary = record.get("menu_ocr") or {}
    ocr_summary["images_discovered"] = images_discovered
    ocr_summary["images_resolved"] = len(analysis_candidates)
    ocr_summary["photo_resolution_error_count"] = len(resolution_errors)
    ocr_summary["all_images_processed"] = all_scraped_photos_processed(
        ocr_summary,
        images_discovered,
        len(resolution_errors),
    )
    if resolution_errors:
        ocr_summary["photo_resolution_errors"] = resolution_errors
    record["menu_ocr"] = ocr_summary
    if not ocr_summary["all_images_processed"]:
        succeeded = int(ocr_summary.get("images_succeeded") or 0)
        raise OCRIncompleteError(
            f"OCR completed successfully for {succeeded} of {images_discovered} scraped photos",
            ocr_summary,
        )
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
    parser.add_argument("--watch", action="store_true", help="Run continuously as the background OCR worker")
    parser.add_argument("--poll-seconds", type=int, default=0, help="Background poll interval (default: OCR_WORKER_POLL_SECONDS)")
    parser.add_argument("-v", "--verbose", action="store_true")
    return parser.parse_args()


def run_once(args: argparse.Namespace) -> int:
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
    budget_conn = None
    budget = None
    try:
        if not args.dry_run:
            budget_conn = get_conn(database_url)
            budget = DurableOCRRequestBudget(
                budget_conn,
                daily_limit=ocr_daily_request_limit(),
            )
            snapshot = budget.snapshot()
            if snapshot.remaining == 0:
                log.info(
                    "OCR daily request budget exhausted (%d/%d); waiting for the UTC reset",
                    snapshot.requests_used,
                    snapshot.daily_limit,
                )
                conn.close()
                budget_conn.close()
                return 0
        analyzer = MenuImageAnalyzer(cfg, request_budget=budget)
    except Exception:
        conn.close()
        if budget_conn is not None:
            budget_conn.close()
        raise

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
                log.info("No pending email-equipped leads with raw_public_data")
                return 0

            log.info("Found %d unverified lead(s)", len(leads))

        for index, (restaurant_id, record, name, claim_id, claim_fingerprint) in enumerate(leads):
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
                        analyzer,
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
            except (OCRDailyBudgetExhausted, OCRTransientError) as exc:
                skipped += 1
                conn.rollback()
                if not args.dry_run:
                    try:
                        with conn.cursor() as cur:
                            for pending in leads[index:]:
                                pending_id, _, _, pending_claim_id, pending_fingerprint = pending
                                if pending_claim_id is None:
                                    continue
                                release_ocr_claim(
                                    cur,
                                    pending_id,
                                    claim_id=pending_claim_id,
                                    claim_fingerprint=pending_fingerprint,
                                    reason=str(exc),
                                )
                        conn.commit()
                    except Exception:
                        conn.rollback()
                        raise
                log.warning("  - %s — OCR deferred safely: %s", label, exc)
                return 0
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
                        summary = (
                            exc.summary
                            if isinstance(exc, OCRIncompleteError)
                            else None
                        )
                        with conn.cursor() as cur:
                            mark_ocr_status(
                                cur,
                                restaurant_id,
                                "failed",
                                claim_id=claim_id,
                                claim_fingerprint=claim_fingerprint,
                                errors=[str(exc)[:500]],
                                summary=summary,
                            )
                        conn.commit()
                    except Exception:
                        conn.rollback()

        log.info("Done: verified=%d failed=%d", verified, failed)
        return 1 if failed else 0
    finally:
        conn.close()
        if budget_conn is not None:
            budget_conn.close()


def main() -> int:
    args = parse_args()
    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s  [%(levelname)s]  %(message)s",
    )

    if not args.watch:
        return run_once(args)

    poll_seconds = max(60, args.poll_seconds or ocr_worker_poll_seconds())
    log.info("OCR background worker started (poll_seconds=%d)", poll_seconds)
    while True:
        try:
            run_once(args)
        except KeyboardInterrupt:
            log.info("OCR background worker stopped")
            return 0
        except Exception:
            log.exception("OCR background cycle failed; retrying after the poll interval")
        time.sleep(poll_seconds)


if __name__ == "__main__":
    raise SystemExit(main())
