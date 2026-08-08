"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import {
  bookConsultation,
  fetchAvailability,
  type BookSuccessResponse,
  type ConsultationSlot,
} from "@/lib/consultationApi";

function formatDayLabel(date: string): string {
  try {
    return new Intl.DateTimeFormat("en-AU", {
      weekday: "long",
      day: "numeric",
      month: "long",
    }).format(new Date(`${date}T12:00:00`));
  } catch {
    return date;
  }
}

function formatTimeLabel(time: string): string {
  const [hours, minutes] = time.split(":").map(Number);
  if (Number.isNaN(hours)) return time;
  const value = new Date();
  value.setHours(hours, minutes || 0, 0, 0);
  return new Intl.DateTimeFormat("en-AU", {
    hour: "numeric",
    minute: "2-digit",
  }).format(value);
}

function groupSlots(slots: ConsultationSlot[]): Map<string, ConsultationSlot[]> {
  const grouped = new Map<string, ConsultationSlot[]>();
  for (const slot of slots) {
    if (!slot.available) continue;
    grouped.set(slot.date, [...(grouped.get(slot.date) || []), slot]);
  }
  return grouped;
}

export default function BookConsultationForm() {
  const [slots, setSlots] = useState<ConsultationSlot[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [selected, setSelected] = useState<ConsultationSlot | null>(null);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [success, setSuccess] = useState<BookSuccessResponse | null>(null);

  const loadSlots = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const data = await fetchAvailability();
      setSlots(data.slots || []);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Could not load available times.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    let cancelled = false;

    void fetchAvailability()
      .then((data) => {
        if (!cancelled) setSlots(data.slots || []);
      })
      .catch((error: unknown) => {
        if (!cancelled) {
          setLoadError(error instanceof Error ? error.message : "Could not load available times.");
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const grouped = useMemo(() => groupSlots(slots), [slots]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selected) return;

    setSubmitting(true);
    setSubmitError(null);
    try {
      const result = await bookConsultation({
        date: selected.date,
        time: selected.time,
        prospect_name: name.trim(),
        prospect_email: email.trim(),
        prospect_phone: phone.trim(),
      });
      setSuccess(result);
    } catch (error) {
      setSubmitError(error instanceof Error ? error.message : "Booking failed.");
      setSelected(null);
      await loadSlots();
    } finally {
      setSubmitting(false);
    }
  }

  if (success) {
    return (
      <section className="rounded-[2rem] border border-border bg-bg p-7 text-center shadow-[0_24px_80px_rgba(15,39,31,0.12)] sm:p-10" aria-live="polite">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-sage text-primary">
          <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" aria-hidden="true">
            <path d="M20 6 9 17l-5-5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </div>
        <p className="mt-5 text-xs font-semibold uppercase tracking-[0.16em] text-primary">
          Consultation booked
        </p>
        <h2 className="mt-2 font-display text-3xl font-semibold tracking-[-0.03em] text-ink">
          You&apos;re all set, {success.prospect_name}.
        </h2>
        <p className="mx-auto mt-3 max-w-md text-sm leading-6 text-muted">
          Your time is reserved in Tuvi&apos;s booking system. Keep the confirmation ID below for your records.
        </p>

        <dl className="mt-7 space-y-3 rounded-2xl border border-border bg-surface p-5 text-left text-sm">
          <div className="flex justify-between gap-4">
            <dt className="font-medium text-muted">Day</dt>
            <dd className="text-right font-semibold text-ink">{formatDayLabel(success.booking_date)}</dd>
          </div>
          <div className="flex justify-between gap-4">
            <dt className="font-medium text-muted">Time</dt>
            <dd className="text-right font-semibold text-ink">{formatTimeLabel(success.booking_time)}</dd>
          </div>
          <div className="flex justify-between gap-4 border-t border-border pt-3">
            <dt className="font-medium text-muted">Confirmation ID</dt>
            <dd className="font-mono font-bold tracking-[0.12em] text-primary">{success.confirmation_code}</dd>
          </div>
        </dl>

        <Link href="/" className="mt-7 inline-flex rounded-full bg-primary px-6 py-3 text-sm font-semibold text-bg transition-colors hover:bg-primary-dim">
          Back to home
        </Link>
      </section>
    );
  }

  return (
    <section className="relative rounded-[2rem] border border-border bg-bg p-6 shadow-[0_24px_80px_rgba(15,39,31,0.12)] sm:p-8">
      {submitting ? (
        <div className="absolute inset-0 z-10 flex items-center justify-center rounded-[2rem] bg-bg/85 backdrop-blur-sm" aria-live="polite">
          <div className="text-center">
            <div className="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
            <p className="mt-3 text-sm font-semibold text-ink">Reserving your time…</p>
          </div>
        </div>
      ) : null}

      <p className="text-xs font-semibold uppercase tracking-[0.16em] text-primary">Free consultation</p>
      <h2 className="mt-2 font-display text-3xl font-semibold tracking-[-0.03em] text-ink sm:text-4xl">
        Pick a time that works for you
      </h2>
      <p className="mt-3 text-sm leading-6 text-muted">
        30-minute call · Australia/Sydney · availability is managed directly in Tuvi
      </p>

      {loadError ? (
        <div className="mt-6 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-800" role="alert">
          {loadError}
          <button type="button" onClick={() => void loadSlots()} className="ml-2 font-semibold underline underline-offset-2">
            Retry
          </button>
        </div>
      ) : null}

      {loading ? (
        <div className="mt-8 flex items-center justify-center gap-3 py-14 text-muted" aria-live="polite">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <span className="text-sm">Loading available times…</span>
        </div>
      ) : grouped.size === 0 ? (
        <p className="mt-8 rounded-2xl bg-surface px-5 py-12 text-center text-sm leading-6 text-muted">
          No times are available right now. Please check back soon or email contact@tuvisolutions.com.
        </p>
      ) : (
        <div className="mt-7 max-h-[22rem] space-y-6 overflow-y-auto pr-1">
          {[...grouped.entries()].map(([date, daySlots]) => (
            <fieldset key={date}>
              <legend className="mb-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted">
                {formatDayLabel(date)}
              </legend>
              <div className="flex flex-wrap gap-2">
                {daySlots.map((slot) => {
                  const isSelected = selected?.iso === slot.iso;
                  return (
                    <button
                      key={slot.iso}
                      type="button"
                      aria-pressed={isSelected}
                      onClick={() => {
                        setSelected(slot);
                        setSubmitError(null);
                      }}
                      className={`rounded-xl border px-4 py-2.5 text-sm font-semibold transition-colors ${
                        isSelected
                          ? "border-primary bg-primary text-bg"
                          : "border-border bg-surface text-ink hover:border-primary hover:bg-sage"
                      }`}
                    >
                      {formatTimeLabel(slot.time)}
                    </button>
                  );
                })}
              </div>
            </fieldset>
          ))}
        </div>
      )}

      {selected ? (
        <form onSubmit={handleSubmit} className="mt-7 border-t border-border pt-7">
          <p className="text-sm text-muted">
            Selected: <strong className="text-ink">{formatDayLabel(selected.date)} at {formatTimeLabel(selected.time)}</strong>
          </p>

          <div className="mt-5 grid gap-4 sm:grid-cols-2">
            <label className="text-sm font-semibold text-ink">
              Name
              <input
                required
                autoComplete="name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                className="mt-2 w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm font-normal text-ink outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15"
                placeholder="Jane Smith"
              />
            </label>
            <label className="text-sm font-semibold text-ink">
              Email
              <input
                required
                type="email"
                autoComplete="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                className="mt-2 w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm font-normal text-ink outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15"
                placeholder="you@company.com"
              />
            </label>
            <label className="text-sm font-semibold text-ink sm:col-span-2">
              Phone
              <input
                required
                type="tel"
                autoComplete="tel"
                value={phone}
                onChange={(event) => setPhone(event.target.value)}
                className="mt-2 w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm font-normal text-ink outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15"
                placeholder="+61 400 000 000"
              />
            </label>
          </div>

          {submitError ? <p className="mt-4 text-sm text-red-700" role="alert">{submitError}</p> : null}

          <button
            type="submit"
            disabled={submitting || !name.trim() || !email.trim() || !phone.trim()}
            className="mt-6 w-full rounded-full bg-primary px-6 py-3.5 text-sm font-semibold text-bg transition-colors hover:bg-primary-dim disabled:cursor-not-allowed disabled:opacity-50"
          >
            Confirm booking
          </button>
          <p className="mt-3 text-center text-xs leading-5 text-muted">
            Confirmed bookings are stored in Tuvi and immediately removed from availability.
          </p>
        </form>
      ) : null}
    </section>
  );
}
