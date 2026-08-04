"""
Gradio test UI for Andre Voice Agent.

Talk via browser mic (WebSocket → /browser-stream) or place an outbound call.

Run:
  uvicorn bot:app --port 8001
  python ui.py
"""

from __future__ import annotations

import asyncio
import json
import os
import struct
from typing import Any

import gradio as gr
import numpy as np
import requests
import websockets
from dotenv import load_dotenv

load_dotenv()

AGENT_URL = os.environ.get("ANDRE_AGENT_URL", f"http://127.0.0.1:{os.environ.get('PORT', '8001')}").rstrip("/")
CALL_API_SECRET = (os.environ.get("CALL_API_SECRET") or "").strip()
UI_PORT = int(os.environ.get("GRADIO_PORT", "7860"))

LANG_CHOICES = [
    ("Auto (mirror caller)", "auto"),
    ("English", "en"),
    ("Hindi", "hi"),
    ("Telugu", "te"),
]


def _ws_base() -> str:
    if AGENT_URL.startswith("https://"):
        return "wss://" + AGENT_URL[len("https://") :]
    if AGENT_URL.startswith("http://"):
        return "ws://" + AGENT_URL[len("http://") :]
    return f"ws://{AGENT_URL}"


def check_health() -> str:
    try:
        r = requests.get(f"{AGENT_URL}/healthz", timeout=5)
        ready = requests.get(f"{AGENT_URL}/readyz/browser", timeout=5)
        return (
            f"Agent: {AGENT_URL}\n"
            f"/healthz → {r.status_code} {r.text}\n"
            f"/readyz/browser → {ready.status_code} {ready.text}"
        )
    except Exception as e:
        return f"Agent unreachable at {AGENT_URL}: {e}"


def place_call(phone: str, language: str) -> str:
    phone = (phone or "").strip()
    if not phone:
        return "Enter a phone number with country code (e.g. +919876543210)."
    headers = {"Content-Type": "application/json"}
    if CALL_API_SECRET:
        headers["X-Call-Api-Key"] = CALL_API_SECRET
    try:
        r = requests.post(
            f"{AGENT_URL}/call",
            headers=headers,
            json={
                "to": phone,
                "language": language,
                "agent": "andre",
                "campaign_id": "gradio",
                "skip_compliance": True,
            },
            timeout=30,
        )
        try:
            data = r.json()
        except Exception:
            return f"HTTP {r.status_code}: {r.text}"
        return json.dumps(data, indent=2)
    except Exception as e:
        return f"Call failed: {e}"


def _float_to_pcm16(audio: np.ndarray) -> bytes:
    audio = np.asarray(audio, dtype=np.float32)
    if audio.ndim > 1:
        audio = audio.mean(axis=1)
    audio = np.clip(audio, -1.0, 1.0)
    return (audio * 32767.0).astype(np.int16).tobytes()


def _pcm16_to_float(pcm: bytes) -> np.ndarray:
    if not pcm:
        return np.zeros(0, dtype=np.float32)
    count = len(pcm) // 2
    samples = struct.unpack(f"<{count}h", pcm[: count * 2])
    return (np.asarray(samples, dtype=np.float32) / 32767.0)


