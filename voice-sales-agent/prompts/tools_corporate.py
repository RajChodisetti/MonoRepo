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
                "Book the consultation after the slot is confirmed and the booking form has "
                "returned name, email, and phone. "
                "May send a confirmation email when delivery is enabled."
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
            "name": "request_booking_details",
            "description": (
                "Open the required on-screen booking form for the visitor to type their name, "
                "email, and phone number. Browser sessions only. Call this after the visitor "
                "confirms an available consultation slot, then wait for the returned details."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "prompt": {
                        "type": "string",
                        "description": "Short message shown above the booking form.",
                    },
                },
                "required": [],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "request_typed_email",
            "description": (
                "Open an on-screen email input in the browser so the visitor can TYPE their email "
                "(never ask them to speak the email). Browser sessions only. "
                "Call this after name/phone (if collected) when you need an email to book. "
                "Wait for the tool result — it returns the typed email."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "prompt": {
                        "type": "string",
                        "description": "Short message shown above the email field.",
                    },
                },
                "required": [],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "place_callback_call",
            "description": (
                "Dial the visitor's phone so the Tuvi AI phone assistant calls them. "
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

# Phone sessions: no browser UI tools / no further outbound dials.
_PHONE_EXCLUDED = {"place_callback_call", "request_typed_email", "request_booking_details"}
CORPORATE_PHONE_TOOLS = [
    t
    for t in CORPORATE_TOOLS
    if t.get("function", {}).get("name") not in _PHONE_EXCLUDED
]
