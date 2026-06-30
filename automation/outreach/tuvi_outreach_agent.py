"""
============================================================
  TUVI SOLUTIONS — Lead Generation & Outreach Agent  v2.0
============================================================
  Author  : Tuvi Solutions — Lead Generation & Outreach Lead
  Version : 2.0.0

  Pipeline per lead:
    1. Load leads  ──  CSV | Google Sheets | inline dict
    2. Scrape      ──  Company website (async, up to 9 paths)
    3. Synthesise  ──  Claude 3.5 Sonnet → structured JSON draft
    4. Draft       ──  Zoho Mail saveDraft (NEVER auto-send)
    5. Write-back  ──  Google Sheets status + draft URL
    6. Notify      ──  Slack review card

  New in v2:
    • Async / concurrent processing (ThreadPoolExecutor)
    • Google Sheets write-back (status, draft URL, confidence, timestamp)
    • Retry logic with exponential back-off for all external calls
    • Per-lead error isolation — one bad lead never kills the batch
    • Rate-limit guard (max N concurrent workers, configurable)
    • JSON + CSV summary report written after every run
============================================================
"""

# ─────────────────────────────────────────────────────────────
# 0. STDLIB
# ─────────────────────────────────────────────────────────────
import os
import re
import csv
import json
import time
import logging
import argparse
import textwrap
import traceback
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime
from pathlib import Path
from typing import Optional

# ─────────────────────────────────────────────────────────────
# 1. THIRD-PARTY  (pip install -r requirements.txt)
# ─────────────────────────────────────────────────────────────
import requests
from bs4 import BeautifulSoup
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

try:
    from openai import OpenAI
except ImportError:
    OpenAI = None

try:
    from notion_client import Client as NotionClient
except ImportError:
    NotionClient = None

try:
    from slack_sdk import WebClient as SlackClient
    from slack_sdk.errors import SlackApiError
except ImportError:
    SlackClient = None


# ─────────────────────────────────────────────────────────────
# 2. LOGGING
# ─────────────────────────────────────────────────────────────
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s  [%(levelname)s]  %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
log = logging.getLogger("tuvi_outreach")


# ═════════════════════════════════════════════════════════════
# 3. CONFIGURATION
# ═════════════════════════════════════════════════════════════
class Config:
    # ── OpenAI ─────────────────────────────────────────
    OPENAI_API_KEY: str           = os.getenv("OPENAI_API_KEY", "")
    OPENAI_MODEL: str             = "gpt-4o"
    OPENAI_MAX_TOKENS: int        = 3000

    # ── Zoho Mail ─────────────────────────────────────────
    ZOHO_ACCOUNT_ID: str          = os.getenv("ZOHO_ACCOUNT_ID", "")
    ZOHO_ACCESS_TOKEN: str        = os.getenv("ZOHO_ACCESS_TOKEN", "")
    ZOHO_REFRESH_TOKEN: str       = os.getenv("ZOHO_REFRESH_TOKEN", "")
    ZOHO_CLIENT_ID: str           = os.getenv("ZOHO_CLIENT_ID", "")
    ZOHO_CLIENT_SECRET: str       = os.getenv("ZOHO_CLIENT_SECRET", "")
    ZOHO_FROM_EMAIL: str          = os.getenv("ZOHO_FROM_EMAIL", "outreach@tuvisolutions.com")
    ZOHO_BASE_URL: str            = "https://mail.zoho.com/api/accounts"
    ZOHO_REGION: str              = os.getenv("ZOHO_REGION", "com")  # com | com.au | eu | in

    # ── Slack ─────────────────────────────────────────────
    SLACK_BOT_TOKEN: str          = os.getenv("SLACK_BOT_TOKEN", "")
    SLACK_REVIEW_CHANNEL: str     = os.getenv("SLACK_REVIEW_CHANNEL", "#outreach-review")

    # ── Notion Tracker ────────────────────────────────────
    NOTION_TOKEN: str = os.getenv("NOTION_TOKEN", "")
    NOTION_DATABASE_ID: str = os.getenv("NOTION_DATABASE_ID", "")

    # ── Scraping ──────────────────────────────────────────
    SCRAPE_TIMEOUT: int           = 15
    SCRAPE_DELAY: float           = 1.0        # polite delay between pages (seconds)
    SCRAPE_MAX_CHARS: int         = 8000

    # ── Concurrency ───────────────────────────────────────
    MAX_WORKERS: int              = 3          # parallel leads processed simultaneously
    RETRY_ATTEMPTS: int           = 3          # retries for external API calls
    RETRY_BACKOFF: float          = 2.0        # exponential back-off multiplier

    # ── Output ────────────────────────────────────────────
    OUTPUT_DIR: str               = "drafts"
    REPORT_DIR: str               = "reports"
    LEADS_DIR: str                = "leads"
    LEADS_FILE: str               = "lead.json"

    # ── APIs ──────────────────────────────────────────────
    APOLLO_API_KEY: str           = os.getenv("APOLLO_API_KEY", "")
    SERPAPI_KEY: str              = os.getenv("SERPAPI_KEY", "")
    PLACES_API_KEY: str           = os.getenv("PLACES_API", "") or os.getenv("GOOGLE_PLACES_API_KEY", "")


# Major Australian cities for restaurant lead sourcing
AUSTRALIAN_RESTAURANT_CITIES: list[str] = [
    "Sydney, Australia",
    "Melbourne, Australia",
    "Perth, Australia",
    "Adelaide, Australia",
    "Brisbane, Australia",
]

AUSTRALIAN_CITY_ALIASES: dict[str, str] = {
    "sydney": "Sydney, Australia",
    "melbourne": "Melbourne, Australia",
    "perth": "Perth, Australia",
    "adelaide": "Adelaide, Australia",
    "brisbane": "Brisbane, Australia",
    "melborney": "Melbourne, Australia",  # common typo
}


def normalize_city_name(city: str) -> str:
    """Convert 'sydney' / 'Sydney' / 'Sydney, Australia' → 'Sydney, Australia'."""
    city = city.strip()
    if not city:
        raise ValueError("City name cannot be empty")

    if "," in city:
        return city

    key = city.lower().replace("_", " ")
    if key in AUSTRALIAN_CITY_ALIASES:
        return AUSTRALIAN_CITY_ALIASES[key]

    for alias, full in AUSTRALIAN_CITY_ALIASES.items():
        if key == alias:
            return full

    return f"{city.title()}, Australia"


def resolve_cities(
    city: str | None = None,
    cities: list[str] | None = None,
) -> list[str]:
    """
    Resolve CLI city filters for Apollo fetch.
      --city Sydney        → ["Sydney, Australia"]
      --cities Sydney Perth → ["Sydney, Australia", "Perth, Australia"]
      (none)               → all 5 default cities
    """
    if city and cities:
        raise ValueError("Use either --city OR --cities, not both")

    if city:
        return [normalize_city_name(city)]

    if cities:
        return [normalize_city_name(c) for c in cities]

    return list(AUSTRALIAN_RESTAURANT_CITIES)


