from __future__ import annotations

import json
import os
from typing import Any

from google import genai
from google.genai import types

from app.prompts import SYSTEM_PROMPT, build_user_prompt
from app.schema import ElectricalAnalysis


def _client() -> genai.Client:
    api_key = (os.getenv("GEMINI_API_KEY") or "").strip()
    if not api_key or api_key.startswith("your_"):
        raise RuntimeError(
            "GEMINI_API_KEY missing. Copy .env.example to .env and set your key."
        )
    return genai.Client(api_key=api_key)


def _model_name() -> str:
    return (os.getenv("GEMINI_MODEL") or "gemini-3.1-flash-lite").strip()


def _response_schema() -> dict[str, Any]:
    # JSON Schema derived from Pydantic for Gemini structured output.
    return ElectricalAnalysis.model_json_schema()


def analyze_image(image_bytes: bytes, mime_type: str = "image/jpeg") -> ElectricalAnalysis:
    """Run Gemini vision and validate into ElectricalAnalysis."""
    client = _client()
    model = _model_name()

    contents = [
        types.Content(
            role="user",
            parts=[
                types.Part.from_text(text=build_user_prompt()),
                types.Part.from_bytes(data=image_bytes, mime_type=mime_type),
            ],
        )
    ]

    config = types.GenerateContentConfig(
        system_instruction=SYSTEM_PROMPT,
        temperature=0.1,
        response_mime_type="application/json",
        response_schema=ElectricalAnalysis,
    )

    try:
        response = client.models.generate_content(
            model=model,
            contents=contents,
            config=config,
        )
    except Exception as first_err:
        # Fallback without response_schema if SDK/model rejects schema binding.
        config_fallback = types.GenerateContentConfig(
            system_instruction=SYSTEM_PROMPT,
            temperature=0.1,
            response_mime_type="application/json",
        )
        try:
            response = client.models.generate_content(
                model=model,
                contents=contents,
                config=config_fallback,
            )
        except Exception as second_err:
            raise RuntimeError(
                f"Gemini request failed: {first_err}; fallback: {second_err}"
            ) from second_err

    text = (response.text or "").strip()
    if not text:
        raise RuntimeError("Gemini returned empty response")

    # Strip accidental markdown fences
    if text.startswith("```"):
        text = text.strip("`")
        if text.lower().startswith("json"):
            text = text[4:].lstrip()

    try:
        data = json.loads(text)
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"Invalid JSON from Gemini: {exc}\n{text[:500]}") from exc

    return ElectricalAnalysis.model_validate(data)
