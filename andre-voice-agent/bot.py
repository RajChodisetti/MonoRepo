"""
Andre Voice Agent — Pipecat + Twilio + Sarvam STT/TTS + OpenAI LLM

Pipeline:
  Audio → SarvamSTT → OpenAI LLM → SarvamTTS → phone/browser
"""

from __future__ import annotations

import asyncio
import json
import logging
import os
import time
import uuid
from contextlib import asynccontextmanager
from datetime import datetime
from pathlib import Path
from urllib.parse import parse_qsl
from zoneinfo import ZoneInfo

import requests as _requests
import uvicorn
from dotenv import load_dotenv
from fastapi import Depends, FastAPI, HTTPException, Request, WebSocket
from fastapi.responses import FileResponse, JSONResponse, PlainTextResponse
from fastapi.staticfiles import StaticFiles
from twilio.request_validator import RequestValidator
from twilio.rest import Client as TwilioClient
import phonenumbers

from pipecat.pipeline.pipeline import Pipeline
from pipecat.pipeline.runner import PipelineRunner
from pipecat.pipeline.task import PipelineParams, PipelineTask
from pipecat.frames.frames import (
    BotStartedSpeakingFrame,
    BotStoppedSpeakingFrame,
    EndFrame,
    LLMFullResponseEndFrame,
    LLMFullResponseStartFrame,
    LLMMessagesFrame,
    FunctionCallResultProperties,
    STTUpdateSettingsFrame,
    TextFrame,
    TranscriptionFrame,
    TTSSpeakFrame,
    TTSUpdateSettingsFrame,
    UserStartedSpeakingFrame,
    UserStoppedSpeakingFrame,
)
from pipecat.transports.websocket.fastapi import FastAPIWebsocketParams, FastAPIWebsocketTransport
from pipecat.serializers.twilio import TwilioFrameSerializer
from pipecat.services.sarvam.stt import SarvamSTTService
from pipecat.services.sarvam.tts import SarvamTTSService
from pipecat.services.openai.llm import OpenAILLMService
from pipecat.processors.aggregators.llm_context import LLMContext
from pipecat.processors.aggregators.llm_response_universal import LLMContextAggregatorPair
from pipecat.audio.vad.silero import SileroVADAnalyzer
from pipecat.audio.vad.vad_analyzer import VADParams
from pipecat.processors.frame_processor import FrameDirection, FrameProcessor
from pipecat.transcriptions.language import Language

from browser_serializer import BrowserFrameSerializer
from caller_id import resolve_caller_id
from logger_db import (
    end_call,
    get_call_number,
    get_call_summary,
    init_db,
    is_opted_out,
    list_calls,
    list_conversation_transcripts,
    log_turn,
    record_opt_out,
    start_call,
    update_call_contact,
)
from prompts.andre import build_andre_greeting, build_andre_prompt
from prompts.tools_andre import (
    ANDRE_BROWSER_TOOLS,
    ANDRE_PHONE_TOOLS,
    andre_tools_schema,
)
import listings_api
import property_store

load_dotenv()

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s — %(message)s",
)
logger = logging.getLogger("AndreVoiceAgent")

init_db()

IST_TZ = ZoneInfo("Asia/Kolkata")

OPT_OUT_PHRASES = [
    "stop calling",
    "don't call me",
    "do not call",
    "remove me",
    "take me off",
    "opt me out",
    "never call again",
    "never contact me",
    "mujhe call mat karo",
    "call mat karo",
    "nannu call cheyyaku",
]

_active_tasks: dict[str, PipelineTask] = {}
_call_db_ids: dict[str, int] = {}
_call_languages: dict[str, str] = {}
_paused_campaigns: set[str] = set()


# ── Language helpers ──────────────────────────────────────────────────────────

def _normalize_language(raw: str | None, default: str | None = None) -> str:
    fallback = (default or os.environ.get("DEFAULT_LANGUAGE", "auto") or "auto").strip().lower()
    lang = (raw or fallback).strip().lower().replace("_", "-")
    aliases = {
        "auto": "auto",
        "en": "en",
        "en-in": "en",
        "en-us": "en",
        "english": "en",
        "hi": "hi",
        "hi-in": "hi",
        "hindi": "hi",
        "te": "te",
        "te-in": "te",
        "telugu": "te",
    }
    return aliases.get(lang, fallback if fallback in ("auto", "en", "hi", "te") else "auto")


def _pipecat_language(lang: str) -> Language | None:
    """Map session language to Pipecat Language for Sarvam. None = auto/unspecified."""
    lang = _normalize_language(lang)
    if lang == "hi":
        return getattr(Language, "HI_IN", None) or Language.HI
    if lang == "te":
        return getattr(Language, "TE_IN", None) or Language.TE
    if lang == "en":
        return getattr(Language, "EN_IN", None) or Language.EN
    return None


def _sarvam_api_language(lang: str) -> str:
    lang = _normalize_language(lang)
    return {"hi": "hi-IN", "te": "te-IN", "en": "en-IN"}.get(lang, "en-IN")


# Best Sarvam bulbul:v3 female voices per language (override via SARVAM_TTS_VOICE_EN/HI/TE).
_DEFAULT_FEMALE_VOICES = {
    "en": "simran",  # English conversational female
    "hi": "priya",  # warm natural Hindi female (less robotic than kavya)
    "te": "kavitha",  # Telugu-friendly female
    "auto": "priya",  # warm default female when auto
}


def _tts_voice_for_language(language: str) -> str:
    """Girl TTS voice for the session language. Per-lang env wins; else female default."""
    lang = _normalize_language(language)
    env_key = {
        "en": "SARVAM_TTS_VOICE_EN",
        "hi": "SARVAM_TTS_VOICE_HI",
        "te": "SARVAM_TTS_VOICE_TE",
    }.get(lang)
    if env_key:
        specific = (os.environ.get(env_key) or "").strip()
        if specific:
            return specific
        return _DEFAULT_FEMALE_VOICES[lang]
    # auto: optional global override, else female default
    return (os.environ.get("SARVAM_TTS_VOICE") or "").strip() or _DEFAULT_FEMALE_VOICES["auto"]


def _build_stt(language: str) -> SarvamSTTService:
    model = os.environ.get("SARVAM_STT_MODEL", "saaras:v3")
    settings_kwargs: dict = {"model": model}
    pipecat_lang = _pipecat_language(language)
    if pipecat_lang is not None:
        settings_kwargs["language"] = pipecat_lang
    return SarvamSTTService(
        api_key=os.environ["SARVAM_API_KEY"],
        settings=SarvamSTTService.Settings(**settings_kwargs),
    )


def _tts_expressiveness() -> tuple[float, float]:
    """Conversational pace + moderate temperature for clear, natural speech."""
    try:
        pace = float(os.environ.get("SARVAM_TTS_PACE", "1.15"))
    except ValueError:
        pace = 1.15
    try:
        temperature = float(os.environ.get("SARVAM_TTS_TEMPERATURE", "0.6"))
    except ValueError:
        temperature = 0.6
    return max(0.5, min(2.0, pace)), max(0.01, min(1.0, temperature))


def _build_tts(language: str, *, sample_rate: int | None = None) -> SarvamTTSService:
    model = os.environ.get("SARVAM_TTS_MODEL", "bulbul:v3")
    voice = _tts_voice_for_language(language)
    pace, temperature = _tts_expressiveness()
    settings_kwargs: dict = {
        "model": model,
        "voice": voice,
        "enable_preprocessing": True,
        "pace": pace,
        "temperature": temperature,
    }
    pipecat_lang = _pipecat_language(language)
    if pipecat_lang is not None:
        settings_kwargs["language"] = pipecat_lang
    else:
        # auto: default to Indian English TTS; LLM mirrors caller language
        settings_kwargs["language"] = getattr(Language, "EN_IN", None) or Language.EN

    kwargs: dict = {
        "api_key": os.environ["SARVAM_API_KEY"],
        "settings": SarvamTTSService.Settings(**settings_kwargs),
    }
    if sample_rate:
        kwargs["sample_rate"] = sample_rate
    logger.info(
        "TTS voice=%s language=%s model=%s pace=%s temp=%s",
        voice,
        language,
        model,
        pace,
        temperature,
    )
    return SarvamTTSService(**kwargs)


