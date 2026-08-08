import type { ProductPageConfig } from "@/content/products/types";

export const posIntegrations: ProductPageConfig = {
  meta: {
    title: "POS Integrations | Tuvi",
    description:
      "Connect Tuvi to your POS. Menus, modifiers, and orders stay aligned so online and in-store tell the same story.",
  },
  hero: {
    heading: "Your POS and Tuvi, finally aligned.",
    subheading:
      "Connect supported POS systems so menus, modifiers, and tickets stay accurate across first-party online and in-store service.",
    primaryCta: { label: "Get a free demo", href: "/book" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-help-hero.png",
      imageAlt: "Avery Knox from Nonna Parcel Kitchen",
      title: "See how Avery stopped double-entering menus",
      attribution: "Avery Knox — Nonna Parcel Kitchen",
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
        body: "Orders flow without retyping — saving the counter and kitchen from busywork during rush.",
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
          "We help you reconnect and rematch menu items so first-party ordering keeps running with minimal downtime.",
      },
      {
        question: "Do I still need to edit menus in two places?",
        answer:
          "With a connected POS, Tuvi is designed to reduce duplicate edits so your source of truth stays clear.",
      },
    ],
  },
};
