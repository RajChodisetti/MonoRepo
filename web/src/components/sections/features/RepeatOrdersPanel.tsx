"use client";

import { useEffect, useMemo, useRef, useState } from "react";

type Beat = {
  id: string;
  label: string;
  detail: string;
  day: string;
};

/** Fictional guest journey — not a real person or venue claim. */
const BEATS: Beat[] = [
  {
    id: "first",
    label: "First order",
    detail: "Jules checks out on your site — guest data stays yours.",
    day: "Day 0",
  },
  {
    id: "offer",
    label: "Timed offer",
    detail: "Tuvi sends a free garlic-bread nudge the next quiet lunch.",
    day: "Day 1",
  },
  {
    id: "picks",
    label: "Dish picks",
    detail: "Email with two favourites Jules already ordered once.",
    day: "Day 3",
  },
  {
    id: "again",
    label: "Orders again",
    detail: "$42.80 first-party ticket — no marketplace fee.",
    day: "Day 4",
  },
  {
    id: "regular",
    label: "Becomes regular",
    detail: "Loyalty unlocked. Third visit this month.",
    day: "Week 2",
  },
];

/** Node positions along a soft S-curve (percent of stage / viewBox). */
const NODES = [
  { x: 8, y: 78 },
  { x: 28, y: 42 },
  { x: 50, y: 70 },
  { x: 72, y: 34 },
  { x: 92, y: 58 },
] as const;

