"""Google Places API (New) discovery and restaurant enrichment."""

from __future__ import annotations

import ipaddress
import logging
import re
import socket
import time
from typing import Any
from urllib.parse import quote, urljoin, urlsplit

import requests
from bs4 import BeautifulSoup

from tuvi_outreach_agent import Config, _lead_from_dict

try:
    from request_budget import BudgetExhausted, RequestBudget
except ImportError:
    BudgetExhausted = Exception  # type: ignore
    RequestBudget = None  # type: ignore

log = logging.getLogger("google_places")

DEFAULT_PLACES_V1_BASE = "https://places.googleapis.com/v1"
EMAIL_RE = re.compile(r"[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}")
OWNER_TITLE_KEYWORDS = ("owner", "founder", "co-founder", "proprietor", "managing director")
CONTACT_LINK_MARKERS = ("contact", "about", "team")
MAX_WEBSITE_BYTES = 1_000_000

# Metro rectangles intentionally constrain categorical Text Search results to the
# supported sales regions. Unknown cities fail closed instead of issuing an
# unbounded search that could consume quota outside the requested market.
CITY_SEARCH_BOUNDS: dict[str, dict[str, float]] = {
    "sydney": {"low_lat": -34.15, "low_lng": 150.55, "high_lat": -33.55, "high_lng": 151.40},
    "melbourne": {"low_lat": -38.15, "low_lng": 144.45, "high_lat": -37.45, "high_lng": 145.65},
    "perth": {"low_lat": -32.30, "low_lng": 115.55, "high_lat": -31.55, "high_lng": 116.20},
    "adelaide": {"low_lat": -35.20, "low_lng": 138.30, "high_lat": -34.55, "high_lng": 139.00},
    "brisbane": {"low_lat": -27.80, "low_lng": 152.70, "high_lat": -27.10, "high_lng": 153.45},
}

PLACE_TYPE_BY_NICHE = {
    "restaurant": "restaurant",
    "dentist": "dentist",
    "plumber": "plumber",
}

DISCOVERY_FIELD_MASK = ",".join([
    "places.id",
    "places.displayName",
    "places.formattedAddress",
    "places.location",
    "places.primaryType",
    "places.types",
    "places.businessStatus",
    "nextPageToken",
])

DETAIL_FIELD_MASK = ",".join([
    "id",
    "displayName",
    "formattedAddress",
    "addressComponents",
    "nationalPhoneNumber",
    "internationalPhoneNumber",
    "websiteUri",
    "googleMapsUri",
    "rating",
    "userRatingCount",
    "priceLevel",
    "types",
    "primaryType",
    "location",
    "reviews",
    "photos",
    "regularOpeningHours",
    "editorialSummary",
    "businessStatus",
])

GENERIC_TYPES = {
    "point_of_interest", "establishment", "food", "store",
    "restaurant", "meal_takeaway", "meal_delivery",
}

PRICE_LEVEL_MAP = {
    "PRICE_LEVEL_FREE": "",
    "PRICE_LEVEL_INEXPENSIVE": "$",
    "PRICE_LEVEL_MODERATE": "$$",
    "PRICE_LEVEL_EXPENSIVE": "$$$",
    "PRICE_LEVEL_VERY_EXPENSIVE": "$$$$",
}


class PlacesAPIError(ValueError):
    """A non-retryable Places 4xx response with a machine-readable status."""

    def __init__(self, message: str, *, status_code: int) -> None:
        super().__init__(message)
        self.status_code = status_code


def get_places_api_key(cfg: Config | None = None) -> str:
    cfg = cfg or Config()
    key = str(getattr(cfg, "PLACES_API_KEY", "") or "").strip()
    if not key:
        raise ValueError(
            "PLACES_API or GOOGLE_PLACES_API_KEY is missing from the configured environment. "
            "Enable Places API (New) in Google Cloud Console."
        )
    return key


def _places_base_url(cfg: Config) -> str:
    return str(getattr(cfg, "PLACES_API_BASE_URL", DEFAULT_PLACES_V1_BASE)).rstrip("/")


