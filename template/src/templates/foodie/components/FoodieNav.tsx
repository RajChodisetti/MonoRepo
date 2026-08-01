"use client";

import { useEffect, useState } from "react";
import type { FoodieContent, FoodieNavLink } from "../lib/foodieContent";

function LogoMark() {
  return (
    <span className="foodie-logo-mark" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="1.6" />
        <path
          d="M8 12c0-2.2 1.8-4 4-4s4 1.8 4 4"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
        />
        <path d="M12 5.5v3M9.5 6.2l.9 2.6M14.5 6.2l-.9 2.6" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
      </svg>
    </span>
  );
}

export default function FoodieNav({
  brand,
  links,
}: {
  brand: FoodieContent["brand"];
  links: FoodieNavLink[];
}) {
  const [open, setOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 12);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <header className={`foodie-nav${scrolled ? " scrolled" : ""}`}>
      <div className="foodie-nav-inner">
        <a
          href="#home"
          className="foodie-logo foodie-anim foodie-anim-logo"
          aria-label={`${brand.name} home`}
        >
          <LogoMark />
          <span className="foodie-logo-text">{brand.name}</span>
        </a>

        <nav className={`foodie-links${open ? " open" : ""}`} aria-label="Primary">
          {links.map((link, i) => (
            <a
              key={link.href}
              href={link.href}
              className={`foodie-link foodie-anim foodie-anim-nav${link.active ? " active" : ""}`}
              style={{ animationDelay: `${120 + i * 55}ms` }}
              onClick={() => setOpen(false)}
            >
              {link.label}
            </a>
          ))}
        </nav>

        <div className="foodie-nav-actions foodie-anim foodie-anim-nav-actions">
          <button type="button" className="foodie-icon-btn" aria-label="Search">
            <svg viewBox="0 0 24 24" fill="none">
              <circle cx="11" cy="11" r="7" stroke="currentColor" strokeWidth="1.7" />
              <path d="m20 20-3.2-3.2" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" />
            </svg>
          </button>
          <button type="button" className="foodie-icon-btn" aria-label="Cart">
            <svg viewBox="0 0 24 24" fill="none">
              <path
                d="M4 5h2l1.6 10.2a1.5 1.5 0 0 0 1.5 1.3h7.2a1.5 1.5 0 0 0 1.48-1.24L20 8H7"
                stroke="currentColor"
                strokeWidth="1.6"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
              <circle cx="10" cy="20" r="1.3" fill="currentColor" />
              <circle cx="17" cy="20" r="1.3" fill="currentColor" />
            </svg>
          </button>
          <button type="button" className="foodie-icon-btn" aria-label="Wishlist">
            <svg viewBox="0 0 24 24" fill="none">
              <path
                d="M12 20s-7-4.4-7-9.3A3.7 3.7 0 0 1 12 8a3.7 3.7 0 0 1 7 2.7C19 15.6 12 20 12 20Z"
                stroke="currentColor"
                strokeWidth="1.6"
                strokeLinejoin="round"
              />
            </svg>
          </button>

          <a href="#login" className="foodie-login">
            Log in
          </a>

          <button
            type="button"
            className={`foodie-burger${open ? " active" : ""}`}
            aria-label="Toggle navigation menu"
            aria-expanded={open}
            onClick={() => setOpen((v) => !v)}
          >
            <span />
            <span />
            <span />
          </button>
        </div>
      </div>
    </header>
  );
}