async def _fetch_sarvam_tts_bytes(transcript: str, language: str = "auto") -> bytes | None:
    """Synthesize PCM16 @ 16kHz via Sarvam HTTP (browser greeting)."""
    api_key = os.environ.get("SARVAM_API_KEY", "")
    if not (api_key and transcript.strip()):
        return None

    model = os.environ.get("SARVAM_TTS_MODEL", "bulbul:v3")
    effective_lang = language if language != "auto" else "en"
    voice = _tts_voice_for_language(effective_lang)
    target_lang = _sarvam_api_language(effective_lang)
    pace, temperature = _tts_expressiveness()

    def _fetch() -> bytes:
        import base64

        resp = _requests.post(
            "https://api.sarvam.ai/text-to-speech",
            headers={
                "api-subscription-key": api_key,
                "Content-Type": "application/json",
            },
            json={
                "text": transcript,
                "target_language_code": target_lang,
                "speaker": voice,
                "model": model,
                "speech_sample_rate": 16000,
                "enable_preprocessing": True,
                "output_audio_codec": "linear16",
                "pace": pace,
                "temperature": temperature,
            },
            timeout=30,
        )
        resp.raise_for_status()
        data = resp.json()
        # Sarvam returns base64 audio in audios[]
        audios = data.get("audios") or data.get("audio") or []
        if isinstance(audios, str):
            return base64.b64decode(audios)
        if audios:
            return base64.b64decode(audios[0])
        raise RuntimeError(f"Unexpected Sarvam TTS response keys: {list(data.keys())}")

    last_exc: Exception | None = None
    for attempt in range(1, 4):
        try:
            return await asyncio.to_thread(_fetch)
        except Exception as exc:
            last_exc = exc
            logger.warning("Sarvam TTS fetch failed (attempt %s/3): %s", attempt, exc)
            if attempt < 3:
                await asyncio.sleep(0.4 * attempt)
    logger.warning("Sarvam TTS fetch gave up: %s", last_exc)
    return None


async def _stream_browser_greeting(
    websocket: WebSocket,
    *,
    call_db_id: int,
    context: LLMContext,
    greeting_text: str,
    language: str,
) -> bool:
    audio = await _fetch_sarvam_tts_bytes(greeting_text, language)
    if not audio:
        return False
    try:
        log_turn(call_db_id, "assistant", greeting_text)
        context.messages.append({"role": "assistant", "content": greeting_text})
        await websocket.send_text(json.dumps({"event": "status", "state": "bot_speaking"}))
        chunk_size = 3200
        for i in range(0, len(audio), chunk_size):
            await websocket.send_bytes(audio[i : i + chunk_size])
        await websocket.send_text(
            json.dumps({"event": "transcript", "role": "assistant", "text": greeting_text})
        )
        await websocket.send_text(json.dumps({"event": "status", "state": "listening"}))
        return True
    except Exception as exc:
        logger.warning("Browser greeting stream failed: %s", exc)
        return False


# ── Helpers ───────────────────────────────────────────────────────────────────

def normalize_e164(number: str, default_region: str = "IN") -> str | None:
    try:
        parsed = phonenumbers.parse(number, default_region)
        if phonenumbers.is_valid_number(parsed):
            return phonenumbers.format_number(parsed, phonenumbers.PhoneNumberFormat.E164)
    except Exception:
        pass
    return None


def _is_calling_allowed() -> tuple[bool, str]:
    """Simple India weekday window 09:00–21:00 IST (relaxed for testing)."""
    now = datetime.now(IST_TZ)
    weekday = now.weekday()
    hour = now.hour + now.minute / 60.0
    if weekday == 6:
        return False, "sunday"
    return (True, "ok") if 9.0 <= hour < 21.0 else (False, "outside_hours")


def _append_spoken_chunk(buf: str, chunk: str) -> str:
    """Join streamed LLM tokens into readable sentences."""
    if not chunk:
        return buf
    if not buf:
        return chunk
    if chunk[:1].isspace() or buf[-1:].isspace():
        return buf + chunk
    if chunk[:1] in ",.!?;:)]}'\"":
        return buf + chunk
    return f"{buf} {chunk}"


class TranscriptLogger(FrameProcessor):
    def __init__(
        self,
        call_id: int,
        role_filter: str,
        timing_state: dict | None = None,
        event_ws: WebSocket | None = None,
    ):
        super().__init__()
        self._call_id = call_id
        self._role = role_filter
        self._timing = timing_state
        self._event_ws = event_ws
        self._assistant_buf = ""

    async def _send_event(self, data: dict):
        if self._event_ws is None:
            return
        try:
            await self._event_ws.send_text(json.dumps(data))
        except Exception:
            pass

    async def _flush_assistant(self):
        text = " ".join(self._assistant_buf.split()).strip()
        self._assistant_buf = ""
        if not text:
            return
        log_turn(self._call_id, "assistant", text)
        logger.info(f"[ASSISTANT] {text}")
        await self._send_event({"event": "transcript", "role": "assistant", "text": text})

    async def process_frame(self, frame: object, direction: FrameDirection):
        await super().process_frame(frame, direction)
        if self._role == "user" and isinstance(frame, TranscriptionFrame):
            text = frame.text.strip()
            if text:
                log_turn(self._call_id, "user", text)
                logger.info(f"[USER]      {text}")
                if self._timing is not None:
                    self._timing["user_turn_ts"] = time.time()
                    self._timing["first_text_pending"] = True
                await self._send_event({"event": "transcript", "role": "user", "text": text})
        elif self._role == "assistant":
            if isinstance(frame, LLMFullResponseStartFrame):
                self._assistant_buf = ""
            elif isinstance(frame, TextFrame):
                chunk = frame.text or ""
                if chunk and not self._assistant_buf:
                    if (
                        self._timing is not None
                        and self._timing.get("first_text_pending")
                        and self._timing.get("user_turn_ts")
                    ):
                        latency_ms = round((time.time() - self._timing["user_turn_ts"]) * 1000)
                        logger.info(f"[LATENCY] turn_end→first_text: {latency_ms}ms")
                        self._timing["first_text_pending"] = False
                self._assistant_buf = _append_spoken_chunk(self._assistant_buf, chunk)
            elif isinstance(frame, (LLMFullResponseEndFrame, EndFrame)):
                await self._flush_assistant()
        await self.push_frame(frame, direction)


class StatusEventSender(FrameProcessor):
    def __init__(self, ws: WebSocket):
        super().__init__()
        self._ws = ws

    async def _send(self, data: dict):
        try:
            await self._ws.send_text(json.dumps(data))
        except Exception:
            pass

    async def process_frame(self, frame: object, direction: FrameDirection):
        await super().process_frame(frame, direction)
        if isinstance(frame, UserStartedSpeakingFrame):
            await self._send({"event": "status", "state": "user_speaking"})
        elif isinstance(frame, UserStoppedSpeakingFrame):
            await self._send({"event": "status", "state": "thinking"})
        elif isinstance(frame, BotStartedSpeakingFrame):
            await self._send({"event": "status", "state": "bot_speaking"})
        elif isinstance(frame, BotStoppedSpeakingFrame):
            await self._send({"event": "status", "state": "listening"})
        await self.push_frame(frame, direction)


class OptOutGuardrail(FrameProcessor):
    def __init__(
        self,
        call_db_id: int,
        to_number_ref: list[str],
        outcome_ref: list[str],
        task_ref: list,
    ):
        super().__init__()
        self._call_db_id = call_db_id
        self._to_number_ref = to_number_ref
        self._outcome_ref = outcome_ref
        self._task_ref = task_ref
        self._triggered = False

    async def process_frame(self, frame: object, direction: FrameDirection):
        await super().process_frame(frame, direction)
        if not self._triggered and isinstance(frame, TranscriptionFrame):
            text = frame.text.lower()
            if any(phrase in text for phrase in OPT_OUT_PHRASES):
                self._triggered = True
                self._outcome_ref[0] = "opted_out"
                phone = self._to_number_ref[0]
                record_opt_out(phone, self._call_db_id)
                logger.warning(f"Opt-out detected: '{frame.text}' — recorded for {phone}")
                task = self._task_ref[0]
                if task:
                    asyncio.create_task(_delayed_end(task, delay=12.0))
        await self.push_frame(frame, direction)


