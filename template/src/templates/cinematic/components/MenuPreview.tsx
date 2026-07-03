"use client";

import { useMemo, useState } from "react";
import Image from "next/image";
import type { RestaurantContent } from "@/data/types/restaurant";
import MenuCategoryTabs from "./MenuCategoryTabs";
import { isFoodMenuImage } from "../lib/menuImages";

export default function MenuPreview({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const categories = restaurant.menuCategories;
  const [active, setActive] = useState("all");

  const items = useMemo(() => {
    const raw =
      active === "all"
        ? categories.flatMap((c) => c.items)
        : categories.find((c) => c.name === active)?.items || [];

    return raw
      .slice(0, 48)
      .map((item) => ({
        ...item,
        image: item.image && isFoodMenuImage(item.image) ? item.image : undefined,
      }));
  }, [active, categories]);

  const totalCount = categories.reduce((n, c) => n + c.items.length, 0);

  return (
    <section id="menu" className="bg-[#141210] py-24">
      <div className="mx-auto max-w-6xl px-6">
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

        {restaurant.menuListImages.length > 0 && (
          <div className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {restaurant.menuListImages.map((img) => (
              <div
                key={img.url}
                className="relative aspect-[3/4] overflow-hidden rounded-xl border border-cream/10 bg-[#1a1614]"
              >
                <Image
                  src={img.url}
                  alt={img.alt}
                  fill
                  loading="lazy"
                  className="object-contain p-2"
                  sizes="(max-width: 768px) 100vw, 33vw"
                />
              </div>
            ))}
          </div>
        )}

        <div className={`grid gap-6 sm:grid-cols-2 lg:grid-cols-3 ${restaurant.menuListImages.length ? "mt-12" : "mt-10"}`}>
          {items.map((item) => (
            <article
              key={item.name + item.category}
              className="group overflow-hidden rounded-xl border border-cream/10 bg-charcoal transition hover:-translate-y-1 hover:border-brass/30 hover:shadow-2xl"
            >
              {item.image && (
                <div className="relative flex aspect-[4/3] items-center justify-center overflow-hidden bg-[#1a1614] p-3">
                  <Image
                    src={item.image}
                    alt={item.name}
                    fill
                    loading="lazy"
                    className="object-contain p-1 transition duration-500 group-hover:scale-[1.02]"
                    sizes="(max-width: 768px) 100vw, 33vw"
                  />
                </div>
              )}
              <div className="p-5">
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
                  <p className="mt-2 line-clamp-2 text-sm text-cream/60">{item.description}</p>
                )}
                <div className="mt-3 flex flex-wrap items-center gap-2">
                  {item.price && <span className="text-sm font-semibold text-brass">{item.price}</span>}
                  {item.tags?.map((tag) => (
                    <span key={tag} className="text-[10px] uppercase tracking-wider text-cream/40">
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
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
