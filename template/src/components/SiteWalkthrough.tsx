"use client";

import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useState,
  type CSSProperties,
} from "react";
import { createPortal } from "react-dom";
import type { TemplateId } from "@/lib/templateConfig";
import { getTemplateSwitchCopy } from "@/lib/templateConfig";
import {
  TOUR_STORAGE_KEY,
  TOUR_TEMPLATE_SWITCH,
  TOUR_VOICE_ASSISTANT,
} from "@/lib/tourTargets";

type Rect = { top: number; left: number; width: number; height: number };

type Step = {
  target: string;
  placement: "bottom" | "top";
  eyebrow: string;
  title: string;
  body: string;
};

function findVisibleTarget(selector: string): HTMLElement | null {
  const nodes = document.querySelectorAll<HTMLElement>(selector);
  for (const node of nodes) {
    const rect = node.getBoundingClientRect();
    if (rect.width > 0 && rect.height > 0) return node;
  }
  return nodes[0] ?? null;
}

function measureTarget(selector: string, padding = 8): Rect | null {
  const el = findVisibleTarget(selector);
  if (!el) return null;
  const rect = el.getBoundingClientRect();
  return {
    top: rect.top - padding,
    left: rect.left - padding,
    width: rect.width + padding * 2,
    height: rect.height + padding * 2,
  };
}

function computeTourLayout(step: Step, stepIndex: number): { rect: Rect | null; style: CSSProperties } {
  const rect = measureTarget(step.target, 10);

  if (!rect) {
    return {
      rect: null,
      style: {
        top: stepIndex === 0 ? 88 : undefined,
        bottom: stepIndex === 1 ? 100 : undefined,
        right: 24,
        maxWidth: 320,
      },
    };
  }

  const tooltipWidth = Math.min(320, window.innerWidth - 32);
  const gap = 14;
  let left = rect.left + rect.width / 2 - tooltipWidth / 2;
  left = Math.max(16, Math.min(left, window.innerWidth - tooltipWidth - 16));

  const style: CSSProperties =
    step.placement === "bottom"
      ? { top: rect.top + rect.height + gap, left, width: tooltipWidth }
      : { bottom: window.innerHeight - rect.top + gap, left, width: tooltipWidth };

  return { rect, style };
}

function rectsEqual(a: Rect | null, b: Rect | null): boolean {
  if (a === b) return true;
  if (!a || !b) return false;
  return a.top === b.top && a.left === b.left && a.width === b.width && a.height === b.height;
}

function stylesEqual(a: CSSProperties, b: CSSProperties): boolean {
  const keys = new Set([...Object.keys(a), ...Object.keys(b)]);
  for (const key of keys) {
    if (a[key as keyof CSSProperties] !== b[key as keyof CSSProperties]) return false;
  }
  return true;
}

function CloseIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M18 6 6 18M6 6l12 12" />
    </svg>
  );
}