class ContextCompactor(FrameProcessor):
    def __init__(self, context: LLMContext, max_pairs: int = 5):
        super().__init__()
        self._context = context
        self._max_pairs = max_pairs

    async def process_frame(self, frame: object, direction: FrameDirection):
        await super().process_frame(frame, direction)
        if isinstance(frame, TranscriptionFrame) and frame.text.strip():
            self._trim()
        await self.push_frame(frame, direction)

    def _trim(self):
        messages = self._context.messages
        if not messages:
            return

        def _is_protected(m: dict) -> bool:
            role = m.get("role")
            if role == "system":
                return True
            if role == "tool":
                return True
            if role == "assistant" and (m.get("tool_calls") or m.get("function_call")):
                return True
            return False

        other_idxs = [i for i, m in enumerate(messages) if not _is_protected(m)]
        max_keep = self._max_pairs * 2
        if len(other_idxs) <= max_keep:
            return
        drop_idxs = other_idxs[: len(other_idxs) - max_keep]
        for i in reversed(drop_idxs):
            del messages[i]
        logger.info(f"[CONTEXT] Compacted: dropped {len(drop_idxs)} msgs")


async def _require_twilio_signature(request: Request):
    env = os.environ.get("ENVIRONMENT", "development").lower()
    if env == "development":
        return
    auth_token = os.environ.get("TWILIO_AUTH_TOKEN", "")
    validator = RequestValidator(auth_token)
    public_base = os.environ.get("PUBLIC_BASE_URL", "").rstrip("/")
    path = request.url.path
    if request.url.query:
        path += f"?{request.url.query}"
    url = f"{public_base}{path}"
    signature = request.headers.get("X-Twilio-Signature", "")
    form_data = await request.form()
    params = dict(form_data)
    if not validator.validate(url, params, signature):
        raise HTTPException(status_code=403, detail="Invalid Twilio signature")


async def _require_call_api_key(request: Request):
    secret = (os.environ.get("CALL_API_SECRET") or "").strip()
    if not secret:
        env = os.environ.get("ENVIRONMENT", "development").lower()
        if env == "development":
            logger.warning("CALL_API_SECRET unset — allowing /call in development")
            return
        raise HTTPException(status_code=503, detail="CALL_API_SECRET is not configured")
    header_key = (request.headers.get("X-Call-Api-Key") or "").strip()
    auth = (request.headers.get("Authorization") or "").strip()
    bearer = auth[7:].strip() if auth.lower().startswith("bearer ") else ""
    provided = header_key or bearer
    if not provided or provided != secret:
        raise HTTPException(status_code=401, detail="Invalid or missing call API key")


def _stream_custom_params(start_payload: dict) -> dict[str, str]:
    custom = start_payload.get("customParameters") or {}
    if isinstance(custom, dict):
        return {str(k): str(v) for k, v in custom.items()}
    if isinstance(custom, list):
        out: dict[str, str] = {}
        for item in custom:
            if isinstance(item, dict) and item.get("name") is not None:
                out[str(item["name"])] = str(item.get("value", ""))
        return out
    return {}


def start_outbound_call(
    to_number: str,
    *,
    campaign_id: str = "default",
    language: str = "auto",
) -> dict:
    public_url = os.environ["PUBLIC_BASE_URL"].rstrip("/")
    ws_host = public_url.replace("https://", "").replace("http://", "")
    safe_campaign = "".join(c for c in campaign_id if c.isalnum() or c in "-_")[:64] or "default"
    lang = _normalize_language(language)

    twilio_client = TwilioClient(
        os.environ["TWILIO_ACCOUNT_SID"],
        os.environ["TWILIO_AUTH_TOKEN"],
    )
    from_e164, verified = resolve_caller_id(None)

    twiml = f"""<?xml version="1.0" encoding="UTF-8"?>
<Response>
  <Connect>
    <Stream url="wss://{ws_host}/stream">
      <Parameter name="campaign_id" value="{safe_campaign}"/>
      <Parameter name="agent" value="andre"/>
      <Parameter name="language" value="{lang}"/>
    </Stream>
  </Connect>
</Response>"""

    call = twilio_client.calls.create(
        to=to_number,
        from_=from_e164,
        twiml=twiml,
        status_callback=f"{public_url}/twilio/status",
        status_callback_event=["initiated", "ringing", "answered", "completed"],
        status_callback_method="POST",
    )

    call_db_id = start_call(call.sid, to_number, channel="phone", agent_mode="andre")
    _call_db_ids[call.sid] = call_db_id
    _call_languages[call.sid] = lang

    logger.info(
        f"Outbound call → {to_number}  SID={call.sid}  lang={lang}  from={from_e164}  DB id={call_db_id}"
    )
    return {
        "status": "calling",
        "to": to_number,
        "call_sid": call.sid,
        "log_id": call_db_id,
        "agent": "andre",
        "language": lang,
        "from": from_e164,
        "from_verified": verified,
    }


async def _prewarm_services():
    try:
        _warmup = SileroVADAnalyzer(params=VADParams(stop_secs=0.4))
        del _warmup
        logger.info("[PREWARM] Silero VAD model loaded")
    except Exception as e:
        logger.warning(f"[PREWARM] Silero pre-warm failed: {e}")


@asynccontextmanager
async def lifespan(app: FastAPI):
    asyncio.create_task(_prewarm_services())
    yield


# Tools that hit the mock listings API — answer is spoken immediately after they return.
_LOOKUP_TOOLS = frozenset(
    {
        "search_properties",
        "get_property_details",
        "get_property_options",
        "create_property",
        "update_property",
        "delete_property",
    }
)


def _format_match_line(match: dict[str, Any], *, include_id: bool = False) -> str:
    bits = [
        match.get("locality") or match.get("city"),
        match.get("size"),
        match.get("price"),
    ]
    line = ", ".join(str(b) for b in bits if b)
    if include_id and match.get("id"):
        line = f"{line} (id {match['id']})" if line else f"id {match['id']}"
    return line or (match.get("title") or match.get("id") or "a listing")


