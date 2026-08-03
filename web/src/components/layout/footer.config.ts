export type FooterLink = {
  label: string;
  href: string;
};

export type FooterColumn = {
  title: string;
  links: FooterLink[];
};

/** Top product row — 4 columns (order matches Owner.com footer) */
export const footerProductColumns: FooterColumn[] = [
  {
    title: "Grow online discovery",
    links: [
      { label: "Restaurant Website", href: "/restaurant-website-ai" },
      { label: "Restaurant SEO", href: "/product/seo" },
      { label: "Online Menu", href: "/product/menu" },
      { label: "Reviews Engine", href: "/product/reviews" },
      { label: "Listings Management", href: "/product/listings" },
    ],
  },
  {
    title: "Grow repeat orders",
    links: [
      { label: "Branded Restaurant App", href: "/product/app" },
      { label: "Marketing Campaigns", href: "/product/campaigns" },
      { label: "Email & SMS Marketing", href: "/product/email" },
      { label: "Push Notifications Marketing", href: "/product/push" },
      { label: "Loyalty & Rewards", href: "/product/loyalty" },
    ],
  },
  {
    title: "Grow online sales",
    links: [
      { label: "Online Ordering", href: "/product/ordering" },
      { label: "Smart Upsells", href: "/product/upsells" },
      { label: "Delivery", href: "/product/delivery" },
      { label: "Catering", href: "/product/catering" },
      { label: "AI Phone Ordering", href: "/product/ai-phone" },
    ],
  },
  {
    title: "Run your restaurant",
    links: [
      { label: "Tuvi App", href: "/product/owner-app" },
      { label: "Reporting & Analytics", href: "/product/analytics" },
      { label: "Kitchen Tablet", href: "/product/kitchen" },
      { label: "POS Integrations", href: "/product/pos" },
    ],
  },
];

/** Bottom row — Resources, Company, Support */
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
    title: "Company",
    links: [
      { label: "About", href: "/company/story" },
      { label: "Careers", href: "/company/careers" },
      { label: "Leadership", href: "/company/leadership" },
      { label: "Builders Wanted", href: "/company/builders" },
      { label: "Press", href: "/company/press" },
      { label: "Partner with Tuvi", href: "/company/partners" },
    ],
  },
  {
    title: "Support",
    links: [
      { label: "1-844-24-TUVI", href: "tel:+1844248884" },
      { label: "support@tuvi.com", href: "mailto:support@tuvi.com" },
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
