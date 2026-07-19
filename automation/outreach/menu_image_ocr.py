#!/usr/bin/env python3
"""
Menu image classifier + OCR for restaurant scrape pipeline.

Uses Hugging Face vision (HUGGING_FACE_API_KEY in backend/.env) by default, then OpenAI,
then Gemini fallback.

Usage:
  python menu_image_ocr.py --url "https://..."
  python menu_image_ocr.py --restaurant-json path/to/record.json
"""

from __future__ import annotations

import argparse
import base64
import ipaddress
import json
import logging
import os
import re
import socket
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any
from urllib.parse import urljoin, urlparse

import requests
from env_loader import load_project_env
from hf_llm import create_hf_client, hf_api_key, hf_enabled, hf_vision_model
from ocr_request_budget import OCRDailyBudgetExhausted

load_project_env()

log = logging.getLogger("menu_image_ocr")

IMAGE_TYPES = frozenset({
    "menu_document",
    "food_photo",
    "drink",
    "interior",
    "exterior",
    "logo",
    "team",
    "event",
    "other",
})

MENU_TYPES = frozenset({"menu_document"})
MAX_IMAGE_DOWNLOAD_BYTES = 10 * 1024 * 1024
HARD_MAX_IMAGE_DOWNLOAD_BYTES = 25 * 1024 * 1024
MAX_IMAGE_REDIRECTS = 3

CLASSIFY_AND_OCR_PROMPT = """You analyze restaurant photos for a data pipeline.

Look at the image and return JSON only (no markdown):

{
  "image_type": "menu_document" | "food_photo" | "drink" | "interior" | "exterior" | "logo" | "team" | "event" | "other",
  "confidence": 0.0-1.0,
  "reason": "one short sentence",
  "caption": "short factual website caption with no invented claims",
  "alt_text": "concise accessible description of what is visibly shown",
  "tags": ["up to 8 short visible-content tags"],
  "quality_score": 0.0-1.0,
  "hero_score": 0.0-1.0,
  "orientation": "landscape" | "portrait" | "square" | "unknown",
  "subject_position": "left" | "center" | "right",
  "contains_people": true | false,
  "contains_text": true | false,
  "menu_items": [
    {
      "name": "dish name",
      "category": "section heading if visible else Menu",
      "description": "subtitle or ingredients if visible else empty string",
      "price": "$12.00 or empty",
      "price_numeric": 12.0 or null
    }
  ]
}

Rules:
- menu_document: printed/digital MENU with multiple dish names and often prices (board, paper, PDF screenshot, menu page). NOT a single plated dish photo.
- food_photo: a prepared dish or close-up of food on a plate/bowl.
- drink: a beverage is the main subject.
- interior: dining room, bar, seating, or indoor decor.
- exterior: storefront, patio, facade, entrance, or outdoor restaurant space.
- logo: brand mark only.
- team: one or more staff members are the main subject.
- event: live entertainment, celebration, or hosted restaurant event.
- other: anything else.

For website metadata:
- Describe only what is visible. Never invent a dish name, award, quality claim, occasion, location, or restaurant identity.
- caption may be empty when a useful factual caption is not possible.
- alt_text should describe the visual, not start with "image of" or "photo of".
- hero_score measures suitability as a wide website hero: composition, clarity, lighting, and useful negative space.
- quality_score measures technical/display quality, not how good the restaurant or food is.

For menu_items:
- ONLY fill menu_items when image_type is menu_document.
- Extract EVERY readable dish line with price when shown.
- Keep original casing for names; normalize prices like "$24.00".
- price_numeric: float without currency symbol, or null if unknown.
- If not a menu, menu_items must be [].
"""


class OCRTransientError(RuntimeError):
    """An OCR input or provider failure that should be retried without penalty."""


def _is_transient_provider_error(exc: Exception) -> bool:
    error_name = type(exc).__name__.lower()
    if "timeout" in error_name or error_name in {
        "apiconnectionerror",
        "ratelimiterror",
        "connecterror",
        "networkerror",
    }:
        return True
    raw_status = getattr(exc, "status_code", None)
    if raw_status is None:
        raw_status = getattr(exc, "code", None)
    try:
        status = int(raw_status)
    except (TypeError, ValueError):
        return False
    return status in (408, 429) or status >= 500


def all_scraped_photos_processed(
    summary: dict,
    images_discovered: int,
    resolution_error_count: int,
) -> bool:
    """Return true only when every discovered photo has a successful result."""
    return (
        images_discovered > 0
        and resolution_error_count == 0
        and int(summary.get("images_analyzed") or 0) == images_discovered
        and int(summary.get("images_succeeded") or 0) == images_discovered
        and int(summary.get("images_failed") or 0) == 0
    )


