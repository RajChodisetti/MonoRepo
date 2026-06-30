"use client";

import Image from "next/image";
import type { RestaurantContent } from "@/data/types/restaurant";

export default function FinalCTA({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  return (
    <section className="relative flex min-h-[70vh] items-center justify-center overflow-hidden">
      {restaurant.heroPoster && (
        <Image
          src={restaurant.heroPoster}
          alt=""
          fill
          loading="lazy"
          className="object-cover"
          sizes="100vw"
        />
      )}
      <div className="absolute inset-0 bg-charcoal/75" />
      <div className="relative z-10 mx-auto max-w-3xl px-6 text-center">
        <h2 className="font-display text-4xl text-cream md:text-6xl">
          Come hungry. Leave with a story.
        </h2>
        <p className="mt-4 text-cream/70">
          Reserve your table and experience food made with fire, patience, and heart at{" "}
          {restaurant.name}.
        </p>
        <div className="mt-10 flex flex-wrap justify-center gap-4">
          <a href={restaurant.primaryCTA.href} className="btn-primary">
            {restaurant.primaryCTA.label}
          </a>
          <a href={restaurant.secondaryCTA.href} className="btn-ghost">
            {restaurant.secondaryCTA.label}
          </a>
        </div>
      </div>
    </section>
  );
}
