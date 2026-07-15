"""Tests for Google Places API (New) discovery and safe enrichment."""

import unittest
from unittest.mock import MagicMock, patch

from google_places_scraper import (
    _is_public_http_url,
    _places_v1_request,
    discover_places_for_city,
    place_to_lead_dict,
    scrape_single_restaurant_places,
)
from known_leads import KnownLeadsRegistry
from request_budget import RequestBudget
from tuvi_outreach_agent import Config


def test_config() -> Config:
    cfg = Config()
    cfg.PLACES_API_KEY = "test-secret-key"
    cfg.PLACES_API_BASE_URL = "https://places.example.test/v1"
    cfg.RETRY_ATTEMPTS = 1
    cfg.SCRAPE_DELAY = 0
    return cfg


class PlacesRequestTest(unittest.TestCase):
    @patch("google_places_scraper.requests.request")
    def test_api_key_is_sent_in_header_not_url(self, request):
        response = MagicMock()
        response.status_code = 200
        response.content = b"{}"
        response.json.return_value = {}
        request.return_value = response

        _places_v1_request(
            "GET",
            "places/place-1",
            test_config(),
            field_mask="id",
            label="test",
        )

        _, url = request.call_args.args[:2]
        headers = request.call_args.kwargs["headers"]
        self.assertNotIn("test-secret-key", url)
        self.assertEqual(headers["X-Goog-Api-Key"], "test-secret-key")
        self.assertEqual(headers["X-Goog-FieldMask"], "id")


class PlacesDiscoveryTest(unittest.TestCase):
    @patch("google_places_scraper._places_v1_request")
    def test_discovers_melbourne_places_and_skips_known_ids(self, places_request):
        places_request.side_effect = [
            {
                "places": [
                    {"id": "known", "displayName": {"text": "Known"}},
                    {"id": "new-1", "displayName": {"text": "New One"}},
                ],
                "nextPageToken": "page-2",
            },
            {
                "places": [
                    {"id": "new-1", "displayName": {"text": "Duplicate"}},
                    {"id": "new-2", "displayName": {"text": "New Two"}},
                ],
            },
        ]
        known = KnownLeadsRegistry(place_ids={"known"})
        budget = RequestBudget(10)

        places = discover_places_for_city(
            city="Melbourne, Australia",
            limit=3,
            cfg=test_config(),
            budget=budget,
            known=known,
        )

        self.assertEqual([place["id"] for place in places], ["new-1", "new-2"])
        self.assertEqual(budget.total_used, 0)  # mocked request owns budget consumption
        second_body = places_request.call_args_list[1].kwargs["body"]
        self.assertEqual(second_body["pageToken"], "page-2")
        bounds = places_request.call_args_list[0].kwargs["body"]["locationRestriction"]["rectangle"]
        self.assertLess(bounds["low"]["latitude"], bounds["high"]["latitude"])

    def test_place_discovery_record_becomes_place_id_lead(self):
        lead = place_to_lead_dict(
            {
                "id": "place-1",
                "displayName": {"text": "Melbourne Cafe"},
                "location": {"latitude": -37.8, "longitude": 144.9},
            },
            "Melbourne, Australia",
        )
        self.assertEqual(lead["google"]["place_id"], "place-1")
        self.assertEqual(lead["source"]["provider"], "google_places_api_new")
        self.assertNotIn("apollo_id", lead.get("extra") or {})


class PlacesEnrichmentTest(unittest.TestCase):
    @patch("google_places_scraper.find_email_from_website", return_value="hello@cafe.test")
    @patch("google_places_scraper.get_place_details")
    def test_expirable_photo_resource_names_are_not_stored(self, details, _email):
        details.return_value = {
            "id": "place-1",
            "displayName": {"text": "Melbourne Cafe"},
            "formattedAddress": "1 Test St, Melbourne VIC, Australia",
            "websiteUri": "https://cafe.test",
            "location": {"latitude": -37.8, "longitude": 144.9},
            "photos": [{"name": "places/place-1/photos/photo-1", "widthPx": 1200, "heightPx": 800}],
        }
        lead = place_to_lead_dict(
            {"id": "place-1", "displayName": {"text": "Melbourne Cafe"}},
            "Melbourne, Australia",
        )

        result = scrape_single_restaurant_places(lead, test_config())

        self.assertEqual(result["scrape_status"], "success")
        self.assertEqual(result["contact"]["email"], "hello@cafe.test")
        self.assertEqual(result["images"]["gallery"], [])
        self.assertEqual(result["images"]["google_photos"], [])
        self.assertEqual(result["images"]["google_photo_count"], 1)


class WebsiteSafetyTest(unittest.TestCase):
    @patch("google_places_scraper.socket.getaddrinfo")
    def test_private_network_website_is_blocked(self, getaddrinfo):
        getaddrinfo.return_value = [(2, 1, 6, "", ("127.0.0.1", 80))]
        self.assertFalse(_is_public_http_url("http://localhost/contact"))

    @patch("google_places_scraper.socket.getaddrinfo")
    def test_public_website_is_allowed(self, getaddrinfo):
        getaddrinfo.return_value = [(2, 1, 6, "", ("93.184.216.34", 443))]
        self.assertTrue(_is_public_http_url("https://example.com/contact"))


if __name__ == "__main__":
    unittest.main()
