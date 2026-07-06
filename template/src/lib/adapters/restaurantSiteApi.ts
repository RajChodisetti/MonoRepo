import type { GalleryImage } from "@/data/types/gallery";
import type { MenuCategory, MenuItem } from "@/data/types/menu";
import type { MenuListImage } from "@/data/types/menuImages";
import type {
  ExperienceCard,
  RestaurantContent,
  StoryStep,
  VideoAssets,
} from "@/data/types/restaurant";

const FOOD_CUISINE_KEYWORDS = [
  "restaurant", "food", "meal", "dining", "cafe", "bistro",
  "bar", "pub", "alcohol", "beer", "wine", "coffee", "cocktail",
  "vegan", "vegetarian", "gluten", "halal", "kosher",
];

export type ApiSiteContent = {
  index: number;
  restaurant_id: string;
  place_id?: string;
  name: string;
  cuisines?: string[] | null;
  rating?: number;
  reviews_count?: number;
  price_level?: string;
  phone?: string;
  email?: string;
  website?: string;
  address?: string;
  city?: string;
  state?: string;
  country?: string;
  latitude?: number;
  longitude?: number;
  hours?: Record<string, string>;
  thumbnail?: string;
  menu_items?: {
    name: string;
    category?: string;
    description?: string;
    price?: string;
    price_numeric?: number;
    image_url?: string;
  }[];
  menu_images?: {
    url: string;
    thumbnail_url?: string;
    image_type?: string;
    confidence?: number;
    title?: string;
  }[];
  gallery_images?: {
    url: string;
    thumbnail_url?: string;
    image_type?: string;
    title?: string;
  }[];
  reviews?: {
    reviewer?: string;
    review?: string;
    stars?: number;
    date?: string;
  }[];
};

export type SiteRestaurantList = {
  count: number;
  restaurants: { index: number; id: string; name: string; place_id?: string; city?: string }[];
};

function apiBase(): string {
  return process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "";
}

function foodCuisines(list?: string[] | null): string[] {
  return (list || []).filter((c) => {
    const low = c.toLowerCase();
    return !FOOD_CUISINE_KEYWORDS.some((k) => low === k || low.includes(`${k} `));
  });
}

function primaryCuisine(cuisines?: string[] | null): string {
  const food = foodCuisines(cuisines);
  if (food.length) return food[0];
  const raw = (cuisines || []).find((c) => /restaurant|cuisine|food/i.test(c));
  return raw || "Fine Dining";
}

function normalizeCategory(cat: string): string {
  return cat.replace(/^A LA CARTE\s*-\s*/i, "").replace(/\s+/g, " ").trim();
}

function galleryType(imageType?: string): GalleryImage["type"] {
  const t = (imageType || "").toLowerCase();
  if (t === "food_photo" || t === "food") return "food";
  if (t === "interior" || t === "ambience") return "ambience";
  return "other";
}

function heroPoster(data: ApiSiteContent): string {
  const food = (data.gallery_images || []).find((g) => galleryType(g.image_type) === "food");
  if (food?.url) return food.url;
  const amb = (data.gallery_images || []).find((g) => galleryType(g.image_type) === "ambience");
  if (amb?.url) return amb.url;
  return data.thumbnail || data.menu_images?.[0]?.url || "";
}

function getVideoAssets(poster: string): VideoAssets {
  return {
    hero: { src: "/videos/hero.mp4", poster },
    kitchen: { src: "/videos/kitchen.mp4", poster },
    ambience: { src: "/videos/ambience.mp4", poster },
  };
}

function buildMenuCategories(data: ApiSiteContent): MenuCategory[] {
  const groups = new Map<string, MenuItem[]>();
  for (const item of data.menu_items || []) {
    const cat = normalizeCategory(item.category || "Menu");
    if (!groups.has(cat)) groups.set(cat, []);
    groups.get(cat)!.push({
      name: item.name,
      description: item.description || "",
      price: item.price,
      image: item.image_url || undefined,
      category: cat,
    });
  }
  return [...groups.entries()].map(([name, items]) => ({ name, items }));
}

function buildMenuListImages(data: ApiSiteContent): MenuListImage[] {
  return (data.menu_images || []).map((img, i) => ({
    url: img.url,
    thumbnail: img.thumbnail_url || img.url,
    alt: img.title ? `${data.name} menu — ${img.title}` : `${data.name} menu ${i + 1}`,
    confidence: img.confidence,
  }));
}

function buildGalleryImages(data: ApiSiteContent): GalleryImage[] {
  return (data.gallery_images || []).slice(0, 24).map((img, i) => ({
    url: img.url,
    alt: img.title ? `${data.name} — ${img.title}` : `${data.name} gallery ${i + 1}`,
    type: galleryType(img.image_type),
  }));
}

function buildSignatureDishes(data: ApiSiteContent): MenuItem[] {
  return (data.menu_items || [])
    .filter((item) => item.image_url)
    .slice(0, 6)
    .map((item) => ({
      name: item.name,
      description: item.description || `${item.name} — a guest favorite.`,
      price: item.price,
      image: item.image_url,
      isChefSpecial: true,
      category: normalizeCategory(item.category || "Menu"),
    }));
}

