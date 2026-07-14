"""Business niche settings for Places discovery and Apollo contact enrichment."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Callable, TYPE_CHECKING

if TYPE_CHECKING:
    from tuvi_outreach_agent import Lead

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

DENTIST_KEYWORD_TAGS: list[str] = [
    "dental",
    "dentistry",
    "dental clinic",
    "dental practice",
    "orthodontics",
]

DENTIST_TITLES: list[str] = [
    "owner",
    "founder",
    "practice owner",
    "principal dentist",
    "director",
    "managing director",
    "office manager",
]

PLUMBER_KEYWORD_TAGS: list[str] = [
    "plumbing",
    "plumber",
    "hvac",
    "trade services",
    "home services",
]

PLUMBER_TITLES: list[str] = [
    "owner",
    "founder",
    "director",
    "managing director",
    "operations manager",
    "general manager",
]


def _accept_restaurant_lead(person: dict, city: str) -> Lead | None:
    from tuvi_outreach_agent import Lead

    org = person.get("organization") or {}
    company_name = (org.get("name") or "").strip()
    if not company_name:
        return None

    org_keywords = " ".join([
        company_name,
        org.get("industry") or "",
        org.get("short_description") or "",
    ]).lower()

    restaurant_signals = (
        "restaurant", "dining", "cafe", "café", "bistro", "eatery", "food",
        "grill", "kitchen", "pizza", "sushi", "bar",
    )
    non_restaurant = ("software", "real estate", "law firm", "accounting", "insurance")

    if any(ex in org_keywords for ex in non_restaurant):
        return None

    if not any(signal in org_keywords for signal in restaurant_signals):
        if "hospitality" not in org_keywords and "catering" not in org_keywords:
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
            "lead_type": "restaurant",
        },
    )


def _accept_generic_lead(lead_type: str, signals: tuple[str, ...], person: dict, city: str) -> Lead | None:
    from tuvi_outreach_agent import Lead

    org = person.get("organization") or {}
    company_name = (org.get("name") or "").strip()
    if not company_name:
        return None

    org_keywords = " ".join([
        company_name,
        org.get("industry") or "",
        org.get("short_description") or "",
    ]).lower()

    if signals and not any(signal in org_keywords for signal in signals):
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
            "lead_type": lead_type,
        },
    )


def _accept_dentist_lead(person: dict, city: str) -> Lead | None:
    return _accept_generic_lead(
        "dentist",
        ("dental", "dentist", "dentistry", "orthodont", "oral"),
        person,
        city,
    )


def _accept_plumber_lead(person: dict, city: str) -> Lead | None:
    return _accept_generic_lead(
        "plumber",
        ("plumb", "hvac", "drain", "pipe", "gas fit"),
        person,
        city,
    )


@dataclass(frozen=True)
class NicheConfig:
    slug: str
    keyword_tags: list[str]
    person_titles: list[str]
    places_query_suffix: str
    lead_type: str
    accept_lead: Callable[[dict, str], Lead | None]

    def apollo_person_to_lead(self, person: dict, city: str) -> Lead | None:
        return self.accept_lead(person, city)


_NICHES: dict[str, NicheConfig] = {
    "restaurant": NicheConfig(
        slug="restaurant",
        keyword_tags=list(RESTAURANT_KEYWORD_TAGS),
        person_titles=list(RESTAURANT_DECISION_MAKER_TITLES),
        places_query_suffix="restaurant",
        lead_type="restaurant",
        accept_lead=_accept_restaurant_lead,
    ),
    "dentist": NicheConfig(
        slug="dentist",
        keyword_tags=list(DENTIST_KEYWORD_TAGS),
        person_titles=list(DENTIST_TITLES),
        places_query_suffix="dentist",
        lead_type="dentist",
        accept_lead=_accept_dentist_lead,
    ),
    "plumber": NicheConfig(
        slug="plumber",
        keyword_tags=list(PLUMBER_KEYWORD_TAGS),
        person_titles=list(PLUMBER_TITLES),
        places_query_suffix="plumber",
        lead_type="plumber",
        accept_lead=_accept_plumber_lead,
    ),
}


def get_niche(niche_type: str) -> NicheConfig:
    slug = (niche_type or "restaurant").strip().lower()
    if slug not in _NICHES:
        allowed = ", ".join(sorted(_NICHES))
        raise ValueError(f"Unknown niche type {niche_type!r}. Allowed: {allowed}")
    return _NICHES[slug]


def list_niche_types() -> list[str]:
    return sorted(_NICHES)
