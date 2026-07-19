import type { ImageClassificationType } from "@/data/types/restaurant";
import type { MenuCategory, MenuItem } from "@/data/types/menu";
import type { GalleryImage } from "@/data/types/gallery";
import type {
  ExperienceCard,
  RestaurantContent,
  StoryStep,
  VideoAssets,
} from "@/data/types/restaurant";

import restaurantsData from "../../../../data/restaurants_data.json";
import imageClassifications from "../../../../data/image_classifications.json";

const FOOD_CUISINE_KEYWORDS = [
  "restaurant", "food", "meal", "dining", "cafe", "bistro",
  "bar", "pub", "alcohol", "beer", "wine", "coffee", "cocktail",
  "vegan", "vegetarian", "gluten", "halal", "kosher",
];

type ScrapedImage = { url?: string; thumbnail?: string; popular?: boolean; image_type?: string; confidence?: number; title?: string };
type ScrapedMenuItem = {
  name: string;
  category?: string;
  description?: string;
  price?: string;
  images?: ScrapedImage[];
};
type ScrapedRestaurant = {
  name: string;
  cuisines?: string[];
  rating?: number;
  reviews_count?: number;
  price_level?: string;
  contact?: { phone?: string; email?: string; website?: string };
  location?: {
    address?: string;
    city?: string;
    state?: string;
    country?: string;
    coordinates?: { latitude: number; longitude: number };
  };
  menu_items?: ScrapedMenuItem[];
  reviews?: { reviewer?: string; review?: string; stars?: number; date?: string }[];
  images?: {
    thumbnail?: string;
    gallery?: ScrapedImage[];
    menu_photos?: ScrapedImage[];
  };
  google?: { place_id?: string };
  hours?: Record<string, string>;
};

function classifyImage(url: string): ImageClassificationType {
  const entry = imageClassifications.classifications[url as keyof typeof imageClassifications.classifications];
  if (!entry) return "other";
  const t = entry.type;
  if (t === "food" || t === "ambience" || t === "menu_list") return t;
  return "other";
}

function pickImageRecord(img: ScrapedImage | string | undefined): ScrapedImage | null {
  if (!img) return null;
  if (typeof img === "string") return { url: img };
  return img;
}

function pickImageUrl(img: ScrapedImage | string | undefined): string {
  const record = pickImageRecord(img);
  if (!record) return "";
  return record.url || record.thumbnail || "";
}

function isMenuImageRecord(img: ScrapedImage | string | undefined): boolean {
  const record = pickImageRecord(img);
  const imageType = (record?.image_type || record?.title || "").toLowerCase();
  return imageType === "menu_document" || imageType === "menu_list" || imageType === "menu_ocr";
}

function pickImage(images?: ScrapedImage[]): string {
  if (!images?.length) return "";
  for (const image of images) {
    if (isMenuImageRecord(image)) continue;
    const url = pickImageUrl(image);
    if (url) return url;
  }
  return "";
}

function galleryTypeFromRecord(img: ScrapedImage): GalleryImage["type"] {
  const t = (img.image_type || img.title || "").toLowerCase();
  if (t === "food_photo" || t === "food" || t === "drink") return "food";
  if (t === "interior" || t === "ambience" || t === "exterior") return "ambience";
  return "other";
}

function buildGallery(r: ScrapedRestaurant): GalleryImage[] {
  const menuPhotoUrls = new Set(
    (r.images?.menu_photos || []).map((img) => pickImageUrl(img)).filter(Boolean),
  );

  const galleryItems = (r.images?.gallery || [])
    .map((img, i) => {
      const url = pickImageUrl(img);
      if (!url || menuPhotoUrls.has(url) || isMenuImageRecord(img)) return null;
      return {
        url,
        alt: img.title ? `${r.name} — ${img.title}` : `${r.name} gallery ${i + 1}`,
        type: galleryTypeFromRecord(img),
      };
    })
    .filter((item): item is GalleryImage => Boolean(item));

  if (galleryItems.length) return galleryItems.slice(0, 24);

  // Fallback before OCR: food/ambience from static classifications only
  const urls = collectImageUrls(r).filter((url) => !menuPhotoUrls.has(url));
  return urls.slice(0, 12).map((url, i) => ({
    url,
    alt: `${r.name} gallery ${i + 1}`,
    type: classifyImage(url) === "food" ? "food" : "ambience",
  }));
}

function foodCuisines(list?: string[]): string[] {
  return (list || []).filter((c) => {
    const low = c.toLowerCase();
    return !FOOD_CUISINE_KEYWORDS.some((k) => low === k || low.includes(`${k} `));
  });
}

