"use client";

import { useMemo, useState } from "react";
import type { RestaurantContent } from "@/data/types/restaurant";
import MenuCategoryTabs from "./MenuCategoryTabs";

export default function MenuPreview({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const categories = restaurant.menuCategories;
  const [active, setActive] = useState("all");

  const items = useMemo(() => {
    if (active === "all") return categories.flatMap((c) => c.items).slice(0, 48);
    return categories.find((c) => c.name === active)?.items.slice(0, 48) || [];
  }, [active, categories]);

  const totalCount = categories.reduce((n, c) => n + c.items.length, 0);

  return (
    <section id="menu" className="bg-[#141210] py-24">
      <div className="mx-auto max-w-4xl px-6">
        <p className="text-xs uppercase tracking-[0.2em] text-brass">Culinary</p>
        <h2 className="font-display mt-3 text-4xl text-cream md:text-5xl">Our Menu</h2>
        <p className="mt-3 text-cream/60">
          {totalCount ? `${totalCount} dishes curated from our kitchen` : "Menu coming soon"}
        </p>

        {categories.length > 1 && (
          <MenuCategoryTabs
            categories={categories}
            active={active}
            onChange={setActive}
          />
        )}

        <div className="mt-10 divide-y divide-cream/10 rounded-2xl border border-cream/10 bg-charcoal/50">
          {items.map((item) => (
            <article
              key={item.name + item.category}
              className="flex flex-col gap-2 px-6 py-5 transition hover:bg-cream/[0.03] sm:flex-row sm:items-start sm:justify-between"
            >
              <div className="min-w-0 flex-1">
                <p className="text-[10px] uppercase tracking-wider text-brass">{item.category}</p>
                <h3 className="font-display mt-1 text-xl text-cream">
                  {item.name}
                  {item.isChefSpecial && (
                    <span className="ml-2 align-middle text-[10px] uppercase tracking-wider text-brass/80">
                      Chef&apos;s Special
                    </span>
                  )}
                </h3>
                {item.description && (
                  <p className="mt-1.5 text-sm leading-relaxed text-cream/60">{item.description}</p>
                )}
                {item.tags && item.tags.length > 0 && (
                  <div className="mt-2 flex flex-wrap gap-2">
                    {item.tags.map((tag) => (
                      <span
                        key={tag}
                        className="rounded-full border border-cream/10 px-2 py-0.5 text-[10px] uppercase tracking-wider text-cream/40"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                )}
              </div>
              {item.price && (
                <p className="shrink-0 text-base font-semibold text-brass sm:pl-6 sm:text-right">
                  {item.price}
                </p>
              )}
            </article>
          ))}
        </div>

        <div className="mt-12 flex flex-wrap gap-4">
          <a href={restaurant.primaryCTA.href} className="btn-primary">
            {restaurant.primaryCTA.label}
          </a>
          {restaurant.website && (
            <a href={restaurant.website} target="_blank" rel="noopener noreferrer" className="btn-ghost">
              Order Online
            </a>
          )}
        </div>
      </div>
    </section>
  );
}
