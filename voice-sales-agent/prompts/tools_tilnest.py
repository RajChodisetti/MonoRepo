"""Tilnest website assistant tools — waitlist FAQ and optional callback."""

TILNEST_TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "place_callback_call",
            "description": (
                "Dial the visitor's phone so the Tilnest AI phone assistant calls them. "
                "Use ONLY after they ask to be called back (or say 'call me') and you have "
                "confirmed their phone number aloud. Browser sessions only."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "phone_number": {
                        "type": "string",
                        "description": "Visitor phone in E.164 if possible (e.g. +61412345678).",
                    },
                    "name": {
                        "type": "string",
                        "description": "Optional visitor name.",
                    },
                },
                "required": ["phone_number"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "end_call",
            "description": "End the voice session after a polite goodbye.",
            "parameters": {"type": "object", "properties": {}},
        },
    },
]

_PHONE_EXCLUDED = {"place_callback_call"}
TILNEST_PHONE_TOOLS = [
    t
    for t in TILNEST_TOOLS
    if t.get("function", {}).get("name") not in _PHONE_EXCLUDED
]
