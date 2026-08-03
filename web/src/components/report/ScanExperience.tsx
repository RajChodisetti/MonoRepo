"use client";

import { useEffect, useMemo, useRef, useState } from "react";

const STEP_INTERVAL_MS = 2200;
const FINISH_BEAT_MS = 700;
const TARGET_SECONDS = 28;
const SAFETY_EXIT_MS = 2500;

function CheckIcon() {
  return (
    <svg viewBox="0 0 16 16" className="h-3.5 w-3.5" fill="none" aria-hidden="true">
      <path
        d="M3.5 8.2 6.4 11l6.1-6.5"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function Spinner() {
  return (
    <svg className="scan-spinner h-4 w-4 text-primary" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeOpacity="0.2" strokeWidth="2.5" />
      <path d="M21 12a9 9 0 0 0-9-9" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" />
    </svg>
  );
}

function StarRow({ rating }: { rating?: number }) {
  const value = typeof rating === "number" && rating > 0 ? rating : 0;
  return (
    <div className="flex items-center gap-0.5" aria-label={value ? `${value} stars` : "No rating"}>
      {Array.from({ length: 5 }, (_, i) => {
        const filled = value >= i + 1 || value >= i + 0.75;
        const half = !filled && value >= i + 0.35;
        return (
          <svg key={i} viewBox="0 0 12 12" className="h-3.5 w-3.5" aria-hidden="true">
            <path
              d="M6 1.1 7.3 4.4l3.6.3-2.8 2.3.9 3.5L6 8.6 3 10.5l.9-3.5L1.1 4.7l3.6-.3L6 1.1Z"
              className={filled ? "fill-[#f5b942]" : half ? "fill-[#f5b942]/60" : "fill-[#e4dfd7]"}
            />
          </svg>
        );
      })}
    </div>
  );
}

function MapPinTile() {
  return (
    <div className="relative flex min-h-[110px] min-w-0 flex-1 self-stretch items-center justify-center overflow-hidden rounded-2xl bg-[#eef2ef] ring-1 ring-black/5 sm:min-h-[140px] lg:min-h-[150px]">
      <div
        className="absolute inset-0 opacity-80"
        style={{
          backgroundImage:
            "linear-gradient(#d7e0d9 1px, transparent 1px), linear-gradient(90deg, #d7e0d9 1px, transparent 1px)",
          backgroundSize: "18px 18px",
        }}
      />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_60%_40%,rgba(47,107,84,0.16),transparent_55%)]" />
      <div className="relative z-[1] flex h-12 w-12 items-center justify-center rounded-full bg-ink text-bg shadow-[0_8px_20px_rgba(15,39,31,0.25)] sm:h-14 sm:w-14">
        <svg viewBox="0 0 16 16" className="h-5 w-5 fill-current sm:h-6 sm:w-6" aria-hidden="true">
          <path d="M8 1.5c-2.6 0-4.7 2-4.7 4.6 0 3.2 4.1 7.8 4.3 8 .2.2.5.2.7 0 .2-.2 4.4-4.8 4.4-8 0-2.6-2.1-4.6-4.7-4.6Zm0 6.3A1.8 1.8 0 1 1 8 4.2a1.8 1.8 0 0 1 0 3.6Z" />
        </svg>
      </div>
    </div>
  );
}

export type ScanExperienceProps = {
  restaurantName?: string;
  rating?: number;
  category?: string;
  photoUrl?: string;
  fetchComplete?: boolean;
  onReady?: () => void;
  className?: string;
};

