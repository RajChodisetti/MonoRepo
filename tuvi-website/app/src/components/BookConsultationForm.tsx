"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import {
  bookConsultation,
  fetchAvailability,
  type BookSuccessResponse,
  type ConsultationSlot,
} from "@/lib/consultationApi";

function formatDayLabel(dateStr: string): string {
  try {
    return new Intl.DateTimeFormat("en-AU", {
      weekday: "long",
      day: "numeric",
      month: "long",
    }).format(new Date(`${dateStr}T12:00:00`));
  } catch {
    return dateStr;
  }
}

function formatTimeLabel(timeStr: string): string {
  const [h, m] = timeStr.split(":").map(Number);
  if (Number.isNaN(h)) return timeStr;
  const d = new Date();
  d.setHours(h, m ?? 0, 0, 0);
  return new Intl.DateTimeFormat("en-AU", {
    hour: "numeric",
    minute: "2-digit",
  }).format(d);
}

function groupSlotsByDate(slots: ConsultationSlot[]): Map<string, ConsultationSlot[]> {
  const map = new Map<string, ConsultationSlot[]>();
  for (const slot of slots) {
    if (!slot.available) continue;
    const list = map.get(slot.date) ?? [];
    list.push(slot);
    map.set(slot.date, list);
  }
  return map;
}

