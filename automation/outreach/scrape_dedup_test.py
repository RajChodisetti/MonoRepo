"""Tests for scrape pre-dedup (no HTTP)."""

import unittest
from unittest.mock import MagicMock, patch

from known_leads import KnownLeadsRegistry
from scrape_restaurant_places import run_places_scrape_pipeline


class ScrapeDedupTest(unittest.TestCase):
    @patch("scrape_restaurant_places.scrape_single_restaurant_places")
    @patch("scrape_restaurant_places.get_places_api_key")
    def test_skips_known_apollo_with_place_id(self, mock_key, mock_scrape):
        mock_key.return_value = "fake-key"
        known = KnownLeadsRegistry()
        known.apollo_ids.add("apollo-99")
        known.apollo_id_to_place_id["apollo-99"] = "ChIJknown"

        lead = {
            "source": {"person_id": "apollo-99"},
            "company": {"name": "Known Cafe"},
            "contact": {"email": "owner@known.com"},
            "location": {"city": "Sydney"},
        }

        records, _ = run_places_scrape_pipeline(
            city="Sydney",
            niche="restaurant",
            known=known,
            leads_data=[lead],
            output_path="/tmp/scrape_dedup_test.json",
        )

        mock_scrape.assert_not_called()
        self.assertEqual(records, [])


if __name__ == "__main__":
    unittest.main()
