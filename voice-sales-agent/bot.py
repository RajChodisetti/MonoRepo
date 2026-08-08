"""Inbound-only corporate and restaurant voice assistant runtime."""

import asyncio
import hmac
import json
import logging
import os
import time
import uuid
from contextlib import asynccontextmanager
from datetime import datetime, timedelta, date
from pathlib import Path
from urllib.parse import parse_qsl
from zoneinfo import ZoneInfo

import requests as _requests
import uvicorn
from dotenv import load_dotenv
from fastapi import FastAPI, WebSocket, Request, Depends, HTTPException
from fastapi.responses import FileResponse, PlainTextResponse, JSONResponse
from fastapi.staticfiles import StaticFiles
from twilio.rest import Client as TwilioClient
from twilio.request_validator import RequestValidator
import phonenumbers

# ── Pipecat core ──────────────────────────────────────────────────────────────
from pipecat.pipeline.pipeline import Pipeline
from pipecat.pipeline.task import PipelineTask, PipelineParams
from pipecat.pipeline.runner import PipelineRunner

# ── Frames ────────────────────────────────────────────────────────────────────
from pipecat.frames.frames import (
    LLMMessagesFrame,
    EndFrame,
    TranscriptionFrame,
    TextFrame,
    TTSSpeakFrame,
    InputAudioRawFrame,
    OutputAudioRawFrame,
    UserStartedSpeakingFrame,
    UserStoppedSpeakingFrame,
    BotStartedSpeakingFrame,
    BotStoppedSpeakingFrame,
)

# ── Twilio Transport ──────────────────────────────────────────────────────────
from pipecat.transports.websocket.fastapi import FastAPIWebsocketTransport, FastAPIWebsocketParams
from pipecat.serializers.twilio import TwilioFrameSerializer
from browser_serializer import BrowserFrameSerializer

# ── STT ───────────────────────────────────────────────────────────────────────
from pipecat.services.deepgram.stt import DeepgramSTTService

# ── LLM ───────────────────────────────────────────────────────────────────────
from pipecat.services.openai.llm import OpenAILLMService
from pipecat.processors.aggregators.llm_context import LLMContext
from pipecat.processors.aggregators.llm_response_universal import LLMContextAggregatorPair

# ── TTS ───────────────────────────────────────────────────────────────────────
from pipecat.services.cartesia.tts import CartesiaTTSService, GenerationConfig

# ── VAD ───────────────────────────────────────────────────────────────────────
from pipecat.audio.vad.silero import SileroVADAnalyzer
from pipecat.audio.vad.vad_analyzer import VADParams

# ── Processors ────────────────────────────────────────────────────────────────
from pipecat.processors.frame_processor import FrameProcessor, FrameDirection
from pipecat.services.cartesia.tts import CartesiaTTSSettings
from pipecat.transcriptions.language import Language

# ── Call Logger (local SQLite) ────────────────────────────────────────────────
from logger_db import (
    init_db, start_call, log_turn, end_call,
    get_call_summary, list_calls, list_conversation_transcripts,
    is_opted_out, record_opt_out, get_call_number, update_call_contact,
)

import api_client
import tuvi_api_client
from prompts.restaurant import build_restaurant_greeting, build_restaurant_prompt
from prompts.tools_restaurant import RESTAURANT_TOOLS, RESTAURANT_BROWSER_TOOLS
from prompts.corporate import build_corporate_greeting, build_corporate_prompt
from prompts.tools_corporate import CORPORATE_TOOLS, CORPORATE_PHONE_TOOLS

load_dotenv()

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s — %(message)s",
)
logger = logging.getLogger("VoiceSalesAgent")

init_db()


def _cartesia_pipecat_language() -> Language:
    """Map CARTESIA_LANGUAGE env (e.g. en_in, en-IN) to Pipecat Language enum."""
    lang_key = os.environ.get("CARTESIA_LANGUAGE", "en_in").upper().replace("-", "_")
    return getattr(Language, lang_key, Language.EN_IN)


def _cartesia_api_language() -> str:
    """Cartesia REST/WS language tag, e.g. en-in."""
    return os.environ.get("CARTESIA_LANGUAGE", "en_in").lower().replace("_", "-")


async def _fetch_cartesia_tts_bytes(transcript: str) -> bytes | None:
    """Synthesize PCM16 @ 16kHz via Cartesia REST (used for instant browser greetings)."""
    cartesia_key = os.environ.get("CARTESIA_API_KEY", "")
    voice_id = os.environ.get("CARTESIA_VOICE_ID", "")
    if not (cartesia_key and voice_id and transcript.strip()):
        return None

    def _fetch() -> bytes:
        resp = _requests.post(
            "https://api.cartesia.ai/tts/bytes",
            headers={
                "Cartesia-Version": "2025-04-16",
                "X-API-Key": cartesia_key,
                "Content-Type": "application/json",
            },
            json={
                "model_id": "sonic-3.5",
                "transcript": transcript,
                "voice": {"mode": "id", "id": voice_id},
                "output_format": {"container": "raw", "encoding": "pcm_s16le", "sample_rate": 16000},
                "language": _cartesia_api_language(),
                "generation_config": {"speed": 1.05, "emotion": "positivity:medium"},
            },
            timeout=30,
        )
        resp.raise_for_status()
        return resp.content

    try:
        return await asyncio.to_thread(_fetch)
    except Exception as exc:
        logger.warning("Cartesia TTS fetch failed: %s", exc)
        return None


async def _stream_browser_greeting(
    websocket: WebSocket,
    *,
    call_db_id: int,
    context: LLMContext,
    greeting_text: str,
) -> bool:
    """
    Stream pre-synthesized greeting audio into the WS buffer before the pipeline starts.
    Returns True if audio was sent (caller should skip LLM-generated greeting).
    """
    audio = await _fetch_cartesia_tts_bytes(greeting_text)
    if not audio:
        return False

    try:
        log_turn(call_db_id, "assistant", greeting_text)
        context.messages.append({"role": "assistant", "content": greeting_text})
        await websocket.send_text(json.dumps({"event": "status", "state": "bot_speaking"}))
        chunk_size = 3200  # 100 ms PCM16 @ 16 kHz mono
        for i in range(0, len(audio), chunk_size):
            await websocket.send_bytes(audio[i:i + chunk_size])
        await websocket.send_text(json.dumps({"event": "transcript", "role": "assistant", "text": greeting_text}))
        await websocket.send_text(json.dumps({"event": "status", "state": "listening"}))
        duration_s = len(audio) / 32000
        logger.info("[GREETING] Streamed %s bytes (%.1fs): %s", f"{len(audio):,}", duration_s, greeting_text[:60])
        return True
    except Exception as exc:
        logger.warning("Browser greeting stream failed: %s", exc)
        return False

# ==============================================================================
#  COMPLIANCE CONSTANTS
#  ACMA telemarketing rules — https://www.acma.gov.au/say-no-to-telemarketers
# ==============================================================================

SYDNEY_TZ = ZoneInfo("Australia/Sydney")

# National public holidays 2026 — update annually or integrate an API for state holidays.
# Current list covers nation-wide holidays only; add your state's additional dates.
_AU_PUBLIC_HOLIDAYS = {
    date(2026, 1, 1),    # New Year's Day
    date(2026, 1, 26),   # Australia Day
    date(2026, 4, 3),    # Good Friday
    date(2026, 4, 4),    # Easter Saturday
    date(2026, 4, 5),    # Easter Sunday
    date(2026, 4, 6),    # Easter Monday
    date(2026, 4, 25),   # Anzac Day
    date(2026, 6, 8),    # King's Birthday (most states — check your state)
    date(2026, 12, 25),  # Christmas Day
    date(2026, 12, 26),  # Boxing Day
}