def _speech_for_tool_result(function_name: str, result: dict[str, Any]) -> str | None:
    """Build an instant spoken answer so the user never waits for a second LLM turn."""
    if not isinstance(result, dict):
        return None

    if function_name == "search_properties":
        matches = result.get("results") or result.get("matches") or []
        count = int(result.get("count") if result.get("count") is not None else len(matches))
        widened = bool(result.get("widened"))
        relaxed = result.get("relaxed_filters") or []
        if not matches:
            type_locs = result.get("type_localities") or []
            if type_locs:
                return (
                    f"I couldn't find an exact match there. "
                    f"For that type, available areas are {', '.join(type_locs[:5])}. "
                    "Want me to search one of those?"
                )
            cities = result.get("available_cities") or []
            city_bit = f" Available cities are {', '.join(cities)}." if cities else ""
            return f"I couldn't find matching listings.{city_bit} Want to widen city, budget, or type?"
        parts = []
        for i, match in enumerate(matches[:2], 1):
            parts.append(f"{i}. {_format_match_line(match)}.")
        more = f" Showing {len(parts)} of {count}." if count > len(parts) else ""
        preface = "I found "
        if widened:
            if "locality" in relaxed and "status" in relaxed:
                preface = "Nothing exact in that area/status, but nearby I found "
            elif "locality" in relaxed:
                preface = "Nothing exact in that area, but nearby I found "
            elif "status" in relaxed:
                preface = "No rentals matched, but for sale I found "
            else:
                preface = "I widened the search and found "
        return f"{preface}{count} option{'s' if count != 1 else ''}.{more} {' '.join(parts)} Want more details on any?"

    if function_name == "get_property_details":
        if result.get("status") == "error" or not result.get("ok", True):
            return result.get("message") or "I couldn't find that listing."
        summary = result.get("summary") or {}
        prop = result.get("property") or {}
        locality = summary.get("locality") or (prop.get("location") or {}).get("locality")
        price = summary.get("price") or (prop.get("pricing") or {}).get("price_display")
        size = summary.get("size")
        title = summary.get("title") or prop.get("title")
        bits = [title, locality, size, price]
        line = ", ".join(str(b) for b in bits if b)
        return f"Here are the details. {line}." if line else "I have the listing details now."

    if function_name == "get_property_options":
        # Options meta is not a search — keep it short; real finds use search_properties.
        if result.get("city") and result.get("localities"):
            locs = ", ".join((result.get("localities") or [])[:5])
            return (
                f"Areas in {result['city']} include {locs}. "
                "Tell me apartment, villa, or plot and I'll search listings."
            )
        cities = result.get("cities") or []
        if isinstance(cities, list) and cities:
            names = []
            for c in cities[:5]:
                if isinstance(c, dict):
                    names.append(c.get("name") or "")
                else:
                    names.append(str(c))
            names = [n for n in names if n]
            if names:
                return f"I cover {', '.join(names)}. Which city and type should I search?"
        return None

    if function_name in {"create_property", "update_property", "delete_property"}:
        if result.get("status") == "error" or result.get("ok") is False:
            return result.get("message") or "Sorry, that inventory change didn't go through."
        if function_name == "delete_property":
            deleted = result.get("deleted")
            return result.get("message") or (f"Done — I deleted listing {deleted}." if deleted else "Done, listing deleted.")
        summary = result.get("summary") or {}
        pid = summary.get("id") or (result.get("property") or {}).get("id")
        line = _format_match_line(summary, include_id=True) if summary else ""
        action = "added" if function_name == "create_property" else "updated"
        if line:
            return f"Done — I {action} {line}."
        if pid:
            return f"Done — I {action} listing {pid}."
        return result.get("message") or f"Done — listing {action}."

    return None


async def _speak_tool_answer(
    task: PipelineTask | None,
    *,
    call_db_id: int,
    function_name: str,
    result: dict[str, Any],
    event_ws: WebSocket | None = None,
) -> str | None:
    """Speak listings/CRUD outcomes immediately (do not wait for another LLM turn)."""
    speech = _speech_for_tool_result(function_name, result)
    if not speech or task is None:
        return None
    try:
        await task.queue_frames([TTSSpeakFrame(text=speech, append_to_context=True)])
        log_turn(call_db_id, "assistant", speech)
        logger.info("[ASSISTANT] (tool-speak %s) %s", function_name, speech)
        if event_ws is not None:
            try:
                await event_ws.send_text(
                    json.dumps({"event": "transcript", "role": "assistant", "text": speech})
                )
            except Exception:
                pass
        return speech
    except Exception as exc:
        logger.warning("Instant tool answer TTS failed: %s", exc)
        return None


async def _apply_language_to_pipeline(task: PipelineTask | None, language: str) -> None:
    """Push STT/TTS language + girl voice updates for the rest of the session."""
    if task is None:
        return
    pipecat_lang = _pipecat_language(language)
    tts_lang = pipecat_lang or getattr(Language, "EN_IN", None) or Language.EN
    voice = _tts_voice_for_language(language)
    frames = [
        TTSUpdateSettingsFrame(
            delta=SarvamTTSService.Settings(language=tts_lang, voice=voice)
        ),
    ]
    if pipecat_lang is not None:
        frames.insert(
            0,
            STTUpdateSettingsFrame(delta=SarvamSTTService.Settings(language=pipecat_lang)),
        )
    try:
        await task.queue_frames(frames)
        logger.info("Applied language=%s voice=%s to pipeline", language, voice)
    except Exception as exc:
        logger.warning("Failed to apply language update frames: %s", exc)


