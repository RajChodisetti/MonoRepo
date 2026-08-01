import type { Metadata } from "next";
import type { RestaurantContent } from "@/data/types/restaurant";

export function buildAuroraMetadata(restaurant: RestaurantContent): Metadata {
  const title = `${restaurant.name} | ${restaurant.cuisine} · ${restaurant.city}`;
  const description = `Experience ${restaurant.name} — ${restaurant.cuisine.toLowerCase()} dining reimagined with premium hospitality in ${restaurant.city}.`;

  const durableHero = restaurant.heroMedia?.sourceKind === "google_places_live" ? "" : restaurant.heroPoster;
  return {
    title,
    description,
    openGraph: {
      title,
      description,
      images: durableHero ? [{ url: durableHero }] : [],
      type: "website",
    },
  };
}

export function buildAuroraJsonLd(restaurant: RestaurantContent): object {
  const durableHero = restaurant.heroMedia?.sourceKind === "google_places_live" ? undefined : restaurant.heroPoster || undefined;
  return {
    "@context": "https://schema.org",
    "@type": "Restaurant",
    name: restaurant.name,
    description: restaurant.subheadline,
    servesCuisine: restaurant.cuisine,
    telephone: restaurant.phone,
    email: restaurant.email,
    url: restaurant.website,
    address: {
      "@type": "PostalAddress",
      streetAddress: restaurant.address,
      addressLocality: restaurant.city,
      addressRegion: restaurant.state,
      addressCountry: restaurant.country,
    },
    aggregateRating: restaurant.rating
      ? {
          "@type": "AggregateRating",
          ratingValue: restaurant.rating,
          reviewCount: restaurant.reviewsCount || 0,
        }
      : undefined,
    image: durableHero,
  };
}