# Phrases that constitute an opt-out request (used by OptOutGuardrail).
# Deliberately specific to avoid false positives from phrases like "not interested in that plan".
OPT_OUT_PHRASES = [
    "stop calling",
    "don't call me",
    "do not call",
    "remove me",
    "take me off",
    "opt me out",
    "this is harassment",
    "add me to your do not call",
    "never call again",
    "not interested and don't call",
    "not interested and do not call",
    "never contact me",
]


# ==============================================================================
#  HELPERS
# ==============================================================================

def normalize_e164(number: str, default_region: str = "AU") -> str | None:
    """Return E.164-formatted number or None if invalid."""
    try:
        parsed = phonenumbers.parse(number, default_region)
        if phonenumbers.is_valid_number(parsed):
            return phonenumbers.format_number(parsed, phonenumbers.PhoneNumberFormat.E164)
    except Exception:
        pass
    return None


def _is_calling_allowed() -> tuple[bool, str]:
    """
    ACMA telemarketing calling windows (Australia/Sydney timezone):
      Mon–Fri  09:00–20:00
      Saturday 09:00–17:00
      Sunday   — never
      Public holidays — never

    Returns (allowed: bool, reason: str).
    Reason is 'ok', 'sunday', 'public_holiday', or 'outside_hours'.
    """
    now = datetime.now(SYDNEY_TZ)
    today = now.date()
    weekday = now.weekday()   # 0=Mon … 6=Sun
    hour = now.hour + now.minute / 60.0

    if today in _AU_PUBLIC_HOLIDAYS:
        return False, "public_holiday"
    if weekday == 6:
        return False, "sunday"
    if weekday == 5:   # Saturday
        return (True, "ok") if 9.0 <= hour < 17.0 else (False, "outside_hours")
    # Monday – Friday
    return (True, "ok") if 9.0 <= hour < 20.0 else (False, "outside_hours")


# ==============================================================================
#  TRANSCRIPT LOGGER
# ==============================================================================
class TranscriptLogger(FrameProcessor):
    """
    Transparent processor that logs user/assistant turns to the DB and optionally
    forwards transcript events as JSON over a browser WebSocket.
    """

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

    async def _send_event(self, data: dict):
        if self._event_ws is None:
            return
        try:
            await self._event_ws.send_text(json.dumps(data))
        except Exception:
            pass

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

        elif self._role == "assistant" and isinstance(frame, TextFrame):
            text = frame.text.strip()
            if text:
                log_turn(self._call_id, "assistant", text)
                logger.info(f"[ASSISTANT] {text}")
                if (
                    self._timing is not None
                    and self._timing.get("first_text_pending")
                    and self._timing.get("user_turn_ts")
                ):
                    latency_ms = round((time.time() - self._timing["user_turn_ts"]) * 1000)
                    logger.info(f"[LATENCY] turn_end→first_text: {latency_ms}ms")
                    self._timing["first_text_pending"] = False
                await self._send_event({"event": "transcript", "role": "assistant", "text": text})

        await self.push_frame(frame, direction)


# ==============================================================================
#  STATUS EVENT SENDER  (browser sessions only)
#  Observes VAD / TTS speaking frames and forwards status events as JSON to
#  the browser WebSocket. Place one instance before STT and one after TTS so
#  both user- and bot-speaking events are captured.
# ==============================================================================
class StatusEventSender(FrameProcessor):
    """Forwards user/bot speaking state as JSON text frames over a WebSocket."""

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


# ==============================================================================
#  OPT-OUT GUARDRAIL
#  Safety net: records opt-out in DB immediately even if LLM misses it.
#  Does NOT block the frame — lets LLM generate the polite goodbye.
#  Schedules a forced call end (12 s) so the call ends even if end_call tool
#  is never called.
# ==============================================================================
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
                    # Allow LLM/TTS to say goodbye, then force-end
                    asyncio.create_task(_delayed_end(task, delay=12.0))

        await self.push_frame(frame, direction)


# ==============================================================================
#  CONTEXT COMPACTOR
#  Sits after OptOutGuardrail, before user_aggregator.
#  At the start of each user turn, trims the LLM context to the last max_pairs
#  user/assistant turn pairs. System prompt is always preserved.
#  This keeps the hot-path prompt small and prevents cost/latency growth on long calls.
#  (Architecture guide §7.4)
# ==============================================================================
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
        system_idxs = [i for i, m in enumerate(messages) if m.get("role") == "system"]
        other_idxs = [i for i, m in enumerate(messages) if m.get("role") != "system"]
        max_keep = self._max_pairs * 2
        if len(other_idxs) <= max_keep:
            return
        drop_idxs = other_idxs[: len(other_idxs) - max_keep]
        for i in reversed(drop_idxs):
            del messages[i]
        logger.info(
            f"[CONTEXT] Compacted: dropped {len(drop_idxs)} msgs, kept last {self._max_pairs} turn pairs"
        )


# ==============================================================================
#  TWILIO SIGNATURE VALIDATION  (architecture guide §16.1)
#  Skip in development (ENVIRONMENT=development) since ngrok/tunnels break HMAC.
#  Set ENVIRONMENT=production to enable.
# ==============================================================================
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
        logger.warning(f"Invalid Twilio signature rejected: {url}")
        raise HTTPException(status_code=403, detail="Invalid Twilio signature")


async def _require_call_api_key(request: Request):
    """Protect transcript and administrative voice-control endpoints."""
    secret = (os.environ.get("CALL_API_SECRET") or "").strip()
    if not secret:
        env = os.environ.get("ENVIRONMENT", "development").lower()
        if env == "development":
            logger.warning("CALL_API_SECRET unset — allowing voice admin access in development")
            return
        raise HTTPException(status_code=503, detail="CALL_API_SECRET is not configured")

    header_key = (request.headers.get("X-Call-Api-Key") or "").strip()
    auth = (request.headers.get("Authorization") or "").strip()
    bearer = ""
    if auth.lower().startswith("bearer "):
        bearer = auth[7:].strip()
    provided = header_key or bearer
    if not provided or not hmac.compare_digest(provided, secret):
        raise HTTPException(status_code=401, detail="Invalid or missing call API key")


def _stream_custom_params(start_payload: dict) -> dict[str, str]:
    """Parse Twilio Media Stream start.customParameters (dict or name/value list)."""
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


def _normalize_agent_mode(raw: str | None, default: str = "corporate") -> str:
    safe_default = (default or "corporate").strip().lower()
    if safe_default not in {"corporate", "restaurant"}:
        safe_default = "corporate"

    mode = (raw or safe_default).strip().lower()
    if mode in {"corporate", "tuvi"}:
        return "corporate"
    if mode == "restaurant":
        return "restaurant"
    return safe_default


# Populated before FastAPI routes; used by start_outbound_call and stream handlers.
_active_tasks: dict[str, PipelineTask] = {}
_call_db_ids: dict[str, int] = {}
_call_agent_modes: dict[str, str] = {}  # call_sid → corporate | restaurant
_paused_campaigns: set[str] = set()


def start_outbound_call(
    to_number: str,
    *,
    campaign_id: str = "default",
    agent_mode: str = "corporate",
    restaurant_index: int | None = None,
    from_number: str | None = None,
    caller_verified: bool = False,
    restaurant_name: str | None = None,
    restaurant_phone_display: str | None = None,
) -> dict:
    """Fail closed: Phase 1 supports inbound voice sessions only."""
    del (
        to_number,
        campaign_id,
        agent_mode,
        restaurant_index,
        from_number,
        caller_verified,
        restaurant_name,
        restaurant_phone_display,
    )
    raise RuntimeError("Outbound AI calls are disabled for this release.")


# ==============================================================================
#  STARTUP PRE-WARMING
#  Eliminates the SileroVADAnalyzer ONNX model load on the first inbound call.
# ==============================================================================
async def _prewarm_services():
    # Pre-load Silero ONNX model so first-call instantiation is fast.
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