def city_slug(city: str) -> str:
    """'Sydney, Australia' → 'sydney'"""
    return city.split(",")[0].strip().lower().replace(" ", "_")


def default_leads_output_path(
    cities: list[str],
    cfg: Config | None = None,
) -> Path:
    """
    Auto-name lead JSON by city:
      all 5 cities  → leads/lead.json
      one city      → leads/lead_sydney.json
      multiple      → leads/lead_sydney_perth.json
    """
    cfg = cfg or Config()
    if len(cities) == 1:
        return Path(cfg.LEADS_DIR) / f"lead_{city_slug(cities[0])}.json"
    if set(cities) == set(AUSTRALIAN_RESTAURANT_CITIES):
        return Path(cfg.LEADS_DIR) / cfg.LEADS_FILE
    slugs = "_".join(city_slug(c) for c in cities)
    return Path(cfg.LEADS_DIR) / f"lead_{slugs}.json"


def default_scrape_output_path(city: str | None, cfg: Config | None = None) -> Path:
    """data/restaurants_data_sydney.json for single-city scrape."""
    cfg = cfg or Config()
    if city:
        return Path("data") / f"restaurants_data_{city_slug(normalize_city_name(city))}.json"
    return Path("data") / "restaurants_data.json"

RESTAURANT_KEYWORD_TAGS: list[str] = [
    "restaurant",
    "restaurants",
    "dining",
    "food and beverages",
    "hospitality",
    "cafe",
    "bistro",
]

RESTAURANT_DECISION_MAKER_TITLES: list[str] = [
    "owner",
    "founder",
    "co-founder",
    "general manager",
    "restaurant manager",
    "operations manager",
    "director",
    "managing director",
    "proprietor",
    "head chef",
]


# ═════════════════════════════════════════════════════════════
# 4. SYSTEM PROMPT  (OpenAI)
# ═════════════════════════════════════════════════════════════
SYSTEM_PROMPT = """\
You are an expert Restaurant Data Analyst and Tech Outreach Specialist for Tuvi Solutions.

TASK 1: DATA EXTRACTION
Analyze the scraped website content and extract the detailed restaurant information based on the JSON structure provided below. 
IMPORTANT: If you cannot find specific information for a field, DO NOT leave it blank. You must use a standard option or placeholder (e.g., "Standard Modern Font", "Light Mode", "Standard Opening Hours", "Menu available via link", or "Not publicly listed"). Look at [IMAGE: ...] and [LINK: ...] tags to find logos, dishes, and booking integrations.

TASK 2: OUTREACH DRAFT
Write ONE highly personalised cold-email draft pitching Tuvi Solutions' custom software, app development, or AI/ML integrations.
- Hook: Reference a specific dish, review, or vibe from the data you extracted.
- Bridge: Connect it to a need for custom tech (e.g., building a proprietary ordering app to avoid UberEats fees, digitizing their menus, or automating reservations).
- CTA: 10-minute discovery call.
- Tone: Professional, warm. Max 180 words. Sign off as "Sri Harsha \\n Tuvi Solutions".

OUTPUT FORMAT: You MUST respond ONLY with a valid JSON object matching this exact structure:
{
  "subject": "<email subject line>",
  "body": "<email body>",
  "confidence": "<high|medium|low>",
  "restaurant_data": {
    "1_brand_identity": {
      "restaurant_name": "...",
      "cuisine_type": "...",
      "logos_and_favicons": "...",
      "theme_colours": "...",
      "typography": "...",
      "tagline": "..."
    },
    "2_business_operations": {
      "address": "...",
      "maps_link": "...",
      "contact_info": "...",
      "internal_email": "...",
      "operating_hours": "...",
      "dietary_capabilities": "..."
    },
    "3_personnel_and_story": {
      "ownership": "...",
      "chef_name": "...",
      "bio_story": "..."
    },
    "4_marketing_content": {
      "best_dishes": "...",
      "customer_reviews": "...",
      "hero_text": "..."
    },
    "5_menu_structure": {
      "delivery_method": "...",
      "interactive_menu_details": "..."
    },
    "6_integrations_and_links": {
      "reservation_link": "...",
      "delivery_links": "...",
      "social_media": "..."
    },
    "7_media_assets": {
      "hero_asset": "...",
      "atmosphere_photos": "...",
      "action_photos": "..."
    },
    "8_technical_metadata": {
      "custom_domain": "...",
      "meta_title_description": "...",
      "analytics_ids": "..."
    }
  }
}
"""


# ═════════════════════════════════════════════════════════════
# 5. DATA MODELS
# ═════════════════════════════════════════════════════════════
class Lead:
    """A single prospect row from any source."""
    def __init__(
        self,
        company_name: str,
        domain: str,
        contact_name: str,
        job_title: str,
        contact_email: str = "",
        sheet_row_index: Optional[int] = None,   # 1-based row in Google Sheet
        extra: dict | None = None,
    ):
        self.company_name    = company_name.strip()
        self.domain          = domain.strip().rstrip("/")
        self.contact_name    = contact_name.strip()
        self.first_name      = contact_name.strip().split()[0] if contact_name.strip() else "there"
        self.job_title       = job_title.strip()
        self.contact_email   = contact_email.strip()
        self.sheet_row_index = sheet_row_index
        self.extra           = extra or {}

    def __repr__(self):
        return f"<Lead {self.company_name} | {self.contact_name} | {self.domain}>"


class OutreachResult:
    """Full outcome for a single lead."""
    def __init__(self, lead: Lead):
        self.lead              = lead
        self.scraped_text: str  = ""
        self.business_model: str = ""
        self.subject: str        = ""
        self.body: str           = ""
        self.confidence: str     = ""
        self.notes: str          = ""
        self.restaurant_data: dict = {}  # <-- ADD THIS LINE
        self.zoho_draft_id: str  = ""
        self.zoho_draft_url: str = ""
        self.error: str          = ""
        self.timestamp: str      = datetime.utcnow().isoformat()

    @property
    def status(self) -> str:
        if self.error:
            return "error"
        if self.zoho_draft_id:
            return "draft_created"
        return "local_only"

    def to_dict(self) -> dict:
        return {
            "company"      : self.lead.company_name,
            "contact"      : self.lead.contact_name,
            "job_title"    : self.lead.job_title,
            "email"        : self.lead.contact_email,
            "domain"       : self.lead.domain,
            "business_model": self.business_model,
            "subject"      : self.subject,
            "confidence"   : self.confidence,
            "status"       : self.status,
            "draft_url"    : self.zoho_draft_url,
            "notes"        : self.notes,
            "error"        : self.error,
            "timestamp"    : self.timestamp,
            "restaurant_data": self.restaurant_data  # <-- Fixes the JSON report inclusion
        }


