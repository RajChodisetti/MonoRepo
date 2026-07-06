"use client";

import { useEffect, useRef, useState } from "react";
import { useVoiceAgentSession, type BookingProgressPhase } from "@/hooks/useVoiceAgentSession";

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

function formatDay(slot: string, bookingDate?: string): string {
  if (slot) {
    try {
      return new Intl.DateTimeFormat("en-AU", {
        weekday: "long",
        day: "numeric",
        month: "long",
        year: "numeric",
      }).format(new Date(slot));
    } catch {
      // fall through
    }
  }
  if (bookingDate) {
    try {
      return new Intl.DateTimeFormat("en-AU", {
        weekday: "long",
        day: "numeric",
        month: "long",
        year: "numeric",
      }).format(new Date(`${bookingDate}T12:00:00`));
    } catch {
      return bookingDate;
    }
  }
  return "—";
}

function formatTime(slot: string, bookingTime?: string): string {
  if (slot) {
    try {
      return new Intl.DateTimeFormat("en-AU", {
        hour: "numeric",
        minute: "2-digit",
      }).format(new Date(slot));
    } catch {
      // fall through
    }
  }
  return bookingTime || "—";
}

function SpinnerIcon({ className }: { className?: string }) {
  return (
    <svg
      className={`animate-spin ${className ?? ""}`}
      width="32"
      height="32"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden
    >
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
      <path
        className="opacity-90"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
      />
    </svg>
  );
}

function CheckIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      width="36"
      height="36"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M20 6 9 17l-5-5" />
    </svg>
  );
}

function BookingProgressOverlay({
  phase,
  message,
}: {
  phase: BookingProgressPhase;
  message: string;
}) {
  const isSuccess = phase === "success";

  return (
    <div
      className="absolute inset-0 z-20 flex items-center justify-center rounded-2xl bg-bg/75 backdrop-blur-md"
      role="status"
      aria-live="polite"
      aria-label={message}
    >
      <div className="flex flex-col items-center gap-3 px-6 text-center">
        {isSuccess ? (
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-emerald-500/20 text-emerald-400">
            <CheckIcon />
          </div>
        ) : (
          <SpinnerIcon className="text-cyan" />
        )}
        <p className={`text-sm font-semibold ${isSuccess ? "text-emerald-300" : "text-text"}`}>
          {message || (phase === "checking_slots" ? "Checking slots…" : "Booking slot…")}
        </p>
        {!isSuccess && (
          <p className="text-xs text-muted">AI is still listening — you can keep talking</p>
        )}
      </div>
    </div>
  );
}