/** Smooth cubic path through nodes (Catmull-Rom → bezier-ish). */
function buildCurvePath(nodes: readonly { x: number; y: number }[]): string {
  if (nodes.length < 2) return "";
  const pts = nodes.map((n) => ({ x: n.x, y: n.y }));
  let d = `M ${pts[0].x} ${pts[0].y}`;
  for (let i = 0; i < pts.length - 1; i += 1) {
    const p0 = pts[i - 1] ?? pts[i];
    const p1 = pts[i];
    const p2 = pts[i + 1];
    const p3 = pts[i + 2] ?? p2;
    const cp1x = p1.x + (p2.x - p0.x) / 6;
    const cp1y = p1.y + (p2.y - p0.y) / 6;
    const cp2x = p2.x - (p3.x - p1.x) / 6;
    const cp2y = p2.y - (p3.y - p1.y) / 6;
    d += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p2.x} ${p2.y}`;
  }
  return d;
}

const PATH_D = buildCurvePath(NODES);

const TRAVEL_MS = 900;
const HOLD_MS = 950;
const FINISH_HOLD_MS = 1700;
const LOOP_PAUSE_MS = 380;

function easeInOutCubic(t: number) {
  return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
}

type RepeatOrdersPanelProps = {
  onProgress?: (progress: number) => void;
  onComplete?: () => void;
};

export default function RepeatOrdersPanel({ onProgress, onComplete }: RepeatOrdersPanelProps) {
  const pathRef = useRef<SVGPathElement>(null);
  const [progress, setProgress] = useState(0); // 0 → 1 along full path
  const [cardStep, setCardStep] = useState(0);
  const [cardVisible, setCardVisible] = useState(true);
  const [tokenPos, setTokenPos] = useState<{ x: number; y: number }>({
    x: NODES[0].x,
    y: NODES[0].y,
  });

  const activeStep = useMemo(() => {
    const idx = Math.round(progress * (BEATS.length - 1));
    return Math.min(BEATS.length - 1, Math.max(0, idx));
  }, [progress]);

  // Keep avatar glued to the SVG path length
  useEffect(() => {
    const path = pathRef.current;
    if (!path) return;
    const len = path.getTotalLength();
    if (!len) return;
    const point = path.getPointAtLength(progress * len);
    setTokenPos({ x: point.x, y: point.y });
  }, [progress]);

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

    const animateProgress = (from: number, to: number, ms: number) =>
      new Promise<void>((resolve) => {
        const start = performance.now();
        const tick = (now: number) => {
          if (cancelled) {
            resolve();
            return;
          }
          const t = Math.min(1, (now - start) / ms);
          const eased = easeInOutCubic(t);
          const value = from + (to - from) * eased;
          setProgress(value);
          report(value);
          if (t < 1) {
            raf = requestAnimationFrame(tick);
          } else {
            setProgress(to);
            report(to);
            resolve();
          }
        };
        raf = requestAnimationFrame(tick);
      });

    const showCard = async (index: number) => {
      setCardVisible(false);
      await wait(160);
      if (cancelled) return;
      setCardStep(index);
      setCardVisible(true);
      await wait(HOLD_MS);
    };

    const run = async () => {
      setProgress(0);
      setCardStep(0);
      setCardVisible(true);
      setTokenPos({ x: NODES[0].x, y: NODES[0].y });
      report(0);
      await wait(LOOP_PAUSE_MS);
      if (cancelled) return;

      // Settle on first stop
      await showCard(0);
      if (cancelled) return;

      for (let i = 1; i < BEATS.length; i += 1) {
        if (cancelled) return;
        const from = (i - 1) / (BEATS.length - 1);
        const to = i / (BEATS.length - 1);
        setCardVisible(false);
        await animateProgress(from, to, TRAVEL_MS);
        if (cancelled) return;
        await showCard(i);
      }

      if (cancelled) return;
      report(1);
      await wait(FINISH_HOLD_MS);
      if (cancelled) return;
      onComplete?.();
    };

    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      timers.push(
        window.setTimeout(() => {
          if (cancelled) return;
          const lastIndex = BEATS.length - 1;
          setProgress(1);
          setCardStep(lastIndex);
          setCardVisible(true);
          setTokenPos({ x: NODES[lastIndex].x, y: NODES[lastIndex].y });
          report(1);
        }, 0),
      );
    } else {
      void run();
    }

    return () => {
      cancelled = true;
      cancelAnimationFrame(raf);
      timers.forEach((id) => window.clearTimeout(id));
      onProgress?.(0);
    };
  }, [onProgress, onComplete]);

  const beat = BEATS[cardStep] ?? BEATS[0];

  return (
    <div
      className="grid min-h-[520px] items-center gap-10 rounded-[28px] px-6 py-14 sm:min-h-[560px] sm:rounded-[32px] sm:px-10 sm:py-16 lg:grid-cols-2 lg:gap-12 lg:px-14 lg:py-20"
      style={{ backgroundColor: "#ebe4d6" }}
    >
      <div>
        <p className="text-[14px] font-medium text-[#7a7268] sm:text-[15px]">Guests who come back</p>
        <h3 className="mt-3 text-[clamp(1.85rem,3.2vw,2.85rem)] font-bold leading-[1.12] tracking-[-0.04em] text-[#1a1a1a]">
          <span className="block">Turn one order into</span>
          <span className="block">a regular — with</span>
          <span className="block">timed, personal nudges</span>
        </h3>
        <p className="mt-4 max-w-[34ch] text-[15px] leading-relaxed text-[#7a7268]">
          Watch Jules move from first checkout to loyal regular — each stop is an automatic Tuvi touchpoint on your brand.
        </p>
      </div>

      <div className="relative mx-auto w-full max-w-[440px]">
        <div
          className={`relative z-20 mb-3 transition-all duration-500 ease-[cubic-bezier(0.22,1,0.36,1)] motion-reduce:transition-none ${
            cardVisible ? "translate-y-0 scale-100 opacity-100" : "translate-y-3 scale-[0.98] opacity-0"
          }`}
        >
          <div className="overflow-hidden rounded-[22px] bg-[#1a1a1a] px-5 py-4 text-[#fff8ee] shadow-[0_20px_44px_rgba(26,26,26,0.28)]">
            <div className="flex items-center justify-between gap-3">
              <p className="text-[11px] font-semibold uppercase tracking-[0.16em] text-[#c45c26]">
                {beat.day}
              </p>
              <p className="text-[11px] tabular-nums text-[#fff8ee]/50">
                Stop {cardStep + 1} / {BEATS.length}
              </p>
            </div>
            <p className="mt-2 font-display text-[1.35rem] font-semibold tracking-[-0.02em]">
              {beat.label}
            </p>
            <p className="mt-1.5 text-[14px] leading-snug text-[#fff8ee]/72">{beat.detail}</p>
          </div>
          <div
            className="mx-auto h-0 w-0 border-x-[10px] border-t-[12px] border-x-transparent border-t-[#1a1a1a]"
            aria-hidden
          />
        </div>

        <div
          className="relative aspect-[5/3.2] overflow-hidden rounded-[24px] border border-[#1a1a1a]/08"
          style={{
            background:
              "radial-gradient(ellipse at 30% 20%, #fff8ee 0%, #e7dfd0 45%, #ddd3c2 100%)",
          }}
        >
          <div
            className="pointer-events-none absolute -left-8 top-6 h-28 w-40 rounded-full bg-[#c45c26]/08 blur-2xl"
            aria-hidden
          />
          <div
            className="pointer-events-none absolute bottom-4 right-4 h-24 w-32 rounded-full bg-[#1a1a1a]/06 blur-2xl"
            aria-hidden
          />

          <svg
            viewBox="0 0 100 100"
            className="absolute inset-0 h-full w-full"
            preserveAspectRatio="none"
            aria-hidden
          >
            <path
              d={PATH_D}
              fill="none"
              stroke="#1a1a1a"
              strokeOpacity="0.12"
              strokeWidth="1.8"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeDasharray="2.4 2"
              vectorEffect="non-scaling-stroke"
            />
            <path
              ref={pathRef}
              d={PATH_D}
              fill="none"
              stroke="#c45c26"
              strokeWidth="2.4"
              strokeLinecap="round"
              strokeLinejoin="round"
              pathLength={1}
              strokeDasharray={1}
              strokeDashoffset={1 - progress}
              vectorEffect="non-scaling-stroke"
            />
          </svg>

          {NODES.map((n, i) => {
            const threshold = i / (BEATS.length - 1);
            const done = progress >= threshold - 0.001;
            const active = activeStep === i && Math.abs(progress - threshold) < 0.04;
            return (
              <div
                key={BEATS[i].id}
                className="absolute -translate-x-1/2 -translate-y-1/2"
                style={{ left: `${n.x}%`, top: `${n.y}%` }}
              >
                <div
                  className={`flex h-8 w-8 items-center justify-center rounded-full text-[11px] font-bold transition-all duration-700 ease-[cubic-bezier(0.22,1,0.36,1)] motion-reduce:transition-none ${
                    active
                      ? "scale-110 bg-[#c45c26] text-[#fff8ee] shadow-[0_0_0_7px_rgba(196,92,38,0.2)]"
                      : done
                        ? "scale-100 bg-[#1a1a1a] text-[#fff8ee]"
                        : "scale-95 bg-[#fff8ee] text-[#8a8074] ring-1 ring-[#1a1a1a]/15"
                  }`}
                >
                  {i + 1}
                </div>
                <p
                  className={`mt-1.5 max-w-[72px] text-center text-[10px] font-semibold leading-tight tracking-[-0.01em] transition-colors duration-500 motion-reduce:transition-none ${
                    active || done ? "text-[#1a1a1a]" : "text-[#8a8074]"
                  }`}
                >
                  {BEATS[i].label}
                </p>
              </div>
            );
          })}

          <div
            className="absolute z-10 will-change-transform"
            style={{
              left: `${tokenPos.x}%`,
              top: `${tokenPos.y}%`,
              transform: "translate(-50%, -118%)",
            }}
          >
            <div
              className="relative flex h-12 w-12 items-center justify-center overflow-hidden rounded-full bg-[#1a1a1a] ring-[3px] ring-[#fff8ee] shadow-[0_12px_28px_rgba(26,26,26,0.28)]"
              aria-hidden="true"
            >
              <span className="font-display text-[18px] font-semibold text-[#fff8ee]">J</span>
            </div>
            <div className="absolute -bottom-1 left-1/2 h-2.5 w-2.5 -translate-x-1/2 rotate-45 bg-[#c45c26]" />
          </div>
        </div>

        <p className="mt-3 text-center text-[12px] text-[#8a8074]">
          Jules · Quillnest Kitchen · first-party only
        </p>
      </div>
    </div>
  );
}
