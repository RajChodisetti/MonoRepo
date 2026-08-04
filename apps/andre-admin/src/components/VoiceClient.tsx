"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { apiFetch } from "@/lib/client-api";

type Lang = "auto" | "en" | "hi" | "te";
type Turn = { role: "user" | "assistant"; text: string };

function agentWsUrl(language: Lang): string {
  const base =
    process.env.NEXT_PUBLIC_ANDRE_AGENT_WS_URL?.replace(/\/$/, "") ||
    "ws://127.0.0.1:8001";
  return `${base}/browser-stream?language=${encodeURIComponent(language)}`;
}

export function VoiceClient() {
  const [language, setLanguage] = useState<Lang>("en");
  const [status, setStatus] = useState("Ready");
  const [active, setActive] = useState(false);
  const [phone, setPhone] = useState("");
  const [callMsg, setCallMsg] = useState("");
  const [calling, setCalling] = useState(false);
  const [turns, setTurns] = useState<Turn[]>([]);
  const [ready, setReady] = useState<{ browserReady: boolean; phoneReady: boolean }>({
    browserReady: false,
    phoneReady: false,
  });

  const wsRef = useRef<WebSocket | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const audioCtxRef = useRef<AudioContext | null>(null);
  const workletRef = useRef<AudioWorkletNode | null>(null);
  const nextPlayRef = useRef(0);
  const activeRef = useRef(false);

  const refreshReady = useCallback(async () => {
    try {
      const data = await apiFetch<{ browserReady: boolean; phoneReady: boolean }>(
        "/api/voice/ready",
      );
      setReady({ browserReady: data.browserReady, phoneReady: data.phoneReady });
    } catch {
      setReady({ browserReady: false, phoneReady: false });
    }
  }, []);

  useEffect(() => {
    void refreshReady();
    const t = window.setInterval(() => void refreshReady(), 15_000);
    return () => window.clearInterval(t);
  }, [refreshReady]);

  useEffect(() => {
    return () => {
      stopSession(false);
    };
  }, []);

  async function playPCM16(buffer: ArrayBuffer) {
    const audioCtx = audioCtxRef.current;
    if (!audioCtx) return;
    if (audioCtx.state === "suspended") {
      try {
        await audioCtx.resume();
      } catch {
        return;
      }
    }
    if (audioCtx.state !== "running") return;
    const int16 = new Int16Array(buffer);
    const float32 = new Float32Array(int16.length);
    for (let i = 0; i < int16.length; i += 1) float32[i] = int16[i]! / 32768;
    const audioBuffer = audioCtx.createBuffer(1, float32.length, 16000);
    audioBuffer.copyToChannel(float32, 0);
    const source = audioCtx.createBufferSource();
    source.buffer = audioBuffer;
    source.connect(audioCtx.destination);
    const start = Math.max(audioCtx.currentTime, nextPlayRef.current);
    source.start(start);
    nextPlayRef.current = start + audioBuffer.duration;
  }

  async function handleMessage(data: string | ArrayBuffer) {
    if (typeof data !== "string") {
      await playPCM16(data);
      return;
    }
    let msg: {
      event?: string;
      state?: string;
      role?: string;
      text?: string;
      message?: string;
    };
    try {
      msg = JSON.parse(data) as typeof msg;
    } catch {
      return;
    }
    if (msg.event === "ready") {
      await startMic();
      activeRef.current = true;
      setActive(true);
      setStatus("Ananya is greeting you…");
    }
    if (msg.event === "status") {
      const map: Record<string, string> = {
        user_speaking: "Listening to you…",
        thinking: "Ananya is thinking…",
        bot_speaking: "Ananya is speaking…",
        listening: "Listening — speak anytime",
      };
      if (msg.state === "user_speaking") {
        nextPlayRef.current = audioCtxRef.current?.currentTime || 0;
      }
      setStatus(map[msg.state || ""] || "Listening — speak anytime");
    }
    if (msg.event === "transcript" && msg.text) {
      const role = msg.role === "user" ? "user" : "assistant";
      const chunk = msg.text.trim();
      if (!chunk) return;
      setTurns((prev) => {
        const last = prev[prev.length - 1];
        if (last && last.role === role) {
          const needsSpace =
            !last.text.endsWith(" ") &&
            !chunk.startsWith(" ") &&
            !",.!?;:)]}'\"".includes(chunk[0] || "");
          const merged = `${last.text}${needsSpace ? " " : ""}${chunk}`.replace(/\s+/g, " ").trim();
          return [...prev.slice(0, -1), { role, text: merged }];
        }
        return [...prev, { role, text: chunk }];
      });
    }
    if (msg.event === "error") {
      setCallMsg(msg.message || "Session error");
      stopSession(false);
    }
  }

  async function startMic() {
    const audioCtx = audioCtxRef.current;
    const stream = streamRef.current;
    if (!audioCtx || !stream) return;
    const source = audioCtx.createMediaStreamSource(stream);
    const workletNode = new AudioWorkletNode(audioCtx, "mic-processor");
    workletNode.port.onmessage = (e: MessageEvent<ArrayBuffer>) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(e.data);
      }
    };
    source.connect(workletNode);
    workletRef.current = workletNode;
  }

  function stopSession(sendStop = true) {
    activeRef.current = false;
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
    nextPlayRef.current = 0;
    setStatus("Ready");
  }

  async function startSession() {
    setCallMsg("");
    setTurns([]);
    setStatus("Requesting microphone…");
    let socket: WebSocket | null = null;
    let setupDone = false;
    try {
      const micPromise = navigator.mediaDevices.getUserMedia({
        audio: {
          channelCount: 1,
          sampleRate: 16000,
          echoCancellation: true,
          noiseSuppression: true,
        },
      });

      socket = new WebSocket(agentWsUrl(language));
      wsRef.current = socket;
      socket.binaryType = "arraybuffer";

      const early: Array<string | ArrayBuffer> = [];
      let audioReady = false;
      let closedDuringSetup = false;

      socket.onerror = () => {
        if (!setupDone) closedDuringSetup = true;
        else setCallMsg("WebSocket error — is the agent running on :8001?");
      };
      socket.onclose = () => {
        if (activeRef.current || setupDone) stopSession(false);
        else closedDuringSetup = true;
      };
      socket.onmessage = async (evt) => {
        if (!audioReady) {
          early.push(evt.data as string | ArrayBuffer);
          return;
        }
        await handleMessage(evt.data as string | ArrayBuffer);
      };

      streamRef.current = await micPromise;
      if (closedDuringSetup || socket.readyState === WebSocket.CLOSED) {
        throw new Error("Server closed the connection — check agent logs");
      }

      const audioCtx = new AudioContext({ sampleRate: 16000 });
      audioCtxRef.current = audioCtx;
      await audioCtx.audioWorklet.addModule("/audio-processor.js");
      if (audioCtx.state !== "running") await audioCtx.resume();
      audioReady = true;

      for (const data of early) await handleMessage(data);
      if (closedDuringSetup || socket.readyState === WebSocket.CLOSED) {
        throw new Error("Server closed the connection — check agent logs");
      }

      socket.onmessage = async (evt) => {
        await handleMessage(evt.data as string | ArrayBuffer);
      };
      setupDone = true;
      setStatus("Connecting…");
    } catch (err) {
      setCallMsg(err instanceof Error ? err.message : "Could not start");
      setStatus("Ready");
      if (socket) {
        socket.close();
        wsRef.current = null;
      }
      streamRef.current?.getTracks().forEach((t) => t.stop());
      streamRef.current = null;
      if (audioCtxRef.current) {
        void audioCtxRef.current.close();
        audioCtxRef.current = null;
      }
    }
  }

  function toggleSession() {
    if (activeRef.current || (wsRef.current && wsRef.current.readyState <= 1)) {
      stopSession(true);
    } else {
      void startSession();
    }
  }

  async function placeCall() {
    setCallMsg("");
    if (!phone.trim()) {
      setCallMsg("Enter a phone number with country code (E.164).");
      return;
    }
    setCalling(true);
    try {
      const data = await apiFetch<{ call_sid?: string; status?: string; to?: string }>(
        "/api/voice/call",
        {
          method: "POST",
          body: { to: phone.trim(), language },
        },
      );
      setCallMsg(`Call started${data.call_sid ? ` (${data.call_sid})` : ""}.`);
    } catch (err) {
      setCallMsg(err instanceof Error ? err.message : "Call failed");
    } finally {
      setCalling(false);
      void refreshReady();
    }
  }

  return (
    <div className="space-y-5 max-w-3xl">
      <header>
        <h1
          className="text-3xl tracking-tight"
          style={{ fontFamily: "var(--font-fraunces), Georgia, serif" }}
        >
          Voice
        </h1>
        <p className="text-sm text-muted mt-1">
          Talk in the browser or place an outbound call. Language applies to both.
        </p>
      </header>

      <section className="rounded-2xl border border-line bg-panel p-5 space-y-4">
        <div className="flex flex-wrap gap-2">
          <span className={`chip ${ready.browserReady ? "" : "chip-warn"}`}>
            Browser {ready.browserReady ? "ready" : "not ready"}
          </span>
          <span className={`chip ${ready.phoneReady ? "" : "chip-warn"}`}>
            Phone {ready.phoneReady ? "ready" : "not ready"}
          </span>
        </div>

        <div className="field max-w-xs">
          <label htmlFor="language">Language</label>
          <select
            id="language"
            value={language}
            disabled={active}
            onChange={(e) => setLanguage(e.target.value as Lang)}
          >
            <option value="auto">Auto</option>
            <option value="en">English</option>
            <option value="hi">Hindi</option>
            <option value="te">Telugu</option>
          </select>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <button type="button" className="btn btn-primary" onClick={toggleSession}>
            {active ? "Stop conversation" : "Start talking"}
          </button>
          <p className="text-sm text-muted">{status}</p>
        </div>

        <div className="rounded-xl border border-line bg-white/70 min-h-48 p-4 space-y-3">
          {turns.length === 0 ? (
            <p className="text-sm text-muted">
              Conversation appears here after you start talking.
            </p>
          ) : (
            turns.map((turn, idx) => (
              <div key={`${idx}-${turn.role}`}>
                <p className="text-xs font-bold uppercase tracking-wide text-muted">
                  {turn.role === "user" ? "You" : "Ananya"}
                </p>
                <p className="text-sm">{turn.text}</p>
              </div>
            ))
          )}
        </div>
      </section>

      <section className="rounded-2xl border border-line bg-panel p-5 space-y-4">
        <h2 className="text-lg font-semibold">Outbound call</h2>
        <div className="grid gap-4 md:grid-cols-[1fr_auto] items-end">
          <div className="field">
            <label htmlFor="phone">Phone (E.164)</label>
            <input
              id="phone"
              placeholder="+9198…"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
            />
          </div>
          <button
            type="button"
            className="btn btn-primary"
            onClick={placeCall}
            disabled={calling || active}
          >
            {calling ? "Calling…" : "Call"}
          </button>
        </div>
        {callMsg && <p className="text-sm text-muted">{callMsg}</p>}
      </section>
    </div>
  );
}