# ═════════════════════════════════════════════════════════════
# 6. UTILITY — RETRY DECORATOR
# ═════════════════════════════════════════════════════════════
def with_retry(fn, attempts: int, backoff: float, label: str = ""):
    """
    Call fn() up to `attempts` times with exponential back-off.
    Raises the last exception if all attempts fail.
    """
    last_exc = None
    for attempt in range(1, attempts + 1):
        try:
            return fn()
        except Exception as exc:
            last_exc = exc
            wait = backoff ** (attempt - 1)
            log.warning(
                f"  [{label}] Attempt {attempt}/{attempts} failed: {exc}"
                + (f" — retrying in {wait:.1f}s" if attempt < attempts else " — giving up")
            )
            if attempt < attempts:
                time.sleep(wait)
    raise last_exc


# ═════════════════════════════════════════════════════════════
# 7. LEAD SOURCES
# ═════════════════════════════════════════════════════════════

def load_leads_from_csv(filepath: str) -> list[Lead]:
    """Load leads from a CSV file (Apollo.io export compatible)."""
    leads = []
    path  = Path(filepath)
    if not path.exists():
        log.error(f"CSV not found: {filepath}")
        return leads

    with open(path, newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        reader.fieldnames = [h.strip().lower().replace(" ", "_") for h in (reader.fieldnames or [])]
        for i, row in enumerate(reader, start=2):  # start=2 → row 1 is header
            try:
                leads.append(Lead(
                    company_name    = row.get("company_name", ""),
                    domain          = row.get("domain", row.get("website", "")),
                    contact_name    = row.get("contact_name", row.get("name", "")),
                    job_title       = row.get("job_title", row.get("title", "")),
                    contact_email   = row.get("contact_email", row.get("email", "")),
                    sheet_row_index = i,
                    extra           = dict(row),
                ))
            except Exception as e:
                log.warning(f"  Skipping row {i}: {e}")

    log.info(f"Loaded {len(leads)} leads from CSV: {filepath}")
    return leads


def lead_to_dict(lead: Lead) -> dict:
    """Serialize a Lead to a clean, nested JSON-friendly structure."""
    city = lead.extra.get("city", "")
    return {
        "company": {
            "name": lead.company_name,
            "domain": lead.domain,
            "industry": lead.extra.get("organization_industry", ""),
        },
        "contact": {
            "name": lead.contact_name,
            "job_title": lead.job_title,
            "email": lead.contact_email,
            "has_email": bool(lead.extra.get("has_email", bool(lead.contact_email))),
        },
        "location": {
            "city": city,
            "state": lead.extra.get("state", ""),
            "country": lead.extra.get("country", "Australia"),
        },
        "source": {
            "provider": "apollo.io",
            "person_id": lead.extra.get("apollo_id", ""),
        },
    }


def _lead_from_dict(data: dict) -> Lead:
    """Reconstruct a Lead from a nested JSON lead object."""
    company = data.get("company") or {}
    contact = data.get("contact") or {}
    location = data.get("location") or {}
    source = data.get("source") or {}

    # Flat legacy format (company_name at top level)
    if not company and data.get("company_name"):
        company = {
            "name": data.get("company_name", ""),
            "domain": data.get("domain", ""),
            "industry": data.get("organization_industry", ""),
        }
        contact = {
            "name": data.get("contact_name", ""),
            "job_title": data.get("job_title", ""),
            "email": data.get("contact_email", data.get("email", "")),
            "has_email": data.get("has_email", False),
        }
        location = {"city": data.get("city", ""), "country": "Australia"}

    return Lead(
        company_name=company.get("name", ""),
        domain=company.get("domain", ""),
        contact_name=contact.get("name", ""),
        job_title=contact.get("job_title", ""),
        contact_email=contact.get("email", ""),
        extra={
            "city": location.get("city", ""),
            "state": location.get("state", ""),
            "country": location.get("country", "Australia"),
            "organization_industry": company.get("industry", ""),
            "has_email": contact.get("has_email", False),
            "apollo_id": source.get("person_id", ""),
        },
    )


def build_leads_document(
    leads: list[Lead],
    cities: list[str],
    per_page: int,
    max_pages: int,
    target_per_city: int | None = None,
) -> dict:
    """Build the full lead.json document with metadata and nested lead records."""
    leads_by_city: dict[str, int] = {}
    for lead in leads:
        city = lead.extra.get("city", "Unknown")
        leads_by_city[city] = leads_by_city.get(city, 0) + 1

    return {
        "meta": {
            "version": "1.0",
            "fetched_at": datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ"),
            "source": "apollo.io",
            "lead_type": "restaurant",
            "total_leads": len(leads),
            "cities_searched": cities,
            "leads_by_city": dict(sorted(leads_by_city.items())),
            "fetch_settings": {
                "per_page": per_page,
                "max_pages": max_pages,
                "target_per_city": target_per_city,
            },
        },
        "leads": [lead_to_dict(lead) for lead in leads],
    }


def save_leads_to_json(
    leads: list[Lead],
    filepath: str | Path,
    cities: list[str] | None = None,
    per_page: int = 100,
    max_pages: int = 20,
    target_per_city: int | None = 100,
) -> Path:
    """Save fetched leads to a well-formatted JSON file."""
    path = Path(filepath)
    path.parent.mkdir(parents=True, exist_ok=True)

    document = build_leads_document(
        leads=leads,
        cities=cities or AUSTRALIAN_RESTAURANT_CITIES,
        per_page=per_page,
        max_pages=max_pages,
        target_per_city=target_per_city,
    )

    path.write_text(
        json.dumps(document, indent=2, ensure_ascii=False) + "\n",
        encoding="utf-8",
    )
    log.info(f"Saved {len(leads)} leads to {path}")
    return path


def load_leads_from_json(filepath: str | Path) -> list[Lead]:
    """Load leads from lead.json (nested format) or a plain leads array."""
    path = Path(filepath)
    if not path.exists():
        log.error(f"JSON not found: {filepath}")
        return []

    data = json.loads(path.read_text(encoding="utf-8"))
    raw_leads = data.get("leads", data) if isinstance(data, dict) else data

    if not isinstance(raw_leads, list):
        log.error(f"Invalid JSON structure in {filepath}: expected 'leads' array")
        return []

    leads = []
    for i, item in enumerate(raw_leads, start=1):
        try:
            leads.append(_lead_from_dict(item))
        except Exception as e:
            log.warning(f"  Skipping lead #{i}: {e}")

    log.info(f"Loaded {len(leads)} leads from JSON: {filepath}")
    return leads


# ── Notion Tracker ─────────────────────────────────────────

def write_to_notion(result: OutreachResult, cfg: Config, notion_client) -> None:
    """Log the outreach result as a new row in the Notion database."""
    if not notion_client or not cfg.NOTION_DATABASE_ID:
        return

    try:
        # Build the properties map matching the exact layout in your new Notion setup
        properties = {
            "Company": {"title": [{"text": {"content": result.lead.company_name}}]},
            "Contact Name": {"rich_text": [{"text": {"content": result.lead.contact_name}}]},
        }
        
        if result.lead.contact_email:
            properties["Email"] = {"email": result.lead.contact_email}
            
        # Updated to capital "Status"
        if result.status:
            properties["Status"] = {"select": {"name": result.status}}
            
        # Updated to capital "Draft URL"
        if result.zoho_draft_url:
            properties["Draft URL"] = {"url": result.zoho_draft_url}
            
        # "Confidence" remains capitalized
        if result.confidence:
            properties["Confidence"] = {"select": {"name": result.confidence.capitalize()}}

        notion_client.pages.create(
            parent={"database_id": cfg.NOTION_DATABASE_ID},
            properties=properties
        )
        log.info(f"  ✅ Successfully logged {result.lead.company_name} to Notion.")

    except Exception as e:
        log.warning(f"  ❌ Failed to write {result.lead.company_name} to Notion: {e}")


def load_leads_inline(data: list[dict]) -> list[Lead]:
    """For programmatic / API usage."""
    leads = []
    for row in data:
        norm = {k.strip().lower().replace(" ", "_"): str(v) for k, v in row.items()}
        leads.append(Lead(
            company_name  = norm.get("company_name", ""),
            domain        = norm.get("domain", norm.get("website", "")),
            contact_name  = norm.get("contact_name", norm.get("name", "")),
            job_title     = norm.get("job_title", norm.get("title", "")),
            contact_email = norm.get("contact_email", norm.get("email", "")),
        ))
    return leads


def _lead_dedup_key(lead: Lead) -> str:
    """Unique key for deduplicating leads across city searches."""
    if lead.contact_email:
        return lead.contact_email.lower()
    return f"{lead.company_name}|{lead.contact_name}|{lead.domain}".lower()


def _apollo_person_to_lead(person: dict, city: str) -> Lead | None:
    """Map one Apollo person record to a Lead, or None if clearly not a restaurant."""
    org = person.get("organization") or {}
    company_name = (org.get("name") or "").strip()
    if not company_name:
        return None

    org_keywords = " ".join([
        company_name,
        org.get("industry") or "",
        org.get("short_description") or "",
    ]).lower()

    restaurant_signals = ("restaurant", "dining", "cafe", "café", "bistro", "eatery", "food", "grill", "kitchen", "pizza", "sushi", "bar")
    non_restaurant = ("software", "real estate", "law firm", "accounting", "insurance")

    if any(ex in org_keywords for ex in non_restaurant):
        return None

    # Apollo query already uses restaurant keyword tags — accept unless clearly excluded
    if not any(signal in org_keywords for signal in restaurant_signals):
        if "hospitality" not in org_keywords and "catering" not in org_keywords:
            # Still allow if Apollo tagged it (no industry data on record)
            if org.get("industry"):
                return None

    city_label = city.split(",")[0].strip()
    return Lead(
        company_name=company_name,
        domain=org.get("primary_domain") or org.get("website_url") or "",
        contact_name=f"{person.get('first_name', '')} {person.get('last_name', '')}".strip(),
        job_title=person.get("title") or "Unknown Title",
        contact_email=person.get("email") or "",
        extra={
            "city": city_label,
            "apollo_id": person.get("id", ""),
            "has_email": person.get("has_email", False),
            "organization_industry": org.get("industry", ""),
        },
    )


def _fetch_restaurant_leads_for_city(
    api_key: str,
    city: str,
    per_page: int = 100,
    max_pages: int = 20,
    target_leads: int | None = 100,
    cfg: Config | None = None,
) -> list[Lead]:
    """Fetch restaurant decision-maker leads for one Australian city via Apollo.io."""
    cfg = cfg or Config()
    url = "https://api.apollo.io/v1/mixed_people/api_search"
    headers = {
        "Cache-Control": "no-cache",
        "Content-Type": "application/json",
        "X-Api-Key": api_key,
    }

    city_leads: list[Lead] = []
    seen_local: set[str] = set()
    city_label = city.split(",")[0].strip()
    page = 1

    while page <= max_pages:
        if target_leads and len(city_leads) >= target_leads:
            log.info(f"  {city_label}: reached target of {target_leads} leads")
            break

        payload = {
            "organization_locations": [city],
            "q_organization_keyword_tags": RESTAURANT_KEYWORD_TAGS,
            "person_titles": RESTAURANT_DECISION_MAKER_TITLES,
            "organization_num_employees_ranges": ["1,10", "11,50", "51,200", "201,500"],
            "per_page": per_page,
            "page": page,
        }

        def _call_apollo():
            response = requests.post(url, headers=headers, json=payload, timeout=30)
            if response.status_code != 200:
                raise RuntimeError(f"Apollo API {response.status_code}: {response.text}")
            return response.json()

        try:
            data = with_retry(
                _call_apollo,
                attempts=cfg.RETRY_ATTEMPTS,
                backoff=cfg.RETRY_BACKOFF,
                label=f"Apollo:{city_label}:p{page}",
            )
        except Exception as exc:
            log.error(f"  Apollo fetch failed for {city_label} (page {page}): {exc}")
            break

        people = data.get("people") or []
        if not people:
            log.info(f"  {city_label}: no more results on page {page}")
            break

        added_this_page = 0
        for person in people:
            lead = _apollo_person_to_lead(person, city)
            if not lead:
                continue
            key = _lead_dedup_key(lead)
            if key in seen_local:
                continue
            seen_local.add(key)
            city_leads.append(lead)
            added_this_page += 1
            if target_leads and len(city_leads) >= target_leads:
                break

        pagination = data.get("pagination") or {}
        total_pages = pagination.get("total_pages")
        log.info(
            f"  {city_label}: page {page}"
            + (f"/{total_pages}" if total_pages else "")
            + f" — {len(people)} raw, +{added_this_page} new, {len(city_leads)} total"
        )

        # Apollo often omits pagination — keep going while pages are full
        if total_pages and page >= total_pages:
            break
        if len(people) < per_page:
            break

        page += 1
        time.sleep(cfg.SCRAPE_DELAY)

    if target_leads and len(city_leads) < target_leads:
        log.warning(
            f"  {city_label}: only found {len(city_leads)}/{target_leads} leads "
            f"(Apollo returned no more results)"
        )

    return city_leads


def fetch_restaurant_leads_from_cities(
    api_key: str,
    cities: list[str] | None = None,
    per_page: int = 100,
    max_pages: int = 20,
    target_per_city: int | None = 100,
    cfg: Config | None = None,
) -> list[Lead]:
    """
    Fetch restaurant leads across major Australian cities.
    Deduplicates by email (or company+contact+domain when email is missing).
    """
    cfg = cfg or Config()
    cities = cities or AUSTRALIAN_RESTAURANT_CITIES
    all_leads: list[Lead] = []
    seen: set[str] = set()

    log.info(f"Fetching restaurant leads from {len(cities)} cities: {', '.join(cities)}")

    for city in cities:
        city_label = city.split(",")[0].strip()
        log.info(f"── City: {city_label} ──")
        city_leads = _fetch_restaurant_leads_for_city(
            api_key, city,
            per_page=per_page,
            max_pages=max_pages,
            target_leads=target_per_city,
            cfg=cfg,
        )

        added = 0
        for lead in city_leads:
            key = _lead_dedup_key(lead)
            if key in seen:
                continue
            seen.add(key)
            all_leads.append(lead)
            added += 1

        log.info(f"  {city_label}: {added} new unique restaurant leads")
        time.sleep(cfg.SCRAPE_DELAY)

    log.info(f"Total unique restaurant leads fetched: {len(all_leads)}")
    return all_leads


def fetch_leads_from_apollo(api_key: str, cfg: Config | None = None) -> list[Lead]:
    """Backward-compatible wrapper — fetches from all default Australian cities."""
    return fetch_restaurant_leads_from_cities(
        api_key=api_key,
        per_page=25,
        max_pages=1,
        cfg=cfg,
    )


def run_lead_fetch_pipeline(
    cfg: Config | None = None,
    cities: list[str] | None = None,
    per_page: int = 100,
    max_pages: int = 20,
    target_per_city: int | None = 100,
    output_path: str | Path | None = None,
) -> tuple[list[Lead], Path]:
    """
    End-to-end lead fetch: Apollo (multi-city) → dedupe → JSON export.
    Returns (leads, json_path).
    """
    cfg = cfg or Config()
    if not cfg.APOLLO_API_KEY:
        raise ValueError("APOLLO_API_KEY is missing from .env")

    cities = cities or AUSTRALIAN_RESTAURANT_CITIES

    leads = fetch_restaurant_leads_from_cities(
        api_key=cfg.APOLLO_API_KEY,
        cities=cities,
        per_page=per_page,
        max_pages=max_pages,
        target_per_city=target_per_city,
        cfg=cfg,
    )

    if output_path is None:
        output_path = default_leads_output_path(cities, cfg)
    else:
        output_path = Path(output_path)

    json_path = save_leads_to_json(
        leads=leads,
        filepath=output_path,
        cities=cities,
        per_page=per_page,
        max_pages=max_pages,
        target_per_city=target_per_city,
    )

    return leads, json_path


# ═════════════════════════════════════════════════════════════
# 8. GOOGLE MAPS & SEARCH SCRAPER (Via SerpApi)
# ═════════════════════════════════════════════════════════════

def scrape_company_google_data(company_name: str, domain: str, cfg: Config) -> str:
    """
    Uses SerpApi to search Google Maps and Google Search for the restaurant's data.
    Returns a combined string of JSON and text for the LLM to parse.
    """
    if not cfg.SERPAPI_KEY:
        log.warning("    No SerpApi key found. Cannot scrape Google.")
        return ""

    collected_data = []

    # --- 1. Google Maps Search (For Address, Hours, Rating, Reviews, Images) ---
    log.info(f"    🔍 Searching Google Maps for: {company_name}")
    maps_url = "https://serpapi.com/search.json"
    maps_params = {
        "engine": "google_maps",
        "q": f"{company_name} restaurant Australia",
        "hl": "en",
        "api_key": cfg.SERPAPI_KEY
    }
    
    try:
        maps_resp = requests.get(maps_url, params=maps_params, timeout=cfg.SCRAPE_TIMEOUT)
        if maps_resp.status_code == 200:
            maps_data = maps_resp.json()
            
            # Extract the top matching place
            if "local_results" in maps_data and len(maps_data["local_results"]) > 0:
                top_place = maps_data["local_results"][0]
                
                extracted_maps_info = {
                    "Name": top_place.get("title"),
                    "Address": top_place.get("address"),
                    "Phone": top_place.get("phone"),
                    "Website": top_place.get("website"),
                    "Rating": top_place.get("rating"),
                    "Reviews_Count": top_place.get("reviews"),
                    "Price_Level": top_place.get("price"),
                    "Operating_Hours": top_place.get("operating_hours", {}),
                    "Service_Options": top_place.get("service_options", {}),
                    "Thumbnail_Image": top_place.get("thumbnail")
                }
                collected_data.append("GOOGLE MAPS PROFILE:\n" + json.dumps(extracted_maps_info, indent=2))
            else:
                collected_data.append("GOOGLE MAPS: No local listing found.")
    except Exception as e:
        log.warning(f"    Google Maps fetch failed: {e}")

    # --- 2. Google Organic Search (For Meta Titles, Snippets, and Links) ---
    log.info(f"    🔍 Searching Google Organic for: {domain}")
    search_params = {
        "engine": "google",
        "q": f"site:{domain} OR \"{company_name}\" menu reservation",
        "hl": "en",
        "gl": "au",
        "api_key": cfg.SERPAPI_KEY
    }
    
    try:
        search_resp = requests.get(maps_url, params=search_params, timeout=cfg.SCRAPE_TIMEOUT)
        if search_resp.status_code == 200:
            search_data = search_resp.json()
            
            organic_results = search_data.get("organic_results", [])[:5] # Grab top 5 search results
            search_snippets = []
            
            for res in organic_results:
                search_snippets.append({
                    "Title": res.get("title"),
                    "Link": res.get("link"),
                    "Snippet": res.get("snippet")
                })
            collected_data.append("\nGOOGLE ORGANIC SEARCH (Top 5 Results):\n" + json.dumps(search_snippets, indent=2))
    except Exception as e:
        log.warning(f"    Google Search fetch failed: {e}")

    # Combine all gathered Google data into a single text block for OpenAI
    return "\n\n".join(collected_data)


# ═════════════════════════════════════════════════════════════
# 9. LLM — CLAUDE 3.5 SONNET
# ═════════════════════════════════════════════════════════════

def generate_email_draft(lead: Lead, scraped_text: str, cfg: Config) -> dict:
    """Call OpenAI → parsed JSON dict."""
    if not OpenAI:
        raise ImportError("openai SDK not installed. Run: pip install openai")
    if not cfg.OPENAI_API_KEY:
        raise ValueError("OPENAI_API_KEY not set in environment variables.")

    client = OpenAI(api_key=cfg.OPENAI_API_KEY)

    user_content = textwrap.dedent(f"""
        PROSPECT DETAILS:
        - Company Name  : {lead.company_name}
        - Domain        : {lead.domain}
        - Contact Name  : {lead.contact_name}
        - First Name    : {lead.first_name}
        - Job Title     : {lead.job_title}
        - Contact Email : {lead.contact_email or "unknown"}

        SCRAPED WEBSITE CONTENT ({cfg.SCRAPE_MAX_CHARS} char cap):
        ---
        {scraped_text if scraped_text else "[No content retrieved — use the FALLBACK template.]"}
        ---

        Generate the personalised email draft JSON now.
    """).strip()

    def _call():
        return client.chat.completions.create(
            model      = cfg.OPENAI_MODEL,
            max_tokens = cfg.OPENAI_MAX_TOKENS,
            temperature= 0.7,
            messages   = [
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": user_content}
            ],
        )

    response = with_retry(_call, cfg.RETRY_ATTEMPTS, cfg.RETRY_BACKOFF, "OpenAI")
    raw      = response.choices[0].message.content.strip()

    # Strip accidental markdown fences
    raw = re.sub(r"^```[a-z]*\n?", "", raw, flags=re.MULTILINE)
    raw = re.sub(r"```$",          "", raw, flags=re.MULTILINE).strip()

    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        log.warning("  LLM returned non-JSON — extracting via regex.")
        subject = re.search(r'"subject"\s*:\s*"([^"]+)"', raw)
        body    = re.search(r'"body"\s*:\s*"([\s\S]+?)"(?:,|\})', raw)
        return {
            "business_model": "Unknown",
            "subject"       : subject.group(1) if subject else f"A quick note for {lead.company_name}",
            "body"          : body.group(1).replace("\\n", "\n") if body else raw,
            "confidence"    : "low",
            "notes"         : "JSON parse failed; regex extraction used.",
        }


# ═════════════════════════════════════════════════════════════
# 10. ZOHO MAIL — DRAFT CREATION
# ═════════════════════════════════════════════════════════════

def refresh_zoho_token(cfg: Config) -> str:
    """Exchange refresh token for a fresh Zoho access token."""
    url     = f"https://accounts.zoho.{cfg.ZOHO_REGION}/oauth/v2/token"
    payload = {
        "grant_type"   : "refresh_token",
        "client_id"    : cfg.ZOHO_CLIENT_ID,
        "client_secret": cfg.ZOHO_CLIENT_SECRET,
        "refresh_token": cfg.ZOHO_REFRESH_TOKEN,
    }
    resp = requests.post(url, data=payload, timeout=15)
    data = resp.json()
    if "access_token" not in data:
        raise RuntimeError(f"Zoho token refresh failed: {data}")
    log.info("  Zoho access token refreshed.")
    return data["access_token"]


def create_zoho_draft(
    lead: Lead,
    subject: str,
    body: str,
    cfg: Config,
    access_token: str,
) -> tuple[str, str]:
    """
    Create a DRAFT in Zoho Mail. Returns (draft_id, draft_url).
    NEVER sends — always uses action=saveDraft.
    """
    url     = f"{cfg.ZOHO_BASE_URL}/{cfg.ZOHO_ACCOUNT_ID}/messages"
    payload = {
        "fromAddress": cfg.ZOHO_FROM_EMAIL,
        "toAddress"  : lead.contact_email or "",
        "subject"    : subject,
        "content"    : body,
        "mailFormat" : "plaintext",
        "action"     : "saveDraft",   # ← SAFETY: never "sendMessage"
    }
    headers = {
        "Authorization": f"Zoho-oauthtoken {access_token}",
        "Content-Type" : "application/json",
    }

    def _call():
        r = requests.post(url, json=payload, headers=headers, timeout=20)
        d = r.json()
        if r.status_code in (200, 201) and d.get("status", {}).get("code") in (200, 201):
            return d
        raise RuntimeError(f"Zoho [{r.status_code}]: {d}")

    data    = with_retry(_call, 3, 2.0, "ZohoDraft")
    msg_id  = data.get("data", {}).get("messageId", "")
    draft_url = (
        f"https://mail.zoho.{cfg.ZOHO_REGION}/zm/#mail/folder/draft/message/{msg_id}"
    )
    log.info(f"  ✅ Zoho draft created: {msg_id}")
    return msg_id, draft_url


# ═════════════════════════════════════════════════════════════
# 11. SLACK NOTIFICATION
# ═════════════════════════════════════════════════════════════

def notify_slack(result: OutreachResult, cfg: Config) -> None:
    """Post a structured review card to Slack."""
    if not SlackClient or not cfg.SLACK_BOT_TOKEN:
        log.info("  Slack not configured — skipping.")
        return

    client = SlackClient(token=cfg.SLACK_BOT_TOKEN)
    emoji  = {"high": "🟢", "medium": "🟡", "low": "🔴"}.get(
        result.confidence.lower(), "⚪"
    )
    status_emoji = "✅" if not result.error else "❌"

    blocks = [
        {
            "type": "header",
            "text": {"type": "plain_text", "text": f"📬 Draft Ready: {result.lead.company_name}"}
        },
        {
            "type": "section",
            "fields": [
                {"type": "mrkdwn", "text": f"*Contact:*\n{result.lead.contact_name} — {result.lead.job_title}"},
                {"type": "mrkdwn", "text": f"*Business Model:*\n{result.business_model or 'N/A'}"},
                {"type": "mrkdwn", "text": f"*AI Confidence:*\n{emoji} {result.confidence.capitalize()}"},
                {"type": "mrkdwn", "text": f"*Status:*\n{status_emoji} {result.status}"},
            ]
        },
        {
            "type": "section",
            "text": {"type": "mrkdwn", "text": f"*Subject:* _{result.subject}_"}
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": f"*Body preview:*\n```{result.body[:500]}{'...' if len(result.body) > 500 else ''}```"
            }
        },
    ]

    if result.zoho_draft_url:
        blocks.append({
            "type": "actions",
            "elements": [{
                "type": "button",
                "text": {"type": "plain_text", "text": "Open in Zoho Mail"},
                "url" : result.zoho_draft_url,
                "style": "primary",
            }]
        })

    if result.error:
        blocks.append({
            "type": "section",
            "text": {"type": "mrkdwn", "text": f"⚠️ *Error:* {result.error}"}
        })

    try:
        client.chat_postMessage(
            channel=cfg.SLACK_REVIEW_CHANNEL,
            text   =f"Draft ready for {result.lead.company_name}",
            blocks =blocks,
        )
        log.info(f"  Slack notified: {cfg.SLACK_REVIEW_CHANNEL}")
    except SlackApiError as e:
        log.warning(f"  Slack failed: {e.response.get('error', e)}")


