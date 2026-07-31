import type { Metadata } from "next";
import type { RestaurantContent } from "@/data/types/restaurant";

export function buildFoodieMetadata(restaurant: RestaurantContent): Metadata {
  const title = `${restaurant.name} | ${restaurant.cuisine || "Restaurant"} · Foodie`;
  const description =
    restaurant.subheadline ||
    restaurant.tagline ||
    `Visit ${restaurant.name} — fresh dining in ${restaurant.city || restaurant.locationLabel || "your city"}.`;

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      images: restaurant.heroPoster ? [{ url: restaurant.heroPoster }] : [{ url: "/foodie/hero-salad.png" }],
      type: "website",
    },
  };
}
