"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import type { FoodieContent, FoodieMenuItem } from "../lib/foodieContent";

const PAGE_SIZE = 4;

function StarIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 3.5l2.6 5.3 5.9.9-4.3 4.1 1 5.8L12 17l-5.2 2.6 1-5.8L3.5 9.7l5.9-.9L12 3.5z" />
    </svg>
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

function DishCard({
  item,
  activeId,
  onHover,
  onImageBroken,
}: {
  item: FoodieMenuItem;
  activeId: string | null;
  onHover: (id: string | null) => void;
  onImageBroken: (id: string) => void;
}) {
  const filled = activeId === item.id || (!activeId && item.featured);

  return (
    <article
      className="foodie-menu-card"
      onMouseEnter={() => onHover(item.id)}
      onMouseLeave={() => onHover(null)}
    >
      <div className="foodie-menu-card-media">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={item.image}
          alt={item.name}
          onError={() => onImageBroken(item.id)}
        />
      </div>

      <div className="foodie-menu-card-body">
        <div className="foodie-menu-card-rating">
          <span>{item.ratingLabel}</span>
          <StarIcon />
        </div>
        <h3 className="foodie-menu-card-title">{item.name}</h3>
        {item.price ? <p className="foodie-menu-card-price">{item.price}</p> : null}
        {item.description ? <p className="foodie-menu-card-desc">{item.description}</p> : null}
        <button type="button" className={`foodie-menu-cart${filled ? " filled" : ""}`}>
          Add to Cart
        </button>
      </div>
    </article>
  );
}

export default function FoodieMenu({ menu }: { menu: FoodieContent["menu"] }) {
  const sectionRef = useRef<HTMLElement>(null);
  const [visible, setVisible] = useState(false);
  const [hoverId, setHoverId] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const [brokenIds, setBrokenIds] = useState<Set<string>>(() => new Set());

  const usableItems = useMemo(
    () => menu.items.filter((item) => item.image && !brokenIds.has(item.id)),
    [menu.items, brokenIds],
  );

  const pageCount = Math.max(1, Math.ceil(usableItems.length / PAGE_SIZE));
  const safePage = Math.min(page, pageCount - 1);

  const visibleItems = useMemo(() => {
    const start = safePage * PAGE_SIZE;
    return usableItems.slice(start, start + PAGE_SIZE);
  }, [usableItems, safePage]);

  useEffect(() => {
    setPage(0);
    setHoverId(null);
    setBrokenIds(new Set());
  }, [menu.items]);

  useEffect(() => {
    if (page > pageCount - 1) setPage(Math.max(0, pageCount - 1));
  }, [page, pageCount]);

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

  const markBroken = (id: string) => {
    setBrokenIds((prev) => {
      if (prev.has(id)) return prev;
      const next = new Set(prev);
      next.add(id);
      return next;
    });
  };

  const goPrev = () => setPage((p) => Math.max(0, p - 1));
  const goNext = () => setPage((p) => Math.min(pageCount - 1, p + 1));

  return (
    <section
      id="menu"
      ref={sectionRef}
      className={`foodie-menu${visible ? " in-view" : ""}`}
      aria-labelledby="foodie-menu-heading"
    >
      <div className="foodie-container">
        <div className="foodie-menu-header">
          <div className="foodie-menu-title-wrap">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="foodie-menu-leaf foodie-menu-leaf-a" src="/foodie/menu-basil.png" alt="" aria-hidden="true" />
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="foodie-menu-leaf foodie-menu-leaf-b" src="/foodie/menu-basil.png" alt="" aria-hidden="true" />
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="foodie-menu-leaf foodie-menu-leaf-c" src="/foodie/menu-basil.png" alt="" aria-hidden="true" />
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="foodie-menu-leaf foodie-menu-leaf-d" src="/foodie/menu-basil.png" alt="" aria-hidden="true" />
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="foodie-menu-leaf foodie-menu-leaf-e" src="/foodie/menu-basil.png" alt="" aria-hidden="true" />
            <h2 id="foodie-menu-heading" className="foodie-menu-title">
              {menu.titleLead}{" "}
              <span className="foodie-menu-title-accent">
                {menu.titleAccent}
                <svg className="foodie-menu-title-ring" viewBox="0 0 200 70" fill="none" aria-hidden="true" preserveAspectRatio="none">
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
          </div>

          <div className="foodie-menu-arrows">
            <button
              type="button"
              className="foodie-menu-arrow"
              aria-label="Previous dishes"
              onClick={goPrev}
              disabled={safePage <= 0}
            >
              <ArrowLeft />
            </button>
            <button
              type="button"
              className="foodie-menu-arrow active"
              aria-label="Next dishes"
              onClick={goNext}
              disabled={safePage >= pageCount - 1}
            >
              <ArrowRight />
            </button>
          </div>
        </div>

        <div className="foodie-menu-track" key={safePage}>
          {visibleItems.map((item) => (
            <DishCard
              key={item.id}
              item={item}
              activeId={hoverId}
              onHover={setHoverId}
              onImageBroken={markBroken}
            />
          ))}
        </div>
      </div>

      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img className="foodie-menu-chilies" src="/foodie/chilies.png" alt="" aria-hidden="true" />
    </section>
  );
}
