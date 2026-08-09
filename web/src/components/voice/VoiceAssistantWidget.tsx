"use client";

import { useEffect, useRef, useState } from "react";
import {
  useVoiceAgentSession,
  type BookingDetails,
  type BookingProgressPhase,
} from "@/hooks/useVoiceAgentSession";
import VoiceOrbCanvas, { statusGlow } from "@/components/voice/VoiceOrbCanvas";

const STATUS_LABELS: Record<string, string> = {
  idle: "Ready",
  checking: "Checking…",
  error: "Unavailable",
  connecting: "Connecting…",
  listening: "Listening",
  thinking: "Thinking",
  speaking: "Speaking",
  "user-speaking": "Listening",
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
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-sage text-accent">
            <CheckIcon />
          </div>
        ) : (
          <SpinnerIcon className="text-accent" />
        )}
        <p className={`text-sm font-semibold ${isSuccess ? "text-accent" : "text-ink"}`}>
          {message || (phase === "checking_slots" ? "Checking slots…" : "Booking slot…")}
        </p>
        {!isSuccess && (
          <p className="text-xs text-muted">AI is still listening — you can keep talking</p>
        )}
      </div>
    </div>
  );
}

function EmailPromptForm({
  prompt,
  onSubmit,
}: {
  prompt: string;
  onSubmit: (email: string) => boolean;
}) {
  const [email, setEmail] = useState("");

  return (
    <div
      className="absolute inset-0 z-30 flex items-center justify-center rounded-2xl bg-bg/80 p-4 backdrop-blur-md"
      role="dialog"
      aria-label="Enter email"
    >
      <form
        className="w-full space-y-3 rounded-xl border border-accent/30 bg-bg p-4 shadow-xl"
        onSubmit={(event) => {
          event.preventDefault();
          onSubmit(email);
        }}
      >
        <p className="text-[11px] font-bold uppercase tracking-[0.16em] text-accent">
          Confirm email
        </p>
        <p className="text-sm text-ink">{prompt}</p>
        <input
          type="email"
          required
          autoFocus
          autoComplete="email"
          placeholder="you@company.com"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-ink outline-none placeholder:text-muted/70 focus:border-primary"
        />
        <button
          type="submit"
          className="w-full cursor-pointer rounded-xl bg-ink px-4 py-2.5 text-sm font-semibold text-bg transition-colors hover:bg-primary-dim"
        >
          Enter
        </button>
        <p className="text-center text-[11px] text-muted">
          Press Enter to confirm — meeting books to this address
        </p>
      </form>
    </div>
  );
}

function BookingDetailsForm({
  prompt,
  onSubmit,
}: {
  prompt: string;
  onSubmit: (details: BookingDetails) => boolean;
}) {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");

  return (
    <div
      className="absolute inset-0 z-30 flex items-start justify-center overflow-y-auto rounded-2xl bg-bg/80 p-4 backdrop-blur-md sm:items-center"
      role="dialog"
      aria-label="Enter booking details"
    >
      <form
        className="w-full space-y-3 rounded-xl border border-accent/30 bg-bg p-4 shadow-xl"
        onSubmit={(event) => {
          event.preventDefault();
          onSubmit({
            prospectName: name,
            prospectEmail: email,
            prospectPhone: phone,
          });
        }}
      >
        <p className="text-[11px] font-bold uppercase tracking-[0.16em] text-accent">
          Confirm your booking
        </p>
        <p className="text-sm text-ink">{prompt}</p>
        <label className="block text-xs font-semibold text-muted">
          Name
          <input
            type="text"
            required
            autoFocus
            autoComplete="name"
            placeholder="Your full name"
            value={name}
            onChange={(event) => setName(event.target.value)}
            className="mt-1.5 w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm font-normal text-ink outline-none placeholder:text-muted/70 focus:border-primary"
          />
        </label>
        <label className="block text-xs font-semibold text-muted">
          Email
          <input
            type="email"
            required
            autoComplete="email"
            placeholder="you@company.com"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            className="mt-1.5 w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm font-normal text-ink outline-none placeholder:text-muted/70 focus:border-primary"
          />
        </label>
        <label className="block text-xs font-semibold text-muted">
          Phone
          <input
            type="tel"
            required
            autoComplete="tel"
            placeholder="+61 4XX XXX XXX"
            value={phone}
            onChange={(event) => setPhone(event.target.value)}
            className="mt-1.5 w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm font-normal text-ink outline-none placeholder:text-muted/70 focus:border-primary"
          />
        </label>
        <button
          type="submit"
          className="w-full cursor-pointer rounded-xl bg-ink px-4 py-2.5 text-sm font-semibold text-bg transition-colors hover:bg-primary-dim"
        >
          Confirm booking
        </button>
        <p className="text-center text-[11px] text-muted">
          Tuvi AI will confirm the appointment after submission.
        </p>
      </form>
    </div>
  );
}

