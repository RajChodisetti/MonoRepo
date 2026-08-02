export type VisualId =
  | "gyro-preview"
  | "ai-search-mock"
  | "ordering-ravioli"
  | "phone-improving"
  | "seo-score"
  | "seo-ai-search"
  | "google-update"
  | "experts-avatars"
  | "menu-items-stack"
  | "rewards-phone"
  | "pita-wraps-menu"
  | "order-tracking-phone"
  | "reviews-owner-photo"
  | "google-reviews-stack"
  | "customers-flow"
  | "reviews-phone-mockup"
  | "listing-map-card"
  | "listings-synced"
  | "address-fix-cards"
  | "listings-experts-photo"
  | "ordering-phone-preview"
  | "ordering-app-showcase"
  | "ordering-customer-list"
  | "fee-savings-toast"
  | "upsell-flow"
  | "upsell-checkout"
  | "upsell-data-avatars"
  | "upsell-improving-photo"
  | "delivery-map-phone"
  | "delivery-tracking-card"
  | "delivery-control-map"
  | "delivery-guest-photo"
  | "catering-menu-stack"
  | "catering-search"
  | "catering-food-collage"
  | "catering-phone-mockup"
  | "ai-phone-mockup"
  | "ai-phone-conversation"
  | "ai-phone-loyalty-photo"
  | "ai-phone-food-tiles"
  | "branded-app-phone"
  | "branded-app-showcase"
  | "app-photo-fill"
  | "campaign-promo"
  | "campaign-calendar"
  | "email-sms-preview"
  | "push-notif-stack"
  | "loyalty-card"
  | "loyalty-rewards-grid"
  | "owner-dashboard"
  | "owner-photo-fill"
  | "analytics-chart"
  | "kitchen-ticket"
  | "kitchen-photo-fill"
  | "pos-sync";

export type IconId =
  | "badge"
  | "percent"
  | "gauge"
  | "chart"
  | "users"
  | "pencil"
  | "diamond"
  | "person"
  | "gear"
  | "bolt"
  | "wallet"
  | "car"
  | "phone"
  | "trophy"
  | "card";

export type CtaLink = {
  label: string;
  href: string;
};

export type ProductHeroConfig = {
  heading: string;
  subheading: string;
  primaryCta: CtaLink;
  secondaryCta?: CtaLink;
  testimonial: {
    imageSrc: string;
    imageAlt: string;
    title?: string;
    attribution?: string;
  };
};

export type ProductFeatureItem = {
  title: string;
  body: string;
  icon: IconId;
};

export type ProductFeatureSplitConfig = {
  heading: string;
  /** muted = gray (website), dark = near-black (SEO) */
  headingTone?: "muted" | "dark";
  /** Panel behind left visual. none = photo flush in rounded frame */
  visualPanel?: "peach" | "green" | "none";
  visual: VisualId;
  features: ProductFeatureItem[];
};

export type FeatureCardTheme =
  | "blue"
  | "beige"
  | "white-green"
  | "green"
  | "sky"
  | "light"
  | "cream-blue"
  | "dark"
  | "white"
  | "indigo";

export type ProductFeatureCard = {
  label: string;
  title: string;
  theme: FeatureCardTheme;
  layout: "full" | "half";
  /** Text color hint for label/title when theme doesn't imply it */
  textTone?: "light" | "dark";
  /** Where the visual sits on full-width cards. Default: left */
  visualSide?: "left" | "right";
  visual: VisualId;
};

export type ProductFeatureCardsConfig = {
  sectionHeading?: string;
  sectionBg?: "beige" | "white";
  cards: ProductFeatureCard[];
};

export type ProductFaqItem = {
  question: string;
  answer: string;
};

export type ProductPageConfig = {
  meta: {
    title: string;
    description: string;
  };
  hero: ProductHeroConfig;
  featureSplit: ProductFeatureSplitConfig;
  featureCards: ProductFeatureCardsConfig;
  faq: {
    items: ProductFaqItem[];
  };
};
