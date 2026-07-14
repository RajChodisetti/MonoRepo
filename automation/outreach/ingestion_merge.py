"""Merge new leads and scrape records into existing JSON files."""

from __future__ import annotations

import json
from datetime import datetime, timezone
from pathlib import Path

from scrape_safety import sanitize_sensitive_urls
from tuvi_outreach_agent import Lead, _lead_dedup_key, lead_to_dict


def _dedup_key_from_lead_dict(data: dict) -> str:
    company = data.get("company") or {}
    contact = data.get("contact") or {}
    email = (contact.get("email") or "").strip().lower()
    if email:
        return email
    return f"{company.get('name', '')}|{contact.get('name', '')}|{company.get('domain', '')}".lower()


def merge_leads_file(
    path: Path,
    new_leads: list[Lead],
    *,
    cities: list[str],
    niche_type: str,
    per_page: int,
    max_pages: int,
    target_per_city: int | None,
) -> int:
    existing: list[dict] = []
    if path.is_file():
        payload = json.loads(path.read_text(encoding="utf-8"))
        existing = list(payload.get("leads") or [])

    seen = {_dedup_key_from_lead_dict(item) for item in existing}
    added = 0
    for lead in new_leads:
        key = _lead_dedup_key(lead)
        if key in seen:
            continue
        seen.add(key)
        existing.append(lead_to_dict(lead))
        added += 1

    leads_by_city: dict[str, int] = {}
    for item in existing:
        loc = item.get("location") or {}
        city = loc.get("city") or "Unknown"
        leads_by_city[city] = leads_by_city.get(city, 0) + 1

    document = {
        "meta": {
            "version": "1.0",
            "fetched_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "source": "apollo.io",
            "lead_type": niche_type,
            "total_leads": len(existing),
            "cities_searched": cities,
            "leads_by_city": dict(sorted(leads_by_city.items())),
            "fetch_settings": {
                "per_page": per_page,
                "max_pages": max_pages,
                "target_per_city": target_per_city,
            },
        },
        "leads": existing,
    }
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    return added


def merge_scrape_file(path: Path, new_records: list[dict]) -> int:
    existing: list[dict] = []
    meta: dict = {}
    if path.is_file():
        payload = sanitize_sensitive_urls(json.loads(path.read_text(encoding="utf-8")))
        existing = list(payload.get("restaurants") or [])
        meta = dict(payload.get("meta") or {})

    seen_place: set[str] = set()
    seen_name_city: set[str] = set()
    for item in existing:
        google = item.get("google") or {}
        pid = (google.get("place_id") or "").strip()
        if pid:
            seen_place.add(pid)
        loc = item.get("location") or {}
        seen_name_city.add(f"{item.get('name', '')}|{loc.get('city', '')}".lower())

    added = 0
    merged = list(existing)
    for raw_record in new_records:
        record = sanitize_sensitive_urls(raw_record)
        google = record.get("google") or {}
        pid = (google.get("place_id") or "").strip()
        loc = record.get("location") or {}
        name_key = f"{record.get('name', '')}|{loc.get('city', '')}".lower()
        if pid and pid in seen_place:
            continue
        if name_key in seen_name_city:
            continue
        if pid:
            seen_place.add(pid)
        seen_name_city.add(name_key)
        merged.append(record)
        added += 1

    meta.setdefault("version", "1.0")
    meta["total_scraped"] = len(merged)
    document = {"meta": meta, "restaurants": merged}
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(document, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    return added
