"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import type {
  ConsultationCalendarMonth,
  ConsultationCalendarSlot,
} from "@/lib/types";
import { EmptyState, ErrorBanner, PageHeader } from "@/components/ui";

const WEEKDAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
const DISPLAY_TIMEZONE = "Australia/Sydney";

function currentMonth(timezone = DISPLAY_TIMEZONE): string {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
  }).formatToParts(new Date());
  const year = parts.find((part) => part.type === "year")?.value;
  const month = parts.find((part) => part.type === "month")?.value;
  return year && month ? `${year}-${month}` : new Date().toISOString().slice(0, 7);
}

function addMonths(month: string, amount: number): string {
  const [year, monthNumber] = month.split("-").map(Number);
  const date = new Date(Date.UTC(year, monthNumber - 1 + amount, 1));
  return `${date.getUTCFullYear()}-${String(date.getUTCMonth() + 1).padStart(2, "0")}`;
}

function monthLabel(month: string): string {
  const [year, monthNumber] = month.split("-").map(Number);
  return new Intl.DateTimeFormat("en-AU", {
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  }).format(new Date(Date.UTC(year, monthNumber - 1, 1)));
}

function dateLabel(date: string): string {
  const [year, month, day] = date.split("-").map(Number);
  return new Intl.DateTimeFormat("en-AU", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  }).format(new Date(Date.UTC(year, month - 1, day)));
}

function timeLabel(time: string): string {
  const [hour, minute] = time.split(":").map(Number);
  const date = new Date(Date.UTC(2000, 0, 1, hour, minute));
  return new Intl.DateTimeFormat("en-AU", {
    hour: "numeric",
    minute: "2-digit",
    timeZone: "UTC",
  }).format(date);
}

function monthCells(month: string): Array<string | null> {
  const [year, monthNumber] = month.split("-").map(Number);
  const first = new Date(Date.UTC(year, monthNumber - 1, 1));
  const leading = (first.getUTCDay() + 6) % 7;
  const count = new Date(Date.UTC(year, monthNumber, 0)).getUTCDate();
  const cells: Array<string | null> = Array.from({ length: leading }, () => null);
  for (let day = 1; day <= count; day += 1) {
    cells.push(`${month}-${String(day).padStart(2, "0")}`);
  }
  while (cells.length % 7 !== 0) cells.push(null);
  return cells;
}

function isOnGrid(slot: ConsultationCalendarSlot): boolean {
  return slot.on_grid === true;
}

function requestStatus(error: unknown): number | undefined {
  if (typeof error !== "object" || error === null || !("status" in error)) return undefined;
  return typeof error.status === "number" ? error.status : undefined;
}

function Chevron({ direction }: { direction: "left" | "right" }) {
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d={direction === "left" ? "m15 18-6-6 6-6" : "m9 18 6-6-6-6"} />
    </svg>
  );
}

