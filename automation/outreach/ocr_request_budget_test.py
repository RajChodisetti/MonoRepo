"""Unit tests for the durable OCR daily request budget."""

from datetime import datetime, timezone
import unittest

from ocr_request_budget import (
    DurableOCRRequestBudget,
    OCRDailyBudgetExhausted,
    seconds_until_utc_reset,
)


class FakeCursor:
    def __init__(self, conn):
        self.conn = conn

    def __enter__(self):
        return self

    def __exit__(self, *_args):
        return False

    def execute(self, query, params):
        self.conn.last_query = query
        self.conn.last_params = params

    def fetchone(self):
        return self.conn.next_row


class FakeConnection:
    def __init__(self, next_row):
        self.next_row = next_row
        self.commits = 0
        self.rollbacks = 0
        self.last_query = ""
        self.last_params = ()

    def cursor(self):
        return FakeCursor(self)

    def commit(self):
        self.commits += 1

    def rollback(self):
        self.rollbacks += 1


class DurableOCRRequestBudgetTest(unittest.TestCase):
    def test_reserve_returns_remaining_and_commits(self):
        conn = FakeConnection((17, 200))
        budget = DurableOCRRequestBudget(conn)

        snapshot = budget.reserve()

        self.assertEqual(snapshot.requests_used, 17)
        self.assertEqual(snapshot.remaining, 183)
        self.assertEqual(conn.commits, 1)
        self.assertIn("ON CONFLICT", conn.last_query)
        self.assertEqual(conn.last_params, ("menu_ocr", 200))

    def test_exhausted_budget_does_not_allow_a_provider_call(self):
        conn = FakeConnection(None)
        budget = DurableOCRRequestBudget(conn)

        with self.assertRaises(OCRDailyBudgetExhausted):
            budget.reserve()

        self.assertEqual(conn.commits, 1)

    def test_limit_cannot_exceed_provider_allowance(self):
        with self.assertRaises(ValueError):
            DurableOCRRequestBudget(FakeConnection(None), daily_limit=201)

    def test_utc_reset_handles_month_boundary(self):
        now = datetime(2026, 7, 31, 23, 59, 30, tzinfo=timezone.utc)
        self.assertEqual(seconds_until_utc_reset(now), 30)


if __name__ == "__main__":
    unittest.main()
