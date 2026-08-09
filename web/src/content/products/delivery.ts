import type { ProductPageConfig } from "@/content/products/types";

export const delivery: ProductPageConfig = {
  meta: {
    title: "Delivery | Tuvi",
    description:
      "Delivery that protects your margin. Get food out on time under your brand — fair rates, no marketplace markups on the menu.",
  },
  hero: {
    heading: "Delivery that still feels like your restaurant.",
    subheading:
      "Get food to guests on time with reliable drivers and clear pricing — while you keep the brand, the guest, and more of the margin.",
    primaryCta: { label: "Get a free demo", href: "/book" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-ordering-hero.png",
      imageAlt: "Jules Maren from Harbourleaf Kitchen",
      title: "See why Jules prefers delivering through Tuvi",
      attribution: "Jules Maren — Harbourleaf Kitchen",
    },
  },
  featureSplit: {
    heading: "Better for the venue and better for the guest",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "delivery-map-phone",
    features: [
      {
        icon: "wallet",
        title: "Predictable delivery costs",
        body: "Use partner drivers at flat rates. Guests can cover delivery on smaller tickets; you can contribute on larger ones.",
      },
      {
        icon: "car",
        title: "Widen your catchment without chaos",
        body: "Reliable drivers cover your service area so more locals can order direct from you.",
      },
      {
        icon: "phone",
        title: "You stay in the conversation",
        body: "Call guests when needed. Delivery issues get handled so your team is not stuck in marketplace chat threads.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Food out the door under your brand",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "green",
        label: "Partner drivers",
        title: "Trusted third-party drivers. Flat fees. No menu markup from Tuvi.",
        visual: "delivery-tracking-card",
      },
      {
        layout: "half",
        theme: "light",
        label: "Your rules",
        title: "Use in-house drivers, partners, or both — you choose the mix.",
        visual: "delivery-control-map",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Guest care",
        title: "Stay reachable when something goes sideways on the road.",
        visual: "delivery-guest-photo",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Why would guests order from my app instead of marketplaces?",
        answer:
          "Lower fees, a smoother ordering experience, and delivery that still feels like your brand. Guests deal with you — not a marketplace — while you keep the relationship and more of the margin.",
      },
      {
        question: "Who pays for delivery — guest or venue?",
        answer:
          "You set the rules. Guests can pay delivery on smaller orders, and you can chip in on larger ones — with flat partner rates so costs stay predictable.",
      },
      {
        question: "How do partner drivers fit in?",
        answer:
          "Drivers still get paid work through their networks. Tuvi routes delivery through rated partners at a fair flat fee — without marketplace markups on your menu.",
      },
    ],
  },
};
