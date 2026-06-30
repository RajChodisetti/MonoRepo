import type { Metadata } from "next";
import type { RestaurantContent } from "@/data/types/restaurant";

export function buildAuroraMetadata(restaurant: RestaurantContent): Metadata {
  const title = `${restaurant.name} | ${restaurant.cuisine} · ${restaurant.city}`;
  const description = `Experience ${restaurant.name} — ${restaurant.cuisine.toLowerCase()} dining reimagined with premium hospitality in ${restaurant.city}.`;

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      images: restaurant.heroPoster ? [{ url: restaurant.heroPoster }] : [],
      type: "website",
    },
  };
}

export function buildAuroraJsonLd(restaurant: RestaurantContent): object {
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
    image: restaurant.heroPoster,
  };
}