# ═════════════════════════════════════════════════════════════
# 12. LOCAL BACKUP & REPORTS
# ═════════════════════════════════════════════════════════════

def save_draft_locally(result: OutreachResult, cfg: Config) -> None:
    """Save text backup AND a JSON file of extracted restaurant data."""
    out_dir = Path(cfg.OUTPUT_DIR)
    out_dir.mkdir(parents=True, exist_ok=True)

    safe_name = re.sub(r"[^\w\-]", "_", result.lead.company_name)
    timestamp = datetime.utcnow().strftime('%Y%m%d_%H%M%S')
    
    # 1. Save the Outreach Email Draft
    txt_filename = out_dir / f"{safe_name}_{timestamp}.txt"
    content = textwrap.dedent(f"""\
        ══════════════════════════════════════════════════════
        TUVI SOLUTIONS — OUTREACH DRAFT  ⚠️  REVIEW BEFORE SENDING
        ══════════════════════════════════════════════════════
        Generated  : {result.timestamp}
        Company    : {result.lead.company_name}
        Contact    : {result.lead.contact_name}
        Zoho Draft : {result.zoho_draft_url or "N/A"}
        ──────────────────────────────────────────────────────
        SUBJECT: {result.subject}
        ──────────────────────────────────────────────────────
        {result.body}
        ══════════════════════════════════════════════════════
    """)
    txt_filename.write_text(content, encoding="utf-8")
    
    # 2. Save the Detailed Restaurant Extraction Profile
    if result.restaurant_data:
        json_filename = out_dir / f"{safe_name}_{timestamp}_DataProfile.json"
        json_filename.write_text(json.dumps(result.restaurant_data, indent=4), encoding="utf-8")
        log.info(f"  ✅ Data Profile saved: {json_filename}")


