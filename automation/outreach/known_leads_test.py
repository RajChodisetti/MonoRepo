"""Tests for known_leads and ingestion_state."""

import json
import tempfile
import unittest
from pathlib import Path

from ingestion_state import get_apollo_page, load_state, save_state, set_apollo_page
from known_leads import KnownLeadsRegistry


class KnownLeadsTest(unittest.TestCase):
    def test_skip_scrape_when_apollo_has_place_id(self):
        registry = KnownLeadsRegistry()
        registry.apollo_ids.add("apollo-1")
        registry.apollo_id_to_place_id["apollo-1"] = "place-abc"
        lead = {
            "source": {"person_id": "apollo-1"},
            "company": {"name": "Test Cafe"},
            "contact": {"email": "a@example.com"},
        }
        skip, reason = registry.should_skip_scrape(lead)
        self.assertTrue(skip)
        self.assertEqual(reason, "known_apollo_with_place_id")

    def test_load_from_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            leads_dir = Path(tmp) / "leads"
            data_dir = Path(tmp) / "data"
            leads_dir.mkdir()
            data_dir.mkdir()
            (leads_dir / "lead_sydney.json").write_text(
                json.dumps({
                    "leads": [{
                        "source": {"person_id": "p1"},
                        "company": {"name": "Cafe", "domain": "cafe.com"},
                        "contact": {"email": "owner@cafe.com", "name": "Owner"},
                        "location": {"city": "Sydney"},
                    }],
                }),
                encoding="utf-8",
            )
            (data_dir / "restaurants_data_sydney.json").write_text(
                json.dumps({
                    "restaurants": [{
                        "name": "Cafe",
                        "google": {"place_id": "place-xyz"},
                        "apollo_lead": {"source": {"person_id": "p1"}},
                        "location": {"city": "Sydney"},
                    }],
                }),
                encoding="utf-8",
            )
            registry = KnownLeadsRegistry.load_from_files(
                leads_dir=leads_dir,
                data_dir=data_dir,
                niche_slug="restaurant",
            )
            self.assertIn("place-xyz", registry.place_ids)
            self.assertIn("p1", registry.apollo_ids)


class IngestionStateTest(unittest.TestCase):
    def test_apollo_page_roundtrip(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "state.json"
            state = load_state(path)
            self.assertEqual(get_apollo_page(state, "restaurant", "Sydney, Australia"), 1)
            set_apollo_page(state, "restaurant", "Sydney, Australia", 4)
            save_state(state, path)
            reloaded = load_state(path)
            self.assertEqual(get_apollo_page(reloaded, "restaurant", "Sydney, Australia"), 4)


if __name__ == "__main__":
    unittest.main()
