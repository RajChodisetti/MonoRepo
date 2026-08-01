"use client";

import type { RestaurantContent } from "@/data/types/restaurant";
import SourceAwareImage, { mediaForURL } from "@/components/SourceAwareImage";
import PhotoAttribution from "@/components/PhotoAttribution";

export default function FinalCTA({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  return (
    <section className="relative flex min-h-[70vh] items-center justify-center overflow-hidden">
      {restaurant.heroPoster && (
        <SourceAwareImage
          media={restaurant.heroMedia || mediaForURL(restaurant.heroPoster, "")}
          fill
          loading="lazy"
          className="object-cover"
          sizes="100vw"
        />
      )}
      <div className="absolute inset-0 bg-charcoal/75" />
      {restaurant.heroMedia?.sourceKind === "google_places_live" ? (
        <div className="absolute bottom-5 left-5 z-20 rounded bg-black/60 px-3 py-2 text-white/75">
          <PhotoAttribution media={restaurant.heroMedia} compact />
        </div>
      ) : null}
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
