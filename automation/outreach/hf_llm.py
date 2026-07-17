"""Hugging Face Inference (OpenAI-compatible router) for scraping + menu OCR."""

from __future__ import annotations

import os

try:
    from openai import OpenAI
except ImportError:
    OpenAI = None  # type: ignore

HF_ROUTER_BASE = "https://router.huggingface.co/v1"


def hf_api_key() -> str:
    return (
        os.getenv("HUGGING_FACE_API_KEY", "")
        or os.getenv("HF_TOKEN", "")
        or os.getenv("HUGGINGFACE_API_KEY", "")
        or (
            os.getenv("LLM_API_KEY", "")
            if os.getenv("LLM_PROVIDER", "").lower() in ("huggingface", "hf")
            else ""
        )
    )


def hf_enabled() -> bool:
    return bool(hf_api_key())


def hf_text_model() -> str:
    return (
        os.getenv("HF_TEXT_MODEL", "").strip()
        or os.getenv("LLM_MODEL", "").strip()
        or "Qwen/Qwen2.5-7B-Instruct"
    )


def hf_vision_model() -> str:
    return (
        os.getenv("HF_VISION_MODEL", "").strip()
        or os.getenv("MENU_OCR_MODEL", "").strip()
        or "google/gemma-3-4b-it:deepinfra"
    )


def create_hf_client() -> "OpenAI":
    if OpenAI is None:
        raise ImportError("openai SDK not installed. Run: pip install openai")
    key = hf_api_key()
    if not key:
        raise ValueError("HUGGING_FACE_API_KEY not set (check backend/.env)")
    return OpenAI(base_url=HF_ROUTER_BASE, api_key=key)


def chat_completion(
    *,
    messages: list[dict],
    model: str | None = None,
    max_tokens: int = 4096,
    temperature: float = 0.7,
    json_mode: bool = False,
):
    client = create_hf_client()
    kwargs: dict = {
        "model": model or hf_text_model(),
        "messages": messages,
        "max_tokens": max_tokens,
        "temperature": temperature,
    }
    if json_mode:
        kwargs["response_format"] = {"type": "json_object"}
    return client.chat.completions.create(**kwargs)
