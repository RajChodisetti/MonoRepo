"""Regression tests for the OCR-free import and outreach enrollment contract."""

from __future__ import annotations

import uuid
import unittest

from import_to_db import (
    ensure_outreach_enrollment,
    record_inferred_business_evidence,
    sync_menu_items_and_reviews,
    without_scraped_media,
)
from scrape_ledger import without_scraped_media as ledger_without_scraped_media


class RecordingCursor:
    def __init__(self) -> None:
        self.calls: list[tuple[str, tuple | None]] = []
        self._fetchone = None

    def execute(self, query: str, params: tuple | None = None) -> None:
        self.calls.append((" ".join(query.split()), params))
        if "SELECT id FROM menus" in query:
            self._fetchone = (uuid.UUID("00000000-0000-0000-0000-000000000002"),)

    def fetchone(self):
        return self._fetchone


class OutreachImportTest(unittest.TestCase):
    def test_enrollment_uses_idempotent_database_helper(self) -> None:
        cur = RecordingCursor()
        restaurant_id = uuid.UUID("00000000-0000-0000-0000-000000000001")

        ensure_outreach_enrollment(cur, restaurant_id)

        self.assertEqual(len(cur.calls), 1)
        query, params = cur.calls[0]
        self.assertIn("ensure_outreach_sequence_enrollment", query)
        self.assertEqual(params, (restaurant_id,))

    def test_import_records_inferred_business_source_without_overwriting_interest(self) -> None:
        cur = RecordingCursor()
        restaurant_id = uuid.UUID("00000000-0000-0000-0000-000000000001")

        record_inferred_business_evidence(cur, restaurant_id, "google_places_business_contact")

        query, params = cur.calls[0]
        self.assertIn("outreach_consent_basis = 'inferred_business'", query)
        self.assertIn("shown_interest = false", query)
        self.assertIn("'express_interest', 'withdrawn'", query)
        self.assertEqual(
            params,
            (
                "google_places_business_contact",
                "google_places_business_contact",
                restaurant_id,
            ),
        )

    def test_scraped_menu_media_is_not_persisted(self) -> None:
        cur = RecordingCursor()
        restaurant_id = uuid.UUID("00000000-0000-0000-0000-000000000001")
        record = {
            "menu_items": [
                {
                    "name": "Pasta",
                    "description": "House sauce",
                    "images": [{"url": "https://third-party.example/pasta.jpg"}],
                }
            ],
            "reviews": [],
        }

        self.assertEqual(
            sync_menu_items_and_reviews(cur, restaurant_id, record),
            (0, 0),
        )

        menu_insert = next(call for call in cur.calls if "INSERT INTO menu_items" in call[0])
        params = menu_insert[1]
        self.assertIsNotNone(params)
        self.assertEqual(params[6], "")
        self.assertEqual(params[7], "[]")
        self.assertFalse(any("menu_images" in query for query, _ in cur.calls))
        self.assertFalse(any("gallery_images" in query for query, _ in cur.calls))

    def test_raw_public_payload_drops_all_scraped_media(self) -> None:
        original = {
            "images": {
                "thumbnail": "https://third-party.example/hero.jpg",
                "gallery": [{"url": "https://third-party.example/gallery.jpg"}],
                "google_photos": ["places/photo-resource"],
            },
            "reviews": [
                {
                    "review": "Great",
                    "images": ["https://third-party.example/review.jpg"],
                }
            ],
            "menu_items": [
                {
                    "name": "Pasta",
                    "image_url": "https://third-party.example/dish.jpg",
                    "image": "https://third-party.example/dish-2.jpg",
                    "images": [{"url": "https://third-party.example/dish-3.jpg"}],
                }
            ],
        }

        durable = without_scraped_media(original)

        self.assertEqual(durable["images"], {})
        self.assertEqual(durable["reviews"][0]["images"], [])
        self.assertEqual(durable["menu_items"][0]["image_url"], "")
        self.assertEqual(durable["menu_items"][0]["image"], "")
        self.assertEqual(durable["menu_items"][0]["images"], [])
        self.assertIn("hero.jpg", original["images"]["thumbnail"])
        self.assertEqual(ledger_without_scraped_media(original), durable)


if __name__ == "__main__":
    unittest.main()
