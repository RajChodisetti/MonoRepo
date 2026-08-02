"use client";

import Image from "next/image";
import { useEffect, useState } from "react";

const QUERY = "Pizza near me";

type Phase = "idle" | "typing" | "results" | "rising" | "hold";

function SearchIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-[18px] w-[18px] shrink-0 text-[#1a1a1a]" aria-hidden="true">
      <circle cx="11" cy="11" r="6.5" fill="none" stroke="currentColor" strokeWidth="1.8" />
      <path d="M16.2 16.2 20 20" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    </svg>
  );
}

function SkeletonRow({ faded = false }: { faded?: boolean }) {
  return (
    <div className={`flex items-center gap-3 px-1 py-2.5 ${faded ? "opacity-40" : "opacity-65"}`}>
      <div className="h-11 w-11 shrink-0 rounded-[12px] bg-[#d9d3cb]" />
      <div className="min-w-0 flex-1 space-y-2">
        <div className="h-2.5 w-[76%] rounded-full bg-[#d9d3cb]" />
        <div className="h-2.5 w-[48%] rounded-full bg-[#d9d3cb]" />
      </div>
    </div>
  );
}

type GoogleTrafficPanelProps = {
  onProgress?: (progress: number) => void;
  onComplete?: () => void;
};

export default function GoogleTrafficPanel({ onProgress, onComplete }: GoogleTrafficPanelProps) {
  const [phase, setPhase] = useState<Phase>("idle");
  const [typed, setTyped] = useState("");

  useEffect(() => {
    let cancelled = false;
    const timers: number[] = [];
    let raf = 0;

    const report = (value: number) => {
      if (!cancelled) onProgress?.(Math.min(1, Math.max(0, value)));
    };

    const wait = (ms: number) =>
      new Promise<void>((resolve) => {
        timers.push(window.setTimeout(resolve, ms));
      });

    const animateProgress = (from: number, to: number, duration: number) =>
      new Promise<void>((resolve) => {
        const start = performance.now();
        const tick = (now: number) => {
          if (cancelled) {
            resolve();
            return;
          }
          const t = Math.min(1, (now - start) / duration);
          const eased = 1 - Math.pow(1 - t, 2);
          report(from + (to - from) * eased);
          if (t < 1) {
            raf = requestAnimationFrame(tick);
          } else {
            report(to);
            resolve();
          }
        };
        raf = requestAnimationFrame(tick);
      });

    const run = async () => {
      setPhase("idle");
      setTyped("");
      report(0);
      await wait(500);
      if (cancelled) return;

      setPhase("typing");
      for (let i = 1; i <= QUERY.length; i += 1) {
        if (cancelled) return;
        setTyped(QUERY.slice(0, i));
        report(0.05 + (i / QUERY.length) * 0.28);
        await wait(75);
      }
      if (cancelled) return;
      await wait(400);
      if (cancelled) return;

      setPhase("results");
      await animateProgress(0.35, 0.48, 900);
      if (cancelled) return;

      setPhase("rising");
      await animateProgress(0.48, 0.96, 3200);
      if (cancelled) return;

      setPhase("hold");
      report(1);
      await wait(700);
      if (cancelled) return;

      onComplete?.();
    };

    void run();

    return () => {
      cancelled = true;
      timers.forEach((id) => window.clearTimeout(id));
      cancelAnimationFrame(raf);
      onProgress?.(0);
    };
  }, [onProgress, onComplete]);

  const showResults = phase === "results" || phase === "rising" || phase === "hold";
  const isRising = phase === "rising" || phase === "hold";

  return (
    <div
      className="grid min-h-[520px] items-center gap-10 rounded-[28px] px-6 py-16 sm:min-h-[560px] sm:rounded-[32px] sm:px-10 sm:py-20 lg:grid-cols-2 lg:gap-12 lg:px-14 lg:py-24"
      style={{ backgroundColor: "#f2ecdf" }}
    >
      <div>
        <p className="text-[14px] font-medium text-[#7a7268] sm:text-[15px]">Upgrade your SEO</p>
        <h3 className="mt-3 text-[clamp(1.85rem,3.2vw,2.85rem)] font-bold leading-[1.12] tracking-[-0.04em] text-[#1a1a1a]">
          <span className="block">With Tuvi, your</span>
          <span className="block">website gets way</span>
          <span className="block">more Google traffic</span>
        </h3>
      </div>

      <div className="relative mx-auto w-full max-w-[380px]">
        <div className="relative z-20 flex items-center gap-3 rounded-2xl bg-white px-4 py-3.5 shadow-[0_10px_40px_rgba(0,0,0,0.1)]">
          <SearchIcon />
          <p className="min-h-[1.25em] flex-1 text-[15px] font-medium tracking-[-0.01em] text-[#1a1a1a]">
            {typed}
            {phase === "typing" ? (
              <span className="ml-0.5 inline-block h-[1.05em] w-[2px] animate-pulse bg-[#1a1a1a] align-[-0.15em]" />
            ) : null}
          </p>
        </div>

        <div
          className={`relative mt-3 overflow-hidden rounded-[22px] bg-[#e7e1d7]/75 p-3 transition-all duration-500 ease-out ${
            showResults ? "translate-y-0 opacity-100" : "pointer-events-none translate-y-3 opacity-0"
          }`}
          style={{ minHeight: 236 }}
        >
          <div
            className={`space-y-0.5 transition-all duration-700 ease-[cubic-bezier(0.22,1,0.36,1)] ${
              isRising ? "pt-[84px]" : "pt-0"
            }`}
          >
            <SkeletonRow />
            <SkeletonRow />
            <SkeletonRow faded />
          </div>

          <div
            className={`absolute left-3 right-3 z-10 flex items-center gap-3.5 rounded-[16px] bg-white p-3.5 shadow-[0_14px_40px_rgba(0,0,0,0.14)] transition-all duration-700 ease-[cubic-bezier(0.22,1,0.36,1)] ${
              isRising ? "scale-100 opacity-100" : "scale-[0.97] opacity-0"
            }`}
            style={{ top: isRising ? 12 : 108 }}
          >
            <div className="relative h-[52px] w-[52px] shrink-0 overflow-hidden rounded-[12px]">
              <Image
                src="https://images.unsplash.com/photo-1513104890138-7c749659a591?auto=format&fit=crop&w=160&q=80"
                alt=""
                fill
                className="object-cover"
                sizes="52px"
              />
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-[15px] font-semibold tracking-[-0.02em] text-[#1a1a1a]">Your restaurant</p>
              <div className="mt-2.5 space-y-2">
                <div className="h-2.5 w-[88%] rounded-full bg-[#ebe7e2]" />
                <div className="h-2.5 w-[56%] rounded-full bg-[#ebe7e2]" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
