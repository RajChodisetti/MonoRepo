"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import type { GalleryImage } from "@/data/types/gallery";
import type { MenuItem } from "@/data/types/menu";
import type { RestaurantContent } from "@/data/types/restaurant";
import PhotoAttribution from "@/components/PhotoAttribution";
import SourceAwareImage, { mediaForURL } from "@/components/SourceAwareImage";
import TemplateSwitchButton from "@/components/TemplateSwitchButton";
import { buildItalianJsonLd } from "./seo";
import "./theme.css";

const DEFAULT_RESERVATION_TIMEZONE = "Australia/Sydney";

type AvailabilityResponse = {
  available_slots?: unknown;
  timezone?: string;
  error?: { message?: string };
};

type ReservationStatus = "pending" | "confirmed" | "cancelled";

type RequestIdentity = {
  fingerprint: string;
  clientRequestId: string;
};

type ItalianImageProps = {
  media: GalleryImage;
  className?: string;
  sizes?: string;
  priority?: boolean;
};

function apiBase(): string {
  return process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "";
}

function imageFromURL(url: string | undefined, alt: string, type: GalleryImage["type"]): GalleryImage | null {
  const value = url?.trim();
  return value ? mediaForURL(value, alt, type) : null;
}

function dedupeImages(images: GalleryImage[]): GalleryImage[] {
  const seen = new Set<string>();
  return images.filter((image) => {
    if (!image.url || seen.has(image.url)) return false;
    seen.add(image.url);
    return true;
  });
}

function allMenuItems(restaurant: RestaurantContent): MenuItem[] {
  return restaurant.menuCategories.flatMap((category) => category.items);
}

function menuImageFor(item: MenuItem, fallback: GalleryImage | undefined): GalleryImage | null {
  return imageFromURL(item.image, item.name, "food") || fallback || null;
}

function formatPriceLevel(value?: string): string {
  if (!value) return "";
  return value.replace(/\s+/g, " ").trim();
}

function buildMapsUrl(restaurant: RestaurantContent): string {
  if (restaurant.coordinates) {
    return `https://www.google.com/maps?q=${restaurant.coordinates.latitude},${restaurant.coordinates.longitude}`;
  }
  return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(restaurant.address)}`;
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
  return `italian-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

function isReservationStatus(value: string | undefined): value is ReservationStatus {
  return value === "pending" || value === "confirmed" || value === "cancelled";
}

function reservationFallbackMessage(status: ReservationStatus): string {
  if (status === "confirmed") return "The restaurant has confirmed your reservation.";
  if (status === "cancelled") return "This reservation request has been cancelled.";
  return "Your reservation request is pending confirmation by the restaurant.";
}

function reservationHeading(status: ReservationStatus | undefined): string {
  if (status === "confirmed") return "Reservation Confirmed";
  if (status === "cancelled") return "Reservation Request Updated";
  return "Reservation Request Received";
}

function ItalianImage({ media, className = "", sizes = "100vw", priority = false }: ItalianImageProps) {
  return (
    <div className={`italian-image ${className}`}>
      <SourceAwareImage
        media={media}
        fill
        sizes={sizes}
        priority={priority}
        className="object-cover"
        fallbackClassName="italian-image-fallback"
      />
      {media.sourceKind === "google_places_live" ? (
        <div className="italian-photo-credit">
          <PhotoAttribution media={media} compact />
        </div>
      ) : null}
    </div>
  );
}

function ItalianNav({ restaurant, showGallery }: { restaurant: RestaurantContent; showGallery: boolean }) {
  const [open, setOpen] = useState(false);
  const links = [
    { href: "#story", label: "Story" },
    { href: "#signature", label: "Signature" },
    { href: "#menu", label: "Menu" },
    ...(showGallery ? [{ href: "#gallery", label: "Gallery" }] : []),
    { href: "#reserve", label: "Reserve" },
    { href: "#visit", label: "Visit" },
  ];

  return (
    <header className="italian-nav">
      <a className="italian-brand" href="#top" aria-label={`${restaurant.name} home`}>
        <span>{restaurant.name}</span>
        <small>{restaurant.locationLabel || restaurant.city}</small>
      </a>

      <nav className="italian-nav-links" aria-label="Restaurant sections">
        {links.map((link) => (
          <a key={link.href} href={link.href}>
            {link.label}
          </a>
        ))}
        <TemplateSwitchButton variant="italian" />
      </nav>

      <div className="italian-nav-actions">
        <a className="italian-nav-reserve" href="#reserve">
          Reserve
        </a>
        <button
          type="button"
          className="italian-menu-button"
          aria-label="Open menu"
          aria-expanded={open}
          onClick={() => setOpen((value) => !value)}
        >
          <span />
          <span />
        </button>
      </div>

      {open ? (
        <div className="italian-mobile-menu">
          {links.map((link) => (
            <a key={link.href} href={link.href} onClick={() => setOpen(false)}>
              {link.label}
            </a>
          ))}
          <TemplateSwitchButton variant="italian" mode="mobile" />
        </div>
      ) : null}
    </header>
  );
}

