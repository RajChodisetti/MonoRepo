"use client";

import { useRouter } from "next/navigation";
import {
  getOtherTemplate,
  getTemplateSwitchCopy,
  type TemplateId,
} from "@/lib/templateConfig";

type TemplateVariant = "aurora" | "cinematic";

const CURRENT_BY_VARIANT: Record<TemplateVariant, TemplateId> = {
  aurora: "2",
  cinematic: "1",
};

export default function TemplateSwitchButton({ variant }: { variant: TemplateVariant }) {
  const router = useRouter();
  const currentTemplateId = CURRENT_BY_VARIANT[variant];
  const otherTemplate = getOtherTemplate(currentTemplateId);
  const copy = getTemplateSwitchCopy(currentTemplateId);

  const switchTemplate = () => {
    const params = new URLSearchParams(window.location.search);
    params.set("template", otherTemplate);
    router.push(`/?${params.toString()}`);
  };

  const className =
    variant === "aurora"
      ? "rounded-lg border border-white/15 px-4 py-2 text-xs font-medium text-white/75 transition hover:border-cyan-400/40 hover:bg-white/5 hover:text-cyan-300"
      : "rounded border border-[#e8e0d4]/25 px-4 py-2 text-[11px] font-medium uppercase tracking-[0.1em] text-[#d4c4b0] transition hover:border-[#b88a44]/50 hover:bg-[#b88a44]/10 hover:text-[#b88a44]";

  return (
    <button type="button" onClick={switchTemplate} className={className}>
      {copy.cta}
    </button>
  );
}
