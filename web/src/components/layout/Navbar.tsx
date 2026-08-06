"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import BrandLogo from "@/components/layout/BrandLogo";
import NavLinks from "@/components/layout/NavLinks";
import NavActions from "@/components/layout/NavActions";
import { primaryNavLinks } from "@/components/layout/nav.config";
import { productMegaSections } from "@/components/layout/productMegaMenu.config";
import { CloseIcon, MenuIcon } from "@/components/icons/NavIcons";

export default function Navbar() {
  const [mobileOpen, setMobileOpen] = useState(false);

  useEffect(() => {
    if (!mobileOpen) return;

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setMobileOpen(false);
    };

    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.body.style.overflow = "";
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [mobileOpen]);

  const closeMobile = () => setMobileOpen(false);

  return (
    <header className="nav-glass sticky top-0 z-50 w-full">
      <nav
        aria-label="Primary"
        className="mx-auto grid h-[4.5rem] w-full max-w-[1200px] grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-4 px-6 md:px-10 lg:px-12"
      >
        <div className="justify-self-start">
          <BrandLogo onNavigate={closeMobile} />
        </div>

        <NavLinks links={primaryNavLinks} className="hidden justify-center lg:flex" />

        <div className="hidden shrink-0 justify-self-end lg:block">
          <NavActions className="flex" />
        </div>

        <button
          type="button"
          className="inline-flex h-10 w-10 cursor-pointer items-center justify-center justify-self-end rounded-full text-ink transition-colors hover:bg-ink/[0.05] lg:hidden"
          aria-label={mobileOpen ? "Close menu" : "Open menu"}
          aria-expanded={mobileOpen}
          onClick={() => setMobileOpen((value) => !value)}
        >
          {mobileOpen ? <CloseIcon className="h-5 w-5" /> : <MenuIcon className="h-5 w-5" />}
        </button>
      </nav>

      {mobileOpen ? (
        <div className="max-h-[calc(100svh-4.5rem)] overflow-y-auto border-t border-border bg-bg lg:hidden">
          <div className="mx-auto flex max-w-[1200px] flex-col gap-6 px-6 py-6">
            <ul className="flex flex-col gap-1">
              {primaryNavLinks.map((item) => {
                if (item.type === "link") {
                  return (
                    <li key={item.label}>
                      <Link
                        href={item.href}
                        onClick={closeMobile}
                        className="block rounded-xl px-3 py-3 text-base font-medium text-[#1a1a1a] hover:bg-black/[0.04]"
                      >
                        {item.label}
                      </Link>
                    </li>
                  );
                }

                if (item.type === "mega") {
                  return (
                    <li key={item.label} className="px-3 py-2">
                      <p className="text-xs font-semibold uppercase tracking-[0.14em] text-black/40">
                        {item.label}
                      </p>
                      <div className="mt-2 space-y-4">
                        {productMegaSections.map((section) => (
                          <div key={section.title}>
                            <p className="px-1 text-[12px] font-medium text-[#8a8a8a]">{section.title}</p>
                            <ul className="mt-1">
                              {section.items.map((sub) => (
                                <li key={sub.href}>
                                  <Link
                                    href={sub.href}
                                    onClick={closeMobile}
                                    className="block rounded-xl px-1 py-2.5 text-base font-medium text-[#1a1a1a] hover:bg-black/[0.04]"
                                  >
                                    {sub.label}
                                  </Link>
                                </li>
                              ))}
                            </ul>
                          </div>
                        ))}
                      </div>
                    </li>
                  );
                }

                return (
                  <li key={item.label} className="px-3 py-2">
                    <p className="text-xs font-semibold uppercase tracking-[0.14em] text-black/40">
                      {item.label}
                    </p>
                    <ul className="mt-1">
                      {item.items.map((sub) => (
                        <li key={sub.href}>
                          <Link
                            href={sub.href}
                            onClick={closeMobile}
                            className="block rounded-xl px-1 py-2.5 text-base font-medium text-[#1a1a1a] hover:bg-black/[0.04]"
                          >
                            {sub.label}
                          </Link>
                        </li>
                      ))}
                    </ul>
                  </li>
                );
              })}
            </ul>
            <NavActions
              onNavigate={closeMobile}
              className="flex flex-col items-stretch gap-3 sm:flex-row sm:items-center"
            />
          </div>
        </div>
      ) : null}
    </header>
  );
}
