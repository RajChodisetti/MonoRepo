"use client";

import { useEffect, useRef, useState } from "react";
import type { FoodieContent } from "../lib/foodieContent";

function Stars({ rating }: { rating: number }) {
  return (
    <div className="foodie-reviews-stars" aria-label={`${rating} out of 5`}>
      {Array.from({ length: 5 }).map((_, i) => (
        <svg key={i} viewBox="0 0 24 24" className={i < rating ? "on" : ""} aria-hidden="true">
          <path d="M12 3.5l2.6 5.3 5.9.9-4.3 4.1 1 5.8L12 17l-5.2 2.6 1-5.8L3.5 9.7l5.9-.9L12 3.5z" />
        </svg>
      ))}
    </div>
  );
}

function ArrowLeft() {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path d="M15 6 9 12l6 6" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function ArrowRight() {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path d="m9 6 6 6-6 6" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export default function FoodieReviews({ reviews }: { reviews: FoodieContent["reviews"] }) {
  const sectionRef = useRef<HTMLElement>(null);
  const [visible, setVisible] = useState(false);
  const [index, setIndex] = useState(0);
  const total = reviews.items.length;
  const active = reviews.items[index] ?? reviews.items[0];

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
      { threshold: 0.18, rootMargin: "0px 0px -8% 0px" },
    );

    io.observe(el);
    return () => io.disconnect();
  }, []);

  const prev = () => setIndex((i) => (i - 1 + total) % total);
  const next = () => setIndex((i) => (i + 1) % total);

  return (
    <section
      id="reviews"
      ref={sectionRef}
      className={`foodie-reviews${visible ? " in-view" : ""}`}
      aria-labelledby="foodie-reviews-heading"
    >
      {/* floating garnishes */}
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img className="foodie-reviews-garnish foodie-reviews-garnish-basil" src="/foodie/basil.png" alt="" aria-hidden="true" />
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img className="foodie-reviews-garnish foodie-reviews-garnish-tomato" src="/foodie/tomato.png" alt="" aria-hidden="true" />
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img className="foodie-reviews-garnish foodie-reviews-garnish-salad" src="/foodie/salad-bits.png" alt="" aria-hidden="true" />

      <div className="foodie-container foodie-reviews-grid">
        <div className="foodie-reviews-visual">
          <div className="foodie-reviews-ring" aria-hidden="true" />
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img className="foodie-reviews-chef" src={reviews.chefImage} alt="Head chef presenting a fresh salad" />

          <div className="foodie-reviews-badge">
            <p className="foodie-reviews-badge-label">{reviews.badgeLabel}</p>
            <div className="foodie-reviews-badge-row">
              <div className="foodie-reviews-avatars">
                {reviews.avatars.map((src) => (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img key={src} src={src} alt="" />
                ))}
              </div>
              <span className="foodie-reviews-count">{reviews.reviewCount}</span>
            </div>
          </div>
        </div>

        <div className="foodie-reviews-copy">
          <h2 id="foodie-reviews-heading" className="foodie-reviews-title">
            {reviews.titleBefore}{" "}
            <span className="foodie-reviews-title-accent">
              {reviews.titleAccent}
              <svg
                className="foodie-reviews-title-ring"
                viewBox="0 0 220 70"
                fill="none"
                aria-hidden="true"
                preserveAspectRatio="none"
              >
                <path
                  d="M16 36c7-16 48-26 96-28 52-2 90 8 98 22 7 12-8 24-38 28-44 6-100 8-132 2-22-4-32-12-24-24Z"
                  stroke="currentColor"
                  strokeWidth="4"
                  strokeLinecap="round"
                  fill="none"
                />
              </svg>
            </span>{" "}
            {reviews.titleAfter}
          </h2>

          <p className="foodie-reviews-desc">{reviews.description}</p>

          <div className="foodie-reviews-card" key={active.id}>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="foodie-reviews-card-avatar" src={active.avatar} alt="" />
            <div className="foodie-reviews-card-body">
              <div className="foodie-reviews-card-meta">
                <strong>{active.name}</strong>
                <span>{active.location}</span>
              </div>
              <Stars rating={active.rating} />
              {active.quote ? <p className="foodie-reviews-card-quote">{active.quote}</p> : null}
            </div>
          </div>

          <div className="foodie-reviews-arrows">
            <button type="button" className="foodie-reviews-arrow" aria-label="Previous review" onClick={prev}>
              <ArrowLeft />
            </button>
            <button type="button" className="foodie-reviews-arrow active" aria-label="Next review" onClick={next}>
              <ArrowRight />
            </button>
          </div>
        </div>
      </div>
    </section>
  );
}
