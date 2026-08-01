export type ImageClassificationType = "food" | "ambience" | "menu_list" | "other";

export interface CTA {
  label: string;
  href: string;
}

export interface StoryStep {
  number: string;
  title: string;
  description: string;
  image: string;
  imageMedia?: import("./gallery").GalleryImage;
}

export interface VideoAssets {
  hero: { src: string; poster: string };
  kitchen: { src: string; poster: string };
  ambience: { src: string; poster: string };
}

export interface ExperienceCard {
  id: string;
  title: string;
  description: string;
  image: string;
  imageMedia?: import("./gallery").GalleryImage;
  cta: CTA;
}

export interface RestaurantContent {
  index: number;
  restaurantId: string;
  name: string;
  tagline: string;
  subheadline: string;
  cuisine: string;
  locationLabel: string;
  metadataLabels: string[];
  rating?: number;
  reviewsCount?: number;
  priceLevel?: string;
  phone?: string;
  email?: string;
  website?: string;
  address: string;
  city: string;
  state: string;
  country: string;
  coordinates?: { latitude: number; longitude: number };
  hours: Record<string, string>;
  primaryCTA: CTA;
  secondaryCTA: CTA;
  heroPoster: string;
  heroMedia?: import("./gallery").GalleryImage;
  videos: VideoAssets;
  storySteps: StoryStep[];
  signatureDishes: import("./menu").MenuItem[];
  menuCategories: import("./menu").MenuCategory[];
  galleryImages: import("./gallery").GalleryImage[];
  reviews: import("./reviews").Review[];
  experienceCards: ExperienceCard[];
}