async def _emit_browser_booking_event(
    websocket,
    *,
    reservation_id: str,
    status: str,
    guest_name: str,
    guest_phone: str,
    party_size: int,
    slot: str,
    booking_date: str = "",
    booking_time: str = "",
    message: str = "",
    delay_seconds: float = 0,
) -> None:
    if delay_seconds > 0:
        await asyncio.sleep(delay_seconds)
    try:
        await websocket.send_text(
            json.dumps(
                {
                    "event": "booking",
                    "reservation_id": reservation_id,
                    "status": status,
                    "guest_name": guest_name,
                    "guest_phone": guest_phone,
                    "party_size": party_size,
                    "slot": slot,
                    "booking_date": booking_date,
                    "booking_time": booking_time,
                    "message": message,
                }
            )
        )
    except Exception as exc:
        logger.warning(f"Failed to emit booking event: {exc}")


async def _emit_browser_consultation_event(
    websocket,
    *,
    confirmation_code: str,
    prospect_name: str,
    prospect_email: str = "",
    slot: str = "",
    booking_date: str = "",
    booking_time: str = "",
    calendly_link: str = "",
    calendar_link: str = "",
    message: str = "",
) -> None:
    link = calendar_link or calendly_link
    try:
        await websocket.send_text(
            json.dumps(
                {
                    "event": "consultation",
                    "confirmation_code": confirmation_code,
                    "prospect_name": prospect_name,
                    "prospect_email": prospect_email,
                    "slot": slot,
                    "booking_date": booking_date,
                    "booking_time": booking_time,
                    "calendly_link": link,
                    "calendar_link": link,
                    "message": message,
                }
            )
        )
    except Exception as exc:
        logger.warning(f"Failed to emit consultation event: {exc}")


async def _emit_booking_progress(websocket, phase: str, message: str = "") -> None:
    """Push slot-check / booking UI state to the browser widget."""
    try:
        await websocket.send_text(
            json.dumps({"event": "booking_progress", "phase": phase, "message": message})
        )
    except Exception as exc:
        logger.warning(f"Failed to emit booking progress: {exc}")


async def _wait_for_typed_email(
    websocket,
    mailbox: dict,
    *,
    call_db_id: int,
    prompt: str = "",
    timeout_s: float = 120.0,
) -> dict:
    """
    Open an on-screen email field in the browser widget and wait for the visitor
    to type + submit it (event: user_email).
    """
    loop = asyncio.get_running_loop()
    # Cancel any previous waiter
    old = mailbox.get("email_future")
    if old is not None and not old.done():
        old.cancel()

    fut: asyncio.Future = loop.create_future()
    mailbox["email_future"] = fut
    message = (prompt or "").strip() or "Enter your email to book the consultation."

    try:
        await websocket.send_text(
            json.dumps({
                "event": "request_email",
                "message": message,
            })
        )
    except Exception as exc:
        mailbox.pop("email_future", None)
        logger.warning("Failed to emit request_email: %s", exc)
        return {"status": "error", "message": "Could not open the email form."}

    log_turn(call_db_id, "assistant", message)
    logger.info("Waiting for typed email (timeout=%ss)", timeout_s)

    try:
        email = await asyncio.wait_for(fut, timeout=timeout_s)
    except asyncio.TimeoutError:
        return {
            "status": "error",
            "message": "Timed out waiting for email. Open the form again with request_typed_email.",
        }
    except asyncio.CancelledError:
        return {"status": "error", "message": "Email collection was cancelled."}
    finally:
        if mailbox.get("email_future") is fut:
            mailbox.pop("email_future", None)
        try:
            await websocket.send_text(json.dumps({"event": "email_prompt_closed"}))
        except Exception:
            pass

    email = (email or "").strip().lower()
    if not email or "@" not in email or "." not in email.split("@")[-1]:
        return {
            "status": "error",
            "message": "That email looks invalid. Call request_typed_email again.",
        }

    log_turn(call_db_id, "user", email)
    update_call_contact(call_db_id, email=email)
    return {
        "status": "success",
        "email": email,
        "message": f"Visitor typed email: {email}. Use this in book_consultation.",
    }


async def _wait_for_booking_details(
    websocket,
    mailbox: dict,
    *,
    call_db_id: int,
    prompt: str = "",
    timeout_s: float = 180.0,
) -> dict:
    """Open the browser booking form and wait for name, email, and phone."""
    loop = asyncio.get_running_loop()
    old = mailbox.get("booking_details_future")
    if old is not None and not old.done():
        old.cancel()

    fut: asyncio.Future = loop.create_future()
    mailbox["booking_details_future"] = fut
    message = (prompt or "").strip() or "Enter your details to confirm the consultation."

    try:
        await websocket.send_text(
            json.dumps({"event": "request_booking_details", "message": message})
        )
    except Exception as exc:
        mailbox.pop("booking_details_future", None)
        logger.warning("Failed to emit request_booking_details: %s", exc)
        return {"status": "error", "message": "Could not open the booking form."}

    log_turn(call_db_id, "assistant", message)
    logger.info("Waiting for typed booking details (timeout=%ss)", timeout_s)

    try:
        details = await asyncio.wait_for(fut, timeout=timeout_s)
    except asyncio.TimeoutError:
        return {
            "status": "error",
            "message": "Timed out waiting for booking details. Open the form again.",
        }
    except asyncio.CancelledError:
        return {"status": "error", "message": "Booking details collection was cancelled."}
    finally:
        if mailbox.get("booking_details_future") is fut:
            mailbox.pop("booking_details_future", None)
        try:
            await websocket.send_text(
                json.dumps({"event": "booking_details_prompt_closed"})
            )
        except Exception:
            pass

    name = (details.get("prospect_name") or "").strip()
    email = (details.get("prospect_email") or "").strip().lower()
    phone = (details.get("prospect_phone") or "").strip()
    if not name:
        return {"status": "error", "message": "Name is required. Open the form again."}
    if not email or "@" not in email or "." not in email.split("@")[-1]:
        return {"status": "error", "message": "Email is invalid. Open the form again."}
    if not phone:
        return {"status": "error", "message": "Phone is required. Open the form again."}

    submitted = {
        "prospect_name": name,
        "prospect_email": email,
        "prospect_phone": phone,
    }
    mailbox["booking_details"] = submitted
    update_call_contact(call_db_id, contact_name=name, email=email, phone=phone)
    return {
        "status": "success",
        **submitted,
        "message": "The visitor submitted the booking form. Book the confirmed slot now.",
    }


