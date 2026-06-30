"use client";

"use client";

import GlassCard from "./ui/GlassCard";
import BlurReveal from "./ui/BlurReveal";
import type { RestaurantContent } from "@/data/types/restaurant";

export default function FeatureCards({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const cuisines = (restaurant.menuCategories.length
    ? restaurant.menuCategories.slice(0, 6).map((c) => ({
        title: c.name,
        desc: `${c.items.length} dishes`,
      }))
    : [{ title: restaurant.cuisine, desc: "Seasonal menu" }]);

  const cards = [
    ...cuisines,
    {
      title: "Opening Hours",
      desc: Object.values(restaurant.hours)[0] || "Open daily",
    },
    {
      title: "Location",
      desc: restaurant.city,
    },
  ].slice(0, 6);

  return (
    <section className="aurora-section">
      <div className="aurora-container">
        <BlurReveal>
          <h2 className="aurora-heading text-center text-3xl font-bold text-white md:text-4xl">
            Why {restaurant.name}
          </h2>
        </BlurReveal>
        <div className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {cards.map((card, i) => (
            <BlurReveal key={card.title + i} delay={i * 0.08}>
              <GlassCard className="p-6">
                <h3 className="aurora-heading font-semibold text-white">{card.title}</h3>
                <p className="mt-2 text-sm text-white/55">{card.desc}</p>
              </GlassCard>
            </BlurReveal>
          ))}
        </div>
      </div>
    </section>
  );
}