def _places_v1_request(
    method: str,
    path: str,
    cfg: Config,
    *,
    field_mask: str,
    label: str,
    budget: RequestBudget | None = None,
    body: dict | None = None,
) -> dict:
    api_key = get_places_api_key(cfg)
    url = f"{_places_base_url(cfg)}/{path.lstrip('/')}"
    headers = {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "X-Goog-Api-Key": api_key,
        "X-Goog-FieldMask": field_mask,
    }

    def _call() -> dict:
        if budget is not None:
            budget.consume(1)
        response = requests.request(
            method,
            url,
            headers=headers,
            json=body,
            timeout=max(cfg.SCRAPE_TIMEOUT, 20),
        )
        if response.status_code < 200 or response.status_code >= 300:
            message = f"Places API HTTP {response.status_code}: {response.text[:300]}"
            if response.status_code != 429 and response.status_code < 500:
                raise PlacesAPIError(message, status_code=response.status_code)
            raise RuntimeError(message)
        if not response.content:
            return {}
        return response.json()

    attempts = max(1, int(cfg.RETRY_ATTEMPTS))
    for attempt in range(1, attempts + 1):
        try:
            return _call()
        except BudgetExhausted:
            raise
        except ValueError:
            raise
        except (requests.RequestException, RuntimeError) as exc:
            if attempt >= attempts:
                raise
            wait = cfg.RETRY_BACKOFF ** (attempt - 1)
            log.warning(
                "[%s] Attempt %d/%d failed: %s — retrying in %.1fs",
                label,
                attempt,
                attempts,
                exc,
                wait,
            )
            time.sleep(wait)
    return {}


def _city_key(city: str) -> str:
    return city.split(",")[0].strip().lower().replace("_", " ")


def get_city_search_bounds(city: str) -> dict[str, float]:
    key = _city_key(city)
    bounds = CITY_SEARCH_BOUNDS.get(key)
    if bounds is None:
        supported = ", ".join(sorted(name.title() for name in CITY_SEARCH_BOUNDS))
        raise ValueError(f"Google Places discovery does not have safe bounds for {city!r}; supported: {supported}")
    return dict(bounds)


def _localized_text(value: Any) -> str:
    if isinstance(value, str):
        return value.strip()
    if isinstance(value, dict):
        return str(value.get("text") or "").strip()
    return ""


def discover_places_page(
    *,
    city: str,
    niche: str = "restaurant",
    bounds: dict[str, float] | None = None,
    page_token: str = "",
    page_size: int = 20,
    cfg: Config | None = None,
    budget: RequestBudget | None = None,
    label: str = "",
) -> tuple[list[dict], str]:
    """Fetch one durable Text Search page for a bounded grid cell.

    The caller owns persistence of ``page_token`` and the returned
    ``nextPageToken``. All search parameters other than the page controls stay
    stable so a persisted token can be reused safely while it remains valid.
    """
    cfg = cfg or Config()
    place_type = PLACE_TYPE_BY_NICHE.get((niche or "restaurant").strip().lower())
    if not place_type:
        raise ValueError(f"Google Places discovery does not support niche {niche!r}")
    if page_size < 1 or page_size > 20:
        raise ValueError("Places Text Search page_size must be between 1 and 20")

    cell_bounds = dict(bounds or get_city_search_bounds(city))
    city_label = city.split(",")[0].strip()
    query_label = "restaurants" if place_type == "restaurant" else f"{place_type}s"
    body: dict[str, Any] = {
        "textQuery": f"{query_label} in {city_label}, Australia",
        "includedType": place_type,
        "strictTypeFiltering": True,
        "locationRestriction": {
            "rectangle": {
                "low": {
                    "latitude": cell_bounds["low_lat"],
                    "longitude": cell_bounds["low_lng"],
                },
                "high": {
                    "latitude": cell_bounds["high_lat"],
                    "longitude": cell_bounds["high_lng"],
                },
            }
        },
        "languageCode": "en",
        "regionCode": "AU",
        "pageSize": page_size,
    }
    if page_token:
        body["pageToken"] = page_token

    data = _places_v1_request(
        "POST",
        "places:searchText",
        cfg,
        field_mask=DISCOVERY_FIELD_MASK,
        label=label or f"places-grid:{_city_key(city)}",
        budget=budget,
        body=body,
    )
    # Return the raw provider page so the durable worker can detect the 60-row
    # result cap accurately. Persistence filters permanently closed listings.
    places = list(data.get("places") or [])
    return places, str(data.get("nextPageToken") or "").strip()


