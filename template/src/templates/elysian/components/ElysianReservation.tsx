"use client";

import { useState } from "react";
import type { RestaurantContent } from "@/data/types/restaurant";

function parseTimeTo24(time: string): { hours: number; minutes: number } | null {
  const m = time.trim().match(/^(\d{1,2}):(\d{2})\s*(AM|PM)$/i);
  if (!m) return null;
  let hours = parseInt(m[1], 10);
  const minutes = parseInt(m[2], 10);
  const ampm = m[3].toUpperCase();
  if (ampm === "PM" && hours !== 12) hours += 12;
  if (ampm === "AM" && hours === 12) hours = 0;
  return { hours, minutes };
}

export default function ElysianReservation({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<{ code: string } | null>(null);
  const [showForm, setShowForm] = useState(true);

  const minDate = new Date().toISOString().split("T")[0];

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    const fd = new FormData(e.currentTarget);
    const name = String(fd.get("name") || "").trim();
    const phone = String(fd.get("phone") || "").trim();
    const date = String(fd.get("date") || "");
    const time = String(fd.get("time") || "");
    const guests = parseInt(String(fd.get("guests") || "2"), 10);

    const parsed = parseTimeTo24(time);
    if (!parsed) {
      setError("Please select a valid time.");
      setLoading(false);
      return;
    }

    const slot = new Date(`${date}T00:00:00`);
    slot.setHours(parsed.hours, parsed.minutes, 0, 0);

    const apiBase = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "");
    if (!apiBase) {
      setError("API URL is not configured.");
      setLoading(false);
      return;
    }

    try {
      const res = await fetch(
        `${apiBase}/api/public/v1/restaurants/${restaurant.restaurantId}/reservations`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            guest_name: name,
            guest_phone: phone,
            party_size: guests,
            slot: slot.toISOString(),
            source: "web_form",
          }),
        },
      );
      const data = (await res.json().catch(() => ({}))) as {
        confirmation_code?: string;
        message?: string;
      };
      if (!res.ok) {
        throw new Error(data.message || "Could not complete reservation.");
      }
      setSuccess({ code: data.confirmation_code || "—" });
      setShowForm(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Reservation failed.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="reservation section" id="reservation">
      <div className="reservation-bg" />
      <div className="container reservation-grid">
        <div className="reservation-copy reveal fade-up">
          <p className="eyebrow">Reserve Your Evening</p>
          <h2 className="section-title">
            Secure Your Seat at <span className="gold-text">Our Table</span>
          </h2>
          <p className="section-sub">
            Book your table at {restaurant.name}. We look forward to welcoming you.
          </p>
          <ul className="reservation-info">
            {restaurant.phone ? <li>Call {restaurant.phone} for same-day requests</li> : null}
            <li>Please arrive within 15 minutes of your reservation time</li>
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
                <input type="date" id="resDate" name="date" required min={minDate} />
              </div>
              <div className="form-field">
                <label htmlFor="resTime">Time</label>
                <select id="resTime" name="time" required defaultValue="">
                  <option value="">Select time</option>
                  <option>6:00 PM</option>
                  <option>6:30 PM</option>
                  <option>7:00 PM</option>
                  <option>7:30 PM</option>
                  <option>8:00 PM</option>
                  <option>8:30 PM</option>
                  <option>9:00 PM</option>
                </select>
              </div>
            </div>
            <div className="form-row">
              <div className="form-field">
                <label htmlFor="resGuests">Guests</label>
                <select id="resGuests" name="guests" required defaultValue="2">
                  <option value="1">1 Guest</option>
                  <option value="2">2 Guests</option>
                  <option value="3">3 Guests</option>
                  <option value="4">4 Guests</option>
                  <option value="5">5 Guests</option>
                  <option value="6">6+ Guests</option>
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
              disabled={loading}
            >
              <span className="btn-text">{loading ? "Confirming..." : "Confirm Reservation"}</span>
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
          <h3>Reservation Confirmed</h3>
          <p>
            Thank you — your confirmation code is{" "}
            <strong style={{ color: "var(--gold)" }}>{success?.code}</strong>.
          </p>
          <button
            type="button"
            className="btn btn-ghost"
            id="resNewBtn"
            onClick={() => {
              setSuccess(null);
              setShowForm(true);
              setError(null);
            }}
          >
            Make Another Reservation
          </button>
        </div>
      </div>
    </section>
  );
}