@dataclass
class MenuOCRConfig:
    huggingface_api_key: str = field(default_factory=hf_api_key)
    hf_vision_model: str = field(default_factory=hf_vision_model)
    openai_api_key: str = os.getenv("OPENAI_API_KEY", "")
    openai_model: str = os.getenv("MENU_OCR_MODEL", "gpt-4o-mini")
    gemini_api_key: str = os.getenv("GEMINI_API_KEY", "")
    gemini_model: str = os.getenv("GEMINI_MODEL", "gemini-2.0-flash")
    max_images: int = int(os.getenv("MENU_OCR_MAX_IMAGES", "15"))
    timeout: int = int(os.getenv("MENU_OCR_TIMEOUT", "45"))
    max_image_bytes: int = int(
        os.getenv("MENU_OCR_MAX_IMAGE_BYTES", str(MAX_IMAGE_DOWNLOAD_BYTES))
    )
    delay: float = float(os.getenv("MENU_OCR_DELAY", "0.5"))
    min_confidence: float = float(os.getenv("MENU_OCR_MIN_CONFIDENCE", "0.55"))

    @property
    def enabled(self) -> bool:
        return bool(self.huggingface_api_key or self.openai_api_key or self.gemini_api_key)

    @property
    def primary_provider(self) -> str:
        if self.huggingface_api_key:
            return "huggingface"
        if self.openai_api_key:
            return "openai"
        if self.gemini_api_key:
            return "gemini"
        return ""


def _price_numeric(value: Any) -> float | None:
    if value is None:
        return None
    if isinstance(value, (int, float)):
        return float(value)
    text = str(value).strip()
    if not text:
        return None
    match = re.search(r"(\d+(?:\.\d{1,2})?)", text.replace(",", ""))
    return float(match.group(1)) if match else None


def _normalize_menu_item(raw: dict) -> dict | None:
    name = (raw.get("name") or "").strip()
    if not name or len(name) < 2:
        return None
    price = (raw.get("price") or "").strip()
    return {
        "name": name,
        "category": (raw.get("category") or "Menu").strip() or "Menu",
        "description": (raw.get("description") or "").strip(),
        "price": price,
        "price_numeric": _price_numeric(raw.get("price_numeric") if raw.get("price_numeric") is not None else price),
        "images": [],
        "source": "menu_ocr",
    }


def _parse_json_response(text: str) -> dict:
    text = (text or "").strip()
    if text.startswith("```"):
        text = re.sub(r"^```(?:json)?\s*", "", text)
        text = re.sub(r"\s*```$", "", text)
    return json.loads(text)


def _response_usage(response: Any) -> dict[str, int]:
    """Normalize OpenAI-compatible token usage without depending on SDK types."""
    usage = getattr(response, "usage", None)
    if usage is None:
        return {}

    normalized: dict[str, int] = {}
    for source_name, target_name in (
        ("prompt_tokens", "input_tokens"),
        ("completion_tokens", "output_tokens"),
        ("total_tokens", "total_tokens"),
    ):
        value = getattr(usage, source_name, None)
        if value is None and isinstance(usage, dict):
            value = usage.get(source_name)
        try:
            parsed = int(value)
        except (TypeError, ValueError):
            continue
        if parsed >= 0:
            normalized[target_name] = parsed
    return normalized


def _validate_image_url(url: str) -> str:
    url = (url or "").strip()
    if not url:
        raise ValueError("Image URL is required.")
    if url.endswith("...") or "/..." in url or url == "https://lh3.googleusercontent.com/...":
        raise ValueError(
            "Invalid URL — you pasted the placeholder '...'. "
            "Copy the full image URL from restaurants_data.json (the entire https://lh3.googleusercontent.com/... string)."
        )
    parsed = urlparse(url)
    if parsed.scheme not in ("http", "https") or not parsed.netloc:
        raise ValueError(f"Invalid URL: {url!r}")
    if parsed.username is not None or parsed.password is not None:
        raise ValueError("Image URLs must not contain credentials.")
    if len(url) > 4096:
        raise ValueError("Image URL is too long.")
    hostname = (parsed.hostname or "").rstrip(".").lower()
    if not hostname:
        raise ValueError(f"Invalid URL host: {url!r}")
    if (
        hostname == "localhost"
        or hostname.endswith(".localhost")
        or hostname.endswith(".local")
        or hostname.endswith(".internal")
        or hostname.endswith(".home.arpa")
    ):
        raise ValueError("Private or local image hosts are not allowed.")
    try:
        port = parsed.port
    except ValueError as exc:
        raise ValueError("Image URL has an invalid port.") from exc
    expected_port = 443 if parsed.scheme == "https" else 80
    if port is not None and port != expected_port:
        raise ValueError("Image URLs must use the standard HTTP or HTTPS port.")
    try:
        address = ipaddress.ip_address(hostname)
    except ValueError:
        address = None
    if address is not None and not address.is_global:
        raise ValueError("Private or local image addresses are not allowed.")
    if len(url) < 40:
        raise ValueError(f"URL looks truncated ({len(url)} chars): {url!r}")
    return url


