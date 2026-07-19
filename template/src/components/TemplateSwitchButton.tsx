"use client";

import { useTransition } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import {
  getNextTemplate,
  getTemplateLabel,
  getTemplateSwitchCopy,
  type TemplateId,
} from "@/lib/templateConfig";
import { TOUR_TEMPLATE_SWITCH } from "@/lib/tourTargets";

type TemplateVariant = "aurora" | "cinematic" | "elysian";
type TemplateSwitchMode = "desktop" | "mobile";

const CURRENT_BY_VARIANT: Record<TemplateVariant, TemplateId> = {
  aurora: "2",
  cinematic: "1",
  elysian: "3",
};

function ArrowIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden>
      <path d="M5 12h14M14 7l5 5-5 5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export default function TemplateSwitchButton({
  variant,
  mode = "desktop",
}: {
  variant: TemplateVariant;
  mode?: TemplateSwitchMode;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [isPending, startTransition] = useTransition();
  const currentTemplateId = CURRENT_BY_VARIANT[variant];
  const nextTemplateId = getNextTemplate(currentTemplateId);
  const copy = getTemplateSwitchCopy(currentTemplateId);
  const currentLabel = getTemplateLabel(currentTemplateId);

  const switchTemplate = () => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("template", nextTemplateId);
    const hash = window.location.hash;
    startTransition(() => {
      router.push(`${pathname}?${params.toString()}${hash}`, { scroll: false });
    });
  };

  if (mode === "mobile") {
    const mobileClass =
      variant === "aurora"
        ? "border-cyan-300/30 bg-cyan-300/[0.06] text-white shadow-[0_12px_36px_rgba(34,211,238,0.08)]"
        : variant === "elysian"
          ? "border-[rgba(212,175,55,0.35)] bg-[rgba(212,175,55,0.07)] text-[#F8F8F8] shadow-[0_12px_36px_rgba(212,175,55,0.08)]"
          : "border-[#b88a44]/40 bg-[#b88a44]/10 text-[#f7f0e6] shadow-[0_12px_36px_rgba(184,138,68,0.08)]";
    const accentClass =
      variant === "aurora"
        ? "text-cyan-300"
        : variant === "elysian"
          ? "text-[#D4AF37]"
          : "text-[#d3a45d]";

    return (
      <button
        type="button"
        onClick={switchTemplate}
        disabled={isPending}
        aria-label={`Current website design: ${currentLabel}. Preview ${copy.targetLabel}. Restaurant details stay the same.`}
        aria-busy={isPending}
        className={`group w-full rounded-2xl border p-4 text-left transition active:scale-[0.99] disabled:cursor-wait disabled:opacity-70 ${mobileClass}`}
        data-tour={TOUR_TEMPLATE_SWITCH}
      >
        <span className="flex items-center justify-between gap-3">
          <span className="text-[10px] font-bold uppercase tracking-[0.18em] opacity-65">Website design</span>
          <span className="rounded-full border border-current/20 px-2 py-1 text-[10px] font-semibold opacity-75">
            {copy.position} of {copy.total}
          </span>
        </span>
        <span className="mt-3 flex items-end justify-between gap-4">
          <span className="min-w-0">
            <span className="block text-[10px] font-semibold uppercase tracking-[0.14em] opacity-55">Current</span>
            <span className="mt-0.5 block text-lg font-semibold leading-tight">{currentLabel}</span>
          </span>
          <span className={`flex shrink-0 items-center gap-2 text-right ${accentClass}`}>
            <span>
              <span className="block text-[10px] font-semibold uppercase tracking-[0.14em] opacity-75">Preview next</span>
              <span className="mt-0.5 block text-sm font-bold">{isPending ? "Loading…" : copy.targetLabel}</span>
            </span>
            <span className="grid h-9 w-9 place-items-center rounded-full border border-current/30 transition group-hover:translate-x-0.5">
              <span className="h-4 w-4"><ArrowIcon /></span>
            </span>
          </span>
        </span>
        <span className="mt-3 block border-t border-current/10 pt-3 text-xs opacity-60">
          Same restaurant details and photos — only the visual style changes.
        </span>
      </button>
    );
  }

  const className =
    variant === "aurora"
      ? "rounded-lg border border-white/15 px-4 py-2 text-xs font-medium text-white/75 transition hover:border-cyan-400/40 hover:bg-white/5 hover:text-cyan-300"
      : variant === "elysian"
        ? "hidden rounded-lg border border-[rgba(212,175,55,0.35)] px-4 py-2 text-xs font-medium text-[#D4AF37] transition hover:border-[#D4AF37] hover:bg-[rgba(212,175,55,0.08)] min-[901px]:inline-flex"
        : "rounded border border-[#e8e0d4]/25 px-4 py-2 text-[11px] font-medium uppercase tracking-[0.1em] text-[#d4c4b0] transition hover:border-[#b88a44]/50 hover:bg-[#b88a44]/10 hover:text-[#b88a44]";

  return (
    <button
      type="button"
      onClick={switchTemplate}
      disabled={isPending}
      aria-label={`Current website design: ${currentLabel}. ${copy.cta}.`}
      aria-busy={isPending}
      className={className}
      data-tour={TOUR_TEMPLATE_SWITCH}
    >
      {isPending ? `Loading ${copy.targetLabel}…` : copy.cta}
    </button>
  );
}
