export type FooterLink = {
  label: string;
  href: string;
};

export type FooterColumn = {
  title: string;
  links: FooterLink[];
};

/** Top product row — 4 columns, sorted by job-to-be-done */
export const footerProductColumns: FooterColumn[] = [
  {
    title: "Get found online",
    links: [
      { label: "Restaurant Website", href: "/restaurant-website-ai" },
      { label: "Restaurant SEO", href: "/product/seo" },
      { label: "Online Menu", href: "/product/menu" },
      { label: "Listings Management", href: "/product/listings" },
      { label: "Reviews Engine", href: "/product/reviews" },
    ],
  },
  {
    title: "Take more orders",
    links: [
      { label: "Online Ordering", href: "/product/ordering" },
      { label: "AI Phone Ordering", href: "/product/ai-phone" },
      { label: "Delivery", href: "/product/delivery" },
      { label: "Catering", href: "/product/catering" },
      { label: "Branded Restaurant App", href: "/product/app" },
    ],
  },
  {
    title: "Bring guests back",
    links: [
      { label: "Marketing Campaigns", href: "/product/campaigns" },
      { label: "Email & SMS Marketing", href: "/product/email" },
      { label: "Push Notifications Marketing", href: "/product/push" },
      { label: "Loyalty & Rewards", href: "/product/loyalty" },
    ],
  },
  {
    title: "Run the floor",
    links: [
      { label: "Owner App", href: "/product/owner-app" },
      { label: "Kitchen Tablet", href: "/product/kitchen" },
      { label: "POS Integrations", href: "/product/pos" },
      { label: "Reporting & Analytics", href: "/product/analytics" },
    ],
  },
];

/** Bottom row — Resources, Support */
export const footerSecondaryColumns: FooterColumn[] = [
  {
    title: "Resources",
    links: [
      { label: "Case Studies", href: "/resources/case-studies" },
      { label: "Restaurant Marketing Guide", href: "/resources/marketing-guide" },
      { label: "SEO for Restaurants", href: "/resources/seo-guide" },
      { label: "Restaurant Email Marketing", href: "/resources/email-marketing" },
      { label: "Restaurant Mobile App", href: "/resources/mobile-app" },
      { label: "Online Ordering Systems", href: "/resources/ordering-systems" },
      { label: "Restaurant Website Builders", href: "/resources/website-builders" },
    ],
  },
  {
    title: "Support",
    links: [
      { label: "contact@tuvisolutions.com", href: "mailto:contact@tuvisolutions.com" },
    ],
  },
];

export const footerLegalLinks: FooterLink[] = [
  { label: "Cookie Settings", href: "/legal/cookies" },
  { label: "Privacy", href: "/legal/privacy" },
  { label: "Website Terms", href: "/legal/website-terms" },
  { label: "Disclaimer", href: "/legal/disclaimer" },
  { label: "Restaurant Agreements", href: "/legal/restaurant-agreements" },
  { label: "Platform Terms", href: "/legal/platform-terms" },
  { label: "Accessibility", href: "/legal/accessibility" },
];