def _validate_public_image_target(url: str) -> str:
    url = _validate_image_url(url)
    parsed = urlparse(url)
    hostname = (parsed.hostname or "").rstrip(".").lower()
    port = parsed.port or (443 if parsed.scheme == "https" else 80)
    try:
        addresses = {
            item[4][0]
            for item in socket.getaddrinfo(
                hostname,
                port,
                type=socket.SOCK_STREAM,
            )
        }
    except socket.gaierror as exc:
        raise ValueError("Image host could not be resolved.") from exc
    if not addresses:
        raise ValueError("Image host did not resolve to an address.")
    for raw_address in addresses:
        try:
            address = ipaddress.ip_address(raw_address.split("%", 1)[0])
        except ValueError as exc:
            raise ValueError("Image host resolved to an invalid address.") from exc
        if not address.is_global:
            raise ValueError("Image host resolved to a private or local address.")
    return url


def _is_google_image_host(hostname: str) -> bool:
    hostname = (hostname or "").rstrip(".").lower()
    return hostname == "googleusercontent.com" or hostname.endswith(".googleusercontent.com") or (
        hostname == "ggpht.com" or hostname.endswith(".ggpht.com")
    )


def is_trusted_automated_image_url(url: str) -> bool:
    """Allow unattended OCR to fetch only HTTPS Google-hosted direct images.

    The durable Places flow resolves first-party photo resources separately.
    Arbitrary restaurant/CDN URLs remain available to explicit manual tooling,
    but are not fetched by the unattended database job where DNS rebinding
    could otherwise turn image retrieval into an internal-network request.
    """
    try:
        parsed = urlparse(str(url or "").strip())
    except ValueError:
        return False
    return parsed.scheme == "https" and _is_google_image_host(parsed.hostname or "")


def _validate_response_peer(resp: requests.Response) -> None:
    connection = getattr(resp.raw, "connection", None) or getattr(
        resp.raw,
        "_connection",
        None,
    )
    sock = getattr(connection, "sock", None)
    if sock is None:
        try:
            sock = resp.raw._fp.fp.raw._sock  # type: ignore[attr-defined]
        except AttributeError as exc:
            raise RuntimeError("Could not verify the image server address.") from exc
    try:
        raw_address = sock.getpeername()[0]
        address = ipaddress.ip_address(str(raw_address).split("%", 1)[0])
    except (OSError, ValueError) as exc:
        raise RuntimeError("Could not verify the image server address.") from exc
    if not address.is_global:
        raise RuntimeError("Image request connected to a private or local address.")


def download_image(
    url: str,
    timeout: int = 45,
    max_bytes: int = MAX_IMAGE_DOWNLOAD_BYTES,
) -> tuple[bytes, str]:
    current_url = _validate_public_image_target(url)
    max_bytes = min(max(1, int(max_bytes)), HARD_MAX_IMAGE_DOWNLOAD_BYTES)
    headers = {
        "User-Agent": (
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
            "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
        ),
        "Accept": "image/avif,image/webp,image/apng,image/*,*/*;q=0.8",
        "Referer": "https://www.google.com/",
    }
    session = requests.Session()
    session.trust_env = False
    try:
        for redirect_count in range(MAX_IMAGE_REDIRECTS + 1):
            with session.get(
                current_url,
                headers=headers,
                timeout=(min(max(1, timeout), 10), max(1, timeout)),
                allow_redirects=False,
                stream=True,
            ) as resp:
                _validate_response_peer(resp)
                if resp.is_redirect or resp.is_permanent_redirect:
                    if redirect_count >= MAX_IMAGE_REDIRECTS:
                        raise RuntimeError("Image download exceeded the redirect limit.")
                    location = (resp.headers.get("Location") or "").strip()
                    if not location:
                        raise RuntimeError("Image redirect did not include a target URL.")
                    next_url = urljoin(current_url, location)
                    if (
                        urlparse(current_url).scheme == "https"
                        and urlparse(next_url).scheme != "https"
                    ):
                        raise RuntimeError("Image redirect attempted an HTTPS downgrade.")
                    current_url = _validate_public_image_target(next_url)
                    continue
                if resp.status_code < 200 or resp.status_code >= 300:
                    raise RuntimeError(
                        f"Could not download image (HTTP {resp.status_code}). "
                        "Use the full URL from restaurants_data.json — not '...' placeholder."
                    )

                content_type = (resp.headers.get("Content-Type") or "").split(";", 1)[0].strip().lower()
                if not content_type.startswith("image/") or content_type in {
                    "image/svg+xml",
                    "image/svg",
                }:
                    raise RuntimeError("Image URL returned a non-raster content type.")

                content_length = (resp.headers.get("Content-Length") or "").strip()
                if content_length:
                    try:
                        declared_length = int(content_length)
                    except ValueError as exc:
                        raise RuntimeError("Image response had an invalid Content-Length.") from exc
                    if declared_length < 0 or declared_length > max_bytes:
                        raise RuntimeError(
                            f"Image exceeds the {max_bytes}-byte download limit."
                        )

                chunks: list[bytes] = []
                total = 0
                for chunk in resp.iter_content(chunk_size=64 * 1024):
                    if not chunk:
                        continue
                    total += len(chunk)
                    if total > max_bytes:
                        raise RuntimeError(
                            f"Image exceeds the {max_bytes}-byte download limit."
                        )
                    chunks.append(chunk)
                if total == 0:
                    raise RuntimeError("Image response was empty.")
                return b"".join(chunks), content_type
    finally:
        session.close()

    raise RuntimeError("Image download did not produce a response.")