export default function SiteWalkthrough({ templateId }: { templateId: TemplateId }) {
  const isAurora = templateId === "2";
  const switchCopy = getTemplateSwitchCopy(templateId);

  const steps = useMemo<Step[]>(
    () => [
      {
        target: `[data-tour="${TOUR_TEMPLATE_SWITCH}"]`,
        placement: "bottom",
        eyebrow: switchCopy.eyebrow,
        title: switchCopy.title,
        body: `Tap "${switchCopy.cta}" up here to preview the ${switchCopy.targetLabel} layout.`,
      },
      {
        target: `[data-tour="${TOUR_VOICE_ASSISTANT}"]`,
        placement: "top",
        eyebrow: "Voice AI",
        title: "Try our AI assistant",
        body: "Book a table, ask about the menu, or get help — start a voice conversation anytime.",
      },
    ],
    [
      switchCopy.cta,
      switchCopy.eyebrow,
      switchCopy.targetLabel,
      switchCopy.title,
    ],
  );

  const [mounted, setMounted] = useState(false);
  const [active, setActive] = useState(false);
  const [stepIndex, setStepIndex] = useState(0);
  const [targetRect, setTargetRect] = useState<Rect | null>(null);
  const [tooltipStyle, setTooltipStyle] = useState<CSSProperties>({});

  const dismiss = useCallback(() => {
    setActive(false);
    try {
      localStorage.setItem(TOUR_STORAGE_KEY, "1");
    } catch {
      // ignore
    }
  }, []);

  const goNext = useCallback(() => {
    setStepIndex((current) => {
      if (current >= steps.length - 1) {
        dismiss();
        return current;
      }
      return current + 1;
    });
  }, [dismiss, steps.length]);

  useEffect(() => {
    setMounted(true);
    try {
      if (localStorage.getItem(TOUR_STORAGE_KEY) === "1") return;
    } catch {
      // ignore
    }
    const timer = window.setTimeout(() => setActive(true), 600);
    return () => window.clearTimeout(timer);
  }, []);

  useLayoutEffect(() => {
    if (!active) return;

    const { rect, style } = computeTourLayout(steps[stepIndex], stepIndex);
    setTargetRect((prev) => (rectsEqual(prev, rect) ? prev : rect));
    setTooltipStyle((prev) => (stylesEqual(prev, style) ? prev : style));
  }, [active, stepIndex, steps]);

  useEffect(() => {
    if (!active) return;

    const onChange = () => {
      const { rect, style } = computeTourLayout(steps[stepIndex], stepIndex);
      setTargetRect((prev) => (rectsEqual(prev, rect) ? prev : rect));
      setTooltipStyle((prev) => (stylesEqual(prev, style) ? prev : style));
    };

    window.addEventListener("resize", onChange);
    window.addEventListener("scroll", onChange, true);
    return () => {
      window.removeEventListener("resize", onChange);
      window.removeEventListener("scroll", onChange, true);
    };
  }, [active, stepIndex, steps]);

  if (!mounted || !active) return null;

  const step = steps[stepIndex];
  const isLast = stepIndex === steps.length - 1;

  const panelClass = isAurora
    ? "border border-white/20 bg-[#09090B]/98 text-white shadow-[0_24px_80px_rgba(0,0,0,0.55)]"
    : "border border-[#e8e0d4]/25 bg-[#1a1614]/98 text-[#f7f0e6] shadow-[0_24px_80px_rgba(0,0,0,0.5)]";

  const eyebrowClass = isAurora ? "text-cyan-300" : "text-[#b88a44]";
  const bodyClass = isAurora ? "text-white/65" : "text-[#a89f96]";
  const primaryBtn = isAurora
    ? "bg-gradient-to-r from-violet-600 to-cyan-500 text-white hover:from-violet-500 hover:to-cyan-400"
    : "bg-[#b88a44] text-[#1a1614] hover:bg-[#c99a54]";
  const ghostBtn = isAurora
    ? "text-white/55 hover:text-white"
    : "text-[#a89f96] hover:text-[#f7f0e6]";

  const arrowClass =
    step.placement === "bottom"
      ? isAurora
        ? "bottom-full border-b-[#09090B]/98 border-l-transparent border-r-transparent border-t-transparent"
        : "bottom-full border-b-[#1a1614]/98 border-l-transparent border-r-transparent border-t-transparent"
      : isAurora
        ? "top-full border-t-[#09090B]/98 border-l-transparent border-r-transparent border-b-transparent"
        : "top-full border-t-[#1a1614]/98 border-l-transparent border-r-transparent border-b-transparent";

  const arrowLeft =
    targetRect && tooltipStyle.left != null && typeof tooltipStyle.width === "number"
      ? Math.min(
          Math.max(20, targetRect.left + targetRect.width / 2 - (tooltipStyle.left as number)),
          (tooltipStyle.width as number) - 20,
        )
      : "50%";

  return createPortal(
    <div className="fixed inset-0 z-[100]" role="presentation" onClick={dismiss}>
      {targetRect ? (
        <div
          className="pointer-events-none absolute animate-pulse rounded-xl ring-2 ring-white/90 transition-all duration-300"
          style={{
            top: targetRect.top,
            left: targetRect.left,
            width: targetRect.width,
            height: targetRect.height,
            boxShadow: "0 0 0 9999px rgba(0,0,0,0.62)",
          }}
        />
      ) : (
        <div className="absolute inset-0 bg-black/60" aria-hidden />
      )}

      <div
        className={`pointer-events-auto absolute z-[101] rounded-2xl p-4 ${panelClass}`}
        onClick={(e) => e.stopPropagation()}
        style={tooltipStyle}
        role="dialog"
        aria-modal="true"
        aria-label={step.title}
      >
        <span
          className={`absolute h-0 w-0 border-[9px] ${arrowClass}`}
          style={{ left: arrowLeft, transform: typeof arrowLeft === "number" ? undefined : "translateX(-50%)" }}
          aria-hidden
        />

        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <p className={`text-[10px] font-bold uppercase tracking-[0.16em] ${eyebrowClass}`}>
              {step.eyebrow}
              <span className={`ml-2 font-medium ${bodyClass}`}>
                {stepIndex + 1}/{steps.length}
              </span>
            </p>
            <h3 className="mt-1 text-base font-semibold leading-snug">{step.title}</h3>
            <p className={`mt-2 text-sm leading-relaxed ${bodyClass}`}>{step.body}</p>
          </div>
          <button
            type="button"
            onClick={dismiss}
            className={`shrink-0 rounded-lg p-1.5 transition hover:bg-white/10 ${ghostBtn}`}
            aria-label="Close walkthrough"
          >
            <CloseIcon />
          </button>
        </div>

        <div className="mt-4 flex items-center justify-between gap-3">
          <button type="button" onClick={dismiss} className={`text-xs font-medium ${ghostBtn}`}>
            Skip tour
          </button>
          <button
            type="button"
            onClick={goNext}
            className={`rounded-xl px-4 py-2 text-sm font-semibold transition ${primaryBtn}`}
          >
            {isLast ? "Got it" : "Next"}
          </button>
        </div>
      </div>
    </div>,
    document.body,
  );
}
