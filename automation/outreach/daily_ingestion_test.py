"""Tests for Places-first daily ingestion with Apollo contact enrichment."""

import os
import unittest
from unittest.mock import patch

from daily_ingestion import run_daily_ingestion
from known_leads import KnownLeadsRegistry
from tuvi_outreach_agent import Config


class DailyIngestionTest(unittest.TestCase):
    @patch("daily_ingestion.save_state")
    @patch("daily_ingestion.record_run_summary")
    @patch("daily_ingestion.merge_scrape_file", return_value=1)
    @patch("daily_ingestion.enrich_missing_contact_with_apollo")
    @patch("daily_ingestion.scrape_single_restaurant_places")
    @patch("daily_ingestion.discover_places_for_city")
    @patch("daily_ingestion.KnownLeadsRegistry.load_combined")
    @patch("daily_ingestion._refuse_when_durable_city_pipeline_is_installed")
    def test_runs_places_then_apollo_contact_enrichment(
        self,
        refuse_legacy,
        load_known,
        discover,
        scrape,
        apollo_enrich,
        _merge,
        _record_summary,
        _save_state,
    ):
        cfg = Config()
        cfg.APOLLO_API_KEY = "apollo-key"
        cfg.APOLLO_ENRICHMENT_ENABLED = True
        cfg.PLACES_API_KEY = "places-key"
        cfg.SCRAPE_DELAY = 0
        load_known.return_value = KnownLeadsRegistry()
        discover.return_value = [{"id": "place-1", "displayName": {"text": "Cafe"}}]
        scrape.return_value = {
            "name": "Cafe",
            "google": {"place_id": "place-1"},
            "location": {"city": "Melbourne"},
            "contact": {"website": "https://cafe.example", "email": ""},
            "owners": [],
            "scrape_status": "success",
        }

        def enrich_after_places(*_args, **_kwargs):
            self.assertTrue(scrape.called)
            return (
                {
                    **scrape.return_value,
                    "contact": {
                        "website": "https://cafe.example",
                        "email": "owner@cafe.example",
                    },
                    "owners": ["Owner Name"],
                    "apollo_lead": {"source": {"person_id": "person-1"}},
                },
                {
                    "status": "enriched",
                    "owner_added": True,
                    "email_added": True,
                },
            )

        apollo_enrich.side_effect = enrich_after_places

        with patch.dict(os.environ, {}, clear=False):
            summary = run_daily_ingestion(
                cities=["Melbourne, Australia"],
                target_per_city=1,
                max_requests=5,
                import_to_db=False,
                cfg=cfg,
            )

        self.assertEqual(
            summary["source"],
            "google_places_api_new+apollo_contact_enrichment",
        )
        self.assertEqual(summary["places_discovered"], 1)
        self.assertEqual(summary["scrape_success"], 1)
        self.assertEqual(summary["apollo_enriched"], 1)
        self.assertEqual(summary["apollo_owner_filled"], 1)
        self.assertEqual(summary["apollo_email_filled"], 1)
        discover.assert_called_once()
        apollo_enrich.assert_called_once()
        self.assertTrue(scrape.called)
        load_known.assert_called_once_with("restaurant", require_database=False)
        refuse_legacy.assert_called_once_with()

    @patch("daily_ingestion._refuse_when_durable_city_pipeline_is_installed")
    def test_requires_apollo_key_when_contact_enrichment_is_enabled(self, refuse_legacy):
        cfg = Config()
        cfg.PLACES_API_KEY = "places-key"
        cfg.APOLLO_API_KEY = ""
        cfg.APOLLO_ENRICHMENT_ENABLED = True

        with self.assertRaisesRegex(ValueError, "APOLLO_API_KEY"):
            run_daily_ingestion(
                cities=["Melbourne, Australia"],
                target_per_city=1,
                max_requests=5,
                import_to_db=False,
                cfg=cfg,
            )
        refuse_legacy.assert_called_once_with()


if __name__ == "__main__":
    unittest.main()