def _image_ref_for_openai(url: str, image_bytes: bytes, content_type: str) -> dict:
    """Prefer URL for Google-hosted images; base64 for others if needed."""
    host = (urlparse(url).hostname or "").lower()
    if _is_google_image_host(host):
        return {"type": "image_url", "image_url": {"url": url, "detail": "high"}}
    encoded = base64.b64encode(image_bytes).decode("ascii")
    data_url = f"data:{content_type};base64,{encoded}"
    return {"type": "image_url", "image_url": {"url": data_url, "detail": "high"}}


class MenuImageAnalyzer:
    def __init__(self, cfg: MenuOCRConfig | None = None, request_budget=None):
        self.cfg = cfg or MenuOCRConfig()
        self._request_budget = request_budget
        self._hf = None
        self._openai = None
        self._gemini = None

        if self.cfg.huggingface_api_key:
            self._hf = create_hf_client(
                timeout=max(1.0, float(self.cfg.timeout)),
                max_retries=0,
            )
        elif self.cfg.openai_api_key:
            from openai import OpenAI
            self._openai = OpenAI(
                api_key=self.cfg.openai_api_key,
                timeout=max(1.0, float(self.cfg.timeout)),
                max_retries=0,
            )
        elif self.cfg.gemini_api_key:
            from google import genai
            from google.genai import types

            self._gemini = genai.Client(
                api_key=self.cfg.gemini_api_key,
                http_options=types.HttpOptions(
                    timeout=max(1, int(self.cfg.timeout)) * 1000,
                    retry_options=types.HttpRetryOptions(attempts=1),
                ),
            )

    def _reserve_provider_request(self) -> None:
        if self._request_budget is None:
            return
        snapshot = self._request_budget.reserve()
        log.info(
            "  OCR daily request budget: %d/%d used",
            snapshot.requests_used,
            snapshot.daily_limit,
        )

    def analyze_image_url(self, url: str) -> dict:
        url = _validate_public_image_target(url)
        host = (urlparse(url).hostname or "").lower()
        is_google = _is_google_image_host(host)

        try:
            if self._hf and is_google:
                return self._analyze_huggingface(url, b"", "image/jpeg")

            try:
                image_bytes, content_type = download_image(
                    url,
                    timeout=self.cfg.timeout,
                    max_bytes=self.cfg.max_image_bytes,
                )
            except requests.Timeout as exc:
                raise OCRTransientError("Image download timed out; the OCR claim will be retried") from exc

            if self._hf:
                return self._analyze_huggingface(url, image_bytes, content_type)
            if self._openai:
                return self._analyze_openai(url, image_bytes, content_type)
            if self._gemini:
                return self._analyze_gemini(url, image_bytes, content_type)
            raise RuntimeError(
                "No vision API configured (set HUGGING_FACE_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY)"
            )
        except (OCRDailyBudgetExhausted, OCRTransientError):
            raise
        except Exception as exc:
            if _is_transient_provider_error(exc):
                raise OCRTransientError(
                    "OCR provider timed out or was temporarily unavailable; the claim will be retried"
                ) from exc
            raise

    def _analyze_huggingface(self, url: str, image_bytes: bytes, content_type: str) -> dict:
        assert self._hf is not None
        image_part = _image_ref_for_openai(url, image_bytes, content_type)
        self._reserve_provider_request()
        response = self._hf.chat.completions.create(
            model=self.cfg.hf_vision_model,
            messages=[{
                "role": "user",
                "content": [
                    {"type": "text", "text": CLASSIFY_AND_OCR_PROMPT + "\n\nReturn valid JSON only."},
                    image_part,
                ],
            }],
            max_tokens=4096,
            temperature=0.1,
        )
        raw = response.choices[0].message.content or "{}"
        parsed = _parse_json_response(raw)
        result = self._normalize_analysis(parsed, url)
        usage = _response_usage(response)
        if usage:
            result["usage"] = usage
        return result

    def _analyze_openai(self, url: str, image_bytes: bytes, content_type: str) -> dict:
        assert self._openai is not None
        image_part = _image_ref_for_openai(url, image_bytes, content_type)
        self._reserve_provider_request()
        response = self._openai.chat.completions.create(
            model=self.cfg.openai_model,
            messages=[{
                "role": "user",
                "content": [
                    {"type": "text", "text": CLASSIFY_AND_OCR_PROMPT},
                    image_part,
                ],
            }],
            response_format={"type": "json_object"},
            max_tokens=4096,
            temperature=0.1,
        )
        raw = response.choices[0].message.content or "{}"
        parsed = _parse_json_response(raw)
        result = self._normalize_analysis(parsed, url)
        usage = _response_usage(response)
        if usage:
            result["usage"] = usage
        return result

    def _analyze_gemini(self, url: str, image_bytes: bytes, content_type: str) -> dict:
        from google import genai
        from google.genai import types

        assert self._gemini is not None
        self._reserve_provider_request()
        response = self._gemini.models.generate_content(
            model=self.cfg.gemini_model,
            contents=[
                types.Content(
                    role="user",
                    parts=[
                        types.Part.from_text(text=CLASSIFY_AND_OCR_PROMPT),
                        types.Part.from_bytes(data=image_bytes, mime_type=content_type),
                    ],
                )
            ],
            config=types.GenerateContentConfig(
                response_mime_type="application/json",
                temperature=0.1,
            ),
        )
        raw = response.text or "{}"
        parsed = _parse_json_response(raw)
        return self._normalize_analysis(parsed, url)

    def _normalize_analysis(self, parsed: dict, url: str) -> dict:
        image_type = (parsed.get("image_type") or "other").strip().lower()
        if image_type not in IMAGE_TYPES:
            image_type = "other"

        confidence = parsed.get("confidence")
        try:
            confidence = float(confidence)
        except (TypeError, ValueError):
            confidence = 0.5
        confidence = max(0.0, min(1.0, confidence))

        def score(name: str, default: float = 0.0) -> float:
            try:
                value = float(parsed.get(name))
            except (TypeError, ValueError):
                value = default
            return max(0.0, min(1.0, value))

        def boolean(name: str) -> bool:
            value = parsed.get(name, False)
            if isinstance(value, bool):
                return value
            return str(value or "").strip().lower() in {"1", "true", "yes"}

        tags: list[str] = []
        seen_tags: set[str] = set()
        for raw_tag in parsed.get("tags") or []:
            tag = str(raw_tag or "").strip()[:40]
            key = tag.lower()
            if not tag or key in seen_tags:
                continue
            seen_tags.add(key)
            tags.append(tag)
            if len(tags) == 8:
                break

        orientation = str(parsed.get("orientation") or "unknown").strip().lower()
        if orientation not in {"landscape", "portrait", "square", "unknown"}:
            orientation = "unknown"
        subject_position = str(parsed.get("subject_position") or "center").strip().lower()
        if subject_position not in {"left", "center", "right"}:
            subject_position = "center"

        menu_items: list[dict] = []
        if image_type in MENU_TYPES:
            for raw in parsed.get("menu_items") or []:
                if not isinstance(raw, dict):
                    continue
                item = _normalize_menu_item(raw)
                if item:
                    item["images"] = [{
                        "url": url,
                        "thumbnail": url,
                        "source": "menu_ocr",
                    }]
                    menu_items.append(item)

        return {
            "url": url,
            "image_type": image_type,
            "confidence": confidence,
            "reason": (parsed.get("reason") or "").strip(),
            "caption": str(parsed.get("caption") or "").strip()[:180],
            "alt_text": str(parsed.get("alt_text") or "").strip()[:180],
            "tags": tags,
            "quality_score": score("quality_score", confidence),
            "hero_score": score("hero_score", 0.0),
            "orientation": orientation,
            "subject_position": subject_position,
            "contains_people": boolean("contains_people"),
            "contains_text": boolean("contains_text"),
            "menu_items": menu_items,
        }


