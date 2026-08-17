"use client";

import { useEffect, useRef, useState } from "react";
import TemplateSwitchButton from "@/components/TemplateSwitchButton";
import { useMediaQuery } from "@/hooks/useMediaQuery";

const LINKS = [
  { href: "#about", label: "About" },
  { href: "#dishes", label: "Dishes" },
  { href: "#menu", label: "Menu" },
  { href: "#gallery", label: "Gallery" },
  { href: "#reservation", label: "Reserve" },
  { href: "#contact", label: "Contact" },
];

export default function ElysianNav({
  name,
  nameAccent,
  scrolled,
  theme,
  onToggleTheme,
  showDishes,
}: {
  name: string;
  nameAccent: string;
  scrolled: boolean;
  theme: "dark" | "light";
  onToggleTheme: () => void;
  showDishes: boolean;
}) {
  const [open, setOpen] = useState(false);
  const menuButtonRef = useRef<HTMLButtonElement>(null);
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  const isDrawer = useMediaQuery("(max-width: 980px)");
  const links = showDishes ? LINKS : LINKS.filter((l) => l.href !== "#dishes");

  useEffect(() => {
    if (!open || !isDrawer) return;
    closeButtonRef.current?.focus();
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      setOpen(false);
      menuButtonRef.current?.focus();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [isDrawer, open]);

  return (
    <header className={`navbar${scrolled ? " scrolled" : ""}`} id="navbar">
      <div className="nav-inner">
        <a href="#home" className="logo" title={`${name}${nameAccent ? ` ${nameAccent}` : ""}`}>
          {name}
          {nameAccent ? <span>{nameAccent}</span> : null}
        </a>
        <nav
          className={`nav-links${open ? " open" : ""}`}
          id="navLinks"
          aria-label="Primary"
          aria-hidden={isDrawer && !open ? true : undefined}
          inert={isDrawer && !open ? true : undefined}
        >
          <button
            ref={closeButtonRef}
            type="button"
            className="menu-close"
            aria-label="Close navigation menu"
            onClick={() => {
              setOpen(false);
              menuButtonRef.current?.focus();
            }}
          >
            <span aria-hidden="true">×</span>
          </button>
          {links.map((l) => (
            <a
              key={l.href}
              href={l.href}
              className="nav-link"
              onClick={() => setOpen(false)}
            >
              {l.label}
            </a>
          ))}
          <div className="mt-4 min-[901px]:hidden">
            <TemplateSwitchButton variant="elysian" mode="mobile" />
          </div>
        </nav>
        <div className="nav-actions">
          <button
            type="button"
            className="theme-toggle"
            id="themeToggle"
            aria-label="Toggle light and dark mode"
            onClick={onToggleTheme}
            data-theme={theme}
          >
            <svg viewBox="0 0 24 24" className="icon-sun">
              <path
                d="M12 4V2M12 22v-2M4 12H2M22 12h-2M5 5 3.5 3.5M20.5 20.5 19 19M5 19l-1.5 1.5M20.5 3.5 19 5"
                stroke="currentColor"
                strokeWidth="1.6"
                strokeLinecap="round"
              />
              <circle cx="12" cy="12" r="4.5" stroke="currentColor" strokeWidth="1.6" />
            </svg>
            <svg viewBox="0 0 24 24" className="icon-moon">
              <path
                d="M20 14.5A8.5 8.5 0 1 1 9.5 4a7 7 0 0 0 10.5 10.5Z"
                stroke="currentColor"
                strokeWidth="1.6"
                strokeLinejoin="round"
              />
            </svg>
          </button>
          <a href="#reservation" className="btn btn-gold nav-cta">
            Reserve a Table
          </a>
          <TemplateSwitchButton variant="elysian" />
          <button
            ref={menuButtonRef}
            type="button"
            className={`menu-toggle${open ? " active" : ""}`}
            id="menuToggle"
            aria-label={open ? "Close navigation menu" : "Open navigation menu"}
            aria-controls="navLinks"
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
