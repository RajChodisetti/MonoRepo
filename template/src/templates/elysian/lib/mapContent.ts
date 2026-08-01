import type { MenuItem } from "@/data/types/menu";
import type { RestaurantContent } from "@/data/types/restaurant";
import type { GalleryImage } from "@/data/types/gallery";

export type ElysianMenuItem = {
  name: string;
  description: string;
  price: string;
  image?: string;
  categories: string[];
};

export type ElysianContent = {
  hero: {
    name: string;
    nameAccent: string;
    eyebrow: string;
    titleLine1: string;
    titleLine2: string;
    subtitle: string;
    poster: string;
    posterMedia?: RestaurantContent["heroMedia"];
    primaryCTA: { label: string; href: string };
    secondaryCTA: { label: string; href: string };
  };
  about: {
    image: string;
    imageMedia?: GalleryImage;
    badgeYears: string;
    badgeLabel: string;
    paragraphs: string[];
    timeline: { year: string; text: string }[];
    showTimeline: boolean;
  };
  dishes: MenuItem[];
  menuItems: ElysianMenuItem[];
  menuTabs: { id: string; label: string }[];
  reviews: RestaurantContent["reviews"];
  gallery: RestaurantContent["galleryImages"];
  experienceCards: RestaurantContent["experienceCards"];
  stats: { value: string; suffix?: string; label: string; animate?: boolean }[];
  faq: { question: string; answer: string }[];
  contact: {
    address: string;
    phone: string;
    email: string;
    hoursLine: string;
    mapsUrl: string;
    embedSrc?: string;
  };
  footer: {
    tagline: string;
    instaImages: RestaurantContent["galleryImages"];
  };
  show: {
    dishes: boolean;
    menu: boolean;
    testimonials: boolean;
    gallery: boolean;
    why: boolean;
    stats: boolean;
    faq: boolean;
    map: boolean;
    insta: boolean;
  };
};

function splitName(name: string): { main: string; accent: string } {
  const parts = name.trim().split(/\s+/);
  if (parts.length <= 1) return { main: name, accent: "" };
  const accent = parts.pop() || "";
  return { main: parts.join(" "), accent };
}

function inferMenuCategories(item: MenuItem, cuisine: string): string[] {
  const cats = new Set<string>(["all"]);
  if (item.category) cats.add(slugify(item.category));
  const text = `${item.name} ${item.description} ${(item.tags || []).join(" ")} ${cuisine}`.toLowerCase();
  if (/vegan|vegetarian|veg\b/.test(text)) cats.add("veg");
  else cats.add("nonveg");
  if (/dessert|soufflé|souffle|tiramisu|brûlée|brulee|chocolate|cake/.test(text)) cats.add("dessert");
  if (/wine|cocktail|martini|drink|beer|whisky|whiskey|spirit/.test(text)) cats.add("drinks");
  if (/pizza/.test(text)) cats.add("pizza");
  if (/pasta|risotto|tagliatelle|spaghetti/.test(text)) cats.add("pasta");
  if (/seafood|fish|scallop|lobster|salmon|bass|prawn|oyster/.test(text)) cats.add("seafood");
  if (/steak|wagyu|ribeye|filet|beef/.test(text)) cats.add("steak");
  return [...cats];
}

function slugify(s: string): string {
  return s.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "menu";
}

function buildMenuItems(restaurant: RestaurantContent): ElysianMenuItem[] {
  const items: ElysianMenuItem[] = [];
  for (const cat of restaurant.menuCategories) {
    for (const item of cat.items) {
      items.push({
        name: item.name,
        description: item.description || "",
        price: item.price || "",
        image: item.image,
        categories: inferMenuCategories(item, restaurant.cuisine),
      });
    }
  }
  return items;
}

function buildMenuTabs(items: ElysianMenuItem[]): { id: string; label: string }[] {
  const tabDefs = [
    { id: "all", label: "All" },
    { id: "veg", label: "Veg" },
    { id: "nonveg", label: "Non-Veg" },
    { id: "dessert", label: "Desserts" },
    { id: "drinks", label: "Drinks" },
    { id: "pizza", label: "Pizza" },
    { id: "pasta", label: "Pasta" },
    { id: "seafood", label: "Seafood" },
    { id: "steak", label: "Steak" },
  ];
  const used = new Set<string>();
  for (const item of items) {
    for (const c of item.categories) used.add(c);
  }
  const fromCategories = [...new Set(items.flatMap((i) => i.categories))]
    .filter((c) => c !== "all" && !tabDefs.some((t) => t.id === c))
    .map((c) => ({ id: c, label: c.replace(/-/g, " ").replace(/\b\w/g, (ch) => ch.toUpperCase()) }));
  return [...tabDefs.filter((t) => t.id === "all" || used.has(t.id)), ...fromCategories];
}

function buildFaq(restaurant: RestaurantContent): { question: string; answer: string }[] {
  const faq: { question: string; answer: string }[] = [];
  if (restaurant.address) {
    faq.push({ question: "Where are you located?", answer: restaurant.address });
  }
  if (restaurant.cuisine) {
    faq.push({
      question: "What cuisine do you serve?",
      answer: `${restaurant.cuisine} — ${restaurant.subheadline}`,
    });
  }
  for (const [day, hours] of Object.entries(restaurant.hours).slice(0, 4)) {
    faq.push({
      question: `What are your hours on ${day}?`,
      answer: hours,
    });
  }
  if (restaurant.phone) {
    faq.push({
      question: "How do I make a reservation?",
      answer: `Call us at ${restaurant.phone} or use the reservation form on this page.`,
    });
  }
  if (restaurant.priceLevel) {
    faq.push({
      question: "What is the price range?",
      answer: restaurant.priceLevel,
    });
  }
  return faq;
}