def _image_url_from_record(img: Any) -> str:
    if isinstance(img, str):
        return img.strip()
    if isinstance(img, dict):
        return (img.get("url") or img.get("thumbnail") or "").strip()
    return ""


def collect_candidate_image_urls(record: dict) -> list[str]:
    """Gather unique image URLs from a scraped restaurant record."""
    urls: list[str] = []
    seen: set[str] = set()

    def add(url: str) -> None:
        url = (url or "").strip()
        if not url or url in seen:
            return
        seen.add(url)
        urls.append(url)

    images = record.get("images") or {}
    add(images.get("thumbnail") or "")

    for img in images.get("gallery") or []:
        add(_image_url_from_record(img))
    for img in images.get("menu_photos") or []:
        add(_image_url_from_record(img))

    for item in record.get("menu_items") or []:
        for img in item.get("images") or []:
            add(_image_url_from_record(img))

    return urls


def _merge_menu_items(existing: list[dict], ocr_items: list[dict]) -> list[dict]:
    seen = {i.get("name", "").strip().lower() for i in existing if i.get("name")}
    merged = list(existing)
    for item in ocr_items:
        key = item.get("name", "").strip().lower()
        if not key or key in seen:
            continue
        seen.add(key)
        merged.append(item)
    return merged


def _attach_food_image_to_item(items: list[dict], url: str, title: str = "") -> None:
    """Try to match a food photo to an existing menu item by name similarity."""
    title_l = title.strip().lower()
    for item in items:
        name_l = item.get("name", "").strip().lower()
        if not name_l:
            continue
        if title_l and (title_l in name_l or name_l in title_l):
            item.setdefault("images", []).append({
                "url": url,
                "thumbnail": url,
                "source": "food_photo_match",
            })
            return

    # Unmatched food photo → synthetic highlight item
    items.append({
        "name": title or "House Special",
        "category": "Popular Highlights",
        "description": "",
        "price": "",
        "price_numeric": None,
        "images": [{"url": url, "thumbnail": url, "source": "food_photo"}],
    })