# ==============================================================================
#  SHARED TOOL DISPATCHER
#  Called by both the Twilio and browser pipeline tool handlers.
# ==============================================================================
async def _dispatch_tool(
    *,
    function_name: str,
    arguments: dict,
    call_db_id: int,
    to_number_ref: list[str],
    outcome_state: list[str],
    task_ref: list,
    booking_state: dict,
    call_sid: str = "unknown",
    is_browser: bool = False,
    restaurant_id: str | None = None,
    restaurant_index: int | None = None,
) -> dict:
    logger.info(f"Tool: {function_name}  args={arguments}  browser={is_browser}  restaurant={restaurant_id}")

    if function_name in {"place_callback_call", "place_restaurant_callback", "send_followup_sms"}:
        return {
            "status": "disabled",
            "message": "Outbound calls and messages are disabled for this release.",
        }

    if function_name == "check_table_availability":
        if not restaurant_id:
            return {"status": "error", "message": "Restaurant booking is not configured for this session."}
        date_str = (arguments.get("date") or "").strip()
        try:
            party_size = int(arguments.get("party_size") or 2)
        except (TypeError, ValueError):
            party_size = 2
        return await api_client.get_table_availability(restaurant_id, date_str, party_size)

    if function_name == "book_table_reservation":
        if not restaurant_id:
            return {"status": "error", "message": "Restaurant booking is not configured for this session."}
        body = {
            "slot": arguments.get("slot", ""),
            "guest_name": arguments.get("guest_name", ""),
            "guest_phone": arguments.get("guest_phone", ""),
            "guest_email": arguments.get("guest_email", ""),
            "party_size": int(arguments.get("party_size") or 2),
            "notes": arguments.get("notes", ""),
            "client_request_id": call_sid,
        }
        result = await api_client.put_reservation(restaurant_id, body)
        if result.get("status") == "pending" and result.get("reservation_id"):
            booking_state.update({
                "type": "table_reservation",
                "reservation_id": result.get("reservation_id"),
                "status": result.get("status"),
                "slot": body["slot"],
                "guest_name": body["guest_name"],
                "guest_phone": body["guest_phone"],
                "party_size": body["party_size"],
                "message": result.get("message"),
            })
            outcome_state[0] = "reservation_requested"
            result = {
                **result,
                "guest_name": body["guest_name"],
                "guest_phone": body["guest_phone"],
                "party_size": body["party_size"],
                "slot": body["slot"],
            }
        return result

    if function_name == "check_calendar_availability":
        return {"status": "success", "available_slots": _mock_available_slots()}

    if function_name == "check_consultation_slot":
        date_str = (arguments.get("date") or "").strip()
        time_str = (arguments.get("time") or "").strip()
        if not date_str or not time_str:
            return {"status": "error", "message": "Need date (YYYY-MM-DD) and time to check availability."}
        result = await tuvi_api_client.check_consultation_slot(date_str, time_str)
        if result.get("status") == "success" and result.get("available"):
            return {
                "status": "success",
                "available": True,
                "date": result.get("date", date_str),
                "time": result.get("time", time_str),
                "slot": result.get("slot", ""),
                "message": f"{date_str} at {time_str} is available.",
            }
        if result.get("status") == "success" and not result.get("available"):
            alts = result.get("alternatives") or []
            return {
                "status": "unavailable",
                "available": False,
                "message": "That slot is not available.",
                "alternatives": alts,
            }
        return result

    if function_name == "check_consultation_slots":
        date_str = (arguments.get("date") or "").strip()
        days_raw = arguments.get("days")
        days = 2
        if days_raw is not None:
            try:
                days = int(days_raw)
            except (TypeError, ValueError):
                days = 2
        return await tuvi_api_client.get_consultation_availability(date_str or None, days)

    if function_name == "book_consultation":
        date_str = (arguments.get("date") or "").strip()
        time_str = (arguments.get("time") or "").strip()
        prospect_name = (arguments.get("prospect_name") or "Guest").strip() or "Guest"
        prospect_email = (arguments.get("prospect_email") or "").strip()
        prospect_phone = (arguments.get("prospect_phone") or "").strip()
        if not date_str or not time_str:
            return {"status": "error", "message": "Need date (YYYY-MM-DD) and time before booking."}
        if not prospect_email:
            return {"status": "error", "message": "Need prospect_email before booking — ask for their email."}
        if not prospect_phone:
            return {"status": "error", "message": "Need prospect_phone before booking - ask for their phone number."}
        result = await tuvi_api_client.book_consultation(
            date=date_str,
            time=time_str,
            prospect_name=prospect_name,
            prospect_email=prospect_email,
            prospect_phone=prospect_phone,
            source="voice",
        )
        if result.get("status") == "success":
            booking_state.update({
                "type": "consultation",
                "confirmation_code": result.get("confirmation_code", ""),
                "slot": result.get("slot", ""),
                "booking_date": result.get("booking_date", date_str),
                "booking_time": result.get("booking_time", time_str),
                "prospect_name": prospect_name,
                "prospect_email": prospect_email,
                "prospect_phone": prospect_phone,
                "calendly_link": result.get("calendar_link") or result.get("calendly_link", ""),
            })
            outcome_state[0] = "booked"
            update_call_contact(
                call_db_id,
                phone=prospect_phone,
                email=prospect_email,
                contact_name=prospect_name,
            )
        return result

    if function_name == "book_appointment":
        slot  = arguments.get("slot", "TBD")
        name  = arguments.get("prospect_name", "Prospect")
        email = arguments.get("prospect_email", "")
        booking_state.update({"slot": slot, "prospect_name": name, "prospect_email": email})
        outcome_state[0] = "booked"
        result = {
            "status": "booked",
            "confirmed_slot": slot,
            "prospect_name": name,
            "calendar_link": f"https://cal.tuvisolutions.com/demo/{slot.replace(':', '-')}",
            "message": f"Discovery call booked for {name} at {slot}.",
        }
        logger.info(f"Booking confirmed: {result}")
        return result

    if function_name == "mark_do_not_call":
        phone = to_number_ref[0]
        record_opt_out(phone, call_db_id)
        outcome_state[0] = "opted_out"
        logger.warning(f"mark_do_not_call tool called for {phone}")
        return {"status": "recorded", "message": "Number added to internal do-not-call list."}

    if function_name == "transfer_to_human":
        if is_browser:
            outcome_state[0] = "transferred"
            if task_ref[0]:
                asyncio.create_task(_delayed_end(task_ref[0], delay=3.0))
            return {
                "status": "info",
                "message": "A team member will follow up with you shortly. Ending this session.",
            }
        human_number = os.environ.get("HUMAN_AGENT_NUMBER", "")
        if human_number and call_sid != "unknown":
            try:
                public_url = os.environ["PUBLIC_BASE_URL"].rstrip("/")
                tc = TwilioClient(os.environ["TWILIO_ACCOUNT_SID"], os.environ["TWILIO_AUTH_TOKEN"])
                tc.calls(call_sid).update(
                    url=f"{public_url}/transfer-twiml?to={human_number}",
                    method="POST",
                )
                outcome_state[0] = "transferred"
                if task_ref[0]:
                    asyncio.create_task(_delayed_end(task_ref[0], delay=2.0))
                logger.info(f"Warm transfer initiated to {human_number}")
                return {"status": "transferring"}
            except Exception as e:
                logger.error(f"Transfer failed: {e}")
                return {"status": "error", "message": str(e)}
        return {"status": "error", "message": "Human transfer not configured. Set HUMAN_AGENT_NUMBER in .env."}

    if function_name == "end_call":
        if outcome_state[0] == "unknown":
            outcome_state[0] = "completed" if is_browser else "not_interested"
        if task_ref[0]:
            asyncio.create_task(_delayed_end(task_ref[0]))
        return {"status": "ending"}

    return {"status": "error", "message": f"Unknown tool: {function_name}"}


# ==============================================================================
#  FASTAPI APP
# ==============================================================================
app = FastAPI(title="Tuvi Inbound Voice Agent", lifespan=lifespan)

_STATIC_DIR = Path(__file__).parent / "static"
if _STATIC_DIR.is_dir():
    app.mount("/static", StaticFiles(directory=str(_STATIC_DIR)), name="static")


# ── Health / readiness ─────────────────────────────────────────────────────────

def _is_placeholder_key(name: str, value: str) -> bool:
    """Detect unset or demo placeholder values in .env."""
    v = (value or "").strip().lower()
    if not v:
        return True
    if v in {"placeholder", "changeme", "todo", "your_key", "your_api_key", "sk-xxx"}:
        return True
    if v.startswith("your_") or v.endswith("_here"):
        return True
    if "xxxx" in v or "xxx" in v:
        return True
    if name == "TWILIO_ACCOUNT_SID" and v.startswith("ac") and "x" in v[2:]:
        return True
    if name.endswith("_API_KEY") and len(v) < 12:
        return True
    return False


async def _send_ws_error(websocket: WebSocket, message: str):
    try:
        await websocket.send_text(json.dumps({"event": "error", "message": message}))
    except Exception:
        pass


@app.get("/healthz")
async def healthz():
    return {"status": "ok"}


@app.get("/health")
async def health():
    return {"status": "ok"}