function primaryCuisine(r: ScrapedRestaurant): string {
  const cuisines = foodCuisines(r.cuisines);
  if (cuisines.length) return cuisines[0];
  const raw = (r.cuisines || []).find((c) => /restaurant|cuisine|food/i.test(c));
  return raw || "Fine Dining";
}

function normalizeCategory(cat: string): string {
  return cat
    .replace(/^A LA CARTE\s*-\s*/i, "")
    .replace(/\s+/g, " ")
    .trim();
}

function inferTags(r: ScrapedRestaurant, item: ScrapedMenuItem): string[] {
  const tags: string[] = [];
  const cuisines = (r.cuisines || []).map((c) => c.toLowerCase());
  if (cuisines.some((c) => c.includes("vegan"))) tags.push("Vegan");
  if (cuisines.some((c) => c.includes("vegetarian"))) tags.push("Vegetarian");
  if (item.images?.[0]?.popular) tags.push("Popular");
  return tags;
}

function collectImageUrls(r: ScrapedRestaurant): string[] {
  const urls = new Set<string>();
  const menuPhotoURLs = new Set(
    (r.images?.menu_photos || []).map((image) => pickImageUrl(image)).filter(Boolean),
  );
  const add = (url: string) => {
    if (url && !menuPhotoURLs.has(url) && classifyImage(url) !== "menu_list") {
      urls.add(url);
    }
  };
  if (r.images?.thumbnail) add(r.images.thumbnail);
  for (const g of r.images?.gallery || []) {
    if (!isMenuImageRecord(g) && g.url) add(g.url);
  }
  for (const item of r.menu_items || []) {
    const u = pickImage(item.images);
    if (u) add(u);
  }
  return [...urls];
}

function getImagesByType(r: ScrapedRestaurant, type: ImageClassificationType): string[] {
  return collectImageUrls(r).filter((url) => classifyImage(url) === type);
}

function heroPoster(r: ScrapedRestaurant): string {
  const food = getImagesByType(r, "food");
  if (food[0]) return food[0];
  const amb = getImagesByType(r, "ambience");
  if (amb[0]) return amb[0];
  return collectImageUrls(r)[0] || "";
}

function websiteMenuItemImage(r: ScrapedRestaurant, item: ScrapedMenuItem): string {
  const candidate = pickImage(item.images);
  return candidate && collectImageUrls(r).includes(candidate) ? candidate : "";
}

function getVideoAssets(poster: string): VideoAssets {
  return {
    hero: { src: "/videos/hero.mp4", poster },
    kitchen: { src: "/videos/kitchen.mp4", poster },
    ambience: { src: "/videos/ambience.mp4", poster },
  };
}

function buildStorySteps(r: ScrapedRestaurant, images: string[]): StoryStep[] {
  const cuisine = primaryCuisine(r).toLowerCase();
  const city = r.location?.city || "town";
  const titles = [
    "Born from tradition",
    "Cooked with fire and patience",
    "Seasonal ingredients, modern plating",
    "A room designed for conversation",
    "Hospitality that feels personal",
  ];
  const descriptions = [
    `Our kitchen at ${r.name} draws on ${cuisine} heritage and the spirit of ${city}.`,
    `Every plate is prepared with care, technique, and the warmth of a team that loves what they do.`,
    `We source thoughtfully and plate beautifully — food that honors tradition while feeling fresh.`,
    `From intimate tables to celebratory evenings, the dining room invites you to slow down.`,
    r.rating
      ? `Rated ${r.rating} stars — we welcome every guest like family.`
      : `We welcome every guest like family.`,
  ];
  return titles.map((title, i) => ({
    number: String(i + 1).padStart(2, "0"),
    title,
    description: descriptions[i],
    image: images[i % Math.max(images.length, 1)] || heroPoster(r),
  }));
}

function buildMenuCategories(r: ScrapedRestaurant): MenuCategory[] {
  const groups = new Map<string, MenuItem[]>();
  for (const item of r.menu_items || []) {
    const cat = normalizeCategory(item.category || "Menu");
    if (!groups.has(cat)) groups.set(cat, []);
    const photo = websiteMenuItemImage(r, item);
    groups.get(cat)!.push({
      name: item.name,
      description: item.description || "",
      price: item.price,
      image: photo || undefined,
      tags: inferTags(r, item),
      isChefSpecial: Boolean(item.images?.[0]?.popular || photo),
      category: cat,
    });
  }
  return [...groups.entries()].map(([name, items]) => ({ name, items }));
}

