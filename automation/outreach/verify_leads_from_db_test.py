"""Unit tests for the durable all-scraped-photos OCR completion gate."""

import unittest
from unittest.mock import patch

from menu_image_ocr import all_scraped_photos_processed
from media_asset_metadata import (
    media_asset_public_url,
    recommended_placement,
    website_media_type,
)


class OCRCompletionGateTest(unittest.TestCase):
    def test_requires_every_discovered_photo_to_succeed(self):
        summary = {
            "images_analyzed": 10,
            "images_succeeded": 9,
            "images_failed": 1,
        }

        self.assertFalse(all_scraped_photos_processed(summary, 10, 0))

    def test_rejects_unresolved_photo_even_if_every_resolved_photo_succeeds(self):
        summary = {
            "images_analyzed": 9,
            "images_succeeded": 9,
            "images_failed": 0,
        }

        self.assertFalse(all_scraped_photos_processed(summary, 10, 1))

    def test_accepts_only_exact_full_success(self):
        summary = {
            "images_analyzed": 10,
            "images_succeeded": 10,
            "images_failed": 0,
        }

        self.assertTrue(all_scraped_photos_processed(summary, 10, 0))

    def test_zero_images_is_not_verified(self):
        self.assertFalse(
            all_scraped_photos_processed(
                {
                    "images_analyzed": 0,
                    "images_succeeded": 0,
                    "images_failed": 0,
                },
                0,
                0,
            )
        )

    def test_owned_media_type_and_placement_are_template_ready(self):
        self.assertEqual(website_media_type("food_photo", "other"), "food")
        self.assertEqual(
            recommended_placement(
                {
                    "image_type": "exterior",
                    "orientation": "landscape",
                    "hero_score": 0.9,
                },
                "gallery",
            ),
            "hero",
        )

    @patch.dict(
        "os.environ",
        {"STORAGE_PUBLIC_BASE_URL": "https://cdn.example.test/media"},
        clear=False,
    )
    def test_owned_media_public_url_escapes_object_key(self):
        self.assertEqual(
            media_asset_public_url("restaurants/a/my image.jpg"),
            "https://cdn.example.test/media/restaurants/a/my%20image.jpg",
        )


if __name__ == "__main__":
    unittest.main()
