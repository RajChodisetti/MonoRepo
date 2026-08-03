"use client";

import Image from "next/image";
import { useCallback, useRef } from "react";
import { trustedOwners, type TrustedOwner } from "@/components/sections/trustedOwners.config";

function ChevronLeft() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" aria-hidden="true">
      <path d="M14.5 6 9 12l5.5 6" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function ChevronRight() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" aria-hidden="true">
      <path d="M9.5 6 15 12l-5.5 6" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function PlayIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" aria-hidden="true">
      <path d="M9 7.5v9l8-4.5-8-4.5Z" fill="currentColor" />
    </svg>
  );
}

function OwnerCard({ owner }: { owner: TrustedOwner }) {
  return (
    <article className="grid h-[360px] w-[min(100%,1040px)] shrink-0 snap-start grid-cols-2 overflow-hidden rounded-[24px] bg-parchment sm:h-[390px] sm:rounded-[28px]">
      {/* Quote side */}
      <div className="relative flex flex-col px-5 py-5 sm:px-8 sm:py-7 md:px-9 md:py-8">
        {/* Subtle arcs */}
        <svg
          className="pointer-events-none absolute inset-0 h-full w-full opacity-[0.35]"
          viewBox="0 0 480 400"
          preserveAspectRatio="none"
          aria-hidden="true"
        >
          {[110, 180, 250, 330].map((r, i) => (
            <circle
              key={r}
              cx="40"
              cy="400"
              r={r}
              fill="none"
              stroke="rgba(0,0,0,0.08)"
              strokeWidth={1.1 - i * 0.05}
            />
          ))}
        </svg>

        <div className="relative z-10 flex flex-1 flex-col">
          <blockquote className="text-[clamp(1.05rem,1.8vw,1.35rem)] font-bold leading-[1.3] tracking-[-0.025em] text-[#111111]">
            &ldquo;{owner.quote}&rdquo;
          </blockquote>

          <p className="mt-3.5 text-[12px] font-medium text-[#8a8580] sm:text-[13px]">
            {owner.name} — {owner.business}
          </p>

          <button
            type="button"
            className="mt-4 inline-flex w-fit items-center gap-1 rounded-lg bg-primary px-3.5 py-2 text-[13px] font-semibold text-bg transition-colors hover:bg-primary-dim"
          >
            Learn more
            <span aria-hidden="true">›</span>
          </button>

          <div className="mt-auto grid grid-cols-2 gap-5 pt-6">
            {owner.metrics.map((metric) => (
              <div key={metric.label}>
                <p className="text-[clamp(1.2rem,2vw,1.55rem)] font-bold tracking-[-0.03em] text-[#111111]">
                  {metric.value}
                </p>
                <p className="mt-0.5 text-[12px] font-medium text-[#8a8580]">{metric.label}</p>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Media side */}
      <div className="relative min-h-0">
        <Image
          src={owner.imageUrl}
          alt=""
          fill
          className="object-cover"
          sizes="520px"
        />
        <button
          type="button"
          className="absolute bottom-3 left-3 flex h-9 w-9 items-center justify-center rounded-full bg-white text-[#111111] shadow-[0_8px_24px_rgba(0,0,0,0.18)] transition-transform hover:scale-105 sm:bottom-4 sm:left-4 sm:h-10 sm:w-10"
          aria-label={`Play video for ${owner.name}`}
        >
          <PlayIcon />
        </button>
      </div>
    </article>
  );
}

export default function TrustedByOwners() {
  const scrollerRef = useRef<HTMLDivElement>(null);

  const scrollByCard = useCallback((dir: -1 | 1) => {
    const el = scrollerRef.current;
    if (!el) return;
    const amount = Math.min(el.clientWidth * 0.92, 1060);
    el.scrollBy({ left: dir * amount, behavior: "smooth" });
  }, []);

  return (
    <section className="overflow-hidden bg-bg py-14 sm:py-18 md:py-20">
      <div className="mx-auto flex max-w-[1100px] items-end justify-between gap-4 px-4 sm:px-8 md:px-12">
        <h2 className="font-display text-[clamp(2rem,4vw,3.25rem)] font-semibold tracking-[-0.03em] text-ink">
          Trusted by owners
        </h2>

        <div className="flex shrink-0 items-center gap-2">
          <button
            type="button"
            onClick={() => scrollByCard(-1)}
            className="flex h-11 w-11 items-center justify-center rounded-full bg-surface text-ink transition-colors hover:bg-parchment"
            aria-label="Previous owner"
          >
            <ChevronLeft />
          </button>
          <button
            type="button"
            onClick={() => scrollByCard(1)}
            className="flex h-11 w-11 items-center justify-center rounded-full bg-surface text-ink transition-colors hover:bg-parchment"
            aria-label="Next owner"
          >
            <ChevronRight />
          </button>
        </div>
      </div>

      <div
        ref={scrollerRef}
        className="mt-8 flex snap-x snap-mandatory gap-4 overflow-x-auto px-4 pb-2 scrollbar-none sm:mt-10 sm:gap-5 sm:px-8 md:px-12"
        style={{ scrollbarWidth: "none", msOverflowStyle: "none" }}
      >
        {trustedOwners.map((owner) => (
          <OwnerCard key={owner.id} owner={owner} />
        ))}
        {/* spacer so last card can align nicely */}
        <div className="w-2 shrink-0 sm:w-4" aria-hidden="true" />
      </div>
    </section>
  );
}
