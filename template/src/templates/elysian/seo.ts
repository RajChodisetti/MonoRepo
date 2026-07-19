import type { Metadata } from "next";
import type { RestaurantContent } from "@/data/types/restaurant";

export function buildElysianMetadata(restaurant: RestaurantContent): Metadata {
  const durableHero = restaurant.heroMedia?.sourceKind === "google_places_live" ? "" : restaurant.heroPoster;
  return {
    title: `${restaurant.name} — ${restaurant.tagline}`,
    description: restaurant.subheadline,
    openGraph: {
      title: restaurant.name,
      description: restaurant.subheadline,
      images: durableHero ? [{ url: durableHero }] : undefined,
    },
  };
}

export function buildElysianJsonLd(restaurant: RestaurantContent) {
  const durableHero = restaurant.heroMedia?.sourceKind === "google_places_live" ? undefined : restaurant.heroPoster || undefined;
  return {
    "@context": "https://schema.org",
    "@type": "Restaurant",
    name: restaurant.name,
    description: restaurant.subheadline,
    image: durableHero,
    telephone: restaurant.phone,
    email: restaurant.email,
    address: {
      "@type": "PostalAddress",
      streetAddress: restaurant.address,
      addressLocality: restaurant.city,
      addressRegion: restaurant.state,
      addressCountry: restaurant.country,
    },
    servesCuisine: restaurant.cuisine,
    aggregateRating: restaurant.rating
      ? {
          "@type": "AggregateRating",
          ratingValue: restaurant.rating,
          reviewCount: restaurant.reviewsCount || restaurant.reviews.length,
        }
      : undefined,
  };
}
