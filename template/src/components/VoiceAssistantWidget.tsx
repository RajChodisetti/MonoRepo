"use client";

import { useEffect, useRef, useState } from "react";
import type { TemplateId } from "@/lib/templateConfig";
import { useVoiceAgentSession } from "@/hooks/useVoiceAgentSession";

const STATUS_LABELS: Record<string, string> = {
  idle: "Ready to connect",
  checking: "Checking configuration…",
  error: "Unavailable",
  connecting: "Connecting…",
  listening: "Listening…",
  thinking: "AI is thinking…",
  speaking: "AI is speaking…",
  "user-speaking": "You are speaking…",
};

function MicIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <rect x="9" y="2" width="6" height="11" rx="3" />
      <path d="M5 10a7 7 0 0 0 14 0M12 19v3M9 22h6" />
    </svg>
  );
}

function CloseIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M18 6 6 18M6 6l12 12" />
    </svg>
  );
}

export default function VoiceAssistantWidget({ templateId }: { templateId: TemplateId }) {
  const [open, setOpen] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);
  const {
    status,
    error,
    transcript,
    active,
    connect,
    disconnect,
    reset,
    prefetchStatus,
    preloadWorklet,
  } = useVoiceAgentSession();

  const isAurora = templateId === "2";

  useEffect(() => {
    if (!open) return;
    void prefetchStatus();
  }, [open, prefetchStatus]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (active) disconnect();
        setOpen(false);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, active, disconnect]);

  const handleTogglePanel = () => {
    setOpen((v) => !v);
  };

  const handlePrimaryAction = async () => {
    if (!open) {
      setOpen(true);
      return;
    }
    if (active) {
      disconnect();
      return;
    }
    if (error) {
      await prefetchStatus();
      return;
    }
    await connect();
  };

  const panelClass = isAurora
    ? "border border-white/15 bg-[#09090B]/95 text-white shadow-[0_20px_60px_rgba(0,0,0,0.55)] backdrop-blur-xl"
    : "border border-[#e8e0d4]/20 bg-[#1a1614]/95 text-[#f7f0e6] shadow-[0_20px_60px_rgba(0,0,0,0.45)] backdrop-blur-xl";

  const buttonClass = isAurora
    ? "bg-gradient-to-r from-violet-600 to-cyan-500 text-white hover:from-violet-500 hover:to-cyan-400"
    : "bg-[#b88a44] text-[#1a1614] hover:bg-[#c99a54]";

  const accentClass = isAurora ? "text-cyan-300" : "text-[#b88a44]";
  const dimClass = isAurora ? "text-white/55" : "text-[#a89f96]";
  const errorClass = isAurora
    ? "border border-red-400/40 bg-red-500/15 text-red-100"
    : "border border-red-400/35 bg-red-950/50 text-red-100";

  const statusDotClass =
    status === "listening" || status === "user-speaking"
      ? "bg-emerald-400"
      : status === "speaking"
        ? "bg-sky-400"
        : status === "thinking" || status === "connecting" || status === "checking"
          ? "bg-amber-400"
          : status === "error"
            ? "bg-red-400"
            : isAurora
              ? "bg-white/35"
              : "bg-[#8a7340]";

  return (
    <div className="pointer-events-none fixed bottom-20 right-4 z-[70] md:bottom-6 md:right-6">
      {open && (
        <div
          ref={panelRef}
          className={`pointer-events-auto mb-3 w-[min(380px,calc(100vw-2rem))] rounded-2xl p-4 ${panelClass}`}
          role="dialog"
          aria-label="AI voice assistant"
        >
          {error && (
            <div className={`mb-3 rounded-xl px-3 py-2.5 ${errorClass}`} role="alert">
              <p className="text-[11px] font-bold uppercase tracking-[0.12em] text-red-300">
                Configuration error
              </p>
              <p className="mt-1 text-sm leading-relaxed">{error}</p>
            </div>
          )}

          <div className="mb-3 flex items-start justify-between gap-3">
            <div>
              <p className={`text-xs font-semibold uppercase tracking-[0.14em] ${accentClass}`}>
                Voice Assistant
              </p>
              <h2 className="mt-1 text-lg font-semibold">Talk to our AI</h2>
            </div>
            <button
              type="button"
              onClick={() => {
                if (active) disconnect();
                setOpen(false);
              }}
              className={`rounded-lg p-1.5 transition hover:bg-white/10 ${dimClass}`}
              aria-label="Close assistant panel"
            >
              <CloseIcon />
            </button>
          </div>

          <div className={`mb-3 flex items-center gap-2 text-sm ${dimClass}`}>
            <span className={`h-2 w-2 shrink-0 rounded-full ${statusDotClass}`} />
            <span>{STATUS_LABELS[status] ?? status}</span>
          </div>

          <div
            className={`mb-3 max-h-44 min-h-[120px] overflow-y-auto rounded-xl p-3 text-sm leading-relaxed ${
              isAurora ? "bg-white/5" : "bg-black/25"
            }`}
          >
            {transcript.length === 0 ? (
              <p className={`italic ${dimClass}`}>
                {error
                  ? "Fix the error above, then tap Start talking again."
                  : "Start a conversation and the transcript will appear here."}
              </p>
            ) : (
              <div className="space-y-3">
                {transcript.map((turn) => (
                  <div key={turn.id}>
                    <p className={`text-[11px] font-semibold uppercase tracking-wide ${accentClass}`}>
                      {turn.role === "user" ? "You" : "AI Assistant"}
                    </p>
                    <p className="mt-0.5">{turn.text}</p>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="flex gap-2">
            <button
              type="button"
              onMouseDown={() => void preloadWorklet()}
              onClick={() => void handlePrimaryAction()}
              className={`flex flex-1 items-center justify-center gap-2 rounded-xl px-4 py-3 text-sm font-semibold transition ${buttonClass}`}
            >
              <MicIcon />
              {active ? "End conversation" : error ? "Check again" : "Start talking"}
            </button>
            {(error || transcript.length > 0) && !active && (
              <button
                type="button"
                onClick={reset}
                className={`rounded-xl px-3 py-3 text-sm font-medium transition ${
                  isAurora
                    ? "border border-white/15 text-white/70 hover:bg-white/5"
                    : "border border-[#e8e0d4]/20 text-[#d4c4b0] hover:bg-white/5"
                }`}
              >
                Reset
              </button>
            )}
          </div>
        </div>
      )}

      <button
        type="button"
        onClick={handleTogglePanel}
        onMouseDown={() => void preloadWorklet()}
        className={`pointer-events-auto flex items-center gap-2.5 rounded-full px-4 py-3 text-sm font-semibold shadow-lg transition hover:scale-[1.02] active:scale-[0.98] ${buttonClass}`}
        aria-expanded={open}
        aria-haspopup="dialog"
      >
        <MicIcon />
        <span>Try our AI assistant</span>
      </button>
    </div>
  );
}