async def _talk_once(audio: tuple[int, np.ndarray] | None, language: str, transcript: str) -> tuple[str, Any]:
    """
    One-shot turn: send mic chunk to Andre browser stream, collect reply audio + transcripts.
    Gradio streaming mic sessions are flaky across versions; this keeps testing simple.
    """
    if audio is None:
        return transcript or "", None

    sample_rate, data = audio
    if data is None or len(data) == 0:
        return transcript or "No audio captured.", None

    # Resample to 16k if needed (simple decimate/repeat)
    arr = np.asarray(data, dtype=np.float32)
    if arr.ndim > 1:
        arr = arr.mean(axis=1)
    if arr.dtype == np.int16:
        arr = arr.astype(np.float32) / 32767.0
    elif np.max(np.abs(arr)) > 1.5:
        arr = arr / 32767.0

    if sample_rate != 16000 and sample_rate > 0:
        duration = len(arr) / float(sample_rate)
        target_len = max(1, int(duration * 16000))
        x_old = np.linspace(0.0, 1.0, num=len(arr), endpoint=False)
        x_new = np.linspace(0.0, 1.0, num=target_len, endpoint=False)
        arr = np.interp(x_new, x_old, arr).astype(np.float32)

    pcm = _float_to_pcm16(arr)
    ws_url = f"{_ws_base()}/browser-stream?language={language}"
    lines = list(filter(None, (transcript or "").splitlines()))
    reply_chunks: list[bytes] = []

    try:
        async with websockets.connect(ws_url, max_size=8 * 1024 * 1024) as ws:
            # Wait for ready
            ready_deadline = asyncio.get_event_loop().time() + 8
            while asyncio.get_event_loop().time() < ready_deadline:
                msg = await asyncio.wait_for(ws.recv(), timeout=8)
                if isinstance(msg, str):
                    evt = json.loads(msg)
                    if evt.get("event") == "ready":
                        break
                    if evt.get("event") == "transcript":
                        role = evt.get("role", "?")
                        text = evt.get("text", "")
                        lines.append(f"{role}: {text}")
                    if evt.get("event") == "error":
                        lines.append(f"error: {evt.get('message')}")
                        return "\n".join(lines), None
                elif isinstance(msg, (bytes, bytearray)):
                    reply_chunks.append(bytes(msg))

            # Send user audio in chunks, then a short silence, then stop
            chunk = 3200
            for i in range(0, len(pcm), chunk):
                await ws.send(pcm[i : i + chunk])
                await asyncio.sleep(0.02)
            # pad silence so VAD can end the turn
            silence = b"\x00" * 16000  # 0.5s
            await ws.send(silence)
            await asyncio.sleep(0.6)

            # Collect replies for a few seconds
            end_at = asyncio.get_event_loop().time() + 18
            while asyncio.get_event_loop().time() < end_at:
                try:
                    msg = await asyncio.wait_for(ws.recv(), timeout=2.5)
                except asyncio.TimeoutError:
                    if reply_chunks:
                        break
                    continue
                if isinstance(msg, (bytes, bytearray)):
                    reply_chunks.append(bytes(msg))
                elif isinstance(msg, str):
                    evt = json.loads(msg)
                    if evt.get("event") == "transcript":
                        role = evt.get("role", "?")
                        text = evt.get("text", "")
                        lines.append(f"{role}: {text}")
                    elif evt.get("event") == "error":
                        lines.append(f"error: {evt.get('message')}")
                        break
                    elif evt.get("event") == "status" and evt.get("state") == "listening" and reply_chunks:
                        # bot finished speaking after our turn
                        await asyncio.sleep(0.3)
                        break

            await ws.send(json.dumps({"event": "stop"}))
    except Exception as e:
        lines.append(f"error: {e}")
        return "\n".join(lines), None

    out_text = "\n".join(lines) if lines else "(no transcript yet)"
    if not reply_chunks:
        return out_text, None

    pcm_out = b"".join(reply_chunks)
    float_out = _pcm16_to_float(pcm_out)
    return out_text, (16000, float_out)


def talk(audio, language, transcript):
    return asyncio.run(_talk_once(audio, language, transcript or ""))


def build_ui() -> gr.Blocks:
    with gr.Blocks(title="Ananya Voice Agent") as demo:
        gr.Markdown(
            """
            # Ananya Voice Agent
            Sarvam STT/TTS · Hindi / English / Telugu · Twilio calling

            Start the agent first: `uvicorn bot:app --host 0.0.0.0 --port 8001`
            """
        )
        health_box = gr.Textbox(label="Health", lines=4, value=check_health())
        refresh = gr.Button("Refresh health")
        refresh.click(fn=check_health, outputs=health_box)

        with gr.Tab("Talk (mic)"):
            lang = gr.Dropdown(
                choices=LANG_CHOICES,
                value="en",
                label="Language (locks reply + girl voice)",
            )
            mic = gr.Audio(sources=["microphone"], type="numpy", label="Speak to Ananya")
            transcript = gr.Textbox(label="Transcript", lines=12)
            reply = gr.Audio(label="Ananya reply", type="numpy")
            go = gr.Button("Send to Ananya", variant="primary")
            go.click(fn=talk, inputs=[mic, lang, transcript], outputs=[transcript, reply])

        with gr.Tab("Call me"):
            phone = gr.Textbox(label="Phone (E.164)", placeholder="+919876543210")
            call_lang = gr.Dropdown(
                choices=LANG_CHOICES,
                value="en",
                label="Language (locks reply + girl voice)",
            )
            call_btn = gr.Button("Place outbound call", variant="primary")
            call_out = gr.Textbox(label="Result", lines=8)
            call_btn.click(fn=place_call, inputs=[phone, call_lang], outputs=call_out)

        gr.Markdown(
            f"Agent API: `{AGENT_URL}` · Gradio port `{UI_PORT}` · "
            "For Twilio, set `PUBLIC_BASE_URL` to your ngrok HTTPS URL (see CALLING.md)."
        )
    return demo


if __name__ == "__main__":
    demo = build_ui()
    demo.launch(server_name="0.0.0.0", server_port=UI_PORT)
