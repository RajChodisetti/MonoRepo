"""Apollo lead fetch with niche, budget, dedup, and pagination."""

from __future__ import annotations

import logging
import time
from typing import TYPE_CHECKING

import requests

from niche_config import NicheConfig, get_niche
from request_budget import BudgetExhausted, RequestBudget
from tuvi_outreach_agent import Config, Lead, _lead_dedup_key, with_retry

if TYPE_CHECKING:
    from known_leads import KnownLeadsRegistry

log = logging.getLogger("lead_fetch")


def fetch_leads_for_city(
    api_key: str,
    city: str,
    niche: NicheConfig | str = "restaurant",
    *,
    per_page: int = 100,
    max_pages: int = 20,
    target_leads: int | None = 100,
    start_page: int = 1,
    budget: RequestBudget | None = None,
    known: KnownLeadsRegistry | None = None,
    cfg: Config | None = None,
) -> tuple[list[Lead], int, dict]:
    """
    Fetch leads for one city. Returns (new_leads, next_apollo_page, stats).
    """
    cfg = cfg or Config()
    if isinstance(niche, str):
        niche = get_niche(niche)

    url = "https://api.apollo.io/v1/mixed_people/api_search"
    headers = {
        "Cache-Control": "no-cache",
        "Content-Type": "application/json",
        "X-Api-Key": api_key,
    }

    city_leads: list[Lead] = []
    seen_local: set[str] = set()
    city_label = city.split(",")[0].strip()
    page = max(1, int(start_page))

    stats = {
        "apollo_requests": 0,
        "pages_fetched": 0,
        "raw_people": 0,
        "skipped_known": 0,
        "added": 0,
    }

    pages_without_new = 0

    while page <= max_pages:
        if target_leads and len(city_leads) >= target_leads:
            break
        if budget is not None and not budget.can_consume(1):
            log.info(f"  {city_label}: Apollo budget exhausted")
            break

        payload = {
            "organization_locations": [city],
            "q_organization_keyword_tags": niche.keyword_tags,
            "person_titles": niche.person_titles,
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
            if budget is not None:
                budget.consume(1)
            stats["apollo_requests"] += 1
            data = with_retry(
                _call_apollo,
                attempts=cfg.RETRY_ATTEMPTS,
                backoff=cfg.RETRY_BACKOFF,
                label=f"Apollo:{city_label}:p{page}",
            )
        except BudgetExhausted:
            break
        except Exception as exc:
            log.error(f"  Apollo fetch failed for {city_label} (page {page}): {exc}")
            break

        stats["pages_fetched"] += 1
        people = data.get("people") or []
        stats["raw_people"] += len(people)
        if not people:
            log.info(f"  {city_label}: no more results on page {page}")
            page += 1
            break

        added_this_page = 0
        for person in people:
            lead = niche.apollo_person_to_lead(person, city)
            if not lead:
                continue
            key = _lead_dedup_key(lead)
            apollo_id = (lead.extra.get("apollo_id") or "").strip()

            if key in seen_local:
                continue
            if known is not None:
                if known.is_dedup_known({"contact": {"email": lead.contact_email}, "company": {"name": lead.company_name, "domain": lead.domain}, "source": {"person_id": apollo_id}}):
                    stats["skipped_known"] += 1
                    continue
                if apollo_id and known.is_apollo_known(apollo_id):
                    stats["skipped_known"] += 1
                    continue

            seen_local.add(key)
            lead.extra["lead_type"] = niche.lead_type
            city_leads.append(lead)
            added_this_page += 1
            stats["added"] += 1
            if target_leads and len(city_leads) >= target_leads:
                break

        pagination = data.get("pagination") or {}
        total_pages = pagination.get("total_pages")
        log.info(
            f"  {city_label}: page {page}"
            + (f"/{total_pages}" if total_pages else "")
            + f" — {len(people)} raw, +{added_this_page} new, {len(city_leads)} total"
        )

        page += 1

        if added_this_page == 0:
            pages_without_new += 1
        else:
            pages_without_new = 0

        if total_pages and page > total_pages:
            break
        if len(people) < per_page:
            break
        if pages_without_new >= 3:
            log.info(f"  {city_label}: stopping after {pages_without_new} pages with no new leads")
            break

        time.sleep(cfg.SCRAPE_DELAY)

    if target_leads and len(city_leads) < target_leads:
        log.warning(
            f"  {city_label}: only found {len(city_leads)}/{target_leads} leads "
            f"(Apollo returned no more results or budget exhausted)"
        )

    return city_leads, page, stats
