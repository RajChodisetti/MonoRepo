"use client";

import Image from "next/image";

type ReportSplashProps = {
  visible: boolean;
  fading: boolean;
};

export default function ReportSplash({ visible, fading }: ReportSplashProps) {
  if (!visible) return null;

  return (
    <div
      className={`absolute inset-0 z-50 flex items-center justify-center bg-white transition-opacity duration-300 ${
        fading ? "opacity-0" : "opacity-100"
      }`}
      aria-hidden={!visible || fading}
    >
      <div className="flex flex-col items-center gap-3">
        <div className="relative h-16 w-16 overflow-hidden rounded-full bg-[#f4f4f4]">
          <Image
            src="/brand/tuvi-solutions-logo.png"
            alt="Tuvi"
            fill
            sizes="64px"
            className="scale-[1.72] object-contain"
            priority
          />
        </div>
        <p className="text-[20px] font-semibold tracking-[-0.03em] text-[#111111]">Tuvi</p>
      </div>
    </div>
  );
}
