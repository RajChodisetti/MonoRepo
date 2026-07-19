"""Query-contract tests for email-only OCR claiming and safe claim release."""

import uuid
import unittest

from import_to_db import fetch_unverified_leads, release_ocr_claim


class FakeCursor:
    def __init__(self):
        self.query = ""
        self.params = ()
        self.rowcount = 1

    def execute(self, query, params=()):
        self.query = query
        self.params = params

    def fetchall(self):
        return []


class OCRClaimQueryTest(unittest.TestCase):
    def test_preview_selects_only_restaurants_with_email(self):
        cur = FakeCursor()

        rows = fetch_unverified_leads(cur, 10, claim=False)

        self.assertEqual(rows, [])
        self.assertIn("NULLIF(BTRIM(r.email), '') IS NOT NULL", cur.query)

    def test_claim_selects_only_restaurants_with_email(self):
        cur = FakeCursor()

        rows = fetch_unverified_leads(cur, 10, claim=True)

        self.assertEqual(rows, [])
        self.assertIn("NULLIF(BTRIM(r.email), '') IS NOT NULL", cur.query)

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

        self.assertIn("ocr_status = 'pending'", cur.query)
        self.assertIn("ocr_attempts = GREATEST(ocr_attempts - 1, 0)", cur.query)
        self.assertEqual(cur.params[1], restaurant_id)
        self.assertEqual(cur.params[2], claim_id)


if __name__ == "__main__":
    unittest.main()
