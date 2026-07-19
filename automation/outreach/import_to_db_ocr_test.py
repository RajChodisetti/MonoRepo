"""Query-contract tests for email-only OCR claiming and safe claim release."""

import uuid
import unittest

from import_to_db import fetch_unverified_leads, mark_ocr_status, release_ocr_claim


class FakeCursor:
    def __init__(self):
        self.query = ""
        self.queries = []
        self.params = ()
        self.params_history = []
        self.rowcount = 1

    def execute(self, query, params=()):
        self.query = query
        self.queries.append(query)
        self.params = params
        self.params_history.append(params)

    def fetchall(self):
        return []


class OCRClaimQueryTest(unittest.TestCase):
    def test_preview_selects_only_restaurants_with_email(self):
        cur = FakeCursor()

        rows = fetch_unverified_leads(cur, 10, claim=False)

        self.assertEqual(rows, [])
        self.assertIn("NULLIF(BTRIM(r.email), '') IS NOT NULL", cur.query)
        self.assertIn("demo.status = 'published'", cur.query)

    def test_claim_selects_only_restaurants_with_email(self):
        cur = FakeCursor()

        rows = fetch_unverified_leads(cur, 10, claim=True)

        self.assertEqual(rows, [])
        self.assertIn("NULLIF(BTRIM(r.email), '') IS NOT NULL", cur.query)
        self.assertIn("demo.status = 'published'", cur.query)

    def test_transient_release_restores_pending_without_attempt_penalty(self):
        cur = FakeCursor()
        restaurant_id = uuid.uuid4()
        claim_id = uuid.uuid4()

        release_ocr_claim(
            cur,
            restaurant_id,
            claim_id=claim_id,
            claim_fingerprint="fingerprint",
            reason="provider timed out",
        )

        executed = "\n".join(cur.queries)
        self.assertIn("ocr_status = 'pending'", executed)
        self.assertIn("ocr_attempts = GREATEST(ocr_attempts - 1, 0)", executed)
        self.assertIn("NULLIF(BTRIM(r.email), '') IS NOT NULL", executed)
        update_params = cur.params_history[1]
        self.assertEqual(update_params[1], restaurant_id)
        self.assertEqual(update_params[2], claim_id)

    def test_verified_status_syncs_demo_ready_lifecycle(self):
        cur = FakeCursor()
        restaurant_id = uuid.uuid4()
        claim_id = uuid.uuid4()

        mark_ocr_status(
            cur,
            restaurant_id,
            "verified",
            claim_id=claim_id,
            claim_fingerprint="fingerprint",
        )

        executed = "\n".join(cur.queries)
        self.assertIn("ocr_status = %s", executed)
        self.assertIn("SET status = CASE WHEN eligibility.eligible THEN 'demo_ready' ELSE 'lead' END", executed)
        self.assertIn("NULLIF(BTRIM(r.email), '') IS NOT NULL", executed)


if __name__ == "__main__":
    unittest.main()
