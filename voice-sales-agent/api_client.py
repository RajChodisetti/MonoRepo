"""Async GET/PUT client for MonoRepo public APIs (voice agent tools)."""

from __future__ import annotations

import logging
import os
from typing import Any

import httpx

log = logging.getLogger("api_client")

DEFAULT_BASE = "http://localhost:8080"
DEFAULT_TIMEOUT_MS = 5000


def _base_url() -> str:
    return os.getenv("MONOREPO_API_URL", DEFAULT_BASE).rstrip("/")


def _timeout() -> float:
    try:
        ms = int(os.getenv("API_TIMEOUT_MS", str(DEFAULT_TIMEOUT_MS)))
    except ValueError:
        ms = DEFAULT_TIMEOUT_MS
    return max(ms, 200) / 1000.0


def _headers() -> dict[str, str]:
    headers = {"Accept": "application/json", "Content-Type": "application/json"}
    token = os.getenv("API_AUTH_TOKEN", "").strip()
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return headers


def _error(message: str) -> dict[str, Any]:
    return {"status": "error", "message": message}


async def get_site_restaurant(index: int) -> dict[str, Any]:
    url = f"{_base_url()}/api/public/v1/site/restaurants/{index}"
    try:
        async with httpx.AsyncClient(timeout=_timeout()) as client:
            resp = await client.get(url, headers=_headers())
            if resp.status_code == 404:
                return _error("Restaurant not found.")
            resp.raise_for_status()
            return resp.json()
    except httpx.TimeoutException:
        log.warning("get_site_restaurant timeout index=%s", index)
        return _error("Could not reach the restaurant system — request timed out.")
    except Exception as exc:
        log.warning("get_site_restaurant failed: %s", exc)
        return _error("Could not load restaurant details right now.")


async def get_table_availability(
    restaurant_id: str,
    date: str,
    party_size: int,
) -> dict[str, Any]:
    url = f"{_base_url()}/api/public/v1/restaurants/{restaurant_id}/table-availability"
    params = {"date": date, "party_size": str(party_size)}
    try:
        async with httpx.AsyncClient(timeout=_timeout()) as client:
            resp = await client.get(url, params=params, headers=_headers())
            if resp.status_code == 404:
                return _error("Restaurant not found.")
            if resp.status_code == 400:
                body = resp.json()
                return _error(body.get("message") or "Invalid availability request.")
            resp.raise_for_status()
            data = resp.json()
            return {"status": "success", **data}
    except httpx.TimeoutException:
        log.warning("get_table_availability timeout")
        return _error("Could not check availability — the booking system timed out.")
    except Exception as exc:
        log.warning("get_table_availability failed: %s", exc)
        return _error("Could not reach the booking system right now.")


async def put_reservation(restaurant_id: str, body: dict[str, Any]) -> dict[str, Any]:
    url = f"{_base_url()}/api/public/v1/restaurants/{restaurant_id}/reservations"
    try:
        async with httpx.AsyncClient(timeout=_timeout()) as client:
            resp = await client.put(url, json=body, headers=_headers())
            if resp.status_code == 404:
                return _error("Restaurant not found.")
            if resp.status_code in (400, 409):
                payload = resp.json()
                return _error(payload.get("message") or "Could not book that table.")
            resp.raise_for_status()
            data = resp.json()
            return {"status": "success", **data}
    except httpx.TimeoutException:
        log.warning("put_reservation timeout")
        return _error("Could not complete the booking — the system timed out.")
    except Exception as exc:
        log.warning("put_reservation failed: %s", exc)
        return _error("Could not complete the booking right now.")
