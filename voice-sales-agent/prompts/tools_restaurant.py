"""Restaurant receptionist tool schemas (demo + API tools)."""

RESTAURANT_TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "demo_book_table",
            "description": (
                "Complete a demo table booking after collecting party size, date, and time "
                "from the guest. Call ONLY when all three are known — never on the first request."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "party_size": {
                        "type": "integer",
                        "description": "Number of guests dining.",
                    },
                    "date": {
                        "type": "string",
                        "description": "Reservation date in YYYY-MM-DD format.",
                    },
                    "time": {
                        "type": "string",
                        "description": "Reservation time, e.g. 19:00, 7:00 PM, or 7pm.",
                    },
                    "guest_name": {
                        "type": "string",
                        "description": "Name for the booking if provided.",
                    },
                },
                "required": ["party_size", "date", "time"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "check_table_availability",
            "description": (
                "Check available table times for a date and party size. "
                "Always call this before offering specific times — never guess slots."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "date": {
                        "type": "string",
                        "description": "Reservation date in YYYY-MM-DD format.",
                    },
                    "party_size": {
                        "type": "integer",
                        "description": "Number of guests, 1–20.",
                    },
                },
                "required": ["date", "party_size"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "book_table_reservation",
            "description": (
                "Submit a table reservation after the guest confirms name, phone, and slot. "
                "Use the exact slot ISO string returned by check_table_availability. "
                "Returns a 6-character confirmation_code to share with the guest."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "slot": {
                        "type": "string",
                        "description": "ISO-8601 datetime from check_table_availability, e.g. 2026-07-03T19:00:00+10:00",
                    },
                    "guest_name": {"type": "string"},
                    "guest_phone": {"type": "string"},
                    "party_size": {"type": "integer"},
                    "guest_email": {
                        "type": "string",
                        "description": "Optional email address.",
                    },
                    "notes": {
                        "type": "string",
                        "description": "Optional special requests.",
                    },
                },
                "required": ["slot", "guest_name", "guest_phone", "party_size"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "transfer_to_human",
            "description": "Offer to connect the guest with staff if they ask for a person or manager.",
            "parameters": {"type": "object", "properties": {}, "required": []},
        },
    },
    {
        "type": "function",
        "function": {
            "name": "end_call",
            "description": "End the session gracefully after saying goodbye.",
            "parameters": {"type": "object", "properties": {}, "required": []},
        },
    },
]