def discover_places_for_city(
    *,
    city: str,
    niche: str = "restaurant",
    limit: int = 100,
    cfg: Config | None = None,
    budget: RequestBudget | None = None,
    known=None,
) -> list[dict]:
    """Discover new businesses through bounded, paginated Places Text Search."""
    if limit < 1:
        return []
    cfg = cfg or Config()
    place_type = PLACE_TYPE_BY_NICHE.get((niche or "restaurant").strip().lower())
    if not place_type:
        raise ValueError(f"Google Places discovery does not support niche {niche!r}")

    bounds = get_city_search_bounds(city)
    city_label = city.split(",")[0].strip()
    query_label = "restaurants" if place_type == "restaurant" else f"{place_type}s"
    base_body = {
        "textQuery": f"{query_label} in {city_label}, Australia",
        "includedType": place_type,
        "strictTypeFiltering": True,
        "locationRestriction": {
            "rectangle": {
                "low": {"latitude": bounds["low_lat"], "longitude": bounds["low_lng"]},
                "high": {"latitude": bounds["high_lat"], "longitude": bounds["high_lng"]},
            }
        },
        "languageCode": "en",
        "regionCode": "AU",
    }

    discovered: list[dict] = []
    seen_place_ids: set[str] = set()
    next_page_token = ""
    page = 0

    while len(discovered) < limit:
        if budget is not None and not budget.can_consume(1):
            break
        page += 1
        request_body = {
            **base_body,
            "pageSize": min(20, max(1, limit - len(discovered))),
        }
        if next_page_token:
            request_body["pageToken"] = next_page_token

        try:
            data = _places_v1_request(
                "POST",
                "places:searchText",
                cfg,
                field_mask=DISCOVERY_FIELD_MASK,
                label=f"places-discovery:{_city_key(city)}:p{page}",
                budget=budget,
                body=request_body,
            )
        except BudgetExhausted:
            break

        for place in data.get("places") or []:
            place_id = str(place.get("id") or "").strip()
            if not place_id or place_id in seen_place_ids:
                continue
            seen_place_ids.add(place_id)
            if str(place.get("businessStatus") or "").upper() == "CLOSED_PERMANENTLY":
                continue
            if known is not None and known.is_place_known(place_id):
                continue
            discovered.append(place)
            if len(discovered) >= limit:
                break

        next_page_token = str(data.get("nextPageToken") or "").strip()
        if not next_page_token:
            break
        time.sleep(cfg.SCRAPE_DELAY)

    return discovered


def place_to_lead_dict(place: dict, city: str, niche: str = "restaurant") -> dict:
    """Build the legacy-compatible input shape from a Places discovery record."""
    place_id = str(place.get("id") or "").strip()
    location = place.get("location") or {}
    name = _localized_text(place.get("displayName")) or "Unknown"
    city_label = city.split(",")[0].strip()
    return {
        "company": {"name": name, "domain": ""},
        "contact": {"name": "", "job_title": "", "email": ""},
        "location": {"city": city_label, "state": "", "country": "Australia"},
        "source": {"provider": "google_places_api_new", "place_id": place_id},
        "extra": {"city": city_label, "lead_type": niche},
        "google": {
            "place_id": place_id,
            "formatted_address": str(place.get("formattedAddress") or ""),
            "latitude": location.get("latitude"),
            "longitude": location.get("longitude"),
        },
    }


