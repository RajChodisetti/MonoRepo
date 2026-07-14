#!/usr/bin/env python3
"""Run durable Places-first city scrape jobs created by the Go API."""

from __future__ import annotations

import argparse
import logging
import os
import socket
import time
import uuid

from env_loader import load_project_env

load_project_env()

from apollo_enrichment import (  # noqa: E402
    enrich_missing_contact_with_apollo,
    get_apollo_api_key,
    needs_apollo_enrichment,
)
from google_places_scraper import (  # noqa: E402
    discover_places_page,
    get_city_search_bounds,
    get_places_api_key,
    place_to_lead_dict,
    PlacesAPIError,
    scrape_single_restaurant_places,
)
from niche_config import get_niche  # noqa: E402
from request_budget import BudgetExhausted  # noqa: E402
from scrape_job_store import (  # noqa: E402
    PersistentRequestBudget,
    ScrapeJobStore,
    ScrapeLeaseLost,
)
from tuvi_outreach_agent import Config  # noqa: E402

log = logging.getLogger("city_scrape_worker")


class PermanentScrapeJobError(RuntimeError):
    """A deterministic job/configuration error that must not retry in 24 hours."""


class CoverageIncomplete(RuntimeError):
    """A provider result cap prevented proof of complete cell coverage."""


def _env_int(name: str, default: int, minimum: int = 1) -> int:
    raw = os.getenv(name, str(default)).strip()
    try:
        value = int(raw)
    except ValueError as exc:
        raise ValueError(f"{name} must be an integer") from exc
    if value < minimum:
        raise ValueError(f"{name} must be at least {minimum}")
    return value


