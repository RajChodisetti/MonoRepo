"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import BrandLogo from "@/components/layout/BrandLogo";
import Button from "@/components/ui/Button";

function ChevronDownIcon({ open }: { open: boolean }) {
  return (
    <svg
      className={`h-4 w-4 transition-transform duration-200 ${open ? "rotate-180" : ""}`}
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

export default function Nav() {
  const [scrolled, setScrolled] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [servicesOpen, setServicesOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24);
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
    <header className="pointer-events-none fixed inset-x-0 top-0 z-50 px-3 pt-3 md:px-5">
      <nav
        aria-label="Primary navigation"
        className={`nav-glass pointer-events-auto mx-auto flex h-[72px] max-w-6xl items-center justify-between gap-4 rounded-2xl px-3.5 transition duration-300 md:px-5 ${
          scrolled ? "nav-glass-scrolled" : ""
        }`}
      >
        <BrandLogo priority />

        <ul className="hidden min-w-0 items-center gap-1 xl:flex">
          <li className="group relative shrink-0">
            <button
              type="button"
              className="inline-flex cursor-pointer items-center gap-1.5 whitespace-nowrap rounded-full px-3.5 py-2 text-sm font-semibold text-muted transition-colors duration-200 hover:bg-surface hover:text-ink"
              aria-haspopup="true"
              aria-expanded={servicesOpen}
              aria-controls="desktop-services-menu"
              onClick={() => setServicesOpen((value) => !value)}
            >
              {siteContent.servicesNav.label}
              <ChevronDownIcon open={servicesOpen} />
            </button>
            <div
              id="desktop-services-menu"
              className={`absolute left-1/2 top-full min-w-80 -translate-x-1/2 pt-4 transition duration-200 ${
                servicesOpen
                  ? "visible opacity-100"
                  : "invisible opacity-0 group-focus-within:visible group-focus-within:opacity-100 group-hover:visible group-hover:opacity-100"
              }`}
            >
              <div className="rounded-2xl border border-border bg-bg-elevated p-2 shadow-[0_24px_60px_rgba(15,39,31,0.14)]">
                {siteContent.servicesNav.items.map((item) => (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={closeMenus}
                    className="block rounded-xl px-4 py-3 transition-colors duration-200 hover:bg-surface"
                  >
                    <span className="block font-display text-base font-semibold text-ink">
                      {item.label}
                    </span>
                    <span className="mt-1 block text-xs leading-5 text-muted">
                      {item.description}
                    </span>
                  </Link>
                ))}
              </div>
            </div>
          </li>
          {siteContent.nav.map((link) => (
            <li key={link.href} className="shrink-0">
              <Link
                href={link.href}
                className="whitespace-nowrap rounded-full px-3.5 py-2 text-sm font-semibold text-muted transition-colors duration-200 hover:bg-surface hover:text-ink"
              >
                {link.label}
              </Link>
            </li>
          ))}
          <li className="ml-3 shrink-0">
            <Button href={getBookCallUrl()} className="!px-5 !py-2.5 !text-sm">
              Book a call
            </Button>
          </li>
        </ul>

        <div className="flex shrink-0 items-center gap-2 xl:hidden">
          <Button href={getBookCallUrl()} className="!px-4 !py-2.5 text-xs sm:text-sm">
            Book a call
          </Button>
          <button
            type="button"
            className="flex h-11 w-11 cursor-pointer flex-col items-center justify-center gap-1.5 rounded-full border border-border bg-bg-elevated text-ink transition-colors hover:bg-surface"
            aria-label={mobileOpen ? "Close menu" : "Open menu"}
            aria-expanded={mobileOpen}
            aria-controls="mobile-navigation"
            onClick={() => {
              setMobileOpen((value) => !value);
              setServicesOpen(false);
            }}
          >
            <span className={`block h-0.5 w-5 bg-current transition ${mobileOpen ? "translate-y-1 rotate-45" : ""}`} />
            <span className={`block h-0.5 w-5 bg-current transition ${mobileOpen ? "-translate-y-1 -rotate-45" : ""}`} />
          </button>
        </div>
      </nav>

      {mobileOpen ? (
        <div
          id="mobile-navigation"
          className="nav-glass pointer-events-auto mx-auto mt-2 max-w-6xl rounded-2xl p-3 shadow-[0_24px_60px_rgba(15,39,31,0.14)] xl:hidden"
        >
          <button
            type="button"
            className="flex w-full cursor-pointer items-center justify-between rounded-xl px-3 py-3 text-left text-sm font-semibold text-ink transition-colors hover:bg-surface"
            aria-expanded={servicesOpen}
            aria-controls="mobile-services-menu"
            onClick={() => setServicesOpen((value) => !value)}
          >
            {siteContent.servicesNav.label}
            <ChevronDownIcon open={servicesOpen} />
          </button>
          {servicesOpen ? (
            <div id="mobile-services-menu" className="mb-2 grid gap-1">
              {siteContent.servicesNav.items.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={closeMenus}
                  className="rounded-xl bg-surface px-3 py-3 text-sm text-ink transition-colors hover:bg-sage/60"
                >
                  <span className="block font-semibold">{item.label}</span>
                  <span className="mt-1 block text-xs leading-5 text-muted">{item.description}</span>
                </Link>
              ))}
            </div>
          ) : null}
          {siteContent.nav.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              onClick={closeMenus}
              className="block rounded-xl px-3 py-3 text-sm font-semibold text-muted transition-colors hover:bg-surface hover:text-ink"
            >
              {link.label}
            </Link>
          ))}
        </div>
      ) : null}
    </header>
  );
}
