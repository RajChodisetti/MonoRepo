"""
Local real-estate inventory loader and query helpers.

Seed data: data/properties.json (structured mock inventory).
"""

from __future__ import annotations

import json
import re
import tempfile
import uuid
from datetime import date
from functools import lru_cache
from pathlib import Path
from typing import Any

_DATA_PATH = Path(__file__).parent / "data" / "properties.json"

# Fallback bands if JSON omits them
_DEFAULT_BUDGET_BANDS = [
    {"id": "under_50L", "label": "Under 50 lakh", "min_price": 0, "max_price": 5_000_000, "status": "sale"},
    {"id": "50L_1Cr", "label": "50 lakh to 1 crore", "min_price": 5_000_000, "max_price": 10_000_000, "status": "sale"},
    {"id": "1Cr_2Cr", "label": "1 to 2 crore", "min_price": 10_000_000, "max_price": 20_000_000, "status": "sale"},
    {"id": "above_2Cr", "label": "Above 2 crore", "min_price": 20_000_000, "max_price": None, "status": "sale"},
    {"id": "rent_under_25k", "label": "Rent under 25 thousand", "min_price": 0, "max_price": 25_000, "status": "rent"},
    {"id": "rent_25k_50k", "label": "Rent 25 to 50 thousand", "min_price": 25_000, "max_price": 50_000, "status": "rent"},
]


def _norm(s: str | None) -> str:
    return (s or "").strip().lower()


# Spoken / typed city aliases → canonical inventory name
_CITY_CANONICAL: dict[str, str] = {
    "bengaluru": "Bengaluru",
    "bangalore": "Bengaluru",
    "blr": "Bengaluru",
    "bengalooru": "Bengaluru",
    "hyderabad": "Hyderabad",
    "hyd": "Hyderabad",
    "secunderabad": "Hyderabad",
}


def canonicalize_city(city: str | None) -> str | None:
    """Map Bangalore/BLR/etc. to the catalog city name used in properties.json."""
    if city is None:
        return None
    raw = str(city).strip()
    if not raw:
        return None
    cleaned = re.sub(r"[^a-zA-Z]+", " ", raw).strip().lower()
    cleaned = re.sub(r"\s+city$", "", cleaned).strip()
    if cleaned in _CITY_CANONICAL:
        return _CITY_CANONICAL[cleaned]
    # Fuzzy: "bengaluruu" / "bangaluru" → Bengaluru
    for alias, canon in _CITY_CANONICAL.items():
        if alias in cleaned or cleaned in alias:
            return canon
    return raw


def city_matches(query: str | None, stored: str | None) -> bool:
    if not query or not str(query).strip():
        return True
    q = canonicalize_city(query) or str(query).strip()
    s = canonicalize_city(stored) or str(stored or "").strip()
    qn, sn = _norm(q), _norm(s)
    if not qn:
        return True
    if not sn:
        return False
    return qn == sn or qn in sn or sn in qn


# Back-compat for callers that imported the private name
_city_matches = city_matches


def format_price_inr(price: int | float | None, *, status: str = "sale") -> str:
    if price is None:
        return "price on request"
    n = float(price)
    if status == "rent":
        if n >= 100_000:
            return f"₹{n / 100_000:.1f} lakh per month".replace(".0 ", " ")
        return f"₹{int(n):,} per month"
    if n >= 10_000_000:
        return f"₹{n / 10_000_000:.2f} crore".replace(".00", "")
    if n >= 100_000:
        return f"₹{n / 100_000:.1f} lakh".replace(".0 ", " ")
    return f"₹{int(n):,}"