class CityScrapeWorker:
    def __init__(self, store: ScrapeJobStore, cfg: Config | None = None) -> None:
        self.store = store
        self.cfg = cfg or Config()
        self.worker_id = os.getenv("SCRAPE_WORKER_ID", "").strip() or (
            f"{socket.gethostname()}:{os.getpid()}"
        )
        self.lease_seconds = _env_int("SCRAPE_JOB_LEASE_SECONDS", 900, 60)
        self.initial_rows = _env_int("SCRAPE_INITIAL_GRID_ROWS", 4)
        self.initial_columns = _env_int("SCRAPE_INITIAL_GRID_COLUMNS", 4)
        self.page_limit = _env_int("SCRAPE_CELL_PAGE_LIMIT", 3)
        self.max_depth = _env_int("SCRAPE_GRID_MAX_DEPTH", 12, 0)

    def claim_and_run(self, job_id: str | None = None) -> bool:
        job = self.store.claim_due_job(
            self.worker_id,
            lease_seconds=self.lease_seconds,
            job_id=job_id,
        )
        if not job:
            return False

        job_id_value = str(job["id"])
        try:
            self._run_job(job)
        except ScrapeLeaseLost as exc:
            log.warning(
                "scrape_job_lease_lost job_id=%s error=%s",
                job_id_value,
                exc,
            )
        except PermanentScrapeJobError as exc:
            log.error(
                "scrape_job_permanent_failure job_id=%s error=%s",
                job_id_value,
                exc,
            )
            try:
                self.store.fail_job(job_id_value, str(exc))
            except ScrapeLeaseLost as lease_exc:
                log.warning(
                    "scrape_job_failure_not_recorded_after_lease_loss job_id=%s error=%s",
                    job_id_value,
                    lease_exc,
                )
        except BudgetExhausted:
            log.info("scrape_request_window_exhausted", extra={"job_id": job_id_value})
            try:
                self.store.pause_for_request_limit(job_id_value)
            except ScrapeLeaseLost as exc:
                log.warning(
                    "scrape_job_limit_pause_skipped_after_lease_loss job_id=%s error=%s",
                    job_id_value,
                    exc,
                )
        except CoverageIncomplete as exc:
            log.error(
                "scrape_job_coverage_incomplete job_id=%s error=%s",
                job_id_value,
                exc,
            )
            try:
                self.store.pause_for_coverage_incomplete(job_id_value, str(exc))
            except ScrapeLeaseLost as lease_exc:
                log.warning(
                    "scrape_job_coverage_pause_skipped_after_lease_loss job_id=%s error=%s",
                    job_id_value,
                    lease_exc,
                )
        except Exception as exc:
            log.exception("scrape_job_provider_pause", extra={"job_id": job_id_value})
            try:
                self.store.pause_for_provider_error(job_id_value, str(exc))
            except ScrapeLeaseLost as lease_exc:
                log.warning(
                    "scrape_job_provider_pause_skipped_after_lease_loss job_id=%s error=%s",
                    job_id_value,
                    lease_exc,
                )
        return True

    def _run_job(self, job: dict) -> None:
        job_id = str(job["id"])
        city = str(job["city"])
        try:
            niche = get_niche(str(job["niche"]))
            get_places_api_key(self.cfg)
            apollo_enabled = bool(getattr(self.cfg, "APOLLO_ENRICHMENT_ENABLED", True))
            if apollo_enabled:
                get_apollo_api_key(self.cfg)
            bounds = get_city_search_bounds(city)
        except ValueError as exc:
            raise PermanentScrapeJobError(
                f"scrape job preflight configuration failed: {exc}"
            ) from exc
        self.store.ensure_initial_grid(
            job_id,
            int(job["cycle_number"]),
            bounds,
            rows=self.initial_rows,
            columns=self.initial_columns,
        )
        budget = PersistentRequestBudget(
            self.store,
            job_id,
            self.worker_id,
            used=int(job["requests_used_window"]),
            maximum=int(job["max_requests_per_window"]),
            lease_seconds=self.lease_seconds,
        )

        if budget.exhausted:
            raise BudgetExhausted("request window was already exhausted")

        log.info(
            "scrape_job_started job_id=%s city=%s niche=%s cycle=%s used=%s/%s",
            job_id,
            city,
            niche.slug,
            job["cycle_number"],
            budget.total_used,
            budget.max_requests,
        )

        while True:
            if budget.exhausted:
                raise BudgetExhausted("request window reached its configured maximum")

            candidate = self.store.next_candidate(job_id)
            if candidate:
                self._process_candidate(job, candidate, niche, budget, apollo_enabled)
                continue

            cell = self.store.claim_next_cell(job_id)
            if cell:
                self._process_cell(job, cell, niche.slug, budget)
                continue

            if self.store.has_pending_work(job_id):
                continue

            incomplete_count = self.store.incomplete_coverage_count(job_id)
            if incomplete_count:
                raise CoverageIncomplete(
                    f"{incomplete_count} grid leaf/leaves remain provider-capped at maximum depth"
                )

            self.store.finish_cycle(job_id)
            log.info(
                "scrape_cycle_completed job_id=%s city=%s cycle=%s; revisit scheduled in 24h",
                job_id,
                city,
                job["cycle_number"],
            )
            return

    def _process_cell(
        self,
        job: dict,
        cell: dict,
        niche: str,
        budget: PersistentRequestBudget,
    ) -> None:
        job_id = str(job["id"])
        cell_id = str(cell["id"])
        page_token = str(cell.get("page_token") or "")
        cell_bounds = {
            "low_lat": float(cell["low_lat"]),
            "low_lng": float(cell["low_lng"]),
            "high_lat": float(cell["high_lat"]),
            "high_lng": float(cell["high_lng"]),
        }
        try:
            places, next_page_token = discover_places_page(
                city=str(job["city"]),
                niche=niche,
                bounds=cell_bounds,
                page_token=page_token,
                page_size=20,
                cfg=self.cfg,
                budget=budget,
                label=f"scrape-job:{job_id}:cell:{cell['cell_key']}:p{int(cell['page_number']) + 1}",
            )
        except PlacesAPIError as exc:
            # Places page tokens are provider-owned and may expire during a
            # 24-hour pause. Restarting this cell is safe because Place IDs are
            # deduplicated in both candidates and restaurant_profiles.
            if page_token and exc.status_code == 400:
                self.store.reset_expired_page_token(
                    job_id,
                    cell_id,
                    "Persisted Places page token expired; cell restarted from page one.",
                )
                return
            raise PermanentScrapeJobError(
                f"Places discovery rejected the durable grid request: {exc}"
            ) from exc

        self.store.add_discoveries(
            job_id,
            cell_id,
            int(job["cycle_number"]),
            places,
        )
        updated = self.store.checkpoint_cell_page(
            job_id,
            cell_id,
            result_count=len(places),
            next_page_token=next_page_token,
        )

        reached_density_threshold = (
            int(updated["page_number"]) >= self.page_limit
            and (
                bool(next_page_token)
                or int(updated["results_seen_cycle"]) >= self.page_limit * 20
            )
        )
        if reached_density_threshold and int(updated["depth"]) < self.max_depth:
            self.store.subdivide_cell(
                job_id,
                cell_id,
                int(job["cycle_number"]),
            )
        elif next_page_token:
            self.store.continue_cell(job_id, cell_id)
        elif reached_density_threshold:
            message = (
                "coverage incomplete: grid cell "
                f"{cell['cell_key']} remained saturated at maximum depth "
                f"{self.max_depth} after {int(updated['results_seen_cycle'])} results"
            )
            log.warning(
                "scrape_cell_saturated_at_max_depth job_id=%s cell_id=%s error=%s",
                job_id,
                cell_id,
                message,
            )
            self.store.mark_cell_coverage_incomplete(job_id, cell_id, message)
            return
        else:
            self.store.complete_cell(
                job_id,
                cell_id,
                saturated=reached_density_threshold,
            )

    def _process_candidate(
        self,
        job: dict,
        candidate: dict,
        niche,
        budget: PersistentRequestBudget,
        apollo_enabled: bool,
    ) -> None:
        job_id = str(job["id"])
        candidate_id = str(candidate["id"])
        status = str(candidate["status"])
        record = candidate.get("scrape_record") or {}

        if status == "discovered":
            discovery = candidate.get("discovery_data") or {}
            lead = place_to_lead_dict(discovery, str(job["city"]), niche.slug)
            record = scrape_single_restaurant_places(
                lead,
                self.cfg,
                max_reviews=5,
                query_suffix=niche.places_query_suffix,
                budget=budget,
                # The durable flow relies on targeted Apollo for missing email
                # rather than issuing arbitrary restaurant-website requests.
                lookup_website_email=False,
            )
            scrape_status = str(record.get("scrape_status") or "")
            if scrape_status == "budget_exhausted":
                raise BudgetExhausted("request window exhausted during Place Details")
            if scrape_status == "error":
                detail_error = "; ".join(
                    str(item) for item in (record.get("errors") or [])
                )
                raise RuntimeError(
                    "Google Places detail enrichment failed: "
                    + (detail_error or "provider error")
                )
            if scrape_status == "permanent_error":
                detail_error = "; ".join(
                    str(item) for item in (record.get("errors") or [])
                )
                raise PermanentScrapeJobError(
                    "Google Places permanently rejected Place Details: "
                    + (detail_error or "provider error")
                )
            if scrape_status != "success":
                self.store.save_candidate(
                    candidate_id,
                    "failed",
                    record,
                    error="; ".join(str(item) for item in (record.get("errors") or [])),
                    increment_attempts=True,
                )
                return
            self.store.save_candidate(candidate_id, "details_ready", record)
            status = "details_ready"

        if status == "details_ready":
            if apollo_enabled and needs_apollo_enrichment(record):
                record, apollo_stats = enrich_missing_contact_with_apollo(
                    record,
                    self.cfg,
                    niche,
                    budget=budget,
                )
                record["apollo_enrichment"] = apollo_stats
                apollo_status = str(apollo_stats.get("status") or "")
                if apollo_status == "budget_exhausted":
                    self.store.save_candidate(candidate_id, "details_ready", record)
                    raise BudgetExhausted("request window exhausted during Apollo enrichment")
                if apollo_status == "error":
                    error_code = apollo_stats.get("error_code")
                    if isinstance(error_code, int) and 400 <= error_code < 500 and error_code != 429:
                        if error_code == 404:
                            # Apollo is optional enrichment after Places. A
                            # person-match miss must not discard an otherwise
                            # usable Places lead.
                            apollo_stats["status"] = "no_match"
                            apollo_stats.pop("error", None)
                            record["apollo_enrichment"] = apollo_stats
                            log.info(
                                "Apollo returned no owner/contact match for %s; importing the Places lead",
                                record.get("name") or record.get("google_place_id") or candidate_id,
                            )
                        else:
                            raise PermanentScrapeJobError(
                                "Apollo permanently rejected contact enrichment "
                                f"with HTTP {error_code}"
                            )
                    else:
                        raise RuntimeError(
                            "Apollo contact enrichment failed: "
                            f"{apollo_stats.get('error') or 'provider error'}"
                        )
            self.store.save_candidate(candidate_id, "enriched", record)
            status = "enriched"

        if status == "enriched":
            self.store.import_candidate(candidate_id, job_id, record)


