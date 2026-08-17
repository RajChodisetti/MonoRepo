"use client";

import { useEffect, useRef, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import type { RestaurantContent } from "@/data/types/restaurant";
import TemplateSwitchButton from "@/components/TemplateSwitchButton";

const LINKS = [
  { href: "#menu", label: "Menu" },
  { href: "#story", label: "Story" },
  { href: "#gallery", label: "Gallery" },
  { href: "#reviews", label: "Reviews" },
  { href: "#contact", label: "Contact" },
];

export default function Navigation({ restaurant }: { restaurant: RestaurantContent }) {
  const [scrolled, setScrolled] = useState(false);
  const [open, setOpen] = useState(false);
  const menuButtonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 40);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  useEffect(() => {
    if (!open) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      setOpen(false);
      menuButtonRef.current?.focus();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [open]);

  return (
    <header
      className={`fixed inset-x-0 top-0 z-50 border-b transition-all duration-500 ${
        scrolled
          ? "border-[#f7f0e6]/12 bg-[#10100e]/98 shadow-[0_4px_24px_rgba(0,0,0,0.45)] backdrop-blur-lg"
          : "border-[#f7f0e6]/8 bg-[#10100e]/88 backdrop-blur-md"
      }`}
    >
      <nav className="mx-auto flex h-[72px] max-w-6xl items-center justify-between px-6">
        <a
          href="#hero"
          className="font-display min-w-0 max-w-[min(48vw,18rem)] truncate text-xl tracking-wide text-[#f7f0e6] drop-shadow-sm"
          title={restaurant.name}
        >
          {restaurant.name}
        </a>

        <ul className="hidden items-center gap-8 min-[1100px]:flex">
          {LINKS.map((link) => (
            <li key={link.href}>
              <a
                href={link.href}
                className="text-[11px] font-medium uppercase tracking-[0.14em] text-[#f7f0e6]/85 transition hover:text-[#b88a44]"
              >
                {link.label}
              </a>
            </li>
          ))}
          <li>
            <TemplateSwitchButton variant="cinematic" />
          </li>
          <li>
            <a
              href={restaurant.primaryCTA.href}
              className="rounded border border-[#b88a44] bg-[#b88a44]/10 px-4 py-2 text-[11px] font-semibold uppercase tracking-[0.12em] text-[#b88a44] transition hover:bg-[#b88a44]/20"
            >
              {restaurant.primaryCTA.label}
            </a>
          </li>
        </ul>

        <button
          ref={menuButtonRef}
          type="button"
          className="flex h-11 w-11 flex-col items-center justify-center gap-1.5 min-[1100px]:hidden"
          aria-label={open ? "Close navigation menu" : "Open navigation menu"}
          aria-controls="cinematic-mobile-menu"
          aria-expanded={open}
          onClick={() => setOpen((v) => !v)}
        >
          <span className="block h-0.5 w-6 bg-[#f7f0e6]" />
          <span className="block h-0.5 w-6 bg-[#f7f0e6]" />
          <span className="block h-0.5 w-6 bg-[#f7f0e6]" />
        </button>
      </nav>

      <AnimatePresence>
        {open && (
          <motion.div
            id="cinematic-mobile-menu"
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            className="border-t border-[#f7f0e6]/10 bg-[#10100e]/98 px-6 py-4 min-[1100px]:hidden"
          >
            {LINKS.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={() => setOpen(false)}
                className="block border-b border-[#f7f0e6]/10 py-3 text-sm uppercase tracking-widest text-[#f7f0e6]/70"
              >
                {link.label}
              </a>
            ))}
            <div className="mt-4">
              <TemplateSwitchButton variant="cinematic" mode="mobile" />
            </div>
            <a
              href={restaurant.primaryCTA.href}
              className="mt-4 block rounded bg-[#b88a44] py-3 text-center text-sm font-semibold uppercase tracking-widest text-[#10100e]"
            >
              {restaurant.primaryCTA.label}
            </a>
          </motion.div>
        )}
      </AnimatePresence>
    </header>
  );
}
