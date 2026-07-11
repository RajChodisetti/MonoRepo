"use client";

import { siteContent } from "@/content/site";
import AnimatedCounter from "@/components/ui/AnimatedCounter";
import Reveal from "@/components/ui/Reveal";

export default function StatsStrip() {
  return (
    <section className="border-y border-border bg-white px-5 py-14 md:px-8">
      <div className="mx-auto grid max-w-6xl grid-cols-2 gap-x-6 gap-y-10 md:grid-cols-4">
        {siteContent.stats.map((stat, i) => (
          <Reveal key={stat.label} delay={i * 0.08} className="text-center">
            <p className="font-display text-4xl font-bold tracking-tight text-ink md:text-5xl">
              <AnimatedCounter value={stat.value} suffix={stat.suffix} />
            </p>
            <p className="mt-2 text-sm font-medium text-muted">{stat.label}</p>
          </Reveal>
        ))}
      </div>
    </section>
  );
}
