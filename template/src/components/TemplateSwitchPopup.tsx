"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  getOtherTemplate,
  getTemplateLabel,
  getTemplateSwitchCopy,
  type TemplateId,
} from "@/lib/templateConfig";

const POPUP_DELAY_MS = 5000;
const DISMISS_KEY = "template-switch-popup-dismissed";

function CloseIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M18 6 6 18M6 6l12 12" />
    </svg>
  );
}

export default function TemplateSwitchPopup({
  currentTemplateId,
}: {
  currentTemplateId: TemplateId;
}) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [visible, setVisible] = useState(false);

  const otherTemplate = getOtherTemplate(currentTemplateId);
  const copy = getTemplateSwitchCopy(currentTemplateId);
  const isAurora = currentTemplateId === "2";

  useEffect(() => {
    if (sessionStorage.getItem(DISMISS_KEY) === "1") return;

    const timer = window.setTimeout(() => {
      setVisible(true);
    }, POPUP_DELAY_MS);

    return () => window.clearTimeout(timer);
  }, []);

  const dismiss = () => {
    sessionStorage.setItem(DISMISS_KEY, "1");
    setVisible(false);
  };

  const switchTemplate = () => {
    sessionStorage.setItem(DISMISS_KEY, "1");
    setVisible(false);

    const params = new URLSearchParams(searchParams.toString());
    params.set("template", otherTemplate);
    router.push(`/?${params.toString()}`);
  };

  if (!visible) return null;

  const panelClass = isAurora
    ? "border border-white/15 bg-[#09090B]/95 text-white shadow-[0_24px_80px_rgba(0,0,0,0.55)] backdrop-blur-xl"
    : "border border-[#e8e0d4]/25 bg-[#1a1614]/95 text-[#f7f0e6] shadow-[0_24px_80px_rgba(0,0,0,0.45)] backdrop-blur-xl";

  const accentClass = isAurora ? "text-cyan-300" : "text-[#b88a44]";
  const dimClass = isAurora ? "text-white/60" : "text-[#a89f96]";
  const primaryBtn = isAurora
    ? "bg-gradient-to-r from-violet-600 to-cyan-500 text-white hover:from-violet-500 hover:to-cyan-400"
    : "bg-[#b88a44] text-[#1a1614] hover:bg-[#c99a54]";
  const ghostBtn = isAurora
    ? "border border-white/15 text-white/75 hover:bg-white/5"
    : "border border-[#e8e0d4]/25 text-[#d4c4b0] hover:bg-white/5";

  return (
    <div
      className="fixed inset-0 z-[100] flex items-end justify-center p-4 sm:items-center sm:p-6"
      role="dialog"
      aria-modal="true"
      aria-label="Switch website template"
    >
      <button
        type="button"
        className="absolute inset-0 bg-black/45 backdrop-blur-[2px]"
        aria-label="Dismiss template switch popup"
        onClick={dismiss}
      />

      <div
        className={`relative w-full max-w-md rounded-2xl p-5 sm:p-6 ${panelClass}`}
      >
        <button
          type="button"
          onClick={dismiss}
          className={`absolute right-4 top-4 rounded-lg p-1.5 transition hover:bg-white/10 ${dimClass}`}
          aria-label="Close"
        >
          <CloseIcon />
        </button>

        <p className={`text-xs font-semibold uppercase tracking-[0.16em] ${accentClass}`}>
          {copy.eyebrow}
        </p>
        <h2 className="mt-2 pr-8 text-2xl font-semibold leading-tight">{copy.title}</h2>
        <p className={`mt-3 text-sm leading-relaxed ${dimClass}`}>{copy.description}</p>

        <div className="mt-5 flex flex-col gap-2 sm:flex-row">
          <button
            type="button"
            onClick={switchTemplate}
            className={`rounded-xl px-4 py-3 text-sm font-semibold transition ${primaryBtn}`}
          >
            {copy.cta}
          </button>
          <button
            type="button"
            onClick={dismiss}
            className={`rounded-xl px-4 py-3 text-sm font-medium transition ${ghostBtn}`}
          >
            Stay on {getTemplateLabel(currentTemplateId)}
          </button>
        </div>
      </div>
    </div>
  );
}