async def _dispatch_tool(
    *,
    function_name: str,
    arguments: dict,
    call_db_id: int,
    to_number_ref: list[str],
    outcome_state: list[str],
    task_ref: list,
    language_state: list[str],
    call_sid: str = "unknown",
    is_browser: bool = False,
) -> dict:
    logger.info(f"Tool: {function_name}  args={arguments}  browser={is_browser}")

    if function_name == "get_property_options":
        # Dummy listings API ← mock JSON inventory
        result = listings_api.api_meta(city=(arguments.get("city") or None))
        result["status"] = "ok" if result.get("ok") else "error"
        return result

    if function_name == "search_properties":
        bhk_raw = arguments.get("bhk")
        bhk = None
        if bhk_raw is not None and str(bhk_raw).strip() != "":
            try:
                bhk = int(bhk_raw)
            except (TypeError, ValueError):
                bhk = None
        limit = 3
        if arguments.get("limit") is not None:
            try:
                limit = int(arguments.get("limit"))
            except (TypeError, ValueError):
                limit = 3
        # Dummy listings API ← mock JSON inventory
        result = listings_api.api_search(
            city=(arguments.get("city") or None),
            locality=(arguments.get("locality") or None),
            property_type=(arguments.get("property_type") or None),
            bhk=bhk,
            min_price=arguments.get("min_price"),
            max_price=arguments.get("max_price"),
            status=(arguments.get("status") or None),
            budget_band=(arguments.get("budget_band") or None),
            limit=limit,
        )
        # Keep both keys so prompts that mention "matches" still work
        result["status"] = "ok" if result.get("ok") else "error"
        result["matches"] = result.get("results") or []
        return result

    if function_name == "get_property_details":
        pid = (arguments.get("property_id") or "").strip()
        result = listings_api.api_details(pid)
        if not result.get("ok"):
            return {
                "status": "error",
                "message": result.get("message")
                or f"Property '{pid}' not found. Use an id from search_properties.",
            }
        return {
            "status": "ok",
            "source": "mock",
            "property": result["property"],
            "summary": result["summary"],
        }

    if function_name == "create_property":
        payload = {
            "title": arguments.get("title"),
            "type": arguments.get("property_type") or arguments.get("type"),
            "status": arguments.get("status") or "sale",
            "city": arguments.get("city"),
            "locality": arguments.get("locality"),
            "state": arguments.get("state"),
            "pincode": arguments.get("pincode"),
            "bhk": arguments.get("bhk"),
            "size_sqft": arguments.get("size_sqft"),
            "size_sqyd": arguments.get("size_sqyd"),
            "price_inr": arguments.get("price_inr"),
            "description": arguments.get("description"),
            "amenities": arguments.get("amenities"),
            "nearby": arguments.get("nearby"),
            "furnishing": arguments.get("furnishing"),
            "possession": arguments.get("possession"),
            "negotiable": arguments.get("negotiable", True),
        }
        result = listings_api.api_create(payload)
        if not result.get("ok"):
            return {"status": "error", "message": result.get("message") or "Could not create listing."}
        return {
            "status": "ok",
            "property": result["property"],
            "summary": result["summary"],
            "message": result.get("message"),
        }

    if function_name == "update_property":
        pid = (arguments.get("property_id") or "").strip()
        if not pid:
            return {"status": "error", "message": "property_id is required to update a listing."}
        payload: dict = {}
        field_map = {
            "title": "title",
            "property_type": "type",
            "status": "status",
            "city": "city",
            "locality": "locality",
            "state": "state",
            "pincode": "pincode",
            "bhk": "bhk",
            "size_sqft": "size_sqft",
            "size_sqyd": "size_sqyd",
            "price_inr": "price_inr",
            "description": "description",
            "amenities": "amenities",
            "nearby": "nearby",
            "furnishing": "furnishing",
            "possession": "possession",
            "negotiable": "negotiable",
        }
        for src, dest in field_map.items():
            if src in arguments and arguments.get(src) is not None:
                payload[dest] = arguments.get(src)
        if "type" not in payload and arguments.get("type") is not None:
            payload["type"] = arguments.get("type")
        if not payload:
            return {"status": "error", "message": "No fields to update. Ask what they want to change."}
        result = listings_api.api_update(pid, payload)
        if not result.get("ok"):
            return {"status": "error", "message": result.get("message") or "Could not update listing."}
        return {
            "status": "ok",
            "property": result["property"],
            "summary": result["summary"],
            "message": result.get("message"),
        }

    if function_name == "delete_property":
        pid = (arguments.get("property_id") or "").strip()
        confirmed_raw = arguments.get("confirmed")
        if isinstance(confirmed_raw, str):
            confirmed = confirmed_raw.strip().lower() in {"1", "true", "yes", "y"}
        else:
            confirmed = bool(confirmed_raw)

        # Resolve listing when user pointed by place instead of id
        if not pid:
            matches = property_store.search_properties(
                city=(arguments.get("city") or None),
                locality=(arguments.get("locality") or None),
                property_type=(arguments.get("property_type") or arguments.get("type") or None),
                limit=5,
            )
            if len(matches) == 1:
                pid = matches[0]["id"]
            elif len(matches) > 1:
                return {
                    "status": "error",
                    "error": "ambiguous",
                    "message": (
                        "Multiple listings match. Ask which one to delete and use a property_id."
                    ),
                    "candidates": matches,
                }
            else:
                return {
                    "status": "error",
                    "error": "not_found",
                    "message": (
                        "Could not resolve which listing to delete. "
                        "Search first, then delete with a property_id from results."
                    ),
                }

        if not confirmed:
            summary = property_store.get_property(pid)
            label = ""
            if summary:
                s = property_store.summarize_property(summary)
                bits = " ".join(
                    str(x) for x in (s.get("locality"), s.get("type"), s.get("price")) if x
                )
                if bits:
                    label = f" ({bits})"
            return {
                "status": "error",
                "error": "confirmation_required",
                "property_id": pid,
                "message": (
                    f"Ready to delete {pid}{label}. "
                    "Ask the user to confirm, then call again with confirmed=true."
                ),
            }

        result = listings_api.api_delete(pid, confirmed=True)
        if not result.get("ok"):
            return {
                "status": "error",
                "error": result.get("error"),
                "message": result.get("message") or "Could not delete listing.",
            }
        return {"status": "ok", "deleted": result.get("deleted"), "message": result.get("message")}

    if function_name == "set_language":
        lang = _normalize_language(arguments.get("language"))
        language_state[0] = lang
        if call_sid and call_sid != "unknown":
            _call_languages[call_sid] = lang
        await _apply_language_to_pipeline(task_ref[0] if task_ref else None, lang)
        return {
            "status": "ok",
            "language": lang,
            "voice": _tts_voice_for_language(lang),
            "message": (
                f"Language locked to {lang} with matching girl voice. "
                "Reply only in that language for the rest of the session."
            ),
        }

    if function_name == "mark_do_not_call":
        phone = to_number_ref[0]
        record_opt_out(phone, call_db_id)
        outcome_state[0] = "opted_out"
        return {"status": "recorded", "message": "Number added to do-not-call list."}

    if function_name == "place_callback_call":
        if not is_browser:
            return {
                "status": "error",
                "message": "Callback dialing is only available from the browser assistant.",
            }
        raw_phone = (arguments.get("phone_number") or "").strip()
        to_number = normalize_e164(raw_phone)
        if not to_number:
            return {
                "status": "error",
                "message": "Invalid phone number. Ask for a number with country code, e.g. +91…",
            }
        if is_opted_out(to_number):
            return {
                "status": "blocked",
                "reason": "internal_opt_out",
                "message": "That number is on the do-not-call list.",
            }
        skip_compliance = os.environ.get("ENVIRONMENT", "development").lower() == "development"
        allowed, reason = _is_calling_allowed()
        if not allowed and not skip_compliance:
            return {
                "status": "queued",
                "reason": reason,
                "message": f"Outside the allowed calling window ({reason}).",
            }
        lang = _normalize_language(arguments.get("language") or language_state[0])
        try:
            result = start_outbound_call(
                to_number,
                campaign_id="andre_callback",
                language=lang,
            )
            update_call_contact(
                call_db_id,
                phone=to_number,
                contact_name=(arguments.get("name") or "").strip(),
            )
            return {
                "status": result.get("status", "calling"),
                "call_sid": result.get("call_sid"),
                "to": to_number,
                "language": lang,
                "name": (arguments.get("name") or "").strip() or None,
            }
        except Exception as e:
            logger.error(f"place_callback_call failed: {e}")
            return {"status": "error", "message": "Could not place the call right now."}

    if function_name == "end_call":
        if outcome_state[0] == "unknown":
            outcome_state[0] = "completed"
        if task_ref[0]:
            asyncio.create_task(_delayed_end(task_ref[0]))
        return {"status": "ending"}

    return {"status": "error", "message": f"Unknown tool: {function_name}"}


# ── FastAPI app ───────────────────────────────────────────────────────────────

app = FastAPI(title="Andre Real Estate Voice Agent", lifespan=lifespan)

_STATIC_DIR = Path(__file__).parent / "static"
if _STATIC_DIR.is_dir():
    app.mount("/static", StaticFiles(directory=str(_STATIC_DIR)), name="static")


@app.get("/api/v1/listings/meta")
@app.get("/api/properties/meta")
async def properties_meta(city: str | None = None):
    """Dummy listings meta API (mock JSON)."""
    return listings_api.api_meta(city=city)


@app.get("/api/v1/listings/search")
@app.get("/api/properties/search")
async def properties_search(
    city: str | None = None,
    locality: str | None = None,
    property_type: str | None = None,
    bhk: int | None = None,
    min_price: float | None = None,
    max_price: float | None = None,
    status: str | None = None,
    budget_band: str | None = None,
    limit: int = 5,
):
    """Dummy listings search API — reads data/properties.json."""
    return listings_api.api_search(
        city=city,
        locality=locality,
        property_type=property_type,
        bhk=bhk,
        min_price=min_price,
        max_price=max_price,
        status=status,
        budget_band=budget_band,
        limit=limit,
    )


@app.get("/api/properties")
async def properties_list_admin(_: None = Depends(_require_call_api_key)):
    """Admin list of all mock inventory summaries."""
    items = property_store.list_properties_admin()
    return {"ok": True, "count": len(items), "results": items, **property_store.catalog_meta()}


@app.post("/api/properties")
async def properties_create(request: Request, _: None = Depends(_require_call_api_key)):
    """Create a listing in data/properties.json and reload cache."""
    try:
        payload = await request.json()
    except Exception as exc:
        raise HTTPException(status_code=400, detail="Invalid JSON body") from exc
    if not isinstance(payload, dict):
        raise HTTPException(status_code=400, detail="Body must be a JSON object")
    try:
        created = property_store.create_property(payload)
    except ValueError as exc:
        raise HTTPException(status_code=409, detail=str(exc)) from exc
    return {"ok": True, "property": created, "summary": property_store.summarize_property(created)}


@app.put("/api/properties/{property_id}")
async def properties_update(property_id: str, request: Request, _: None = Depends(_require_call_api_key)):
    """Update a listing in data/properties.json and reload cache."""
    try:
        payload = await request.json()
    except Exception as exc:
        raise HTTPException(status_code=400, detail="Invalid JSON body") from exc
    if not isinstance(payload, dict):
        raise HTTPException(status_code=400, detail="Body must be a JSON object")
    try:
        updated = property_store.update_property(property_id, payload)
    except KeyError as exc:
        raise HTTPException(status_code=404, detail="Property not found") from exc
    except ValueError as exc:
        raise HTTPException(status_code=409, detail=str(exc)) from exc
    return {"ok": True, "property": updated, "summary": property_store.summarize_property(updated)}


@app.delete("/api/properties/{property_id}")
async def properties_delete(property_id: str, _: None = Depends(_require_call_api_key)):
    """Delete a listing from data/properties.json and reload cache."""
    if not property_store.delete_property(property_id):
        raise HTTPException(status_code=404, detail="Property not found")
    return {"ok": True, "deleted": property_id}


@app.get("/api/v1/listings/{property_id}")
@app.get("/api/properties/{property_id}")
async def properties_get(property_id: str):
    """Dummy listing details API — full structured mock property."""
    result = listings_api.api_details(property_id)
    if not result.get("ok"):
        raise HTTPException(status_code=404, detail=result.get("message") or "Property not found")
    return result