def search_place_id(
    company_name: str,
    city: str,
    cfg: Config,
    *,
    query_suffix: str = "restaurant",
    budget: RequestBudget | None = None,
) -> tuple[str | None, dict]:
    """Resolve a known business name through Places API (New) Text Search."""
    body: dict[str, Any] = {
        "textQuery": f"{company_name} {query_suffix} {city} Australia",
        "pageSize": 1,
        "languageCode": "en",
        "regionCode": "AU",
    }
    try:
        bounds = get_city_search_bounds(city)
        body["locationRestriction"] = {
            "rectangle": {
                "low": {"latitude": bounds["low_lat"], "longitude": bounds["low_lng"]},
                "high": {"latitude": bounds["high_lat"], "longitude": bounds["high_lng"]},
            }
        }
    except ValueError:
        pass

    try:
        data = _places_v1_request(
            "POST",
            "places:searchText",
            cfg,
            field_mask="places.id,places.displayName,places.formattedAddress",
            label=f"places-search:{company_name}",
            budget=budget,
            body=body,
        )
    except BudgetExhausted:
        raise
    places = data.get("places") or []
    if not places:
        return None, {}
    top = places[0]
    return str(top.get("id") or "").strip() or None, top


def get_place_details(place_id: str, cfg: Config, budget: RequestBudget | None = None) -> dict:
    """Retrieve contact, hours, reviews and safe photo resource metadata."""
    try:
        return _places_v1_request(
            "GET",
            f"places/{quote(place_id, safe='')}",
            cfg,
            field_mask=DETAIL_FIELD_MASK,
            label=f"places-details:{place_id[:12]}",
            budget=budget,
        )
    except BudgetExhausted:
        raise


def _extract_cuisines_from_types(types: list[str]) -> list[str]:
    cuisines = []
    for place_type in types or []:
        label = place_type.replace("_", " ").strip()
        if place_type.lower() in GENERIC_TYPES:
            continue
        cuisines.append(label.title())
    return cuisines


def _parse_hours(opening_hours: dict | None) -> dict:
    if not opening_hours:
        return {}
    hours: dict[str, Any] = {}
    descriptions = opening_hours.get("weekdayDescriptions") or opening_hours.get("weekday_text") or []
    for line in descriptions:
        if ": " in line:
            day, times = line.split(": ", 1)
            hours[day.lower()] = times
        else:
            hours[line.lower()] = ""
    open_now = opening_hours.get("openNow")
    if open_now is None:
        open_now = opening_hours.get("open_now")
    if open_now is not None:
        hours["open_now"] = open_now
    return hours


def _extract_address_component(details: dict, wanted_type: str) -> str:
    for component in details.get("addressComponents") or []:
        if wanted_type in (component.get("types") or []):
            return str(component.get("longText") or component.get("shortText") or "").strip()
    return ""


def _extract_owners(lead_contact: dict, details: dict) -> list[str]:
    owners: list[str] = []
    name = (lead_contact.get("name") or "").strip()
    title = (lead_contact.get("job_title") or "").lower()
    if name and any(keyword in title for keyword in OWNER_TITLE_KEYWORDS):
        owners.append(name)
    summary = _localized_text(details.get("editorialSummary"))
    for match in re.findall(
        r"(?:owned by|founder|chef)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)?)",
        summary,
        re.I,
    ):
        owners.append(match.strip())
    return list(dict.fromkeys(owners))


def _extract_emails_from_text(text: str) -> list[str]:
    found = []
    for raw in EMAIL_RE.findall(text):
        email = raw.rstrip(".")
        if any(value in email.lower() for value in ("example.com", "wixpress", "sentry", "schema.org")):
            continue
        found.append(email)
    return list(dict.fromkeys(found))


def _is_public_http_url(raw_url: str) -> bool:
    try:
        parsed = urlsplit(raw_url)
    except ValueError:
        return False
    if parsed.scheme not in ("http", "https") or not parsed.hostname:
        return False
    if parsed.username or parsed.password:
        return False
    try:
        addresses = socket.getaddrinfo(parsed.hostname, parsed.port or (443 if parsed.scheme == "https" else 80))
    except socket.gaierror:
        return False
    for address in addresses:
        try:
            ip = ipaddress.ip_address(address[4][0])
        except ValueError:
            return False
        if not ip.is_global:
            return False
    return True


