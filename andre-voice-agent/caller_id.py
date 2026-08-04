"""Resolve outbound Twilio caller ID."""

from __future__ import annotations

import os

import phonenumbers


def _normalize_e164(number: str, default_region: str = "IN") -> str | None:
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


def resolve_caller_id(preferred_phone: str | None = None) -> tuple[str, bool]:
    """
    Pick Twilio `from` number for outbound calls.
    Uses preferred_phone when it is in TWILIO_VERIFIED_CALLER_IDS; else TWILIO_PHONE_NUMBER.
    """
    fallback = os.environ.get("TWILIO_PHONE_NUMBER", "").strip()
    if not fallback:
        raise RuntimeError("TWILIO_PHONE_NUMBER is not configured")

    if not preferred_phone:
        return fallback, False

    candidate = _normalize_e164(preferred_phone)
    if candidate and candidate in _verified_caller_ids():
        return candidate, True

    return fallback, False
