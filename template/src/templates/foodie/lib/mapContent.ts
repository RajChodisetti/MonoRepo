import type { MenuItem } from "@/data/types/menu";
import type { RestaurantContent } from "@/data/types/restaurant";
import {
  foodieContent,
  type FoodieContent,
  type FoodieMenuItem,
  type FoodieReviewItem,
} from "./foodieContent";

const AVATAR_FALLBACKS = [
  "/foodie/avatar-1.png",
  "/foodie/avatar-2.png",
  "/foodie/avatar-3.png",
  "/foodie/avatar-4.png",
] as const;

function slugify(s: string): string {
  return s.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "dish";
}

function splitTitleAccent(text: string, fallbackAccent: string): { lead: string; accent: string } {
  const cleaned = text.trim();
  if (!cleaned) return { lead: foodieContent.hero.titleLead, accent: fallbackAccent };

  const words = cleaned.split(/\s+/);
  if (words.length === 1) return { lead: cleaned, accent: fallbackAccent };

  // Prefer a short trailing accent (1–2 words) for the orange oval
  const accentCount = words.length >= 4 ? 2 : 1;
  const accent = words.slice(-accentCount).join(" ");
  const lead = words.slice(0, -accentCount).join(" ");
  return { lead, accent };
}

function formatHours(hours: Record<string, string>): string {
  const entries = Object.entries(hours);
  if (!entries.length) return foodieContent.hero.hours;
  if (entries.length === 1) return `Open: ${entries[0][1]}`;
  const first = entries[0];
  const last = entries[entries.length - 1];
  return `${first[0]} – ${last[0]} · ${first[1]}`;
}

function formatHoursLine(hours: Record<string, string>): string {
  const entries = Object.entries(hours);
  if (!entries.length) return foodieContent.contact.hoursLine;
  if (entries.length === 1) return `${entries[0][0]} · ${entries[0][1]}`;
  const first = entries[0];
  const last = entries[entries.length - 1];
  return `${first[0]} – ${last[0]} · ${first[1]}`;
}

function formatReviewCount(count: number): string {
  if (count >= 1000) {
    const k = count / 1000;
    return `${k >= 10 ? Math.round(k) : k.toFixed(1).replace(/\.0$/, "")}k+`;
  }
  return count > 0 ? `${count}+` : foodieContent.reviews.reviewCount;
}

function flattenMenuItems(restaurant: RestaurantContent): MenuItem[] {
  const seen = new Set<string>();
  const out: MenuItem[] = [];

  const push = (item: MenuItem) => {
    const key = `${item.name}|${item.price || ""}`;
    if (seen.has(key)) return;
    seen.add(key);
    out.push(item);
  };

  for (const item of restaurant.signatureDishes) push(item);
  for (const cat of restaurant.menuCategories) {
    for (const item of cat.items) push(item);
  }
  return out;
}

function isUsableImageUrl(url?: string): boolean {
  if (!url) return false;
  const u = url.trim();
  if (!u || u === "null" || u === "undefined") return false;
  if (u.startsWith("/")) return true;
  try {
    const parsed = new URL(u);
    return parsed.protocol === "http:" || parsed.protocol === "https:";
  } catch {
    return false;
  }
}

function shortenText(text: string, max = 68): string {
  const t = text.replace(/\s+/g, " ").trim();
  if (!t) return "";
  if (t.length <= max) return t;
  const cut = t.slice(0, max - 1);
  return `${cut.replace(/\s+\S*$/, "").replace(/[,:;.\-–—]+$/, "")}…`;
}

function buildMenuItems(restaurant: RestaurantContent): FoodieMenuItem[] {
  const source = flattenMenuItems(restaurant).filter((item) => isUsableImageUrl(item.image));
  const ratingLabel =
    restaurant.reviewsCount && restaurant.reviewsCount > 0
      ? `(${formatReviewCount(restaurant.reviewsCount).replace(/\+$/, "")})`
      : foodieContent.menu.items[0]?.ratingLabel || "(5.6k)";

  if (!source.length) {
    // Seed cards only when API has no imaged dishes
    return foodieContent.menu.items.map((item) => ({
      ...item,
      description: shortenText(item.description),
    }));
  }

  return source.map((item, i) => ({
    id: `${slugify(item.name)}-${i}`,
    name: item.name,
    ratingLabel,
    price: item.price || "",
    description: shortenText(
      item.description || foodieContent.menu.items[0]?.description || "",
    ),
    image: item.image!,
    featured: item.isChefSpecial || i === Math.min(2, source.length - 1),
  }));
}