export default function VoiceAssistantWidget() {
  const [open, setOpen] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);
  const {
    status,
    error,
    active,
    consultation,
    bookingProgress,
    emailPrompt,
    bookingDetailsPrompt,
    connect,
    disconnect,
    reset,
    dismissConsultation,
    submitEmail,
    submitBookingDetails,
    prefetchStatus,
    preloadWorklet,
  } = useVoiceAgentSession();

  useEffect(() => {
    if (open) void prefetchStatus();
  }, [open, prefetchStatus]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (emailPrompt || bookingDetailsPrompt) return;
        if (active) disconnect();
        setOpen(false);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, active, disconnect, emailPrompt, bookingDetailsPrompt]);

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

  const closePanel = () => {
    if (active) disconnect();
    setOpen(false);
  };

  const activeBookingProgress: {
    phase: "checking_slots" | "booking_slot" | "success";
    message: string;
  } | null =
    bookingProgress?.phase === "checking_slots" ||
    bookingProgress?.phase === "booking_slot" ||
    bookingProgress?.phase === "success"
      ? { phase: bookingProgress.phase, message: bookingProgress.message }
      : null;

  const glow = statusGlow(status);

  return (
    <>
      {consultation && (
        <div
          className="pointer-events-auto fixed inset-0 z-[90] flex items-center justify-center bg-black/55 p-4 backdrop-blur-sm"
          role="dialog"
          aria-modal="true"
          aria-label="Consultation confirmation"
        >
          <div className="w-full max-w-sm rounded-2xl border border-border bg-bg p-6 text-center shadow-2xl">
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-accent">
              Consultation booked
            </p>
            <h2 className="mt-2 text-xl font-semibold text-ink">Your call is confirmed</h2>
            <p className="mt-2 text-sm text-muted">
              {consultation.prospectName ? `Thanks, ${consultation.prospectName}. ` : ""}
              Save your confirmation number below.
            </p>

            <div className="mt-5 space-y-3 rounded-xl border border-border bg-surface px-4 py-4 text-left">
              <div className="flex justify-between gap-3">
                <span className="text-[11px] font-semibold uppercase tracking-wider text-muted">Day</span>
                <span className="text-sm font-medium text-ink">
                  {formatDay(consultation.slot, consultation.bookingDate)}
                </span>
              </div>
              <div className="flex justify-between gap-3">
                <span className="text-[11px] font-semibold uppercase tracking-wider text-muted">Time</span>
                <span className="text-sm font-medium text-ink">
                  {formatTime(consultation.slot, consultation.bookingTime)}
                </span>
              </div>
            </div>

            <div className="mt-4 rounded-xl border border-primary/30 bg-primary/5 px-4 py-4">
              <p className="text-[11px] font-semibold uppercase tracking-wider text-muted">
                Confirmation ID
              </p>
              <p className="mt-2 font-mono text-3xl font-bold tracking-[0.2em] text-primary">
                {consultation.confirmationCode}
              </p>
            </div>

            <div className="mt-4 flex flex-col gap-2">
              <button
                type="button"
                onClick={dismissConsultation}
                className="rounded-xl border border-border px-4 py-3 text-sm font-medium text-muted transition hover:text-ink"
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
            className="pointer-events-auto relative mb-3 w-[min(440px,calc(100vw-2.5rem))] overflow-hidden rounded-2xl border border-border bg-bg/95 p-5 text-ink shadow-2xl backdrop-blur-xl"
            role="dialog"
            aria-label="Tuvi AI assistant"
          >
            {activeBookingProgress && (
              <BookingProgressOverlay
                phase={activeBookingProgress.phase}
                message={activeBookingProgress.message}
              />
            )}

            {emailPrompt && <EmailPromptForm prompt={emailPrompt} onSubmit={submitEmail} />}

            {bookingDetailsPrompt && (
              <BookingDetailsForm
                prompt={bookingDetailsPrompt}
                onSubmit={submitBookingDetails}
              />
            )}

            <div className="mb-3 flex items-center justify-between gap-3">
              <p className="text-xs font-semibold uppercase tracking-wider text-accent">Tuvi AI</p>
              <button
                type="button"
                onClick={closePanel}
                className="rounded-lg p-1.5 text-muted transition hover:bg-surface hover:text-ink"
                aria-label="Close assistant"
              >
                <CloseIcon />
              </button>
            </div>

            {error && (
              <div
                className="mb-3 rounded-xl border border-[#b42318]/30 bg-[#b42318]/10 px-3 py-2.5 text-[#b42318]"
                role="alert"
              >
                <p className="text-[11px] font-bold uppercase tracking-wider text-[#b42318]">
                  Configuration error
                </p>
                <p className="mt-1 text-sm leading-relaxed">{error}</p>
              </div>
            )}

            <div className="flex flex-col items-center py-1">
              <VoiceOrbCanvas status={status} size={176} className="mx-auto bg-transparent" />

              <div
                className="mt-3 inline-flex items-center gap-2 rounded-full border border-border bg-surface px-3 py-1.5"
                style={{ borderColor: `${glow}40` }}
              >
                <span
                  className={`h-2 w-2 shrink-0 rounded-full ${
                    status === "thinking" || status === "connecting" ? "animate-pulse" : ""
                  }`}
                  style={{ backgroundColor: glow, boxShadow: `0 0 8px ${glow}` }}
                />
                <span className="text-[11px] font-semibold uppercase tracking-[0.14em] text-muted">
                  {STATUS_LABELS[status] ?? status}
                </span>
              </div>

              <div className="mt-4 flex w-full items-center gap-2">
                <button
                  type="button"
                  onMouseDown={() => void preloadWorklet()}
                  onClick={() => void handlePrimaryAction()}
                  className="flex h-12 flex-1 cursor-pointer items-center justify-center gap-2 rounded-full bg-ink px-4 text-sm font-semibold text-bg transition-colors hover:bg-primary-dim"
                >
                  <MicIcon />
                  {active ? "End" : error ? "Check again" : "Start talking"}
                </button>

                {error && !active && (
                  <button
                    type="button"
                    onClick={reset}
                    className="h-12 shrink-0 rounded-full border border-border px-3 text-sm text-muted transition hover:bg-surface"
                  >
                    Reset
                  </button>
                )}
              </div>
            </div>
          </div>
        )}

        <button
          type="button"
          onClick={() => (open ? closePanel() : handlePrimaryAction())}
          onMouseDown={() => void preloadWorklet()}
          className="pointer-events-auto flex cursor-pointer items-center gap-2.5 rounded-full bg-white px-3.5 py-3.5 text-sm font-semibold text-black shadow-[0_12px_34px_rgba(0,0,0,0.35)] transition-colors hover:bg-white/90 sm:px-5"
          aria-expanded={open}
          aria-haspopup="dialog"
          aria-label="Talk to Tuvi AI"
        >
          <MicIcon />
          <span className="hidden sm:inline">Talk to Tuvi AI</span>
        </button>
      </div>
    </>
  );
}
