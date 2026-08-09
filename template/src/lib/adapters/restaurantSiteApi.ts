import type { GalleryImage } from "@/data/types/gallery";
import type { MenuCategory, MenuItem } from "@/data/types/menu";
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
  menu_items?: {
    name: string;
    category?: string;
    description?: string;
    price?: string;
    price_numeric?: number;
  }[];
  media?: ApiPublicMedia[];
  reviews?: {
    reviewer?: string;
    review?: string;
    stars?: number;
    date?: string;
  }[];
};

type ApiPublicMedia = {
  url: string;
  source_kind: "google_places_live" | "owner_upload" | "licensed";
  media_type: "exterior" | "interior" | "food" | "drink" | "logo" | "team" | "event" | "other";
  caption?: string;
  alt_text?: string;
  tags?: string[];
  quality_score?: number;
  hero_score?: number;
  width_px?: number;
  height_px?: number;
  orientation?: string;
  subject_position?: string;
  placement_role?: string;
  unoptimized?: boolean;
  author_attributions?: { display_name: string; uri?: string; photo_uri?: string }[];
  google_maps_uri?: string;
  flag_content_uri?: string;
};

export type SiteRestaurantList = {
  count: number;
  restaurants: { index: number; id: string; name: string; place_id?: string; city?: string }[];
};

export type SignedDemoPayload = {
  restaurant_id: string;
  restaurant_name: string;
  cuisine?: string;
  hours?: Record<string, string>;
  address?: string;
  phone?: string;
  menu_sections?: {
    name?: string;
    items?: {
      name: string;
      description?: string;
      price?: string;
    }[];
  }[];
  reservation_cta?: string;
  ai_receptionist_cta?: string;
  media?: ApiPublicMedia[];
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
  if (t === "food_photo" || t === "food" || t === "drink") return "food";
  if (t === "interior" || t === "ambience" || t === "exterior") return "ambience";
  return "other";
}

function publicMediaToGallery(data: ApiSiteContent, item: ApiPublicMedia): GalleryImage {
  return {
    url: item.url,
    alt: item.alt_text || `${data.name} ${item.media_type || "photo"}`,
    type: galleryType(item.media_type),
    mediaType: item.media_type,
    sourceKind: item.source_kind,
    caption: item.caption,
    tags: item.tags,
    qualityScore: item.quality_score,
    heroScore: item.hero_score,
    width: item.width_px,
    height: item.height_px,
    orientation: item.orientation,
    subjectPosition: item.subject_position,
    placementRole: item.placement_role,
    unoptimized: item.source_kind === "google_places_live" || item.unoptimized,
    authorAttributions: (item.author_attributions || []).map((attribution) => ({
      displayName: attribution.display_name,
      uri: attribution.uri,
      photoUri: attribution.photo_uri,
    })),
    googleMapsUri: item.google_maps_uri,
    flagContentUri: item.flag_content_uri,
  };
}

function selectHeroMedia(data: ApiSiteContent): GalleryImage | undefined {
  const media = (data.media || []).map((item) => publicMediaToGallery(data, item));
  return (
    media.find((item) => item.placementRole === "hero") ||
    [...media]
      .filter((item) => item.orientation === "landscape" && item.heroScore != null)
      .sort((left, right) => (right.heroScore || 0) - (left.heroScore || 0))[0] ||
    media.find((item) => item.mediaType === "exterior") ||
    media.find((item) => item.mediaType === "interior") ||
    media.find((item) => item.type === "food") ||
    media[0]
  );
}

function heroPoster(data: ApiSiteContent): string {
  const selected = selectHeroMedia(data);
  return selected?.url || "";
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
      category: cat,
    });
  }
  return [...groups.entries()].map(([name, items]) => ({ name, items }));
}

function buildGalleryImages(data: ApiSiteContent): GalleryImage[] {
  return (data.media || [])
    .map((item) => publicMediaToGallery(data, item))
    .slice(0, 24);
}

function buildSignatureDishes(data: ApiSiteContent): MenuItem[] {
  const foodMedia = (data.media || [])
    .filter(
      (item) =>
        item.source_kind !== "google_places_live" &&
        (item.media_type === "food" || item.media_type === "drink"),
    )
    .slice(0, 6)
    .map((item, index): MenuItem => ({
      name: `From the kitchen${index > 0 ? ` · ${index + 1}` : ""}`,
      description: item.caption || `A glimpse of what is served at ${data.name}.`,
      image: item.url,
      isChefSpecial: true,
      category: item.media_type === "drink" ? "Drinks" : "Kitchen",
    }));
  return foodMedia;
}

