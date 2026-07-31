"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import BrandLogo from "@/components/layout/BrandLogo";

function ChevronDownIcon({ open }: { open: boolean }) {
  return (
    <svg
      className={`h-3.5 w-3.5 transition-transform duration-200 ${open ? "rotate-180" : ""}`}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="m6 9 6 6 6-6" />
    </svg>
  );
}

const linkClass =
  "whitespace-nowrap rounded-full px-3.5 py-2 text-[13px] font-medium tracking-[-0.01em] text-white/55 transition-colors duration-200 hover:bg-white/[0.06] hover:text-white";

export default function Nav() {
  const [scrolled, setScrolled] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [servicesOpen, setServicesOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 16);
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setMobileOpen(false);
        setServicesOpen(false);
      }
    };

    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("scroll", onScroll);
      window.removeEventListener("keydown", onKey);
    };
  }, []);

  const closeMenus = () => {
    setMobileOpen(false);
    setServicesOpen(false);
  };

  return (
    <header className="pointer-events-none fixed inset-x-0 top-0 z-50 px-3 pt-4 md:px-6">
      <nav
        aria-label="Primary navigation"
        className={`nav-glass pointer-events-auto relative mx-auto flex h-16 max-w-6xl items-center justify-between gap-3 rounded-full px-3 transition duration-300 md:h-[68px] md:px-5 ${
          scrolled ? "nav-glass-scrolled" : ""
        }`}
      >
        {/* Zone 1 — logo */}
        <div className="relative z-10 shrink-0">
          <BrandLogo priority />
        </div>

        {/* Zone 2 — centered links (desktop) */}
        <ul className="absolute left-1/2 top-1/2 hidden -translate-x-1/2 -translate-y-1/2 items-center gap-0.5 xl:flex">
          <li className="group relative">
            <button
              type="button"
              className={`${linkClass} inline-flex cursor-pointer items-center gap-1.5`}
              aria-haspopup="true"
              aria-expanded={servicesOpen}
              aria-controls="desktop-services-menu"
              onClick={() => setServicesOpen((value) => !value)}
            >
              Services
              <ChevronDownIcon open={servicesOpen} />
            </button>
            <div
              id="desktop-services-menu"
              className={`absolute left-1/2 top-full min-w-[20rem] -translate-x-1/2 pt-4 transition duration-200 ${
                servicesOpen
                  ? "visible opacity-100"
                  : "invisible opacity-0 group-focus-within:visible group-focus-within:opacity-100 group-hover:visible group-hover:opacity-100"
              }`}
            >
              <div className="overflow-hidden rounded-2xl border border-white/10 bg-[#0c0c0c]/95 p-1.5 shadow-[0_28px_70px_rgba(0,0,0,0.55)] backdrop-blur-xl">
                {siteContent.servicesNav.items.map((item) => (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={closeMenus}
                    className="block rounded-xl px-4 py-3 transition-colors duration-200 hover:bg-white/[0.06]"
                  >
                    <span className="block text-sm font-semibold text-white">{item.label}</span>
                    <span className="mt-1 block text-xs leading-5 text-white/45">{item.description}</span>
                  </Link>
                ))}
              </div>
            </div>
          </li>
          {siteContent.nav.map((link) => (
            <li key={link.href}>
              <Link href={link.href} className={linkClass}>
                {link.label}
              </Link>
            </li>
          ))}
        </ul>

        {/* Zone 3 — CTA / mobile */}
        <div className="relative z-10 flex shrink-0 items-center gap-2">
          <a
            href={getBookCallUrl()}
            className="hidden rounded-full bg-white px-5 py-2.5 text-[13px] font-semibold text-black shadow-[0_8px_24px_rgba(255,255,255,0.12)] transition hover:bg-white/90 sm:inline-flex"
          >
            Book a call
          </a>
          <button
            type="button"
            className="flex h-10 w-10 cursor-pointer flex-col items-center justify-center gap-1.5 rounded-full border border-white/12 bg-white/[0.04] text-white transition-colors hover:bg-white/[0.08] xl:hidden"
            aria-label={mobileOpen ? "Close menu" : "Open menu"}
            aria-expanded={mobileOpen}
            aria-controls="mobile-navigation"
            onClick={() => {
              setMobileOpen((value) => !value);
              setServicesOpen(false);
            }}
          >
            <span className={`block h-0.5 w-4 bg-current transition ${mobileOpen ? "translate-y-[3px] rotate-45" : ""}`} />
            <span className={`block h-0.5 w-4 bg-current transition ${mobileOpen ? "opacity-0" : ""}`} />
            <span className={`block h-0.5 w-4 bg-current transition ${mobileOpen ? "-translate-y-[3px] -rotate-45" : ""}`} />
          </button>
        </div>
      </nav>

      {mobileOpen ? (
        <div
          id="mobile-navigation"
          className="nav-glass pointer-events-auto mx-auto mt-2 max-w-6xl rounded-3xl p-2 shadow-[0_24px_60px_rgba(0,0,0,0.5)] xl:hidden"
        >
          <button
            type="button"
            className="flex w-full cursor-pointer items-center justify-between rounded-2xl px-4 py-3.5 text-left text-sm font-semibold text-white transition-colors hover:bg-white/[0.05]"
            aria-expanded={servicesOpen}
            aria-controls="mobile-services-menu"
            onClick={() => setServicesOpen((value) => !value)}
          >
            Services
            <ChevronDownIcon open={servicesOpen} />
          </button>
          {servicesOpen ? (
            <div id="mobile-services-menu" className="mb-1 grid gap-1 px-1">
              {siteContent.servicesNav.items.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={closeMenus}
                  className="rounded-2xl bg-white/[0.04] px-4 py-3 text-sm text-white transition-colors hover:bg-white/[0.08]"
                >
                  <span className="block font-semibold">{item.label}</span>
                  <span className="mt-1 block text-xs leading-5 text-white/45">{item.description}</span>
                </Link>
              ))}
            </div>
          ) : null}
          {siteContent.nav.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              onClick={closeMenus}
              className="block rounded-2xl px-4 py-3.5 text-sm font-semibold text-white/65 transition-colors hover:bg-white/[0.05] hover:text-white"
            >
              {link.label}
            </Link>
          ))}
          <a
            href={getBookCallUrl()}
            onClick={closeMenus}
            className="mt-1 flex items-center justify-center rounded-full bg-white px-5 py-3 text-sm font-semibold text-black"
          >
            Book a call
          </a>
        </div>
      ) : null}
    </header>
  );
}