def save_run_report(results: list[OutreachResult], cfg: Config) -> None:
    """
    Write two files after a full run:
      reports/run_YYYYMMDD_HHMMSS.json  — machine-readable with full restaurant extraction
      reports/run_YYYYMMDD_HHMMSS.csv   — spreadsheet-friendly
    """
    rep_dir   = Path(cfg.REPORT_DIR)
    rep_dir.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S")

    # 1. Save Complete JSON (Matches what you are viewing in image_d157a8.jpg)
    json_path = rep_dir / f"run_{timestamp}.json"
    json_path.write_text(
        json.dumps([r.to_dict() for r in results], indent=2, ensure_ascii=False),
        encoding="utf-8",
    )

    # 2. Save Spreadsheet-Friendly CSV
    csv_path = rep_dir / f"run_{timestamp}.csv"
    if results:
        fieldnames = list(results[0].to_dict().keys())
        with open(csv_path, "w", newline="", encoding="utf-8") as f:
            writer = csv.DictWriter(f, fieldnames=fieldnames)
            writer.writeheader()
            
            for r in results:
                row_data = r.to_dict()
                # If restaurant data exists, flatten it into text string so it doesn't break CSV cells
                if isinstance(row_data.get("restaurant_data"), dict):
                    row_data["restaurant_data"] = json.dumps(row_data["restaurant_data"])
                writer.writerow(row_data)

    log.info(f"  Run report saved: {json_path} | {csv_path}")


