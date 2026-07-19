"""Unit tests for the menu image OCR candidate and usage contracts."""

import unittest

from menu_image_ocr import (
    MenuOCRConfig,
    OCRTransientError,
    enrich_restaurant_with_menu_ocr,
)


class FakeAnalyzer:
    def __init__(self):
        self.urls: list[str] = []

    def analyze_image_url(self, url: str) -> dict:
        self.urls.append(url)
        return {
            "url": url,
            "image_type": "food_photo",
            "confidence": 0.9,
            "reason": "Prepared food is visible.",
            "menu_items": [],
            "usage": {
                "input_tokens": 10,
                "output_tokens": 5,
                "total_tokens": 15,
            },
        }


class TimeoutAnalyzer:
    def analyze_image_url(self, _url: str) -> dict:
        raise OCRTransientError("provider timed out")


class MenuImageOCRTest(unittest.TestCase):
    def test_database_verifier_can_process_every_candidate_without_config_cap(self):
        analyzer = FakeAnalyzer()
        cfg = MenuOCRConfig(huggingface_api_key="test", max_images=1, delay=0)
        candidates = [
            {
                "analysis_url": f"https://images.example.test/photo-{index}.jpg",
                "persistent_url": "",
                "source": "google_places_photo",
            }
            for index in range(3)
        ]

        record = enrich_restaurant_with_menu_ocr(
            {},
            cfg,
            analyzer,
            candidates,
            process_all_candidates=True,
        )

        self.assertEqual(len(analyzer.urls), 3)
        self.assertEqual(record["menu_ocr"]["images_analyzed"], 3)
        self.assertEqual(record["menu_ocr"]["images_succeeded"], 3)
        self.assertEqual(record["menu_ocr"]["images_failed"], 0)
        self.assertEqual(record["menu_ocr"]["input_tokens"], 30)
        self.assertEqual(record["menu_ocr"]["output_tokens"], 15)
        self.assertEqual(record["menu_ocr"]["total_tokens"], 45)

    def test_manual_ocr_keeps_configured_candidate_cap(self):
        analyzer = FakeAnalyzer()
        cfg = MenuOCRConfig(huggingface_api_key="test", max_images=1, delay=0)
        candidates = [
            {
                "analysis_url": f"https://images.example.test/photo-{index}.jpg",
                "persistent_url": "",
            }
            for index in range(3)
        ]

        record = enrich_restaurant_with_menu_ocr({}, cfg, analyzer, candidates)

        self.assertEqual(len(analyzer.urls), 1)
        self.assertEqual(record["menu_ocr"]["images_analyzed"], 1)

    def test_transient_timeout_is_not_collapsed_into_permanent_image_failure(self):
        cfg = MenuOCRConfig(huggingface_api_key="test", delay=0)
        candidates = [{
            "analysis_url": "https://images.example.test/photo.jpg",
            "persistent_url": "",
            "source": "google_places_photo",
        }]

        with self.assertRaises(OCRTransientError):
            enrich_restaurant_with_menu_ocr(
                {},
                cfg,
                TimeoutAnalyzer(),
                candidates,
                process_all_candidates=True,
            )


if __name__ == "__main__":
    unittest.main()
