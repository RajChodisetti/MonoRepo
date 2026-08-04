"""
Dummy listings API used by Ananya voice tools.

All search / details responses come from data/properties.json (mock inventory).
"""

from __future__ import annotations

from typing import Any

import property_store


def api_meta(*, city: str | None = None) -> dict[str, Any]:
    meta = property_store.get_meta()
    if city and city.strip():
        wanted = property_store.canonicalize_city(city) or city.strip()
        match = next(
            (
                c
                for c in meta["cities"]
                if property_store.city_matches(wanted, c["name"])
            ),
            None,
        )
        if match:
            return {
                "ok": True,
                "source": "mock",
                "city": match["name"],
                "localities": match["localities"],
                "types": meta["types"],
                "statuses": meta["statuses"],
                "budget_bands": meta["budget_bands"],
                "catalog": {
                    "version": meta.get("version"),
                    "updated_at": meta.get("updated_at"),
                    "count": meta.get("count"),
                },
            }
        return {
            "ok": True,
            "source": "mock",
            "message": f"No localities for {city}. Available cities listed below.",
            "cities": meta["cities"],
            "types": meta["types"],
            "budget_bands": meta["budget_bands"],
        }
    return {"ok": True, "source": "mock", **meta}


def api_search(
    *,
    city: str | None = None,
    locality: str | None = None,
    property_type: str | None = None,
    bhk: int | None = None,
    min_price: float | None = None,
    max_price: float | None = None,
    status: str | None = None,
    budget_band: str | None = None,
    limit: int = 3,
) -> dict[str, Any]:
    query = {
        "city": city,
        "locality": locality,
        "property_type": property_type,
        "bhk": bhk,
        "min_price": min_price,
        "max_price": max_price,
        "status": status,
        "budget_band": budget_band,
        "limit": limit,
    }

    def _run(**kwargs):
        return property_store.search_properties(limit=limit, **kwargs)

    # Progressive widen: keep city/type when possible so "find plot in Hyderabad"
    # still returns plots even if locality/status was too strict.
    attempts: list[tuple[str | None, dict[str, Any]]] = [
        (
            None,
            {
                "city": city,
                "locality": locality,
                "property_type": property_type,
                "bhk": bhk,
                "min_price": min_price,
                "max_price": max_price,
                "status": status,
                "budget_band": budget_band,
            },
        ),
    ]
    if locality:
        attempts.append(
            (
                "locality",
                {
                    "city": city,
                    "locality": None,
                    "property_type": property_type,
                    "bhk": bhk,
                    "min_price": min_price,
                    "max_price": max_price,
                    "status": status,
                    "budget_band": budget_band,
                },
            )
        )
    if status:
        attempts.append(
            (
                "status",
                {
                    "city": city,
                    "locality": None if locality else locality,
                    "property_type": property_type,
                    "bhk": bhk,
                    "min_price": min_price,
                    "max_price": max_price,
                    "status": None,
                    "budget_band": None,
                },
            )
        )
    if bhk is not None:
        attempts.append(
            (
                "bhk",
                {
                    "city": city,
                    "locality": None,
                    "property_type": property_type,
                    "bhk": None,
                    "min_price": min_price,
                    "max_price": max_price,
                    "status": None,
                    "budget_band": None,
                },
            )
        )
    if min_price is not None or max_price is not None or budget_band:
        attempts.append(
            (
                "budget",
                {
                    "city": city,
                    "locality": None,
                    "property_type": property_type,
                    "bhk": None,
                    "min_price": None,
                    "max_price": None,
                    "status": None,
                    "budget_band": None,
                },
            )
        )

    matches: list[dict[str, Any]] = []
    relaxed: list[str] = []
    used_filters = attempts[0][1]
    for reason, filters in attempts:
        matches = _run(**filters)
        used_filters = filters
        if matches:
            if reason:
                relaxed.append(reason)
            break
        if reason:
            relaxed.append(reason)

    meta = property_store.get_meta()
    available_cities = [c["name"] for c in meta.get("cities") or []]
    resolved_city = property_store.canonicalize_city(city) if city else None
    if city and resolved_city and resolved_city != city:
        query["city_resolved"] = resolved_city

    # Localities that actually have this type in the city (for helpful empty/widen speech)
    type_localities: list[str] = []
    if property_type and (resolved_city or city):
        want_city = resolved_city or city
        want_type = property_type.strip().lower()
        for p in property_store.load_properties():
            if property_store.city_matches(want_city, p.get("city")) and (
                p.get("type") or ""
            ).lower() == want_type:
                loc = p.get("locality")
                if loc and loc not in type_localities:
                    type_localities.append(loc)

    widened = bool(matches and relaxed)
    if matches and widened:
        message = (
            f"No exact match for the strict filters; widened by dropping {', '.join(relaxed)}. "
            f"Found {len(matches)} nearby options from mock inventory."
        )
    elif matches:
        message = f"Found {len(matches)} matching properties from mock inventory."
    else:
        hint = ""
        if type_localities:
            hint = f" For {property_type} in {resolved_city or city}, try: {', '.join(type_localities[:5])}."
        message = (
            "No matches in mock inventory. Available cities: "
            + (", ".join(available_cities) if available_cities else "none")
            + "."
            + hint
            + " Plots are typically listed for sale (not rent)."
        )

    return {
        "ok": True,
        "source": "mock",
        "query": {k: v for k, v in query.items() if v is not None},
        "applied_filters": {k: v for k, v in used_filters.items() if v is not None},
        "relaxed_filters": relaxed,
        "widened": widened,
        "count": len(matches),
        "results": matches,
        "available_cities": available_cities,
        "type_localities": type_localities,
        "message": message,
    }