def _sanitize_menu_item_images(record: dict, url_types: dict[str, str] | None = None) -> None:
    """Remove menu board URLs from dish cards; keep only food photos on menu_items."""
    url_types = url_types or {}
    images = record.get("images") or {}
    menu_board_urls: set[str] = set()

    for img in images.get("menu_photos") or []:
        u = _image_url_from_record(img)
        if u:
            menu_board_urls.add(u)

    for cls in (record.get("menu_ocr") or {}).get("classifications") or []:
        u = (cls.get("url") or "").strip()
        t = (cls.get("image_type") or "").strip().lower()
        if u and t:
            url_types[u] = t
            if t in MENU_TYPES:
                menu_board_urls.add(u)

    def is_food_url(url: str) -> bool:
        if not url:
            return False
        if url in menu_board_urls:
            return False
        t = url_types.get(url, "").lower()
        if t in MENU_TYPES:
            return False
        if t == "food_photo":
            return True
        # Unknown URL on a dish card: allow only if NOT a known menu board
        return url not in menu_board_urls

    for item in record.get("menu_items") or []:
        cleaned: list[dict] = []
        for img in item.get("images") or []:
            u = _image_url_from_record(img)
            if is_food_url(u):
                entry = img if isinstance(img, dict) else {"url": u, "thumbnail": u}
                entry = {**entry, "image_type": entry.get("image_type") or url_types.get(u) or "food_photo"}
                cleaned.append(entry)
        item["images"] = cleaned