@app.get("/readyz")
async def readyz():
    """Readiness check — verifies required env vars are present and not placeholders."""
    required = ("TWILIO_ACCOUNT_SID", "TWILIO_AUTH_TOKEN", "TWILIO_PHONE_NUMBER",
                "DEEPGRAM_API_KEY", "OPENAI_API_KEY", "CARTESIA_API_KEY",
                "CARTESIA_VOICE_ID", "PUBLIC_BASE_URL")
    missing = [k for k in required if not os.environ.get(k)]
    invalid = [k for k in required if k not in missing and _is_placeholder_key(k, os.environ.get(k, ""))]
    problems = missing + invalid
    if problems:
        return JSONResponse({
            "status": "not_ready",
            "missing": problems,
            "message": "Missing or placeholder API keys. Update voice-sales-agent/.env",
        }, status_code=503)
    return {"status": "ready"}


@app.get("/readyz/browser")
async def readyz_browser():
    """Readiness for browser voice sessions (no Twilio required)."""
    voice_keys = ("DEEPGRAM_API_KEY", "OPENAI_API_KEY", "CARTESIA_API_KEY", "CARTESIA_VOICE_ID")
    missing = [k for k in voice_keys if not os.environ.get(k)]
    invalid = [k for k in voice_keys if k not in missing and _is_placeholder_key(k, os.environ.get(k, ""))]
    problems = missing + invalid
    if problems:
        return JSONResponse({
            "status": "not_ready",
            "missing": problems,
            "message": "Missing or invalid API keys for voice AI. Add real keys to voice-sales-agent/.env",
        }, status_code=503)
    return {"status": "ready"}


# ── Call log REST endpoints ────────────────────────────────────────────────────

@app.get("/calls")
async def get_calls(limit: int = 20, _: None = Depends(_require_call_api_key)):
    return JSONResponse(list_calls(limit))


@app.get("/calls/{call_id}")
async def get_call(call_id: int, _: None = Depends(_require_call_api_key)):
    summary = get_call_summary(call_id)
    if not summary:
        return JSONResponse({"error": "not found"}, status_code=404)
    return JSONResponse(summary)


@app.get("/transcripts")
async def get_transcripts(
    phone: str | None = None,
    email: str | None = None,
    call_id: int | None = None,
    limit: int = 500,
    _: None = Depends(_require_call_api_key),
):
    """
    List conversation transcripts (user + assistant), filterable by phone or email.

    Examples:
      GET /transcripts?phone=%2B61412345678
      GET /transcripts?email=jane%40example.com
      GET /transcripts?call_id=42
    """
    rows = list_conversation_transcripts(
        phone=phone,
        email=email,
        call_id=call_id,
        limit=limit,
    )
    return JSONResponse({"count": len(rows), "transcripts": rows})


# ── Admin endpoints ────────────────────────────────────────────────────────────

@app.post("/admin/call/{call_sid}/hangup")
async def admin_hangup(call_sid: str, _: None = Depends(_require_call_api_key)):
    """Manually terminate an active call pipeline."""
    task = _active_tasks.get(call_sid)
    if not task:
        return JSONResponse({"error": "call not found or already ended"}, status_code=404)
    await task.queue_frames([EndFrame()])
    logger.warning(f"Admin hangup triggered for {call_sid}")
    return {"status": "hangup_queued", "call_sid": call_sid}


@app.post("/admin/campaign/{campaign_id}/pause")
async def admin_pause_campaign(campaign_id: str, _: None = Depends(_require_call_api_key)):
    _paused_campaigns.add(campaign_id)
    logger.warning(f"Campaign {campaign_id} paused")
    return {"status": "paused", "campaign_id": campaign_id}


@app.post("/admin/campaign/{campaign_id}/resume")
async def admin_resume_campaign(campaign_id: str, _: None = Depends(_require_call_api_key)):
    _paused_campaigns.discard(campaign_id)
    return {"status": "resumed", "campaign_id": campaign_id}


# ── Disabled outbound call trigger ─────────────────────────────────────────────

@app.post("/call")
async def initiate_call():
    """Retain a stable denial response without exposing a dialing path."""
    return JSONResponse(
        {
            "status": "disabled",
            "message": "Outbound AI calls are disabled for this release.",
        },
        status_code=403,
    )


# ── TwiML webhook (inbound / fallback) ────────────────────────────────────────

@app.post("/twiml")
async def twiml_webhook(
    request: Request,
    _: None = Depends(_require_twilio_signature),
):
    public_url = os.environ["PUBLIC_BASE_URL"].rstrip("/")
    ws_host = public_url.replace("https://", "").replace("http://", "")
    form = await request.form()
    # Inbound caller identity for transcript storage
    caller_raw = (form.get("From") or "").strip()
    caller_phone = "".join(c for c in caller_raw if c.isdigit() or c == "+")
    # Inbound website number → corporate AI assistant
    twiml = f"""<?xml version="1.0" encoding="UTF-8"?>
<Response>
  <Connect>
    <Stream url="wss://{ws_host}/stream">
      <Parameter name="agent" value="corporate"/>
      <Parameter name="caller_phone" value="{caller_phone}"/>
    </Stream>
  </Connect>
</Response>"""
    return PlainTextResponse(content=twiml, media_type="application/xml")


# ── Twilio call status callbacks ───────────────────────────────────────────────

@app.post("/twilio/status")
async def twilio_status(
    request: Request,
    _: None = Depends(_require_twilio_signature),
):
    """
    Receives Twilio call status callbacks.
    Only records unanswered-call outcomes (no-answer, busy, failed).
    Connected-call outcomes are handled by the WebSocket pipeline.
    """
    form = await request.form()
    call_sid = form.get("CallSid", "")
    call_status = form.get("CallStatus", "")
    logger.info(f"Twilio status callback: SID={call_sid}  status={call_status}")

    if call_status in ("no-answer", "busy", "failed", "canceled"):
        db_id = _call_db_ids.get(call_sid)
        if db_id is not None:
            end_call(db_id, call_status, duration_s=None)
            _call_db_ids.pop(call_sid, None)

    return PlainTextResponse("", media_type="text/plain")


# ── Warm transfer TwiML ────────────────────────────────────────────────────────

@app.post("/transfer-twiml")
async def transfer_twiml(request: Request):
    """Returns TwiML that dials the human agent after a warm transfer."""
    to = request.query_params.get("to", os.environ.get("HUMAN_AGENT_NUMBER", ""))
    if not to:
        return PlainTextResponse(
            '<?xml version="1.0" encoding="UTF-8"?><Response><Say>Transfer unavailable.</Say><Hangup/></Response>',
            media_type="application/xml",
        )
    twiml = f"""<?xml version="1.0" encoding="UTF-8"?>
<Response>
  <Say>Please hold while I connect you with a member of our team.</Say>
  <Dial>{to}</Dial>
</Response>"""
    return PlainTextResponse(content=twiml, media_type="application/xml")


# ── WebSocket: Twilio Media Stream ─────────────────────────────────────────────

