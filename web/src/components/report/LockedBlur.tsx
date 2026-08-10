"use client";

import type { ReactNode } from "react";

/** Soft blur gate — content stays in layout; unlocks after email verify. */
export default function LockedBlur({
  locked,
  children,
  label = "Verify email to unlock",
  className = "",
  onUnlock,
}: {
  locked: boolean;
  children: ReactNode;
  label?: string;
  className?: string;
  onUnlock?: () => void;
}) {
  if (!locked) return <>{children}</>;

  return (
    <div className={`relative overflow-hidden ${className}`}>
      <div className="pointer-events-none select-none blur-[7px] saturate-50" aria-hidden="true">
        {children}
      </div>
      <div className="absolute inset-0 flex items-center justify-center bg-gradient-to-b from-white/25 via-white/55 to-white/70 px-3">
        {onUnlock ? (
          <button
            type="button"
            onClick={onUnlock}
            className="min-h-11 cursor-pointer rounded-full bg-white/95 px-4 py-2 text-center text-[11.5px] font-semibold leading-snug text-[#111111] shadow-[0_4px_18px_rgba(0,0,0,0.08)] ring-1 ring-black/5 transition-transform hover:-translate-y-0.5 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
          >
            {label}
          </button>
        ) : (
          <p className="rounded-full bg-white/95 px-3.5 py-1.5 text-center text-[11.5px] font-semibold leading-snug text-[#111111] shadow-[0_4px_18px_rgba(0,0,0,0.08)]">
            {label}
          </p>
        )}
      </div>
    </div>
  );
}