def _normalize_property(raw: dict[str, Any]) -> dict[str, Any]:
    """
    Accept nested mock schema or legacy flat schema; return a rich record
    that also exposes flat fields for filtering.
    """
    location = raw.get("location") if isinstance(raw.get("location"), dict) else {}
    specs = raw.get("specs") if isinstance(raw.get("specs"), dict) else {}
    pricing = raw.get("pricing") if isinstance(raw.get("pricing"), dict) else {}

    city = location.get("city") or raw.get("city")
    locality = location.get("locality") or raw.get("locality")
    state = location.get("state") or raw.get("state")
    pincode = location.get("pincode") or raw.get("pincode")

    bhk = specs.get("bhk") if "bhk" in specs else raw.get("bhk")
    size_sqft = specs.get("size_sqft") if "size_sqft" in specs else raw.get("size_sqft")
    size_sqyd = specs.get("size_sqyd") if "size_sqyd" in specs else raw.get("size_sqyd")
    furnishing = specs.get("furnishing") if "furnishing" in specs else raw.get("furnishing")
    facing = specs.get("facing") if "facing" in specs else raw.get("facing")
    floor = specs.get("floor") if "floor" in specs else raw.get("floor")
    total_floors = specs.get("total_floors") if "total_floors" in specs else raw.get("total_floors")
    possession = specs.get("possession") if "possession" in specs else raw.get("possession")

    price_inr = pricing.get("price_inr") if "price_inr" in pricing else raw.get("price_inr")
    negotiable = pricing.get("negotiable") if "negotiable" in pricing else raw.get("negotiable")

    amenities = raw.get("amenities") or raw.get("features") or []
    nearby = raw.get("nearby") or []
    status = raw.get("status") or "sale"

    return {
        "id": raw["id"],
        "title": raw.get("title") or raw["id"],
        "type": raw.get("type"),
        "status": status,
        "description": raw.get("description") or "",
        # Nested (canonical mock shape)
        "location": {
            "city": city,
            "locality": locality,
            "state": state,
            "pincode": pincode,
        },
        "specs": {
            "bhk": bhk,
            "size_sqft": size_sqft,
            "size_sqyd": size_sqyd,
            "furnishing": furnishing,
            "facing": facing,
            "floor": floor,
            "total_floors": total_floors,
            "possession": possession,
        },
        "pricing": {
            "price_inr": price_inr,
            "price_display": format_price_inr(price_inr, status=status),
            "negotiable": bool(negotiable) if negotiable is not None else False,
            "currency": "INR",
        },
        "amenities": list(amenities),
        "nearby": list(nearby),
        # Flat aliases for search / older callers
        "city": city,
        "locality": locality,
        "bhk": bhk,
        "size_sqft": size_sqft,
        "size_sqyd": size_sqyd,
        "price_inr": price_inr,
        "features": list(amenities),
    }


@lru_cache(maxsize=1)
def _load_raw_catalog() -> dict[str, Any]:
    with _DATA_PATH.open(encoding="utf-8") as f:
        data = json.load(f)

    if isinstance(data, list):
        return {
            "version": "legacy",
            "source": "tuvi_mock_inventory",
            "currency": "INR",
            "budget_bands": _DEFAULT_BUDGET_BANDS,
            "properties": data,
        }

    if not isinstance(data, dict) or "properties" not in data:
        raise ValueError("properties.json must be a list or {properties: [...]}")

    bands = data.get("budget_bands") or _DEFAULT_BUDGET_BANDS
    return {
        "version": data.get("version") or "1.0",
        "source": data.get("source") or "tuvi_mock_inventory",
        "currency": data.get("currency") or "INR",
        "updated_at": data.get("updated_at"),
        "budget_bands": bands,
        "properties": data.get("properties") or [],
    }


@lru_cache(maxsize=1)
def load_properties() -> tuple[dict[str, Any], ...]:
    catalog = _load_raw_catalog()
    return tuple(_normalize_property(p) for p in catalog["properties"] if p.get("id"))


def budget_bands() -> list[dict[str, Any]]:
    return list(_load_raw_catalog().get("budget_bands") or _DEFAULT_BUDGET_BANDS)


def catalog_meta() -> dict[str, Any]:
    catalog = _load_raw_catalog()
    return {
        "version": catalog.get("version"),
        "source": catalog.get("source"),
        "currency": catalog.get("currency"),
        "updated_at": catalog.get("updated_at"),
        "count": len(load_properties()),
    }


def reload_properties() -> list[dict[str, Any]]:
    _load_raw_catalog.cache_clear()
    load_properties.cache_clear()
    return list(load_properties())


def _slug_part(value: str | None, fallback: str) -> str:
    text = re.sub(r"[^a-z0-9]+", "-", (value or "").strip().lower()).strip("-")
    return text[:24] or fallback


def _generate_property_id(payload: dict[str, Any]) -> str:
    raw_city = (
        (payload.get("location") or {}).get("city")
        if isinstance(payload.get("location"), dict)
        else payload.get("city")
    )
    city = _slug_part(canonicalize_city(raw_city) or raw_city, "city")
    ptype = _slug_part(payload.get("type") or payload.get("property_type"), "prop")
    short = uuid.uuid4().hex[:6]
    return f"{city[:3]}-{ptype[:3]}-{short}"


