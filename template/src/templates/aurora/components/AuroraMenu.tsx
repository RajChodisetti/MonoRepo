"use client";

import { useMemo, useState } from "react";
import GlassCard from "./ui/GlassCard";
import BlurReveal from "./ui/BlurReveal";
import type { RestaurantContent } from "@/data/types/restaurant";

export default function AuroraMenu({ restaurant }: { restaurant: RestaurantContent }) {
  const categories = restaurant.menuCategories;
  const [active, setActive] = useState("all");

  const items = useMemo(() => {
    if (active === "all") return categories.flatMap((c) => c.items).slice(0, 24);
    return categories.find((c) => c.name === active)?.items.slice(0, 24) || [];
  }, [active, categories]);

  return (
    <section id="menu" className="aurora-section">
      <div className="aurora-container">
        <BlurReveal>
          <p className="text-xs uppercase tracking-[0.2em] text-cyan-400">Culinary</p>
          <h2 className="aurora-heading mt-3 text-4xl font-bold text-white">Menu Preview</h2>
        </BlurReveal>

        {categories.length > 1 && (
          <div className="mt-8 flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => setActive("all")}
              className={`rounded-full border px-4 py-2 text-xs uppercase tracking-wider ${
                active === "all"
                  ? "border-purple-500/50 bg-purple-500/10 text-purple-300"
                  : "border-white/10 text-white/50"
              }`}
            >
              All
            </button>
            {categories.slice(0, 8).map((cat) => (
              <button
                key={cat.name}
                type="button"
                onClick={() => setActive(cat.name)}
                className={`rounded-full border px-4 py-2 text-xs uppercase tracking-wider ${
                  active === cat.name
                    ? "border-purple-500/50 bg-purple-500/10 text-purple-300"
                    : "border-white/10 text-white/50"
                }`}
              >
                {cat.name}
              </button>
            ))}
          </div>
        )}

        <div className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((item, i) => (
            <BlurReveal key={item.name + item.category} delay={(i % 6) * 0.05}>
              <GlassCard hover={false} className="p-5">
                <p className="text-[10px] uppercase tracking-wider text-purple-400">{item.category}</p>
                <h3 className="aurora-heading mt-1 font-semibold text-white">{item.name}</h3>
                {item.price && <p className="mt-2 text-sm text-cyan-400">{item.price}</p>}
              </GlassCard>
            </BlurReveal>
          ))}
        </div>
      </div>
    </section>
  );
}
