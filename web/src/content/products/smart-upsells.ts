import type { ProductPageConfig } from "@/content/products/types";

export const smartUpsells: ProductPageConfig = {
  meta: {
    title: "Guided Add-ons | Tuvi",
    description:
      "Grow ticket size with guided add-ons. Suggest sides, drinks, and extras that fit the cart — without hard-selling guests.",
  },
  hero: {
    heading: "Grow every ticket with guided add-ons.",
    subheading:
      "Tuvi surfaces sensible extras at the right moment — so guests build a fuller order and you lift average check size on direct sales.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-email-hero.png",
      imageAlt: "Riley Quinn from Quillnest Kitchen",
      title: "See how guided add-ons lifted Riley's average ticket",
      attribution: "Riley Quinn — Quillnest Kitchen",
    },
  },
  featureSplit: {
    heading: "Ordering that helps guests build a better ticket",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "upsell-flow",
    features: [
      {
        icon: "bolt",
        title: "The right add-on at the right step",
        body: "Show sides, drinks, and desserts that fit what is already in the cart — helpful, not pushy.",
      },
      {
        icon: "gauge",
        title: "Tuned to how guests actually order",
        body: "Suggestions lean on real cart patterns so attachment feels natural in your menu flow.",
      },
      {
        icon: "chart",
        title: "Keeps refining with you",
        body: "As menus and seasons change, guided add-ons stay aligned so tickets keep climbing without extra staff coaching.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Lift average check size without nagging the guest.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "indigo",
        label: "More per order",
        title: "Sensible extras appear at checkout when they fit the cart.",
        visual: "upsell-checkout",
      },
      {
        layout: "half",
        theme: "white",
        label: "Operator control",
        title: "Set preferences and exclusions so suggestions match your kitchen.",
        visual: "upsell-data-avatars",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Quietly effective",
        title: "Small lifts per ticket compound across every first-party order.",
        visual: "upsell-improving-photo",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How do guided add-ons work?",
        answer:
          "When a guest adds an item, Tuvi suggests extras that commonly pair well with that cart — right in the ordering flow — so more tickets include a drink, side, or dessert without awkward hard sells.",
      },
      {
        question: "Can I control what gets suggested?",
        answer:
          "Yes. You can set preferences and exclusions while still letting the flow highlight add-ons that fit each cart, so ticket size grows on your terms.",
      },
      {
        question: "How does this increase revenue?",
        answer:
          "Higher attachment on drinks, sides, and desserts lifts average check size. Even a few extra dollars per order compounds across every commission-free order you take.",
      },
    ],
  },
};
