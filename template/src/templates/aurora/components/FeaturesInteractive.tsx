"use client";

import GlassCard from "./ui/GlassCard";
import BlurReveal from "./ui/BlurReveal";
import type { AuroraContent } from "../lib/mapContent";

export default function FeaturesInteractive({
  features,
}: {
  features: AuroraContent["features"];
}) {
  return (
    <section id="features" className="aurora-section">
      <div className="aurora-container">
        <BlurReveal className="text-center">
          <p className="text-xs uppercase tracking-[0.2em] text-purple-400">Experiences</p>
          <h2 className="aurora-heading mt-3 text-4xl font-bold text-white md:text-5xl">
            Interactive Features
          </h2>
          <p className="mx-auto mt-4 max-w-xl text-white/50">
            Everything you need for an unforgettable dining experience.
          </p>
        </BlurReveal>

        <div className="mt-16 grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {features.map((feature, i) => (
            <BlurReveal key={feature.title} delay={i * 0.08}>
              <GlassCard className="flex h-full flex-col p-6">
                <div className="mb-4 h-10 w-10 rounded-lg bg-gradient-to-br from-purple-500/30 to-cyan-500/30" />
                <h3 className="aurora-heading text-xl font-semibold text-white">
                  {feature.title}
                </h3>
                <p className="mt-2 flex-1 text-sm text-white/55">{feature.description}</p>
                <a
                  href={feature.href}
                  className="mt-4 text-xs font-semibold uppercase tracking-wider text-cyan-400 hover:text-cyan-300"
                >
                  {feature.label} →
                </a>
              </GlassCard>
            </BlurReveal>
          ))}
        </div>
      </div>
    </section>
  );
}
