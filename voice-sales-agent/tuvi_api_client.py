"""Async client for Tuvi company consultation endpoints on the unified API."""

from __future__ import annotations

import logging
import os
from typing import Any

import httpx

log = logging.getLogger("tuvi_api_client")

DEFAULT_BASE = "http://localhost:8080"
DEFAULT_TIMEOUT_MS = 20000


def _base_url() -> str:
    return os.getenv(
        "MONOREPO_API_URL",
        os.getenv("TUVI_WEBSITE_API_URL", DEFAULT_BASE),
    ).rstrip("/")


def _timeout() -> float:
    try:
        ms = int(os.getenv("TUVI_API_TIMEOUT_MS", os.getenv("API_TIMEOUT_MS", str(DEFAULT_TIMEOUT_MS))))
    except ValueError:
        ms = DEFAULT_TIMEOUT_MS
    return max(ms, 200) / 1000.0


def _headers() -> dict[str, str]:
    headers = {"Accept": "application/json", "Content-Type": "application/json"}
    token = os.getenv("TUVI_API_TOKEN", "").strip()
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return headers


def _error(message: str, **extra: Any) -> dict[str, Any]:
    return {"status": "error", "message": message, **extra}


async def get_consultation_availability(
    date: str | None = None,
    days: int | None = None,
) -> dict[str, Any]:
    params: dict[str, str] = {}
    if date and date.strip():
        params["date"] = date.strip()
    if days is not None and days > 0:
        params["days"] = str(days)

    url = f"{_base_url()}/api/v1/company/consultations/availability"
    try:
        async with httpx.AsyncClient(timeout=_timeout()) as client:
            resp = await client.get(url, params=params, headers=_headers())
            if resp.status_code == 400:
                body = resp.json()
                return _error(body.get("message") or "Invalid availability request.")
            resp.raise_for_status()
            data = resp.json()
            slots = data.get("slots") or []
            iso_slots = [s.get("iso") for s in slots if s.get("iso")]
            return {
                "status": "success",
                "slots": slots,
                "available_slots": iso_slots,
            }
    except httpx.TimeoutException:
        log.warning("get_consultation_availability timeout")
        return _error("Could not check consultation availability — request timed out.")
    except Exception as exc:
        log.warning("get_consultation_availability failed: %s", exc)
        return _error("Could not reach the consultation booking system right now.")


async def check_consultation_slot(date: str, time: str) -> dict[str, Any]:
    url = f"{_base_url()}/api/v1/company/consultations/availability/check"
    params = {"date": date.strip(), "time": time.strip()}
    try:
        async with httpx.AsyncClient(timeout=_timeout()) as client:
            resp = await client.get(url, params=params, headers=_headers())
            if resp.status_code == 400:
                body = resp.json()
                return _error(body.get("message") or "Invalid slot check.")
            resp.raise_for_status()
            return {"status": "success", **resp.json()}
    except httpx.TimeoutException:
        log.warning("check_consultation_slot timeout")
        return _error("Could not verify that slot — request timed out.")
    except Exception as exc:
        log.warning("check_consultation_slot failed: %s", exc)
        return _error("Could not verify slot availability right now.")


async def book_consultation(
    *,
    date: str,
    time: str,
    prospect_name: str,
    prospect_email: str = "",
    source: str = "voice",
) -> dict[str, Any]:
    url = f"{_base_url()}/api/v1/company/consultations"
    body = {
        "date": date.strip(),
        "time": time.strip(),
        "prospect_name": prospect_name.strip(),
        "prospect_email": prospect_email.strip(),
        "source": source,
    }
    try:
        async with httpx.AsyncClient(timeout=_timeout()) as client:
            resp = await client.post(url, json=body, headers=_headers())
            if resp.status_code == 409:
                payload = resp.json()
                return {
                    "status": "conflict",
                    "message": payload.get("message") or "That slot is already booked.",
                    "alternatives": payload.get("alternatives") or [],
                }
            if resp.status_code == 400:
                payload = resp.json()
                return _error(payload.get("message") or "Could not book that consultation.")
            resp.raise_for_status()
            data = resp.json()
            return {
                "status": "success",
                "confirmation_code": data.get("confirmation_code", ""),
                "prospect_name": data.get("prospect_name", prospect_name),
                "prospect_email": data.get("prospect_email", prospect_email),
                "slot": data.get("slot", ""),
                "booking_date": data.get("booking_date", date),
                "booking_time": data.get("booking_time", time),
                "calendar_link": data.get("calendar_link", ""),
                "calendly_link": data.get("calendar_link", ""),
                "message": data.get("message", "Your consultation is booked."),
            }
    except httpx.TimeoutException:
        log.warning("book_consultation timeout")
        return _error("Could not complete the booking — request timed out.")
    except Exception as exc:
        log.warning("book_consultation failed: %s", exc)
        return _error("Could not complete the consultation booking right now.")
