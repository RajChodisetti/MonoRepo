"use client";

import Image from "next/image";
import { useEffect, useState } from "react";

type StepType = "chip" | "wait" | "pill" | "note" | "dot";

type Step = {
  type: StepType;
  label?: string;
  sublabel?: string;
  icon?: "tag" | "flame" | "mail";
  /** Nodes that receive the blue hover glow when the strip reaches them */
  highlightable?: boolean;
};

const TIMELINE: Step[] = [
  { type: "chip", label: "Ciara", sublabel: "New customer", highlightable: true },
  { type: "wait", label: "wait 1 day" },
  { type: "pill", label: "Sent special offer", icon: "tag", highlightable: true },
  { type: "pill", label: "Sent recommended dishes email", icon: "flame", highlightable: true },
  { type: "note", label: "Ciara orders again" },
  { type: "dot" },
  { type: "wait", label: "wait 1 day" },
  { type: "pill", label: "Sent upcoming holiday special", icon: "mail", highlightable: true },
  { type: "note", label: "Ciara orders again & becomes a regular" },
];

/** Indices of highlightable steps — strip pauses / glows on these */
const HIGHLIGHT_INDICES = TIMELINE.map((s, i) => (s.highlightable ? i : -1)).filter((i) => i >= 0);

const STEP_MS = 850;
const HOLD_MS = 1800;
const LOOP_PAUSE_MS = 400;

function TagIcon({ active }: { active?: boolean }) {
  const c = active ? "#174c3a" : "#9b6b38";
  return (
    <svg viewBox="0 0 20 20" className="h-4 w-4" aria-hidden="true">
      <path
        d="M3 3h6.2c.4 0 .8.2 1.1.5l6 6a1.6 1.6 0 0 1 0 2.2l-4.8 4.8a1.6 1.6 0 0 1-2.2 0l-6-6A1.6 1.6 0 0 1 3 9.4V3Z"
        fill="none"
        stroke={c}
        strokeWidth="1.5"
      />
      <circle cx="6.6" cy="6.6" r="1.1" fill={c} />
    </svg>
  );
}

function FlameIcon({ active }: { active?: boolean }) {
  const c = active ? "#174c3a" : "#2f6b54";
  return (
    <svg viewBox="0 0 20 20" className="h-4 w-4" aria-hidden="true">
      <path
        d="M10 2.5c1.2 2.2-.2 3.6-1.2 4.6C7.5 8.4 6.8 9.4 6.8 11a3.2 3.2 0 0 0 6.4 0c0-1.4-.5-2.4-1.4-3.4-.7-.8-1.5-1.6-1.3-3.1.1-.6.4-1.3.5-2Z"
        fill={c}
        opacity="0.95"
      />
      <path
        d="M10 8.2c.6 1-.1 1.7-.6 2.2-.5.5-.9 1-.9 1.8a1.6 1.6 0 0 0 3.2 0c0-.7-.3-1.2-.7-1.7-.4-.4-.8-.8-.7-1.6.05-.3.2-.6.25-.7Z"
        fill="#fff"
        opacity="0.55"
      />
    </svg>
  );
}