def enrich_restaurant_with_menu_ocr(
    record: dict,
    cfg: MenuOCRConfig | None = None,
    analyzer: MenuImageAnalyzer | None = None,
    analysis_candidates: list[dict] | None = None,
    *,
    process_all_candidates: bool = False,
) -> dict:
    """
    Classify images, split menu_photos vs gallery, OCR menus, merge menu_items.
    Mutates and returns record.
    """
    cfg = cfg or MenuOCRConfig()
    if not cfg.enabled:
        log.debug("Menu OCR skipped — no API key")
        return record

    analyzer = analyzer or MenuImageAnalyzer(cfg)
    if analysis_candidates is None:
        analysis_candidates = [
            {
                "analysis_url": url,
                "persistent_url": url,
                "source": "public_url",
            }
            for url in collect_candidate_image_urls(record)
        ]
    candidates = [
        candidate
        for candidate in analysis_candidates
        if isinstance(candidate, dict) and str(candidate.get("analysis_url") or "").strip()
    ]
    if not process_all_candidates:
        candidates = candidates[: cfg.max_images]
    if not candidates:
        return record

    images_block = record.setdefault("images", {})
    menu_photos: list[dict] = []
    gallery: list[dict] = []
    ocr_items: list[dict] = []
    classifications: list[dict] = []
    url_types: dict[str, str] = {}
    images_succeeded = 0
    images_failed = 0
    input_tokens = 0
    output_tokens = 0
    total_tokens = 0

    log.info(f"  Menu OCR: analyzing {len(candidates)} image(s)…")

    for idx, candidate in enumerate(candidates, start=1):
        url = str(candidate.get("analysis_url") or "").strip()
        persistent_url = str(candidate.get("persistent_url") or "").strip()
        source = str(candidate.get("source") or "public_url").strip()
        source_ref = str(candidate.get("source_ref") or "").strip()
        google_place_id = str(candidate.get("google_place_id") or "").strip()
        source_index = candidate.get("source_index")
        source_fingerprint = str(candidate.get("source_fingerprint") or "").strip()
        attribution = candidate.get("author_attributions") or []
        source_metadata = {"source": source}
        if source == "google_places_photo":
            if google_place_id:
                source_metadata["google_place_id"] = google_place_id
            if isinstance(source_index, int) and source_index >= 0:
                source_metadata["source_index"] = source_index
            if re.fullmatch(r"[0-9a-f]{64}", source_fingerprint):
                source_metadata["source_fingerprint"] = source_fingerprint
        elif source_ref:
            source_metadata["source_ref"] = source_ref
        if attribution:
            source_metadata["author_attributions"] = attribution
        try:
            result = analyzer.analyze_image_url(url)
            images_succeeded += 1
            usage = result.get("usage") or {}
            input_tokens += int(usage.get("input_tokens") or 0)
            output_tokens += int(usage.get("output_tokens") or 0)
            total_tokens += int(usage.get("total_tokens") or 0)
            image_type = result["image_type"]
            confidence = result["confidence"]
            if persistent_url:
                url_types[persistent_url] = image_type

            normalized_items = list(result.get("menu_items") or [])
            if not persistent_url:
                # Photo Media photoUri values are short-lived. Keep extracted text,
                # but strip the transient URI from every persisted menu item.
                normalized_items = [
                    {**item, "images": []}
                    for item in normalized_items
                    if isinstance(item, dict)
                ]

            if persistent_url:
                persisted_classification = {
                    key: value
                    for key, value in result.items()
                    if key not in ("url", "menu_items", "usage")
                }
            else:
                # Google Places media is live-only. Persist its stable Place ID,
                # source index, and internal classification, but not generated
                # website copy or the temporary media URL/name.
                persisted_classification = {
                    "image_type": image_type,
                    "confidence": confidence,
                    "reason": result.get("reason", ""),
                }
            persisted_classification["public_eligible"] = bool(
                image_type not in MENU_TYPES
                and confidence >= cfg.min_confidence
                and not (
                    image_type == "other"
                    and bool(result.get("contains_text", False))
                )
            )
            persisted_classification["menu_items"] = normalized_items
            if persistent_url:
                persisted_classification["url"] = persistent_url
            else:
                persisted_classification.update(source_metadata)
            classifications.append(persisted_classification)

            entry = {
                "image_type": image_type,
                "confidence": confidence,
                "reason": result.get("reason", ""),
            }
            if persistent_url:
                entry.update({
                    "url": persistent_url,
                    "thumbnail": persistent_url,
                    "caption": result.get("caption", ""),
                    "alt_text": result.get("alt_text", ""),
                    "tags": result.get("tags", []),
                    "quality_score": result.get("quality_score"),
                    "hero_score": result.get("hero_score"),
                    "orientation": result.get("orientation", "unknown"),
                    "subject_position": result.get("subject_position", "center"),
                    "contains_people": result.get("contains_people", False),
                    "contains_text": result.get("contains_text", False),
                })
            else:
                entry.update(source_metadata)

            if image_type in MENU_TYPES and confidence >= cfg.min_confidence:
                menu_photos.append(entry)
                ocr_items.extend(normalized_items)
                log.info(f"    [{idx}] menu_document ({confidence:.0%}) — {len(normalized_items)} items")
            elif image_type in ("food_photo", "drink"):
                gallery.append({**entry, "title": ""})
                if persistent_url and image_type == "food_photo":
                    _attach_food_image_to_item(record.get("menu_items") or [], persistent_url)
                log.info(f"    [{idx}] {image_type} ({confidence:.0%})")
            elif image_type in ("interior", "exterior", "logo", "team", "event", "other"):
                gallery.append({**entry, "title": image_type})
                log.info(f"    [{idx}] {image_type} ({confidence:.0%})")
            else:
                gallery.append({**entry, "title": "uncertain"})

        except (OCRDailyBudgetExhausted, OCRTransientError):
            raise
        except Exception as exc:
            images_failed += 1
            display_ref = persistent_url or source_ref or source
            if persistent_url:
                log.warning(f"    [{idx}] OCR failed for {display_ref[:60]}…: {exc}")
            else:
                log.warning(f"    [{idx}] OCR failed for {display_ref[:60]}…")
            persisted_error = str(exc) if persistent_url else "Google Places photo analysis failed"
            error_entry = {"image_type": "error", "error": persisted_error}
            if persistent_url:
                error_entry.update({"url": persistent_url, "thumbnail": persistent_url})
            else:
                error_entry.update(source_metadata)
            gallery.append(error_entry)

        if idx < len(candidates):
            time.sleep(cfg.delay)

    # Prefer classified lists; keep unprocessed originals if OCR found nothing
    if menu_photos:
        images_block["menu_photos"] = menu_photos
    if gallery:
        existing_gallery = images_block.get("gallery") or []
        seen_g = {(_image_url_from_record(g)) for g in gallery}
        for g in existing_gallery:
            u = _image_url_from_record(g)
            if u and u not in seen_g:
                gallery.append(g if isinstance(g, dict) else {"url": u, "title": ""})
        images_block["gallery"] = gallery

    if ocr_items:
        record["menu_items"] = _merge_menu_items(record.get("menu_items") or [], ocr_items)

    record.setdefault("menu_ocr", {})
    provider = cfg.primary_provider
    model = (
        cfg.hf_vision_model
        if provider == "huggingface"
        else cfg.openai_model
        if provider == "openai"
        else cfg.gemini_model
    )
    record["menu_ocr"].update({
        "provider": provider,
        "model": model,
        "images_analyzed": len(candidates),
        "images_succeeded": images_succeeded,
        "images_failed": images_failed,
        "menu_photos_found": len(menu_photos),
        "items_extracted": len(ocr_items),
        "input_tokens": input_tokens,
        "output_tokens": output_tokens,
        "total_tokens": total_tokens,
        "classifications": classifications,
    })

    _sanitize_menu_item_images(record, url_types)

    return record


