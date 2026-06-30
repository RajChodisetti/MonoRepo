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

export type TranscriptTurn = {
  id: string;
  role: "user" | "assistant";
  text: string;
};

type ReadinessResult =
  | { ok: true }
  | { ok: false; message: string; missing?: string[] };

const READY_TIMEOUT_MS = 12_000;
const THINKING_TIMEOUT_MS = 15_000;

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

export function useVoiceAgentSession() {
  const [status, setStatus] = useState<VoiceSessionStatus>("idle");
  const [error, setError] = useState<string | null>(null);
  const [transcript, setTranscript] = useState<TranscriptTurn[]>([]);
  const [active, setActive] = useState(false);

  const wsRef = useRef<WebSocket | null>(null);
  const audioCtxRef = useRef<AudioContext | null>(null);
  const workletRef = useRef<AudioWorkletNode | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const nextPlayTimeRef = useRef(0);
  const workletReadyRef = useRef(false);
  const preloadCtxRef = useRef<AudioContext | null>(null);
  const sessionActiveRef = useRef(false);
  const readyReceivedRef = useRef(false);
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

  const addTurn = useCallback((role: "user" | "assistant", text: string) => {
    setTranscript((prev) => [
      ...prev,
      { id: `${Date.now()}-${prev.length}`, role, text },
    ]);
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
    clearThinkingTimer();
    sessionActiveRef.current = false;
    readyReceivedRef.current = false;
    setActive(false);

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
    clearThinkingTimer();
    thinkingTimerRef.current = setTimeout(() => {
      fail(
        "AI response timed out. Check API keys (OpenAI, Deepgram, Cartesia) in voice-sales-agent/.env",
      );
      cleanup(false);
    }, THINKING_TIMEOUT_MS);
  }, [cleanup, clearThinkingTimer, fail]);

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

        if (msg.event === "transcript" && msg.role && msg.text) {
          clearThinkingTimer();
          addTurn(msg.role, msg.text);
        }
        return;
      }

      clearThinkingTimer();
      playPcm16(data);
    },
    [
      addTurn,
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
    setError(null);
    setStatus("checking");

    const readiness = await checkVoiceAgentReadiness();
    if (!readiness.ok) {
      fail(readiness.message);
      return;
    }

    try {
      setStatus("connecting");
      await preloadWorklet();

      const ws = new WebSocket(getVoiceAgentWsUrl());
      ws.binaryType = "arraybuffer";
      wsRef.current = ws;

      const earlyMessages: (string | ArrayBuffer)[] = [];
      let audioReady = false;
      let readyTimer: ReturnType<typeof setTimeout> | null = null;

      ws.onclose = () => {
        if (readyTimer) clearTimeout(readyTimer);
        if (sessionActiveRef.current) {
          cleanup(false);
        } else if (!readyReceivedRef.current) {
          fail("Connection closed before session started. Check voice agent logs.");
          cleanup(false);
        }
      };

      ws.onmessage = async (evt) => {
        if (!audioReady) {
          earlyMessages.push(evt.data);
          return;
        }
        await handleMessage(evt.data);
      };

      await waitForWsOpen(ws, 8000);

      readyTimer = setTimeout(() => {
        if (!readyReceivedRef.current) {
          fail("Session start timed out. Check API keys in voice-sales-agent/.env");
          cleanup(false);
        }
      }, READY_TIMEOUT_MS);

      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          channelCount: 1,
          sampleRate: 16000,
          echoCancellation: true,
          noiseSuppression: true,
        },
      });
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

      if (audioCtx.state !== "running") await audioCtx.resume();
      audioCtxRef.current = audioCtx;
      audioReady = true;

      for (const msg of earlyMessages) {
        await handleMessage(msg);
      }

      ws.onmessage = async (evt) => {
        await handleMessage(evt.data);
      };
    } catch (e) {
      const message = e instanceof Error ? e.message : "Could not start voice session";
      fail(message);
      cleanup(false);
    }
  }, [cleanup, fail, handleMessage, preloadWorklet]);

  const disconnect = useCallback(() => {
    cleanup(true);
    setStatus("idle");
  }, [cleanup]);

  const reset = useCallback(() => {
    setError(null);
    setStatus("idle");
    setTranscript([]);
  }, []);

  useEffect(() => () => cleanup(false), [cleanup]);

  return {
    status,
    error,
    transcript,
    active,
    connect,
    disconnect,
    reset,
    prefetchStatus,
    preloadWorklet,
  };
}
