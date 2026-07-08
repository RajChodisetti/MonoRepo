import type { Metadata } from "next";
import type { RestaurantContent } from "@/data/types/restaurant";

export function buildElysianMetadata(restaurant: RestaurantContent): Metadata {
  return {
    title: `${restaurant.name} — ${restaurant.tagline}`,
    description: restaurant.subheadline,
    openGraph: {
      title: restaurant.name,
      description: restaurant.subheadline,
      images: restaurant.heroPoster ? [{ url: restaurant.heroPoster }] : undefined,
    },
  };
}

export function buildElysianJsonLd(restaurant: RestaurantContent) {
  return {
    "@context": "https://schema.org",
    "@type": "Restaurant",
    name: restaurant.name,
    description: restaurant.subheadline,
    image: restaurant.heroPoster,
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
