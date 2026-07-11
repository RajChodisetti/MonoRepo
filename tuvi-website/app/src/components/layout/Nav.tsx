"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import Button from "@/components/ui/Button";

function ChevronDownIcon({ open }: { open: boolean }) {
  return (
    <svg
      className={`h-4 w-4 transition-transform ${open ? "rotate-180" : ""}`}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="m6 9 6 6 6-6" />
    </svg>
  );
}

export default function Nav() {
  const [scrolled, setScrolled] = useState(false);
  const [open, setOpen] = useState(false);
  const [servicesOpen, setServicesOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  const toggleMenu = () => {
    setOpen((current) => {
      const next = !current;
      if (!next) setServicesOpen(false);
      return next;
    });
  };

  return (
    <header
      className={`nav-glass fixed inset-x-0 top-0 z-50 transition-[background,box-shadow,backdrop-filter] duration-500 ${
        scrolled ? "nav-glass-scrolled" : ""
      }`}
    >
      <nav className="mx-auto flex h-16 max-w-6xl items-center justify-between gap-4 px-5 md:h-[72px] md:px-8">
        <Link
          href="/"
          className="shrink-0 font-display text-xl font-bold tracking-tight text-ink"
        >
          Tuvi<span className="text-primary">.</span>
        </Link>

        <ul className="hidden min-w-0 items-center gap-1 xl:flex">
          <li className="group relative shrink-0">
            <button
              type="button"
              className="whitespace-nowrap rounded-full px-3.5 py-2 text-sm font-medium text-muted transition hover:bg-zinc-100 hover:text-ink focus:text-ink focus:outline-none"
              aria-haspopup="true"
            >
              {siteContent.servicesNav.label}
            </button>
            <div className="invisible absolute left-1/2 top-full min-w-72 -translate-x-1/2 pt-4 opacity-0 transition duration-200 group-focus-within:visible group-focus-within:opacity-100 group-hover:visible group-hover:opacity-100">
              <div className="rounded-2xl border border-border bg-white p-2 shadow-xl shadow-zinc-900/10">
                {siteContent.servicesNav.items.map((item) => (
                  <Link
                    key={item.href}
                    href={item.href}
                    className="block rounded-xl px-4 py-3 transition hover:bg-zinc-50"
                  >
                    <span className="block text-sm font-semibold text-ink">{item.label}</span>
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
                className="whitespace-nowrap rounded-full px-3.5 py-2 text-sm font-medium text-muted transition hover:bg-zinc-100 hover:text-ink"
              >
                {link.label}
              </Link>
            </li>
          ))}
          <li className="ml-3 shrink-0">
            <Button href={getBookCallUrl()} variant="primary" className="!px-5 !py-2.5 !text-sm">
              Book a call
            </Button>
          </li>
        </ul>

        <div className="flex shrink-0 items-center gap-3 xl:hidden">
          <Button href={getBookCallUrl()} variant="primary" className="!px-4 !py-2 text-xs">
            Book a call
          </Button>
          <button
            type="button"
            className="flex flex-col gap-1.5 p-1"
            aria-label="Open menu"
            onClick={toggleMenu}
          >
            <span className="block h-0.5 w-5 bg-ink" />
            <span className="block h-0.5 w-5 bg-ink" />
          </button>
        </div>
      </nav>

      {open && (
        <div className="border-t border-border bg-white/95 backdrop-blur-xl xl:hidden">
          <div className="flex flex-col gap-1 px-5 py-4">
            <div className="py-2">
              <button
                type="button"
                className="flex w-full items-center justify-between rounded-xl px-1 py-2 text-left text-sm font-semibold text-ink transition hover:text-primary"
                aria-expanded={servicesOpen}
                aria-controls="mobile-services-menu"
                onClick={() => setServicesOpen((value) => !value)}
              >
                {siteContent.servicesNav.label}
                <ChevronDownIcon open={servicesOpen} />
              </button>
              {servicesOpen && (
                <div id="mobile-services-menu" className="mt-2 flex flex-col gap-1">
                  {siteContent.servicesNav.items.map((item) => (
                    <Link
                      key={item.href}
                      href={item.href}
                      onClick={() => {
                        setOpen(false);
                        setServicesOpen(false);
                      }}
                      className="rounded-xl border border-border bg-zinc-50 px-3 py-3 text-sm text-ink transition hover:border-primary/40"
                    >
                      <span className="block font-semibold">{item.label}</span>
                      <span className="mt-1 block text-xs leading-5 text-muted">
                        {item.description}
                      </span>
                    </Link>
                  ))}
                </div>
              )}
            </div>
            {siteContent.nav.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                onClick={() => {
                  setOpen(false);
                  setServicesOpen(false);
                }}
                className="py-2.5 text-sm font-medium text-muted transition hover:text-ink"
              >
                {link.label}
              </Link>
            ))}
            <Button href={getBookCallUrl()} variant="primary" className="mt-3 w-full">
              Book a call
            </Button>
          </div>
        </div>
      )}
    </header>
  );
}
