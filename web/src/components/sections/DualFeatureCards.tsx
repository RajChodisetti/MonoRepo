"use client";

import Image from "next/image";
import { useEffect, useState } from "react";

/* ─── Left: Google rankings animation ─── */

const QUERY = "Tacos near me";

type RankPhase = "idle" | "typing" | "results" | "rising" | "hold";

function SearchIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-[18px] w-[18px] shrink-0 text-[#1a1a1a]" aria-hidden="true">
      <circle cx="11" cy="11" r="6.5" fill="none" stroke="currentColor" strokeWidth="1.8" />
      <path d="M16.2 16.2 20 20" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  );
}

function StarRow() {
  return (
    <span className="inline-flex items-center gap-0.5 text-[#f5a623]" aria-hidden="true">
      {Array.from({ length: 5 }).map((_, i) => (
        <svg key={i} viewBox="0 0 12 12" className="h-3 w-3">
          <path
            fill="currentColor"
            d="M6 1.1l1.35 2.74 3.02.44-2.18 2.13.51 3-2.7-1.42L3.3 9.41l.51-3L1.63 4.28l3.02-.44L6 1.1z"
          />
        </svg>
      ))}
    </span>
  );
}

function GreenSkeleton({ faded = false }: { faded?: boolean }) {
  return (
    <div className={`flex items-center gap-3 px-1 py-2.5 ${faded ? "opacity-35" : "opacity-55"}`}>
      <div className="h-11 w-11 shrink-0 rounded-[12px] bg-white/25" />
      <div className="min-w-0 flex-1 space-y-2">
        <div className="h-2.5 w-[76%] rounded-full bg-white/25" />
        <div className="h-2.5 w-[48%] rounded-full bg-white/25" />
      </div>
    </div>
  );
}

