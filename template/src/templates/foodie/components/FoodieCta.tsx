"use client";

import { useEffect, useRef, useState } from "react";
import type { FoodieContent } from "../lib/foodieContent";

export default function FoodieCta({ cta }: { cta: FoodieContent["cta"] }) {
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
      { threshold: 0.18, rootMargin: "0px 0px -8% 0px" },
    );

    io.observe(el);
    return () => io.disconnect();
  }, []);

  return (
    <section
      id="reservation"
      ref={sectionRef}
      className={`foodie-cta${visible ? " in-view" : ""}`}
      aria-labelledby="foodie-cta-heading"
    >
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img className="foodie-cta-side foodie-cta-wrap" src={cta.wrapImage} alt="" aria-hidden="true" />
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img className="foodie-cta-side foodie-cta-fries" src={cta.friesImage} alt="" aria-hidden="true" />

      <div className="foodie-container foodie-cta-inner">
        <h2 id="foodie-cta-heading" className="foodie-cta-title">
          {cta.titleLead}
          <br />
          <span className="foodie-cta-title-accent">
            {cta.titleAccent}
            <svg
              className="foodie-cta-title-ring"
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
          </span>
        </h2>

        <p className="foodie-cta-desc">{cta.description}</p>

        <div className="foodie-cta-actions">
          <a href="#reservation" className="foodie-btn foodie-btn-primary">
            {cta.primaryCta}
          </a>
          <a href="#menu" className="foodie-btn foodie-btn-outline">
            {cta.secondaryCta}
          </a>
        </div>
      </div>
    </section>
  );
}
