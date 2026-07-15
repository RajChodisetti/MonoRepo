"""Tests for niche_config."""

import unittest

from niche_config import RESTAURANT_KEYWORD_TAGS, get_niche, list_niche_types


class NicheConfigTest(unittest.TestCase):
    def test_restaurant_is_default(self):
        niche = get_niche("restaurant")
        self.assertEqual(niche.slug, "restaurant")
        self.assertEqual(niche.keyword_tags, RESTAURANT_KEYWORD_TAGS)
        self.assertEqual(niche.places_query_suffix, "restaurant")

    def test_unknown_type_raises(self):
        with self.assertRaises(ValueError):
            get_niche("lawyer")

    def test_list_includes_beta_niches(self):
        types = list_niche_types()
        self.assertIn("restaurant", types)
        self.assertIn("dentist", types)
        self.assertIn("plumber", types)


if __name__ == "__main__":
    unittest.main()