# ═════════════════════════════════════════════════════════════
# 13. PER-LEAD PIPELINE
# ═════════════════════════════════════════════════════════════

def process_lead(
    lead: Lead,
    cfg: Config,
    zoho_token: str,
    notion_client=None,
) -> OutreachResult:
    """
    Full pipeline for one lead:
      scrape → LLM → Zoho draft → local save → Slack → sheet write-back
    Errors are caught and stored in result.error — never propagated.
    """
    result = OutreachResult(lead)
    thread_name = f"[{lead.company_name}]"

    try:
        # ── 1. Scrape ──────────────────────────────────────
        log.info(f"{thread_name} [1/5] Scraping Google Data for {lead.company_name}...")
        result.scraped_text = scrape_company_google_data(lead.company_name, lead.domain, cfg)
        if not result.scraped_text:
            log.warning(f"{thread_name}  No content scraped — using generic template.")

        # ── 2. Generate email ──────────────────────────────
        log.info(f"{thread_name} [2/5] Calling OpenAI…")
        try:
            llm = generate_email_draft(lead, result.scraped_text, cfg)
        except Exception as e:
            raise RuntimeError(f"LLM error: {e}")

        result.business_model = llm.get("business_model", "Unknown")
        result.subject        = llm.get("subject", "")
        result.body           = llm.get("body", "").replace("\\n", "\n")
        result.confidence     = llm.get("confidence", "low")
        result.restaurant_data = llm.get("restaurant_data", {})
        result.notes          = llm.get("notes", "")

        # ── 3. Zoho draft ──────────────────────────────────
        if zoho_token:
            log.info(f"{thread_name} [3/5] Creating Zoho draft…")
            try:
                result.zoho_draft_id, result.zoho_draft_url = create_zoho_draft(
                    lead, result.subject, result.body, cfg, zoho_token
                )
            except Exception as e:
                log.warning(f"{thread_name}  Zoho draft failed: {e} — saving locally only.")
                result.notes += f" | Zoho error: {e}"
        else:
            log.info(f"{thread_name} [3/5] Skipping Zoho (no token).")

        # ── 4. Local save ──────────────────────────────────
        log.info(f"{thread_name} [4/5] Saving local backup…")
        save_draft_locally(result, cfg)

        # ── 5. Slack ───────────────────────────────────────
        log.info(f"{thread_name} [5/5] Sending Slack notification…")
        notify_slack(result, cfg)

    except Exception as e:
        result.error = str(e)
        log.error(f"{thread_name}  ❌ {e}\n{traceback.format_exc()}")
        save_draft_locally(result, cfg)   # always save even on error

    finally:
        # ── Notion Logging (always, even on error) ──
        if notion_client is not None:
            write_to_notion(result, cfg, notion_client)

    return result