function buildStorySteps(data: ApiSiteContent, images: string[]): StoryStep[] {
  const cuisine = primaryCuisine(data.cuisines).toLowerCase();
  const city = data.city || "town";
  const titles = [
    "Born from tradition",
    "Cooked with fire and patience",
    "Seasonal ingredients, modern plating",
    "A room designed for conversation",
    "Hospitality that feels personal",
  ];
  const poster = heroPoster(data);
  const descriptions = [
    `Our kitchen at ${data.name} draws on ${cuisine} heritage and the spirit of ${city}.`,
    "Every plate is prepared with care, technique, and warmth.",
    "We source thoughtfully and plate beautifully.",
    "From intimate tables to celebratory evenings, the dining room invites you to slow down.",
    data.rating ? `Rated ${data.rating} stars — we welcome every guest like family.` : "We welcome every guest like family.",
  ];
  const imgs = images.length ? images : poster ? [poster] : [""];
  return titles.map((title, i) => ({
    number: String(i + 1).padStart(2, "0"),
    title,
    description: descriptions[i],
    image: imgs[i % imgs.length],
  }));
}

function buildExperienceCards(data: ApiSiteContent, poster: string): ExperienceCard[] {
  const phone = data.phone?.replace(/\s/g, "") || "";
  const cards: ExperienceCard[] = [
    {
      id: "dine-in",
      title: "Dine In",
      description: "Reserve a table for dinner, drinks, and celebrations.",
      image: poster,
      cta: { label: "Reserve a Table", href: phone ? `tel:${phone}` : "#reserve" },
    },
    {
      id: "catering",
      title: "Catering",
      description: "Office lunches, parties, weddings, and private events.",
      image: poster,
      cta: { label: "Contact Us", href: "#contact" },
    },
  ];
  if (data.website) {
    cards.splice(1, 0, {
      id: "order",
      title: "Order Online",
      description: "Your favorite dishes ready for pickup or delivery.",
      image: poster,
      cta: { label: "Order Now", href: data.website },
    });
  }
  return cards;
}

export function adaptSiteContent(data: ApiSiteContent): RestaurantContent {
  const cuisine = primaryCuisine(data.cuisines);
  const city = data.city || "";
  const state = data.state || "";
  const country = data.country || "";
  const locationLabel = [city, state, country].filter(Boolean).join(", ");
  const poster = heroPoster(data);
  const gallery = buildGalleryImages(data);
  const storyImages = gallery.filter((g) => g.type !== "other").map((g) => g.url).slice(0, 5);
  const phone = data.phone || "";

  return {
    index: data.index,
    name: data.name,
    tagline: "Fire, flavor, and a table waiting for you.",
    subheadline: `A ${cuisine.toLowerCase()} experience built around seasonal ingredients and warm hospitality in ${city || "your city"}.`,
    cuisine,
    locationLabel,
    metadataLabels: [
      locationLabel.toUpperCase(),
      "OPEN DAILY",
      `${cuisine.toUpperCase()} • DINING • COCKTAILS`,
    ].filter(Boolean),
    rating: data.rating,
    reviewsCount: data.reviews_count,
    priceLevel: data.price_level,
    phone,
    email: data.email,
    website: data.website,
    address: data.address || locationLabel,
    city,
    state,
    country,
    coordinates:
      data.latitude != null && data.longitude != null
        ? { latitude: data.latitude, longitude: data.longitude }
        : undefined,
    hours: data.hours || {},
    primaryCTA: {
      label: "Reserve a Table",
      href: phone ? `tel:${phone.replace(/\s/g, "")}` : "#reserve",
    },
    secondaryCTA: { label: "View Menu", href: "#menu" },
    heroPoster: poster,
    videos: getVideoAssets(poster),
    storySteps: buildStorySteps(data, storyImages),
    signatureDishes: buildSignatureDishes(data),
    menuCategories: buildMenuCategories(data),
    menuListImages: buildMenuListImages(data),
    galleryImages: gallery,
    reviews: (data.reviews || []).slice(0, 6).map((rev) => ({
      reviewer: rev.reviewer || "Guest",
      review: rev.review || "",
      stars: rev.stars || 5,
      date: rev.date,
    })),
    experienceCards: buildExperienceCards(data, poster),
  };
}

export async function fetchSiteRestaurantList(): Promise<SiteRestaurantList | null> {
  const base = apiBase();
  if (!base) return null;
  try {
    const res = await fetch(`${base}/api/public/v1/site/restaurants`, { next: { revalidate: 60 } });
    if (!res.ok) return null;
    return (await res.json()) as SiteRestaurantList;
  } catch {
    return null;
  }
}

export async function fetchSiteRestaurant(index: number): Promise<RestaurantContent | null> {
  const base = apiBase();
  if (!base) return null;
  try {
    const res = await fetch(`${base}/api/public/v1/site/restaurants/${index}`, {
      next: { revalidate: 60 },
    });
    if (!res.ok) return null;
    const data = (await res.json()) as ApiSiteContent;
    if (Array.isArray(data.cuisines)) {
      // already parsed
    } else if (typeof data.cuisines === "string") {
      try {
        data.cuisines = JSON.parse(data.cuisines as unknown as string);
      } catch {
        data.cuisines = [];
      }
    }
    if (typeof data.hours === "string") {
      try {
        data.hours = JSON.parse(data.hours as unknown as string);
      } catch {
        data.hours = {};
      }
    }
    return adaptSiteContent(data);
  } catch {
    return null;
  }
}

export async function getRestaurantCountFromApi(): Promise<number | null> {
  const list = await fetchSiteRestaurantList();
  return list?.count ?? null;
}
