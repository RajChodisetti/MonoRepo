"""Pure helpers for durable restaurant media URL and placement metadata."""

from __future__ import annotations

import os
from urllib.parse import quote


WEBSITE_MEDIA_TYPES = frozenset({
    "exterior", "interior", "food", "drink", "logo", "team", "event", "other"
})


def media_asset_public_url(storage_key: str) -> str:
    base = os.getenv("STORAGE_PUBLIC_BASE_URL", "").strip().rstrip("/")
    key = str(storage_key or "").strip().lstrip("/")
    if not base or not key:
        return ""
    return f"{base}/{quote(key, safe='/')}"


def website_media_type(image_type: str, fallback: str) -> str:
    normalized = str(image_type or "").strip().lower()
    if normalized == "food_photo":
        return "food"
    if normalized in WEBSITE_MEDIA_TYPES:
        return normalized
    return fallback if fallback in WEBSITE_MEDIA_TYPES else "other"


def recommended_placement(result: dict, current: str) -> str:
    image_type = website_media_type(result.get("image_type"), "other")
    orientation = str(result.get("orientation") or "unknown").lower()
    try:
        hero_score = float(result.get("hero_score") or 0)
    except (TypeError, ValueError):
        hero_score = 0
    if hero_score >= 0.78 and orientation == "landscape" and image_type in {"exterior", "interior"}:
        return "hero"
    if image_type in {"exterior", "interior"}:
        return "ambience_gallery"
    if image_type in {"food", "drink"}:
        return "food_gallery"
    if image_type == "logo":
        return "logo"
    return current if current in {
        "hero", "about", "gallery", "food_gallery", "ambience_gallery", "logo"
    } else "gallery"
