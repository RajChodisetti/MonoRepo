"use client";

import { useEffect, useState } from "react";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import Button from "@/components/ui/Button";

export default function Nav() {
  const [scrolled, setScrolled] = useState(false);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <header
      className={`nav-glass fixed inset-x-0 top-0 z-50 transition-[background,box-shadow,backdrop-filter] duration-500 ${
        scrolled ? "nav-glass-scrolled" : ""
      }`}
    >
      <nav className="mx-auto flex h-16 max-w-6xl items-center justify-between gap-4 px-5 md:h-[72px] md:px-8">
        <a href="#" className="shrink-0 font-display text-lg font-bold tracking-tight text-text md:text-xl">
          Tuvi<span className="text-gold">.</span>
        </a>

        <ul className="hidden min-w-0 items-center gap-4 xl:flex xl:gap-7">
          {siteContent.nav.map((link) => (
            <li key={link.href} className="shrink-0">
              <a
                href={link.href}
                className="whitespace-nowrap text-[10px] font-medium uppercase tracking-[0.12em] text-muted transition hover:text-cyan xl:text-xs xl:tracking-[0.14em]"
              >
                {link.label}
              </a>
            </li>
          ))}
          <li className="shrink-0">
            <Button href={getBookCallUrl()} variant="primary" className="!px-4 !py-2.5 !text-xs xl:!px-5">
              Book a Call
            </Button>
          </li>
        </ul>

        <div className="flex shrink-0 items-center gap-3 xl:hidden">
          <Button href={getBookCallUrl()} variant="primary" className="!px-4 !py-2 text-xs">
            Book a Call
          </Button>
          <button
            type="button"
            className="flex flex-col gap-1.5 p-1"
            aria-label="Open menu"
            onClick={() => setOpen((v) => !v)}
          >
            <span className="block h-0.5 w-5 bg-text" />
            <span className="block h-0.5 w-5 bg-text" />
          </button>
        </div>
      </nav>

      {open && (
        <div className="border-t border-white/5 bg-bg/80 backdrop-blur-xl xl:hidden">
          <div className="flex flex-col gap-1 px-5 py-4">
            {siteContent.nav.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={() => setOpen(false)}
                className="py-2.5 text-sm text-muted transition hover:text-text"
              >
                {link.label}
              </a>
            ))}
            <Button
              href={getBookCallUrl()}
              variant="primary"
              className="mt-3 w-full"
            >
              Book a Call
            </Button>
          </div>
        </div>
      )}
    </header>
  );
}
