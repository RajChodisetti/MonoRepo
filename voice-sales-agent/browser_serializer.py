"""
BrowserFrameSerializer — PCM16 serializer for browser WebSocket sessions.

Protocol:
  Browser → Server:
    binary  = raw PCM16 audio @ 16kHz mono  →  InputAudioRawFrame
    text    = {"event": "stop"}             →  EndFrame
    text    = {"event": "user_email", "email": "..."}  → mailbox (no frame)

  Server → Browser:
    binary  = raw PCM16 audio @ 16kHz mono  ←  OutputAudioRawFrame
    text    = {"event": ..., ...}           ←  sent directly via ws.send_text()
"""

from __future__ import annotations

import json
from typing import Any

from pipecat.frames.frames import AudioRawFrame, EndFrame, Frame, InputAudioRawFrame, OutputAudioRawFrame
from pipecat.serializers.base_serializer import FrameSerializer


class BrowserFrameSerializer(FrameSerializer):
    """
    Converts between raw PCM16 WebSocket binary frames and pipecat audio frames.
    Optional `mailbox` receives typed form events (e.g. user_email) during a session.
    """

    def __init__(self, mailbox: dict[str, Any] | None = None):
        super().__init__()
        self._mailbox = mailbox if mailbox is not None else {}

    async def serialize(self, frame: Frame) -> str | bytes | None:
        if isinstance(frame, (OutputAudioRawFrame, AudioRawFrame)):
            return frame.audio
        return None

    async def deserialize(self, data: str | bytes) -> Frame | None:
        if isinstance(data, bytes):
            if not data:
                return None
            return InputAudioRawFrame(audio=data, sample_rate=16000, num_channels=1)

        try:
            msg = json.loads(data)
        except Exception:
            return None

        event = msg.get("event")
        if event == "stop":
            return EndFrame()

        if event == "user_email":
            fut = self._mailbox.get("email_future")
            if fut is not None and not fut.done():
                fut.set_result((msg.get("email") or "").strip())
            return None

        return None