@app.websocket("/stream")
async def stream_websocket(websocket: WebSocket):
    """
    Twilio connects here when the prospect picks up.
    Full Pipecat pipeline boots per call.

    Pipeline:
      TwilioTransport.input()
        → Deepgram STT
        → TranscriptLogger(user)
        → OptOutGuardrail
        → LLMContextAggregator (user side)
        → OpenAI LLM
        → TranscriptLogger(assistant)
        → Cartesia TTS
        → TwilioTransport.output()
        → LLMContextAggregator (assistant side)
    """
    await websocket.accept()
    call_sid = "unknown"
    stream_sid = "unknown"
    call_db_id = -1
    agent_mode = "corporate"
    custom_params: dict[str, str] = {}

    logger.info("Twilio WebSocket connected — reading call SID...")

    # Peek at the first two Twilio envelope messages (connected + start events)
    # to extract call SID before handing the socket to Pipecat.
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
            agent_mode = _normalize_agent_mode(
                custom_params.get("agent") or _call_agent_modes.get(call_sid),
                default="corporate",
            )
            logger.info(f"Linked to call SID: {call_sid}  agent={agent_mode}")
            caller_phone = (
                custom_params.get("caller_phone")
                or custom_params.get("from")
                or start_payload.get("from")
                or start_payload.get("to")
                or ""
            ).strip()
            if call_sid in _call_db_ids:
                call_db_id = _call_db_ids[call_sid]
                # Outbound already stored dialled number; fill if missing
                if caller_phone and not get_call_number(call_db_id):
                    update_call_contact(call_db_id, phone=caller_phone)
            else:
                # Inbound / unregistered call — store caller phone for transcripts
                call_db_id = start_call(
                    call_sid,
                    caller_phone or "unknown",
                    channel="phone",
                    agent_mode=agent_mode,
                )
                _call_db_ids[call_sid] = call_db_id
            _call_agent_modes[call_sid] = agent_mode
    except Exception as e:
        logger.warning(f"Could not read call SID from Twilio envelope: {e}")
        call_db_id = start_call("unknown", "unknown", channel="phone", agent_mode=agent_mode)

    # Resolve prospect's phone number (needed for SMS and opt-out tools)
    to_number_ref: list[str] = [get_call_number(call_db_id) or "unknown"]

    # ── Transport ─────────────────────────────────────────────────────────────
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

    # ── STT ───────────────────────────────────────────────────────────────────
    stt = DeepgramSTTService(api_key=os.environ["DEEPGRAM_API_KEY"])

    restaurant_id: str | None = None
    if agent_mode == "restaurant":
        try:
            restaurant_index = int(custom_params.get("restaurant_index", "0"))
        except ValueError:
            restaurant_index = 0
        site = await api_client.get_site_restaurant(restaurant_index)
        if site.get("status") == "error":
            site = {"name": "the restaurant"}
        restaurant_id = str(site.get("restaurant_id") or "")
        phone_tools = RESTAURANT_TOOLS
        system_prompt = build_restaurant_prompt(site, channel="phone")
        greeting_text = build_restaurant_greeting(site)
    else:
        phone_tools = CORPORATE_PHONE_TOOLS
        system_prompt = build_corporate_prompt(channel="phone")
        greeting_text = build_corporate_greeting()

    # ── LLM ───────────────────────────────────────────────────────────────────
    llm = OpenAILLMService(
        api_key=os.environ["OPENAI_API_KEY"],
        model=os.environ.get("OPENAI_MODEL", "gpt-4o-mini"),
        tools=phone_tools,
    )

    # Mutable closure state
    booking_state: dict = {}
    outcome_state: list[str] = ["unknown"]
    task_ref: list = [None]   # populated after Pipeline/PipelineTask are created

    async def handle_tool_call(function_name, tool_call_id, arguments, llm, context, result_callback):
        result = await _dispatch_tool(
            function_name=function_name,
            arguments=arguments,
            call_db_id=call_db_id,
            to_number_ref=to_number_ref,
            outcome_state=outcome_state,
            task_ref=task_ref,
            booking_state=booking_state,
            call_sid=call_sid,
            is_browser=False,
            restaurant_id=restaurant_id,
        )
        await result_callback(json.dumps(result))

    llm.register_function(None, handle_tool_call)

    # ── LLM Context ───────────────────────────────────────────────────────────
    context = LLMContext(messages=[{"role": "system", "content": system_prompt}])
    user_aggregator, assistant_aggregator = LLMContextAggregatorPair(context)

    # ── TTS ───────────────────────────────────────────────────────────────────
    # Language + voice from .env (default Indian English receptionist).
    _lang = _cartesia_pipecat_language()
    tts = CartesiaTTSService(
        api_key=os.environ["CARTESIA_API_KEY"],
        voice_id=os.environ["CARTESIA_VOICE_ID"],
        settings=CartesiaTTSSettings(
            language=_lang,
            generation_config=GenerationConfig(
                speed=1.05,         # Cartesia range: 0.6–1.5, 1.0=default; 1.05=slightly snappier
                emotion="positivity:medium",  # warm, upbeat but not fake
            ),
        ),
    )

    # ── Processors ────────────────────────────────────────────────────────────
    # Shared timing state: records user-turn-end timestamp, cleared on first assistant text
    turn_timing: dict = {"user_turn_ts": 0.0, "first_text_pending": False}

    user_logger      = TranscriptLogger(call_db_id, "user", timing_state=turn_timing)
    assistant_logger = TranscriptLogger(call_db_id, "assistant", timing_state=turn_timing)
    opt_out_guardrail = OptOutGuardrail(call_db_id, to_number_ref, outcome_state, task_ref)
    context_compactor = ContextCompactor(context, max_pairs=int(os.environ.get("LLM_CONTEXT_PAIRS", "5")))

    # ── Pipeline ──────────────────────────────────────────────────────────────
    #  user_logger → records turn timestamp
    #  opt_out_guardrail → hard opt-out detection
    #  context_compactor → trims context to last N pairs before LLM receives new turn
    #  assistant_logger → measures first-text latency
    pipeline = Pipeline([
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
    ])

    task = PipelineTask(
        pipeline,
        params=PipelineParams(
            allow_interruptions=True,
            enable_metrics=True,
        ),
    )
    task_ref[0] = task   # make available to tool handlers and guardrail
    _active_tasks[call_sid] = task

    # Track answer time for duration_s calculation
    call_start_time: list[float] = [0.0]
    _flushed: list[bool] = [False]
    max_duration_task: asyncio.Task | None = None

    def _do_flush():
        if _flushed[0]:
            return
        _flushed[0] = True
        duration = round(time.time() - call_start_time[0], 1) if call_start_time[0] else None
        _flush_call_log(call_sid, call_db_id, outcome_state[0], booking_state or None, duration)

    def _cancel_max_duration():
        nonlocal max_duration_task
        if max_duration_task and not max_duration_task.done():
            max_duration_task.cancel()
        max_duration_task = None

    # ── Events ────────────────────────────────────────────────────────────────

    @transport.event_handler("on_client_connected")
    async def on_connected(transport, client):
        nonlocal max_duration_task
        call_start_time[0] = time.time()
        logger.info(
            "Prospect answered — speaking greeting (agent=%s): %s",
            agent_mode,
            greeting_text[:80],
        )
        # Speak immediately via TTS (do not wait for LLM). Pipeline then listens.
        log_turn(call_db_id, "assistant", greeting_text)
        await task.queue_frames([
            TTSSpeakFrame(text=greeting_text, append_to_context=True),
        ])

        max_minutes = _phone_call_max_minutes()
        if max_minutes is not None:
            logger.info(
                "Phone call max duration: %.1f minute(s) (PHONE_CALL_MAX_MINUTES)",
                max_minutes,
            )
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

    # ── Run ───────────────────────────────────────────────────────────────────
    runner = PipelineRunner()
    try:
        await runner.run(task)
    finally:
        _cancel_max_duration()

    _do_flush()  # no-op if already flushed in on_disconnected

    _active_tasks.pop(call_sid, None)
    _call_db_ids.pop(call_sid, None)
    _call_agent_modes.pop(call_sid, None)
    logger.info(f"Pipeline done. DB id={call_db_id}  outcome={outcome_state[0]}")


# ── Static UI ─────────────────────────────────────────────────────────────────

@app.get("/")
async def serve_index():
    index = _STATIC_DIR / "index.html"
    if not index.is_file():
        return JSONResponse({"error": "UI not found. static/index.html is missing."}, status_code=404)
    return FileResponse(str(index))


# ── WebSocket: Browser Direct Voice ───────────────────────────────────────────

