"""Unit tests for the menu image OCR candidate and usage contracts."""

import unittest

from menu_image_ocr import (
    MenuImageAnalyzer,
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
            "caption": "A plated entree",
            "alt_text": "Entree on a ceramic plate",
            "tags": ["entree"],
            "quality_score": 0.8,
            "hero_score": 0.6,
            "orientation": "landscape",
            "subject_position": "center",
            "contains_people": False,
            "contains_text": False,
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
    def test_normalizes_template_metadata_without_accepting_unbounded_values(self):
        analyzer = MenuImageAnalyzer.__new__(MenuImageAnalyzer)
        result = analyzer._normalize_analysis(
            {
                "image_type": "exterior",
                "confidence": 1.5,
                "caption": "A bright storefront",
                "alt_text": "Brick restaurant entrance",
                "tags": ["Storefront", "storefront", "patio"],
                "quality_score": 0.8,
                "hero_score": 0.9,
                "orientation": "landscape",
                "subject_position": "left",
                "contains_people": False,
                "contains_text": True,
                "menu_items": [],
            },
            "https://images.example.test/photo.jpg",
        )

        self.assertEqual(result["image_type"], "exterior")
        self.assertEqual(result["confidence"], 1.0)
        self.assertEqual(result["tags"], ["Storefront", "patio"])
        self.assertEqual(result["hero_score"], 0.9)

    def test_database_verifier_can_process_every_candidate_without_config_cap(self):
        analyzer = FakeAnalyzer()
        cfg = MenuOCRConfig(huggingface_api_key="test", max_images=1, delay=0)
        candidates = [
            {
                "analysis_url": f"https://images.example.test/photo-{index}.jpg",
                "persistent_url": "",
                "source": "google_places_photo",
                "google_place_id": "place-123",
                "source_index": index,
                "source_fingerprint": f"{index:064x}",
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
        self.assertNotIn("caption", record["menu_ocr"]["classifications"][0])
        self.assertNotIn("alt_text", record["menu_ocr"]["classifications"][0])
        self.assertEqual(record["menu_ocr"]["classifications"][0]["google_place_id"], "place-123")
        self.assertEqual(record["menu_ocr"]["classifications"][0]["source_index"], 0)
        self.assertEqual(record["menu_ocr"]["classifications"][0]["source_fingerprint"], "0" * 64)
        self.assertTrue(record["menu_ocr"]["classifications"][0]["public_eligible"])
        self.assertNotIn("url", record["menu_ocr"]["classifications"][0])

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
