"use client";

import GlassCard from "./ui/GlassCard";
import BlurReveal from "./ui/BlurReveal";
import MagneticButton from "./ui/MagneticButton";
import type { AuroraContent } from "../lib/mapContent";

export default function PricingCards({
  tiers,
  ctaHref,
}: {
  tiers: AuroraContent["pricingTiers"];
  ctaHref: string;
}) {
  return (
    <section className="aurora-section">
      <div className="aurora-container">
        <BlurReveal className="text-center">
          <p className="text-xs uppercase tracking-[0.2em] text-cyan-400">Menu Tiers</p>
          <h2 className="aurora-heading mt-3 text-4xl font-bold text-white">Pricing Cards</h2>
        </BlurReveal>

        <div className="mt-12 grid gap-6 md:grid-cols-3">
          {tiers.map((tier, i) => (
            <BlurReveal key={tier.name} delay={i * 0.1}>
              <GlassCard
                className={`flex h-full flex-col p-8 ${
                  i === 1 ? "border-purple-500/40 shadow-[0_0_40px_rgba(124,58,237,0.15)]" : ""
                }`}
              >
                <h3 className="aurora-heading text-xl font-bold text-white">{tier.name}</h3>
                <p className="mt-2 text-sm text-white/50">{tier.description}</p>
                <p className="aurora-gradient-text aurora-heading mt-6 text-3xl font-bold">
                  {tier.price}
                </p>
                <ul className="mt-6 flex-1 space-y-2">
                  {tier.features.map((f) => (
                    <li key={f} className="flex items-start gap-2 text-sm text-white/60">
                      <span className="text-cyan-400">✓</span> {f}
                    </li>
                  ))}
                </ul>
                <MagneticButton href={ctaHref} className="mt-8 w-full justify-center">
                  View Menu
                </MagneticButton>
              </GlassCard>
            </BlurReveal>
          ))}
        </div>
      </div>
    </section>
  );
}
