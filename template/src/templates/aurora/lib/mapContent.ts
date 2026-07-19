import type { RestaurantContent } from "@/data/types/restaurant";
import type { MenuCategory } from "@/data/types/menu";

export interface AuroraContent {
  hero: {
    name: string;
    tagline: string;
    subheadline: string;
    poster: string;
    posterMedia?: RestaurantContent["heroMedia"];
    supportingMedia: RestaurantContent["galleryImages"];
    rating?: number;
    reviewsCount?: number;
    priceLevel?: string;
    primaryCTA: { label: string; href: string };
    secondaryCTA: { label: string; href: string };
    marqueeItems: string[];
  };
  stats: { label: string; value: string; suffix?: string }[];
  features: { title: string; description: string; href: string; label: string }[];
  faq: { question: string; answer: string }[];
  pricingTiers: { name: string; description: string; price: string; features: string[] }[];
}

export function mapAuroraContent(restaurant: RestaurantContent): AuroraContent {
  const menuCount = restaurant.menuCategories.reduce(
    (n, c) => n + c.items.length,
    0
  );

  const marqueeItems = [
    restaurant.cuisine,
    restaurant.city,
    restaurant.priceLevel || "Premium Dining",
    `${restaurant.rating || "5"}★ Rated`,
    "Reserve Now",
    "Seasonal Menu",
  ].filter(Boolean);

  const stats = [
    { label: "Guest Rating", value: String(restaurant.rating || "4.9"), suffix: "★" },
    { label: "Reviews", value: String(restaurant.reviewsCount || restaurant.reviews.length), suffix: "+" },
    { label: "Menu Items", value: String(menuCount), suffix: "+" },
    { label: "Price Range", value: restaurant.priceLevel || "Premium", suffix: "" },
  ];

  const features = restaurant.experienceCards.map((card) => ({
    title: card.title,
    description: card.description,
    href: card.cta.href,
    label: card.cta.label,
  }));

  const faq: { question: string; answer: string }[] = [
    {
      question: "Where are you located?",
      answer: restaurant.address,
    },
    {
      question: "What cuisine do you serve?",
      answer: `${restaurant.cuisine} — ${restaurant.subheadline}`,
    },
    ...Object.entries(restaurant.hours).slice(0, 3).map(([day, hours]) => ({
      question: `What are your hours on ${day}?`,
      answer: hours,
    })),
    {
      question: "How do I make a reservation?",
      answer: restaurant.phone
        ? `Call us at ${restaurant.phone} or use the Reserve button on this page.`
        : "Use the Reserve button on this page to get in touch.",
    },
  ];

  const pricingTiers = buildPricingTiers(restaurant.menuCategories, restaurant.priceLevel);

  return {
    hero: {
      name: restaurant.name,
      tagline: restaurant.tagline,
      subheadline: restaurant.subheadline,
      poster: restaurant.heroPoster,
      posterMedia: restaurant.heroMedia,
      supportingMedia: restaurant.galleryImages
        .filter((image) => image.url !== restaurant.heroPoster && image.sourceKind !== "google_places_live")
        .slice(0, 2),
      rating: restaurant.rating,
      reviewsCount: restaurant.reviewsCount,
      priceLevel: restaurant.priceLevel,
      primaryCTA: restaurant.primaryCTA,
      secondaryCTA: restaurant.secondaryCTA,
      marqueeItems,
    },
    stats,
    features,
    faq,
    pricingTiers,
  };
}

function buildPricingTiers(
  categories: MenuCategory[],
  priceLevel?: string
): AuroraContent["pricingTiers"] {
  if (!categories.length) {
    return [
      {
        name: "Dine In",
        description: "Full restaurant experience",
        price: priceLevel || "Visit us",
        features: ["Table service", "Full menu", "Bar & cocktails"],
      },
    ];
  }

  return categories.slice(0, 3).map((cat) => {
    const prices = cat.items
      .map((i) => i.price)
      .filter(Boolean) as string[];
    const sampleItems = cat.items.slice(0, 4).map((i) => i.name);
    return {
      name: cat.name,
      description: `${cat.items.length} curated dishes`,
      price: prices[0] || priceLevel || "Seasonal",
      features: sampleItems.length ? sampleItems : ["Chef curated", "Seasonal ingredients"],
    };
  });
}