export default function ConsultationCalendarPage() {
  const [month, setMonth] = useState(() => currentMonth());
  const [calendar, setCalendar] = useState<ConsultationCalendarMonth | null>(null);
  const [draft, setDraft] = useState<Record<string, boolean>>({});
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const loadGeneration = useRef(0);

  const loadMonth = useCallback(async (targetMonth: string) => {
    const generation = ++loadGeneration.current;
    setLoading(true);
    setCalendar(null);
    setDraft({});
    setSelectedDate(null);
    setError(null);
    setMessage(null);
    try {
      const data = await adminFetch<ConsultationCalendarMonth>(
        `admin/consultation-calendar/${targetMonth}`,
      );
      const nextDraft = Object.fromEntries(
        (data.slots || [])
          .filter(isOnGrid)
          .map((slot) => [slot.iso, slot.is_available]),
      );
      const dates = [...new Set((data.slots || []).map((slot) => slot.date))];
      const firstFuture = dates.find((date) =>
        data.slots.some((slot) => slot.date === date && !slot.past),
      );
      if (generation !== loadGeneration.current) return;
      setCalendar(data);
      setDraft(nextDraft);
      setSelectedDate(firstFuture || dates[0] || `${targetMonth}-01`);
    } catch (err) {
      if (generation !== loadGeneration.current) return;
      setError(err instanceof Error ? err.message : "Failed to load consultation calendar");
    } finally {
      if (generation === loadGeneration.current) setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadMonth(month);
    return () => {
      loadGeneration.current += 1;
    };
  }, [loadMonth, month]);

  const dirty = useMemo(
    () =>
      Boolean(
        calendar?.slots.some(
          (slot) =>
            isOnGrid(slot) &&
            !slot.past &&
            draft[slot.iso] !== slot.is_available,
        ),
      ),
    [calendar, draft],
  );

  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  const slotsByDate = useMemo(() => {
    const grouped = new Map<string, ConsultationCalendarSlot[]>();
    for (const slot of calendar?.slots || []) {
      const list = grouped.get(slot.date) || [];
      list.push(slot);
      grouped.set(slot.date, list);
    }
    for (const slots of grouped.values()) {
      slots.sort((a, b) => a.iso.localeCompare(b.iso));
    }
    return grouped;
  }, [calendar]);

  const selectedSlots = selectedDate ? slotsByDate.get(selectedDate) || [] : [];
  const editableSelectedSlots = selectedSlots.filter(
    (slot) => isOnGrid(slot) && !slot.past && !slot.booked,
  );

  const summary = useMemo(() => {
    let open = 0;
    let closed = 0;
    for (const slot of calendar?.slots || []) {
      if (!isOnGrid(slot) || slot.past || slot.booked) continue;
      if (draft[slot.iso]) open += 1;
      else closed += 1;
    }
    return { open, booked: calendar?.booked_call_count || 0, closed };
  }, [calendar, draft]);

  function navigate(target: string) {
    if (loading || saving) return;
    if (dirty && !window.confirm("Discard the unsaved availability changes for this month?")) {
      return;
    }
    setLoading(true);
    setCalendar(null);
    setDraft({});
    setSelectedDate(null);
    setMonth(target);
  }

  function toggleSlot(slot: ConsultationCalendarSlot) {
    if (loading || saving || slot.past || slot.booked || !isOnGrid(slot)) return;
    setDraft((current) => ({ ...current, [slot.iso]: !current[slot.iso] }));
    setMessage(null);
  }

  function setSelectedDayAvailability(isAvailable: boolean) {
    if (loading || saving) return;
    setDraft((current) => {
      const next = { ...current };
      for (const slot of editableSelectedSlots) next[slot.iso] = isAvailable;
      return next;
    });
    setMessage(null);
  }

  async function saveMonth() {
    if (!calendar || loading || saving) return;
    const generation = loadGeneration.current;
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      const slots = calendar.slots
        .filter((slot) => isOnGrid(slot) && !slot.past)
        .map((slot) => ({
          iso: slot.iso,
          is_available: slot.booked ? slot.is_available : Boolean(draft[slot.iso]),
        }));
      const saved = await adminFetch<ConsultationCalendarMonth>(
        `admin/consultation-calendar/${calendar.month}`,
        {
          method: "PUT",
          body: { expected_revision: calendar.revision, slots },
        },
      );
      if (generation !== loadGeneration.current) return;
      setCalendar(saved);
      setDraft(
        Object.fromEntries(
          saved.slots.filter(isOnGrid).map((slot) => [slot.iso, slot.is_available]),
        ),
      );
      setMessage(`${monthLabel(saved.month)} availability saved to the scheduling database.`);
    } catch (err) {
      if (generation !== loadGeneration.current) return;
      if (requestStatus(err) === 409) {
        setError(
          "This month changed in another admin session. Your edits are still here; refresh the month, then reapply them before saving.",
        );
      } else {
        setError(err instanceof Error ? err.message : "Failed to save availability");
      }
    } finally {
      setSaving(false);
    }
  }

  return (
    <div>
      <style>{`
        .consultation-summary {
          display: grid;
          grid-template-columns: repeat(3, minmax(0, 1fr));
          gap: 0.7rem;
          margin-bottom: 1rem;
        }
        .consultation-layout {
          display: grid;
          grid-template-columns: minmax(0, 1.35fr) minmax(18rem, 0.65fr);
          gap: 1rem;
          align-items: start;
        }
        .calendar-weekdays,
        .calendar-grid {
          display: grid;
          grid-template-columns: repeat(7, minmax(0, 1fr));
        }
        .calendar-weekdays span {
          padding: 0.45rem 0.25rem;
          color: var(--muted);
          font-size: 0.7rem;
          font-weight: 700;
          letter-spacing: 0.04em;
          text-align: center;
          text-transform: uppercase;
        }
        .calendar-day,
        .calendar-blank {
          min-height: 6.25rem;
          border-right: 1px solid var(--line);
          border-top: 1px solid var(--line);
        }
        .calendar-day:nth-child(7n),
        .calendar-blank:nth-child(7n) { border-right: 0; }
        .calendar-day {
          appearance: none;
          background: white;
          border-bottom: 0;
          border-left: 0;
          color: var(--ink);
          cursor: pointer;
          font: inherit;
          padding: 0.55rem;
          text-align: left;
          transition: background 160ms ease, box-shadow 160ms ease;
        }
        .calendar-day:hover:not(:disabled) { background: var(--accent-soft); }
        .calendar-day:focus-visible {
          outline: 3px solid color-mix(in srgb, var(--accent) 35%, transparent);
          outline-offset: -3px;
        }
        .calendar-day[data-selected="true"] {
          background: var(--accent-soft);
          box-shadow: inset 0 0 0 2px var(--accent);
        }
        .calendar-day-counts {
          display: grid;
          gap: 0.18rem;
          margin-top: 0.65rem;
          font-size: 0.72rem;
          line-height: 1.25;
        }
        .slot-grid {
          display: grid;
          grid-template-columns: repeat(2, minmax(0, 1fr));
          gap: 0.5rem;
        }
        .slot-toggle {
          display: grid;
          gap: 0.15rem;
          min-height: 3.55rem;
          padding: 0.65rem 0.7rem;
          border: 1px solid var(--line);
          background: white;
          color: var(--ink);
          cursor: pointer;
          font: inherit;
          text-align: left;
          transition: background 160ms ease, border-color 160ms ease;
        }
        .slot-toggle:hover:not(:disabled) { border-color: var(--accent); }
        .slot-toggle:focus-visible {
          outline: 2px solid var(--accent);
          outline-offset: 2px;
        }
        .slot-toggle[data-available="true"] {
          background: var(--ok-bg);
          border-color: color-mix(in srgb, var(--ok) 45%, var(--line));
        }
        .slot-toggle[data-booked="true"] {
          background: var(--warn-bg);
          border-color: color-mix(in srgb, var(--warn) 45%, var(--line));
        }
        .slot-toggle:disabled { cursor: not-allowed; opacity: 0.78; }
        @media (max-width: 1050px) {
          .consultation-layout { grid-template-columns: 1fr; }
        }
        @media (max-width: 680px) {
          .consultation-summary { grid-template-columns: 1fr; }
          .calendar-day, .calendar-blank { min-height: 4.75rem; }
          .calendar-day { padding: 0.4rem; }
          .calendar-day-counts { font-size: 0.64rem; }
          .slot-grid { grid-template-columns: 1fr; }
        }
        @media (prefers-reduced-motion: reduce) {
          .calendar-day, .slot-toggle { transition: none; }
        }
      `}</style>

      <PageHeader
        title="Consultation calendar"
        subtitle="Set the times the website and corporate voice assistant may offer"
        actions={
          <>
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => navigate(currentMonth(calendar?.timezone || DISPLAY_TIMEZONE))}
              disabled={loading || saving || month === currentMonth(calendar?.timezone || DISPLAY_TIMEZONE)}
            >
              Today
            </button>
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => {
                if (!dirty || window.confirm("Reload this month and discard unsaved changes?")) {
                  void loadMonth(month);
                }
              }}
              disabled={loading || saving}
            >
              Refresh
            </button>
            <button
              type="button"
              className="btn btn-primary"
              onClick={() => void saveMonth()}
              disabled={!calendar || loading || saving || !dirty}
            >
              {saving ? "Saving…" : dirty ? "Save month" : "Saved"}
            </button>
          </>
        }
      />

      <ErrorBanner message={error} />
      {message ? (
        <div className="alert alert-info" style={{ marginBottom: "1rem" }} role="status">
          {message}
        </div>
      ) : null}

      {loading && !calendar ? <EmptyState message="Loading consultation availability…" /> : null}

      {calendar ? (
        <>
          <div className="consultation-summary" aria-label="Month summary">
            <div className="card">
              <div style={{ color: "var(--muted)", fontSize: "0.78rem" }}>Open times</div>
              <div style={{ color: "var(--ok)", fontSize: "1.65rem", fontWeight: 700 }}>
                {summary.open}
              </div>
            </div>
            <div className="card">
              <div style={{ color: "var(--muted)", fontSize: "0.78rem" }}>Calls booked</div>
              <div style={{ color: "var(--warn)", fontSize: "1.65rem", fontWeight: 700 }}>
                {summary.booked}
              </div>
            </div>
            <div className="card">
              <div style={{ color: "var(--muted)", fontSize: "0.78rem" }}>Closed times</div>
              <div style={{ fontSize: "1.65rem", fontWeight: 700 }}>{summary.closed}</div>
            </div>
          </div>

          <div className="consultation-layout">
            <section className="card" style={{ padding: 0, overflow: "hidden" }}>
              <div
                style={{
                  alignItems: "center",
                  display: "flex",
                  gap: "0.75rem",
                  justifyContent: "space-between",
                  padding: "0.85rem 1rem",
                }}
              >
                <button
                  type="button"
                  className="btn btn-secondary"
                  aria-label="Previous month"
                  onClick={() => navigate(addMonths(month, -1))}
                  disabled={loading || saving}
                >
                  <Chevron direction="left" />
                </button>
                <div style={{ textAlign: "center" }}>
                  <h2 style={{ fontSize: "1.15rem", margin: 0 }}>{monthLabel(month)}</h2>
                  <div style={{ color: "var(--muted)", fontSize: "0.76rem", marginTop: "0.15rem" }}>
                    {calendar.timezone} · {calendar.slot_duration_minutes}-minute times
                  </div>
                </div>
                <button
                  type="button"
                  className="btn btn-secondary"
                  aria-label="Next month"
                  onClick={() => navigate(addMonths(month, 1))}
                  disabled={loading || saving}
                >
                  <Chevron direction="right" />
                </button>
              </div>
              <div className="calendar-weekdays" aria-hidden="true">
                {WEEKDAYS.map((weekday) => <span key={weekday}>{weekday}</span>)}
              </div>
              <div className="calendar-grid" aria-label={`${monthLabel(month)} calendar`}>
                {monthCells(month).map((date, index) => {
                  if (!date) return <div className="calendar-blank" key={`blank-${index}`} />;
                  const daySlots = slotsByDate.get(date) || [];
                  const openCount = daySlots.filter(
                    (slot) => isOnGrid(slot) && !slot.past && !slot.booked && draft[slot.iso],
                  ).length;
                  const bookedCount = daySlots.filter((slot) => slot.booked).length;
                  const past = daySlots.length > 0 && daySlots.every((slot) => slot.past);
                  return (
                    <button
                      type="button"
                      className="calendar-day"
                      key={date}
                      data-selected={selectedDate === date}
                      aria-pressed={selectedDate === date}
                      aria-label={`${dateLabel(date)}. ${openCount} open, ${bookedCount} call-blocked times.`}
                      disabled={loading || saving}
                      onClick={() => setSelectedDate(date)}
                    >
                      <span style={{ fontSize: "0.85rem", fontWeight: 700 }}>
                        {Number(date.slice(-2))}
                      </span>
                      <span className="calendar-day-counts">
                        {openCount > 0 ? <span style={{ color: "var(--ok)" }}>{openCount} open</span> : null}
                        {bookedCount > 0 ? <span style={{ color: "var(--warn)" }}>{bookedCount} call-blocked</span> : null}
                        {openCount === 0 && bookedCount === 0 ? (
                          <span style={{ color: "var(--muted)" }}>{past ? "Past" : "No open times"}</span>
                        ) : null}
                      </span>
                    </button>
                  );
                })}
              </div>
            </section>

            <aside className="card">
              <div style={{ display: "grid", gap: "0.3rem" }}>
                <h2 style={{ fontSize: "1.1rem", margin: 0 }}>
                  {selectedDate ? dateLabel(selectedDate) : "Choose a date"}
                </h2>
                <p style={{ color: "var(--muted)", fontSize: "0.82rem", margin: 0 }}>
                  Business window {calendar.business_hour_start}–{calendar.business_hour_end} {calendar.timezone}
                </p>
              </div>

              {editableSelectedSlots.length > 0 ? (
                <div style={{ display: "flex", flexWrap: "wrap", gap: "0.45rem", marginTop: "0.85rem" }}>
                  <button
                    type="button"
                    className="btn btn-secondary"
                    onClick={() => setSelectedDayAvailability(true)}
                    disabled={loading || saving}
                  >
                    Open all
                  </button>
                  <button
                    type="button"
                    className="btn btn-secondary"
                    onClick={() => setSelectedDayAvailability(false)}
                    disabled={loading || saving}
                  >
                    Close all
                  </button>
                </div>
              ) : null}

              <div className="slot-grid" style={{ marginTop: "0.9rem" }}>
                {selectedSlots.map((slot) => {
                  const available = Boolean(draft[slot.iso]);
                  const readOnly = slot.past || slot.booked || !isOnGrid(slot);
                  const status = slot.booked
                    ? isOnGrid(slot) ? "Booked" : "Booked · legacy time"
                    : slot.past
                      ? "Past"
                      : available
                        ? "Open"
                        : "Closed";
                  return (
                    <button
                      type="button"
                      className="slot-toggle"
                      key={slot.iso}
                      data-available={!slot.booked && available}
                      data-booked={slot.booked}
                      aria-pressed={!readOnly ? available : undefined}
                      aria-label={`${timeLabel(slot.time)}. ${status}.`}
                      disabled={readOnly || loading || saving}
                      onClick={() => toggleSlot(slot)}
                    >
                      <strong>{timeLabel(slot.time)}</strong>
                      <span style={{ color: slot.booked ? "var(--warn)" : "var(--muted)", fontSize: "0.74rem" }}>
                        {status}
                      </span>
                    </button>
                  );
                })}
              </div>

              {selectedSlots.length === 0 ? (
                <div style={{ border: "1px dashed var(--line)", color: "var(--muted)", marginTop: "0.9rem", padding: "1rem", textAlign: "center" }}>
                  No configurable times on this date.
                </div>
              ) : null}

              <div className="alert alert-info" style={{ marginTop: "1rem" }}>
                Booked calls are read-only here and are never cancelled by saving availability.
                Unsaved months keep the configured weekday defaults until you save them.
              </div>
            </aside>
          </div>
        </>
      ) : null}
    </div>
  );
}
