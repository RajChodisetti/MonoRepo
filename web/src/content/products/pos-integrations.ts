import type { ProductPageConfig } from "@/content/products/types";

export const posIntegrations: ProductPageConfig = {
  meta: {
    title: "POS Integrations | Tuvi",
    description:
      "Connect Tuvi to your POS. Menus, orders, and items stay in sync so online and in-store stay aligned.",
  },
  hero: {
    heading: "Your POS and Tuvi, finally in sync.",
    subheading:
      "Connect supported POS systems so menus, modifiers, and orders stay accurate across online and in-store.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/mo-farraj.jpg",
      imageAlt: "Operator managing POS and online orders",
      title: "See how POS sync removes double entry",
      attribution: "Mo Farraj — Talkin' Tacos",
    },
  },
  featureSplit: {
    heading: "Less double work. Fewer mismatched menus.",
    headingTone: "dark",
    visualPanel: "green",
    visual: "pos-sync",
    features: [
      {
        icon: "card",
        title: "Connect once",
        body: "Link supported POS platforms so Tuvi can read the menu and send orders where they belong.",
      },
      {
        icon: "gear",
        title: "Stay accurate",
        body: "Item names, prices, and modifiers stay aligned so guests never order something you stopped selling.",
      },
      {
        icon: "bolt",
        title: "Faster service",
        body: "Orders flow without retyping — saving the counter and kitchen from busywork.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Integrations that keep service smooth.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "sky",
        label: "Synced",
        title: "See connection status at a glance across locations.",
        visual: "owner-dashboard",
      },
      {
        layout: "half",
        theme: "white",
        label: "Kitchen-ready",
        title: "Tickets arrive with the details your line needs.",
        visual: "kitchen-ticket",
      },
      {
        layout: "half",
        theme: "light",
        label: "Less chaos",
        title: "Stop updating the same menu in three different places.",
        visual: "campaign-calendar",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Which POS systems do you support?",
        answer:
          "Tuvi supports major restaurant POS platforms and continues expanding. We will confirm fit for your setup on a demo.",
      },
      {
        question: "What if I change POS later?",
        answer:
          "We help you reconnect and rematch menu items so online ordering keeps running with minimal downtime.",
      },
      {
        question: "Do I still need to edit menus in two places?",
        answer:
          "With a connected POS, Tuvi is designed to reduce duplicate edits so your source of truth stays clear.",
      },
    ],
  },
};
