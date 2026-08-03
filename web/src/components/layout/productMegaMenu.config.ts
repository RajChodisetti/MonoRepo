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

export const productMegaSections: MegaMenuSection[] = [
  {
    title: "Grow online discovery",
    items: [
      { label: "Restaurant Website", href: "/restaurant-website-ai", icon: "globe" },
      { label: "Restaurant SEO", href: "/product/seo", icon: "search" },
      { label: "Online Menu", href: "/product/menu", icon: "list" },
      { label: "Reviews Engine", href: "/product/reviews", icon: "star" },
      { label: "Listings Management", href: "/product/listings", icon: "mapPin" },
    ],
  },
  {
    title: "Grow online sales",
    items: [
      { label: "Online Ordering", href: "/product/ordering", icon: "bag" },
      { label: "Smart Upsells", href: "/product/upsells", icon: "chartUp" },
      { label: "Delivery", href: "/product/delivery", icon: "truck" },
      { label: "Catering", href: "/product/catering", icon: "cup" },
      {
        label: "AI Phone Ordering",
        href: "/product/ai-phone",
        icon: "phone",
        badge: "Waitlist",
      },
    ],
  },
  {
    title: "Grow repeat orders",
    items: [
      { label: "Branded Restaurant App", href: "/product/app", icon: "phoneDevice" },
      { label: "Marketing Campaigns", href: "/product/campaigns", icon: "megaphone" },
      { label: "Email & Text Marketing", href: "/product/email", icon: "mail" },
      { label: "Push Notifications Marketing", href: "/product/push", icon: "bell" },
      { label: "Loyalty & Rewards", href: "/product/loyalty", icon: "medal" },
    ],
  },
  {
    title: "Run your restaurant",
    items: [
      { label: "Owner App", href: "/product/owner-app", icon: "laptop" },
      { label: "Reporting & Analytics", href: "/product/analytics", icon: "pie" },
      { label: "Kitchen Tablet", href: "/product/kitchen", icon: "tablet" },
      { label: "POS Integrations", href: "/product/pos", icon: "card" },
    ],
  },
];

/** Featured story cards — images from Unsplash */
export const productMegaFeatured: MegaMenuFeatured[] = [
  {
    href: "/stories/talkin-tacos",
    imageSrc:
      "https://images.unsplash.com/photo-1559339352-11d035aa65de?auto=format&fit=crop&w=900&q=80",
    imageAlt: "Restaurant founders smiling together",
    title: "How Mo and Omar from Talkin Tacos grew direct online sales to $120K/m",
  },
  {
    href: "/stories/hillcrust-pizza",
    imageSrc:
      "https://images.unsplash.com/photo-1577219491135-ce391730fb2c?auto=format&fit=crop&w=900&q=80",
    imageAlt: "Chef standing in a restaurant kitchen",
    title: "How HillCrust Pizza saved thousands and ranked higher on Google with Tuvi",
  },
];
