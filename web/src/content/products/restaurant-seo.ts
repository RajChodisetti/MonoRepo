import type { ProductPageConfig } from "@/content/products/types";

export const restaurantSeo: ProductPageConfig = {
  meta: {
    title: "Restaurant SEO | Tuvi",
    description:
      "Show up when hungry locals search nearby. Tuvi keeps your restaurant visible on Google so first-party traffic finds you — not a marketplace.",
  },
  hero: {
    heading: "Show up when locals are ready to order.",
    subheading:
      "Tuvi keeps your restaurant discoverable on Google so nearby guests land on your site — and order direct, commission-free.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-seo-hero.png",
      imageAlt: "Mira Chen from Tileoven Co.",
      title: "See how Tuvi helped Mira climb local search results",
      attribution: "Mira Chen — Tileoven Co.",
    },
  },
  featureSplit: {
    heading: "Local search tuned for restaurants that want direct orders",
    headingTone: "dark",
    visualPanel: "green",
    visual: "seo-score",
    features: [
      {
        icon: "chart",
        title: "Pull in traffic from your catchment",
        body: "We focus on the neighbourhoods and suburbs that actually feed your kitchen — not vanity national rankings.",
      },
      {
        icon: "users",
        title: "Compete on the searches that matter",
        body: "Tuvi studies nearby rivals and shapes your presence so hungry guests choose your site first.",
      },
      {
        icon: "pencil",
        title: "Stay steady when search shifts",
        body: "When Google changes, your pages and listings get updated so visibility does not quietly slip overnight.",
      },
    ],
  },
  featureCards: {
    sectionBg: "white",
    cards: [
      {
        layout: "full",
        theme: "green",
        label: "Search-ready pages",
        title: "Restaurant pages structured so Google understands your menu and location.",
        visual: "seo-ai-search",
      },
      {
        layout: "half",
        theme: "sky",
        label: "Ongoing care",
        title: "We watch ranking moves so your local presence keeps improving.",
        visual: "google-update",
      },
      {
        layout: "half",
        theme: "light",
        label: "Operator-first",
        title: "You run service. Tuvi keeps search and listings tidy in the background.",
        visual: "experts-avatars",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How is Tuvi SEO different for restaurants?",
        answer:
          "It is built around local demand, menus, and first-party ordering — not generic agency checklists. We pair ongoing optimisation with humans who watch search changes so rankings keep moving instead of stalling after a one-off audit.",
      },
      {
        question: "How long before we see movement?",
        answer:
          "Many venues notice clearer local visibility within a few weeks, with stronger gains over the first few months. Timing depends on competition nearby and how quickly we can tighten your site and listings.",
      },
      {
        question: "Do I need to maintain SEO myself?",
        answer:
          "Very little. Your team focuses on food and service. Tuvi keeps pages, listings, and technical basics healthy so search stays a growth channel for direct orders.",
      },
    ],
  },
};
