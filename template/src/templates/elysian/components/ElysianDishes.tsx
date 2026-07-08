"use client";

import { useState } from "react";
import type { MenuItem } from "@/data/types/menu";
import ElysianImage from "./ElysianImage";

function DishCard({ dish }: { dish: MenuItem }) {
  const [fav, setFav] = useState(false);

  return (
    <article
      className="dish-card reveal fade-up"
      data-tilt
      onMouseMove={(e) => {
        const card = e.currentTarget;
        const rect = card.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        const rotateX = (y / rect.height - 0.5) * -8;
        const rotateY = (x / rect.width - 0.5) * 8;
        card.style.transform = `perspective(900px) rotateX(${rotateX}deg) rotateY(${rotateY}deg) translateY(-10px)`;
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.transform = "";
      }}
    >
      <div className="dish-media">
        {dish.image ? (
          <ElysianImage
            src={dish.image}
            alt={dish.name}
            fill
            sizes="(max-width: 768px) 100vw, 33vw"
          />
        ) : null}
        <button
          type="button"
          className={`fav-btn${fav ? " active" : ""}`}
          aria-label="Add to favorites"
          onClick={() => setFav((v) => !v)}
        >
          <svg viewBox="0 0 24 24">
            <path
              d="M12 21s-7.2-4.6-9.8-9.1C.5 8.6 2 5 5.6 5c2 0 3.4 1 4.9 2.8C12 6 13.4 5 15.4 5 19 5 20.5 8.6 18.8 11.9 19.2 4.6 12 21 12 21Z"
              stroke="currentColor"
              strokeWidth="1.6"
            />
          </svg>
        </button>
        {dish.isChefSpecial ? <span className="dish-tag">Chef&apos;s Signature</span> : null}
      </div>
      <div className="dish-body">
        <div className="dish-top">
          <h3>{dish.name}</h3>
          {dish.price ? <span className="dish-price">{dish.price}</span> : null}
        </div>
        {dish.description ? <p>{dish.description}</p> : null}
      </div>
    </article>
  );
}

export default function ElysianDishes({ dishes }: { dishes: MenuItem[] }) {
  if (!dishes.length) return null;

  return (
    <section className="dishes section" id="dishes">
      <div className="container">
        <div className="section-head reveal fade-up">
          <p className="eyebrow">Signature Dishes</p>
          <h2 className="section-title">
            Crafted for the <span className="gold-text">Unforgettable</span>
          </h2>
          <p className="section-sub">
            A curated selection from our menu, each plate born of season and story.
          </p>
        </div>
        <div className="dish-grid">
          {dishes.map((dish) => (
            <DishCard key={dish.name} dish={dish} />
          ))}
        </div>
      </div>
    </section>
  );
}
