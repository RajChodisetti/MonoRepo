#!/usr/bin/env python3
"""
Scrape Google Images for Melbourne restaurant dishes.

Downloads the first usable image per dish, stores files in automation/images/
as <dish name>.jpg|.png|.webp, and writes metadata to automation/data/catalog.json.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import sys
import time
from pathlib import Path
from urllib.parse import quote_plus, unquote

import httpx

SCRIPT_DIR = Path(__file__).resolve().parent
IMAGES_DIR = SCRIPT_DIR / "images"
DATA_DIR = SCRIPT_DIR / "data"
DISHES_FILE = SCRIPT_DIR / "dishes_melbourne.json"
CATALOG_FILE = DATA_DIR / "catalog.json"
PLAYWRIGHT_BROWSERS_PATH = SCRIPT_DIR / ".playwright-browsers"

MIN_IMAGE_BYTES = 8_000
DELAY_SECONDS = 3
DEFAULT_USER_AGENT = (
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
    "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
)

OU_PATTERN = re.compile(r'"ou":"(https?://[^"\\]+)"')
IMG_SRC_PATTERN = re.compile(r'"(https?://[^"\\]+\.(?:jpg|jpeg|png|webp)[^"\\]*)"')
BING_MURL_PATTERN = re.compile(r'murl&quot;:&quot;(https?://[^&]+)')
UNSAFE_FILENAME_CHARS = re.compile(r'[<>:"/\\|?*]')


def sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def guess_extension(data: bytes) -> str:
    if data.startswith(b"\x89PNG"):
        return ".png"
    if data.startswith(b"\xff\xd8"):
        return ".jpg"
    if data.startswith(b"RIFF") and data[8:12] == b"WEBP":
        return ".webp"
    if data.startswith(b"GIF"):
        return ".gif"
    return ".jpg"


def dish_image_filename(name: str, extension: str) -> str:
    cleaned = UNSAFE_FILENAME_CHARS.sub("", name).strip()
    if not cleaned:
        cleaned = "dish"
    return f"{cleaned}{extension}"


def dish_image_path(name: str, extension: str) -> Path:
    return IMAGES_DIR / dish_image_filename(name, extension)


def rename_catalog_images(catalog: list[dict]) -> list[dict]:
    IMAGES_DIR.mkdir(parents=True, exist_ok=True)
    updated_catalog: list[dict] = []

    for entry in catalog:
        name = entry.get("name", "").strip()
        if not name:
            updated_catalog.append(entry)
            continue

        old_rel = entry.get("image_file", "")
        old_path = SCRIPT_DIR / old_rel if old_rel else None
        if old_path is None or not old_path.exists():
            updated_catalog.append(entry)
            print(f"skip rename (file missing): {name}")
            continue

        extension = old_path.suffix or ".jpg"
        new_path = dish_image_path(name, extension)
        new_rel = f"images/{new_path.name}"

        if old_path.resolve() != new_path.resolve():
            if new_path.exists():
                new_path.unlink()
            old_path.rename(new_path)
            print(f"renamed: {old_path.name} -> {new_path.name}")

        entry = dict(entry)
        entry["image_file"] = new_rel
        updated_catalog.append(entry)

    save_catalog(updated_catalog)
    return updated_catalog


def load_json(path: Path, default):
    if not path.exists():
        return default
    return json.loads(path.read_text(encoding="utf-8"))


def save_catalog(catalog: list[dict]) -> None:
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    CATALOG_FILE.write_text(json.dumps(catalog, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def build_query(dish: dict) -> str:
    template = dish.get("search_query")
    if template:
        return template
    return f"melbourne restaurant {dish['name']} food plated"


def normalize_url(url: str) -> str:
    return unquote(url.replace("\\u003d", "=").replace("\\u0026", "&"))


def is_bad_image_url(url: str) -> bool:
    lowered = url.lower()
    blocked = (
        "googlelogo",
        "gstatic.com/images/branding",
        "favicon",
        ".svg",
    )
    return any(token in lowered for token in blocked)


def extract_google_image_urls(html: str) -> list[str]:
    urls: list[str] = []

    for match in OU_PATTERN.findall(html):
        url = normalize_url(match)
        if not is_bad_image_url(url):
            urls.append(url)

    if not urls:
        for match in IMG_SRC_PATTERN.findall(html):
            url = normalize_url(match)
            if not is_bad_image_url(url):
                urls.append(url)

    deduped: list[str] = []
    seen: set[str] = set()
    for url in urls:
        if url in seen:
            continue
        seen.add(url)
        deduped.append(url)
    return deduped


def configure_playwright_path() -> None:
    PLAYWRIGHT_BROWSERS_PATH.mkdir(parents=True, exist_ok=True)
    os.environ.setdefault("PLAYWRIGHT_BROWSERS_PATH", str(PLAYWRIGHT_BROWSERS_PATH))


def launch_browser(playwright):
    configure_playwright_path()
    launch_args = {"headless": True}

    try:
        return playwright.chromium.launch(channel="chrome", **launch_args)
    except Exception:
        pass

    try:
        return playwright.chromium.launch(**launch_args)
    except Exception as exc:
        raise RuntimeError(
            "Could not launch a browser. Install Google Chrome or run:\n"
            "  playwright install chromium"
        ) from exc


def search_google_images_browser(query: str) -> list[str]:
    from playwright.sync_api import TimeoutError as PlaywrightTimeoutError
    from playwright.sync_api import sync_playwright

    with sync_playwright() as playwright:
        browser = launch_browser(playwright)
        context = browser.new_context(user_agent=DEFAULT_USER_AGENT, locale="en-AU")
        page = context.new_page()
        search_url = f"https://www.google.com/search?q={quote_plus(query)}&tbm=isch&hl=en&gl=au"
        page.goto(search_url, wait_until="networkidle", timeout=60_000)

        for selector in ('button:has-text("Accept all")', "#L2AGLb"):
            try:
                page.locator(selector).first.click(timeout=2000)
                time.sleep(0.5)
                break
            except PlaywrightTimeoutError:
                continue
            except Exception:
                continue

        try:
            page.wait_for_selector("a[data-ou], img[src^='http']", timeout=25_000)
        except PlaywrightTimeoutError:
            pass

        page.evaluate("window.scrollTo(0, 800)")
        time.sleep(1.5)

        urls = page.evaluate(
            """
            () => {
                const found = [];

                document.querySelectorAll('a[data-ou]').forEach((anchor) => {
                    const url = anchor.getAttribute('data-ou');
                    if (url && url.startsWith('http')) found.push(url);
                });

                if (found.length === 0) {
                    document.querySelectorAll('img[src^="http"]').forEach((img) => {
                        const src = img.getAttribute('src') || '';
                        if (
                            src.startsWith('http')
                            && !src.includes('googlelogo')
                            && !src.includes('gstatic.com/images/branding')
                        ) {
                            found.push(src);
                        }
                    });
                }

                return [...new Set(found)];
            }
            """
        )
        context.close()
        browser.close()

    return [normalize_url(url) for url in urls if not is_bad_image_url(url)]


def search_bing_images(client: httpx.Client, query: str) -> list[str]:
    url = f"https://www.bing.com/images/search?q={quote_plus(query)}&form=HDRSC2"
    response = client.get(url, timeout=30)
    response.raise_for_status()

    urls: list[str] = []
    seen: set[str] = set()
    for match in BING_MURL_PATTERN.findall(response.text):
        image_url = normalize_url(match)
        if image_url in seen or is_bad_image_url(image_url):
            continue
        seen.add(image_url)
        urls.append(image_url)
    return urls


def search_duckduckgo_images(query: str) -> list[str]:
    try:
        from ddgs import DDGS
    except ImportError:
        from duckduckgo_search import DDGS

    results = DDGS().images(query, max_results=8)
    urls: list[str] = []
    for item in results:
        url = item.get("image") or item.get("thumbnail") or ""
        if url.startswith("http") and not is_bad_image_url(url):
            urls.append(url)
    return urls


def download_image(client: httpx.Client, url: str) -> bytes | None:
    try:
        response = client.get(url, timeout=30)
        response.raise_for_status()
    except Exception:
        return None

    content = response.content
    if len(content) < MIN_IMAGE_BYTES:
        return None

    content_type = (response.headers.get("content-type") or "").lower()
    looks_like_image = content[:4] in (b"\x89PNG", b"\xff\xd8", b"RIFF", b"GIF8")
    if "image" not in content_type and not looks_like_image:
        return None

    return content


def find_image_urls(
    client: httpx.Client,
    query: str,
    *,
    google_only: bool,
    skip_google: bool,
) -> tuple[list[str], str]:
    if not skip_google:
        try:
            urls = search_google_images_browser(query)
            if urls:
                return urls, "google"
        except Exception as exc:
            if google_only:
                raise
            print(f"  google failed ({exc}); trying bing image search")
    elif google_only:
        return [], "google"

    if google_only:
        return [], "google"

    try:
        urls = search_bing_images(client, query)
        if urls:
            return urls, "bing"
    except Exception as exc:
        print(f"  bing failed ({exc}); trying duckduckgo image search")

    try:
        urls = search_duckduckgo_images(query)
        if urls:
            return urls, "duckduckgo"
    except Exception as exc:
        print(f"  duckduckgo failed ({exc})")

    return [], "none"


def scrape_dishes(
    dishes: list[dict],
    catalog: list[dict],
    *,
    delay_seconds: float,
    limit: int | None,
    google_only: bool,
    skip_google: bool,
) -> list[dict]:
    IMAGES_DIR.mkdir(parents=True, exist_ok=True)

    existing_names = {entry["name"] for entry in catalog}
    updated_catalog = list(catalog)

    to_process = [dish for dish in dishes if dish["name"] not in existing_names]
    if limit is not None:
        to_process = to_process[:limit]

    if not to_process:
        print("Nothing to scrape. All dishes are already in catalog.json")
        return updated_catalog

    print(f"Scraping {len(to_process)} Melbourne dish(es) via Google Images (browser)")
    print(f"Saving images to {IMAGES_DIR}")

    headers = {
        "User-Agent": DEFAULT_USER_AGENT,
        "Accept-Language": "en-AU,en;q=0.9",
    }

    with httpx.Client(headers=headers, follow_redirects=True) as client:
        for index, dish in enumerate(to_process, start=1):
            name = dish["name"]
            query = build_query(dish)
            print(f"[{index}/{len(to_process)}] {name}")
            print(f"  query: {query}")

            try:
                image_urls, source = find_image_urls(
                    client,
                    query,
                    google_only=google_only,
                    skip_google=skip_google,
                )
            except Exception as exc:
                print(f"  failed: search error ({exc})")
                save_catalog(updated_catalog)
                time.sleep(delay_seconds)
                continue

            if source != "google":
                print(f"  note: used fallback image search ({source})")

            if not image_urls:
                print("  failed: no image URLs found")
                save_catalog(updated_catalog)
                time.sleep(delay_seconds)
                continue

            saved = False
            for candidate_url in image_urls[:10]:
                image_bytes = download_image(client, candidate_url)
                if image_bytes is None:
                    continue

                image_hash = sha256_bytes(image_bytes)
                extension = guess_extension(image_bytes)
                image_path = dish_image_path(name, extension)

                image_path.write_bytes(image_bytes)

                entry = {
                    "name": name,
                    "cuisines": dish.get("cuisines", []),
                    "location": "Melbourne, Australia",
                    "search_query": query,
                    "image_hash": image_hash,
                    "image_file": f"images/{image_path.name}",
                    "source_url": candidate_url,
                    "image_source": source,
                }
                updated_catalog.append(entry)
                save_catalog(updated_catalog)
                print(f"  saved: {image_path.name} ({source})")
                saved = True
                break

            if not saved:
                print("  failed: could not download a valid image")

            time.sleep(delay_seconds)

    return updated_catalog


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Scrape Google Images for Melbourne restaurant dishes."
    )
    parser.add_argument(
        "--limit",
        type=int,
        default=None,
        help="Only scrape this many dishes that are not already in catalog.json",
    )
    parser.add_argument(
        "--delay",
        type=float,
        default=DELAY_SECONDS,
        help=f"Seconds to wait between searches (default: {DELAY_SECONDS})",
    )
    parser.add_argument(
        "--google-only",
        action="store_true",
        help="Do not fall back to Bing/DuckDuckGo if Google fails",
    )
    parser.add_argument(
        "--skip-google",
        action="store_true",
        help="Skip Google browser search and use Bing image search directly (faster)",
    )
    parser.add_argument(
        "--reset",
        action="store_true",
        help="Ignore existing catalog.json and scrape everything again",
    )
    parser.add_argument(
        "--rename-images",
        action="store_true",
        help="Rename existing catalog images to <dish name>.jpg|.png format",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()

    if not DISHES_FILE.exists():
        print(f"Missing dishes file: {DISHES_FILE}", file=sys.stderr)
        return 1

    dishes = load_json(DISHES_FILE, [])
    if not isinstance(dishes, list) or not dishes:
        print(f"No dishes found in {DISHES_FILE}", file=sys.stderr)
        return 1

    catalog = [] if args.reset else load_json(CATALOG_FILE, [])
    if not isinstance(catalog, list):
        catalog = []

    if args.rename_images:
        if not catalog:
            print("No catalog entries to rename.", file=sys.stderr)
            return 1
        rename_catalog_images(catalog)
        print(f"\nDone. Renamed images in {IMAGES_DIR}")
        print(f"Updated catalog: {CATALOG_FILE}")
        return 0

    scrape_dishes(
        dishes,
        catalog,
        delay_seconds=args.delay,
        limit=args.limit,
        google_only=args.google_only,
        skip_google=args.skip_google,
    )

    final_catalog = load_json(CATALOG_FILE, [])
    print(f"\nDone. Catalog: {CATALOG_FILE}")
    print(f"Images: {IMAGES_DIR}")
    print(f"Total scraped dishes: {len(final_catalog)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
