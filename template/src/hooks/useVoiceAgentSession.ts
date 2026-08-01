"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { getVoiceAgentWsUrl } from "@/lib/voiceAgentConfig";

export type VoiceSessionStatus =
  | "idle"
  | "checking"
  | "error"
  | "connecting"
  | "listening"
  | "thinking"
  | "speaking"
  | "user-speaking";

export type ReservationRequestReceipt = {
  reservationId: string;
  status: string;
  guestName: string;
  guestPhone: string;
  partySize: number;
  slot: string;
  bookingDate: string;
  bookingTime: string;
  message: string;
};

type ReadinessResult =
  | { ok: true }
  | { ok: false; message: string; missing?: string[] };

const READY_TIMEOUT_MS = 12_000;

export async function checkVoiceAgentReadiness(): Promise<ReadinessResult> {
  try {
    const res = await fetch("/api/voice-agent/status", {
      cache: "no-store",
      signal: AbortSignal.timeout(5000),
    });
    const data = (await res.json()) as {
      status?: string;
      missing?: string[];
      message?: string;
    };

    if (res.ok && data.status === "ready") {
      return { ok: true };
    }

    if (data.missing?.length) {
      return {
        ok: false,
        missing: data.missing,
        message:
          data.message ||
          `Missing or invalid API keys: ${data.missing.join(", ")}. Update voice-sales-agent/.env`,
      };
    }

    if (data.status === "unavailable") {
      return {
        ok: false,
        message:
          data.message ||
          "Voice agent server is not running. Start it from the voice-sales-agent folder.",
      };
    }

    return {
      ok: false,
      message: data.message || "Voice assistant is not ready. Check voice-sales-agent/.env",
    };
  } catch {
    return {
      ok: false,
      message: "Could not reach voice assistant. Is the server running on port 8000?",
    };
  }
}

function waitForWsOpen(ws: WebSocket, timeoutMs: number): Promise<void> {
  return new Promise((resolve, reject) => {
    if (ws.readyState === WebSocket.OPEN) {
      resolve();
      return;
    }

    const timer = setTimeout(() => {
      cleanup();
      reject(new Error("Connection timed out. Voice agent server may be down."));
    }, timeoutMs);

    const onOpen = () => {
      cleanup();
      resolve();
    };
    const onError = () => {
      cleanup();
      reject(new Error("WebSocket connection failed. Is the voice agent server running?"));
    };

    const cleanup = () => {
      clearTimeout(timer);
      ws.removeEventListener("open", onOpen);
      ws.removeEventListener("error", onError);
    };

    ws.addEventListener("open", onOpen);
    ws.addEventListener("error", onError);
  });
}