@app.get("/healthz")
async def healthz():
    return {"status": "ok", "agent": "andre", "role": "real_estate"}


@app.get("/health")
async def health():
    return {"status": "ok", "agent": "andre", "role": "real_estate"}


@app.get("/readyz")
async def readyz():
    missing = []
    for key in ("SARVAM_API_KEY", "OPENAI_API_KEY", "TWILIO_ACCOUNT_SID", "TWILIO_AUTH_TOKEN", "TWILIO_PHONE_NUMBER"):
        val = (os.environ.get(key) or "").strip()
        if not val or _is_placeholder_key(key, val):
            missing.append(key)
    if missing:
        return JSONResponse(
            status_code=503,
            content={"ready": False, "missing": missing},
        )
    return {"ready": True, "agent": "andre"}


@app.get("/readyz/browser")
async def readyz_browser():
    missing = []
    for key in ("SARVAM_API_KEY", "OPENAI_API_KEY"):
        val = (os.environ.get(key) or "").strip()
        if not val or _is_placeholder_key(key, val):
            missing.append(key)
    if missing:
        return JSONResponse(status_code=503, content={"ready": False, "missing": missing})
    return {"ready": True, "agent": "andre", "mode": "browser"}


@app.get("/calls")
async def get_calls(limit: int = 20):
    return list_calls(limit)


@app.get("/calls/{call_id}")
async def get_call(call_id: int):
    summary = get_call_summary(call_id)
    if not summary:
        raise HTTPException(status_code=404, detail="Call not found")
    return summary


@app.get("/transcripts")
async def get_transcripts(
    phone: str | None = None,
    email: str | None = None,
    call_id: int | None = None,
    limit: int = 500,
):
    return list_conversation_transcripts(phone=phone, email=email, call_id=call_id, limit=limit)


@app.post("/admin/call/{call_sid}/hangup")
async def admin_hangup(call_sid: str):
    _hangup_twilio_call(call_sid)
    task = _active_tasks.get(call_sid)
    if task:
        await task.queue_frames([EndFrame()])
    return {"status": "hangup_requested", "call_sid": call_sid}


def _public_base_reachable(timeout: float = 5.0) -> tuple[bool, str]:
    """Twilio Media Streams need a live HTTPS tunnel; fail fast if it's dead."""
    public_url = (os.environ.get("PUBLIC_BASE_URL") or "").rstrip("/")
    if not public_url:
        return False, "PUBLIC_BASE_URL is not set"
    if public_url.startswith("http://") and "localhost" not in public_url and "127.0.0.1" not in public_url:
        return False, "PUBLIC_BASE_URL must be HTTPS for Twilio"
    try:
        resp = _requests.get(f"{public_url}/healthz", timeout=timeout)
        if resp.status_code != 200:
            return False, f"Tunnel health returned HTTP {resp.status_code}"
        return True, public_url
    except Exception as exc:
        return False, f"Tunnel unreachable ({exc}). Start: cloudflared tunnel --url http://127.0.0.1:8001"


@app.post("/call")
async def initiate_call(request: Request, _: None = Depends(_require_call_api_key)):
    body = await request.json()
    raw_to = (body.get("to") or body.get("phone") or "").strip()
    to_number = normalize_e164(raw_to)
    if not to_number:
        raise HTTPException(status_code=400, detail="Invalid phone number (use E.164, e.g. +91…)")

    if is_opted_out(to_number):
        return {"status": "blocked", "reason": "internal_opt_out", "to": to_number}

    skip_compliance = bool(body.get("skip_compliance")) or (
        os.environ.get("ENVIRONMENT", "development").lower() == "development"
    )
    allowed, reason = _is_calling_allowed()
    if not allowed and not skip_compliance:
        return {
            "status": "queued",
            "reason": reason,
            "message": f"Outside calling window ({reason}).",
            "to": to_number,
        }

    campaign_id = (body.get("campaign_id") or "default").strip() or "default"
    if campaign_id in _paused_campaigns:
        return {"status": "blocked", "reason": "campaign_paused", "campaign_id": campaign_id}

    language = _normalize_language(body.get("language"))
    if not (os.environ.get("PUBLIC_BASE_URL") or "").strip():
        raise HTTPException(status_code=503, detail="PUBLIC_BASE_URL is not set")

    ok, detail = await asyncio.to_thread(_public_base_reachable)
    if not ok:
        logger.error("Refusing outbound call — public tunnel dead: %s", detail)
        raise HTTPException(
            status_code=503,
            detail=f"Public tunnel down — calls would drop on pickup. {detail}",
        )

    try:
        return start_outbound_call(to_number, campaign_id=campaign_id, language=language)
    except Exception as e:
        logger.error(f"/call failed: {e}")
        raise HTTPException(status_code=500, detail=str(e)) from e


@app.post("/twiml")
async def twiml_webhook(
    request: Request,
    _: None = Depends(_require_twilio_signature),
):
    """Inbound Twilio webhook. Language via ?language= or form Language / DEFAULT_LANGUAGE."""
    public_url = os.environ["PUBLIC_BASE_URL"].rstrip("/")
    ws_host = public_url.replace("https://", "").replace("http://", "")
    form = await request.form()
    lang = _normalize_language(
        request.query_params.get("language")
        or form.get("language")
        or form.get("Language")
        or os.environ.get("DEFAULT_LANGUAGE", "auto")
    )
    call_sid = str(form.get("CallSid") or "")
    from_number = str(form.get("From") or "")
    if call_sid:
        call_db_id = start_call(call_sid, from_number or "unknown", channel="phone", agent_mode="andre")
        _call_db_ids[call_sid] = call_db_id
        _call_languages[call_sid] = lang

    twiml = f"""<?xml version="1.0" encoding="UTF-8"?>
<Response>
  <Connect>
    <Stream url="wss://{ws_host}/stream">
      <Parameter name="agent" value="andre"/>
      <Parameter name="language" value="{lang}"/>
      <Parameter name="caller_phone" value="{from_number}"/>
    </Stream>
  </Connect>
</Response>"""
    return PlainTextResponse(twiml, media_type="application/xml")


@app.post("/twilio/status")
async def twilio_status(request: Request, _: None = Depends(_require_twilio_signature)):
    form = await request.form()
    call_sid = str(form.get("CallSid") or "")
    status = str(form.get("CallStatus") or "")
    logger.info(f"Twilio status: {call_sid} → {status}")
    return {"ok": True}