function buildStorySteps(data: ApiSiteContent, images: GalleryImage[], fallbackMedia?: GalleryImage): StoryStep[] {
  const cuisine = primaryCuisine(data.cuisines).toLowerCase();
  const city = data.city || "town";
  const titles = [
    "Born from tradition",
    "Cooked with fire and patience",
    "Seasonal ingredients, modern plating",
    "A room designed for conversation",
    "Hospitality that feels personal",
  ];
  const descriptions = [
    `Our kitchen at ${data.name} draws on ${cuisine} heritage and the spirit of ${city}.`,
    "Every plate is prepared with care, technique, and warmth.",
    "We source thoughtfully and plate beautifully.",
    "From intimate tables to celebratory evenings, the dining room invites you to slow down.",
    data.rating ? `Rated ${data.rating} stars — we welcome every guest like family.` : "We welcome every guest like family.",
  ];
  const media = images.length ? images : fallbackMedia ? [fallbackMedia] : [];
  return titles.map((title, i) => ({
    number: String(i + 1).padStart(2, "0"),
    title,
    description: descriptions[i],
    image: media.length ? media[i % media.length].url : "",
    imageMedia: media.length ? media[i % media.length] : undefined,
  }));
}

function buildExperienceCards(data: ApiSiteContent, media?: GalleryImage): ExperienceCard[] {
  const phone = data.phone?.replace(/\s/g, "") || "";
  const image = media?.url || "";
  const cards: ExperienceCard[] = [
    {
      id: "dine-in",
      title: "Dine In",
      description: "Reserve a table for dinner, drinks, and celebrations.",
      image,
      imageMedia: media,
      cta: { label: "Reserve a Table", href: phone ? `tel:${phone}` : "#reserve" },
    },
    {
      id: "catering",
      title: "Catering",
      description: "Office lunches, parties, weddings, and private events.",
      image,
      imageMedia: media,
      cta: { label: "Contact Us", href: "#contact" },
    },
  ];
  if (data.website) {
    cards.splice(1, 0, {
      id: "order",
      title: "Order Online",
      description: "Your favorite dishes ready for pickup or delivery.",
      image,
      imageMedia: media,
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
  const selectedHeroMedia = selectHeroMedia(data) || gallery[0];
  const visualMedia = gallery.filter((g) => g.type !== "other").slice(0, 5);
  const storyMedia = visualMedia.length ? visualMedia : gallery.slice(0, 5);
  const phone = data.phone || "";

  return {
    index: data.index,
    restaurantId: data.restaurant_id,
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
    heroMedia: selectedHeroMedia,
    videos: getVideoAssets(poster),
    storySteps: buildStorySteps(data, storyMedia, selectedHeroMedia),
    signatureDishes: buildSignatureDishes(data),
    menuCategories: buildMenuCategories(data),
    galleryImages: gallery,
    reviews: (data.reviews || []).slice(0, 6).map((rev) => ({
      reviewer: rev.reviewer || "Guest",
      review: rev.review || "",
      stars: rev.stars || 5,
      date: rev.date,
    })),
    experienceCards: buildExperienceCards(data, selectedHeroMedia || gallery[0]),
  };
}

function adaptSignedDemoPayload(payload: SignedDemoPayload, fallbackIndex: number): RestaurantContent {
  const restaurantId = payload.restaurant_id?.trim();
  if (
    !restaurantId ||
    !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(
      restaurantId,
    ) ||
    restaurantId === "00000000-0000-0000-0000-000000000000"
  ) {
    throw new Error("Signed demo payload has no valid restaurant identity.");
  }
  const menuItems = (payload.menu_sections || []).flatMap((section) =>
    (section.items || []).map((item) => ({
      ...item,
      category: section.name || "Menu",
    })),
  );
  return adaptSiteContent({
    index: fallbackIndex,
    restaurant_id: restaurantId,
    name: payload.restaurant_name,
    cuisines: payload.cuisine ? [payload.cuisine] : [],
    phone: payload.phone,
    address: payload.address,
    hours: payload.hours || {},
    menu_items: menuItems,
    media: payload.media,
  });
}

export async function fetchSignedDemo(
  slug: string,
  token: string,
  index: number,
): Promise<RestaurantContent | null> {
  const base = apiBase();
  if (!base || !slug.trim() || !token.trim()) return null;
  try {
    const query = new URLSearchParams({ token });
    const res = await fetch(
      `${base}/api/public/v1/demo/${encodeURIComponent(slug)}?${query.toString()}`,
      { cache: "no-store" },
    );
    if (!res.ok) return null;
    return adaptSignedDemoPayload((await res.json()) as SignedDemoPayload, index);
  } catch {
    return null;
  }
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
      cache: "no-store",
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

export async function fetchSiteRestaurantByID(restaurantID: string): Promise<RestaurantContent | null> {
  const base = apiBase();
  if (!base || !/^[0-9a-f-]{36}$/i.test(restaurantID)) return null;
  try {
    const res = await fetch(
      `${base}/api/public/v1/site/restaurants/by-id/${encodeURIComponent(restaurantID)}`,
      { cache: "no-store" },
    );
    if (!res.ok) return null;
    return adaptSiteContent((await res.json()) as ApiSiteContent);
  } catch {
    return null;
  }
}

export async function getRestaurantCountFromApi(): Promise<number | null> {
  const list = await fetchSiteRestaurantList();
  return list?.count ?? null;
}