@app.websocket("/browser-stream")
async def browser_stream(websocket: WebSocket):
    """
    Browser clients connect here for a direct voice session without a phone.

    Query params:
      restaurant_index — site index for MonoRepo public API (default 0)
      agent — "restaurant" (default) or "corporate" for Tuvi website assistant

    Audio protocol:
      Browser → Server: binary PCM16 @ 16kHz mono
      Server → Browser: binary PCM16 @ 16kHz mono  (TTS output)
      Server → Browser: text JSON  {"event": ..., ...}  (transcripts, status)
    """
    await websocket.accept()

    query = websocket.scope.get("query_string", b"").decode()
    params = dict(parse_qsl(query))
    agent_mode = _normalize_agent_mode(params.get("agent"), default="restaurant")
    restaurant_index = 0

    if agent_mode == "corporate":
        site = {"name": "Tuvi Solutions"}
        restaurant_id = ""
        system_prompt = build_corporate_prompt(channel="browser")
        greeting_text = build_corporate_greeting()
        agent_tools = CORPORATE_TOOLS
    else:
        try:
            restaurant_index = int(params.get("restaurant_index", "0"))
        except ValueError:
            restaurant_index = 0

        site = await api_client.get_site_restaurant(restaurant_index)
        if site.get("status") == "error":
            await _send_ws_error(websocket, site.get("message", "Restaurant unavailable."))
            await websocket.close()
            return

        restaurant_id = str(site.get("restaurant_id") or "")
        system_prompt = build_restaurant_prompt(site)
        greeting_text = build_restaurant_greeting(site)
        agent_tools = RESTAURANT_BROWSER_TOOLS

    session_id = f"browser-{uuid.uuid4().hex[:8]}"
    contact_name = str(site.get("name") or "").strip()
    call_db_id = start_call(
        session_id,
        "",
        channel="browser",
        agent_mode=agent_mode,
        contact_name=contact_name,
        restaurant_index=restaurant_index if agent_mode == "restaurant" else None,
    )

    logger.info(
        f"Browser session started: {session_id}  DB id={call_db_id}  "
        f"agent={agent_mode}  restaurant={contact_name or '—'}"
    )

    # Tell the browser immediately so it can start mic streaming before greeting audio.
    try:
        await websocket.send_text(
            json.dumps({
                "event": "ready",
                "session_id": session_id,
                "call_db_id": call_db_id,
                "restaurant": contact_name or site.get("name"),
            })
        )
    except Exception:
        pass

    to_number_ref: list[str] = ["browser"]
    booking_state: dict = {}
    outcome_state: list[str] = ["unknown"]
    task_ref: list = [None]
    # Shared mailbox: browser_serializer puts typed contact-form futures here.
    client_mailbox: dict = {}

    # ── Transport ─────────────────────────────────────────────────────────────
    _lang = _cartesia_pipecat_language()

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

    # ── Services ──────────────────────────────────────────────────────────────
    stt = DeepgramSTTService(api_key=os.environ["DEEPGRAM_API_KEY"])
    llm = OpenAILLMService(
        api_key=os.environ["OPENAI_API_KEY"],
        model=os.environ.get("OPENAI_MODEL", "gpt-4o-mini"),
        tools=agent_tools,
    )
    tts = CartesiaTTSService(
        api_key=os.environ["CARTESIA_API_KEY"],
        voice_id=os.environ["CARTESIA_VOICE_ID"],
        sample_rate=16000,
        settings=CartesiaTTSSettings(
            language=_lang,
            generation_config=GenerationConfig(
                speed=1.05,
                emotion="positivity:medium",
            ),
        ),
    )

    # ── Tool handler ──────────────────────────────────────────────────────────
    async def handle_browser_tool(function_name, tool_call_id, arguments, llm, context, result_callback):
        check_tools = (
            "check_consultation_slot",
            "check_consultation_slots",
            "check_table_availability",
        )

        if function_name == "request_typed_email":
            result = await _wait_for_typed_email(
                websocket,
                client_mailbox,
                call_db_id=call_db_id,
                prompt=(arguments.get("prompt") or "").strip(),
            )
            await result_callback(json.dumps(result))
            return

        if function_name == "request_booking_details":
            result = await _wait_for_booking_details(
                websocket,
                client_mailbox,
                call_db_id=call_db_id,
                prompt=(arguments.get("prompt") or "").strip(),
            )
            await result_callback(json.dumps(result))
            return

        # Every browser consultation must use the on-screen contact-details form.
        if function_name == "book_consultation":
            details = client_mailbox.pop("booking_details", None)
            if details is None:
                details = await _wait_for_booking_details(
                    websocket,
                    client_mailbox,
                    call_db_id=call_db_id,
                    prompt="Enter your name, email, and phone to confirm the consultation.",
                )
                if details.get("status") != "success":
                    await result_callback(json.dumps(details))
                    return
            arguments = {
                **arguments,
                "prospect_name": details["prospect_name"],
                "prospect_email": details["prospect_email"],
                "prospect_phone": details["prospect_phone"],
            }
            client_mailbox.pop("booking_details", None)

        if function_name in check_tools:
            await _emit_booking_progress(websocket, "checking_slots", "Checking slots…")
        elif function_name in ("book_consultation", "book_table_reservation"):
            message = (
                "Submitting request…"
                if function_name == "book_table_reservation"
                else "Booking slot…"
            )
            await _emit_booking_progress(websocket, "booking_slot", message)

        result = await _dispatch_tool(
            function_name=function_name,
            arguments=arguments,
            call_db_id=call_db_id,
            to_number_ref=to_number_ref,
            outcome_state=outcome_state,
            task_ref=task_ref,
            booking_state=booking_state,
            call_sid=session_id,
            is_browser=True,
            restaurant_id=restaurant_id,
            restaurant_index=restaurant_index,
        )

        if function_name in check_tools:
            await _emit_booking_progress(websocket, "idle", "")
        elif function_name in ("book_consultation", "book_table_reservation"):
            if result.get("status") == "success":
                await _emit_booking_progress(websocket, "success", "Slot booked successfully!")
            elif result.get("status") == "pending":
                await _emit_booking_progress(websocket, "success", "Reservation request received.")
            else:
                await _emit_booking_progress(websocket, "idle", "")

        if (
            function_name == "book_table_reservation"
            and result.get("status") == "pending"
            and result.get("reservation_id")
        ):
            await _emit_browser_booking_event(
                websocket,
                reservation_id=result.get("reservation_id", ""),
                status=result.get("status", ""),
                guest_name=result.get("guest_name") or arguments.get("guest_name", "Guest"),
                guest_phone=arguments.get("guest_phone", ""),
                party_size=int(result.get("party_size") or arguments.get("party_size") or 2),
                slot=result.get("slot") or arguments.get("slot", ""),
                booking_date=result.get("booking_date") or arguments.get("date", ""),
                booking_time=result.get("booking_time") or arguments.get("time", ""),
                message=result.get("message", ""),
            )
        if function_name == "book_consultation" and result.get("status") == "success":
            await _emit_browser_consultation_event(
                websocket,
                confirmation_code=result.get("confirmation_code", ""),
                prospect_name=result.get("prospect_name") or arguments.get("prospect_name", "Guest"),
                prospect_email=result.get("prospect_email") or arguments.get("prospect_email", ""),
                slot=result.get("slot") or "",
                booking_date=result.get("booking_date") or arguments.get("date", ""),
                booking_time=result.get("booking_time") or arguments.get("time", ""),
                calendly_link=result.get("calendly_link", "") or result.get("calendar_link", ""),
                calendar_link=result.get("calendar_link", "") or result.get("calendly_link", ""),
                message=result.get("message", ""),
            )
        await result_callback(json.dumps(result))

    llm.register_function(None, handle_browser_tool)

    # ── LLM Context ───────────────────────────────────────────────────────────
    context = LLMContext(messages=[{"role": "system", "content": system_prompt}])
    user_aggregator, assistant_aggregator = LLMContextAggregatorPair(context)

    # ── Processors ────────────────────────────────────────────────────────────
    turn_timing: dict = {"user_turn_ts": 0.0, "first_text_pending": False}
    max_pairs = int(os.environ.get("LLM_CONTEXT_PAIRS", "5"))

    user_logger      = TranscriptLogger(call_db_id, "user",      timing_state=turn_timing, event_ws=websocket)
    assistant_logger = TranscriptLogger(call_db_id, "assistant", timing_state=turn_timing, event_ws=websocket)
    opt_out_guardrail = OptOutGuardrail(call_db_id, to_number_ref, outcome_state, task_ref)
    context_compactor = ContextCompactor(context, max_pairs=max_pairs)
    status_user = StatusEventSender(websocket)   # catches UserStarted/StoppedSpeakingFrame
    status_bot  = StatusEventSender(websocket)   # catches BotStarted/StoppedSpeakingFrame

    # ── Pipeline ──────────────────────────────────────────────────────────────
    pipeline = Pipeline([
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
    ])

    task = PipelineTask(
        pipeline,
        params=PipelineParams(allow_interruptions=True, enable_metrics=True),
    )
    task_ref[0] = task

    call_start_time: list[float] = [0.0]
    _flushed: list[bool] = [False]

    def _do_flush():
        if _flushed[0]:
            return
        _flushed[0] = True
        duration = round(time.time() - call_start_time[0], 1) if call_start_time[0] else None
        _flush_call_log(session_id, call_db_id, outcome_state[0], booking_state or None, duration)

    # ── Events ────────────────────────────────────────────────────────────────
    @transport.event_handler("on_client_connected")
    async def on_browser_connected(transport, client):
        call_start_time[0] = time.time()
        logger.info(f"Browser session live: {session_id}  restaurant={site.get('name')}")
        # Greeting already streamed before pipeline start — wait for guest to speak.
        if not greeting_streamed:
            await task.queue_frames([LLMMessagesFrame(context.messages)])

    @transport.event_handler("on_client_disconnected")
    async def on_browser_disconnected(transport, client):
        logger.info(f"Browser session ended: {session_id}")
        if outcome_state[0] == "unknown":
            outcome_state[0] = "completed"
        _do_flush()
        await task.queue_frames([EndFrame()])

    # Stream restaurant greeting audio BEFORE pipeline (instant playback, Indian voice).
    greeting_streamed = await _stream_browser_greeting(
        websocket,
        call_db_id=call_db_id,
        context=context,
        greeting_text=greeting_text,
    )
    if not greeting_streamed:
        try:
            log_turn(call_db_id, "assistant", greeting_text)
            context.messages.append({"role": "assistant", "content": greeting_text})
            await websocket.send_text(json.dumps({"event": "transcript", "role": "assistant", "text": greeting_text}))
        except Exception as e:
            logger.warning(f"Greeting transcript fallback failed: {e}")

    runner = PipelineRunner()
    try:
        await runner.run(task)
    except Exception as e:
        logger.error(f"Browser pipeline error: {e}")
        await _send_ws_error(websocket, f"Voice session failed: {e}")
        raise

    _do_flush()
    logger.info(f"Browser pipeline done. DB id={call_db_id}  outcome={outcome_state[0]}")


