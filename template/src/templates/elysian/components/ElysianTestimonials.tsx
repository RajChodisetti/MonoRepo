"use client";

import { useTestimonialCarousel } from "../hooks/useTestimonialCarousel";
import type { RestaurantContent } from "@/data/types/restaurant";
import { excerptReview, reviewsForCarousel } from "../lib/reviewExcerpt";

export default function ElysianTestimonials({
  reviews,
}: {
  reviews: RestaurantContent["reviews"];
}) {
  const slides = reviewsForCarousel(reviews);
  const { index, goTo, setPaused } = useTestimonialCarousel(slides.length);

  if (!slides.length) return null;

  return (
    <section className="testimonials section" id="testimonials">
      <div className="container">
        <div className="section-head reveal fade-up">
          <p className="eyebrow">Guest Reflections</p>
          <h2 className="section-title">
            Moments Our Guests <span className="gold-text">Remember</span>
          </h2>
        </div>

        <div
          className="testi-carousel reveal fade-up"
          onMouseEnter={() => setPaused(true)}
          onMouseLeave={() => setPaused(false)}
        >
          <div className="testi-card">
            <p className="testi-quote">&ldquo;{excerptReview(slides[index].review)}&rdquo;</p>
            <div className="testi-person">
              <strong>{slides[index].reviewer}</strong>
              <span className="testi-rating">
                {"★".repeat(Math.round(slides[index].stars || 5))}
              </span>
            </div>
          </div>
          {slides.length > 1 ? (
            <div className="testi-dots" id="testiDots">
              {slides.map((_, i) => (
                <button
                  key={i}
                  type="button"
                  className={i === index ? "active" : ""}
                  aria-label={`Go to review ${i + 1}`}
                  onClick={() => goTo(i)}
                />
              ))}
            </div>
          ) : null}
        </div>
      </div>
    </section>
  );
}
