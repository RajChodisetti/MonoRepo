import type { ProductPageConfig } from "@/content/products/types";

export const aiPhoneOrdering: ProductPageConfig = {
  meta: {
    title: "AI Phone Ordering | Tuvi",
    description:
      "Never miss another phone order. Tuvi answers the call, takes the order, and sends it to your kitchen — while growing your guest list.",
  },
  hero: {
    heading: "Answer every call you are currently missing.",
    subheading:
      "Tuvi picks up, takes the order in a natural conversation, and sends it straight to your kitchen — then keeps the caller in your world.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-help-hero.png",
      imageAlt: "Marco Reyes from Copperwick Table",
      title: "See how Marco captures phone orders without missing service",
      attribution: "Marco Reyes — Copperwick Table",
    },
  },
  featureSplit: {
    heading: "Phone ordering that brings callers back as regulars",
    headingTone: "dark",
    visualPanel: "green",
    visual: "ai-phone-mockup",
    features: [
      {
        icon: "trophy",
        title: "Turn callers into loyal regulars",
        body: "Every caller can join your loyalty program and guest list, then get follow-ups — so a one-time call becomes a repeat habit.",
      },
      {
        icon: "bolt",
        title: "Every call answered, instantly",
        body: "Tuvi picks up on the first ring and handles orders, hours, and delivery questions — even when the line will not stop ringing.",
      },
      {
        icon: "diamond",
        title: "Connected to everything you run",
        body: "Orders go to your kitchen, callers into your list, points and all. Most phone tools stop at the order. Tuvi keeps going.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Answers your phone. Grows your first-party sales.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "sky",
        label: "Natural conversation",
        title: "Feels like talking to a sharp team member. Available around the clock.",
        visual: "ai-phone-conversation",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Bring callers back",
        title: "Callers join your list, earn points, and get follow-ups.",
        visual: "ai-phone-loyalty-photo",
      },
      {
        layout: "half",
        theme: "light",
        label: "Real orders",
        title: "Real phone orders for the kitchen — not a text with a link.",
        visual: "ai-phone-food-tiles",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How does AI phone ordering work?",
        answer:
          "Tuvi answers your restaurant phone, takes the order in a natural conversation, confirms details, and sends it straight to your kitchen — while adding the caller to your customer list.",
      },
      {
        question: "What happens if the AI can't handle a call?",
        answer:
          "Complex or edge cases can transfer to your staff. Most routine orders, hours, and delivery questions are handled automatically so your team stays focused in the kitchen.",
      },
      {
        question: "How is this different from generic phone bots?",
        answer:
          "Tuvi does not stop at taking the order. Callers join your loyalty and marketing flows, orders connect to your kitchen ops, and the experience is built specifically for restaurants.",
      },
    ],
  },
};