function ItalianReservationForm({ restaurant }: { restaurant: RestaurantContent }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<{
    status: ReservationStatus;
    reservationId: string;
    message: string;
  } | null>(null);
  const [selectedDate, setSelectedDate] = useState("");
  const [partySize, setPartySize] = useState(2);
  const [selectedSlot, setSelectedSlot] = useState("");
  const [availableSlots, setAvailableSlots] = useState<string[]>([]);
  const [availabilityTimeZone, setAvailabilityTimeZone] = useState(DEFAULT_RESERVATION_TIMEZONE);
  const [availabilityLoading, setAvailabilityLoading] = useState(false);
  const [availabilityError, setAvailabilityError] = useState<string | null>(null);
  const requestIdentityRef = useRef<RequestIdentity | null>(null);
  const baseURL = apiBase();
  const minDate = dateInTimeZone(DEFAULT_RESERVATION_TIMEZONE);

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
        setAvailabilityError(reason instanceof Error ? reason.message : "Could not load available times.");
      })
      .finally(() => {
        if (!controller.signal.aborted) setAvailabilityLoading(false);
      });

    return () => controller.abort();
  }, [baseURL, partySize, restaurant.restaurantId, selectedDate]);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setLoading(true);

    const fd = new FormData(event.currentTarget);
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
        throw new Error(data.error?.message || data.message || "Could not submit the reservation request.");
      }
      if (!isReservationStatus(data.status) || !data.reservation_id) {
        throw new Error("The reservation request response was incomplete.");
      }
      setSuccess({
        status: data.status,
        reservationId: data.reservation_id,
        message: data.message || reservationFallbackMessage(data.status),
      });
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Reservation request failed.");
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <div className="italian-reservation-panel" role="status">
        <p className="italian-kicker">Request received</p>
        <h3>{reservationHeading(success.status)}</h3>
        <p>{success.message}</p>
        <dl className="italian-receipt">
          <div>
            <dt>Status</dt>
            <dd>{success.status}</dd>
          </div>
          <div>
            <dt>Request ID</dt>
            <dd>{success.reservationId}</dd>
          </div>
        </dl>
        <button
          type="button"
          className="italian-button italian-button-secondary"
          onClick={() => {
            requestIdentityRef.current = null;
            setSuccess(null);
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
    );
  }

  return (
    <form className="italian-reservation-panel" onSubmit={handleSubmit}>
      <p className="italian-kicker">Request a table</p>
      <h3>Plan your next evening at {restaurant.name}</h3>
      <p>
        Send a reservation request. Your table is pending until the restaurant confirms availability.
      </p>

      {error ? <p className="italian-form-error">{error}</p> : null}

      <div className="italian-form-grid">
        <label>
          Full Name
          <input type="text" name="name" required placeholder="Your name" />
        </label>
        <label>
          Email
          <input type="email" name="email" placeholder="you@example.com" />
        </label>
        <label>
          Date
          <input
            type="date"
            name="date"
            required
            min={minDate}
            value={selectedDate}
            onChange={(event) => {
              setSelectedDate(event.target.value);
              setSelectedSlot("");
              setAvailableSlots([]);
            }}
          />
        </label>
        <label>
          Time
          <select
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
          {availabilityError ? <span className="italian-field-note">{availabilityError}</span> : null}
        </label>
        <label>
          Guests
          <select
            name="guests"
            required
            value={partySize}
            onChange={(event) => {
              setPartySize(Number(event.target.value));
              setSelectedSlot("");
              setAvailableSlots([]);
            }}
          >
            {[1, 2, 3, 4, 5, 6].map((value) => (
              <option key={value} value={value}>
                {value} {value === 1 ? "Guest" : "Guests"}
              </option>
            ))}
          </select>
        </label>
        <label>
          Phone
          <input type="tel" name="phone" required placeholder="Phone number" />
        </label>
      </div>

      <label className="italian-note-field">
        Special Request
        <textarea name="note" rows={3} placeholder="Allergies, seating preference, occasion" />
      </label>

      <button
        type="submit"
        className="italian-button italian-button-primary"
        disabled={loading || availabilityLoading || !selectedSlot}
      >
        {loading ? "Sending Request..." : "Request Reservation"}
      </button>
    </form>
  );
}

