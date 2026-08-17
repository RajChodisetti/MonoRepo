"use client";

import { useEffect, useRef, useState } from "react";
import type { TemplateId } from "@/lib/templateConfig";
import { TOUR_VOICE_ASSISTANT } from "@/lib/tourTargets";
import { useVoiceAgentSession } from "@/hooks/useVoiceAgentSession";
import VoiceOrbCanvas, { statusGlow } from "@/components/VoiceOrbCanvas";

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

function formatBookingDay(slot: string, bookingDate?: string): string {
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

function formatBookingTime(slot: string, bookingTime?: string): string {
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

export default function VoiceAssistantWidget({
  templateId,
  restaurantIndex = 0,
}: {
  templateId: TemplateId;
  restaurantIndex?: number;
}) {
  const [open, setOpen] = useState(false);
  const [restaurantName, setRestaurantName] = useState<string | undefined>();
  const panelRef = useRef<HTMLDivElement>(null);
  const {
    status,
    error,
    active,
    booking,
    connect,
    disconnect,
    reset,
    dismissBooking,
    prefetchStatus,
    preloadWorklet,
  } = useVoiceAgentSession(restaurantIndex);

  const isAurora = templateId === "2";
  const isElysian = templateId === "3";

  const panelClass = isElysian
    ? "border border-[rgba(212,175,55,0.25)] bg-[#050505]/95 text-[#F8F8F8] shadow-[0_20px_60px_rgba(0,0,0,0.55)] backdrop-blur-xl"
    : isAurora
      ? "border border-white/15 bg-[#09090B]/95 text-white shadow-[0_20px_60px_rgba(0,0,0,0.55)] backdrop-blur-xl"
      : "border border-[#e8e0d4]/20 bg-[#1a1614]/95 text-[#f7f0e6] shadow-[0_20px_60px_rgba(0,0,0,0.45)] backdrop-blur-xl";

  const accentClass = isElysian
    ? "text-[#D4AF37]"
    : isAurora
      ? "text-cyan-300"
      : "text-[#d4a853]";
  const dimClass = isElysian
    ? "text-[#A9A9A9]"
    : isAurora
      ? "text-white/55"
      : "text-[#a89f96]";
  const primaryBtnClass = isElysian
    ? "bg-gradient-to-r from-[#e9d38b] via-[#D4AF37] to-[#a9822a] text-[#0a0a0a] hover:opacity-90"
    : isAurora
      ? "bg-gradient-to-r from-violet-600 to-cyan-500 text-white hover:opacity-90"
      : "bg-gradient-to-r from-[#9a6f2e] to-[#d4a853] text-[#1a1614] hover:opacity-90";
  const ghostBtnClass = isElysian
    ? "border border-[rgba(255,255,255,0.15)] bg-white/5 text-[#F8F8F8] hover:border-[#D4AF37]/50 hover:text-[#D4AF37]"
    : isAurora
      ? "border border-white/15 bg-white/5 text-white hover:border-cyan-400/40 hover:text-cyan-300"
      : "border border-[#e8e0d4]/20 bg-white/5 text-[#f7f0e6] hover:border-[#d4a853]/40 hover:text-[#d4a853]";
  const errorClass = isElysian
    ? "border border-red-400/40 bg-red-500/15 text-red-100"
    : isAurora
      ? "border border-red-400/40 bg-red-500/15 text-red-100"
      : "border border-red-400/35 bg-red-950/50 text-red-100";
  const fabPositionClass =
    isElysian
      ? "bottom-5 right-4 sm:right-5 md:bottom-8 md:right-6"
      : "bottom-20 right-4 md:bottom-6 md:right-6";

  useEffect(() => {
    if (open) void prefetchStatus();
  }, [open, prefetchStatus]);

  useEffect(() => {
    if (!open) return;
    const apiBase = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "");
    if (!apiBase) return;
    let cancelled = false;
    void (async () => {
      try {
        const res = await fetch(`${apiBase}/api/public/v1/site/restaurants/${restaurantIndex}`, {
          cache: "no-store",
        });
        if (!res.ok) return;
        const data = (await res.json()) as { name?: string };
        if (cancelled) return;
        if (data.name) setRestaurantName(data.name);
      } catch {
        // ignore
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [open, restaurantIndex]);

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

  const closePanel = () => {
    if (active) disconnect();
    setOpen(false);
  };

  const glow = statusGlow(status);
  const headerLabel = restaurantName ? `${restaurantName} AI` : "Voice Assistant";

  return (
    <>
      {booking && (
        <div
          className="pointer-events-auto fixed inset-0 z-[90] flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm"
          role="dialog"
          aria-modal="true"
          aria-label="Reservation request receipt"
        >
          <div className={`w-full max-w-sm rounded-2xl p-6 text-center shadow-2xl ${panelClass}`}>
            <p className={`text-xs font-semibold uppercase tracking-[0.16em] ${accentClass}`}>
              Request received
            </p>
            <h2 className="mt-2 text-xl font-semibold">Your reservation request is pending</h2>
            <p className={`mt-2 text-sm ${dimClass}`}>
              {booking.guestName ? `Thanks, ${booking.guestName}. ` : ""}
              {booking.message || "The restaurant will contact you to confirm availability."}
            </p>

            <div
              className={`mt-5 space-y-3 rounded-xl border px-4 py-4 text-left ${
                isAurora ? "border-white/10 bg-white/5" : "border-[#e8e0d4]/15 bg-black/20"
              }`}
            >
              <div className="flex items-start justify-between gap-3">
                <span className={`text-[11px] font-semibold uppercase tracking-[0.12em] ${dimClass}`}>
                  Requested day
                </span>
                <span className="text-right text-sm font-medium">
                  {formatBookingDay(booking.slot, booking.bookingDate)}
                </span>
              </div>
              <div className="flex items-start justify-between gap-3">
                <span className={`text-[11px] font-semibold uppercase tracking-[0.12em] ${dimClass}`}>
                  Requested time
                </span>
                <span className="text-right text-sm font-medium">
                  {formatBookingTime(booking.slot, booking.bookingTime)}
                </span>
              </div>
              <div className="flex items-start justify-between gap-3">
                <span className={`text-[11px] font-semibold uppercase tracking-[0.12em] ${dimClass}`}>
                  Guests
                </span>
                <span className="text-right text-sm font-medium">
                  {booking.partySize} {booking.partySize === 1 ? "person" : "people"}
                </span>
              </div>
            </div>

            <div
              className={`mt-4 rounded-xl border px-4 py-4 ${
                isAurora ? "border-cyan-400/30 bg-white/5" : "border-[#d4a853]/35 bg-black/20"
              }`}
            >
              <p className={`text-[11px] font-semibold uppercase tracking-[0.14em] ${dimClass}`}>
                Request ID
              </p>
              <p className={`mt-2 break-all font-mono text-sm font-bold ${accentClass}`}>
                {booking.reservationId}
              </p>
              <p className={`mt-2 text-xs capitalize ${dimClass}`}>Status: {booking.status}</p>
            </div>

            <button
              type="button"
              onClick={dismissBooking}
              className={`mt-6 w-full rounded-xl px-4 py-3 text-sm font-semibold transition ${primaryBtnClass}`}
            >
              Got it
            </button>
          </div>
        </div>
      )}

      <div className={`pointer-events-none fixed z-[70] ${fabPositionClass}`}>
        {open && (
          <div
            ref={panelRef}
            className={`pointer-events-auto relative mb-3 max-h-[calc(100svh-7rem)] w-[min(440px,calc(100vw-1.5rem))] overflow-y-auto rounded-2xl p-5 ${panelClass}`}
            role="dialog"
            aria-label="AI voice assistant"
          >
            <div className="mb-3 flex items-center justify-between gap-3">
              <p className={`text-xs font-semibold uppercase tracking-wider ${accentClass}`}>
                {headerLabel}
              </p>
              <button
                type="button"
                onClick={closePanel}
                className={`rounded-lg p-1.5 transition hover:bg-white/10 ${dimClass}`}
                aria-label="Close assistant"
              >
                <CloseIcon />
              </button>
            </div>

            {error && (
              <div className={`mb-3 rounded-xl px-3 py-2.5 ${errorClass}`} role="alert">
                <p className="text-[11px] font-bold uppercase tracking-wider text-red-300">
                  Configuration error
                </p>
                <p className="mt-1 text-sm leading-relaxed">{error}</p>
              </div>
            )}

            <div className="flex flex-col items-center py-1">
              <VoiceOrbCanvas status={status} size={176} className="mx-auto bg-transparent" />

              <div
                className="mt-3 inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1.5"
                style={{ borderColor: `${glow}40` }}
              >
                <span
                  className={`h-2 w-2 shrink-0 rounded-full ${
                    status === "thinking" || status === "connecting" ? "animate-pulse" : ""
                  }`}
                  style={{ backgroundColor: glow, boxShadow: `0 0 8px ${glow}` }}
                />
                <span className={`text-[11px] font-semibold uppercase tracking-[0.14em] ${dimClass}`}>
                  {STATUS_LABELS[status] ?? status}
                </span>
              </div>

              <div className="mt-4 flex w-full items-center gap-2">
                <button
                  type="button"
                  onMouseDown={() => void preloadWorklet()}
                  onClick={() => void handlePrimaryAction()}
                  className={`flex h-12 flex-1 items-center justify-center gap-2 rounded-full px-4 text-sm font-semibold transition ${primaryBtnClass}`}
                >
                  <MicIcon />
                  {active ? "End" : error ? "Check again" : "Start talking"}
                </button>

                {error && !active && (
                  <button
                    type="button"
                    onClick={reset}
                    className={`h-12 shrink-0 rounded-full px-3 text-sm transition ${ghostBtnClass}`}
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
          className={`pointer-events-auto flex items-center gap-2.5 rounded-full px-5 py-3.5 text-sm font-semibold shadow-lg transition hover:scale-[1.02] active:scale-[0.98] ${primaryBtnClass}`}
          aria-expanded={open}
          aria-haspopup="dialog"
          data-tour={TOUR_VOICE_ASSISTANT}
        >
          <MicIcon />
          <span>Talk to our AI</span>
        </button>
      </div>
    </>
  );
}
