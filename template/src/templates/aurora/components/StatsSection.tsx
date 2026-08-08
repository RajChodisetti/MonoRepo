"use client";

import { useEffect, useRef, useState } from "react";
import { useInView } from "framer-motion";
import { usePrefersReducedMotion } from "@/hooks/usePrefersReducedMotion";
import GlassCard from "./ui/GlassCard";
import BlurReveal from "./ui/BlurReveal";
import type { AuroraContent } from "../lib/mapContent";

function AnimatedStat({
  value,
  suffix = "",
  label,
}: {
  value: string;
  suffix?: string;
  label: string;
}) {
  const ref = useRef(null);
  const inView = useInView(ref, { once: true });
  const reduced = usePrefersReducedMotion();
  const numeric = parseFloat(value);
  const [display, setDisplay] = useState(reduced ? value : "0");

  useEffect(() => {
    if (!inView || reduced || Number.isNaN(numeric)) {
      setDisplay(value);
      return;
    }
    const duration = 1500;
    const startTime = performance.now();
    const tick = (now: number) => {
      const p = Math.min(1, (now - startTime) / duration);
      const current = numeric * p;
      setDisplay(Number.isInteger(numeric) ? String(Math.floor(current)) : current.toFixed(1));
      if (p < 1) requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  }, [inView, reduced, numeric, value]);

  return (
    <GlassCard hover={false} className="p-8 text-center">
      <p ref={ref} className="aurora-heading text-4xl font-bold aurora-gradient-text md:text-5xl">
        {display}{suffix}
      </p>
      <p className="mt-2 text-sm uppercase tracking-wider text-white/50">{label}</p>
    </GlassCard>
  );
}

export default function StatsSection({ stats }: { stats: AuroraContent["stats"] }) {
  return (
    <section className="aurora-section">
      <div className="aurora-container">
        <BlurReveal className="text-center">
          <h2 className="aurora-heading text-3xl font-bold text-white md:text-4xl">
            By the Numbers
          </h2>
        </BlurReveal>
        <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {stats.map((stat, i) => (
            <BlurReveal key={stat.label} delay={i * 0.1}>
              <AnimatedStat {...stat} />
            </BlurReveal>
          ))}
        </div>
      </div>
    </section>
  );
}
