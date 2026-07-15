"""Targeted Apollo enrichment after Google Places discovery and details."""

from __future__ import annotations

import logging
import re
import time
from copy import deepcopy
from typing import Any
from urllib.parse import urlsplit

import requests

from niche_config import NicheConfig
from request_budget import BudgetExhausted, RequestBudget
from tuvi_outreach_agent import Config

log = logging.getLogger("apollo_enrichment")

DEFAULT_APOLLO_API_BASE = "https://api.apollo.io/api/v1"
EMAIL_RE = re.compile(r"^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$")
NON_ORGANIZATION_DOMAINS = {
    "facebook.com",
    "gmail.com",
    "hotmail.com",
    "icloud.com",
    "instagram.com",
    "linktr.ee",
    "live.com",
    "outlook.com",
    "proton.me",
    "protonmail.com",
    "yahoo.com",
}


class ApolloAPIError(RuntimeError):
    """Safe Apollo error that never includes response bodies or credentials."""

    def __init__(self, message: str, *, status_code: int | None = None) -> None:
        super().__init__(message)
        self.status_code = status_code


def get_apollo_api_key(cfg: Config | None = None) -> str:
    cfg = cfg or Config()
    key = str(getattr(cfg, "APOLLO_API_KEY", "") or "").strip()
    if not key:
        raise ValueError(
            "APOLLO_API_KEY is required when Apollo contact enrichment is enabled"
        )
    return key


def _apollo_base_url(cfg: Config) -> str:
    return str(
        getattr(cfg, "APOLLO_API_BASE_URL", DEFAULT_APOLLO_API_BASE)
        or DEFAULT_APOLLO_API_BASE
    ).rstrip("/")


def _apollo_request(
    method: str,
    path: str,
    cfg: Config,
    *,
    label: str,
    budget: RequestBudget | None = None,
    body: dict | None = None,
    params: dict | None = None,
) -> dict:
    url = f"{_apollo_base_url(cfg)}/{path.lstrip('/')}"
    headers = {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "Cache-Control": "no-cache",
        "X-Api-Key": get_apollo_api_key(cfg),
    }
    attempts = max(1, int(cfg.RETRY_ATTEMPTS))

    for attempt in range(1, attempts + 1):
        if budget is not None:
            budget.consume(1)
        try:
            response = requests.request(
                method,
                url,
                headers=headers,
                json=body,
                params=params,
                timeout=max(cfg.SCRAPE_TIMEOUT, 20),
            )
        except requests.RequestException as exc:
            if attempt >= attempts:
                raise ApolloAPIError("Apollo API network request failed") from exc
            wait = cfg.RETRY_BACKOFF ** (attempt - 1)
            log.warning(
                "[%s] Attempt %d/%d failed — retrying in %.1fs",
                label,
                attempt,
                attempts,
                wait,
            )
            time.sleep(wait)
            continue

        if 200 <= response.status_code < 300:
            if not response.content:
                return {}
            try:
                return response.json()
            except ValueError as exc:
                raise ApolloAPIError("Apollo API returned invalid JSON") from exc

        if response.status_code != 429 and response.status_code < 500:
            raise ApolloAPIError(
                f"Apollo API HTTP {response.status_code}",
                status_code=response.status_code,
            )
        if attempt >= attempts:
            raise ApolloAPIError(
                f"Apollo API HTTP {response.status_code}",
                status_code=response.status_code,
            )

        retry_after = response.headers.get("Retry-After") or ""
        try:
            wait = min(30.0, max(0.0, float(retry_after)))
        except ValueError:
            wait = cfg.RETRY_BACKOFF ** (attempt - 1)
        if wait == 0:
            wait = cfg.RETRY_BACKOFF ** (attempt - 1)
        log.warning(
            "[%s] Apollo HTTP %d — retrying in %.1fs",
            label,
            response.status_code,
            wait,
        )
        time.sleep(wait)

    return {}


def _domain_from_value(raw_value: str) -> str:
    raw = str(raw_value or "").strip().lower()
    if not raw:
        return ""
    if "@" in raw and "://" not in raw:
        raw = raw.rsplit("@", 1)[-1]
    if "://" not in raw:
        raw = f"https://{raw}"
    try:
        host = (urlsplit(raw).hostname or "").lower()
    except ValueError:
        return ""
    return host.removeprefix("www.").strip(".")


def _record_domain(record: dict) -> str:
    contact = record.get("contact") or {}
    website_domain = _domain_from_value(contact.get("website") or "")
    if website_domain and not _is_non_organization_domain(website_domain):
        return website_domain
    email_domain = _domain_from_value(contact.get("email") or "")
    if _is_non_organization_domain(email_domain):
        return ""
    return email_domain


def _is_non_organization_domain(domain: str) -> bool:
    domain = str(domain or "").lower().strip(".")
    return any(
        domain == blocked or domain.endswith(f".{blocked}")
        for blocked in NON_ORGANIZATION_DOMAINS
    )


