"""Fetch and resolve Google Places photos for immediate OCR use.

Google documents photo resource names as expirable, so OCR refreshes them from
Place Details immediately before resolving Photo Media. Neither resource names
nor the short-lived ``photoUri`` are persisted to PostgreSQL or lead JSON.
"""

from __future__ import annotations

import re
import time
from urllib.parse import quote, urlparse

import requests

from tuvi_outreach_agent import Config


PHOTO_RESOURCE_RE = re.compile(r"^places/[^/?#]+/photos/[^/?#]+$")


class PhotoRequestTransientError(RuntimeError):
    """A timeout, rate limit, or temporary Places failure safe to retry later."""


def _validated_photo_resources(photos: list[dict]) -> list[dict]:
    """Return unique, validated photo metadata from a fresh Place response."""
    resources: list[dict] = []
    seen: set[str] = set()
    for photo in photos:
        if not isinstance(photo, dict):
            continue
        name = str(photo.get("name") or "").strip()
        if not PHOTO_RESOURCE_RE.fullmatch(name) or name in seen:
            continue
        seen.add(name)
        resources.append({"name": name})
    return resources


def fetch_fresh_google_photo_resources(place_id: str, cfg: Config) -> list[dict]:
    """Fetch current photo resource names from Place Details just before OCR."""
    place_id = str(place_id or "").strip()
    if not place_id:
        return []

    api_key = str(getattr(cfg, "PLACES_API_KEY", "") or "").strip()
    if not api_key:
        raise RuntimeError("Google Places API key is not configured for OCR photo refresh")

    base_url = str(
        getattr(cfg, "PLACES_API_BASE_URL", "https://places.googleapis.com/v1")
        or "https://places.googleapis.com/v1"
    ).rstrip("/")
    url = f"{base_url}/places/{quote(place_id, safe='')}"
    headers = {
        "Accept": "application/json",
        "X-Goog-Api-Key": api_key,
        "X-Goog-FieldMask": "photos",
    }
    attempts = max(1, int(getattr(cfg, "RETRY_ATTEMPTS", 3) or 3))
    backoff = max(1.0, float(getattr(cfg, "RETRY_BACKOFF", 2.0) or 2.0))
    for attempt in range(1, attempts + 1):
        try:
            response = requests.get(
                url,
                headers=headers,
                timeout=max(int(getattr(cfg, "SCRAPE_TIMEOUT", 20) or 20), 20),
                allow_redirects=False,
            )
        except requests.RequestException as exc:
            if attempt >= attempts:
                raise PhotoRequestTransientError(
                    "Google Places photo resource refresh timed out or was unavailable"
                ) from exc
        else:
            if 200 <= response.status_code < 300:
                try:
                    payload = response.json() or {}
                except ValueError as exc:
                    raise RuntimeError("Google Places photo resource refresh returned invalid JSON") from exc
                return _validated_photo_resources(payload.get("photos") or [])

            if response.status_code < 500 and response.status_code not in (408, 429):
                raise RuntimeError(
                    f"Google Places photo resource refresh HTTP {response.status_code}"
                )
            if attempt >= attempts:
                raise PhotoRequestTransientError(
                    f"Google Places photo resource refresh HTTP {response.status_code}"
                )

        time.sleep(backoff ** (attempt - 1))

    raise PhotoRequestTransientError("Google Places photo resource refresh failed")


def resolve_google_photo_uri(
    photo_name: str,
    cfg: Config,
    *,
    max_width_px: int = 1600,
) -> str:
    """Resolve one photo resource to a short-lived HTTPS media URI."""
    photo_name = str(photo_name or "").strip()
    if not PHOTO_RESOURCE_RE.fullmatch(photo_name):
        raise ValueError("Invalid Google Places photo resource name")

    api_key = str(getattr(cfg, "PLACES_API_KEY", "") or "").strip()
    if not api_key:
        raise RuntimeError("Google Places API key is not configured for OCR photo resolution")

    width = max(1, min(int(max_width_px), 4800))
    base_url = str(
        getattr(cfg, "PLACES_API_BASE_URL", "https://places.googleapis.com/v1")
        or "https://places.googleapis.com/v1"
    ).rstrip("/")
    url = f"{base_url}/{photo_name}/media"
    headers = {
        "Accept": "application/json",
        "X-Goog-Api-Key": api_key,
    }
    params = {
        "maxWidthPx": width,
        "skipHttpRedirect": "true",
    }

    attempts = max(1, int(getattr(cfg, "RETRY_ATTEMPTS", 3) or 3))
    backoff = max(1.0, float(getattr(cfg, "RETRY_BACKOFF", 2.0) or 2.0))
    for attempt in range(1, attempts + 1):
        try:
            response = requests.get(
                url,
                headers=headers,
                params=params,
                timeout=max(int(getattr(cfg, "SCRAPE_TIMEOUT", 20) or 20), 20),
                allow_redirects=False,
            )
        except requests.RequestException as exc:
            if attempt >= attempts:
                raise PhotoRequestTransientError(
                    "Google Places photo media request timed out or was unavailable"
                ) from exc
        else:
            if 200 <= response.status_code < 300:
                try:
                    photo_uri = str((response.json() or {}).get("photoUri") or "").strip()
                except ValueError as exc:
                    raise RuntimeError("Google Places photo media returned invalid JSON") from exc
                parsed = urlparse(photo_uri)
                if parsed.scheme != "https" or not parsed.netloc:
                    raise RuntimeError("Google Places photo media returned an invalid URI")
                return photo_uri

            if response.status_code < 500 and response.status_code not in (408, 429):
                raise RuntimeError(f"Google Places photo media HTTP {response.status_code}")
            if attempt >= attempts:
                raise PhotoRequestTransientError(
                    f"Google Places photo media HTTP {response.status_code}"
                )

        time.sleep(backoff ** (attempt - 1))

    raise PhotoRequestTransientError("Google Places photo media request failed")
