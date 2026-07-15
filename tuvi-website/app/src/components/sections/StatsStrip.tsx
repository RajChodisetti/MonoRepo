"use client";

import Reveal from "@/components/ui/Reveal";

const commitments = [
  ["01", "Outcome before output"],
  ["02", "Working progress, early"],
  ["03", "Human-reviewed automation"],
  ["04", "One accountable team"],
] as const;

export default function StatsStrip() {
  return (
    <section aria-label="How Tuvi works" className="border-y border-border bg-bg-elevated px-5 py-10 md:px-8 md:py-12">
      <div className="mx-auto grid max-w-6xl gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {commitments.map(([number, label], index) => (
          <Reveal key={number} delay={index * 0.05}>
            <div className="flex items-center gap-4 rounded-2xl border border-border bg-surface/60 px-4 py-4">
              <span className="font-display text-2xl font-semibold text-primary">{number}</span>
              <span className="text-sm font-semibold leading-5 text-ink">{label}</span>
            </div>
          </Reveal>
        ))}
      </div>
    </section>
  );
}
