import type { ProductPageConfig } from "@/content/products/types";

export const smartUpsells: ProductPageConfig = {
  meta: {
    title: "Smart Upsells | Tuvi",
    description:
      "More revenue on every order with smart upsells. Tuvi uses data from thousands of restaurants to suggest perfect add-ons.",
  },
  hero: {
    heading: "More revenue on every order with smart upsells.",
    subheading:
      "Tuvi uses data from thousands of restaurants to suggest perfect upsells, maximizing your sales per ticket.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/rahul-bhatia.jpg",
      imageAlt: "Rahul Bhatia from Saffron Indian Kitchen",
      title: "See how Smart Upsells helps Rahul grow average ticket size",
      attribution: "Rahul Bhatia — Saffron Indian Kitchen",
    },
  },
  featureSplit: {
    heading: "Online ordering that upsells like your best server.",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "upsell-flow",
    features: [
      {
        icon: "bolt",
        title: "The perfect upsells every time",
        body: "We use data to show the right upsell at the right moment, so guests spend more without feeling pushed.",
      },
      {
        icon: "gauge",
        title: "Data from thousands of restaurants",
        body: "Tuvi knows what to suggest and when, based on millions of datapoints and online guest interactions.",
      },
      {
        icon: "chart",
        title: "Gets even smarter over time",
        body: "Our team runs tests across the platform, so your upsells keep improving over time.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Grow your average check size without lifting a finger.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "indigo",
        label: "More per order",
        title: "The right add-on at checkout every time.",
        visual: "upsell-checkout",
      },
      {
        layout: "half",
        theme: "white",
        label: "Based on real data",
        title: "Built on what works across thousands of restaurants.",
        visual: "upsell-data-avatars",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Always improving",
        title: "Our team keeps testing so your upsells keep getting better.",
        visual: "upsell-improving-photo",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How do Smart Upsells work?",
        answer:
          "When a guest adds an item, Tuvi suggests the add-ons most likely to convert for that order — based on restaurant data and real guest behavior — right at checkout.",
      },
      {
        question: "Can I control what gets suggested?",
        answer:
          "Yes. You can set preferences and exclusions, while still letting the system optimize which add-ons show for each cart so tickets grow without awkward hard sells.",
      },
      {
        question: "How does this actually increase revenue?",
        answer:
          "Higher attachment rates on drinks, sides, and desserts lift average check size. Even a few extra dollars per order compounds across every direct order you take.",
      },
    ],
  },
};
