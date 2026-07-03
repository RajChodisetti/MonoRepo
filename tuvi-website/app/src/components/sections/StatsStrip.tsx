"use client";

import { siteContent } from "@/content/site";
import AnimatedCounter from "@/components/ui/AnimatedCounter";
import Reveal from "@/components/ui/Reveal";

export default function StatsStrip() {
  return (
    <section className="border-y border-border bg-bg-elevated/50 px-5 py-12 md:px-8">
      <div className="mx-auto grid max-w-6xl grid-cols-2 gap-8 md:grid-cols-4">
        {siteContent.stats.map((stat, i) => (
          <Reveal key={stat.label} delay={i * 0.08} className="text-center md:text-left">
            <p className="font-display text-4xl font-bold text-gradient md:text-5xl">
              <AnimatedCounter value={stat.value} suffix={stat.suffix} />
            </p>
            <p className="mt-2 text-sm uppercase tracking-[0.12em] text-muted">{stat.label}</p>
          </Reveal>
        ))}
      </div>
    </section>
  );
}