export default function ItalianTemplate({ restaurant }: { restaurant: RestaurantContent }) {
  const menuItems = useMemo(() => allMenuItems(restaurant), [restaurant]);
  const foodImages = restaurant.galleryImages.filter((image) => image.type === "food");
  const ambienceImages = restaurant.galleryImages.filter((image) => image.type === "ambience");
  const heroMedia = restaurant.heroMedia || imageFromURL(restaurant.heroPoster, restaurant.name, "ambience");
  const storyImage = ambienceImages[0] || restaurant.galleryImages[0] || heroMedia;
  const featuredDishes = (restaurant.signatureDishes.length ? restaurant.signatureDishes : menuItems).slice(0, 4);
  const gallery = dedupeImages([
    ...restaurant.galleryImages,
    ...(heroMedia ? [heroMedia] : []),
  ]).slice(0, 8);
  const openHours = Object.entries(restaurant.hours).slice(0, 7);
  const mapsUrl = buildMapsUrl(restaurant);
  const cuisine = restaurant.cuisine || "Italian cuisine";
  const placeLabel = restaurant.city || restaurant.locationLabel || restaurant.address || "the neighborhood";
  const priceLevel = formatPriceLevel(restaurant.priceLevel);
  const jsonLd = buildItalianJsonLd(restaurant);

  return (
    <div className="italian-root" id="top">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <ItalianNav restaurant={restaurant} showGallery={gallery.length > 0} />

      <main>
        <section className="italian-hero">
          {heroMedia ? <ItalianImage media={heroMedia} className="italian-hero-media" priority /> : null}
          <div className="italian-hero-shade" />
          <div className="italian-hero-content">
            <p className="italian-kicker">{cuisine} in {placeLabel}</p>
            <h1>{restaurant.name}</h1>
            <p className="italian-hero-subtitle">{restaurant.subheadline || restaurant.tagline}</p>
            <div className="italian-hero-actions">
              <a className="italian-button italian-button-primary" href="#reserve">
                Reserve a Table
              </a>
              <a className="italian-button italian-button-glass" href="#menu">
                View Menu
              </a>
            </div>
            <dl className="italian-hero-facts">
              {restaurant.rating ? (
                <div>
                  <dt>Rating</dt>
                  <dd>{restaurant.rating} / 5</dd>
                </div>
              ) : null}
              {priceLevel ? (
                <div>
                  <dt>Price</dt>
                  <dd>{priceLevel}</dd>
                </div>
              ) : null}
              <div>
                <dt>Neighborhood</dt>
                <dd>{restaurant.locationLabel || restaurant.city}</dd>
              </div>
            </dl>
          </div>
        </section>

        <section className="italian-intro" id="story">
          <div className="italian-section-copy">
            <p className="italian-kicker">Italian hospitality</p>
            <h2>A dining room built around craft, warmth, and the table.</h2>
            <p>
              {restaurant.tagline || `${restaurant.name} brings regional flavor, polished service, and a relaxed Italian pace to ${restaurant.city}.`}
            </p>
          </div>
          <div className="italian-intro-grid">
            <article>
              <span>01</span>
              <h3>Seasonal Rhythm</h3>
              <p>{restaurant.storySteps[0]?.description || "Fresh pasta, bright sauces, and produce-led plates guide the experience from first bite to final espresso."}</p>
            </article>
            <article>
              <span>02</span>
              <h3>Service With Pace</h3>
              <p>{restaurant.storySteps[1]?.description || "A polished host flow makes reservations, arrivals, and special requests feel clear before guests walk in."}</p>
            </article>
            <article>
              <span>03</span>
              <h3>Memorable Evenings</h3>
              <p>{restaurant.storySteps[2]?.description || "From weekday pasta nights to celebratory tables, the site keeps the path to a reservation immediate."}</p>
            </article>
          </div>
          {storyImage ? <ItalianImage media={storyImage} className="italian-story-image" sizes="(min-width: 900px) 42vw, 100vw" /> : null}
        </section>

        {featuredDishes.length > 0 ? (
          <section className="italian-signature" id="signature">
            <div className="italian-section-copy">
              <p className="italian-kicker">Signature plates</p>
              <h2>A first taste of the dishes guests remember.</h2>
            </div>
            <div className="italian-dish-grid">
              {featuredDishes.map((dish, index) => {
                const media = menuImageFor(dish, foodImages[index % Math.max(1, foodImages.length)]);
                return (
                  <article className="italian-dish" key={`${dish.name}-${index}`}>
                    {media ? <ItalianImage media={media} className="italian-dish-image" sizes="(min-width: 900px) 25vw, 100vw" /> : null}
                    <div>
                      <p>{dish.category || "Chef selection"}</p>
                      <h3>{dish.name}</h3>
                      <span>{dish.price || "Seasonal"}</span>
                    </div>
                  </article>
                );
              })}
            </div>
          </section>
        ) : null}

        {restaurant.menuCategories.length > 0 ? (
          <section className="italian-menu-section" id="menu">
            <div className="italian-section-copy">
              <p className="italian-kicker">Menu chapters</p>
              <h2>A focused look at what is coming from the kitchen.</h2>
              <p>
                Start with antipasti, pasta, mains, or sweets, then reserve when the evening is ready.
              </p>
            </div>
            <div className="italian-menu-grid">
              {restaurant.menuCategories.slice(0, 6).map((category) => (
                <section className="italian-menu-card" key={category.name}>
                  <h3>{category.name}</h3>
                  <div>
                    {category.items.slice(0, 5).map((item, index) => (
                      <article className="italian-menu-item" key={`${category.name}-${item.name}-${index}`}>
                        <div>
                          <h4>{item.name}</h4>
                          {item.description ? <p>{item.description}</p> : null}
                        </div>
                        <span>{item.price || ""}</span>
                      </article>
                    ))}
                  </div>
                </section>
              ))}
            </div>
          </section>
        ) : null}

        {gallery.length > 0 ? (
          <section className="italian-gallery" id="gallery">
            <div className="italian-section-copy">
              <p className="italian-kicker">Atmosphere</p>
              <h2>A room, a table, and the details around the meal.</h2>
            </div>
            <div className="italian-gallery-grid">
              {gallery.map((image, index) => (
                <ItalianImage
                  key={`${image.url}-${index}`}
                  media={image}
                  className={index === 0 ? "italian-gallery-image italian-gallery-image-large" : "italian-gallery-image"}
                  sizes={index === 0 ? "(min-width: 900px) 50vw, 100vw" : "(min-width: 900px) 25vw, 50vw"}
                />
              ))}
            </div>
          </section>
        ) : null}

        {restaurant.reviews.length > 0 ? (
          <section className="italian-reviews">
            <div className="italian-section-copy">
              <p className="italian-kicker">Guest proof</p>
              <h2>What guests are saying after the last pour.</h2>
            </div>
            <div className="italian-review-grid">
              {restaurant.reviews.slice(0, 3).map((review, index) => (
                <blockquote key={`${review.reviewer}-${index}`}>
                  <p>{review.review}</p>
                  <footer>
                    <strong>{review.reviewer}</strong>
                    <span>{review.stars ? `${review.stars} star review` : "Guest review"}</span>
                  </footer>
                </blockquote>
              ))}
            </div>
          </section>
        ) : null}

        <section className="italian-reserve-band" id="reserve">
          <div className="italian-reserve-copy">
            <p className="italian-kicker">Reservations</p>
            <h2>Save a seat for the next service.</h2>
            <p>
              Request a table online or use the floating AI assistant to ask about menu, hours, and availability.
            </p>
            <ul>
              {restaurant.phone ? <li>Same-day phone requests: {restaurant.phone}</li> : null}
              <li>Reservation requests remain pending until confirmed.</li>
              {restaurant.email ? <li>Private dining inquiries: {restaurant.email}</li> : null}
            </ul>
          </div>
          <ItalianReservationForm restaurant={restaurant} />
        </section>

        <section className="italian-visit" id="visit">
          <div className="italian-section-copy">
            <p className="italian-kicker">Visit</p>
            <h2>{restaurant.locationLabel || restaurant.city}</h2>
            <p>{restaurant.address}</p>
            <a className="italian-button italian-button-secondary" href={mapsUrl} target="_blank" rel="noreferrer">
              Open in Maps
            </a>
          </div>
          <div className="italian-hours-panel">
            <h3>Hours</h3>
            {openHours.length > 0 ? (
              <dl>
                {openHours.map(([day, hours]) => (
                  <div key={day}>
                    <dt>{day}</dt>
                    <dd>{hours}</dd>
                  </div>
                ))}
              </dl>
            ) : (
              <p>Hours are available by phone.</p>
            )}
          </div>
        </section>
      </main>

      <footer className="italian-footer">
        <div>
          <strong>{restaurant.name}</strong>
          <span>{cuisine}</span>
        </div>
        <a href="#reserve">Reserve a Table</a>
      </footer>

      <div className="italian-mobile-bar">
        <a href="#reserve">Reserve</a>
        {restaurant.phone ? <a href={`tel:${restaurant.phone}`}>Call</a> : <a href="#visit">Visit</a>}
      </div>
    </div>
  );
}
