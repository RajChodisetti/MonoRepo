"use client";

import { useEffect, useRef, useState } from "react";
import type { RestaurantContent } from "@/data/types/restaurant";

const DEFAULT_RESERVATION_TIMEZONE = "Australia/Sydney";

type AvailabilityResponse = {
  available_slots?: unknown;
  timezone?: string;
  date?: string;
  party_size?: number;
  error?: { message?: string };
};

type RequestIdentity = {
  fingerprint: string;
  clientRequestId: string;
};

type ReservationStatus = "pending" | "confirmed" | "cancelled";

function apiBase(): string {
  return process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "";
}

function dateInTimeZone(timeZone: string): string {
  const parts = new Intl.DateTimeFormat("en-AU", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(new Date());
  const values = Object.fromEntries(parts.map((part) => [part.type, part.value]));
  return `${values.year}-${values.month}-${values.day}`;
}

function formatSlot(slot: string, timeZone: string): string {
  const value = new Date(slot);
  if (Number.isNaN(value.getTime())) return slot;
  return new Intl.DateTimeFormat("en-AU", {
    timeZone,
    hour: "numeric",
    minute: "2-digit",
  }).format(value);
}

function newClientRequestId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return `web-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

function isReservationStatus(value: string | undefined): value is ReservationStatus {
  return value === "pending" || value === "confirmed" || value === "cancelled";
}

function reservationHeading(status: ReservationStatus | undefined): string {
  if (status === "confirmed") return "Reservation Confirmed";
  if (status === "cancelled") return "Reservation Request Updated";
  return "Reservation Request Received";
}

function reservationFallbackMessage(status: ReservationStatus): string {
  if (status === "confirmed") return "The restaurant has confirmed your reservation.";
  if (status === "cancelled") return "This reservation request has been cancelled.";
  return "Your reservation request is pending confirmation by the restaurant.";
}

export default function ElysianReservation({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<{
    status: ReservationStatus;
    reservationId: string;
    message: string;
  } | null>(null);
  const [showForm, setShowForm] = useState(true);
  const [selectedDate, setSelectedDate] = useState("");
  const [partySize, setPartySize] = useState(2);
  const [selectedSlot, setSelectedSlot] = useState("");
  const [availableSlots, setAvailableSlots] = useState<string[]>([]);
  const [availabilityTimeZone, setAvailabilityTimeZone] = useState(
    DEFAULT_RESERVATION_TIMEZONE,
  );
  const [availabilityLoading, setAvailabilityLoading] = useState(false);
  const [availabilityError, setAvailabilityError] = useState<string | null>(null);
  const requestIdentityRef = useRef<RequestIdentity | null>(null);

  const minDate = dateInTimeZone(DEFAULT_RESERVATION_TIMEZONE);
  const baseURL = apiBase();

  useEffect(() => {
    setSelectedSlot("");
    setAvailableSlots([]);
    setAvailabilityError(null);

    if (!selectedDate) {
      setAvailabilityLoading(false);
      return;
    }
    if (!baseURL) {
      setAvailabilityLoading(false);
      setAvailabilityError("API URL is not configured.");
      return;
    }

    const controller = new AbortController();
    const query = new URLSearchParams({
      date: selectedDate,
      party_size: String(partySize),
    });
    setAvailabilityLoading(true);

    void fetch(
      `${baseURL}/api/public/v1/restaurants/${restaurant.restaurantId}/table-availability?${query.toString()}`,
      { cache: "no-store", signal: controller.signal },
    )
      .then(async (response) => {
        const payload = (await response.json().catch(() => ({}))) as AvailabilityResponse;
        if (!response.ok) {
          throw new Error(payload.error?.message || "Could not load available times.");
        }
        if (!Array.isArray(payload.available_slots)) {
          throw new Error("Availability response did not include any slots.");
        }
        return {
          slots: payload.available_slots.filter(
            (slot): slot is string => typeof slot === "string" && slot.trim() !== "",
          ),
          timeZone: payload.timezone?.trim() || DEFAULT_RESERVATION_TIMEZONE,
        };
      })
      .then(({ slots, timeZone }) => {
        if (controller.signal.aborted) return;
        setAvailableSlots(slots);
        setAvailabilityTimeZone(timeZone);
      })
      .catch((reason: unknown) => {
        if (controller.signal.aborted) return;
        setAvailabilityError(
          reason instanceof Error ? reason.message : "Could not load available times.",
        );
      })
      .finally(() => {
        if (!controller.signal.aborted) setAvailabilityLoading(false);
      });

    return () => controller.abort();
  }, [baseURL, partySize, restaurant.restaurantId, selectedDate]);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    const fd = new FormData(e.currentTarget);
    const name = String(fd.get("name") || "").trim();
    const email = String(fd.get("email") || "").trim();
    const phone = String(fd.get("phone") || "").trim();
    const note = String(fd.get("note") || "").trim();

    if (!selectedSlot || !availableSlots.includes(selectedSlot)) {
      setError("Please select one of the currently available times.");
      setLoading(false);
      return;
    }

    if (!baseURL) {
      setError("API URL is not configured.");
      setLoading(false);
      return;
    }

    try {
      const requestBody: {
        guest_name: string;
        guest_phone: string;
        guest_email?: string;
        party_size: number;
        slot: string;
        source: "web_form";
        notes?: string;
        client_request_id?: string;
      } = {
        guest_name: name,
        guest_phone: phone,
        party_size: partySize,
        slot: selectedSlot,
        source: "web_form",
      };
      if (email) requestBody.guest_email = email;
      if (note) requestBody.notes = note;

      const fingerprint = JSON.stringify(requestBody);
      if (!requestIdentityRef.current || requestIdentityRef.current.fingerprint !== fingerprint) {
        requestIdentityRef.current = {
          fingerprint,
          clientRequestId: newClientRequestId(),
        };
      }
      requestBody.client_request_id = requestIdentityRef.current.clientRequestId;

      const res = await fetch(
        `${baseURL}/api/public/v1/restaurants/${restaurant.restaurantId}/reservations`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(requestBody),
        },
      );
      const data = (await res.json().catch(() => ({}))) as {
        status?: string;
        reservation_id?: string;
        message?: string;
        error?: { message?: string };
      };
      if (!res.ok) {
        throw new Error(
          data.error?.message || data.message || "Could not submit the reservation request.",
        );
      }
      if (!isReservationStatus(data.status) || !data.reservation_id) {
        throw new Error("The reservation request response was incomplete.");
      }
      setSuccess({
        status: data.status,
        reservationId: data.reservation_id,
        message: data.message || reservationFallbackMessage(data.status),
      });
      setShowForm(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Reservation request failed.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="reservation section" id="reservation">
      <div className="reservation-bg" />
      <div className="container reservation-grid">
        <div className="reservation-copy reveal fade-up">
          <p className="eyebrow">Request Your Evening</p>
          <h2 className="section-title">
            Request a Table at <span className="gold-text">Our Restaurant</span>
          </h2>
          <p className="section-sub">
            Send a reservation request to {restaurant.name}. The restaurant will confirm
            availability with you directly.
          </p>
          <ul className="reservation-info">
            {restaurant.phone ? <li>Call {restaurant.phone} for same-day requests</li> : null}
            <li>Your table is not confirmed until the restaurant contacts you</li>
            {restaurant.email ? <li>Email {restaurant.email} for private dining inquiries</li> : null}
          </ul>
        </div>

        {showForm ? (
          <form
            className="reservation-form glass reveal fade-up"
            id="reservationForm"
            onSubmit={handleSubmit}
          >
            {error ? (
              <p className="menu-empty" style={{ padding: "0 0 16px", textAlign: "left" }}>
                {error}
              </p>
            ) : null}
            <div className="form-row">
              <div className="form-field">
                <label htmlFor="resName">Full Name</label>
                <input type="text" id="resName" name="name" required placeholder="Your name" />
              </div>
              <div className="form-field">
                <label htmlFor="resEmail">Email</label>
                <input type="email" id="resEmail" name="email" placeholder="you@email.com" />
              </div>
            </div>
            <div className="form-row">
              <div className="form-field">
                <label htmlFor="resDate">Date</label>
                <input
                  type="date"
                  id="resDate"
                  name="date"
                  required
                  min={minDate}
                  value={selectedDate}
                  onChange={(event) => {
                    setSelectedSlot("");
                    setAvailableSlots([]);
                    setSelectedDate(event.target.value);
                  }}
                />
              </div>
              <div className="form-field">
                <label htmlFor="resTime">Time</label>
                <select
                  id="resTime"
                  name="time"
                  required
                  value={selectedSlot}
                  disabled={!selectedDate || availabilityLoading || availableSlots.length === 0}
                  onChange={(event) => setSelectedSlot(event.target.value)}
                >
                  <option value="">
                    {!selectedDate
                      ? "Select a date first"
                      : availabilityLoading
                        ? "Loading times..."
                        : availableSlots.length === 0
                          ? "No times available"
                          : "Select time"}
                  </option>
                  {availableSlots.map((slot) => (
                    <option key={slot} value={slot}>
                      {formatSlot(slot, availabilityTimeZone)}
                    </option>
                  ))}
                </select>
                {availabilityError ? (
                  <p className="menu-empty" style={{ padding: "8px 0 0", textAlign: "left" }}>
                    {availabilityError}
                  </p>
                ) : null}
              </div>
            </div>
            <div className="form-row">
              <div className="form-field">
                <label htmlFor="resGuests">Guests</label>
                <select
                  id="resGuests"
                  name="guests"
                  required
                  value={partySize}
                  onChange={(event) => {
                    setSelectedSlot("");
                    setAvailableSlots([]);
                    setPartySize(Number(event.target.value));
                  }}
                >
                  <option value="1">1 Guest</option>
                  <option value="2">2 Guests</option>
                  <option value="3">3 Guests</option>
                  <option value="4">4 Guests</option>
                  <option value="5">5 Guests</option>
                  <option value="6">6 Guests</option>
                </select>
              </div>
              <div className="form-field">
                <label htmlFor="resPhone">Phone</label>
                <input type="tel" id="resPhone" name="phone" required placeholder="+61..." />
              </div>
            </div>
            <div className="form-field">
              <label htmlFor="resNote">Special Request</label>
              <textarea id="resNote" name="note" rows={3} placeholder="Allergies, seating preference..." />
            </div>
            <button
              type="submit"
              className={`btn btn-gold btn-lg ripple form-submit${loading ? " loading" : ""}`}
              disabled={loading || availabilityLoading || !selectedSlot}
            >
              <span className="btn-text">
                {loading ? "Sending Request..." : "Request Reservation"}
              </span>
            </button>
          </form>
        ) : null}

        <div className={`reservation-success${success ? " show" : ""}`} id="reservationSuccess">
          <div className="success-check">
            <svg viewBox="0 0 52 52">
              <circle cx="26" cy="26" r="24" />
              <path d="M14 27l7 7 17-17" />
            </svg>
          </div>
          <h3>{reservationHeading(success?.status)}</h3>
          <p>{success?.message}</p>
          <p>
            Status: <strong style={{ color: "var(--gold)" }}>{success?.status}</strong>
            <br />
            Request ID: {success?.reservationId}
          </p>
          <button
            type="button"
            className="btn btn-ghost"
            id="resNewBtn"
            onClick={() => {
              requestIdentityRef.current = null;
              setSuccess(null);
              setShowForm(true);
              setError(null);
              setSelectedDate("");
              setPartySize(2);
              setSelectedSlot("");
              setAvailableSlots([]);
              setAvailabilityError(null);
            }}
          >
            Submit Another Request
          </button>
        </div>
      </div>
    </section>
  );
}