function buildStats(restaurant: RestaurantContent, menuCount: number): ElysianContent["stats"] {
  const stats: ElysianContent["stats"] = [];
  if (restaurant.rating) {
    stats.push({ value: String(restaurant.rating), suffix: "★", label: "Guest Rating", animate: false });
  }
  if (restaurant.reviewsCount) {
    stats.push({
      value: String(restaurant.reviewsCount),
      suffix: "+",
      label: "Reviews",
      animate: true,
    });
  }
  if (menuCount > 0) {
    stats.push({ value: String(menuCount), suffix: "+", label: "Menu Items", animate: true });
  }
  if (restaurant.galleryImages.length > 0) {
    stats.push({
      value: String(restaurant.galleryImages.length),
      suffix: "+",
      label: "Gallery Photos",
      animate: true,
    });
  }
  return stats;
}

function formatHours(hours: Record<string, string>): string {
  const entries = Object.entries(hours);
  if (!entries.length) return "";
  if (entries.length === 1) return `${entries[0][0]} · ${entries[0][1]}`;
  const first = entries[0];
  const last = entries[entries.length - 1];
  return `${first[0]} – ${last[0]} · ${first[1]}`;
}

export function mapElysianContent(restaurant: RestaurantContent): ElysianContent {
  const { main, accent } = splitName(restaurant.name);
  const menuItems = buildMenuItems(restaurant);
  const menuCount = menuItems.length;
  const stats = buildStats(restaurant, menuCount);
  const faq = buildFaq(restaurant);
  const aboutMedia =
    restaurant.galleryImages.find((image) => image.type === "ambience") ||
    restaurant.heroMedia ||
    restaurant.galleryImages[0];

  const timeline = restaurant.storySteps.map((step, i) => ({
    year: step.number || String(2010 + i * 5),
    text: `${step.title} — ${step.description}`,
  }));

  const paragraphs = [
    restaurant.subheadline,
    restaurant.storySteps[0]?.description,
    restaurant.storySteps[1]?.description,
  ].filter(Boolean) as string[];

  const mapsUrl = restaurant.coordinates
    ? `https://www.google.com/maps?q=${restaurant.coordinates.latitude},${restaurant.coordinates.longitude}`
    : `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(restaurant.address)}`;

  const embedSrc = restaurant.coordinates
    ? `https://maps.google.com/maps?q=${restaurant.coordinates.latitude},${restaurant.coordinates.longitude}&z=15&output=embed`
    : undefined;

  const eyebrowParts = [
    restaurant.locationLabel,
    restaurant.rating ? `${restaurant.rating}★ Rated` : "",
    restaurant.priceLevel || "",
  ].filter(Boolean);

  return {
    hero: {
      name: main || restaurant.name,
      nameAccent: accent,
      eyebrow: eyebrowParts.join(" — ") || restaurant.cuisine,
      titleLine1: restaurant.tagline.split(/[.!?]/)[0] || restaurant.tagline,
      titleLine2: restaurant.tagline.includes(".")
        ? restaurant.tagline.split(/[.!?]/).slice(1).join(".").trim() || restaurant.name
        : restaurant.name,
      subtitle: restaurant.subheadline,
      poster: restaurant.heroPoster,
      posterMedia: restaurant.heroMedia,
      primaryCTA: restaurant.primaryCTA,
      secondaryCTA: restaurant.secondaryCTA,
    },
    about: {
      image: aboutMedia?.url || restaurant.heroPoster,
      imageMedia: aboutMedia,
      badgeYears: restaurant.reviewsCount ? String(Math.min(99, Math.floor(restaurant.reviewsCount / 100) + 1)) : "1",
      badgeLabel: "Years of<br>Culinary Mastery",
      paragraphs,
      timeline,
      showTimeline: timeline.length > 0,
    },
    dishes: restaurant.signatureDishes,
    menuItems,
    menuTabs: buildMenuTabs(menuItems),
    reviews: restaurant.reviews,
    gallery: restaurant.galleryImages,
    experienceCards: restaurant.experienceCards,
    stats,
    faq,
    contact: {
      address: restaurant.address,
      phone: restaurant.phone || "",
      email: restaurant.email || "",
      hoursLine: formatHours(restaurant.hours),
      mapsUrl,
      embedSrc,
    },
    footer: {
      tagline: restaurant.subheadline,
      instaImages: restaurant.galleryImages.slice(0, 4),
    },
    show: {
      dishes: restaurant.signatureDishes.length > 0,
      menu: menuItems.length > 0,
      testimonials: restaurant.reviews.length > 0,
      gallery: restaurant.galleryImages.length > 0,
      why: restaurant.experienceCards.length > 0,
      stats: stats.length > 0,
      faq: faq.length > 0,
      map: Boolean(restaurant.coordinates || restaurant.address),
      insta: restaurant.galleryImages.length > 0,
    },
  };
}