export default function BookConsultationForm() {
  const [slots, setSlots] = useState<ConsultationSlot[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [selected, setSelected] = useState<ConsultationSlot | null>(null);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [success, setSuccess] = useState<BookSuccessResponse | null>(null);

  const loadSlots = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const data = await fetchAvailability(7);
      setSlots(data.slots ?? []);
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : "Could not load slots.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadSlots();
  }, [loadSlots]);

  const grouped = useMemo(() => groupSlotsByDate(slots), [slots]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selected) return;
    setSubmitting(true);
    setSubmitError(null);
    try {
      const result = await bookConsultation({
        date: selected.date,
        time: selected.time,
        prospect_name: name.trim(),
        prospect_email: email.trim(),
      });
      setSuccess(result);
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : "Booking failed.");
    } finally {
      setSubmitting(false);
    }
  };

  if (success) {
    return (
      <div className="rounded-3xl border border-white/10 bg-bg-elevated/80 p-8 text-center shadow-2xl backdrop-blur-xl md:p-10">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-400">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M20 6 9 17l-5-5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </div>
        <p className="mt-5 text-xs font-bold uppercase tracking-[0.16em] text-cyan">
          Consultation booked
        </p>
        <h2 className="mt-2 font-display text-2xl font-bold text-text">You&apos;re all set, {success.prospect_name}!</h2>
        <p className="mt-2 text-sm text-muted">
          Confirmation email sent to {success.prospect_email}
        </p>

        <div className="mt-6 space-y-3 rounded-xl border border-white/10 bg-white/5 px-4 py-4 text-left text-sm">
          <div className="flex justify-between gap-3">
            <span className="text-[11px] font-semibold uppercase tracking-wider text-muted">Day</span>
            <span className="font-medium text-text">{formatDayLabel(success.booking_date)}</span>
          </div>
          <div className="flex justify-between gap-3">
            <span className="text-[11px] font-semibold uppercase tracking-wider text-muted">Time</span>
            <span className="font-medium text-text">{formatTimeLabel(success.booking_time)}</span>
          </div>
        </div>

        <div className="mt-4 rounded-xl border border-gold/30 bg-gold/5 px-4 py-4">
          <p className="text-[11px] font-semibold uppercase tracking-wider text-muted">Confirmation ID</p>
          <p className="mt-2 font-mono text-3xl font-bold tracking-[0.2em] text-gold">
            {success.confirmation_code}
          </p>
        </div>

        <div className="mt-6 flex flex-col gap-3">
          {success.calendar_link ? (
            <a
              href={success.calendar_link}
              target="_blank"
              rel="noopener noreferrer"
              className="rounded-full bg-gradient-to-r from-gold-dim to-gold px-6 py-3 text-sm font-semibold text-bg transition hover:opacity-90"
            >
              Open in Google Calendar
            </a>
          ) : null}
          <Link
            href="/"
            className="rounded-full border border-white/15 px-6 py-3 text-sm font-medium text-muted transition hover:text-text"
          >
            Back to home
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="relative rounded-3xl border border-white/10 bg-bg-elevated/80 p-6 shadow-2xl backdrop-blur-xl md:p-8">
      {submitting && (
        <div className="absolute inset-0 z-10 flex items-center justify-center rounded-3xl bg-bg/75 backdrop-blur-md">
          <div className="text-center">
            <div className="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-cyan border-t-transparent" />
            <p className="mt-3 text-sm font-semibold text-text">Booking your slot…</p>
          </div>
        </div>
      )}

      <p className="text-xs font-bold uppercase tracking-[0.16em] text-cyan">Free consultation</p>
      <h1 className="mt-2 font-display text-2xl font-bold text-text md:text-3xl">
        Pick a time that works for you
      </h1>
      <p className="mt-2 text-sm text-muted">
        30-minute call · Australia/Sydney business hours · Synced to Google Calendar
      </p>

      {loadError && (
        <div className="mt-5 rounded-xl border border-red-400/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">
          {loadError}
          <button
            type="button"
            onClick={() => void loadSlots()}
            className="ml-2 underline hover:no-underline"
          >
            Retry
          </button>
        </div>
      )}

      {loading ? (
        <div className="mt-8 flex items-center justify-center gap-3 py-16 text-muted">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-cyan border-t-transparent" />
          <span className="text-sm">Loading available slots…</span>
        </div>
      ) : grouped.size === 0 ? (
        <p className="mt-8 py-12 text-center text-sm text-muted">
          No slots available right now. Please check back soon or email us.
        </p>
      ) : (
        <div className="mt-6 max-h-[340px] space-y-5 overflow-y-auto pr-1">
          {[...grouped.entries()].map(([date, daySlots]) => (
            <div key={date}>
              <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted">
                {formatDayLabel(date)}
              </p>
              <div className="flex flex-wrap gap-2">
                {daySlots.map((slot) => {
                  const isSelected =
                    selected?.date === slot.date && selected?.time === slot.time;
                  return (
                    <button
                      key={slot.iso}
                      type="button"
                      onClick={() => {
                        setSelected(slot);
                        setSubmitError(null);
                      }}
                      className={`rounded-xl border px-4 py-2.5 text-sm font-medium transition ${
                        isSelected
                          ? "border-gold bg-gold/15 text-gold"
                          : "border-white/10 bg-white/5 text-text hover:border-cyan/40 hover:text-cyan"
                      }`}
                    >
                      {formatTimeLabel(slot.time)}
                    </button>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      )}

      {selected && (
        <form onSubmit={(e) => void handleSubmit(e)} className="mt-6 border-t border-white/10 pt-6">
          <p className="text-sm text-muted">
            Selected:{" "}
            <span className="font-medium text-text">
              {formatDayLabel(selected.date)} at {formatTimeLabel(selected.time)}
            </span>
          </p>

          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <label className="block">
              <span className="text-[11px] font-semibold uppercase tracking-wider text-muted">
                Your name
              </span>
              <input
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="mt-1.5 w-full rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-text outline-none transition focus:border-cyan/50"
                placeholder="Jane Smith"
              />
            </label>
            <label className="block">
              <span className="text-[11px] font-semibold uppercase tracking-wider text-muted">
                Email
              </span>
              <input
                required
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1.5 w-full rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-text outline-none transition focus:border-cyan/50"
                placeholder="you@company.com"
              />
            </label>
          </div>

          {submitError && (
            <p className="mt-3 text-sm text-red-300" role="alert">
              {submitError}
            </p>
          )}

          <button
            type="submit"
            disabled={submitting || !name.trim() || !email.trim()}
            className="mt-5 w-full rounded-full bg-gradient-to-r from-gold-dim to-gold px-6 py-3.5 text-sm font-semibold text-bg transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Confirm booking
          </button>
        </form>
      )}
    </div>
  );
}
