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
import json
import logging
import os
import re
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

import requests
from env_loader import load_project_env
from hf_llm import create_hf_client, hf_api_key, hf_enabled, hf_vision_model

load_project_env()

log = logging.getLogger("menu_image_ocr")

IMAGE_TYPES = frozenset({
    "menu_document",
    "food_photo",
    "interior",
    "logo",
    "other",
})

MENU_TYPES = frozenset({"menu_document"})

CLASSIFY_AND_OCR_PROMPT = """You analyze restaurant photos for a data pipeline.

Look at the image and return JSON only (no markdown):

{
  "image_type": "menu_document" | "food_photo" | "interior" | "logo" | "other",
  "confidence": 0.0-1.0,
  "reason": "one short sentence",
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
- food_photo: a prepared dish, drink, or close-up of food on a plate/bowl.
- interior: dining room, bar, storefront, staff, decor.
- logo: brand mark only.
- other: anything else.

For menu_items:
- ONLY fill menu_items when image_type is menu_document.
- Extract EVERY readable dish line with price when shown.
- Keep original casing for names; normalize prices like "$24.00".
- price_numeric: float without currency symbol, or null if unknown.
- If not a menu, menu_items must be [].
"""


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
    if len(url) < 40:
        raise ValueError(f"URL looks truncated ({len(url)} chars): {url!r}")
    return url


def download_image(url: str, timeout: int = 45) -> tuple[bytes, str]:
    url = _validate_image_url(url)
    headers = {
        "User-Agent": (
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
            "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
        ),
        "Accept": "image/avif,image/webp,image/apng,image/*,*/*;q=0.8",
        "Referer": "https://www.google.com/",
    }
    resp = requests.get(url, headers=headers, timeout=timeout, allow_redirects=True)
    if resp.status_code >= 400:
        raise RuntimeError(
            f"Could not download image (HTTP {resp.status_code}). "
            "Use the full URL from restaurants_data.json — not '...' placeholder."
        )
    content_type = (resp.headers.get("Content-Type") or "image/jpeg").split(";")[0].strip()
    if not content_type.startswith("image/"):
        content_type = "image/jpeg"
    return resp.content, content_type


def _image_ref_for_openai(url: str, image_bytes: bytes, content_type: str) -> dict:
    """Prefer URL for Google-hosted images; base64 for others if needed."""
    host = urlparse(url).netloc.lower()
    if "googleusercontent.com" in host or "ggpht.com" in host:
        return {"type": "image_url", "image_url": {"url": url, "detail": "high"}}
    encoded = base64.b64encode(image_bytes).decode("ascii")
    data_url = f"data:{content_type};base64,{encoded}"
    return {"type": "image_url", "image_url": {"url": data_url, "detail": "high"}}


class MenuImageAnalyzer:
    def __init__(self, cfg: MenuOCRConfig | None = None):
        self.cfg = cfg or MenuOCRConfig()
        self._hf = None
        self._openai = None
        self._gemini = None

        if self.cfg.huggingface_api_key:
            self._hf = create_hf_client()
        elif self.cfg.openai_api_key:
            from openai import OpenAI
            self._openai = OpenAI(api_key=self.cfg.openai_api_key)
        elif self.cfg.gemini_api_key:
            from google import genai
            self._gemini = genai.Client(api_key=self.cfg.gemini_api_key)

    def analyze_image_url(self, url: str) -> dict:
        url = _validate_image_url(url)
        host = urlparse(url).netloc.lower()
        is_google = "googleusercontent.com" in host or "ggpht.com" in host

        if self._hf and is_google:
            return self._analyze_huggingface(url, b"", "image/jpeg")

        image_bytes, content_type = download_image(url, timeout=self.cfg.timeout)

        if self._hf:
            return self._analyze_huggingface(url, image_bytes, content_type)
        if self._openai:
            return self._analyze_openai(url, image_bytes, content_type)
        if self._gemini:
            return self._analyze_gemini(url, image_bytes, content_type)
        raise RuntimeError(
            "No vision API configured (set HUGGING_FACE_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY)"
        )

    def _analyze_huggingface(self, url: str, image_bytes: bytes, content_type: str) -> dict:
        assert self._hf is not None
        image_part = _image_ref_for_openai(url, image_bytes, content_type)
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
        return self._normalize_analysis(parsed, url)

    def _analyze_openai(self, url: str, image_bytes: bytes, content_type: str) -> dict:
        assert self._openai is not None
        image_part = _image_ref_for_openai(url, image_bytes, content_type)
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
        return self._normalize_analysis(parsed, url)

    def _analyze_gemini(self, url: str, image_bytes: bytes, content_type: str) -> dict:
        from google import genai
        from google.genai import types

        assert self._gemini is not None
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
    candidates = collect_candidate_image_urls(record)[: cfg.max_images]
    if not candidates:
        return record

    images_block = record.setdefault("images", {})
    menu_photos: list[dict] = []
    gallery: list[dict] = []
    ocr_items: list[dict] = []
    classifications: list[dict] = []
    url_types: dict[str, str] = {}

    log.info(f"  Menu OCR: analyzing {len(candidates)} image(s)…")

    for idx, url in enumerate(candidates, start=1):
        try:
            result = analyzer.analyze_image_url(url)
            classifications.append(result)
            image_type = result["image_type"]
            confidence = result["confidence"]
            url_types[url] = image_type

            entry = {
                "url": url,
                "thumbnail": url,
                "image_type": image_type,
                "confidence": confidence,
                "reason": result.get("reason", ""),
            }

            if image_type in MENU_TYPES and confidence >= cfg.min_confidence:
                menu_photos.append(entry)
                ocr_items.extend(result.get("menu_items") or [])
                log.info(f"    [{idx}] menu_document ({confidence:.0%}) — {len(result.get('menu_items') or [])} items")
            elif image_type == "food_photo":
                gallery.append({**entry, "title": ""})
                _attach_food_image_to_item(record.get("menu_items") or [], url)
                log.info(f"    [{idx}] food_photo ({confidence:.0%})")
            elif image_type in ("interior", "logo", "other"):
                gallery.append({**entry, "title": image_type})
                log.info(f"    [{idx}] {image_type} ({confidence:.0%})")
            else:
                gallery.append({**entry, "title": "uncertain"})

        except Exception as exc:
            log.warning(f"    [{idx}] OCR failed for {url[:60]}…: {exc}")
            gallery.append({"url": url, "thumbnail": url, "image_type": "error", "error": str(exc)})

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
        "menu_photos_found": len(menu_photos),
        "items_extracted": len(ocr_items),
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
