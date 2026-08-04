"""
dial.py — Trigger an outbound Andre call.

Usage:
    python dial.py +919876543210
    python dial.py +919876543210 --language hi
    python dial.py +919876543210 --campaign demo
"""

import os
import sys

import requests
from dotenv import load_dotenv

load_dotenv()


def dial(number: str, campaign_id: str = "default", language: str = "auto"):
    # Hit the local agent API for dialing; Twilio Media Stream still uses PUBLIC_BASE_URL.
    port = os.environ.get("PORT", "8001")
    base_url = (
        os.environ.get("ANDRE_AGENT_URL")
        or f"http://127.0.0.1:{port}"
    ).rstrip("/")
    public = (os.environ.get("PUBLIC_BASE_URL") or "").rstrip("/")
    if not public.startswith("https://"):
        print("WARNING: PUBLIC_BASE_URL should be HTTPS (ngrok) for Twilio to stream audio.")

    url = f"{base_url}/call"
    print(f"Dialling {number} via {url} (campaign={campaign_id}, language={language}) ...")
    if public:
        print(f"Twilio stream host will be: {public}")

    headers = {"ngrok-skip-browser-warning": "true"}
    secret = (os.environ.get("CALL_API_SECRET") or "").strip()
    if secret:
        headers["X-Call-Api-Key"] = secret

    resp = requests.post(
        url,
        json={
            "to": number,
            "campaign_id": campaign_id,
            "language": language,
            "agent": "andre",
            "skip_compliance": True,
        },
        headers=headers,
        timeout=30,
    )

    print(f"Status Code: {resp.status_code}")
    try:
        data = resp.json()
        print("Response:", data)
    except Exception:
        print("Failed to parse JSON. Raw response:")
        print(resp.text)
        sys.exit(1)

    status = data.get("status")
    if status == "calling":
        print(f"Call initiated! SID: {data.get('call_sid')}")
    elif status == "blocked":
        print(f"Call blocked: {data.get('reason')}")
    elif status == "queued":
        print(f"Call queued: {data.get('reason')} — {data.get('message', '')}")
    else:
        print(f"Unexpected response: {data}")


if __name__ == "__main__":
    args = sys.argv[1:]
    if not args:
        print("Usage: python dial.py +<countrycode><number> [--language en|hi|te|auto] [--campaign <id>]")
        sys.exit(1)

    number = args[0]
    campaign = "default"
    language = "auto"
    if "--campaign" in args:
        campaign = args[args.index("--campaign") + 1]
    if "--language" in args:
        language = args[args.index("--language") + 1]

    dial(number, campaign, language)