function buildSignatureDishes(r: ScrapedRestaurant): MenuItem[] {
  const withImages = (r.menu_items || [])
    .map((item) => ({
      name: item.name,
      description: item.description || `${item.name} — a guest favorite from our kitchen.`,
      price: item.price,
      image: websiteMenuItemImage(r, item) || undefined,
      tags: inferTags(r, item),
      isChefSpecial: true,
      category: normalizeCategory(item.category || "Menu"),
    }))
    .filter((item) => item.image);
  return withImages.slice(0, 6);
}

function buildExperienceCards(r: ScrapedRestaurant, poster: string): ExperienceCard[] {
  const phone = r.contact?.phone?.replace(/\s/g, "") || "";
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
    {
      id: "private",
      title: "Private Dining",
      description: "Host intimate events with curated menus.",
      image: poster,
      cta: { label: "Learn More", href: "#contact" },
    },
  ];
  if (r.contact?.website) {
    cards.splice(1, 0, {
      id: "order",
      title: "Order Online",
      description: "Your favorite dishes ready for pickup or delivery.",
      image: poster,
      cta: { label: "Order Now", href: r.contact.website },
    });
  }
  return cards;
}

export function getRestaurantCount(): number {
  return (restaurantsData.restaurants as ScrapedRestaurant[]).length;
}

export function adaptRestaurant(index: number): RestaurantContent {
  const restaurants = restaurantsData.restaurants as ScrapedRestaurant[];
  if (index < 0 || index >= restaurants.length) {
    throw new Error(`Restaurant index ${index} out of range (0–${restaurants.length - 1})`);
  }
  const r = restaurants[index];
  const cuisine = primaryCuisine(r);
  const city = r.location?.city || "";
  const state = r.location?.state || "";
  const country = r.location?.country || "";
  const locationLabel = [city, state, country].filter(Boolean).join(", ");
  const poster = heroPoster(r);
  const storyImages = [
    ...getImagesByType(r, "ambience"),
    ...getImagesByType(r, "food"),
  ].slice(0, 5);
  if (!storyImages.length && poster) storyImages.push(poster);

  const phone = r.contact?.phone || "";

  return {
    index,
    restaurantId: r.google?.place_id || `json-${index}`,
    name: r.name,
    tagline: "Fire, flavor, and a table waiting for you.",
    subheadline: `A ${cuisine.toLowerCase()} experience built around seasonal ingredients and warm hospitality in ${city || "your city"}.`,
    cuisine,
    locationLabel,
    metadataLabels: [
      locationLabel.toUpperCase(),
      "OPEN DAILY",
      `${cuisine.toUpperCase()} • DINING • COCKTAILS`,
    ].filter(Boolean),
    rating: r.rating,
    reviewsCount: r.reviews_count,
    priceLevel: r.price_level,
    phone,
    email: r.contact?.email,
    website: r.contact?.website,
    address: r.location?.address || locationLabel,
    city,
    state,
    country,
    coordinates: r.location?.coordinates,
    hours: r.hours || {},
    primaryCTA: {
      label: "Reserve a Table",
      href: phone ? `tel:${phone.replace(/\s/g, "")}` : "#reserve",
    },
    secondaryCTA: { label: "View Menu", href: "#menu" },
    heroPoster: poster,
    videos: getVideoAssets(poster),
    storySteps: buildStorySteps(r, storyImages),
    signatureDishes: buildSignatureDishes(r),
    menuCategories: buildMenuCategories(r),
    galleryImages: buildGallery(r),
    reviews: (r.reviews || []).slice(0, 6).map((rev) => ({
      reviewer: rev.reviewer || "Guest",
      review: rev.review || "",
      stars: rev.stars || 5,
      date: rev.date,
    })),
    experienceCards: buildExperienceCards(r, poster),
  };
}

export function loadRestaurant(index: number): RestaurantContent {
  return adaptRestaurant(index);
}

export function getPlaceID(index: number): string {
  const restaurants = restaurantsData.restaurants as ScrapedRestaurant[];
  if (index < 0 || index >= restaurants.length) return "";
  return (restaurants[index].google?.place_id || "").trim();
}

export async function loadRestaurantAsync(index: number): Promise<RestaurantContent> {
  const base = adaptRestaurant(index);
  const placeID = getPlaceID(index);
  if (!placeID) return base;

  const { fetchSiteImagesByPlaceID } = await import("./restaurantImagesApi");
  const images = await fetchSiteImagesByPlaceID(placeID, base.name);
  if (!images) return base;

  return {
    ...base,
    galleryImages: images.galleryImages.length ? images.galleryImages : base.galleryImages,
  };
}

export function parseRestaurantIndex(raw?: string | string[] | null): number {
  const value = Array.isArray(raw) ? raw[0] : raw;
  const n = parseInt(value ?? "0", 10);
  if (Number.isNaN(n) || n < 0) return 0;
  return n;
}
