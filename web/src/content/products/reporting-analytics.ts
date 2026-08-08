import type { ProductPageConfig } from "@/content/products/types";

export const reportingAnalytics: ProductPageConfig = {
  meta: {
    title: "Reporting & Analytics | Tuvi",
    description:
      "Restaurant reporting that makes sense. Sales, guests, and channel mix in clear dashboards — built for operators.",
  },
  hero: {
    heading: "Reports you can actually use.",
    subheading:
      "Understand sales, channels, and guest behaviour without exporting five CSVs every Monday.",
    primaryCta: { label: "Get a free demo", href: "/book" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-marketing-hero.png",
      imageAlt: "Marco Reyes from Copperwick Table",
      title: "See how clearer reports changed Marco's weekly decisions",
      attribution: "Marco Reyes — Copperwick Table",
    },
  },
  featureSplit: {
    heading: "From gut feel to confident decisions",
    headingTone: "dark",
    visualPanel: "green",
    visual: "analytics-chart",
    features: [
      {
        icon: "chart",
        title: "Sales that tell a story",
        body: "Daypart, channel, and item trends that show what is working — and what needs a nudge.",
      },
      {
        icon: "users",
        title: "Guest insights you own",
        body: "New vs returning, loyalty engagement, and campaign lift from your first-party list.",
      },
      {
        icon: "gauge",
        title: "Always current",
        body: "Dashboards update with live order data so you are never managing from last week.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Analytics built for restaurant operators.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "sky",
        label: "Performance",
        title: "Spot slow nights early and double down on winners.",
        visual: "owner-dashboard",
      },
      {
        layout: "half",
        theme: "white",
        label: "Channels",
        title: "Compare website, app, and delivery contribution clearly.",
        visual: "campaign-calendar",
      },
      {
        layout: "half",
        theme: "light",
        label: "Simple",
        title: "No analyst required — just answers you can act on.",
        visual: "loyalty-card",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Can I export reports?",
        answer:
          "Yes. Pull the views you need for accounting or partners while still using live dashboards day to day.",
      },
      {
        question: "Does reporting include loyalty and marketing?",
        answer:
          "It connects sales with guest and campaign activity so you see the full picture of what drives repeats under your brand.",
      },
      {
        question: "How often does data refresh?",
        answer:
          "Core operational metrics refresh continuously as first-party orders come in through Tuvi.",
      },
    ],
  },
};
