"use client";

import { useEffect, useId, useRef, useState } from "react";
import { createPortal } from "react-dom";
import Image from "next/image";
import Link from "next/link";
import { ChevronDown } from "@/components/icons/NavIcons";
import MegaMenuIcon from "@/components/icons/MegaMenuIcons";
import {
  productMegaFeatured,
  productMegaSections,
} from "@/components/layout/productMegaMenu.config";

function MegaSection({
  section,
}: {
  section: (typeof productMegaSections)[number];
}) {
  return (
    <div>
      <p className="mb-4 text-[13px] font-medium text-[#9a9a9a]">{section.title}</p>
      <ul className="flex flex-col gap-0.5">
        {section.items.map((item) => (
          <li key={item.href}>
            <Link
              href={item.href}
              role="menuitem"
              className="flex items-center gap-3 rounded-lg px-1 py-2 text-[15px] font-semibold text-[#171717] transition-colors hover:bg-black/[0.04]"
            >
              <MegaMenuIcon name={item.icon} className="h-[18px] w-[18px] shrink-0 text-[#555]" />
              <span className="inline-flex items-center gap-2">
                {item.label}
                {item.badge ? (
                  <span className="rounded-full bg-[#ebebeb] px-2 py-0.5 text-[11px] font-medium text-[#6f6f6f]">
                    {item.badge}
                  </span>
                ) : null}
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default function ProductMegaMenu() {
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

  const leftCol = [productMegaSections[0], productMegaSections[1]];
  const midCol = [productMegaSections[2], productMegaSections[3]];

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
              className="fixed inset-x-0 top-[4.35rem] z-50 px-6 pt-1 md:px-10 lg:px-12"
              onMouseEnter={openMenu}
              onMouseLeave={scheduleClose}
            >
              {/* Fixed width box, left-aligned — same size whether left or right */}
              <div
                className="flex overflow-hidden rounded-2xl border border-black/10 bg-white shadow-[0_24px_80px_rgba(0,0,0,0.14)]"
                style={{ width: "min(1180px, calc(100vw - 3rem))" }}
              >
                <div
                  className="grid flex-1 grid-cols-2 gap-x-16 px-10 py-10"
                  style={{ minWidth: 0 }}
                >
                  <div className="flex flex-col gap-10">
                    {leftCol.map((section) => (
                      <MegaSection key={section.title} section={section} />
                    ))}
                  </div>
                  <div className="flex flex-col gap-10">
                    {midCol.map((section) => (
                      <MegaSection key={section.title} section={section} />
                    ))}
                  </div>
                </div>

                <div className="flex w-[220px] shrink-0 flex-col gap-3 border-l border-black/8 p-4">
                  {productMegaFeatured.map((story) => (
                    <Link
                      key={story.href}
                      href={story.href}
                      className="group relative h-[120px] w-full shrink-0 overflow-hidden rounded-xl"
                      onClick={() => setOpen(false)}
                    >
                      <Image
                        src={story.imageSrc}
                        alt={story.imageAlt}
                        fill
                        sizes="188px"
                        className="object-cover transition-transform duration-500 group-hover:scale-[1.03]"
                      />
                      <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/25 to-transparent" />
                      <p className="absolute inset-x-0 bottom-0 p-3 text-[12px] font-semibold leading-snug text-white">
                        {story.title}
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
        Product
        <ChevronDown className={`h-3.5 w-3.5 transition-transform duration-200 ${open ? "rotate-180" : ""}`} />
      </button>
      {panel}
    </div>
  );
}
