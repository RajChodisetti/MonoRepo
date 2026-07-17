"""Unit tests for the durable all-scraped-photos OCR completion gate."""

import unittest

from menu_image_ocr import all_scraped_photos_processed


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


if __name__ == "__main__":
    unittest.main()