@app.websocket("/stream")
async def stream_websocket(websocket: WebSocket):
    await websocket.accept()
    call_sid = "unknown"
    stream_sid = "unknown"
    call_db_id = -1
    language = _normalize_language(None)
    custom_params: dict[str, str] = {}

    logger.info("Twilio WebSocket connected — reading call SID...")

    try:
        raw = await asyncio.wait_for(websocket.receive_text(), timeout=5.0)
        envelope = json.loads(raw)
        if envelope.get("event") != "connected":
            logger.warning(f"Unexpected first event: {envelope.get('event')}")

        raw2 = await asyncio.wait_for(websocket.receive_text(), timeout=5.0)
        envelope2 = json.loads(raw2)
        if envelope2.get("event") == "start":
            start_payload = envelope2.get("start", {}) or {}
            call_sid = start_payload.get("callSid", "unknown")
            stream_sid = start_payload.get("streamSid", "unknown")
            custom_params = _stream_custom_params(start_payload)
            language = _normalize_language(
                custom_params.get("language") or _call_languages.get(call_sid)
            )
            logger.info(f"Linked to call SID: {call_sid}  language={language}")
            caller_phone = (
                custom_params.get("caller_phone")
                or custom_params.get("from")
                or start_payload.get("from")
                or start_payload.get("to")
                or ""
            ).strip()
            if call_sid in _call_db_ids:
                call_db_id = _call_db_ids[call_sid]
                if caller_phone and not get_call_number(call_db_id):
                    update_call_contact(call_db_id, phone=caller_phone)
            else:
                call_db_id = start_call(
                    call_sid,
                    caller_phone or "unknown",
                    channel="phone",
                    agent_mode="andre",
                )
                _call_db_ids[call_sid] = call_db_id
            _call_languages[call_sid] = language
    except Exception as e:
        logger.warning(f"Could not read call SID from Twilio envelope: {e}")
        call_db_id = start_call("unknown", "unknown", channel="phone", agent_mode="andre")

    to_number_ref: list[str] = [get_call_number(call_db_id) or "unknown"]
    language_state: list[str] = [language]

    transport = FastAPIWebsocketTransport(
        websocket=websocket,
        params=FastAPIWebsocketParams(
            audio_in_enabled=True,
            audio_out_enabled=True,
            vad_enabled=True,
            vad_analyzer=SileroVADAnalyzer(params=VADParams(stop_secs=0.4)),
            serializer=TwilioFrameSerializer(
                stream_sid,
                call_sid=call_sid,
                account_sid=os.environ.get("TWILIO_ACCOUNT_SID"),
                auth_token=os.environ.get("TWILIO_AUTH_TOKEN"),
            ),
        ),
    )

    stt = _build_stt(language_state[0])
    system_prompt = build_andre_prompt(channel="phone", language=language_state[0])
    greeting_text = build_andre_greeting(language_state[0])

    llm = OpenAILLMService(
        api_key=os.environ["OPENAI_API_KEY"],
        model=os.environ.get("OPENAI_MODEL", "gpt-4o-mini"),
    )

    outcome_state: list[str] = ["unknown"]
    task_ref: list = [None]

    async def handle_tool_call(function_name, tool_call_id, arguments, llm, context, result_callback):
        result = await _dispatch_tool(
            function_name=function_name,
            arguments=arguments,
            call_db_id=call_db_id,
            to_number_ref=to_number_ref,
            outcome_state=outcome_state,
            task_ref=task_ref,
            language_state=language_state,
            call_sid=call_sid,
            is_browser=False,
        )
        if function_name == "set_language" and result.get("status") == "ok":
            lang = result.get("language") or language_state[0]
            new_prompt = build_andre_prompt(channel="phone", language=lang)
            msgs = getattr(context, "messages", None) or []
            if msgs and isinstance(msgs[0], dict) and msgs[0].get("role") == "system":
                msgs[0]["content"] = new_prompt
        spoken = None
        if function_name in _LOOKUP_TOOLS and isinstance(result, dict):
            spoken = await _speak_tool_answer(
                task_ref[0],
                call_db_id=call_db_id,
                function_name=function_name,
                result=result,
            )
            if spoken:
                result = {**result, "already_spoken": True, "spoken_text": spoken}
        # Pass the dict through — the aggregator JSON-encodes it. Double-encoding
        # (json.dumps here) confuses the next LLM turn and it goes silent after "one moment".
        # If we already spoke the listings answer, skip a second LLM turn (avoids silence + repeats).
        await result_callback(
            result,
            properties=FunctionCallResultProperties(run_llm=not bool(spoken)),
        )

    llm.register_function(None, handle_tool_call, cancel_on_interruption=False)

    context = LLMContext(
        messages=[{"role": "system", "content": system_prompt}],
        tools=andre_tools_schema(ANDRE_PHONE_TOOLS),
    )
    user_aggregator, assistant_aggregator = LLMContextAggregatorPair(context)
    tts = _build_tts(language_state[0])

    turn_timing: dict = {"user_turn_ts": 0.0, "first_text_pending": False}
    user_logger = TranscriptLogger(call_db_id, "user", timing_state=turn_timing)
    assistant_logger = TranscriptLogger(call_db_id, "assistant", timing_state=turn_timing)
    opt_out_guardrail = OptOutGuardrail(call_db_id, to_number_ref, outcome_state, task_ref)
    context_compactor = ContextCompactor(
        context, max_pairs=int(os.environ.get("LLM_CONTEXT_PAIRS", "8"))
    )

    pipeline = Pipeline(
        [
            transport.input(),
            stt,
            user_logger,
            opt_out_guardrail,
            context_compactor,
            user_aggregator,
            llm,
            assistant_logger,
            tts,
            transport.output(),
            assistant_aggregator,
        ]
    )

    task = PipelineTask(
        pipeline,
        params=PipelineParams(allow_interruptions=True, enable_metrics=True),
    )
    task_ref[0] = task
    _active_tasks[call_sid] = task

    call_start_time: list[float] = [0.0]
    _flushed: list[bool] = [False]
    max_duration_task: asyncio.Task | None = None

    def _do_flush():
        if _flushed[0]:
            return
        _flushed[0] = True
        duration = round(time.time() - call_start_time[0], 1) if call_start_time[0] else None
        _flush_call_log(call_sid, call_db_id, outcome_state[0], None, duration)

    def _cancel_max_duration():
        nonlocal max_duration_task
        if max_duration_task and not max_duration_task.done():
            max_duration_task.cancel()
        max_duration_task = None

    @transport.event_handler("on_client_connected")
    async def on_connected(transport, client):
        nonlocal max_duration_task
        call_start_time[0] = time.time()
        logger.info("Prospect answered — speaking greeting: %s", greeting_text[:80])
        log_turn(call_db_id, "assistant", greeting_text)
        await task.queue_frames([TTSSpeakFrame(text=greeting_text, append_to_context=True)])

        max_minutes = _phone_call_max_minutes()
        if max_minutes is not None:
            max_duration_task = asyncio.create_task(
                _max_duration_watchdog(
                    minutes=max_minutes,
                    call_sid=call_sid,
                    call_db_id=call_db_id,
                    outcome_state=outcome_state,
                    task_ref=task_ref,
                    flushed=_flushed,
                )
            )

    @transport.event_handler("on_client_disconnected")
    async def on_disconnected(transport, client):
        logger.info("Call ended")
        _cancel_max_duration()
        if outcome_state[0] == "unknown":
            outcome_state[0] = "completed"
        _do_flush()
        await task.queue_frames([EndFrame()])

    runner = PipelineRunner()
    try:
        await runner.run(task)
    finally:
        _cancel_max_duration()

    _do_flush()
    _active_tasks.pop(call_sid, None)
    _call_db_ids.pop(call_sid, None)
    _call_languages.pop(call_sid, None)
    logger.info(f"Pipeline done. DB id={call_db_id}  outcome={outcome_state[0]}")


@app.get("/ui-config")
async def ui_config():
    """Non-secret UI bootstrap. In development, also returns CALL_API_SECRET for local dialing."""
    env = os.environ.get("ENVIRONMENT", "development").lower()
    cfg: dict = {
        "agent": "andre",
        "environment": env,
        "default_language": _normalize_language(None),
    }
    if env == "development":
        cfg["call_api_key"] = (os.environ.get("CALL_API_SECRET") or "").strip()
    return cfg


@app.get("/")
async def serve_index():
    index = _STATIC_DIR / "index.html"
    if index.is_file():
        return FileResponse(index)
    return {"agent": "andre", "docs": "/healthz"}


