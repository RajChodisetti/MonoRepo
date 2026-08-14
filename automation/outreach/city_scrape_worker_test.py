"""Tests for best-effort Apollo behavior in the durable city worker."""

import unittest
from unittest.mock import MagicMock, patch

from city_scrape_worker import CityScrapeWorker, _apollo_available
from niche_config import get_niche
from request_budget import BudgetExhausted, RequestBudget
from tuvi_outreach_agent import Config


def test_config() -> Config:
    cfg = Config()
    cfg.APOLLO_ENRICHMENT_ENABLED = True
    cfg.APOLLO_API_KEY = "apollo-test-key"
    return cfg


def details_ready_candidate() -> dict:
    return {
        "id": "00000000-0000-4000-8000-000000000001",
        "status": "details_ready",
        "scrape_record": {
            "name": "Example Cafe",
            "google_place_id": "place-1",
            "scrape_status": "success",
            "contact": {"website": "https://cafe.example", "email": ""},
            "owners": [],
        },
    }


class CityScrapeWorkerApolloTest(unittest.TestCase):
    def worker(self) -> CityScrapeWorker:
        worker = CityScrapeWorker.__new__(CityScrapeWorker)
        worker.store = MagicMock()
        worker.cfg = test_config()
        return worker

    def test_missing_apollo_key_disables_only_optional_enrichment(self):
        cfg = test_config()
        cfg.APOLLO_API_KEY = ""

        with self.assertLogs("city_scrape_worker", level="WARNING") as captured:
            available = _apollo_available(cfg, "job-1")

        self.assertFalse(available)
        self.assertIn("continuing with Google Places only", " ".join(captured.output))

    @patch("city_scrape_worker.enrich_missing_contact_with_apollo")
    def test_apollo_provider_error_imports_verified_places_record(self, enrich):
        worker = self.worker()
        candidate = details_ready_candidate()
        enriched = dict(candidate["scrape_record"])
        enrich.return_value = (
            enriched,
            {
                "status": "error",
                "error": "Apollo API HTTP 403",
                "error_code": 403,
                "search_requests": 1,
                "match_requests": 0,
                "owner_added": False,
                "email_added": False,
            },
        )

        worker._process_candidate(
            {"id": "job-1", "city": "Melbourne"},
            candidate,
            get_niche("restaurant"),
            RequestBudget(10),
            True,
        )

        saved_record = worker.store.save_candidate.call_args.args[2]
        self.assertEqual(worker.store.save_candidate.call_args.args[1], "enriched")
        self.assertEqual(saved_record["apollo_enrichment"]["status"], "error")
        worker.store.import_candidate.assert_called_once_with(
            candidate["id"], "job-1", saved_record
        )

    @patch("city_scrape_worker.enrich_missing_contact_with_apollo")
    def test_unexpected_apollo_exception_still_imports_places_record(self, enrich):
        worker = self.worker()
        candidate = details_ready_candidate()
        enrich.side_effect = RuntimeError("adapter failure")

        worker._process_candidate(
            {"id": "job-1", "city": "Melbourne"},
            candidate,
            get_niche("restaurant"),
            RequestBudget(10),
            True,
        )

        saved_record = worker.store.save_candidate.call_args.args[2]
        self.assertEqual(saved_record["apollo_enrichment"]["status"], "error")
        self.assertNotIn("adapter failure", saved_record["apollo_enrichment"]["error"])
        worker.store.import_candidate.assert_called_once()

    @patch("city_scrape_worker.enrich_missing_contact_with_apollo")
    def test_combined_request_limit_still_checkpoints_before_pause(self, enrich):
        worker = self.worker()
        candidate = details_ready_candidate()
        enrich.return_value = (
            dict(candidate["scrape_record"]),
            {
                "status": "budget_exhausted",
                "search_requests": 1,
                "match_requests": 0,
                "owner_added": False,
                "email_added": False,
            },
        )

        with self.assertRaises(BudgetExhausted):
            worker._process_candidate(
                {"id": "job-1", "city": "Melbourne"},
                candidate,
                get_niche("restaurant"),
                RequestBudget(10),
                True,
            )

        self.assertEqual(worker.store.save_candidate.call_args.args[1], "details_ready")
        worker.store.import_candidate.assert_not_called()


if __name__ == "__main__":
    unittest.main()