def coerce_storage_property(payload: dict[str, Any], *, existing: dict[str, Any] | None = None) -> dict[str, Any]:
    """Normalize admin/API payload into nested catalog property shape."""
    base = dict(existing or {})
    location_in = payload.get("location") if isinstance(payload.get("location"), dict) else {}
    specs_in = payload.get("specs") if isinstance(payload.get("specs"), dict) else {}
    pricing_in = payload.get("pricing") if isinstance(payload.get("pricing"), dict) else {}
    base_loc = base.get("location") if isinstance(base.get("location"), dict) else {}
    base_specs = base.get("specs") if isinstance(base.get("specs"), dict) else {}
    base_pricing = base.get("pricing") if isinstance(base.get("pricing"), dict) else {}

    prop_id = (payload.get("id") or base.get("id") or "").strip()
    if not prop_id:
        prop_id = _generate_property_id(payload)

    amenities = payload.get("amenities")
    if amenities is None:
        amenities = base.get("amenities") or []
    if isinstance(amenities, str):
        amenities = [a.strip() for a in amenities.split(",") if a.strip()]

    nearby = payload.get("nearby")
    if nearby is None:
        nearby = base.get("nearby") or []
    if isinstance(nearby, str):
        nearby = [a.strip() for a in nearby.split(",") if a.strip()]

    bhk = specs_in.get("bhk", payload.get("bhk", base_specs.get("bhk")))
    size_sqft = specs_in.get("size_sqft", payload.get("size_sqft", base_specs.get("size_sqft")))
    size_sqyd = specs_in.get("size_sqyd", payload.get("size_sqyd", base_specs.get("size_sqyd")))
    price_inr = pricing_in.get("price_inr", payload.get("price_inr", base_pricing.get("price_inr")))

    return {
        "id": prop_id,
        "title": payload.get("title") or base.get("title") or prop_id,
        "type": payload.get("type") or base.get("type") or "apartment",
        "status": payload.get("status") or base.get("status") or "sale",
        "location": {
            "city": canonicalize_city(
                location_in.get("city", payload.get("city", base_loc.get("city")))
            )
            or "",
            "locality": location_in.get("locality", payload.get("locality", base_loc.get("locality"))) or "",
            "state": location_in.get("state", payload.get("state", base_loc.get("state"))) or "",
            "pincode": location_in.get("pincode", payload.get("pincode", base_loc.get("pincode"))) or "",
        },
        "specs": {
            "bhk": int(bhk) if bhk not in (None, "") else None,
            "size_sqft": int(size_sqft) if size_sqft not in (None, "") else None,
            "size_sqyd": int(size_sqyd) if size_sqyd not in (None, "") else None,
            "furnishing": specs_in.get("furnishing", payload.get("furnishing", base_specs.get("furnishing"))) or "",
            "facing": specs_in.get("facing", payload.get("facing", base_specs.get("facing"))) or "",
            "floor": specs_in.get("floor", payload.get("floor", base_specs.get("floor"))),
            "total_floors": specs_in.get("total_floors", payload.get("total_floors", base_specs.get("total_floors"))),
            "possession": specs_in.get("possession", payload.get("possession", base_specs.get("possession"))) or "",
        },
        "pricing": {
            "price_inr": int(price_inr) if price_inr not in (None, "") else None,
            "negotiable": bool(
                pricing_in.get(
                    "negotiable",
                    payload.get("negotiable", base_pricing.get("negotiable", True)),
                )
            ),
        },
        "amenities": list(amenities),
        "nearby": list(nearby),
        "description": payload.get("description", base.get("description")) or "",
    }


def _write_catalog(catalog: dict[str, Any]) -> None:
    catalog = dict(catalog)
    catalog["updated_at"] = date.today().isoformat()
    _DATA_PATH.parent.mkdir(parents=True, exist_ok=True)
    raw = json.dumps(catalog, indent=2, ensure_ascii=False) + "\n"
    with tempfile.NamedTemporaryFile(
        "w",
        encoding="utf-8",
        dir=str(_DATA_PATH.parent),
        prefix=".properties-",
        suffix=".tmp",
        delete=False,
    ) as tmp:
        tmp.write(raw)
        tmp_path = Path(tmp.name)
    tmp_path.replace(_DATA_PATH)
    reload_properties()


def list_properties_admin() -> list[dict[str, Any]]:
    return [summarize_property(p) for p in load_properties()]


def create_property(payload: dict[str, Any]) -> dict[str, Any]:
    catalog = dict(_load_raw_catalog())
    props = list(catalog.get("properties") or [])
    stored = coerce_storage_property(payload)
    if any(_norm(p.get("id")) == _norm(stored["id"]) for p in props):
        raise ValueError(f"property id already exists: {stored['id']}")
    props.append(stored)
    catalog["properties"] = props
    _write_catalog(catalog)
    return get_property(stored["id"]) or stored


def update_property(property_id: str, payload: dict[str, Any]) -> dict[str, Any]:
    catalog = dict(_load_raw_catalog())
    props = list(catalog.get("properties") or [])
    pid = _norm(property_id)
    idx = next((i for i, p in enumerate(props) if _norm(p.get("id")) == pid), None)
    if idx is None:
        raise KeyError(property_id)
    merged = coerce_storage_property(payload, existing=props[idx])
    # Keep original id unless explicitly changing to a free id
    new_id = (payload.get("id") or "").strip()
    if new_id and _norm(new_id) != pid:
        if any(_norm(p.get("id")) == _norm(new_id) for j, p in enumerate(props) if j != idx):
            raise ValueError(f"property id already exists: {new_id}")
        merged["id"] = new_id
    else:
        merged["id"] = props[idx]["id"]
    props[idx] = merged
    catalog["properties"] = props
    _write_catalog(catalog)
    return get_property(merged["id"]) or merged


