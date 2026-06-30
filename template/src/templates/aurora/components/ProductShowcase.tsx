"use client";

import Image from "next/image";
import GlassCard from "./ui/GlassCard";
import BlurReveal from "./ui/BlurReveal";
import type { MenuItem } from "@/data/types/menu";

export default function ProductShowcase({ dishes }: { dishes: MenuItem[] }) {
  if (!dishes.length) return null;

  return (
    <section id="showcase" className="aurora-section">
      <div className="aurora-container">
        <BlurReveal>
          <p className="text-xs uppercase tracking-[0.2em] text-cyan-400">Signature Plates</p>
          <h2 className="aurora-heading mt-3 text-4xl font-bold text-white md:text-5xl">
            Animated Product Showcase
          </h2>
        </BlurReveal>

        <div className="mt-16 grid gap-8 md:grid-cols-2 lg:grid-cols-3">
          {dishes.slice(0, 6).map((dish, i) => (
            <BlurReveal key={dish.name} delay={i * 0.1}>
              <GlassCard className="overflow-hidden">
                {dish.image && (
                  <div className="relative aspect-[4/3] overflow-hidden">
                    <Image
                      src={dish.image}
                      alt={dish.name}
                      fill
                      className="object-cover transition duration-700 hover:scale-105"
                      sizes="(max-width: 768px) 100vw, 33vw"
                    />
                    <div className="absolute inset-0 bg-gradient-to-t from-[#09090B] to-transparent" />
                  </div>
                )}
                <div className="p-6">
                  <h3 className="aurora-heading text-lg font-semibold text-white">{dish.name}</h3>
                  <p className="mt-2 line-clamp-2 text-sm text-white/55">{dish.description}</p>
                  {dish.price && (
                    <p className="mt-3 text-sm font-semibold text-purple-400">{dish.price}</p>
                  )}
                </div>
              </GlassCard>
            </BlurReveal>
          ))}
        </div>
      </div>
    </section>
  );
}
