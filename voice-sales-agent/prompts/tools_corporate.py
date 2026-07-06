"""Tuvi corporate website assistant tools."""

CORPORATE_TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "check_consultation_slot",
            "description": (
                "Fast check if ONE specific date and time is free. "
                "Use right after the visitor gives a preferred date and time — before asking name/email."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "date": {
                        "type": "string",
                        "description": "Date YYYY-MM-DD.",
                    },
                    "time": {
                        "type": "string",
                        "description": "Time e.g. 14:00 or 2:00 PM.",
                    },
                },
                "required": ["date", "time"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "check_consultation_slots",
            "description": (
                "List open slots over the next few days. "
                "Use ONLY when the visitor asks what times are available in general."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "date": {
                        "type": "string",
                        "description": "Optional start date YYYY-MM-DD.",
                    },
                    "days": {
                        "type": "integer",
                        "description": "Business days to include (default 2).",
                    },
                },
                "required": [],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "book_consultation",
            "description": (
                "Book the consultation after slot is confirmed and you have name + email. "
                "Sends confirmation email to the visitor."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "date": {
                        "type": "string",
                        "description": "Consultation date YYYY-MM-DD.",
                    },
                    "time": {
                        "type": "string",
                        "description": "Time e.g. 14:00, 2:00 PM.",
                    },
                    "prospect_name": {
                        "type": "string",
                        "description": "Visitor's full name.",
                    },
                    "prospect_email": {
                        "type": "string",
                        "description": "Visitor's email for confirmation (required).",
                    },
                    "prospect_phone": {
                        "type": "string",
                        "description": "Visitor's phone number for the consultation request.",
                    },
                },
                "required": ["date", "time", "prospect_name", "prospect_email", "prospect_phone"],
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
