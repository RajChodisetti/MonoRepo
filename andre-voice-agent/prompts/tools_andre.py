"""OpenAI function tools for Andre real-estate voice agent."""

from __future__ import annotations

ANDRE_TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "get_property_options",
            "description": (
                "Meta endpoint for cities/localities/types/budget bands. "
                "Use ONLY when the user asks which areas/cities exist. "
                "Do NOT use this when they ask to find/search listings — use search_properties instead."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "city": {
                        "type": "string",
                        "description": "Optional city to list localities for (e.g. Hyderabad).",
                    }
                },
                "required": [],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "search_properties",
            "description": (
                "Search mock inventory for listings. Call this whenever the user wants to find / "
                "dhoondna properties (city and/or type is enough). Partial filters are fine; "
                "the API auto-widens if locality/status is too strict. Speak ONLY from returned results."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "city": {
                        "type": "string",
                        "description": "City name. Bangalore and Bengaluru are the same; Hyd = Hyderabad.",
                    },
                    "locality": {"type": "string", "description": "Area / locality within the city"},
                    "property_type": {
                        "type": "string",
                        "enum": ["apartment", "villa", "plot"],
                        "description": "Property type",
                    },
                    "bhk": {"type": "integer", "description": "Number of bedrooms (apartments/villas)"},
                    "min_price": {"type": "number", "description": "Minimum price in INR"},
                    "max_price": {"type": "number", "description": "Maximum price in INR"},
                    "status": {
                        "type": "string",
                        "enum": ["sale", "rent"],
                        "description": "sale or rent",
                    },
                    "budget_band": {
                        "type": "string",
                        "description": "Optional band id from get_property_options, e.g. 50L_1Cr",
                    },
                    "limit": {
                        "type": "integer",
                        "description": "Max results (1–5). Default 3.",
                    },
                },
                "required": [],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "get_property_details",
            "description": (
                "Call the dummy listings details API for one mock listing id "
                "from a previous search_properties result."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "property_id": {
                        "type": "string",
                        "description": "Listing id, e.g. hyd-apt-001",
                    }
                },
                "required": ["property_id"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "create_property",
            "description": (
                "Add a new listing to the mock inventory when the user explicitly asks to create / add "
                "a property. Collect at least title, type, status (sale/rent), city, and price first. "
                "Never invent silently — only create after the user provides the details."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "title": {"type": "string", "description": "Listing title"},
                    "property_type": {
                        "type": "string",
                        "enum": ["apartment", "villa", "plot"],
                    },
                    "status": {"type": "string", "enum": ["sale", "rent"]},
                    "city": {"type": "string"},
                    "locality": {"type": "string"},
                    "state": {"type": "string"},
                    "pincode": {"type": "string"},
                    "bhk": {"type": "integer"},
                    "size_sqft": {"type": "integer"},
                    "size_sqyd": {"type": "integer"},
                    "price_inr": {"type": "number", "description": "Price in INR"},
                    "description": {"type": "string"},
                    "amenities": {
                        "type": "string",
                        "description": "Comma-separated amenities",
                    },
                    "nearby": {
                        "type": "string",
                        "description": "Comma-separated nearby landmarks",
                    },
                    "furnishing": {"type": "string"},
                    "possession": {"type": "string"},
                    "negotiable": {"type": "boolean"},
                },
                "required": ["title", "property_type", "status", "city", "price_inr"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "update_property",
            "description": (
                "Update an existing listing when the user asks to change price, locality, title, etc. "
                "Use a real property_id from search/details. Only send fields that should change."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "property_id": {
                        "type": "string",
                        "description": "Existing listing id to update",
                    },
                    "title": {"type": "string"},
                    "property_type": {
                        "type": "string",
                        "enum": ["apartment", "villa", "plot"],
                    },
                    "status": {"type": "string", "enum": ["sale", "rent"]},
                    "city": {"type": "string"},
                    "locality": {"type": "string"},
                    "state": {"type": "string"},
                    "pincode": {"type": "string"},
                    "bhk": {"type": "integer"},
                    "size_sqft": {"type": "integer"},
                    "size_sqyd": {"type": "integer"},
                    "price_inr": {"type": "number"},
                    "description": {"type": "string"},
                    "amenities": {"type": "string"},
                    "nearby": {"type": "string"},
                    "furnishing": {"type": "string"},
                    "possession": {"type": "string"},
                    "negotiable": {"type": "boolean"},
                },
                "required": ["property_id"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "delete_property",
            "description": (
                "Delete a listing from inventory. Prefer property_id from a prior search. "
                "If the user pointed at a listing by place (e.g. first plot / KPHB / Attibele), "
                "pass city/locality/property_type to resolve it. "
                "Pass confirmed=true when they already said delete this / yes / confirm / go ahead / delete karo."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "property_id": {
                        "type": "string",
                        "description": "Listing id to delete (from search_properties)",
                    },
                    "city": {"type": "string", "description": "Optional city to resolve the listing"},
                    "locality": {
                        "type": "string",
                        "description": "Optional locality / area name to resolve the listing",
                    },
                    "property_type": {
                        "type": "string",
                        "enum": ["apartment", "villa", "plot"],
                    },
                    "confirmed": {
                        "type": "boolean",
                        "description": (
                            "true when user clearly asked to delete this listing or confirmed "
                            "(yes / confirm / go ahead / delete karo / just delete it)"
                        ),
                    },
                },
                "required": ["confirmed"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "set_language",
            "description": (
                "Lock the session language for STT/TTS (girl voice per language) and replies. "
                "Call ONLY when the user explicitly asks to switch language — "
                "never because they spoke a different language while locked."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "language": {
                        "type": "string",
                        "enum": ["auto", "en", "hi", "te"],
                        "description": "auto | en (English) | hi (Hindi) | te (Telugu)",
                    }
                },
                "required": ["language"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "place_callback_call",
            "description": (
                "Place an outbound phone call so Ananya (the AI assistant) calls the user. "
                "Browser sessions only. Confirm the number first."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "phone_number": {
                        "type": "string",
                        "description": "E.164 phone number, e.g. +919876543210",
                    },
                    "name": {
                        "type": "string",
                        "description": "Optional caller name",
                    },
                    "language": {
                        "type": "string",
                        "enum": ["auto", "en", "hi", "te"],
                        "description": "Language for the outbound phone session",
                    },
                },
                "required": ["phone_number"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "mark_do_not_call",
            "description": (
                "Record an opt-out / do-not-call request immediately. "
                "Call before end_call when the user opts out."
            ),
            "parameters": {"type": "object", "properties": {}, "required": []},
        },
    },
    {
        "type": "function",
        "function": {
            "name": "end_call",
            "description": "End the call or browser session after a polite goodbye.",
            "parameters": {"type": "object", "properties": {}, "required": []},
        },
    },
]

ANDRE_PHONE_TOOLS = ANDRE_TOOLS
ANDRE_BROWSER_TOOLS = ANDRE_TOOLS


def andre_tools_schema(tools: list | None = None):
    """Convert OpenAI-style tool dicts into Pipecat ToolsSchema (required on LLMContext)."""
    from pipecat.adapters.schemas.function_schema import FunctionSchema
    from pipecat.adapters.schemas.tools_schema import ToolsSchema

    schemas = []
    for tool in tools or ANDRE_TOOLS:
        fn = tool.get("function") or {}
        params = fn.get("parameters") or {}
        schemas.append(
            FunctionSchema(
                name=fn["name"],
                description=fn.get("description") or "",
                properties=params.get("properties") or {},
                required=list(params.get("required") or []),
            )
        )
    return ToolsSchema(standard_tools=schemas)
