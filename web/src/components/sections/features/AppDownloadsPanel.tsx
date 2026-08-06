"use client";

import Image from "next/image";
import { useEffect } from "react";

const CYCLE_MS = 5200;
const HOLD_MS = 1600;

type AppDownloadsPanelProps = {
  onProgress?: (progress: number) => void;
  onComplete?: () => void;
};

export default function AppDownloadsPanel({ onProgress, onComplete }: AppDownloadsPanelProps) {
  useEffect(() => {
    let cancelled = false;
    let raf = 0;
    const timers: number[] = [];

    const report = (value: number) => {
      if (!cancelled) onProgress?.(Math.min(1, Math.max(0, value)));
    };

    const wait = (ms: number) =>
      new Promise<void>((resolve) => {
        timers.push(window.setTimeout(resolve, ms));
      });

    const animateTo = (from: number, to: number, ms: number) =>
      new Promise<void>((resolve) => {
        const start = performance.now();
        const tick = (now: number) => {
          if (cancelled) {
            resolve();
            return;
          }
          const t = Math.min(1, (now - start) / ms);
          const eased = 1 - Math.pow(1 - t, 3);
          report(from + (to - from) * eased);
          if (t < 1) {
            raf = requestAnimationFrame(tick);
          } else {
            resolve();
          }
        };
        raf = requestAnimationFrame(tick);
      });

    const run = async () => {
      report(0);
      await wait(400);
      if (cancelled) return;

      await animateTo(0, 1, CYCLE_MS);
      if (cancelled) return;

      report(1);
      await wait(HOLD_MS);
      if (cancelled) return;

      onComplete?.();
    };

    void run();

    return () => {
      cancelled = true;
      cancelAnimationFrame(raf);
      timers.forEach((id) => window.clearTimeout(id));
      onProgress?.(0);
    };
  }, [onProgress, onComplete]);

  return (
    <div className="relative min-h-[520px] overflow-hidden rounded-[28px] sm:min-h-[560px] sm:rounded-[32px]">
      <Image
        src="/app/app-downloads-burger.png"
        alt="Fresh burger on a restaurant table"
        fill
        priority
        className="object-cover object-center"
        sizes="(max-width: 1100px) 100vw, 1100px"
      />

      <div
        className="absolute inset-0 bg-gradient-to-r from-black/80 via-black/50 to-black/20 sm:from-black/75 sm:via-black/40 sm:to-transparent"
        aria-hidden="true"
      />

      <div className="relative z-10 flex min-h-[520px] items-center px-6 py-14 sm:min-h-[560px] sm:px-10 sm:py-16 lg:px-14 lg:py-20">
        <div className="max-w-[440px]">
          <p className="text-[14px] font-medium text-white/75 sm:text-[15px]">Your branded app</p>
          <h3 className="mt-3 text-[clamp(1.85rem,3.2vw,2.85rem)] font-bold leading-[1.12] tracking-[-0.04em] text-white">
            <span className="block">Earn loyalty every</span>
            <span className="block">time guests order</span>
            <span className="block">under your brand</span>
          </h3>
          <p className="mt-4 max-w-[36ch] text-[15px] leading-relaxed text-white/70">
            Points, pushes, and reorder shortcuts — on an app that looks like your restaurant, not a marketplace.
          </p>
        </div>
      </div>
    </div>
  );
}
