"use client";

import { useEffect, useRef, useState } from "react";
import { useVoiceAgentSession, type BookingProgressPhase } from "@/hooks/useVoiceAgentSession";
import RequestCallbackForm from "@/components/RequestCallbackForm";
import VoiceOrbCanvas, { statusGlow } from "@/components/VoiceOrbCanvas";
import {
  getDefaultRealEstateLanguage,
  type RealEstateLanguage,
  type VoiceAgentKind,
} from "@/lib/voiceAgentConfig";

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

function PhoneIcon({ className }: { className?: string }) {
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
      <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6A19.79 19.79 0 0 1 2.12 4.18 2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.13.81.36 1.6.68 2.34a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.74-1.74a2 2 0 0 1 2.11-.45c.74.32 1.53.55 2.34.68A2 2 0 0 1 22 16.92z" />
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
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-emerald-100 text-emerald-600">
            <CheckIcon />
          </div>
        ) : (
          <SpinnerIcon className="text-primary" />
        )}
        <p className={`text-sm font-semibold ${isSuccess ? "text-emerald-600" : "text-text"}`}>
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
  const [showCallback, setShowCallback] = useState(false);
  const [agentKind, setAgentKind] = useState<VoiceAgentKind>("corporate");
  const [language, setLanguage] = useState<RealEstateLanguage>(getDefaultRealEstateLanguage());
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
  } = useVoiceAgentSession({ agent: agentKind, language });

  const [emailInput, setEmailInput] = useState("");
  const [bookingName, setBookingName] = useState("");
  const [bookingEmail, setBookingEmail] = useState("");
  const [bookingPhone, setBookingPhone] = useState("");

  const isRealEstate = agentKind === "real_estate";
  const agentTitle = isRealEstate ? "Real estate agent" : "Tuvi agent";

  useEffect(() => {
    if (emailPrompt) setEmailInput("");
  }, [emailPrompt]);

  useEffect(() => {
    if (!bookingDetailsPrompt) return;
    setBookingName("");
    setBookingEmail("");
    setBookingPhone("");
  }, [bookingDetailsPrompt]);

  useEffect(() => {
    if (!open) {
      setShowCallback(false);
      return;
    }
    void prefetchStatus();
  }, [open, agentKind, prefetchStatus]);

  const switchAgent = (next: VoiceAgentKind) => {
    if (next === agentKind) return;
    if (active) disconnect();
    setShowCallback(false);
    setAgentKind(next);
  };

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (emailPrompt || bookingDetailsPrompt) return;
        if (showCallback) {
          setShowCallback(false);
          return;
        }
        if (active) disconnect();
        setOpen(false);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, active, disconnect, emailPrompt, bookingDetailsPrompt, showCallback]);

  const handlePrimaryAction = async () => {
    if (!open) {
      setOpen(true);
      return;
    }
    if (showCallback) setShowCallback(false);
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
    setShowCallback(false);
    setOpen(false);
  };

  const calendarUrl = consultation?.calendarLink || consultation?.calendlyLink;
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
      {consultation && !isRealEstate && (
        <div
          className="pointer-events-auto fixed inset-0 z-[90] flex items-center justify-center bg-black/55 p-4 backdrop-blur-sm"
          role="dialog"
          aria-modal="true"
          aria-label="Consultation confirmation"
        >
          <div className="w-full max-w-sm rounded-2xl border border-border bg-bg-elevated p-6 text-center shadow-2xl">
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-primary">
              Consultation booked
            </p>
            <h2 className="mt-2 text-xl font-semibold text-text">Your call is confirmed</h2>
            <p className="mt-2 text-sm text-muted">
              {consultation.prospectName ? `Thanks, ${consultation.prospectName}. ` : ""}
              Save your confirmation number below.
            </p>

            <div className="mt-5 space-y-3 rounded-xl border border-border bg-surface px-4 py-4 text-left">
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

            <div className="mt-4 rounded-xl border border-primary/30 bg-primary/5 px-4 py-4">
              <p className="text-[11px] font-semibold uppercase tracking-wider text-muted">
                Confirmation ID
              </p>
              <p className="mt-2 font-mono text-3xl font-bold tracking-[0.2em] text-primary">
                {consultation.confirmationCode}
              </p>
            </div>

            <div className="mt-4 flex flex-col gap-2">
              {calendarUrl ? (
                <a
                  href={calendarUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="rounded-xl bg-primary px-4 py-3 text-sm font-semibold text-bg transition-colors hover:bg-primary-dim"
                >
                  Open in Google Calendar
                </a>
              ) : null}
              <button
                type="button"
                onClick={dismissConsultation}
                className="rounded-xl border border-border px-4 py-3 text-sm font-medium text-muted transition hover:text-text"
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
            className="pointer-events-auto relative mb-3 w-[min(380px,calc(100vw-2.5rem))] overflow-hidden rounded-2xl border border-border bg-bg-elevated p-4 text-text shadow-2xl backdrop-blur-xl sm:p-5"
            role="dialog"
            aria-label={agentTitle}
          >
            {activeBookingProgress && !isRealEstate && (
              <BookingProgressOverlay
                phase={activeBookingProgress.phase}
                message={activeBookingProgress.message}
              />
            )}

            {emailPrompt && !isRealEstate && (
              <div
                className="absolute inset-0 z-30 flex items-start justify-center overflow-y-auto rounded-2xl bg-bg/80 p-4 backdrop-blur-md sm:items-center"
                role="dialog"
                aria-label="Enter email"
              >
                <form
                  className="w-full space-y-3 rounded-xl border border-primary/30 bg-bg-elevated p-4 shadow-xl"
                  onSubmit={(e) => {
                    e.preventDefault();
                    submitEmail(emailInput);
                  }}
                >
                  <p className="text-[11px] font-bold uppercase tracking-[0.16em] text-primary">
                    Confirm email
                  </p>
                  <p className="text-sm text-text">{emailPrompt}</p>
                  <input
                    type="email"
                    required
                    autoFocus
                    autoComplete="email"
                    placeholder="you@company.com"
                    value={emailInput}
                    onChange={(e) => setEmailInput(e.target.value)}
                    className="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text outline-none placeholder:text-muted/70 focus:border-primary/50"
                  />
                  <button
                    type="submit"
                    className="w-full cursor-pointer rounded-xl bg-primary px-4 py-2.5 text-sm font-semibold text-bg transition-colors hover:bg-primary-dim"
                  >
                    Enter
                  </button>
                  <p className="text-center text-[11px] text-muted">
                    Press Enter to confirm — meeting books to this address
                  </p>
                </form>
              </div>
            )}

            {bookingDetailsPrompt && !isRealEstate && (
              <div
                className="absolute inset-0 z-30 flex items-center justify-center rounded-2xl bg-bg/80 p-4 backdrop-blur-md"
                role="dialog"
                aria-label="Enter booking details"
              >
                <form
                  className="w-full space-y-3 rounded-xl border border-primary/30 bg-bg-elevated p-4 shadow-xl"
                  onSubmit={(event) => {
                    event.preventDefault();
                    submitBookingDetails({
                      prospectName: bookingName,
                      prospectEmail: bookingEmail,
                      prospectPhone: bookingPhone,
                    });
                  }}
                >
                  <p className="text-[11px] font-bold uppercase tracking-[0.16em] text-primary">
                    Confirm your booking
                  </p>
                  <p className="text-sm text-text">{bookingDetailsPrompt}</p>
                  <label className="block text-xs font-semibold text-muted">
                    Name
                    <input
                      type="text"
                      required
                      autoFocus
                      autoComplete="name"
                      placeholder="Your full name"
                      value={bookingName}
                      onChange={(event) => setBookingName(event.target.value)}
                      className="mt-1.5 w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm font-normal text-text outline-none placeholder:text-muted/70 focus:border-primary/50"
                    />
                  </label>
                  <label className="block text-xs font-semibold text-muted">
                    Email
                    <input
                      type="email"
                      required
                      autoComplete="email"
                      placeholder="you@company.com"
                      value={bookingEmail}
                      onChange={(event) => setBookingEmail(event.target.value)}
                      className="mt-1.5 w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm font-normal text-text outline-none placeholder:text-muted/70 focus:border-primary/50"
                    />
                  </label>
                  <label className="block text-xs font-semibold text-muted">
                    Phone
                    <input
                      type="tel"
                      required
                      autoComplete="tel"
                      placeholder="+61 4XX XXX XXX"
                      value={bookingPhone}
                      onChange={(event) => setBookingPhone(event.target.value)}
                      className="mt-1.5 w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm font-normal text-text outline-none placeholder:text-muted/70 focus:border-primary/50"
                    />
                  </label>
                  <button
                    type="submit"
                    className="w-full cursor-pointer rounded-xl bg-primary px-4 py-2.5 text-sm font-semibold text-bg transition-colors hover:bg-primary-dim"
                  >
                    Confirm booking
                  </button>
                  <p className="text-center text-[11px] text-muted">
                    Tuvi AI will confirm the appointment after submission.
                  </p>
                </form>
              </div>
            )}

            <div className="flex items-center justify-between gap-2">
              <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-primary">
                {agentTitle}
              </p>
              <button
                type="button"
                onClick={closePanel}
                className="rounded-lg p-1.5 text-muted transition hover:bg-surface hover:text-text"
                aria-label="Close assistant"
              >
                <CloseIcon />
              </button>
            </div>

            <div
              className="mt-3 grid grid-cols-2 gap-1 rounded-xl border border-border bg-surface p-1"
              role="tablist"
              aria-label="Choose voice agent"
            >
              <button
                type="button"
                role="tab"
                aria-selected={!isRealEstate}
                disabled={active}
                onClick={() => switchAgent("corporate")}
                className={`rounded-lg px-2.5 py-2 text-[12px] font-semibold transition disabled:opacity-50 ${
                  !isRealEstate
                    ? "bg-primary text-bg shadow-sm"
                    : "text-muted hover:bg-bg hover:text-text"
                }`}
              >
                Tuvi agent
              </button>
              <button
                type="button"
                role="tab"
                aria-selected={isRealEstate}
                disabled={active}
                onClick={() => switchAgent("real_estate")}
                className={`rounded-lg px-2.5 py-2 text-[12px] font-semibold transition disabled:opacity-50 ${
                  isRealEstate
                    ? "bg-primary text-bg shadow-sm"
                    : "text-muted hover:bg-bg hover:text-text"
                }`}
              >
                Real estate
              </button>
            </div>

            <p className="mt-2 text-[12px] leading-snug text-muted">
              {isRealEstate
                ? "Ananya — property search, listings & callbacks."
                : "Software consultations & booking."}
            </p>

            {isRealEstate && !active ? (
              <label className="mt-2.5 flex items-center gap-2 text-[11px] text-muted">
                <span className="shrink-0 font-semibold uppercase tracking-wider">Lang</span>
                <select
                  value={language}
                  onChange={(e) => setLanguage(e.target.value as RealEstateLanguage)}
                  className="min-w-0 flex-1 rounded-lg border border-border bg-bg px-2.5 py-2 text-[12px] font-medium text-text outline-none focus:border-primary/50"
                  aria-label="Real estate agent language"
                >
                  <option value="en">English</option>
                  <option value="hi">Hindi</option>
                  <option value="te">Telugu</option>
                  <option value="auto">Auto-detect</option>
                </select>
              </label>
            ) : null}

            {error && !showCallback && (
              <div
                className="mt-3 rounded-xl border border-danger/35 bg-danger/10 px-3 py-2.5"
                role="alert"
              >
                <p className="text-[11px] font-bold uppercase tracking-wider text-danger">
                  Configuration error
                </p>
                <p className="mt-1 text-sm leading-relaxed text-text">{error}</p>
              </div>
            )}

            {showCallback ? (
              <div className="mt-4 space-y-3">
                <div className="flex items-center justify-between gap-2">
                  <div>
                    <p className="text-sm font-semibold text-text">Call me</p>
                    <p className="mt-0.5 text-xs text-muted">
                      {isRealEstate
                        ? "Drop your number — the real estate agent will dial you."
                        : "Drop your number — our AI will dial you now."}
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() => setShowCallback(false)}
                    className="rounded-lg p-1.5 text-muted transition hover:bg-surface hover:text-text"
                    aria-label="Close callback form"
                  >
                    <CloseIcon />
                  </button>
                </div>
                <RequestCallbackForm
                  compact
                  agent={agentKind}
                  language={language}
                  onSuccess={() => {
                    window.setTimeout(() => setShowCallback(false), 1800);
                  }}
                />
              </div>
            ) : (
              <div className="mt-4 flex flex-col items-center">
                <VoiceOrbCanvas status={status} size={132} className="mx-auto bg-transparent" />

                <div
                  className="mt-2 inline-flex items-center gap-2 rounded-full border border-border bg-surface px-3 py-1.5"
                  style={{ borderColor: `${glow}55` }}
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
                    onClick={() => setShowCallback(true)}
                    className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full border border-border bg-surface text-text transition hover:border-primary/40 hover:text-primary"
                    aria-label="Get a callback"
                    title="Get a callback"
                  >
                    <PhoneIcon />
                  </button>

                  <button
                    type="button"
                    onMouseDown={() => void preloadWorklet()}
                    onClick={() => void handlePrimaryAction()}
                    className="flex h-11 flex-1 cursor-pointer items-center justify-center gap-2 rounded-full bg-primary px-4 text-sm font-semibold text-bg transition-colors hover:bg-primary-dim"
                  >
                    <MicIcon />
                    {active ? "End" : error ? "Check again" : "Start talking"}
                  </button>

                  {error && !active && (
                    <button
                      type="button"
                      onClick={reset}
                      className="h-11 shrink-0 rounded-full border border-border px-3 text-sm text-muted transition hover:bg-surface hover:text-text"
                    >
                      Reset
                    </button>
                  )}
                </div>
              </div>
            )}
          </div>
        )}

        <button
          type="button"
          onClick={() => (open ? closePanel() : handlePrimaryAction())}
          onMouseDown={() => void preloadWorklet()}
          className="pointer-events-auto flex cursor-pointer items-center gap-2.5 rounded-full bg-white px-5 py-3.5 text-sm font-semibold text-black shadow-[0_12px_34px_rgba(0,0,0,0.35)] transition-colors hover:bg-white/90"
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
