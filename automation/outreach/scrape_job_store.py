"""PostgreSQL persistence for resumable city scraping jobs."""

from __future__ import annotations

import json
import os
from typing import Any


def _json(value: Any) -> str:
    return json.dumps(value if value is not None else {})


class ScrapeLeaseLost(Exception):
    """The worker no longer owns a live lease for the scrape job."""


class ScrapeJobStore:
    def __init__(self, database_url: str | None = None) -> None:
        url = (database_url or os.getenv("DATABASE_URL") or "").strip()
        if not url:
            raise ValueError("DATABASE_URL is required by the city scrape worker")
        try:
            import psycopg2
            import psycopg2.extras
        except ImportError as exc:
            raise RuntimeError("psycopg2-binary is required by the city scrape worker") from exc

        self._psycopg2 = psycopg2
        self._extras = psycopg2.extras
        self.conn = psycopg2.connect(url, connect_timeout=10)
        self.conn.autocommit = False
        self._active_worker_id: str | None = None
        self._active_job_id: str | None = None

    def close(self) -> None:
        self.conn.close()

    def _cursor(self):
        return self.conn.cursor(cursor_factory=self._extras.RealDictCursor)

    def _lock_active_job(self, cur, job_id: str) -> None:
        """Fence a mutation to this process's current, unexpired job lease."""
        worker_id = self._active_worker_id
        if not worker_id:
            raise ScrapeLeaseLost("scrape worker does not hold an active job lease")
        if self._active_job_id != job_id:
            raise ScrapeLeaseLost(
                f"scrape job {job_id} is not the job claimed by this worker process"
            )
        cur.execute(
            """
            SELECT id
            FROM scrape_jobs
            WHERE id = %s::uuid
              AND status = 'running'
              AND locked_by = %s
              AND lease_expires_at > now()
            FOR UPDATE
            """,
            (job_id, worker_id),
        )
        if not cur.fetchone():
            raise ScrapeLeaseLost(
                f"scrape job {job_id} lease is expired or owned by another worker"
            )

    def claim_due_job(
        self,
        worker_id: str,
        *,
        lease_seconds: int = 900,
        job_id: str | None = None,
    ) -> dict | None:
        """Claim one queued, due, or lease-expired job.

        A due ``revisit`` starts a new coverage cycle and resets only leaf cells;
        the quadtree built by earlier cycles remains available for future scans.
        """
        try:
            with self._cursor() as cur:
                cur.execute(
                    """
                    SELECT *
                    FROM scrape_jobs
                    WHERE (%s IS NULL OR id = %s::uuid)
                      AND (
                        status = 'queued'
                        OR (status = 'waiting' AND resume_at <= now())
                        OR (status = 'running' AND lease_expires_at < now())
                      )
                    ORDER BY
                      CASE status WHEN 'running' THEN 0 WHEN 'queued' THEN 1 ELSE 2 END,
                      COALESCE(resume_at, created_at), created_at
                    FOR UPDATE SKIP LOCKED
                    LIMIT 1
                    """,
                    (job_id, job_id),
                )
                job = cur.fetchone()
                if not job:
                    self.conn.commit()
                    self._active_worker_id = None
                    self._active_job_id = None
                    return None

                job_id_value = str(job["id"])
                if job["status"] == "running":
                    cur.execute(
                        """
                        UPDATE scrape_job_cells
                        SET status = 'pending', updated_at = now()
                        WHERE scrape_job_id = %s::uuid AND status = 'running'
                        """,
                        (job_id_value,),
                    )

                if job["status"] == "waiting":
                    if job.get("waiting_reason") == "revisit":
                        next_cycle = int(job["cycle_number"]) + 1
                        cur.execute(
                            """
                            UPDATE scrape_job_cells AS cell
                            SET status = 'pending',
                                cycle_number = %s,
                                page_token = NULL,
                                page_number = 0,
                                results_seen_cycle = 0,
                                saturated = false,
                                last_error = NULL,
                                started_at = NULL,
                                completed_at = NULL,
                                updated_at = now()
                            WHERE cell.scrape_job_id = %s::uuid
                              AND cell.status <> 'subdivided'
                              AND NOT EXISTS (
                                SELECT 1 FROM scrape_job_cells child
                                WHERE child.parent_cell_id = cell.id
                              )
                            """,
                            (next_cycle, job_id_value),
                        )
                        cur.execute(
                            """
                            UPDATE scrape_job_candidates
                            SET status = 'discovered', attempts = 0, last_error = NULL,
                                last_seen_cycle = %s, updated_at = now()
                            WHERE scrape_job_id = %s::uuid AND status = 'failed'
                            """,
                            (next_cycle, job_id_value),
                        )
                        job["cycle_number"] = next_cycle
                    elif job.get("waiting_reason") == "coverage_incomplete":
                        cur.execute(
                            """
                            UPDATE scrape_job_cells
                            SET status = 'pending', page_token = NULL, page_number = 0,
                                results_seen_cycle = 0, started_at = NULL,
                                completed_at = NULL, updated_at = now()
                            WHERE scrape_job_id = %s::uuid
                              AND status = 'coverage_incomplete'
                            """,
                            (job_id_value,),
                        )

                    job["requests_used_window"] = 0
                    cur.execute(
                        """
                        UPDATE scrape_jobs
                        SET cycle_number = %s,
                            requests_used_window = 0,
                            window_started_at = now(),
                            resume_at = NULL,
                            waiting_reason = NULL,
                            last_error = NULL,
                            updated_at = now()
                        WHERE id = %s::uuid
                        """,
                        (job["cycle_number"], job_id_value),
                    )

                cur.execute(
                    """
                    UPDATE scrape_jobs
                    SET status = 'running',
                        window_started_at = COALESCE(window_started_at, now()),
                        current_cell_id = NULL,
                        locked_by = %s,
                        locked_at = now(),
                        lease_expires_at = now() + (%s * interval '1 second'),
                        updated_at = now()
                    WHERE id = %s::uuid
                    RETURNING *
                    """,
                    (worker_id, lease_seconds, job_id_value),
                )
                claimed = dict(cur.fetchone())
            self.conn.commit()
            self._active_worker_id = worker_id
            self._active_job_id = job_id_value
            return claimed
        except Exception:
            self.conn.rollback()
            raise

    def renew_lease(self, job_id: str, worker_id: str, lease_seconds: int) -> None:
        if worker_id != self._active_worker_id or job_id != self._active_job_id:
            raise ScrapeLeaseLost("scrape lease renewal came from a non-owning worker")
        try:
            with self._cursor() as cur:
                cur.execute(
                    """
                    UPDATE scrape_jobs
                    SET locked_at = now(),
                        lease_expires_at = now() + (%s * interval '1 second'),
                        updated_at = now()
                    WHERE id = %s::uuid
                      AND status = 'running'
                      AND locked_by = %s
                      AND lease_expires_at > now()
                    """,
                    (lease_seconds, job_id, worker_id),
                )
                if cur.rowcount != 1:
                    raise ScrapeLeaseLost(
                        f"scrape job {job_id} lease could not be renewed"
                    )
            self.conn.commit()
        except Exception:
            self.conn.rollback()
            raise

    def reserve_request(self, job_id: str, worker_id: str, lease_seconds: int) -> tuple[int, int] | None:
        """Atomically reserve one provider call before issuing it."""
        if worker_id != self._active_worker_id or job_id != self._active_job_id:
            raise ScrapeLeaseLost("scrape request reservation came from a non-owning worker")
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                cur.execute(
                    """
                    UPDATE scrape_jobs
                    SET requests_used_window = requests_used_window + 1,
                        requests_used_total = requests_used_total + 1,
                        locked_at = now(),
                        lease_expires_at = now() + (%s * interval '1 second'),
                        updated_at = now()
                    WHERE id = %s::uuid
                      AND requests_used_window < max_requests_per_window
                    RETURNING requests_used_window, max_requests_per_window
                    """,
                    (lease_seconds, job_id),
                )
                row = cur.fetchone()
            self.conn.commit()
            if not row:
                return None
            return int(row["requests_used_window"]), int(row["max_requests_per_window"])
        except Exception:
            self.conn.rollback()
            raise

    def ensure_initial_grid(
        self,
        job_id: str,
        cycle_number: int,
        bounds: dict[str, float],
        *,
        rows: int = 4,
        columns: int = 4,
    ) -> None:
        if rows < 1 or columns < 1:
            raise ValueError("initial grid rows and columns must be positive")
        lat_step = (bounds["high_lat"] - bounds["low_lat"]) / rows
        lng_step = (bounds["high_lng"] - bounds["low_lng"]) / columns
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                cur.execute(
                    "SELECT count(*) AS count FROM scrape_job_cells WHERE scrape_job_id = %s::uuid",
                    (job_id,),
                )
                if int(cur.fetchone()["count"]) > 0:
                    self.conn.commit()
                    return

                for row in range(rows):
                    for column in range(columns):
                        low_lat = bounds["low_lat"] + row * lat_step
                        high_lat = bounds["high_lat"] if row == rows - 1 else low_lat + lat_step
                        low_lng = bounds["low_lng"] + column * lng_step
                        high_lng = bounds["high_lng"] if column == columns - 1 else low_lng + lng_step
                        cur.execute(
                            """
                            INSERT INTO scrape_job_cells (
                                scrape_job_id, cell_key, depth,
                                low_lat, low_lng, high_lat, high_lng,
                                status, cycle_number
                            ) VALUES (%s::uuid, %s, 0, %s, %s, %s, %s, 'pending', %s)
                            ON CONFLICT (scrape_job_id, cell_key) DO NOTHING
                            """,
                            (
                                job_id,
                                f"r{row:02d}c{column:02d}",
                                low_lat,
                                low_lng,
                                high_lat,
                                high_lng,
                                cycle_number,
                            ),
                        )
            self.conn.commit()
        except Exception:
            self.conn.rollback()
            raise

    def claim_next_cell(self, job_id: str) -> dict | None:
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                cur.execute(
                    """
                    SELECT *
                    FROM scrape_job_cells
                    WHERE scrape_job_id = %s::uuid AND status = 'pending'
                    ORDER BY depth, cell_key
                    FOR UPDATE SKIP LOCKED
                    LIMIT 1
                    """,
                    (job_id,),
                )
                cell = cur.fetchone()
                if not cell:
                    self.conn.commit()
                    return None
                cur.execute(
                    """
                    UPDATE scrape_job_cells
                    SET status = 'running', saturated = false, last_error = NULL,
                        started_at = COALESCE(started_at, now()), updated_at = now()
                    WHERE id = %s
                    RETURNING *
                    """,
                    (cell["id"],),
                )
                claimed = dict(cur.fetchone())
                cur.execute(
                    "UPDATE scrape_jobs SET current_cell_id = %s, updated_at = now() WHERE id = %s::uuid",
                    (cell["id"], job_id),
                )
            self.conn.commit()
            return claimed
        except Exception:
            self.conn.rollback()
            raise

    def reset_expired_page_token(self, job_id: str, cell_id: str, message: str) -> None:
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                cur.execute(
                    """
                    UPDATE scrape_job_cells
                    SET status = 'pending', page_token = NULL, page_number = 0,
                        results_seen_cycle = 0, last_error = %s, updated_at = now()
                    WHERE id = %s::uuid AND scrape_job_id = %s::uuid
                    """,
                    (message[:1000], cell_id, job_id),
                )
                if cur.rowcount != 1:
                    raise RuntimeError(
                        "scrape grid cell disappeared while resetting its page token"
                    )
                cur.execute(
                    "UPDATE scrape_jobs SET current_cell_id = NULL, updated_at = now() WHERE id = %s::uuid",
                    (job_id,),
                )
            self.conn.commit()
        except Exception:
            self.conn.rollback()
            raise

    def add_discoveries(
        self,
        job_id: str,
        cell_id: str,
        cycle_number: int,
        places: list[dict],
    ) -> int:
        added = 0
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                for place in places:
                    if str(place.get("businessStatus") or "").upper() == "CLOSED_PERMANENTLY":
                        continue
                    place_id = str(place.get("id") or "").strip()
                    if not place_id:
                        continue
                    cur.execute(
                        """
                        SELECT rp.restaurant_id
                        FROM restaurant_profiles rp
                        WHERE rp.google_place_id = %s
                        LIMIT 1
                        """,
                        (place_id,),
                    )
                    known = cur.fetchone()
                    status = "duplicate" if known else "discovered"
                    restaurant_id = known["restaurant_id"] if known else None
                    cur.execute(
                        """
                        INSERT INTO scrape_job_candidates (
                            scrape_job_id, first_cell_id, google_place_id, status,
                            first_seen_cycle, last_seen_cycle, discovery_data, restaurant_id
                        ) VALUES (%s::uuid, %s::uuid, %s, %s, %s, %s, %s::jsonb, %s)
                        ON CONFLICT (scrape_job_id, google_place_id) DO UPDATE SET
                            last_seen_cycle = EXCLUDED.last_seen_cycle,
                            discovery_data = EXCLUDED.discovery_data,
                            restaurant_id = COALESCE(scrape_job_candidates.restaurant_id, EXCLUDED.restaurant_id),
                            status = CASE
                                WHEN scrape_job_candidates.status IN ('imported', 'duplicate')
                                  THEN scrape_job_candidates.status
                                WHEN scrape_job_candidates.status = 'failed'
                                  AND EXCLUDED.last_seen_cycle > scrape_job_candidates.last_seen_cycle
                                  THEN EXCLUDED.status
                                ELSE scrape_job_candidates.status
                            END,
                            attempts = CASE
                                WHEN scrape_job_candidates.status = 'failed'
                                  AND EXCLUDED.last_seen_cycle > scrape_job_candidates.last_seen_cycle
                                  THEN 0
                                ELSE scrape_job_candidates.attempts
                            END,
                            last_error = CASE
                                WHEN scrape_job_candidates.status = 'failed'
                                  AND EXCLUDED.last_seen_cycle > scrape_job_candidates.last_seen_cycle
                                  THEN NULL
                                ELSE scrape_job_candidates.last_error
                            END,
                            updated_at = now()
                        RETURNING (xmax = 0) AS inserted
                        """,
                        (
                            job_id,
                            cell_id,
                            place_id,
                            status,
                            cycle_number,
                            cycle_number,
                            _json(place),
                            restaurant_id,
                        ),
                    )
                    if cur.fetchone()["inserted"]:
                        added += 1
            self.conn.commit()
            return added
        except Exception:
            self.conn.rollback()
            raise

    def checkpoint_cell_page(
        self,
        job_id: str,
        cell_id: str,
        *,
        result_count: int,
        next_page_token: str,
    ) -> dict:
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                cur.execute(
                    """
                    UPDATE scrape_job_cells
                    SET page_number = page_number + 1,
                        page_token = NULLIF(%s, ''),
                        results_seen_cycle = results_seen_cycle + %s,
                        results_seen_total = results_seen_total + %s,
                        updated_at = now()
                    WHERE id = %s::uuid AND scrape_job_id = %s::uuid
                    RETURNING *
                    """,
                    (next_page_token, result_count, result_count, cell_id, job_id),
                )
                row = cur.fetchone()
                if not row:
                    raise RuntimeError("scrape grid cell disappeared during checkpoint")
                cell = dict(row)
            self.conn.commit()
            return cell
        except Exception:
            self.conn.rollback()
            raise

    def continue_cell(self, job_id: str, cell_id: str) -> None:
        self._finish_cell_state(job_id, cell_id, "pending", completed=False, saturated=False)

    def complete_cell(
        self,
        job_id: str,
        cell_id: str,
        *,
        saturated: bool = False,
        error: str = "",
    ) -> None:
        self._finish_cell_state(
            job_id,
            cell_id,
            "completed",
            completed=True,
            saturated=saturated,
            error=error,
        )

    def _finish_cell_state(
        self,
        job_id: str,
        cell_id: str,
        status: str,
        *,
        completed: bool,
        saturated: bool,
        error: str = "",
    ) -> None:
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                cur.execute(
                    """
                    UPDATE scrape_job_cells
                    SET status = %s,
                        saturated = saturated OR %s,
                        last_error = CASE
                            WHEN %s <> '' THEN %s
                            ELSE last_error
                        END,
                        completed_at = CASE WHEN %s THEN now() ELSE completed_at END,
                        updated_at = now()
                    WHERE id = %s::uuid AND scrape_job_id = %s::uuid
                    """,
                    (
                        status,
                        saturated,
                        error[:1000],
                        error[:1000],
                        completed,
                        cell_id,
                        job_id,
                    ),
                )
                if cur.rowcount != 1:
                    raise RuntimeError("scrape grid cell disappeared while changing state")
                cur.execute(
                    "UPDATE scrape_jobs SET current_cell_id = NULL, updated_at = now() WHERE id = %s::uuid",
                    (job_id,),
                )
            self.conn.commit()
        except Exception:
            self.conn.rollback()
            raise

    def subdivide_cell(self, job_id: str, cell_id: str, cycle_number: int) -> None:
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                cur.execute(
                    "SELECT * FROM scrape_job_cells WHERE id = %s::uuid AND scrape_job_id = %s::uuid FOR UPDATE",
                    (cell_id, job_id),
                )
                cell = cur.fetchone()
                if not cell:
                    raise RuntimeError("scrape grid cell disappeared during subdivision")
                mid_lat = (float(cell["low_lat"]) + float(cell["high_lat"])) / 2
                mid_lng = (float(cell["low_lng"]) + float(cell["high_lng"])) / 2
                children = (
                    (float(cell["low_lat"]), float(cell["low_lng"]), mid_lat, mid_lng),
                    (float(cell["low_lat"]), mid_lng, mid_lat, float(cell["high_lng"])),
                    (mid_lat, float(cell["low_lng"]), float(cell["high_lat"]), mid_lng),
                    (mid_lat, mid_lng, float(cell["high_lat"]), float(cell["high_lng"])),
                )
                for index, bounds in enumerate(children):
                    cur.execute(
                        """
                        INSERT INTO scrape_job_cells (
                            scrape_job_id, parent_cell_id, cell_key, depth,
                            low_lat, low_lng, high_lat, high_lng,
                            status, cycle_number
                        ) VALUES (%s::uuid, %s::uuid, %s, %s, %s, %s, %s, %s, 'pending', %s)
                        ON CONFLICT (scrape_job_id, cell_key) DO UPDATE SET
                            status = 'pending', cycle_number = EXCLUDED.cycle_number,
                            page_token = NULL, page_number = 0, results_seen_cycle = 0,
                            saturated = false, completed_at = NULL, updated_at = now()
                        """,
                        (
                            job_id,
                            cell_id,
                            f"{cell['cell_key']}/{index}",
                            int(cell["depth"]) + 1,
                            *bounds,
                            cycle_number,
                        ),
                    )
                cur.execute(
                    """
                    UPDATE scrape_job_cells
                    SET status = 'subdivided', saturated = true, page_token = NULL,
                        completed_at = now(), updated_at = now()
                    WHERE id = %s::uuid
                    """,
                    (cell_id,),
                )
                cur.execute(
                    "UPDATE scrape_jobs SET current_cell_id = NULL, updated_at = now() WHERE id = %s::uuid",
                    (job_id,),
                )
            self.conn.commit()
        except Exception:
            self.conn.rollback()
            raise

    def next_candidate(self, job_id: str) -> dict | None:
        try:
            with self._cursor() as cur:
                cur.execute(
                    """
                    SELECT *
                    FROM scrape_job_candidates
                    WHERE scrape_job_id = %s::uuid
                      AND status IN ('discovered', 'details_ready', 'enriched')
                    ORDER BY created_at, id
                    LIMIT 1
                    """,
                    (job_id,),
                )
                row = cur.fetchone()
            self.conn.commit()
            return dict(row) if row else None
        except Exception:
            self.conn.rollback()
            raise

    def save_candidate(
        self,
        candidate_id: str,
        status: str,
        record: dict,
        *,
        error: str = "",
        increment_attempts: bool = False,
    ) -> None:
        try:
            with self._cursor() as cur:
                cur.execute(
                    "SELECT scrape_job_id::text FROM scrape_job_candidates WHERE id = %s::uuid",
                    (candidate_id,),
                )
                candidate = cur.fetchone()
                if not candidate:
                    raise RuntimeError("scrape candidate disappeared before it could be saved")
                self._lock_active_job(cur, str(candidate["scrape_job_id"]))
                cur.execute(
                    """
                    UPDATE scrape_job_candidates
                    SET status = %s,
                        scrape_record = %s::jsonb,
                        attempts = attempts + CASE WHEN %s THEN 1 ELSE 0 END,
                        last_error = NULLIF(%s, ''),
                        updated_at = now()
                    WHERE id = %s::uuid
                    """,
                    (status, _json(record), increment_attempts, error[:1000], candidate_id),
                )
                if cur.rowcount != 1:
                    raise RuntimeError("scrape candidate disappeared before it could be saved")
            self.conn.commit()
        except Exception:
            self.conn.rollback()
            raise

    def import_candidate(self, candidate_id: str, job_id: str, record: dict) -> str:
        from import_to_db import upsert_restaurant_record, verify_schema

        try:
            with self.conn.cursor() as cur:
                self._lock_active_job(cur, job_id)
                cur.execute(
                    """
                    SELECT 1
                    FROM scrape_job_candidates
                    WHERE id = %s::uuid AND scrape_job_id = %s::uuid
                    FOR UPDATE
                    """,
                    (candidate_id, job_id),
                )
                if not cur.fetchone():
                    raise RuntimeError("scrape candidate does not belong to the active job")
                verify_schema(cur)
                restaurant_id, _ = upsert_restaurant_record(
                    cur,
                    record,
                    f"scrape_job:{job_id}",
                )
                cur.execute(
                    """
                    UPDATE scrape_job_candidates
                    SET status = 'imported', restaurant_id = %s, scrape_record = %s::jsonb,
                        last_error = NULL, updated_at = now()
                    WHERE id = %s::uuid AND scrape_job_id = %s::uuid
                    """,
                    (restaurant_id, _json(record), candidate_id, job_id),
                )
            self.conn.commit()
            return str(restaurant_id)
        except Exception:
            self.conn.rollback()
            raise

    def has_pending_work(self, job_id: str) -> bool:
        try:
            with self._cursor() as cur:
                cur.execute(
                    """
                    SELECT EXISTS (
                        SELECT 1 FROM scrape_job_cells
                        WHERE scrape_job_id = %s::uuid AND status IN ('pending', 'running')
                    ) OR EXISTS (
                        SELECT 1 FROM scrape_job_candidates
                        WHERE scrape_job_id = %s::uuid
                          AND status IN ('discovered', 'details_ready', 'enriched')
                    ) AS pending
                    """,
                    (job_id, job_id),
                )
                pending = bool(cur.fetchone()["pending"])
            self.conn.commit()
            return pending
        except Exception:
            self.conn.rollback()
            raise

    def pause_for_request_limit(self, job_id: str) -> None:
        self._set_waiting(job_id, "request_limit")

    def finish_cycle(self, job_id: str) -> None:
        self._set_waiting(job_id, "revisit", cycle_completed=True)

    def pause_for_provider_error(self, job_id: str, error: str) -> None:
        self._set_waiting(job_id, "provider_error", error=error)

    def pause_for_coverage_incomplete(self, job_id: str, error: str) -> None:
        self._set_waiting(job_id, "coverage_incomplete", error=error)

    def incomplete_coverage_count(self, job_id: str) -> int:
        try:
            with self._cursor() as cur:
                cur.execute(
                    """
                    SELECT count(*) AS count
                    FROM scrape_job_cells
                    WHERE scrape_job_id = %s::uuid
                      AND status = 'coverage_incomplete'
                    """,
                    (job_id,),
                )
                count = int(cur.fetchone()["count"])
            self.conn.commit()
            return count
        except Exception:
            self.conn.rollback()
            raise

    def mark_cell_coverage_incomplete(self, job_id: str, cell_id: str, error: str) -> None:
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                cur.execute(
                    """
                    UPDATE scrape_job_cells
                    SET status = 'coverage_incomplete', saturated = true, last_error = %s,
                        page_token = NULL, page_number = 0,
                        results_seen_cycle = 0, started_at = NULL,
                        completed_at = NULL, updated_at = now()
                    WHERE id = %s::uuid AND scrape_job_id = %s::uuid
                    """,
                    (error[:1000], cell_id, job_id),
                )
                if cur.rowcount != 1:
                    raise RuntimeError(
                        "scrape grid cell disappeared while recording incomplete coverage"
                    )
                cur.execute(
                    "UPDATE scrape_jobs SET current_cell_id = NULL, updated_at = now() WHERE id = %s::uuid",
                    (job_id,),
                )
            self.conn.commit()
        except Exception:
            self.conn.rollback()
            raise

    def _set_waiting(
        self,
        job_id: str,
        reason: str,
        *,
        cycle_completed: bool = False,
        error: str = "",
    ) -> None:
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                cur.execute(
                    """
                    UPDATE scrape_jobs
                    SET status = 'waiting', waiting_reason = %s,
                        resume_at = now() + interval '24 hours',
                        last_cycle_completed_at = CASE WHEN %s THEN now() ELSE last_cycle_completed_at END,
                        last_error = NULLIF(%s, ''),
                        current_cell_id = NULL, locked_by = NULL, locked_at = NULL,
                        lease_expires_at = NULL, updated_at = now()
                    WHERE id = %s::uuid
                    """,
                    (reason, cycle_completed, error[:2000], job_id),
                )
                cur.execute(
                    """
                    UPDATE scrape_job_cells SET status = 'pending', updated_at = now()
                    WHERE scrape_job_id = %s::uuid AND status = 'running'
                    """,
                    (job_id,),
                )
            self.conn.commit()
            self._active_worker_id = None
            self._active_job_id = None
        except Exception:
            self.conn.rollback()
            raise

    def fail_job(
        self,
        job_id: str,
        error: str,
        *,
        cell_id: str | None = None,
        saturated: bool = False,
    ) -> None:
        try:
            with self._cursor() as cur:
                self._lock_active_job(cur, job_id)
                if cell_id:
                    cur.execute(
                        """
                        UPDATE scrape_job_cells
                        SET status = 'failed', saturated = saturated OR %s,
                            last_error = %s, page_token = NULL, updated_at = now()
                        WHERE id = %s::uuid AND scrape_job_id = %s::uuid
                        """,
                        (saturated, error[:1000], cell_id, job_id),
                    )
                    if cur.rowcount != 1:
                        raise RuntimeError(
                            "saturated scrape grid cell disappeared before the job could fail"
                        )
                cur.execute(
                    """
                    UPDATE scrape_jobs
                    SET status = 'failed', last_error = %s, current_cell_id = NULL,
                        locked_by = NULL, locked_at = NULL, lease_expires_at = NULL,
                        updated_at = now()
                    WHERE id = %s::uuid
                    """,
                    (error[:2000], job_id),
                )
                cur.execute(
                    """
                    UPDATE scrape_job_cells SET status = 'pending', last_error = %s, updated_at = now()
                    WHERE scrape_job_id = %s::uuid AND status = 'running'
                      AND (%s::uuid IS NULL OR id <> %s::uuid)
                    """,
                    (error[:1000], job_id, cell_id, cell_id),
                )
            self.conn.commit()
            self._active_worker_id = None
            self._active_job_id = None
        except Exception:
            self.conn.rollback()
            raise


class PersistentRequestBudget:
    """RequestBudget-compatible facade backed by ``scrape_jobs``."""

    def __init__(
        self,
        store: ScrapeJobStore,
        job_id: str,
        worker_id: str,
        *,
        used: int,
        maximum: int,
        lease_seconds: int,
    ) -> None:
        self.store = store
        self.job_id = job_id
        self.worker_id = worker_id
        self.total_used = int(used)
        self.max_requests = int(maximum)
        self.lease_seconds = lease_seconds

    @property
    def remaining(self) -> int:
        return max(0, self.max_requests - self.total_used)

    @property
    def exhausted(self) -> bool:
        return self.total_used >= self.max_requests

    def can_consume(self, count: int = 1) -> bool:
        return count > 0 and self.total_used + count <= self.max_requests

    def consume(self, count: int = 1) -> None:
        from request_budget import BudgetExhausted

        if count < 1:
            return
        for _ in range(count):
            reserved = self.store.reserve_request(
                self.job_id,
                self.worker_id,
                self.lease_seconds,
            )
            if reserved is None:
                raise BudgetExhausted(
                    f"request budget exhausted ({self.total_used}/{self.max_requests} used)"
                )
            self.total_used, self.max_requests = reserved