export default function VoiceAssistantWidget() {
  const [open, setOpen] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);
  const {
    status,
    error,
    transcript,
    active,
    consultation,
    bookingProgress,
    connect,
    disconnect,
    reset,
    dismissConsultation,
    prefetchStatus,
    preloadWorklet,
  } = useVoiceAgentSession();

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

  const calendarUrl = consultation?.calendarLink || consultation?.calendlyLink;
  const showBookingOverlay =
    bookingProgress &&
    bookingProgress.phase !== "idle" &&
    (bookingProgress.phase === "checking_slots" ||
      bookingProgress.phase === "booking_slot" ||
      bookingProgress.phase === "success");

  return (
    <>
      {consultation && (
        <div
          className="pointer-events-auto fixed inset-0 z-[90] flex items-center justify-center bg-black/55 p-4 backdrop-blur-sm"
          role="dialog"
          aria-modal="true"
          aria-label="Consultation confirmation"
        >
          <div className="w-full max-w-sm rounded-2xl border border-white/15 bg-bg-elevated p-6 text-center shadow-2xl">
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-cyan">
              Consultation booked
            </p>
            <h2 className="mt-2 text-xl font-semibold text-text">Your call is confirmed</h2>
            <p className="mt-2 text-sm text-muted">
              {consultation.prospectName ? `Thanks, ${consultation.prospectName}. ` : ""}
              Save your confirmation number below.
            </p>

            <div className="mt-5 space-y-3 rounded-xl border border-white/10 bg-white/5 px-4 py-4 text-left">
              <div className="flex justify-between gap-3">
                <span className="text-[11px] font-semibold uppercase tracking-wider text-muted">Day</span>
                <span className="text-sm font-medium text-text">
                  {formatDay(consultation.slot, consultation.bookingDate)}
                </span>
              </div>
              <div className="flex justify-between gap-3">
                <span className="text-[11px] font-semibold uppercase tracking-wider text-muted">Time</span>
                <span className="text-sm font-medium text-text">
                  {formatTime(consultation.slot, consultation.bookingTime)}
                </span>
              </div>
            </div>

            <div className="mt-4 rounded-xl border border-gold/30 bg-gold/5 px-4 py-4">
              <p className="text-[11px] font-semibold uppercase tracking-wider text-muted">
                Confirmation ID
              </p>
              <p className="mt-2 font-mono text-3xl font-bold tracking-[0.2em] text-gold">
                {consultation.confirmationCode}
              </p>
            </div>

            <div className="mt-4 flex flex-col gap-2">
              {calendarUrl ? (
                <a
                  href={calendarUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="rounded-xl bg-gradient-to-r from-gold-dim to-gold px-4 py-3 text-sm font-semibold text-bg transition hover:opacity-90"
                >
                  Open in Google Calendar
                </a>
              ) : null}
              <button
                type="button"
                onClick={dismissConsultation}
                className="rounded-xl border border-white/15 px-4 py-3 text-sm font-medium text-muted transition hover:text-text"
              >
                Got it
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="pointer-events-none fixed bottom-5 right-5 z-[70] md:bottom-6 md:right-6">
        {open && (
          <div
            ref={panelRef}
            className="pointer-events-auto relative mb-3 w-[min(380px,calc(100vw-2.5rem))] rounded-2xl border border-white/15 bg-bg-elevated/95 p-4 text-text shadow-2xl backdrop-blur-xl"
            role="dialog"
            aria-label="Tuvi AI assistant"
          >
            {showBookingOverlay && (
              <BookingProgressOverlay
                phase={bookingProgress.phase}
                message={bookingProgress.message}
              />
            )}
            {error && (
              <div
                className="mb-3 rounded-xl border border-red-400/40 bg-red-500/15 px-3 py-2.5 text-red-100"
                role="alert"
              >
                <p className="text-[11px] font-bold uppercase tracking-wider text-red-300">
                  Configuration error
                </p>
                <p className="mt-1 text-sm leading-relaxed">{error}</p>
              </div>
            )}

            <div className="mb-3 flex items-start justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-cyan">
                  Tuvi AI Assistant
                </p>
                <h2 className="mt-1 text-lg font-semibold">Ask about our services</h2>
                <p className="mt-1 text-xs text-muted">
                  Book a free consultation or learn what we build.
                </p>
              </div>
              <button
                type="button"
                onClick={() => {
                  if (active) disconnect();
                  setOpen(false);
                }}
                className="rounded-lg p-1.5 text-muted transition hover:bg-white/10 hover:text-text"
                aria-label="Close assistant"
              >
                <CloseIcon />
              </button>
            </div>

            <div className="mb-3 flex items-center gap-2 text-sm text-muted">
              <span
                className={`h-2 w-2 shrink-0 rounded-full ${
                  status === "listening" || status === "user-speaking"
                    ? "bg-emerald-400"
                    : status === "speaking"
                      ? "bg-sky-400"
                      : status === "thinking" || status === "connecting"
                        ? "bg-amber-400"
                        : status === "error"
                          ? "bg-red-400"
                          : "bg-white/35"
                }`}
              />
              <span>{STATUS_LABELS[status] ?? status}</span>
            </div>

            <div className="mb-3 max-h-44 min-h-[100px] overflow-y-auto rounded-xl bg-white/5 p-3 text-sm leading-relaxed">
              {transcript.length === 0 ? (
                <p className="italic text-muted">
                  {error
                    ? "Fix the error above, then try again."
                    : "Start talking — ask about custom software, AI, or book a call."}
                </p>
              ) : (
                <div className="space-y-3">
                  {transcript.map((turn) => (
                    <div key={turn.id}>
                      <p className="text-[11px] font-semibold uppercase tracking-wide text-cyan">
                        {turn.role === "user" ? "You" : "Tuvi AI"}
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
                className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-gold-dim to-gold px-4 py-3 text-sm font-semibold text-bg transition hover:opacity-90"
              >
                <MicIcon />
                {active ? "End conversation" : error ? "Check again" : "Start talking"}
              </button>
              {(error || transcript.length > 0) && !active && (
                <button
                  type="button"
                  onClick={reset}
                  className="rounded-xl border border-white/15 px-3 py-3 text-sm text-muted transition hover:bg-white/5"
                >
                  Reset
                </button>
              )}
            </div>
          </div>
        )}

        <button
          type="button"
          onClick={() => (open ? setOpen(false) : handlePrimaryAction())}
          onMouseDown={() => void preloadWorklet()}
          className="pointer-events-auto flex items-center gap-2.5 rounded-full bg-gradient-to-r from-gold-dim to-gold px-5 py-3.5 text-sm font-semibold text-bg shadow-[0_8px_32px_rgba(212,168,83,0.35)] transition hover:scale-[1.02] active:scale-[0.98]"
          aria-expanded={open}
          aria-haspopup="dialog"
        >
          <MicIcon />
          <span>Talk to Tuvi AI</span>
        </button>
      </div>
    </>
  );
}
