"""Safety helpers shared by the public-data scraping and import paths."""

from __future__ import annotations

from copy import deepcopy
from typing import Any
from urllib.parse import parse_qsl, urlsplit


_GOOGLE_API_HOSTS = {
    "maps.googleapis.com",
    "places.googleapis.com",
}
_SECRET_QUERY_KEYS = {"key", "api_key", "apikey"}


def has_embedded_google_api_key(value: str) -> bool:
    """Return True when a Google API URL contains a credential query value."""
    raw = (value or "").strip()
    if not raw:
        return False
    try:
        parsed = urlsplit(raw)
    except ValueError:
        return False
    host = (parsed.hostname or "").lower()
    if host not in _GOOGLE_API_HOSTS:
        return False
    return any(key.lower() in _SECRET_QUERY_KEYS for key, _ in parse_qsl(parsed.query))


def sanitize_sensitive_urls(value: Any) -> Any:
    """Deep-copy JSON-compatible data while removing credential-bearing URLs."""
    if isinstance(value, dict):
        return {key: sanitize_sensitive_urls(item) for key, item in value.items()}
    if isinstance(value, list):
        return [sanitize_sensitive_urls(item) for item in value]
    if isinstance(value, tuple):
        return tuple(sanitize_sensitive_urls(item) for item in value)
    if isinstance(value, str) and has_embedded_google_api_key(value):
        return ""
    return deepcopy(value)
