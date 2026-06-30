"""
Google Places API scraper (legacy Places API — Text Search + Place Details).
Reads PLACES_API or GOOGLE_PLACES_API_KEY from environment.
"""

from __future__ import annotations

import logging
import re
import time
from typing import Any

import requests

from tuvi_outreach_agent import Config, _lead_from_dict, with_retry

log = logging.getLogger("google_places")

LEGACY_BASE = "https://maps.googleapis.com/maps/api/place"
EMAIL_RE = re.compile(r"[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}")
OWNER_TITLE_KEYWORDS = ("owner", "founder", "co-founder", "proprietor", "managing director")

PRICE_LEVEL_MAP = {0: "", 1: "$", 2: "$$", 3: "$$$", 4: "$$$$"}

GENERIC_TYPES = {
    "point_of_interest", "establishment", "food", "store",
    "restaurant", "meal_takeaway", "meal_delivery",
}


def get_places_api_key(cfg: Config | None = None) -> str:
    cfg = cfg or Config()
    key = getattr(cfg, "PLACES_API_KEY", "") or ""
    if not key:
        raise ValueError(
            "PLACES_API or GOOGLE_PLACES_API_KEY is missing from .env. "
            "Enable 'Places API' in Google Cloud Console."
        )
    return key


def _places_get(url: str, params: dict, cfg: Config, label: str) -> dict:
    api_key = get_places_api_key(cfg)
    params = {**params, "key": api_key}

    def _call():
        resp = requests.get(url, params=params, timeout=cfg.SCRAPE_TIMEOUT)
        if resp.status_code != 200:
            raise RuntimeError(f"HTTP {resp.status_code}: {resp.text[:300]}")
        data = resp.json()
        status = data.get("status", "")
        if status not in ("OK", "ZERO_RESULTS"):
            raise RuntimeError(data.get("error_message") or status or "Places API error")
        return data

    return with_retry(_call, cfg.RETRY_ATTEMPTS, cfg.RETRY_BACKOFF, label=label)


def search_place_id(company_name: str, city: str, cfg: Config) -> tuple[str | None, dict]:
    """Text Search → place_id and basic result."""
    query = f"{company_name} restaurant {city} Australia"
    data = _places_get(
        f"{LEGACY_BASE}/textsearch/json",
        {"query": query, "language": "en"},
        cfg,
        label=f"places-search:{company_name}",
    )
    results = data.get("results") or []
    if not results:
        return None, {}
    top = results[0]
    return top.get("place_id"), top


def get_place_details(place_id: str, cfg: Config) -> dict:
    """Place Details with reviews, photos, hours, contact."""
    fields = ",".join([
        "name", "place_id", "formatted_address", "formatted_phone_number",
        "international_phone_number", "website", "url", "rating",
        "user_ratings_total", "opening_hours", "price_level", "types",
        "geometry", "reviews", "photos", "editorial_summary", "business_status",
    ])
    data = _places_get(
        f"{LEGACY_BASE}/details/json",
        {"place_id": place_id, "fields": fields, "language": "en"},
        cfg,
        label=f"places-details:{place_id[:12]}",
    )
    return data.get("result") or {}


def build_photo_url(photo_reference: str, cfg: Config, maxwidth: int = 1200) -> str:
    api_key = get_places_api_key(cfg)
    return (
        f"{LEGACY_BASE}/photo"
        f"?maxwidth={maxwidth}&photo_reference={photo_reference}&key={api_key}"
    )


def _extract_cuisines_from_types(types: list[str]) -> list[str]:
    cuisines = []
    for t in types or []:
        label = t.replace("_", " ").strip()
        if t.lower() in GENERIC_TYPES:
            continue
        cuisines.append(label.title())
    return cuisines


def _parse_hours(opening_hours: dict | None) -> dict:
    if not opening_hours:
        return {}
    hours: dict[str, Any] = {}
    if opening_hours.get("weekday_text"):
        for line in opening_hours["weekday_text"]:
            if ": " in line:
                day, times = line.split(": ", 1)
                hours[day.lower()] = times
            else:
                hours[line.lower()] = ""
    if opening_hours.get("open_now") is not None:
        hours["open_now"] = opening_hours["open_now"]
    return hours


def _extract_owners(lead_contact: dict, details: dict) -> list[str]:
    owners: list[str] = []
    name = (lead_contact.get("name") or "").strip()
    title = (lead_contact.get("job_title") or "").lower()
    if name and any(kw in title for kw in OWNER_TITLE_KEYWORDS):
        owners.append(name)
    summary = (details.get("editorial_summary") or {}).get("overview") or ""
    for match in re.findall(
        r"(?:owned by|founder|chef)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)?)", summary, re.I
    ):
        owners.append(match.strip())
    return list(dict.fromkeys(owners))


