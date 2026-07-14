"""Tests for targeted Apollo contact enrichment after Places."""

import unittest
from unittest.mock import MagicMock, patch

from apollo_enrichment import enrich_missing_contact_with_apollo
from niche_config import get_niche
from request_budget import RequestBudget
from tuvi_outreach_agent import Config


def test_config() -> Config:
    cfg = Config()
    cfg.APOLLO_API_KEY = "apollo-test-key"
    cfg.APOLLO_API_BASE_URL = "https://apollo.example.test/api/v1"
    cfg.RETRY_ATTEMPTS = 1
    return cfg


class ApolloEnrichmentTest(unittest.TestCase):
    @patch("apollo_enrichment.requests.request")
    def test_fills_missing_owner_and_work_email_after_domain_search(self, request):
        search_response = MagicMock()
        search_response.status_code = 200
        search_response.content = b"search"
        search_response.json.return_value = {
            "people": [
                {
                    "id": "person-1",
                    "first_name": "Alex",
                    "last_name_obfuscated": "Sm***h",
                    "title": "Owner",
                    "has_email": True,
                }
            ]
        }
        match_response = MagicMock()
        match_response.status_code = 200
        match_response.content = b"match"
        match_response.json.return_value = {
            "person": {
                "id": "person-1",
                "first_name": "Alex",
                "last_name": "Smith",
                "title": "Owner",
                "email": "alex@cafe.example",
                "personal_emails": ["private@example.net"],
                "organization": {"name": "Example Cafe"},
            }
        }
        request.side_effect = [search_response, match_response]
        budget = RequestBudget(5)
        record = {
            "name": "Example Cafe",
            "contact": {"website": "https://www.cafe.example", "email": ""},
            "owners": [],
        }

        enriched, stats = enrich_missing_contact_with_apollo(
            record,
            test_config(),
            get_niche("restaurant"),
            budget=budget,
        )

        self.assertEqual(enriched["owners"], ["Alex Smith"])
        self.assertEqual(enriched["contact"]["email"], "alex@cafe.example")
        self.assertEqual(enriched["apollo_lead"]["source"]["person_id"], "person-1")
        self.assertNotIn("personal_emails", str(enriched["apollo_lead"]))
        self.assertEqual(stats["status"], "enriched")
        self.assertEqual(budget.total_used, 2)

        search_call, match_call = request.call_args_list
        self.assertNotIn("apollo-test-key", search_call.args[1])
        self.assertEqual(search_call.kwargs["headers"]["X-Api-Key"], "apollo-test-key")
        self.assertEqual(
            search_call.kwargs["json"]["q_organization_domains_list"],
            ["cafe.example"],
        )
        self.assertEqual(match_call.kwargs["params"]["reveal_personal_emails"], "false")
        self.assertEqual(match_call.kwargs["params"]["reveal_phone_number"], "false")

    @patch("apollo_enrichment.requests.request")
    def test_does_not_call_apollo_when_owner_and_email_are_present(self, request):
        record = {
            "contact": {"website": "https://cafe.example", "email": "owner@cafe.example"},
            "owners": ["Alex Smith"],
        }

        enriched, stats = enrich_missing_contact_with_apollo(
            record,
            test_config(),
            get_niche("restaurant"),
        )

        self.assertEqual(enriched, record)
        self.assertEqual(stats["status"], "not_needed")
        request.assert_not_called()

    @patch("apollo_enrichment.requests.request")
    def test_skips_ambiguous_business_without_domain(self, request):
        record = {
            "name": "Common Cafe",
            "contact": {
                "website": "https://instagram.com/common-cafe",
                "email": "common-cafe@gmail.com",
            },
            "owners": [],
        }

        _enriched, stats = enrich_missing_contact_with_apollo(
            record,
            test_config(),
            get_niche("restaurant"),
        )

        self.assertEqual(stats["status"], "skipped_no_domain")
        request.assert_not_called()

    @patch("apollo_enrichment.requests.request")
    def test_provider_errors_do_not_expose_response_body(self, request):
        response = MagicMock()
        response.status_code = 403
        response.content = b'{"error":"owner@example.com apollo-test-key"}'
        request.return_value = response
        record = {
            "name": "Example Cafe",
            "contact": {"website": "https://cafe.example", "email": ""},
            "owners": [],
        }

        _enriched, stats = enrich_missing_contact_with_apollo(
            record,
            test_config(),
            get_niche("restaurant"),
        )

        self.assertEqual(stats["status"], "error")
        self.assertEqual(stats["error_code"], 403)
        self.assertNotIn("owner@example.com", stats["error"])
        self.assertNotIn("apollo-test-key", stats["error"])


if __name__ == "__main__":
    unittest.main()