def _fetch_public_html(raw_url: str, cfg: Config) -> tuple[str, str]:
    current = raw_url
    for _ in range(4):
        if not _is_public_http_url(current):
            return "", ""
        response = requests.get(
            current,
            timeout=cfg.SCRAPE_TIMEOUT,
            headers={"User-Agent": "Mozilla/5.0 (compatible; TuviBot/1.0)"},
            allow_redirects=False,
            stream=True,
        )
        if response.status_code in (301, 302, 303, 307, 308):
            destination = response.headers.get("Location") or ""
            response.close()
            if not destination:
                return "", ""
            current = urljoin(current, destination)
            continue
        if response.status_code != 200:
            response.close()
            return "", ""
        content_type = (response.headers.get("Content-Type") or "").lower()
        if content_type and "html" not in content_type and not content_type.startswith("text/"):
            response.close()
            return "", ""
        chunks: list[bytes] = []
        total = 0
        for chunk in response.iter_content(chunk_size=16_384):
            if not chunk:
                continue
            remaining = MAX_WEBSITE_BYTES - total
            if remaining <= 0:
                break
            chunks.append(chunk[:remaining])
            total += min(len(chunk), remaining)
        encoding = response.encoding or "utf-8"
        response.close()
        return b"".join(chunks).decode(encoding, errors="replace"), current
    return "", ""


def find_email_from_website(website: str, cfg: Config) -> str:
    """Extract a public business email while blocking private-network requests."""
    website = (website or "").strip()
    if not website:
        return ""
    if not website.startswith(("http://", "https://")):
        website = f"https://{website}"

    pending = [website]
    visited: set[str] = set()
    candidate_emails: list[str] = []
    base_host = (urlsplit(website).hostname or "").lower().removeprefix("www.")

    try:
        while pending and len(visited) < 3:
            target = pending.pop(0)
            if target in visited:
                continue
            visited.add(target)
            html, final_url = _fetch_public_html(target, cfg)
            if not html:
                continue
            candidate_emails.extend(_extract_emails_from_text(html))
            if candidate_emails:
                break

            soup = BeautifulSoup(html, "html.parser")
            for link in soup.find_all("a", href=True):
                href = str(link.get("href") or "").strip()
                label = f"{href} {link.get_text(' ', strip=True)}".lower()
                if not any(marker in label for marker in CONTACT_LINK_MARKERS):
                    continue
                absolute = urljoin(final_url, href)
                host = (urlsplit(absolute).hostname or "").lower().removeprefix("www.")
                if host == base_host and absolute not in visited and absolute not in pending:
                    pending.append(absolute)
                if len(pending) >= 2:
                    break
    except Exception as exc:
        log.warning("Website email scrape failed for host %s: %s", base_host or "unknown", exc)
        return ""

    for email in candidate_emails:
        if base_host and base_host.split(".")[0] in email.lower():
            return email
    return candidate_emails[0] if candidate_emails else ""