function buildReviews(restaurant: RestaurantContent): FoodieContent["reviews"] {
  const location =
    restaurant.locationLabel ||
    [restaurant.city, restaurant.state].filter(Boolean).join(", ") ||
    foodieContent.reviews.items[0]?.location ||
    "";

  const items: FoodieReviewItem[] =
    restaurant.reviews.length > 0
      ? restaurant.reviews.map((r, i) => ({
          id: `review-${i}`,
          name: r.reviewer || "Guest",
          location,
          avatar: AVATAR_FALLBACKS[i % AVATAR_FALLBACKS.length],
          rating: Math.max(1, Math.min(5, Math.round(r.stars || 5))),
          quote: r.review,
        }))
      : foodieContent.reviews.items;

  const count =
    restaurant.reviewsCount ??
    (restaurant.reviews.length > 0 ? restaurant.reviews.length : 0);

  return {
    ...foodieContent.reviews,
    description: restaurant.subheadline || foodieContent.reviews.description,
    chefImage: foodieContent.reviews.chefImage,
    avatars: [...AVATAR_FALLBACKS],
    reviewCount: formatReviewCount(count || items.length),
    items,
  };
}

function buildHeroTitle(restaurant: RestaurantContent): { lead: string; accent: string } {
  const tag = restaurant.tagline?.trim();
  if (tag) {
    const sentence = tag.split(/[.!?]/)[0]?.trim() || tag;
    return splitTitleAccent(sentence, foodieContent.hero.titleAccent);
  }
  return splitTitleAccent(
    `${restaurant.name} and Enjoy`,
    foodieContent.hero.titleAccent,
  );
}

export function mapFoodieContent(restaurant: RestaurantContent): FoodieContent {
  const seed = foodieContent;
  const { lead, accent } = buildHeroTitle(restaurant);
  const menuItems = buildMenuItems(restaurant);
  const firstSig = restaurant.signatureDishes[0];
  const nameParts = restaurant.name.trim().split(/\s+/);
  const contactAccent = nameParts[nameParts.length - 1] || restaurant.name;

  const eyebrowParts = [
    restaurant.cuisine ? `Welcome to` : seed.hero.eyebrow,
  ];

  return {
    brand: {
      name: restaurant.name || seed.brand.name,
      logo: seed.brand.logo,
    },
    nav: seed.nav,
    hero: {
      ...seed.hero,
      eyebrow: eyebrowParts[0] || seed.hero.eyebrow,
      titleLead: lead,
      titleAccent: accent,
      description: restaurant.subheadline || seed.hero.description,
      primaryCta: restaurant.primaryCTA?.label || seed.hero.primaryCta,
      secondaryCta: restaurant.secondaryCTA?.label || seed.hero.secondaryCta,
      hours: formatHours(restaurant.hours),
      // Keep all decorative Foodie images
      plate: seed.hero.plate,
      garnish: seed.hero.garnish,
      badge: seed.hero.badge,
    },
    dish: {
      name: firstSig?.name || seed.dish.name,
      rating: Math.max(1, Math.min(5, Math.round(restaurant.rating || seed.dish.rating))),
      description: firstSig?.description || restaurant.subheadline || seed.dish.description,
      price: firstSig?.price || seed.dish.price,
      image: seed.dish.image,
    },
    menu: {
      titleLead: seed.menu.titleLead,
      titleAccent: seed.menu.titleAccent,
      items: menuItems,
    },
    reviews: buildReviews(restaurant),
    cta: {
      ...seed.cta,
      description: restaurant.subheadline || seed.cta.description,
      primaryCta: restaurant.primaryCTA?.label || seed.cta.primaryCta,
      secondaryCta: restaurant.secondaryCTA?.label || seed.cta.secondaryCta,
      wrapImage: seed.cta.wrapImage,
      friesImage: seed.cta.friesImage,
    },
    contact: {
      ...seed.contact,
      titleAccent: contactAccent,
      address: restaurant.address || seed.contact.address,
      phone: restaurant.phone || seed.contact.phone,
      email: restaurant.email || seed.contact.email,
      hoursLine: formatHoursLine(restaurant.hours),
      coordinates: restaurant.coordinates,
    },
    footer: {
      tagline: restaurant.tagline || restaurant.subheadline || seed.footer.tagline,
      links: seed.footer.links,
    },
  };
}
