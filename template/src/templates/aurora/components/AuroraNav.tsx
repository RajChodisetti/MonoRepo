"use client";

import { useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import type { RestaurantContent } from "@/data/types/restaurant";

const LINKS = [
  { href: "#features", label: "Features" },
  { href: "#showcase", label: "Showcase" },
  { href: "#menu", label: "Menu" },
  { href: "#reviews", label: "Reviews" },
  { href: "#contact", label: "Contact" },
];

export default function AuroraNav({ restaurant }: { restaurant: RestaurantContent }) {
  const [scrolled, setScrolled] = useState(false);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <header
      className={`fixed inset-x-0 top-0 z-50 transition-all duration-500 ${
        scrolled
          ? "border-b border-white/10 bg-[#09090B]/80 backdrop-blur-xl"
          : "bg-transparent"
      }`}
    >
      <nav className="aurora-container flex h-16 items-center justify-between">
        <a href="#hero" className="aurora-heading text-lg font-semibold text-white">
          {restaurant.name}
        </a>

        <ul className="hidden items-center gap-8 md:flex">
          {LINKS.map((link) => (
            <li key={link.href}>
              <a
                href={link.href}
                className="text-xs font-medium uppercase tracking-wider text-white/50 transition hover:text-cyan-400"
              >
                {link.label}
              </a>
            </li>
          ))}
          <li>
            <a href={restaurant.primaryCTA.href} className="aurora-btn-primary !px-5 !py-2 !text-xs">
              {restaurant.primaryCTA.label}
            </a>
          </li>
        </ul>

        <button
          type="button"
          className="md:hidden"
          aria-label="Menu"
          onClick={() => setOpen((v) => !v)}
        >
          <span className="block h-0.5 w-6 bg-white" />
          <span className="mt-1.5 block h-0.5 w-6 bg-white" />
        </button>
      </nav>

      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            className="border-t border-white/10 bg-[#09090B]/95 px-6 py-4 md:hidden"
          >
            {LINKS.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={() => setOpen(false)}
                className="block py-2 text-sm text-white/70"
              >
                {link.label}
              </a>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </header>
  );
}
