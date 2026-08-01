"use client";

import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import ServiceIcon, { type ServiceIconName } from "@/components/ui/ServiceIcon";

const ease = [0.22, 1, 0.36, 1] as const;

type Props = {
  label: string;
  icon: ServiceIconName;
  imageSrc: string;
  /** Shared across all chips so flips stay in sync */
  showImage: boolean;
  float?: string;
  align?: "left" | "right";
  delay?: number;
  compact?: boolean;
  className?: string;
};

/**
 * Label face keeps the glass pill.
 * Image face is free-floating (no glass box / ring) — black asset
 * backgrounds blend out on the dark page via mix-blend-mode.
 */
export default function HeroServiceChip({
  label,
  icon,
  imageSrc,
  showImage,
  float = "",
  align = "left",
  delay = 0,
  compact = false,
  className = "",
}: Props) {
  const reduceMotion = useReducedMotion();
  const imageMode = showImage && !reduceMotion;

  const iconBox = compact ? "h-7 w-7" : "h-8 w-8";
  const iconSize = compact ? "h-3.5 w-3.5" : "h-4 w-4";
  const imgSize = compact ? "h-20 w-20" : "h-28 w-28";
  const labelClass = compact
    ? "pr-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-white/85"
    : "pr-1 text-[11px] font-semibold uppercase tracking-[0.14em] text-white/85";

  return (
    <motion.div
      initial={reduceMotion ? false : { opacity: 0, y: 18, scale: 0.92 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={{ duration: 0.55, delay, ease }}
      className={`relative flex h-16 min-w-[11rem] items-center justify-center ${
        align === "right" ? "self-end" : "self-start"
      } ${compact ? "h-16 min-w-[9.5rem] shrink-0" : ""} ${
        reduceMotion ? "" : float
      } ${className}`}
      aria-label={label}
    >
      <AnimatePresence mode="wait" initial={false}>
        {imageMode ? (
          <motion.div
            key="image"
            initial={{ opacity: 0, scale: 0.7 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.7 }}
            transition={{ duration: 0.45, ease }}
            className="pointer-events-none flex items-center justify-center bg-transparent"
          >
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={imageSrc}
              alt=""
              width={112}
              height={112}
              className={`${imgSize} bg-transparent object-contain`}
              style={{
                // Hide baked-in black square on #050505 page
                mixBlendMode: "screen",
                filter: "drop-shadow(0 8px 18px rgba(0,0,0,0.45))",
              }}
              draggable={false}
            />
          </motion.div>
        ) : (
          <motion.div
            key="label"
            initial={{ opacity: 0, scale: 0.92 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.92 }}
            transition={{ duration: 0.45, ease }}
            className={`hero-chip flex items-center gap-2.5 rounded-full px-3.5 py-2.5 ${
              compact ? "px-3 py-2" : ""
            }`}
          >
            <span
              className={`flex shrink-0 items-center justify-center rounded-full bg-white/10 text-white ring-1 ring-white/20 ${iconBox}`}
            >
              <ServiceIcon name={icon} className={iconSize} />
            </span>
            <span className={labelClass}>{label}</span>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}