def _valid_job_id(value: str) -> str:
    try:
        return str(uuid.UUID(value))
    except ValueError as exc:
        raise argparse.ArgumentTypeError("job id must be a valid UUID") from exc


def main() -> int:
    parser = argparse.ArgumentParser(description="Run durable city scrape jobs from PostgreSQL")
    parser.add_argument("--once", action="store_true", help="Claim at most one due job, then exit")
    parser.add_argument("--job-id", type=_valid_job_id, help="Run one specific due job")
    parser.add_argument(
        "--poll-seconds",
        type=int,
        default=_env_int("SCRAPE_WORKER_POLL_SECONDS", 15),
    )
    parser.add_argument("-v", "--verbose", action="store_true")
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s [%(levelname)s] %(name)s %(message)s",
    )

    store = ScrapeJobStore()
    worker = CityScrapeWorker(store)
    try:
        if args.job_id:
            ran = worker.claim_and_run(args.job_id)
            if not ran:
                log.info("requested scrape job is not due or is held by another worker")
            return 0
        if args.once:
            worker.claim_and_run()
            return 0

        while True:
            ran = worker.claim_and_run()
            if not ran:
                time.sleep(max(1, args.poll_seconds))
    except KeyboardInterrupt:
        log.info("city scrape worker stopped")
        return 0
    except (ValueError, RuntimeError) as exc:
        log.error("%s", exc)
        return 1
    finally:
        store.close()


if __name__ == "__main__":
    raise SystemExit(main())