# ═════════════════════════════════════════════════════════════
# 14. CONCURRENT ORCHESTRATOR
# ═════════════════════════════════════════════════════════════

def run_outreach(
    leads: list[Lead],
    cfg: Config | None = None,
    notion_client=None,
) -> list[OutreachResult]:
    """
    Process all leads concurrently using a thread pool.
    """
    cfg = cfg or Config()

    if not leads:
        log.warning("No leads to process.")
        return []

    # ── Obtain a Zoho token once for the whole batch ───────
    zoho_token = ""
    
    # 1. Prioritize the Refresh Token Flow (Best Practice)
    if cfg.ZOHO_REFRESH_TOKEN and cfg.ZOHO_CLIENT_ID and cfg.ZOHO_CLIENT_SECRET:
        log.info("🔄 Initiating Zoho OAuth2 handshake via Refresh Token...")
        try:
            zoho_token = refresh_zoho_token(cfg)
            log.info("✅ Fresh Zoho Access Token successfully generated!")
        except Exception as e:
            log.warning(f"⚠️ Zoho token refresh failed: {e} — drafts will be saved locally only.")
            
    # 2. Fallback to a hardcoded Access Token if no Refresh Token exists
    elif cfg.ZOHO_ACCESS_TOKEN:
        log.info("ℹ️ Using static Zoho Access Token (Note: This expires after 1 hour).")
        zoho_token = cfg.ZOHO_ACCESS_TOKEN
        
    # 3. No credentials provided
    else:
        log.warning("⚠️ No Zoho credentials configured. Drafts will be saved locally only.")

    log.info(
        f"\n{'═'*60}\n"
        f"  TUVI OUTREACH  —  {len(leads)} lead(s)  |  "
        f"{cfg.MAX_WORKERS} worker(s)\n"
        f"{'═'*60}"
    )

    results_map: dict[int, OutreachResult] = {}

    with ThreadPoolExecutor(max_workers=cfg.MAX_WORKERS) as pool:
        future_to_idx = {
            pool.submit(process_lead, lead, cfg, zoho_token, notion_client): idx
            for idx, lead in enumerate(leads)
        }

        for future in as_completed(future_to_idx):
            idx = future_to_idx[future]
            lead = leads[idx]
            try:
                result = future.result()
                results_map[idx] = result
                status = "✅" if not result.error else "❌"
                log.info(
                    f"{status} Done [{idx+1}/{len(leads)}]: "
                    f"{lead.company_name} — {result.confidence} confidence"
                )
            except Exception as e:
                r       = OutreachResult(lead)
                r.error = f"Unexpected: {e}"
                results_map[idx] = r
                log.error(f"❌ Unexpected error for {lead.company_name}: {e}")

    # Reassemble in original order
    results = [results_map[i] for i in range(len(leads))]

    # ── Post-run report ────────────────────────────────────
    save_run_report(results, cfg)

    ok  = sum(1 for r in results if not r.error)
    err = len(results) - ok
    log.info(
        f"\n{'═'*60}\n"
        f"  COMPLETE — {ok} succeeded / {err} failed / {len(results)} total\n"
        f"{'═'*60}\n"
    )

    return results


