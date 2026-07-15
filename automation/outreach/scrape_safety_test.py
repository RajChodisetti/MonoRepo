"""Tests for credential-bearing URL removal."""

import unittest

from scrape_safety import has_embedded_google_api_key, sanitize_sensitive_urls


class ScrapeSafetyTest(unittest.TestCase):
    def test_removes_google_photo_url_with_api_key(self):
        unsafe = "https://maps.googleapis.com/maps/api/place/photo?maxwidth=1200&key=secret"
        record = {"images": {"thumbnail": unsafe, "gallery": [{"url": unsafe}]}}
        sanitized = sanitize_sensitive_urls(record)
        self.assertEqual(sanitized["images"]["thumbnail"], "")
        self.assertEqual(sanitized["images"]["gallery"][0]["url"], "")

    def test_keeps_noncredential_public_url(self):
        safe = "https://images.example.com/photo.jpg"
        self.assertFalse(has_embedded_google_api_key(safe))
        self.assertEqual(sanitize_sensitive_urls(safe), safe)


if __name__ == "__main__":
    unittest.main()