@app.websocket("/browser-stream")
async def browser_stream(websocket: WebSocket):
    await websocket.accept()

    query = websocket.scope.get("query_string", b"").decode()
    params = dict(parse_qsl(query))
    language = _normalize_language(params.get("language"))

    session_id = f"browser-{uuid.uuid4().hex[:8]}"
    call_db_id = start_call(session_id, "", channel="browser", agent_mode="andre")
    language_state: list[str] = [language]
    _call_languages[session_id] = language

    system_prompt = build_andre_prompt(channel="browser", language=language)
    greeting_text = build_andre_greeting(language)

    logger.info(f"Browser session started: {session_id}  lang={language}  DB id={call_db_id}")

    try:
        await websocket.send_text(
            json.dumps(
                {
                    "event": "ready",
                    "session_id": session_id,
                    "call_db_id": call_db_id,
                    "language": language,
                    "agent": "andre",
                }
            )
        )
    except Exception:
        pass

    to_number_ref: list[str] = ["browser"]
    outcome_state: list[str] = ["unknown"]
    task_ref: list = [None]
    client_mailbox: dict = {}

    transport = FastAPIWebsocketTransport(
        websocket=websocket,
        params=FastAPIWebsocketParams(
            audio_in_enabled=True,
            audio_out_enabled=True,
            audio_in_sample_rate=16000,
            audio_out_sample_rate=16000,
            vad_enabled=True,
            vad_analyzer=SileroVADAnalyzer(params=VADParams(stop_secs=0.4)),
            serializer=BrowserFrameSerializer(mailbox=client_mailbox),
        ),
    )

    stt = _build_stt(language)
    llm = OpenAILLMService(
        api_key=os.environ["OPENAI_API_KEY"],
        model=os.environ.get("OPENAI_MODEL", "gpt-4o-mini"),
    )
    tts = _build_tts(language, sample_rate=16000)

    async def handle_browser_tool(function_name, tool_call_id, arguments, llm, context, result_callback):
        result = await _dispatch_tool(
            function_name=function_name,
            arguments=arguments,
            call_db_id=call_db_id,
            to_number_ref=to_number_ref,
            outcome_state=outcome_state,
            task_ref=task_ref,
            language_state=language_state,
            call_sid=session_id,
            is_browser=True,
        )
        if function_name == "set_language" and result.get("status") == "ok":
            lang = result.get("language") or language_state[0]
            new_prompt = build_andre_prompt(channel="browser", language=lang)
            msgs = getattr(context, "messages", None) or []
            if msgs and isinstance(msgs[0], dict) and msgs[0].get("role") == "system":
                msgs[0]["content"] = new_prompt
        spoken = None
        if function_name in _LOOKUP_TOOLS and isinstance(result, dict):
            spoken = await _speak_tool_answer(
                task_ref[0],
                call_db_id=call_db_id,
                function_name=function_name,
                result=result,
                event_ws=websocket,
            )
            if spoken:
                result = {**result, "already_spoken": True, "spoken_text": spoken}
        await result_callback(
            result,
            properties=FunctionCallResultProperties(run_llm=not bool(spoken)),
        )

    llm.register_function(None, handle_browser_tool, cancel_on_interruption=False)

    @llm.event_handler("on_function_calls_started")
    async def on_browser_function_calls_started(service, function_calls):
        # Local inventory lookups are instant — skip "just a moment" filler so
        # the tool answer can speak immediately without being blocked/interrupted.
        return

    context = LLMContext(
        messages=[{"role": "system", "content": system_prompt}],
        tools=andre_tools_schema(ANDRE_BROWSER_TOOLS),
    )
    user_aggregator, assistant_aggregator = LLMContextAggregatorPair(context)

    turn_timing: dict = {"user_turn_ts": 0.0, "first_text_pending": False}
    max_pairs = int(os.environ.get("LLM_CONTEXT_PAIRS", "8"))
    user_logger = TranscriptLogger(call_db_id, "user", timing_state=turn_timing, event_ws=websocket)
    assistant_logger = TranscriptLogger(
        call_db_id, "assistant", timing_state=turn_timing, event_ws=websocket
    )
    opt_out_guardrail = OptOutGuardrail(call_db_id, to_number_ref, outcome_state, task_ref)
    context_compactor = ContextCompactor(context, max_pairs=max_pairs)
    status_user = StatusEventSender(websocket)
    status_bot = StatusEventSender(websocket)

    pipeline = Pipeline(
        [
            transport.input(),
            status_user,
            stt,
            user_logger,
            opt_out_guardrail,
            context_compactor,
            user_aggregator,
            llm,
            assistant_logger,
            tts,
            status_bot,
            transport.output(),
            assistant_aggregator,
        ]
    )

    task = PipelineTask(
        pipeline,
        params=PipelineParams(allow_interruptions=True, enable_metrics=True),
    )
    task_ref[0] = task

    call_start_time: list[float] = [0.0]
    _flushed: list[bool] = [False]
    greeting_streamed = False

    def _do_flush():
        if _flushed[0]:
            return
        _flushed[0] = True
        duration = round(time.time() - call_start_time[0], 1) if call_start_time[0] else None
        _flush_call_log(session_id, call_db_id, outcome_state[0], None, duration)

    @transport.event_handler("on_client_connected")
    async def on_browser_connected(transport, client):
        call_start_time[0] = time.time()
        logger.info(f"Browser session live: {session_id}")
        if not greeting_streamed:
            # HTTP TTS failed — speak greeting via pipeline Sarvam TTS websocket
            logger.info("Speaking browser greeting via pipeline TTS fallback")
            await task.queue_frames(
                [TTSSpeakFrame(text=greeting_text, append_to_context=False)]
            )

    @transport.event_handler("on_client_disconnected")
    async def on_browser_disconnected(transport, client):
        logger.info(f"Browser session ended: {session_id}")
        if outcome_state[0] == "unknown":
            outcome_state[0] = "completed"
        _do_flush()
        await task.queue_frames([EndFrame()])

    greeting_streamed = await _stream_browser_greeting(
        websocket,
        call_db_id=call_db_id,
        context=context,
        greeting_text=greeting_text,
        language=language,
    )
    if not greeting_streamed:
        try:
            log_turn(call_db_id, "assistant", greeting_text)
            context.messages.append({"role": "assistant", "content": greeting_text})
            await websocket.send_text(
                json.dumps({"event": "transcript", "role": "assistant", "text": greeting_text})
            )
        except Exception as e:
            logger.warning(f"Greeting transcript fallback failed: {e}")

    runner = PipelineRunner()
    try:
        await runner.run(task)
    except Exception as e:
        logger.error(f"Browser pipeline error: {e}")
        try:
            await websocket.send_text(json.dumps({"event": "error", "message": str(e)}))
        except Exception:
            pass
        raise

    _do_flush()
    _call_languages.pop(session_id, None)
    logger.info(f"Browser pipeline done. DB id={call_db_id}  outcome={outcome_state[0]}")


# ── Helpers ───────────────────────────────────────────────────────────────────

def _flush_call_log(
    call_sid: str,
    call_db_id: int,
    outcome: str,
    booking: dict | None,
    duration_s: float | None = None,
):
    try:
        end_call(call_db_id, outcome, booking, duration_s=duration_s)
        logger.info(f"Call log flushed: sid={call_sid} id={call_db_id} outcome={outcome}")
    except Exception as e:
        logger.error(f"Failed to flush call log: {e}")


async def _delayed_end(task: PipelineTask, delay: float = 3.0):
    await asyncio.sleep(delay)
    try:
        await task.queue_frames([EndFrame()])
    except Exception:
        pass


def _phone_call_max_minutes() -> float | None:
    raw = (os.environ.get("PHONE_CALL_MAX_MINUTES") or "").strip()
    if not raw:
        return None
    try:
        minutes = float(raw)
    except ValueError:
        return None
    return None if minutes <= 0 else minutes


def _hangup_twilio_call(call_sid: str) -> None:
    try:
        tc = TwilioClient(os.environ["TWILIO_ACCOUNT_SID"], os.environ["TWILIO_AUTH_TOKEN"])
        tc.calls(call_sid).update(status="completed")
    except Exception as e:
        logger.warning(f"Hangup failed for {call_sid}: {e}")


async def _max_duration_watchdog(
    *,
    minutes: float,
    call_sid: str,
    call_db_id: int,
    outcome_state: list[str],
    task_ref: list,
    flushed: list[bool],
):
    try:
        await asyncio.sleep(minutes * 60)
        logger.info(f"Max duration reached ({minutes} min) — hanging up {call_sid}")
        if outcome_state[0] == "unknown":
            outcome_state[0] = "max_duration"
        _hangup_twilio_call(call_sid)
        task = task_ref[0]
        if task:
            await task.queue_frames([EndFrame()])
        if not flushed[0]:
            flushed[0] = True
            _flush_call_log(call_sid, call_db_id, outcome_state[0], None, duration_s=minutes * 60)
    except asyncio.CancelledError:
        pass


def _is_placeholder_key(name: str, value: str) -> bool:
    v = value.lower()
    placeholders = ("your_", "change-me", "xxxxxxxx", "placeholder", "example")
    return any(p in v for p in placeholders)


if __name__ == "__main__":
    port = int(os.environ.get("PORT", "8001"))
    uvicorn.run("bot:app", host="0.0.0.0", port=port, reload=False)