# ═════════════════════════════════════════════════════════════
# 15. CLI & MAIN ORCHESTRATION
# ═════════════════════════════════════════════════════════════


def _build_arg_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        description="Tuvi Solutions — Restaurant Lead Fetch & Outreach Agent",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=textwrap.dedent("""\
            Examples:
              # Fetch Sydney restaurant leads only
              python tuvi_outreach_agent.py --fetch-leads --city Sydney

              # Fetch Perth leads
              python tuvi_outreach_agent.py --fetch-leads --city Perth

              # Fetch + run outreach (dry-run)
              python tuvi_outreach_agent.py --fetch-leads --city Sydney --run-outreach --no-zoho --no-slack
        """),
    )

    # Pipeline mode
    mode = p.add_mutually_exclusive_group()
    mode.add_argument(
        "--fetch-leads", action="store_true",
        help="Fetch restaurant leads from Apollo and save to leads/lead.json",
    )
    mode.add_argument(
        "--json", nargs="?", const="leads/lead.json", metavar="FILE",
        help="Load leads from JSON (default file: leads/lead.json)",
    )
    mode.add_argument(
        "--csv", metavar="FILE",
        help="Load leads from a CSV file instead of Apollo",
    )

    p.add_argument(
        "--run-outreach", action="store_true",
        help="After --fetch-leads, run the full outreach pipeline on fetched leads",
    )
    city_group = p.add_mutually_exclusive_group()
    city_group.add_argument(
        "--city", metavar="CITY",
        help="Fetch leads for one city only (e.g. --city Sydney, --city Perth)",
    )
    city_group.add_argument(
        "--cities", nargs="+", metavar="CITY",
        help="Fetch leads for multiple cities (e.g. --cities Sydney Melbourne)",
    )
    p.add_argument("--per-page", type=int, default=100,
                   help="Apollo results per page (default: 100)")
    p.add_argument("--max-pages", type=int, default=20,
                   help="Max pages per city (default: 20)")
    p.add_argument("--total", type=int, default=None, metavar="N",
                   help="Target leads per city (default: 100 for single --city)")
    p.add_argument("--output", metavar="FILE",
                   help="JSON output path for --fetch-leads (default: leads/lead.json)")

    # Concurrency & Flags
    p.add_argument("--workers", type=int, default=Config.MAX_WORKERS,
                   help=f"Parallel workers (default: {Config.MAX_WORKERS})")
    p.add_argument("--no-zoho",   action="store_true", help="Skip Zoho upload")
    p.add_argument("--no-slack",  action="store_true", help="Skip Slack notifications")
    p.add_argument("--no-notion", action="store_true", help="Skip Notion database logging")
    p.add_argument("--verbose",   "-v", action="store_true", help="Debug logging")

    return p


def _resolve_cities(args) -> list[str]:
    try:
        return resolve_cities(city=getattr(args, "city", None), cities=getattr(args, "cities", None))
    except ValueError as exc:
        log.error(f"❌ {exc}")
        raise SystemExit(2) from exc


def main():
    parser = _build_arg_parser()
    args   = parser.parse_args()

    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)

    cfg = Config()
    cfg.MAX_WORKERS = args.workers

    if args.no_zoho:
        cfg.ZOHO_ACCESS_TOKEN  = ""
        cfg.ZOHO_REFRESH_TOKEN = ""
        log.info("--no-zoho: Zoho upload disabled.")

    if args.no_slack:
        cfg.SLACK_BOT_TOKEN = ""
        log.info("--no-slack: Slack notifications disabled.")

    cities = _resolve_cities(args)
    target_per_city = getattr(args, "total", None)
    if target_per_city is None and getattr(args, "city", None):
        target_per_city = 100
    if getattr(args, "city", None) or getattr(args, "cities", None):
        log.info(f"City filter: {', '.join(c.split(',')[0] for c in cities)}")
    raw_leads: list[Lead] = []

    # ── Mode 1: Fetch leads from Apollo (multi-city) ───────
    if args.fetch_leads:
        log.info("🚀 Starting Restaurant Lead Fetch Pipeline...")
        if not cfg.APOLLO_API_KEY:
            log.error("❌ APOLLO_API_KEY is missing from .env")
            return

        try:
            raw_leads, json_path = run_lead_fetch_pipeline(
                cfg=cfg,
                cities=cities,
                per_page=args.per_page,
                max_pages=args.max_pages,
                target_per_city=target_per_city,
                output_path=args.output,
            )
        except ValueError as e:
            log.error(f"❌ {e}")
            return

        log.info(f"✅ Fetch complete — {len(raw_leads)} leads saved to {json_path}")

        if not args.run_outreach:
            log.info("Tip: run outreach with  --fetch-leads --run-outreach  or  --json leads/lead.json")
            return

    # ── Mode 2: Load leads from JSON ─────────────────────
    elif args.json is not None:
        log.info(f"📂 Loading leads from JSON: {args.json}")
        raw_leads = load_leads_from_json(args.json)

    # ── Mode 3: Load leads from CSV ────────────────────────
    elif args.csv:
        log.info(f"📂 Loading leads from CSV: {args.csv}")
        raw_leads = load_leads_from_csv(args.csv)

    # ── Mode 3: Default — fetch from Apollo then outreach ──
    else:
        log.info("🚀 Starting End-to-End Apollo + Outreach Workflow...")
        if not cfg.APOLLO_API_KEY:
            log.error("❌ APOLLO_API_KEY is missing from .env")
            return
        raw_leads = fetch_restaurant_leads_from_cities(
            api_key=cfg.APOLLO_API_KEY,
            cities=cities,
            per_page=args.per_page,
            max_pages=args.max_pages,
            target_per_city=target_per_city,
            cfg=cfg,
        )

    if not raw_leads:
        log.warning("⚠️ No leads to process. Exiting.")
        return

    notion_client = None
    if not args.no_notion and cfg.NOTION_TOKEN:
        if NotionClient is None:
            log.warning("⚠️ notion-client not installed. Skipping Notion logging.")
        else:
            try:
                notion_client = NotionClient(auth=cfg.NOTION_TOKEN)
                log.info("✅ Notion client authenticated.")
            except Exception as e:
                log.error(f"❌ Failed to initialize Notion client: {e}")
    elif not args.no_notion:
        log.warning("⚠️ Notion token missing. Skipping database logging.")

    results = run_outreach(raw_leads, cfg, notion_client=notion_client)
    ok = sum(1 for r in results if not r.error)
    log.info(f"Outreach finished: {ok}/{len(results)} succeeded")


if __name__ == "__main__":
    main()
