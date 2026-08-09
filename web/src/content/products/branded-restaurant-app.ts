import type { ProductPageConfig } from "@/content/products/types";

export const brandedRestaurantApp: ProductPageConfig = {
  meta: {
    title: "Branded Restaurant App | Tuvi",
    description:
      "Your own restaurant app. Guests order, earn rewards, and come back — under your brand, commission-free.",
  },
  hero: {
    heading: "Your restaurant. Your app. Your customers.",
    subheading:
      "Tuvi gives independent venues a branded mobile app so guests order direct, earn loyalty, and skip the marketplaces.",
    primaryCta: { label: "Get a free demo", href: "/book" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-app-hero.png",
      imageAlt: "Priya Mehta from Steamrail Noodles",
      title: "See how Priya grows first-party orders with a branded app",
      attribution: "Priya Mehta — Steamrail Noodles",
    },
  },
  featureSplit: {
    heading: "An app experience that feels polished — and stays yours",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "branded-app-phone",
    features: [
      {
        icon: "phone",
        title: "Order ahead on your brand",
        body: "Guests browse your menu, customise items, and checkout in an app that looks and feels like yours.",
      },
      {
        icon: "trophy",
        title: "Loyalty built in",
        body: "Points, rewards, and offers live inside the app — so every order brings guests back to you.",
      },
      {
        icon: "bolt",
        title: "Push when it matters",
        body: "Send offers, order updates, and reminders without relying on third-party marketplaces.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Own the guest relationship on every device.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "green",
        label: "Your brand",
        title: "A polished app that puts your name front and centre.",
        visual: "branded-app-showcase",
      },
      {
        layout: "half",
        theme: "light",
        label: "Repeat visits",
        title: "Rewards and reorder shortcuts that bring guests back.",
        visual: "loyalty-card",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Direct sales",
        title: "Keep the margin and the customer list — not the marketplace.",
        visual: "app-photo-fill",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Do I need a developer to launch the app?",
        answer:
          "No. Tuvi builds and publishes your branded app with your menu, branding, and ordering already connected — you focus on the restaurant.",
      },
      {
        question: "Is the app available on iPhone and Android?",
        answer:
          "Yes. Guests can download your restaurant app on both iOS and Android, with the same ordering and loyalty experience.",
      },
      {
        question: "How does the app connect to online ordering?",
        answer:
          "Orders flow into the same Tuvi system as your website — kitchen tickets, customer list, and loyalty stay in one place.",
      },
    ],
  },
};
