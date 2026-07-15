"""Known business registry for database- and file-backed scrape deduplication."""

from __future__ import annotations

import json
import os
from dataclasses import dataclass, field
from pathlib import Path

SCRIPT_DIR = Path(__file__).resolve().parent
LEADS_DIR = SCRIPT_DIR / "leads"
DATA_DIR = SCRIPT_DIR / "data"


def _apollo_id_from_lead_dict(data: dict) -> str:
    source = data.get("source") or {}
    person_id = (source.get("person_id") or "").strip()
    if person_id:
        return person_id
    extra = data.get("extra") or {}
    return (extra.get("apollo_id") or "").strip()


def _place_id_from_lead_dict(data: dict) -> str:
    google = data.get("google") or {}
    if google.get("place_id"):
        return str(google["place_id"]).strip()
    source = data.get("source") or {}
    return str(source.get("place_id") or "").strip()


def _dedup_key_from_lead_dict(data: dict) -> str:
    contact = data.get("contact") or {}
    company = data.get("company") or {}
    email = (contact.get("email") or "").strip()
    if email:
        return email.lower()
    return f"{company.get('name', '')}|{contact.get('name', '')}|{company.get('domain', '')}".lower()


@dataclass
class KnownLeadsRegistry:
    place_ids: set[str] = field(default_factory=set)
    apollo_ids: set[str] = field(default_factory=set)
    dedup_keys: set[str] = field(default_factory=set)
    apollo_id_to_place_id: dict[str, str] = field(default_factory=dict)

    def is_apollo_known(self, apollo_id: str) -> bool:
        return bool(apollo_id) and apollo_id in self.apollo_ids

    def is_place_known(self, place_id: str) -> bool:
        return bool(place_id) and place_id in self.place_ids

    def is_dedup_known(self, lead_dict: dict) -> bool:
        key = _dedup_key_from_lead_dict(lead_dict)
        return bool(key) and key in self.dedup_keys

    def should_skip_scrape(self, lead_dict: dict) -> tuple[bool, str]:
        place_id = _place_id_from_lead_dict(lead_dict)
        if self.is_place_known(place_id):
            return True, "known_place_id"
        apollo_id = _apollo_id_from_lead_dict(lead_dict)
        if apollo_id and apollo_id in self.apollo_id_to_place_id:
            mapped_place_id = self.apollo_id_to_place_id[apollo_id]
            if mapped_place_id:
                return True, "known_apollo_with_place_id"
        return False, ""

    def register_lead_dict(self, lead_dict: dict) -> None:
        key = _dedup_key_from_lead_dict(lead_dict)
        if key:
            self.dedup_keys.add(key)
        apollo_id = _apollo_id_from_lead_dict(lead_dict)
        if apollo_id:
            self.apollo_ids.add(apollo_id)

    def register_scrape_record(self, record: dict) -> None:
        google = record.get("google") or {}
        place_id = (google.get("place_id") or "").strip()
        if place_id:
            self.place_ids.add(place_id)
        apollo = record.get("apollo_lead") or {}
        apollo_id = _apollo_id_from_lead_dict(apollo)
        if apollo_id and place_id:
            self.apollo_id_to_place_id[apollo_id] = place_id
        if apollo_id:
            self.apollo_ids.add(apollo_id)

    @classmethod
    def load_from_db(
        cls,
        database_url: str | None = None,
        *,
        required: bool = False,
    ) -> KnownLeadsRegistry:
        registry = cls()
        url = (database_url or os.getenv("DATABASE_URL") or "").strip()
        if not url:
            if required:
                raise ValueError("DATABASE_URL is required for database-backed ingestion deduplication")
            return registry

        try:
            import psycopg2
        except ImportError as exc:
            if required:
                raise RuntimeError("psycopg2 is required for database-backed ingestion deduplication") from exc
            return registry

        try:
            conn = psycopg2.connect(url, connect_timeout=10)
        except Exception as exc:
            if required:
                raise RuntimeError("Could not connect to PostgreSQL for ingestion deduplication") from exc
            return registry

        try:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT google_place_id
                    FROM restaurant_profiles
                    WHERE google_place_id IS NOT NULL AND trim(google_place_id) <> ''
                    """
                )
                for (place_id,) in cur.fetchall():
                    registry.place_ids.add(str(place_id).strip())

                cur.execute(
                    """
                    SELECT apollo_lead->'source'->>'person_id',
                           apollo_lead->'extra'->>'apollo_id',
                           google_place_id
                    FROM restaurant_profiles
                    WHERE apollo_lead IS NOT NULL AND apollo_lead <> '{}'::jsonb
                    """
                )
                for source_id, extra_id, place_id in cur.fetchall():
                    apollo_id = (source_id or extra_id or "").strip()
                    if apollo_id:
                        registry.apollo_ids.add(apollo_id)
                        pid = (place_id or "").strip()
                        if pid:
                            registry.apollo_id_to_place_id[apollo_id] = pid

                cur.execute(
                    """
                    SELECT lower(trim(email))
                    FROM restaurants
                    WHERE trim(email) <> ''
                    """
                )
                for (email,) in cur.fetchall():
                    if email:
                        registry.dedup_keys.add(str(email).strip().lower())
        finally:
            conn.close()

        return registry

    @classmethod
    def load_from_files(
        cls,
        *,
        leads_dir: Path | None = None,
        data_dir: Path | None = None,
        niche_slug: str = "restaurant",
    ) -> KnownLeadsRegistry:
        registry = cls()
        leads_dir = leads_dir or LEADS_DIR
        data_dir = data_dir or DATA_DIR

        if leads_dir.is_dir():
            for path in sorted(leads_dir.glob("*.json")):
                try:
                    payload = json.loads(path.read_text(encoding="utf-8"))
                except (json.JSONDecodeError, OSError):
                    continue
                for lead in payload.get("leads") or []:
                    registry.register_lead_dict(lead)

        data_glob = "restaurants_data_*.json" if niche_slug == "restaurant" else f"{niche_slug}_data_*.json"
        if data_dir.is_dir():
            for path in sorted(data_dir.glob(data_glob)):
                try:
                    payload = json.loads(path.read_text(encoding="utf-8"))
                except (json.JSONDecodeError, OSError):
                    continue
                for record in payload.get("restaurants") or []:
                    registry.register_scrape_record(record)

        return registry

    @classmethod
    def load_combined(
        cls,
        niche_slug: str = "restaurant",
        *,
        require_database: bool = False,
    ) -> KnownLeadsRegistry:
        db_registry = cls.load_from_db(required=require_database)
        file_registry = cls.load_from_files(niche_slug=niche_slug)
        merged = cls()
        merged.place_ids = db_registry.place_ids | file_registry.place_ids
        merged.apollo_ids = db_registry.apollo_ids | file_registry.apollo_ids
        merged.dedup_keys = db_registry.dedup_keys | file_registry.dedup_keys
        merged.apollo_id_to_place_id = {**file_registry.apollo_id_to_place_id, **db_registry.apollo_id_to_place_id}
        return merged