export default function ScanExperience({
  restaurantName = "Your restaurant",
  rating,
  category = "Restaurant",
  photoUrl,
  fetchComplete = false,
  onReady,
  className = "",
}: ScanExperienceProps) {
  const steps = useMemo(
    () => [
      `${restaurantName} & competitors`,
      "Google business profile",
      "Google review sentiment",
      "Photo quality and quantity",
      "Website / listing signals",
      "Mobile experience",
    ],
    [restaurantName],
  );

  const [activeIndex, setActiveIndex] = useState(0);
  const [secondsLeft, setSecondsLeft] = useState(TARGET_SECONDS);
  const [finishing, setFinishing] = useState(false);
  const onReadyRef = useRef(onReady);
  onReadyRef.current = onReady;

  // Countdown copy
  useEffect(() => {
    const started = Date.now();
    const id = window.setInterval(() => {
      const elapsed = Math.floor((Date.now() - started) / 1000);
      setSecondsLeft(Math.max(0, TARGET_SECONDS - elapsed));
    }, 400);
    return () => window.clearInterval(id);
  }, []);

  // Advance steps while waiting
  useEffect(() => {
    if (fetchComplete) return;
    const id = window.setInterval(() => {
      setActiveIndex((i) => Math.min(i + 1, steps.length - 1));
    }, STEP_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [fetchComplete, steps.length]);

  // When fetch completes: finish checklist then exit.
  // Do NOT guard on `finishing` state — Strict Mode remount clears the timeout;
  // a state guard would block the second effect and leave the UI stuck on "Wrapping up…".
  useEffect(() => {
    if (!fetchComplete) return;

    setFinishing(true);
    setActiveIndex(steps.length);

    let cancelled = false;
    const beat = window.setTimeout(() => {
      if (!cancelled) onReadyRef.current?.();
    }, FINISH_BEAT_MS);

    // Absolute safety net if something interrupts the beat timer
    const safety = window.setTimeout(() => {
      if (!cancelled) onReadyRef.current?.();
    }, SAFETY_EXIT_MS);

    return () => {
      cancelled = true;
      window.clearTimeout(beat);
      window.clearTimeout(safety);
    };
  }, [fetchComplete, steps.length]);

  const progress = finishing
    ? 1
    : Math.min(0.92, (activeIndex + 0.35) / steps.length);

  return (
    <div className={`hero-atmosphere relative min-h-[calc(100dvh-4.5rem)] overflow-hidden ${className}`}>
      <div className="hero-grid pointer-events-none absolute inset-0 opacity-40" aria-hidden="true" />
      <div
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(36rem_24rem_at_55%_45%,rgba(47,107,84,0.07),transparent_62%)]"
        aria-hidden="true"
      />

      <div className="relative grid min-h-[calc(100dvh-4.5rem)] w-full lg:grid-cols-[260px_minmax(0,1fr)] xl:grid-cols-[280px_minmax(0,1fr)]">
        <aside className="relative z-10 flex flex-col justify-center border-b border-border/70 bg-bg/95 px-5 py-6 backdrop-blur-md sm:px-6 lg:border-b-0 lg:border-r lg:px-5 lg:py-7">
          <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-muted">AI grader</p>
          <h1 className="mt-1.5 font-display text-[1.65rem] font-semibold leading-tight tracking-[-0.035em] text-ink lg:text-[1.85rem]">
            Scanning…
          </h1>
          <p className="mt-1.5 text-[13px] leading-snug text-muted">
            Free SEO report — rankings, reviews, website &amp; listing signals.
          </p>

          <ul className="mt-6 space-y-2.5">
            {steps.map((label, index) => {
              const done = finishing || index < activeIndex;
              const active = !done && index === activeIndex;
              return (
                <li key={`${index}-${label}`} className="flex items-start gap-2.5">
                  <span
                    className={`mt-0.5 flex h-[20px] w-[20px] shrink-0 items-center justify-center rounded-full transition-all duration-300 ${
                      done
                        ? "bg-ink text-bg shadow-sm"
                        : active
                          ? "scan-step-active border-2 border-primary bg-white text-primary"
                          : "border border-[#d8d2c8] bg-transparent"
                    }`}
                  >
                    {done ? <CheckIcon /> : null}
                  </span>
                  <span
                    className={`pt-px text-[13.5px] leading-snug transition-colors duration-300 ${
                      done || active ? "font-medium text-ink" : "text-muted/65"
                    }`}
                  >
                    {label}
                  </span>
                </li>
              );
            })}
          </ul>

          <div className="mt-6 border-t border-border/60 pt-4">
            <div className="h-1.5 overflow-hidden rounded-full bg-[#ebe6de]">
              <div
                className="scan-progress-fill h-full rounded-full transition-[width] duration-500 ease-out"
                style={{ width: `${Math.max(10, progress * 100)}%` }}
              />
            </div>
            <div className="mt-2.5 flex items-center gap-2 text-[12.5px] text-muted">
              <Spinner />
              <span className="font-medium">
                {finishing
                  ? "Wrapping up your report…"
                  : secondsLeft > 0
                    ? `About ${secondsLeft}s remaining`
                    : "Almost done…"}
              </span>
            </div>
          </div>
        </aside>

        <section className="relative flex min-h-0 min-w-0 flex-col items-center justify-center p-4 sm:p-5 lg:p-6">
          <div className="scan-card-float relative w-full max-w-[720px]">
            <div
              className="pointer-events-none absolute inset-0 z-20 overflow-hidden rounded-[22px]"
              aria-hidden="true"
            >
              <div className="scan-beam absolute left-0 right-0 h-[3px]" />
            </div>

            <article className="relative z-10 flex w-full flex-col overflow-hidden rounded-[22px] border border-border/50 bg-white p-5 shadow-[0_22px_70px_rgba(15,39,31,0.12)] sm:p-6 lg:p-7">
              <div className="flex w-full gap-3 sm:gap-4">
                <div className="relative h-[110px] w-[110px] shrink-0 overflow-hidden rounded-2xl bg-[#efebe6] ring-1 ring-black/5 sm:h-[140px] sm:w-[140px] lg:h-[150px] lg:w-[150px]">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={
                      photoUrl ||
                      "https://images.unsplash.com/photo-1414235077428-338989a2e8c0?auto=format&fit=crop&w=400&q=80"
                    }
                    alt=""
                    className="h-full w-full object-cover"
                  />
                </div>
                <MapPinTile />
              </div>

              <div className="mt-5 sm:mt-6">
                <h2 className="font-display text-[1.55rem] font-semibold leading-tight tracking-[-0.03em] text-ink sm:text-[1.85rem] lg:text-[2rem]">
                  {restaurantName}
                </h2>

                <div className="mt-2.5 flex flex-wrap items-center gap-2 text-[14px] text-muted sm:mt-3 sm:text-[15px]">
                  <StarRow rating={rating} />
                  {typeof rating === "number" && rating > 0 ? (
                    <span className="font-semibold tabular-nums text-ink">{rating.toFixed(1)}</span>
                  ) : null}
                  <span className="text-[#d0c9bd]">|</span>
                  <span className="capitalize">{category}</span>
                </div>
              </div>
            </article>
          </div>
        </section>
      </div>
    </div>
  );
}
