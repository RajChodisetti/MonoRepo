"""PostgreSQL-backed request budget for the unattended menu OCR worker."""

from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime, timedelta, timezone


class OCRDailyBudgetExhausted(RuntimeError):
    """Raised before a provider call when the UTC daily budget is exhausted."""


@dataclass(frozen=True)
class OCRBudgetSnapshot:
    requests_used: int
    daily_limit: int

    @property
    def remaining(self) -> int:
        return max(0, self.daily_limit - self.requests_used)


class DurableOCRRequestBudget:
    """Atomically reserves one global OCR request in an independent transaction."""

    def __init__(self, conn, *, budget_key: str = "menu_ocr", daily_limit: int = 200):
        budget_key = str(budget_key or "").strip()
        if not budget_key:
            raise ValueError("budget_key is required")
        if daily_limit < 1 or daily_limit > 200:
            raise ValueError("daily_limit must be between 1 and 200")
        self.conn = conn
        self.budget_key = budget_key
        self.daily_limit = daily_limit

    def snapshot(self) -> OCRBudgetSnapshot:
        with self.conn.cursor() as cur:
            cur.execute(
                """
                SELECT requests_used, daily_limit
                FROM ocr_daily_request_usage
                WHERE usage_date = (now() AT TIME ZONE 'UTC')::date
                  AND budget_key = %s
                """,
                (self.budget_key,),
            )
            row = cur.fetchone()
        self.conn.commit()
        if row is None:
            return OCRBudgetSnapshot(0, self.daily_limit)
        return OCRBudgetSnapshot(int(row[0]), int(row[1]))

    def reserve(self) -> OCRBudgetSnapshot:
        """Reserve before the call, so timeouts still consume provider budget."""
        try:
            with self.conn.cursor() as cur:
                cur.execute(
                    """
                    INSERT INTO ocr_daily_request_usage (
                      usage_date, budget_key, requests_used, daily_limit,
                      last_requested_at
                    ) VALUES (
                      (now() AT TIME ZONE 'UTC')::date, %s, 1, %s, now()
                    )
                    ON CONFLICT (usage_date, budget_key) DO UPDATE
                    SET requests_used = ocr_daily_request_usage.requests_used + 1,
                        daily_limit = EXCLUDED.daily_limit,
                        last_requested_at = now(),
                        updated_at = now()
                    WHERE ocr_daily_request_usage.requests_used < EXCLUDED.daily_limit
                    RETURNING requests_used, daily_limit
                    """,
                    (self.budget_key, self.daily_limit),
                )
                row = cur.fetchone()
            self.conn.commit()
        except Exception:
            self.conn.rollback()
            raise
        if row is None:
            raise OCRDailyBudgetExhausted(
                f"OCR daily request budget exhausted (limit={self.daily_limit}, resets at 00:00 UTC)"
            )
        return OCRBudgetSnapshot(int(row[0]), int(row[1]))


def seconds_until_utc_reset(now: datetime | None = None) -> int:
    now = now or datetime.now(timezone.utc)
    tomorrow = now.replace(hour=0, minute=0, second=0, microsecond=0)
    # Adding one day through a timestamp avoids month/year boundary handling.
    tomorrow += timedelta(days=1)
    return max(1, int((tomorrow - now).total_seconds()))
