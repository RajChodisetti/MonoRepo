/* eslint-disable @next/next/no-img-element */
import type { FoodieContent, FoodieDishCard } from "../lib/foodieContent";

function Stars({ rating }: { rating: number }) {
  return (
    <div className="foodie-stars" aria-label={`${rating} out of 5`}>
      {Array.from({ length: 5 }).map((_, i) => (
        <svg key={i} viewBox="0 0 24 24" className={i < rating ? "on" : ""} aria-hidden="true">
          <path d="M12 3.5l2.6 5.3 5.9.9-4.3 4.1 1 5.8L12 17l-5.2 2.6 1-5.8L3.5 9.7l5.9-.9L12 3.5z" />
        </svg>
      ))}
    </div>
  );
}

export default function FoodieHero({
  hero,
  dish,
}: {
  hero: FoodieContent["hero"];
  dish: FoodieDishCard;
}) {
  return (
    <section className="foodie-hero">
      <svg className="foodie-deco foodie-deco-pizza foodie-anim foodie-anim-deco" viewBox="0 0 64 64" fill="none" aria-hidden="true">
        <path d="M32 6 8 20l24 38 24-38L32 6Z" stroke="currentColor" strokeWidth="2" strokeLinejoin="round" />
        <path d="M14 23h36" stroke="currentColor" strokeWidth="2" />
        <circle cx="26" cy="30" r="2.4" fill="currentColor" />
        <circle cx="36" cy="34" r="2.4" fill="currentColor" />
        <circle cx="31" cy="43" r="2.4" fill="currentColor" />
      </svg>
      <svg className="foodie-deco foodie-deco-spoon foodie-anim foodie-anim-deco" viewBox="0 0 64 64" fill="none" aria-hidden="true" style={{ animationDelay: "0.35s" }}>
        <path d="M40 8c-6 0-10 5-10 12s4 10 10 10 10-3 10-10S46 8 40 8Z" stroke="currentColor" strokeWidth="2" />
        <path d="M40 42v14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
        <path d="M18 8c-3 6-3 12 0 16s3 8 0 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
        <path d="M18 8v48" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      </svg>

      <div className="foodie-container foodie-hero-grid">
        <div className="foodie-hero-copy">
          <p className="foodie-eyebrow foodie-anim foodie-anim-up" style={{ animationDelay: "0.15s" }}>
            {hero.eyebrow}
          </p>
          <h1 className="foodie-title foodie-anim foodie-anim-up" style={{ animationDelay: "0.28s" }}>
            {hero.titleLead}{" "}
            <span className="foodie-title-accent">
              {hero.titleAccent}
              <svg className="foodie-underline" viewBox="0 0 280 90" fill="none" aria-hidden="true" preserveAspectRatio="none">
                <path
                  className="foodie-underline-path"
                  d="M28 48c8-22 60-34 118-36 62-2 108 10 120 28 10 16-6 32-42 38-48 8-118 10-162 2-28-6-42-18-34-32Z"
                  stroke="currentColor"
                  strokeWidth="5.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  fill="none"
                />
              </svg>
            </span>
          </h1>
          <p className="foodie-desc foodie-anim foodie-anim-up" style={{ animationDelay: "0.42s" }}>
            {hero.description}
          </p>

          <div className="foodie-cta-row foodie-anim foodie-anim-up" style={{ animationDelay: "0.55s" }}>
            <a href="#reservation" className="foodie-btn foodie-btn-primary">
              {hero.primaryCta}
            </a>
            <a href="#order" className="foodie-btn foodie-btn-outline">
              {hero.secondaryCta}
            </a>
          </div>

          <p className="foodie-hours foodie-anim foodie-anim-up" style={{ animationDelay: "0.68s" }}>
            <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="1.7" />
              <path d="M12 7v5l3.2 2" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
            {hero.hours}
          </p>
        </div>

        <div className="foodie-hero-visual foodie-anim foodie-anim-visual">
          <div className="foodie-ring" aria-hidden="true" />
          <div className="foodie-plate-wrap">
            <img className="foodie-plate" src={hero.plate} alt="Signature fresh salad plate" />
          </div>

          <img className="foodie-garnish foodie-garnish-tomato" src={hero.garnish.tomato} alt="" aria-hidden="true" />
          <img className="foodie-garnish foodie-garnish-basil" src={hero.garnish.basil} alt="" aria-hidden="true" />
          <img className="foodie-garnish foodie-garnish-onion" src={hero.garnish.onion} alt="" aria-hidden="true" />

          <div className="foodie-badge">
            <span>{hero.badge}</span>
            <span className="foodie-badge-emoji" aria-hidden="true">🍔</span>
          </div>

          <div className="foodie-dish-card">
            <img className="foodie-dish-thumb" src={dish.image} alt={dish.name} />
            <div className="foodie-dish-body">
              <h3 className="foodie-dish-name">{dish.name}</h3>
              <Stars rating={dish.rating} />
              <p className="foodie-dish-desc">{dish.description}</p>
              <span className="foodie-dish-price">{dish.price}</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