def api_details(property_id: str) -> dict[str, Any]:
    pid = (property_id or "").strip()
    prop = property_store.get_property(pid) if pid else None
    if not prop:
        return {
            "ok": False,
            "source": "mock",
            "error": "not_found",
            "message": f"Property '{pid}' not found. Use an id from search results.",
        }
    return {
        "ok": True,
        "source": "mock",
        "property": prop,
        "summary": property_store.summarize_property(prop),
    }


def api_create(payload: dict[str, Any]) -> dict[str, Any]:
    try:
        created = property_store.create_property(payload)
    except ValueError as exc:
        return {"ok": False, "source": "mock", "error": "conflict", "message": str(exc)}
    except Exception as exc:
        return {"ok": False, "source": "mock", "error": "failed", "message": str(exc)}
    return {
        "ok": True,
        "source": "mock",
        "property": created,
        "summary": property_store.summarize_property(created),
        "message": f"Created listing {created.get('id')}.",
    }


def api_update(property_id: str, payload: dict[str, Any]) -> dict[str, Any]:
    try:
        updated = property_store.update_property(property_id, payload)
    except KeyError:
        return {
            "ok": False,
            "source": "mock",
            "error": "not_found",
            "message": f"Property '{property_id}' not found.",
        }
    except ValueError as exc:
        return {"ok": False, "source": "mock", "error": "conflict", "message": str(exc)}
    except Exception as exc:
        return {"ok": False, "source": "mock", "error": "failed", "message": str(exc)}
    return {
        "ok": True,
        "source": "mock",
        "property": updated,
        "summary": property_store.summarize_property(updated),
        "message": f"Updated listing {updated.get('id')}.",
    }


def api_delete(property_id: str, *, confirmed: bool) -> dict[str, Any]:
    if not confirmed:
        return {
            "ok": False,
            "source": "mock",
            "error": "confirmation_required",
            "message": "Delete not confirmed. Ask the user to confirm, then call again with confirmed=true.",
        }
    if not property_store.delete_property(property_id):
        return {
            "ok": False,
            "source": "mock",
            "error": "not_found",
            "message": f"Property '{property_id}' not found.",
        }
    return {
        "ok": True,
        "source": "mock",
        "deleted": property_id,
        "message": f"Deleted listing {property_id}.",
    }