# ==============================================================================
#  HELPERS
# ==============================================================================

def _flush_call_log(
    call_sid: str,
    call_db_id: int,
    outcome: str,
    booking: dict | None,
    duration_s: float | None = None,
):
    if call_db_id < 0:
        return
    try:
        end_call(call_db_id, outcome, booking, duration_s)
        logger.info(f"Call log saved → DB id={call_db_id}  outcome={outcome}  duration={duration_s}s")
    except Exception as e:
        logger.error(f"Failed to flush call log: {e}")


async def _delayed_end(task: PipelineTask, delay: float = 3.0):
    """Let TTS finish saying goodbye, then end the pipeline."""
    await asyncio.sleep(delay)
    await task.queue_frames([EndFrame()])


def _phone_call_max_minutes() -> float | None:
    """
    Max talk time for phone calls, in minutes.
    Env: PHONE_CALL_MAX_MINUTES (or TIME as alias). 2 = 2 minutes, 3 = 3 minutes.
    0 / unset / invalid = no auto hangup.
    """
    raw = (
        os.environ.get("PHONE_CALL_MAX_MINUTES")
        or os.environ.get("TIME")
        or ""
    ).strip()
    if not raw:
        return None
    try:
        minutes = float(raw)
    except ValueError:
        logger.warning("Invalid PHONE_CALL_MAX_MINUTES=%r — ignoring", raw)
        return None
    if minutes <= 0:
        return None
    return minutes


def _hangup_twilio_call(call_sid: str) -> None:
    """Force-end the Twilio call leg so the user phone hangs up."""
    if not call_sid or call_sid == "unknown":
        return
    sid = os.environ.get("TWILIO_ACCOUNT_SID", "")
    token = os.environ.get("TWILIO_AUTH_TOKEN", "")
    if not (sid and token):
        return
    try:
        TwilioClient(sid, token).calls(call_sid).update(status="completed")
        logger.info("Twilio hangup sent for %s", call_sid)
    except Exception as exc:
        logger.warning("Twilio hangup failed for %s: %s", call_sid, exc)


async def _max_duration_watchdog(
    *,
    minutes: float,
    call_sid: str,
    call_db_id: int,
    outcome_state: list[str],
    task_ref: list,
    flushed: list[bool],
) -> None:
    """After `minutes`, speak a short wrap-up and auto-cut the phone call."""
    try:
        await asyncio.sleep(minutes * 60.0)
    except asyncio.CancelledError:
        return

    if flushed[0]:
        return

    logger.warning(
        "Phone call max duration reached (%.1f min) — auto hangup SID=%s",
        minutes,
        call_sid,
    )
    outcome_state[0] = "max_duration"
    wrap_up = (
        "Thanks so much for contacting Tuvi Solutions. "
        "It was a pleasure speaking with you. "
        "If you'd like to continue, call us back anytime or book a free consultation on our website. "
        "Take care, and goodbye."
    )
    try:
        log_turn(call_db_id, "assistant", wrap_up)
    except Exception:
        pass

    task = task_ref[0]
    if task is not None:
        try:
            await task.queue_frames([TTSSpeakFrame(text=wrap_up, append_to_context=True)])
        except Exception as exc:
            logger.warning("Max-duration wrap-up TTS failed: %s", exc)
        # Let closing message finish speaking before hangup
        await asyncio.sleep(10.0)
        try:
            await task.queue_frames([EndFrame()])
        except Exception:
            pass

    _hangup_twilio_call(call_sid)


def _mock_available_slots() -> list[str]:
    """Return 3 business-day slots starting from tomorrow (Australia/Sydney)."""
    slots = []
    base = datetime.now(SYDNEY_TZ) + timedelta(days=1)
    for _ in range(3):
        while base.weekday() >= 5:   # skip weekends
            base += timedelta(days=1)
        slots.append(base.replace(hour=14, minute=0, second=0, microsecond=0).isoformat())
        base += timedelta(days=1)
    return slots


# ==============================================================================
#  ENTRY POINT
# ==============================================================================
if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8000))
    logger.info(
        f"\n{'='*60}\n"
        f"  Tuvi Inbound Voice Agent\n"
        f"\n"
        f"  NEXT STEPS:\n"
        f"  1. Expose:     ngrok http {port}\n"
        f"  2. Set PUBLIC_BASE_URL in .env to your ngrok URL\n"
        f"  3. Configure:  Twilio inbound voice webhook → POST /twiml\n"
        f"  4. View logs:  GET http://localhost:{port}/calls\n"
        f"  5. Terminal:   python show_calls.py\n"
        f"  6. Readiness:  GET http://localhost:{port}/readyz\n"
        f"{'='*60}\n"
    )
    uvicorn.run(app, host="0.0.0.0", port=port)
