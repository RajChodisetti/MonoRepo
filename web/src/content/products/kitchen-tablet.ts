import type { ProductPageConfig } from "@/content/products/types";

export const kitchenTablet: ProductPageConfig = {
  meta: {
    title: "Kitchen Tablet | Tuvi",
    description:
      "Kitchen tablet for clear tickets. Online and phone orders land where your line can cook them fast.",
  },
  hero: {
    heading: "Tickets the kitchen can cook from.",
    subheading:
      "Online, app, and phone orders show up clearly on a kitchen tablet — so nothing gets lost between the counter and the line.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-seo-hero.png",
      imageAlt: "Mira Chen from Tileoven Co.",
      title: "See how clear tickets keep Mira's line moving",
      attribution: "Mira Chen — Tileoven Co.",
    },
  },
  featureSplit: {
    heading: "Designed for the pass — not a laptop in the office",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "kitchen-ticket",
    features: [
      {
        icon: "bolt",
        title: "Orders in real time",
        body: "New tickets appear instantly with modifiers, timing, and pickup notes your cooks need.",
      },
      {
        icon: "gear",
        title: "Bump and complete",
        body: "Mark items done as you cook so the front of house always knows status.",
      },
      {
        icon: "diamond",
        title: "Fewer mistakes",
        body: "Readable tickets reduce remakes and “what was on this?” moments mid-rush.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Keep the line focused when volume spikes.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "cream-blue",
        label: "Clarity",
        title: "Every modifier and allergy note where you can see it.",
        visual: "owner-dashboard",
      },
      {
        layout: "half",
        theme: "white",
        label: "Channels",
        title: "Website, app, and AI phone orders in one queue.",
        visual: "pos-sync",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Built for service",
        title: "A screen that belongs in the kitchen, not the office.",
        visual: "kitchen-photo-fill",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Do I need special hardware?",
        answer:
          "Tuvi Kitchen works on standard tablets. We help you set up a dedicated kitchen device that stays on the pass.",
      },
      {
        question: "Can it work alongside my POS?",
        answer:
          "Yes. With supported POS integrations, tickets stay aligned so the kitchen and counter are not working from different truths.",
      },
      {
        question: "What about phone orders?",
        answer:
          "AI phone and staff-entered orders can land in the same kitchen queue as online tickets.",
      },
    ],
  },
};
