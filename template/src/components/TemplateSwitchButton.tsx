"use client";

import { useRouter } from "next/navigation";
import {
  getNextTemplate,
  getTemplateSwitchCopy,
  type TemplateId,
} from "@/lib/templateConfig";
import { TOUR_TEMPLATE_SWITCH } from "@/lib/tourTargets";

type TemplateVariant = "aurora" | "cinematic" | "elysian";

const CURRENT_BY_VARIANT: Record<TemplateVariant, TemplateId> = {
  aurora: "2",
  cinematic: "1",
  elysian: "3",
};

export default function TemplateSwitchButton({ variant }: { variant: TemplateVariant }) {
  const router = useRouter();
  const currentTemplateId = CURRENT_BY_VARIANT[variant];
  const copy = getTemplateSwitchCopy(currentTemplateId);

  const switchTemplate = () => {
    const params = new URLSearchParams(window.location.search);
    params.set("template", getNextTemplate(currentTemplateId));
    router.push(`/?${params.toString()}`);
  };

  const className =
    variant === "aurora"
      ? "rounded-lg border border-white/15 px-4 py-2 text-xs font-medium text-white/75 transition hover:border-cyan-400/40 hover:bg-white/5 hover:text-cyan-300"
      : variant === "elysian"
        ? "rounded-full border border-[rgba(212,175,55,0.35)] px-4 py-2 text-xs font-medium text-[#D4AF37] transition hover:border-[#D4AF37] hover:bg-[rgba(212,175,55,0.08)]"
        : "rounded border border-[#e8e0d4]/25 px-4 py-2 text-[11px] font-medium uppercase tracking-[0.1em] text-[#d4c4b0] transition hover:border-[#b88a44]/50 hover:bg-[#b88a44]/10 hover:text-[#b88a44]";

  return (
    <button
      type="button"
      onClick={switchTemplate}
      className={className}
      data-tour={TOUR_TEMPLATE_SWITCH}
    >
      {copy.cta}
    </button>
  );
}
