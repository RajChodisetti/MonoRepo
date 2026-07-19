import type { Metadata } from "next";
import type { RestaurantContent } from "@/data/types/restaurant";

export function buildItalianMetadata(restaurant: RestaurantContent): Metadata {
  const title = `${restaurant.name} | Italian Restaurant in ${restaurant.city}`;
  const description =
    restaurant.subheadline ||
    `Reserve a table at ${restaurant.name}, an Italian restaurant in ${restaurant.city}.`;
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

export function buildItalianJsonLd(restaurant: RestaurantContent) {
  const durableHero = restaurant.heroMedia?.sourceKind === "google_places_live" ? undefined : restaurant.heroPoster || undefined;

  return {
    "@context": "https://schema.org",
    "@type": "Restaurant",
    name: restaurant.name,
    description: restaurant.subheadline,
    image: durableHero,
    servesCuisine: restaurant.cuisine || "Italian",
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
          reviewCount: restaurant.reviewsCount || restaurant.reviews.length,
        }
      : undefined,
    openingHoursSpecification: Object.entries(restaurant.hours).map(([day, hours]) => ({
      "@type": "OpeningHoursSpecification",
      dayOfWeek: day.charAt(0).toUpperCase() + day.slice(1),
      description: hours,
    })),
  };
}