def scrape_single_restaurant_places(
    lead_dict: dict,
    cfg: Config,
    max_reviews: int = 5,
    *,
    query_suffix: str = "restaurant",
    budget: RequestBudget | None = None,
    lookup_website_email: bool = True,
) -> dict:
    """Enrich one discovered business through Places API (New)."""
    lead = _lead_from_dict(lead_dict)
    company = lead_dict.get("company") or {}
    contact = lead_dict.get("contact") or {}
    location = lead_dict.get("location") or {}
    discovered_google = lead_dict.get("google") or {}
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
            "address": discovered_google.get("formatted_address") or "",
            "city": city,
            "state": location.get("state", ""),
            "country": location.get("country", "Australia"),
            "coordinates": {
                "latitude": discovered_google.get("latitude"),
                "longitude": discovered_google.get("longitude"),
            },
        },
        "menu_items": [],
        "reviews": [],
        "images": {
            "thumbnail": "",
            "gallery": [],
            "menu_photos": [],
            "google_photo_count": 0,
            # Photo resource names expire and are fetched again just before
            # OCR; they must not be cached in durable lead data.
            "google_photos": [],
        },
        "hours": {},
        "price_level": "",
        "apollo_lead": {},
        "google": {
            "place_id": str(discovered_google.get("place_id") or "").strip(),
            "maps_url": "",
            "source": "places_api_new",
        },
        "scrape_status": "pending",
        "errors": [],
    }

    try:
        place_id = result["google"]["place_id"]
        if not place_id:
            place_id, _ = search_place_id(
                company_name,
                city,
                cfg,
                query_suffix=query_suffix,
                budget=budget,
            )
        if not place_id:
            result["scrape_status"] = "not_found"
            result["errors"].append("No Google Maps listing found (Places API New)")
            return result

        details = get_place_details(place_id, cfg, budget=budget)
        if not details:
            result["scrape_status"] = "not_found"
            result["errors"].append("Place details empty")
            return result

        result["name"] = _localized_text(details.get("displayName")) or company_name
        result["cuisines"] = _extract_cuisines_from_types(details.get("types") or [])
        result["rating"] = details.get("rating")
        result["reviews_count"] = details.get("userRatingCount")
        result["contact"]["phone"] = (
            details.get("nationalPhoneNumber")
            or details.get("internationalPhoneNumber")
            or ""
        )
        result["contact"]["website"] = details.get("websiteUri") or result["contact"]["website"]
        result["price_level"] = PRICE_LEVEL_MAP.get(str(details.get("priceLevel") or ""), "")
        result["hours"] = _parse_hours(details.get("regularOpeningHours"))
        result["location"]["address"] = details.get("formattedAddress") or result["location"]["address"]
        result["location"]["state"] = _extract_address_component(details, "administrative_area_level_1") or result["location"]["state"]
        result["location"]["country"] = _extract_address_component(details, "country") or result["location"]["country"]
        geo = details.get("location") or {}
        result["location"]["coordinates"] = {
            "latitude": geo.get("latitude"),
            "longitude": geo.get("longitude"),
        }
        result["google"]["place_id"] = details.get("id") or place_id
        result["google"]["maps_url"] = details.get("googleMapsUri") or ""
        result["google"]["business_status"] = details.get("businessStatus") or ""

        result["images"]["google_photo_count"] = len(details.get("photos") or [])

        for review in (details.get("reviews") or [])[:max_reviews]:
            author = review.get("authorAttribution") or {}
            result["reviews"].append({
                "reviewer": author.get("displayName") or "",
                "review": _localized_text(review.get("originalText")) or _localized_text(review.get("text")),
                "stars": review.get("rating"),
                "date": review.get("relativePublishTimeDescription") or review.get("publishTime") or "",
                "images": [],
                "source": "google_places_api_new",
            })

        result["owners"] = _extract_owners(contact, details)
        if (
            lookup_website_email
            and not result["contact"]["email"]
            and result["contact"]["website"]
        ):
            result["contact"]["email"] = find_email_from_website(result["contact"]["website"], cfg)

        result["scrape_status"] = "success"
        if not result["menu_items"]:
            result["errors"].append("Menu items are not provided by Places API")

    except BudgetExhausted:
        result["scrape_status"] = "budget_exhausted"
        result["errors"].append("Request budget exhausted during Places scrape")
    except PlacesAPIError as exc:
        result["scrape_status"] = "not_found" if exc.status_code == 404 else "permanent_error"
        result["provider_error_code"] = exc.status_code
        result["errors"].append(str(exc))
        log.error("Places rejected %s with HTTP %d", company_name, exc.status_code)
    except Exception as exc:
        result["scrape_status"] = "error"
        result["errors"].append(str(exc))
        log.error("Places scrape failed for %s: %s", company_name, exc)

    return result