export function useVoiceAgentSession(restaurantIndex = 0) {
  const [status, setStatus] = useState<VoiceSessionStatus>("idle");
  const [error, setError] = useState<string | null>(null);
  const [active, setActive] = useState(false);
  const [booking, setBooking] = useState<ReservationRequestReceipt | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const audioCtxRef = useRef<AudioContext | null>(null);
  const workletRef = useRef<AudioWorkletNode | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const nextPlayTimeRef = useRef(0);
  const workletReadyRef = useRef(false);
  const preloadCtxRef = useRef<AudioContext | null>(null);
  const sessionActiveRef = useRef(false);
  const readyReceivedRef = useRef(false);
  const connectInFlightRef = useRef(false);
  const connectGenRef = useRef(0);
  const thinkingTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const fail = useCallback((message: string) => {
    setError(message);
    setStatus("error");
  }, []);

  const clearThinkingTimer = useCallback(() => {
    if (thinkingTimerRef.current) {
      clearTimeout(thinkingTimerRef.current);
      thinkingTimerRef.current = null;
    }
  }, []);

  const playPcm16 = useCallback((buffer: ArrayBuffer) => {
    const audioCtx = audioCtxRef.current;
    if (!audioCtx) return;

    const int16 = new Int16Array(buffer);
    const float32 = new Float32Array(int16.length);
    for (let i = 0; i < int16.length; i++) {
      float32[i] = int16[i] / 32768;
    }

    const audioBuffer = audioCtx.createBuffer(1, float32.length, 16000);
    audioBuffer.copyToChannel(float32, 0);
    const source = audioCtx.createBufferSource();
    source.buffer = audioBuffer;
    source.connect(audioCtx.destination);
    const start = Math.max(audioCtx.currentTime, nextPlayTimeRef.current);
    source.start(start);
    nextPlayTimeRef.current = start + audioBuffer.duration;
  }, []);

  const resetPlayback = useCallback(() => {
    nextPlayTimeRef.current = audioCtxRef.current?.currentTime ?? 0;
  }, []);

  const cleanup = useCallback((sendStop = true) => {
    connectGenRef.current += 1;
    connectInFlightRef.current = false;
    clearThinkingTimer();
    sessionActiveRef.current = false;
    readyReceivedRef.current = false;
    setActive(false);
    setSessionId(null);

    const ws = wsRef.current;
    if (ws && ws.readyState === WebSocket.OPEN) {
      if (sendStop) ws.send(JSON.stringify({ event: "stop" }));
      ws.close();
    }
    wsRef.current = null;

    workletRef.current?.disconnect();
    workletRef.current = null;

    streamRef.current?.getTracks().forEach((t) => t.stop());
    streamRef.current = null;

    if (audioCtxRef.current) {
      void audioCtxRef.current.close();
      audioCtxRef.current = null;
    }
    workletReadyRef.current = false;
    nextPlayTimeRef.current = 0;
  }, [clearThinkingTimer]);

  const armThinkingTimer = useCallback(() => {
    // No hard timeout — table booking API calls can take 30s+.
    clearThinkingTimer();
  }, [clearThinkingTimer]);

  const preloadWorklet = useCallback(async () => {
    if (workletReadyRef.current) return;
    try {
      preloadCtxRef.current = new AudioContext({ sampleRate: 16000 });
      await preloadCtxRef.current.audioWorklet.addModule("/voice-agent/audio-processor.js");
      workletReadyRef.current = true;
    } catch {
      // Non-fatal
    }
  }, []);

  const startMic = useCallback(async () => {
    const audioCtx = audioCtxRef.current;
    const stream = streamRef.current;
    const ws = wsRef.current;
    if (!audioCtx || !stream || !ws) return;

    const source = audioCtx.createMediaStreamSource(stream);
    const worklet = new AudioWorkletNode(audioCtx, "mic-processor");
    worklet.port.onmessage = (e: MessageEvent<ArrayBuffer>) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(e.data);
      }
    };
    source.connect(worklet);
    worklet.connect(audioCtx.destination);
    workletRef.current = worklet;
  }, []);

  const handleMessage = useCallback(
    async (data: string | ArrayBuffer) => {
      if (typeof data === "string") {
        let msg: {
          event?: string;
          state?: string;
          role?: "user" | "assistant";
          text?: string;
          message?: string;
          session_id?: string;
          call_db_id?: number;
          status?: string;
          reservation_id?: string;
          guest_name?: string;
          guest_phone?: string;
          party_size?: number;
          slot?: string;
          booking_date?: string;
          booking_time?: string;
        };
        try {
          msg = JSON.parse(data);
        } catch {
          return;
        }

        if (msg.event === "error" && msg.message) {
          fail(msg.message);
          cleanup(false);
          return;
        }

        if (msg.event === "ready") {
          readyReceivedRef.current = true;
          clearThinkingTimer();
          if (msg.session_id) setSessionId(msg.session_id);
          await startMic();
          sessionActiveRef.current = true;
          setActive(true);
          setStatus("speaking");
        }

        if (msg.event === "status" && msg.state) {
          const map: Record<string, VoiceSessionStatus> = {
            user_speaking: "user-speaking",
            thinking: "thinking",
            bot_speaking: "speaking",
            listening: "listening",
          };
          if (msg.state === "user_speaking") resetPlayback();
          const next = map[msg.state] ?? "listening";
          setStatus(next);
          if (next === "thinking") {
            armThinkingTimer();
          } else {
            clearThinkingTimer();
          }
        }

        // The voice service keeps its operational log; signed demo pages also
        // forward each turn to the main engagement ledger for restaurant review.
        if (msg.event === "transcript" && msg.role && msg.text) {
          clearThinkingTimer();
          window.dispatchEvent(
            new CustomEvent("tuvi:voice-transcript", {
              detail: { role: msg.role, text: msg.text },
            }),
          );
        }

        if (msg.event === "booking" && msg.status === "pending" && msg.reservation_id) {
          const receipt: ReservationRequestReceipt = {
            reservationId: msg.reservation_id,
            status: msg.status,
            guestName: msg.guest_name ?? "",
            guestPhone: msg.guest_phone ?? "",
            partySize: msg.party_size ?? 2,
            slot: msg.slot ?? "",
            bookingDate: msg.booking_date ?? "",
            bookingTime: msg.booking_time ?? "",
            message: msg.message ?? "",
          };
          setBooking(receipt);
        }
        return;
      }

      clearThinkingTimer();
      playPcm16(data);
    },
    [
      armThinkingTimer,
      cleanup,
      clearThinkingTimer,
      fail,
      playPcm16,
      resetPlayback,
      startMic,
    ],
  );

  const prefetchStatus = useCallback(async () => {
    setStatus("checking");
    const readiness = await checkVoiceAgentReadiness();
    if (!readiness.ok) {
      fail(readiness.message);
      return readiness;
    }
    setError(null);
    setStatus("idle");
    return readiness;
  }, [fail]);

  const connect = useCallback(async () => {
    if (connectInFlightRef.current || sessionActiveRef.current) return;

    connectInFlightRef.current = true;
    const gen = ++connectGenRef.current;

    setError(null);
    setStatus("checking");
    setSessionId(null);

    const readiness = await checkVoiceAgentReadiness();
    if (gen !== connectGenRef.current) return;
    if (!readiness.ok) {
      connectInFlightRef.current = false;
      fail(readiness.message);
      return;
    }

    try {
      setStatus("connecting");
      await preloadWorklet();
      if (gen !== connectGenRef.current) return;

      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          channelCount: 1,
          sampleRate: 16000,
          echoCancellation: true,
          noiseSuppression: true,
        },
      });
      if (gen !== connectGenRef.current) {
        stream.getTracks().forEach((t) => t.stop());
        return;
      }
      streamRef.current = stream;

      let audioCtx: AudioContext;
      const preloaded = preloadCtxRef.current;
      if (workletReadyRef.current && preloaded && preloaded.state !== "closed") {
        audioCtx = preloaded;
        preloadCtxRef.current = null;
      } else {
        audioCtx = new AudioContext({ sampleRate: 16000 });
        if (!workletReadyRef.current) {
          await audioCtx.audioWorklet.addModule("/voice-agent/audio-processor.js");
          workletReadyRef.current = true;
        }
      }
      if (gen !== connectGenRef.current) {
        void audioCtx.close();
        return;
      }

      if (audioCtx.state !== "running") await audioCtx.resume();
      audioCtxRef.current = audioCtx;

      const ws = new WebSocket(getVoiceAgentWsUrl(restaurantIndex));
      ws.binaryType = "arraybuffer";
      wsRef.current = ws;

      let readyTimer: ReturnType<typeof setTimeout> | null = null;

      ws.onclose = () => {
        if (gen !== connectGenRef.current) return;
        if (readyTimer) clearTimeout(readyTimer);
        if (sessionActiveRef.current) {
          cleanup(false);
        } else if (!readyReceivedRef.current) {
          fail("Connection closed before session started. Check voice agent logs.");
          cleanup(false);
        }
      };

      ws.onmessage = async (evt) => {
        if (gen !== connectGenRef.current) return;
        await handleMessage(evt.data);
      };

      await waitForWsOpen(ws, 8000);
      if (gen !== connectGenRef.current) return;

      readyTimer = setTimeout(() => {
        if (gen !== connectGenRef.current) return;
        if (!readyReceivedRef.current) {
          fail("Session start timed out. Check API keys in voice-sales-agent/.env");
          cleanup(false);
        }
      }, READY_TIMEOUT_MS);
    } catch (e) {
      if (gen !== connectGenRef.current) return;
      if (e instanceof DOMException && e.name === "NotAllowedError") {
        fail("Microphone access is required. Allow mic permission and try again.");
      } else {
        const message = e instanceof Error ? e.message : "Could not start voice session";
        fail(message);
      }
      cleanup(false);
    } finally {
      if (gen === connectGenRef.current) {
        connectInFlightRef.current = false;
      }
    }
  }, [cleanup, fail, handleMessage, preloadWorklet, restaurantIndex]);

  const disconnect = useCallback(() => {
    cleanup(true);
    setStatus("idle");
  }, [cleanup]);

  const reset = useCallback(() => {
    setError(null);
    setStatus("idle");
    setBooking(null);
    setSessionId(null);
  }, []);

  const dismissBooking = useCallback(() => {
    setBooking(null);
  }, []);

  useEffect(() => () => cleanup(false), [cleanup]);

  return {
    status,
    error,
    active,
    booking,
    sessionId,
    connect,
    disconnect,
    reset,
    dismissBooking,
    prefetchStatus,
    preloadWorklet,
  };
}
