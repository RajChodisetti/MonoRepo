"""
Restaurant identity hashing for scrape deduplication.

Identity = normalized name + rounded lat/lng (NOT search rank or list position).
"""

from __future__ import annotations

import hashlib
import json
import re
from typing import Any

COORD_DECIMALS = 5  # ~1.1m precision — avoids float noise


def normalize_name(name: str) -> str:
    """Lowercase, trim, collapse whitespace, strip most punctuation."""
    if not name:
        return ""
    text = name.strip().lower()
    text = re.sub(r"[^\w\s&'-]", "", text, flags=re.UNICODE)
    text = re.sub(r"\s+", " ", text).strip()
    return text


def round_coord(value: float | int | None) -> str:
    if value is None:
        return ""
    try:
        return f"{round(float(value), COORD_DECIMALS):.{COORD_DECIMALS}f}"
    except (TypeError, ValueError):
        return ""


def compute_identity_hash(
    name: str,
    latitude: float | int | None,
    longitude: float | int | None,
) -> str:
    """
    SHA-256 of normalized name|lat|lng.
    Same restaurant at same location always yields same hash regardless of search rank.
    """
    norm_name = normalize_name(name)
    lat = round_coord(latitude)
    lng = round_coord(longitude)
    if not norm_name or not lat or not lng:
        return ""
    key = f"{norm_name}|{lat}|{lng}"
    return hashlib.sha256(key.encode("utf-8")).hexdigest()


def identity_hash_from_record(record: dict) -> str:
    """Compute identity hash from a scraped restaurant record."""
    name = (record.get("name") or "").strip()
    coords = (record.get("location") or {}).get("coordinates") or {}
    return compute_identity_hash(
        name,
        coords.get("latitude"),
        coords.get("longitude"),
    )


def identity_hash_from_search(name: str, search_hit: dict) -> str:
    """Compute identity hash from Places Text Search result (before Details call)."""
    geo = (search_hit.get("geometry") or {}).get("location") or {}
    display_name = (search_hit.get("name") or name or "").strip()
    return compute_identity_hash(
        display_name,
        geo.get("lat"),
        geo.get("lng"),
    )


def content_hash(record: dict) -> str:
    """
    Hash of scraped payload for change detection on refresh.
    Excludes volatile fields like scrape timestamps and discovery rank.
    """
    stable = {
        "name": record.get("name"),
        "rating": record.get("rating"),
        "reviews_count": record.get("reviews_count"),
        "contact": record.get("contact"),
        "location": record.get("location"),
        "hours": record.get("hours"),
        "price_level": record.get("price_level"),
        "cuisines": record.get("cuisines"),
        "google": record.get("google"),
    }
    payload = json.dumps(stable, sort_keys=True, ensure_ascii=False, default=str)
    return hashlib.sha256(payload.encode("utf-8")).hexdigest()


def coords_from_search_hit(search_hit: dict) -> dict[str, Any]:
    geo = (search_hit.get("geometry") or {}).get("location") or {}
    return {
        "latitude": geo.get("lat"),
        "longitude": geo.get("lng"),
    }
