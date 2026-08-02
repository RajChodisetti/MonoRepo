"use client";

import { useEffect, useId, useRef, useState } from "react";
import { createPortal } from "react-dom";
import Image from "next/image";
import Link from "next/link";
import { ChevronDown } from "@/components/icons/NavIcons";
import {
  companyFeatured,
  companyLinks,
} from "@/components/layout/companyMegaMenu.config";

export default function CompanyMegaMenu() {
  const [open, setOpen] = useState(false);
  const [mounted, setMounted] = useState(false);
  const menuId = useId();
  const closeTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    setMounted(true);
  }, []);

  const clearClose = () => {
    if (closeTimer.current) {
      clearTimeout(closeTimer.current);
      closeTimer.current = null;
    }
  };

  const openMenu = () => {
    clearClose();
    setOpen(true);
  };

  const scheduleClose = () => {
    clearClose();
    closeTimer.current = setTimeout(() => setOpen(false), 140);
  };

  useEffect(() => {
    if (!open) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [open]);

  useEffect(() => () => clearClose(), []);

  const panel =
    open && mounted
      ? createPortal(
          <>
            <div
              className="fixed inset-x-0 top-[4.35rem] bottom-0 z-40 bg-white/55 backdrop-blur-[14px]"
              aria-hidden="true"
              onClick={() => setOpen(false)}
            />

            <div
              id={menuId}
              role="menu"
              className="fixed inset-x-0 top-[4.35rem] z-50 flex justify-center px-6 pt-1 md:px-10 lg:px-12"
              onMouseEnter={openMenu}
              onMouseLeave={scheduleClose}
            >
              <div
                className="flex overflow-hidden rounded-2xl border border-black/10 bg-white shadow-[0_24px_80px_rgba(0,0,0,0.14)]"
                style={{ width: "min(980px, calc(100vw - 3rem))" }}
              >
                {/* Links */}
                <div className="flex w-[240px] shrink-0 flex-col justify-center border-r border-black/8 px-8 py-8">
                  <ul className="flex flex-col gap-1">
                    {companyLinks.map((item) => (
                      <li key={item.href}>
                        <Link
                          href={item.href}
                          role="menuitem"
                          className="block rounded-lg px-1 py-2.5 text-[16px] font-semibold tracking-[-0.01em] text-[#171717] transition-colors hover:bg-black/[0.04]"
                          onClick={() => setOpen(false)}
                        >
                          {item.label}
                        </Link>
                      </li>
                    ))}
                  </ul>
                </div>

                {/* Featured cards */}
                <div className="grid min-w-0 flex-1 grid-cols-2 gap-4 p-5">
                  {companyFeatured.map((card) => (
                    <Link
                      key={card.href + card.title}
                      href={card.href}
                      className="group relative min-h-[280px] overflow-hidden rounded-2xl"
                      onClick={() => setOpen(false)}
                    >
                      {card.variant === "hiring" ? (
                        <div
                          className="absolute inset-0"
                          style={{
                            background:
                              "linear-gradient(180deg, #5b8def 0%, #2f5fd0 48%, #1e3f9a 100%)",
                          }}
                        >
                          <svg
                            className="absolute inset-0 h-full w-full opacity-35"
                            viewBox="0 0 400 500"
                            preserveAspectRatio="none"
                            aria-hidden="true"
                          >
                            <path
                              d="M0 80 C80 60 120 110 200 90 C280 70 320 120 400 100"
                              fill="none"
                              stroke="white"
                              strokeWidth="1.2"
                            />
                            <path
                              d="M0 160 C90 140 140 190 220 170 C300 150 340 200 400 180"
                              fill="none"
                              stroke="white"
                              strokeWidth="1.2"
                            />
                            <path
                              d="M0 250 C70 230 130 280 210 255 C290 230 340 290 400 265"
                              fill="none"
                              stroke="white"
                              strokeWidth="1.2"
                            />
                            <path
                              d="M0 340 C85 315 145 370 230 345 C310 320 350 380 400 355"
                              fill="none"
                              stroke="white"
                              strokeWidth="1.2"
                            />
                            <path
                              d="M0 420 C95 400 150 450 240 430 C320 410 360 460 400 440"
                              fill="none"
                              stroke="white"
                              strokeWidth="1.2"
                            />
                          </svg>
                        </div>
                      ) : (
                        <>
                          <Image
                            src={card.imageSrc!}
                            alt={card.imageAlt ?? ""}
                            fill
                            sizes="360px"
                            className="object-cover transition-transform duration-500 group-hover:scale-[1.03]"
                          />
                          <div className="absolute inset-0 bg-gradient-to-t from-black/75 via-black/25 to-transparent" />
                        </>
                      )}
                      <p className="absolute inset-x-0 bottom-0 p-5 text-[20px] font-semibold leading-snug tracking-[-0.02em] text-white">
                        {card.title}
                      </p>
                    </Link>
                  ))}
                </div>
              </div>
            </div>
          </>,
          document.body,
        )
      : null;

  return (
    <div className="relative" onMouseEnter={openMenu} onMouseLeave={scheduleClose}>
      <button
        type="button"
        className={`inline-flex cursor-pointer items-center gap-1 rounded-lg px-2.5 py-1.5 text-[15px] font-semibold text-[#1a1a1a] transition-colors hover:bg-black/[0.04] ${
          open ? "bg-black/[0.04]" : ""
        }`}
        aria-expanded={open}
        aria-haspopup="true"
        aria-controls={menuId}
        onClick={() => setOpen((value) => !value)}
      >
        Company
        <ChevronDown className={`h-3.5 w-3.5 transition-transform duration-200 ${open ? "rotate-180" : ""}`} />
      </button>
      {panel}
    </div>
  );
}