def _sample_image_url_from_data() -> str:
    """Pick first menu photo URL from bundled restaurant scrape data."""
    candidates = [
        Path(__file__).resolve().parents[2] / "data" / "restaurants_data.json",
        Path("data/restaurants_data.json"),
        Path("../../data/restaurants_data.json"),
    ]
    for path in candidates:
        if not path.is_file():
            continue
        doc = json.loads(path.read_text(encoding="utf-8"))
        for restaurant in doc.get("restaurants") or []:
            for img in (restaurant.get("images") or {}).get("menu_photos") or []:
                url = _image_url_from_record(img)
                if url:
                    return url
            for item in restaurant.get("menu_items") or []:
                for img in item.get("images") or []:
                    url = _image_url_from_record(img)
                    if url:
                        return url
    raise FileNotFoundError("No sample image found — pass --url with a full Google image URL")


def main() -> int:
    parser = argparse.ArgumentParser(description="Classify restaurant images + OCR menus")
    parser.add_argument("--url", help="Full image URL to analyze (not '...' placeholder)")
    parser.add_argument("--sample", action="store_true",
                        help="Use first image URL from data/restaurants_data.json")
    parser.add_argument("--restaurant-json", help="Path to one restaurant JSON object")
    parser.add_argument("--batch-data", type=Path,
                        help="Run OCR on every restaurant in a JSON data file")
    parser.add_argument("--sanitize-data", type=Path,
                        help="Strip menu board URLs from dish cards (no API calls)")
    parser.add_argument("--limit", type=int, default=0, help="Max restaurants for --batch-data")
    parser.add_argument("-v", "--verbose", action="store_true")
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(asctime)s  [%(levelname)s]  %(message)s",
    )

    cfg = MenuOCRConfig()
    analyzer = MenuImageAnalyzer(cfg) if cfg.enabled else None

    if args.sanitize_data:
        path = args.sanitize_data
        doc = json.loads(path.read_text(encoding="utf-8"))
        restaurants = doc.get("restaurants") or []
        if args.limit:
            restaurants = restaurants[: args.limit]
        for rec in restaurants:
            _sanitize_menu_item_images(rec)
            log.info(f"  Sanitized: {rec.get('name', '?')}")
        path.write_text(json.dumps(doc, indent=2, ensure_ascii=False), encoding="utf-8")
        log.info(f"Wrote {len(restaurants)} restaurant(s) → {path}")
        return 0

    if not cfg.enabled:
        log.error(
            "Set HUGGING_FACE_API_KEY (backend/.env), OPENAI_API_KEY, or GEMINI_API_KEY"
        )
        return 1

    if args.batch_data:
        path = args.batch_data
        doc = json.loads(path.read_text(encoding="utf-8"))
        all_restaurants = doc.get("restaurants") or []
        end = args.limit if args.limit else len(all_restaurants)
        total = end
        for i in range(end):
            rec = all_restaurants[i]
            name = rec.get("name") or "?"
            log.info(f"[{i + 1}/{total}] OCR: {name}")
            try:
                enrich_restaurant_with_menu_ocr(rec, cfg, analyzer)
            except Exception as exc:
                log.warning(f"  OCR failed for {name}: {exc}")
        doc["restaurants"] = all_restaurants
        path.write_text(json.dumps(doc, indent=2, ensure_ascii=False), encoding="utf-8")
        log.info(f"Batch OCR complete → {path} ({total} restaurant(s))")
        return 0

    if args.sample:
        args.url = _sample_image_url_from_data()
        log.info(f"Sample URL: {args.url[:80]}…")

    if args.url:
        try:
            result = analyzer.analyze_image_url(args.url)
        except (ValueError, RuntimeError) as exc:
            log.error(str(exc))
            return 1
        print(json.dumps(result, indent=2, ensure_ascii=False))
        return 0

    if args.restaurant_json:
        data = json.loads(open(args.restaurant_json, encoding="utf-8").read())
        enriched = enrich_restaurant_with_menu_ocr(data, cfg, analyzer)
        print(json.dumps(enriched, indent=2, ensure_ascii=False))
        return 0

    parser.print_help()
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
