"""Tilnest website assistant tools — inbound voice only."""

TILNEST_TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "end_call",
            "description": "End the voice session after a polite goodbye.",
            "parameters": {"type": "object", "properties": {}},
        },
    },
]

TILNEST_PHONE_TOOLS = TILNEST_TOOLS