def delete_property(property_id: str) -> bool:
    catalog = dict(_load_raw_catalog())
    props = list(catalog.get("properties") or [])
    pid = _norm(property_id)
    next_props = [p for p in props if _norm(p.get("id")) != pid]
    if len(next_props) == len(props):
        return False
    catalog["properties"] = next_props
    _write_catalog(catalog)
    return True


# Back-compat alias used by older imports / docs
BUDGET_BANDS = _DEFAULT_BUDGET_BANDS


def summarize_property(p: dict[str, Any]) -> dict[str, Any]:
    """Compact spoken-friendly summary for the LLM."""
    status = p.get("status") or "sale"
    size_bits = []
    if p.get("bhk"):
        size_bits.append(f"{p['bhk']} BHK")
    if p.get("size_sqft"):
        size_bits.append(f"{p['size_sqft']} sq ft")
    if p.get("size_sqyd"):
        size_bits.append(f"{p['size_sqyd']} sq yd")
    return {
        "id": p["id"],
        "title": p.get("title") or p["id"],
        "type": p.get("type"),
        "status": status,
        "city": p.get("city"),
        "locality": p.get("locality"),
        "size": ", ".join(size_bits) if size_bits else None,
        "price": format_price_inr(p.get("price_inr"), status=status),
        "price_inr": p.get("price_inr"),
        "features": (p.get("features") or p.get("amenities") or [])[:4],
        "nearby": (p.get("nearby") or [])[:4],
        "possession": (p.get("specs") or {}).get("possession"),
        "furnishing": (p.get("specs") or {}).get("furnishing"),
    }


def get_meta() -> dict[str, Any]:
    props = load_properties()
    cities: dict[str, set[str]] = {}
    types: set[str] = set()
    statuses: set[str] = set()
    for p in props:
        city = p.get("city") or ""
        loc = p.get("locality") or ""
        if city:
            cities.setdefault(city, set())
            if loc:
                cities[city].add(loc)
        if p.get("type"):
            types.add(p["type"])
        if p.get("status"):
            statuses.add(p["status"])
    return {
        **catalog_meta(),
        "cities": [
            {"name": city, "localities": sorted(locs)}
            for city, locs in sorted(cities.items(), key=lambda x: x[0])
        ],
        "types": sorted(types),
        "statuses": sorted(statuses),
        "budget_bands": budget_bands(),
    }


def get_property(property_id: str) -> dict[str, Any] | None:
    pid = _norm(property_id)
    for p in load_properties():
        if _norm(p.get("id")) == pid:
            return dict(p)
    return None


def search_properties(
    *,
    city: str | None = None,
    locality: str | None = None,
    property_type: str | None = None,
    bhk: int | None = None,
    min_price: int | float | None = None,
    max_price: int | float | None = None,
    status: str | None = None,
    budget_band: str | None = None,
    limit: int = 5,
) -> list[dict[str, Any]]:
    props = list(load_properties())
    bands = budget_bands()

    if budget_band:
        band = next((b for b in bands if b["id"] == budget_band), None)
        if band:
            if min_price is None:
                min_price = band.get("min_price")
            if max_price is None:
                max_price = band.get("max_price")
            if status is None and band.get("status"):
                status = band["status"]

    loc_n = _norm(locality)
    type_n = _norm(property_type)
    status_n = _norm(status)
    city_query = canonicalize_city(city) or city

    results: list[dict[str, Any]] = []
    for p in props:
        if not city_matches(city_query, p.get("city")):
            continue
        if loc_n and loc_n not in _norm(p.get("locality")):
            continue
        if type_n and type_n != _norm(p.get("type")):
            continue
        if status_n and status_n != _norm(p.get("status")):
            continue
        if bhk is not None:
            pbhk = p.get("bhk")
            if pbhk is None or int(pbhk) != int(bhk):
                continue
        price = p.get("price_inr")
        if price is not None:
            if min_price is not None and price < float(min_price):
                continue
            if max_price is not None and price > float(max_price):
                continue
        results.append(p)

    if status_n != "rent":
        results.sort(key=lambda x: (0 if x.get("status") == "sale" else 1, x.get("price_inr") or 0))
    else:
        results.sort(key=lambda x: x.get("price_inr") or 0)

    lim = max(1, min(int(limit or 5), 10))
    return [summarize_property(p) for p in results[:lim]]
