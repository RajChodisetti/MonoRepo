"""Persist Apollo pagination and run summaries for daily ingestion."""

from __future__ import annotations

import json
from datetime import datetime, timezone
from pathlib import Path

STATE_DIR = Path(__file__).resolve().parent / "state"
DEFAULT_STATE_FILE = STATE_DIR / "ingestion_state.json"


def _now_iso() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def load_state(path: Path | None = None) -> dict:
    path = path or DEFAULT_STATE_FILE
    if not path.is_file():
        return {}
    return json.loads(path.read_text(encoding="utf-8"))


def save_state(state: dict, path: Path | None = None) -> Path:
    path = path or DEFAULT_STATE_FILE
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(state, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    return path


def get_apollo_page(state: dict, niche: str, city: str) -> int:
    niche_state = state.get(niche) or {}
    city_key = city.split(",")[0].strip().lower().replace(" ", "_")
    entry = niche_state.get(city_key) or {}
    page = entry.get("apollo_page", 1)
    try:
        return max(1, int(page))
    except (TypeError, ValueError):
        return 1


def set_apollo_page(state: dict, niche: str, city: str, page: int) -> None:
    niche_state = state.setdefault(niche, {})
    city_key = city.split(",")[0].strip().lower().replace(" ", "_")
    entry = niche_state.setdefault(city_key, {})
    entry["apollo_page"] = max(1, int(page))
    entry["last_run_at"] = _now_iso()


def record_run_summary(
    state: dict,
    niche: str,
    city: str,
    summary: dict,
) -> None:
    niche_state = state.setdefault(niche, {})
    city_key = city.split(",")[0].strip().lower().replace(" ", "_")
    entry = niche_state.setdefault(city_key, {})
    entry["last_summary"] = summary
    entry["last_run_at"] = _now_iso()
