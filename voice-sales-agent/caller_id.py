"""Resolve outbound caller ID for restaurant template calls."""

from __future__ import annotations

import os

import phonenumbers


def _normalize_e164(number: str, default_region: str = "AU") -> str | None:
    try:
        parsed = phonenumbers.parse(number, default_region)
        if phonenumbers.is_valid_number(parsed):
            return phonenumbers.format_number(parsed, phonenumbers.PhoneNumberFormat.E164)
    except Exception:
        pass
    return None


def _verified_caller_ids() -> set[str]:
    raw = os.environ.get("TWILIO_VERIFIED_CALLER_IDS", "")
    out: set[str] = set()
    for part in raw.split(","):
        token = part.strip()
        if not token:
            continue
        e164 = _normalize_e164(token)
        if e164:
            out.add(e164)
    return out


def resolve_caller_id(restaurant_phone: str | None) -> tuple[str, bool]:
    """
    Pick Twilio `from` number for outbound restaurant calls.

    Uses the scraped restaurant phone when it appears in TWILIO_VERIFIED_CALLER_IDS;
    otherwise falls back to TWILIO_PHONE_NUMBER.
    """
    fallback = os.environ.get("TWILIO_PHONE_NUMBER", "").strip()
    if not fallback:
        raise RuntimeError("TWILIO_PHONE_NUMBER is not configured")

    if not restaurant_phone:
        return fallback, False

    candidate = _normalize_e164(restaurant_phone)
    if candidate and candidate in _verified_caller_ids():
        return candidate, True

    return fallback, False