function RankingCard() {
  const [phase, setPhase] = useState<RankPhase>("idle");
  const [typed, setTyped] = useState("");

  useEffect(() => {
    let cancelled = false;
    const timers: number[] = [];

    const wait = (ms: number) =>
      new Promise<void>((resolve) => {
        timers.push(window.setTimeout(resolve, ms));
      });

    const run = async () => {
      while (!cancelled) {
        setPhase("idle");
        setTyped("");
        await wait(600);
        if (cancelled) break;

        setPhase("typing");
        for (let i = 1; i <= QUERY.length; i += 1) {
          if (cancelled) break;
          setTyped(QUERY.slice(0, i));
          await wait(75);
        }
        if (cancelled) break;
        await wait(450);
        if (cancelled) break;

        setPhase("results");
        await wait(900);
        if (cancelled) break;

        setPhase("rising");
        await wait(2800);
        if (cancelled) break;

        setPhase("hold");
        await wait(1600);
      }
    };

    void run();
    return () => {
      cancelled = true;
      timers.forEach((id) => window.clearTimeout(id));
    };
  }, []);

  const showResults = phase === "results" || phase === "rising" || phase === "hold";
  const isRising = phase === "rising" || phase === "hold";

  return (
    <div
      className="relative flex min-h-[560px] flex-col overflow-hidden rounded-[28px] px-5 pb-7 pt-8 sm:min-h-[620px] sm:rounded-[32px] sm:px-7 sm:pb-8 sm:pt-10 md:min-h-[660px]"
      style={{
        background:
          "linear-gradient(180deg, #3d8f6e 0%, #2f6b54 42%, #174c3a 78%, #0f271f 100%)",
      }}
    >
      <div className="relative mx-auto w-full max-w-[340px] flex-1">
        <div className="relative z-20 flex items-center gap-3 rounded-2xl bg-white px-4 py-3.5 shadow-[0_10px_40px_rgba(0,0,0,0.12)]">
          <SearchIcon />
          <p className="min-h-[1.25em] flex-1 text-[15px] font-medium tracking-[-0.01em] text-[#1a1a1a]">
            {typed}
            {phase === "typing" ? (
              <span className="ml-0.5 inline-block h-[1.05em] w-[2px] animate-pulse bg-[#1a1a1a] align-[-0.15em]" />
            ) : null}
          </p>
        </div>

        <div
          className={`relative mt-3 overflow-hidden rounded-[22px] bg-white/20 p-3 backdrop-blur-[2px] transition-all duration-500 ease-out ${
            showResults ? "translate-y-0 opacity-100" : "pointer-events-none translate-y-3 opacity-0"
          }`}
          style={{ minHeight: 236 }}
        >
          <div
            className={`space-y-0.5 transition-all duration-700 ease-[cubic-bezier(0.22,1,0.36,1)] ${
              isRising ? "pt-[92px]" : "pt-0"
            }`}
          >
            <GreenSkeleton />
            <GreenSkeleton />
            <GreenSkeleton faded />
          </div>

          <div
            className={`absolute left-3 right-3 z-10 flex items-center gap-3.5 rounded-[16px] bg-white p-3.5 shadow-[0_14px_40px_rgba(0,0,0,0.16)] transition-all duration-700 ease-[cubic-bezier(0.22,1,0.36,1)] ${
              isRising ? "scale-100 opacity-100" : "scale-[0.97] opacity-0"
            }`}
            style={{ top: isRising ? 12 : 108 }}
          >
            <div className="relative h-[52px] w-[52px] shrink-0 overflow-hidden rounded-[12px]">
              <Image src="/menu/tacos.jpg" alt="" fill className="object-cover" sizes="52px" />
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-[15px] font-semibold tracking-[-0.02em] text-[#1a1a1a]">Talkin&apos; Tacos</p>
              <div className="mt-1.5 flex items-center gap-1.5">
                <span className="text-[12px] font-semibold text-[#1a1a1a]">4.8</span>
                <StarRow />
              </div>
              <div className="mt-2 space-y-1.5">
                <div className="h-2 w-[88%] rounded-full bg-[#ebe7e2]" />
                <div className="h-2 w-[56%] rounded-full bg-[#ebe7e2]" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <p className="relative z-10 mt-8 max-w-[22ch] text-[clamp(1.15rem,2vw,1.35rem)] font-semibold leading-snug tracking-[-0.025em] text-white">
        Get higher Google rankings with your AI-powered restaurant website.
      </p>
    </div>
  );
}

/* ─── Right: static product card ─── */

function PlusIcon() {
  return (
    <svg viewBox="0 0 20 20" className="h-5 w-5" aria-hidden="true">
      <path d="M10 4v12M4 10h12" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  );
}

function OrderingCard() {
  return (
    <div
      className="relative flex min-h-[560px] flex-col overflow-hidden rounded-[28px] px-5 pb-7 pt-8 sm:min-h-[620px] sm:rounded-[32px] sm:px-7 sm:pb-8 sm:pt-10 md:min-h-[660px]"
      style={{ backgroundColor: "#f2ecdf" }}
    >
      <div className="flex flex-1 items-center justify-center px-2">
        <div className="flex w-full max-w-[340px] items-center gap-3.5 rounded-[22px] bg-bg p-3.5 shadow-[0_14px_40px_rgba(15,39,31,0.08)] ring-1 ring-border">
          <div className="relative h-[88px] w-[88px] shrink-0 overflow-hidden rounded-2xl bg-sage">
            <Image src="/menu/birria-tacos.jpg" alt="" fill className="object-cover" sizes="88px" />
            <span className="absolute left-2 top-2 rounded-full bg-accent px-2 py-0.5 text-[10px] font-bold text-bg">
              New
            </span>
          </div>

          <div className="min-w-0 flex-1">
            <p className="text-[16px] font-semibold tracking-[-0.02em] text-ink">Birria Tacos</p>
            <div className="mt-2.5 space-y-2">
              <div className="h-2 w-[88%] rounded-full bg-sage" />
              <div className="h-2 w-[58%] rounded-full bg-sage" />
            </div>
            <p className="mt-2.5 text-[14px] font-medium text-secondary">$13.99</p>
          </div>

          <span className="mb-0.5 ml-1 flex h-10 w-10 shrink-0 self-end items-center justify-center rounded-full bg-bg text-ink ring-1 ring-border">
            <PlusIcon />
          </span>
        </div>
      </div>

      <p className="relative z-10 mt-8 max-w-[24ch] text-[clamp(1.15rem,2vw,1.35rem)] font-semibold leading-snug tracking-[-0.025em] text-ink">
        Grow your sales with an online ordering system modeled after the big brands.
      </p>
    </div>
  );
}

export default function DualFeatureCards() {
  return (
    <section className="bg-bg px-4 pb-16 pt-4 sm:px-8 sm:pb-20 sm:pt-6 md:px-12">
      <div className="mx-auto grid max-w-[1100px] gap-4 sm:gap-5 lg:grid-cols-2">
        <RankingCard />
        <OrderingCard />
      </div>
    </section>
  );
}