function MailIcon({ active }: { active?: boolean }) {
  const c = active ? "#174c3a" : "#7b927f";
  return (
    <svg viewBox="0 0 20 20" className="h-4 w-4" aria-hidden="true">
      <rect x="3" y="5" width="14" height="10" rx="2" fill="none" stroke={c} strokeWidth="1.5" />
      <path d="m4 6.5 6 4 6-4" fill="none" stroke={c} strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

function StepIcon({ icon, active }: { icon?: Step["icon"]; active?: boolean }) {
  if (icon === "tag") return <TagIcon active={active} />;
  if (icon === "flame") return <FlameIcon active={active} />;
  if (icon === "mail") return <MailIcon active={active} />;
  return null;
}

type RepeatOrdersPanelProps = {
  onProgress?: (progress: number) => void;
  onComplete?: () => void;
};

export default function RepeatOrdersPanel({ onProgress, onComplete }: RepeatOrdersPanelProps) {
  /** Index of the currently hovered/glowing highlightable node, or -1 */
  const [activeIdx, setActiveIdx] = useState(-1);
  /** How far the blue strip has filled (0–1) along the full timeline */
  const [fill, setFill] = useState(0);

  useEffect(() => {
    let cancelled = false;
    const timers: number[] = [];

    const report = (value: number) => {
      if (!cancelled) onProgress?.(Math.min(1, Math.max(0, value)));
    };

    const wait = (ms: number) =>
      new Promise<void>((resolve) => {
        timers.push(window.setTimeout(resolve, ms));
      });

    /**
     * Animate blue strip from `from` → `to` (0–1) over `ms`,
     * updating fill continuously via rAF.
     */
    const animateFill = (from: number, to: number, ms: number) =>
      new Promise<void>((resolve) => {
        const start = performance.now();
        const tick = (now: number) => {
          if (cancelled) {
            resolve();
            return;
          }
          const t = Math.min(1, (now - start) / ms);
          const eased = t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
          const value = from + (to - from) * eased;
          setFill(value);
          report(value);
          if (t < 1) {
            requestAnimationFrame(tick);
          } else {
            resolve();
          }
        };
        requestAnimationFrame(tick);
      });

    const run = async () => {
      setActiveIdx(-1);
      setFill(0);
      report(0);
      await wait(LOOP_PAUSE_MS);
      if (cancelled) return;

      // Walk highlightable nodes: strip flows to each, then glow that box
      for (let h = 0; h < HIGHLIGHT_INDICES.length; h += 1) {
        if (cancelled) return;
        const idx = HIGHLIGHT_INDICES[h];
        const target = (idx + 0.5) / TIMELINE.length;
        const from = h === 0 ? 0 : (HIGHLIGHT_INDICES[h - 1] + 0.5) / TIMELINE.length;

        setActiveIdx(-1);
        await animateFill(from, target, STEP_MS);
        if (cancelled) return;

        setActiveIdx(idx);
        await wait(STEP_MS * 0.7);
      }

      if (cancelled) return;

      const last = HIGHLIGHT_INDICES[HIGHLIGHT_INDICES.length - 1];
      await animateFill((last + 0.5) / TIMELINE.length, 1, STEP_MS * 0.6);
      if (cancelled) return;

      report(1);
      await wait(HOLD_MS);
      if (cancelled) return;

      onComplete?.();
    };

    void run();

    return () => {
      cancelled = true;
      timers.forEach((id) => window.clearTimeout(id));
      onProgress?.(0);
    };
  }, [onProgress, onComplete]);

  return (
    <div className="tuvi-forest-panel relative grid min-h-[520px] items-center gap-10 overflow-hidden rounded-[28px] px-6 py-16 sm:min-h-[560px] sm:rounded-[32px] sm:px-10 sm:py-20 lg:grid-cols-2 lg:gap-12 lg:px-14 lg:py-24">
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.15] mix-blend-overlay"
        style={{
          backgroundImage:
            "url(\"data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E\")",
        }}
        aria-hidden="true"
      />

      <div className="relative z-10">
        <p className="text-[14px] font-medium text-bg/70 sm:text-[15px]">Create more regulars</p>
        <h3 className="mt-3 font-display text-[clamp(1.85rem,3.2vw,2.85rem)] font-semibold leading-[1.12] tracking-[-0.03em] text-bg">
          <span className="block">Tuvi uses smart</span>
          <span className="block">follow-ups that grow</span>
          <span className="block">your repeat orders</span>
        </h3>
      </div>

      <div className="relative z-10 mx-auto w-full max-w-[380px]">
        <div className="relative">
          {/* Static track + flowing accent strip */}
          <div
            className="absolute left-1/2 top-4 bottom-4 w-[2px] -translate-x-1/2 overflow-hidden rounded-full bg-white/30"
            aria-hidden="true"
          >
            <div
              className="absolute left-0 top-0 w-full rounded-full"
              style={{
                height: `${fill * 100}%`,
                background: "linear-gradient(180deg, #2f6b54 0%, #3d8f6e 55%, #7b927f 100%)",
                boxShadow: "0 0 10px rgba(47, 107, 84, 0.45)",
                transition: "none",
              }}
            />
          </div>

          {/* Timeline fully visible from the start */}
          <ul className="relative flex flex-col items-center gap-3.5">
            {TIMELINE.map((step, index) => {
              const isActive = activeIdx === index;
              const glow = isActive
                ? "scale-[1.04] shadow-[0_0_0_3px_rgba(47,107,84,0.4),0_10px_28px_rgba(15,39,31,0.28)]"
                : "scale-100 shadow-[0_8px_24px_rgba(15,39,31,0.16)]";

              if (step.type === "dot") {
                const lit = fill >= (index + 0.5) / TIMELINE.length;
                return (
                  <li key={`dot-${index}`} className="relative z-10 flex h-3 items-center justify-center">
                    <span
                      className={`block h-2.5 w-2.5 rounded-full transition-all duration-300 ${
                        lit
                          ? "scale-110 bg-accent shadow-[0_0_8px_rgba(47,107,84,0.55)]"
                          : "bg-white/50"
                      }`}
                    />
                  </li>
                );
              }

              if (step.type === "wait") {
                return (
                  <li key={`wait-${index}`} className="relative z-10 py-0.5">
                    <p className="text-center text-[12px] font-medium text-bg/65">{step.label}</p>
                  </li>
                );
              }

              if (step.type === "note") {
                return (
                  <li key={`note-${index}`} className="relative z-10 py-0.5">
                    <p className="text-center text-[13px] font-medium text-bg/85">{step.label}</p>
                  </li>
                );
              }

              if (step.type === "chip") {
                return (
                  <li key={`chip-${index}`} className="relative z-10 flex w-full justify-center">
                    <div
                      className={`flex items-center gap-3 rounded-full bg-bg px-3 py-2 transition-all duration-300 ease-out ${glow}`}
                    >
                      <span className="relative h-9 w-9 shrink-0 overflow-hidden rounded-full">
                        <Image src="/people/ciara.jpg" alt="" fill className="object-cover" sizes="36px" />
                      </span>
                      <span className="pr-2 text-left leading-tight">
                        <span className="block text-[11px] font-medium text-secondary">{step.sublabel}</span>
                        <span className="block text-[15px] font-semibold text-ink">{step.label}</span>
                      </span>
                    </div>
                  </li>
                );
              }

              // pill
              return (
                <li key={`pill-${index}`} className="relative z-10 flex w-full justify-center">
                  <div
                    className={`inline-flex items-center gap-2 rounded-full bg-bg px-3.5 py-2 transition-all duration-300 ease-out ${glow}`}
                  >
                    <StepIcon icon={step.icon} active={isActive} />
                    <span className="whitespace-nowrap text-[13px] font-semibold text-ink">
                      {step.label}
                    </span>
                  </div>
                </li>
              );
            })}
          </ul>
        </div>
      </div>
    </div>
  );
}
