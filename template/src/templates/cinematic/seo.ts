import type { Metadata } from "next";
import type { RestaurantContent } from "@/data/types/restaurant";

export function buildMetadata(restaurant: RestaurantContent): Metadata {
  const title = `${restaurant.name} | ${restaurant.cuisine} in ${restaurant.city}`;
  const description = `Reserve a table at ${restaurant.name}, a ${restaurant.cuisine.toLowerCase()} restaurant in ${restaurant.city} serving memorable dining, cocktails, and warm hospitality.`;

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

export function buildRestaurantJsonLd(restaurant: RestaurantContent): object {
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
    openingHoursSpecification: Object.entries(restaurant.hours).map(([day, hours]) => ({
      "@type": "OpeningHoursSpecification",
      dayOfWeek: day.charAt(0).toUpperCase() + day.slice(1),
      description: hours,
    })),
    image: restaurant.heroPoster,
  };
}
