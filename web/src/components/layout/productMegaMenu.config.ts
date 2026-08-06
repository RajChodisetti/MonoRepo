export type MegaMenuItem = {
  label: string;
  href: string;
  icon: MegaMenuIconName;
  badge?: string;
};

export type MegaMenuSection = {
  title: string;
  items: MegaMenuItem[];
};

export type MegaMenuFeatured = {
  href: string;
  imageSrc: string;
  imageAlt: string;
  title: string;
};

export type MegaMenuIconName =
  | "globe"
  | "search"
  | "list"
  | "star"
  | "mapPin"
  | "bag"
  | "chartUp"
  | "truck"
  | "cup"
  | "phone"
  | "phoneDevice"
  | "megaphone"
  | "mail"
  | "bell"
  | "medal"
  | "laptop"
  | "pie"
  | "tablet"
  | "card";

/**
 * Product mega menu — grouped by job-to-be-done for restaurants.
 */
export const productMegaSections: MegaMenuSection[] = [
  {
    title: "Get found online",
    items: [
      { label: "Restaurant Website", href: "/restaurant-website-ai", icon: "globe" },
      { label: "Restaurant SEO", href: "/product/seo", icon: "search" },
      { label: "Online Menu", href: "/product/menu", icon: "list" },
      { label: "Listings Management", href: "/product/listings", icon: "mapPin" },
      { label: "Reviews Engine", href: "/product/reviews", icon: "star" },
    ],
  },
  {
    title: "Take more orders",
    items: [
      { label: "Online Ordering", href: "/product/ordering", icon: "bag" },
      { label: "AI Phone Ordering", href: "/product/ai-phone", icon: "phone" },
      { label: "Delivery", href: "/product/delivery", icon: "truck" },
      { label: "Catering", href: "/product/catering", icon: "cup" },
      { label: "Branded Restaurant App", href: "/product/app", icon: "phoneDevice" },
    ],
  },
  {
    title: "Bring guests back",
    items: [
      { label: "Marketing Campaigns", href: "/product/campaigns", icon: "megaphone" },
      { label: "Email & SMS Marketing", href: "/product/email", icon: "mail" },
      { label: "Push Notifications Marketing", href: "/product/push", icon: "bell" },
      { label: "Loyalty & Rewards", href: "/product/loyalty", icon: "medal" },
    ],
  },
  {
    title: "Run the floor",
    items: [
      { label: "Owner App", href: "/product/owner-app", icon: "laptop" },
      { label: "Kitchen Tablet", href: "/product/kitchen", icon: "tablet" },
      { label: "POS Integrations", href: "/product/pos", icon: "card" },
      { label: "Reporting & Analytics", href: "/product/analytics", icon: "pie" },
    ],
  },
];

/** Featured story cards — fictional demo venues + Tuvi-generated imagery */
export const productMegaFeatured: MegaMenuFeatured[] = [
  {
    href: "/resources/case-studies/quillnest-kitchen",
    imageSrc: "/resources/resource-marketing-hero.png",
    imageAlt: "Warm evening dining room at Quillnest Kitchen",
    title:
      "How Riley from Quillnest Kitchen cut marketplace fees and grew +$72K in online sales",
  },
  {
    href: "/resources/case-studies/brightkiln-kitchen",
    imageSrc: "/resources/resource-seo-hero.png",
    imageAlt: "Local search moment for Brightkiln Kitchen",
    title:
      "How Brightkiln Kitchen doubled Google bookings and filled weeknights with Tuvi SEO",
  },
];