def _extract_emails_from_text(text: str) -> list[str]:
    found = []
    for raw in EMAIL_RE.findall(text):
        email = raw.rstrip(".")
        if any(x in email.lower() for x in ("example.com", "wixpress", "sentry", "schema.org")):
            continue
        found.append(email)
    return list(dict.fromkeys(found))


def find_email_from_website(website: str, cfg: Config) -> str:
    if not website:
        return ""
    if not website.startswith("http"):
        website = f"https://{website}"
    try:
        resp = requests.get(
            website,
            timeout=cfg.SCRAPE_TIMEOUT,
            headers={"User-Agent": "Mozilla/5.0 (compatible; TuviBot/1.0)"},
            allow_redirects=True,
        )
        if resp.status_code != 200:
            return ""
        emails = _extract_emails_from_text(resp.text)
        domain = website.split("//")[-1].split("/")[0].replace("www.", "")
        for email in emails:
            if domain.split(".")[0] in email.lower():
                return email
        return emails[0] if emails else ""
    except Exception as exc:
        log.warning(f"    Website email scrape failed: {exc}")
        return ""


def scrape_single_restaurant_places(
    lead_dict: dict,
    cfg: Config,
    max_reviews: int = 5,
) -> dict:
    """Scrape one restaurant via Google Places API (Maps data)."""
    lead = _lead_from_dict(lead_dict)
    company = lead_dict.get("company") or {}
    contact = lead_dict.get("contact") or {}
    location = lead_dict.get("location") or {}
    city = location.get("city") or lead.extra.get("city", "")
    company_name = company.get("name") or lead.company_name

    result: dict[str, Any] = {
        "name": company_name,
        "cuisines": [],
        "rating": None,
        "reviews_count": None,
        "contact": {
            "phone": "",
            "email": contact.get("email") or "",
            "website": company.get("domain") or "",
        },
        "owners": _extract_owners(contact, {}),
        "location": {
            "address": "",
            "city": city,
            "state": location.get("state", ""),
            "country": location.get("country", "Australia"),
            "coordinates": {},
        },
        "menu_items": [],
        "reviews": [],
        "images": {
            "thumbnail": "",
            "gallery": [],
            "menu_photos": [],
        },
        "hours": {},
        "price_level": "",
        "apollo_lead": lead_dict,
        "google": {
            "place_id": "",
            "maps_url": "",
        },
        "scrape_status": "pending",
        "errors": [],
    }

    try:
        place_id, _search_hit = search_place_id(company_name, city, cfg)
        if not place_id:
            result["scrape_status"] = "not_found"
            result["errors"].append("No Google Maps listing found (Places API)")
            return result

        time.sleep(cfg.SCRAPE_DELAY)
        details = get_place_details(place_id, cfg)
        if not details:
            result["scrape_status"] = "not_found"
            result["errors"].append("Place details empty")
            return result

        result["name"] = details.get("name") or company_name
        result["cuisines"] = _extract_cuisines_from_types(details.get("types") or [])
        result["rating"] = details.get("rating")
        result["reviews_count"] = details.get("user_ratings_total")
        result["contact"]["phone"] = (
            details.get("formatted_phone_number")
            or details.get("international_phone_number")
            or ""
        )
        result["contact"]["website"] = details.get("website") or result["contact"]["website"]
        result["price_level"] = PRICE_LEVEL_MAP.get(details.get("price_level"), "")
        result["hours"] = _parse_hours(details.get("opening_hours"))
        result["location"]["address"] = details.get("formatted_address") or ""
        geo = (details.get("geometry") or {}).get("location") or {}
        result["location"]["coordinates"] = {
            "latitude": geo.get("lat"),
            "longitude": geo.get("lng"),
        }
        result["google"]["place_id"] = details.get("place_id") or place_id
        result["google"]["maps_url"] = details.get("url") or ""

        photos = details.get("photos") or []
        for i, photo in enumerate(photos[:15]):
            ref = photo.get("photo_reference")
            if not ref:
                continue
            url = build_photo_url(ref, cfg)
            entry = {"title": f"Photo {i + 1}", "url": url}
            if i == 0:
                result["images"]["thumbnail"] = url
            result["images"]["gallery"].append(entry)

        for rev in (details.get("reviews") or [])[:max_reviews]:
            result["reviews"].append({
                "reviewer": rev.get("author_name", ""),
                "review": rev.get("text", ""),
                "stars": rev.get("rating"),
                "date": rev.get("relative_time_description", ""),
                "images": [],
                "source": "google_places_api",
            })

        result["owners"] = _extract_owners(contact, details)

        if not result["contact"]["email"] and result["contact"]["website"]:
            result["contact"]["email"] = find_email_from_website(result["contact"]["website"], cfg)

        result["scrape_status"] = "success"
        if not result["menu_items"]:
            result["errors"].append(
                "Menu items not available via Places API (use Playwright for menu)"
            )

    except Exception as exc:
        result["scrape_status"] = "error"
        result["errors"].append(str(exc))
        log.error(f"  Places scrape failed for {company_name}: {exc}")

    return result