def needs_apollo_enrichment(record: dict) -> bool:
    contact = record.get("contact") or {}
    owners = [
        str(owner).strip()
        for owner in (record.get("owners") or [])
        if str(owner).strip()
    ]
    return not owners or not str(contact.get("email") or "").strip()


def _candidate_score(person: dict, titles: list[str], email_missing: bool) -> tuple[int, int]:
    title = str(person.get("title") or "").strip().lower()
    title_rank = len(titles) + 1
    for index, preferred in enumerate(titles):
        preferred = preferred.lower()
        if title == preferred or re.search(rf"\b{re.escape(preferred)}\b", title):
            title_rank = index
            break
    has_email_penalty = 0 if not email_missing or person.get("has_email") else 1
    if email_missing:
        return has_email_penalty, title_rank
    return title_rank, 0


def _choose_candidate(people: list[dict], niche: NicheConfig, email_missing: bool) -> dict:
    candidates = [person for person in people if str(person.get("id") or "").strip()]
    if not candidates:
        return {}
    return min(
        candidates,
        key=lambda person: _candidate_score(person, niche.person_titles, email_missing),
    )


def _full_person_name(person: dict) -> str:
    name = str(person.get("name") or "").strip()
    if not name:
        name = " ".join(
            part
            for part in (
                str(person.get("first_name") or "").strip(),
                str(person.get("last_name") or "").strip(),
            )
            if part
        ).strip()
    if "*" in name:
        return ""
    return name


def enrich_missing_contact_with_apollo(
    record: dict,
    cfg: Config,
    niche: NicheConfig,
    *,
    budget: RequestBudget | None = None,
) -> tuple[dict, dict[str, Any]]:
    """Fill only missing owner/work-email fields using targeted Apollo calls."""
    enriched = deepcopy(record)
    stats: dict[str, Any] = {
        "status": "not_needed",
        "search_requests": 0,
        "match_requests": 0,
        "owner_added": False,
        "email_added": False,
    }
    if not needs_apollo_enrichment(enriched):
        return enriched, stats

    domain = _record_domain(enriched)
    if not domain:
        stats["status"] = "skipped_no_domain"
        return enriched, stats

    contact = enriched.setdefault("contact", {})
    owners = enriched.setdefault("owners", [])
    email_missing = not str(contact.get("email") or "").strip()
    search_body: dict[str, Any] = {
        "q_organization_domains_list": [domain],
        "person_titles": list(niche.person_titles),
        "per_page": 10,
        "page": 1,
    }
    if email_missing:
        search_body["contact_email_status"] = ["verified"]

    try:
        stats["search_requests"] = 1
        search_data = _apollo_request(
            "POST",
            "mixed_people/api_search",
            cfg,
            label=f"apollo-search:{domain}",
            budget=budget,
            body=search_body,
        )
        candidate = _choose_candidate(search_data.get("people") or [], niche, email_missing)
        if not candidate:
            stats["status"] = "no_candidate"
            return enriched, stats

        person_id = str(candidate.get("id") or "").strip()
        stats["match_requests"] = 1
        match_data = _apollo_request(
            "POST",
            "people/match",
            cfg,
            label=f"apollo-match:{person_id[:10]}",
            budget=budget,
            params={
                "reveal_personal_emails": "false",
                "reveal_phone_number": "false",
            },
            body={
                "id": person_id,
                "domain": domain,
                "organization_name": str(enriched.get("name") or "").strip(),
            },
        )
    except BudgetExhausted:
        stats["status"] = "budget_exhausted"
        return enriched, stats
    except ApolloAPIError as exc:
        stats["status"] = "error"
        stats["error"] = str(exc)
        stats["error_code"] = exc.status_code
        return enriched, stats

    person = match_data.get("person") or {}
    if not person:
        stats["status"] = "no_match"
        return enriched, stats

    owner_name = _full_person_name(person)
    owner_names = {str(owner).strip().lower() for owner in owners}
    if owner_name and owner_name.lower() not in owner_names:
        owners.append(owner_name)
        stats["owner_added"] = True

    work_email = str(person.get("email") or "").strip().lower()
    if email_missing and EMAIL_RE.fullmatch(work_email):
        contact["email"] = work_email
        stats["email_added"] = True

    enriched["apollo_lead"] = {
        "company": {
            "name": str(
                (person.get("organization") or {}).get("name")
                or enriched.get("name")
                or ""
            ).strip(),
            "domain": domain,
        },
        "contact": {
            "name": owner_name,
            "job_title": str(person.get("title") or candidate.get("title") or "").strip(),
            "email": work_email if stats["email_added"] else "",
            "has_email": bool(person.get("email")),
        },
        "source": {
            "provider": "apollo.io",
            "person_id": person_id,
            "mode": "places_contact_enrichment",
        },
    }
    stats["status"] = (
        "enriched"
        if stats["owner_added"] or stats["email_added"]
        else "matched_no_new_data"
    )
    return enriched, stats
