"use client";

import { useEffect, useRef, useState } from "react";
import type { FoodieContent } from "../lib/foodieContent";
import { mapsEmbedSrc, mapsHref, telHref } from "@/lib/reservation";

function PinIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path
        d="M12 21s7-5.4 7-11a7 7 0 1 0-14 0c0 5.6 7 11 7 11Z"
        stroke="currentColor"
        strokeWidth="1.8"
      />
      <circle cx="12" cy="10" r="2.4" stroke="currentColor" strokeWidth="1.8" />
    </svg>
  );
}

function PhoneIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path
        d="M7.2 3.8c.6-.6 1.6-.5 2 .3l1.4 2.6c.3.6.2 1.3-.3 1.8l-1 1a12.5 12.5 0 0 0 5.2 5.2l1-1c.5-.5 1.2-.6 1.8-.3l2.6 1.4c.8.4.9 1.4.3 2l-1.3 1.3c-.6.6-1.5.9-2.4.7A16.5 16.5 0 0 1 5.2 7.5c-.2-.9.1-1.8.7-2.4l1.3-1.3Z"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function MailIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <rect x="3.5" y="5.5" width="17" height="13" rx="2.2" stroke="currentColor" strokeWidth="1.8" />
      <path d="m4.5 7.5 7.5 6 7.5-6" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  );
}

function ClockIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <circle cx="12" cy="12" r="8.5" stroke="currentColor" strokeWidth="1.8" />
      <path d="M12 7.5V12l3.2 2" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  );
}

export default function FoodieContact({
  contact,
  brandName,
}: {
  contact: FoodieContent["contact"];
  brandName: string;
}) {
  const sectionRef = useRef<HTMLElement>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const el = sectionRef.current;
    if (!el) return;

    const io = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setVisible(true);
          io.disconnect();
        }
      },
      { threshold: 0.16, rootMargin: "0px 0px -8% 0px" },
    );

    io.observe(el);
    return () => io.disconnect();
  }, []);

  const embedSrc = contact.coordinates
    ? mapsEmbedSrc(contact.coordinates)
    : `https://maps.google.com/maps?q=${encodeURIComponent(contact.address)}&z=15&output=embed`;
  const directionsUrl = mapsHref(contact.address, contact.coordinates);
  const phoneUrl = telHref(contact.phone);

  return (
    <section
      id="contact"
      ref={sectionRef}
      className={`foodie-contact${visible ? " in-view" : ""}`}
      aria-labelledby="foodie-contact-heading"
    >
      <div className="foodie-container foodie-contact-grid">
        <div className="foodie-contact-info">
          <p className="foodie-contact-eyebrow">{contact.eyebrow}</p>
          <h2 id="foodie-contact-heading" className="foodie-contact-title">
            {contact.titleLead}{" "}
            <span className="foodie-contact-title-accent">
              {contact.titleAccent}
              <svg
                className="foodie-contact-title-ring"
                viewBox="0 0 200 70"
                fill="none"
                aria-hidden="true"
                preserveAspectRatio="none"
              >
                <path
                  d="M18 36c6-16 42-26 86-28 48-2 82 8 90 22 7 12-8 24-36 28-40 6-92 8-122 2-20-4-30-12-24-24Z"
                  stroke="currentColor"
                  strokeWidth="4"
                  strokeLinecap="round"
                  fill="none"
                />
              </svg>
            </span>
          </h2>

          <ul className="foodie-contact-list">
            <li>
              <span className="foodie-contact-icon">
                <PinIcon />
              </span>
              <span>{contact.address}</span>
            </li>
            <li>
              <span className="foodie-contact-icon">
                <PhoneIcon />
              </span>
              <a href={phoneUrl}>{contact.phone}</a>
            </li>
            <li>
              <span className="foodie-contact-icon">
                <MailIcon />
              </span>
              <a href={`mailto:${contact.email}`}>{contact.email}</a>
            </li>
            <li>
              <span className="foodie-contact-icon">
                <ClockIcon />
              </span>
              <span>{contact.hoursLine}</span>
            </li>
          </ul>

          <div className="foodie-contact-actions">
            <a
              href={directionsUrl}
              className="foodie-btn foodie-btn-primary"
              target="_blank"
              rel="noopener noreferrer"
            >
              {contact.directionsLabel}
            </a>
            <a href={phoneUrl} className="foodie-btn foodie-btn-outline">
              {contact.callLabel}
            </a>
          </div>
        </div>

        <div className="foodie-contact-map">
          <iframe
            title={`Map of ${brandName}`}
            src={embedSrc}
            loading="lazy"
            referrerPolicy="no-referrer-when-downgrade"
            allowFullScreen
          />
        </div>
      </div>
    </section>
  );
}
