import type { ProductPageConfig } from "@/content/products/types";

export const ownerApp: ProductPageConfig = {
  meta: {
    title: "Operator App | Tuvi",
    description:
      "Run the floor from your phone. Sales, orders, and guest signals in one Tuvi operator app — wherever you are.",
  },
  hero: {
    heading: "The venue in your pocket.",
    subheading:
      "Check sales, manage orders, and stay across performance — whether you are on the pass or away from the dining room.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-app-hero.png",
      imageAlt: "Sam Rivera from Brightkiln Kitchen",
      title: "See how Sam stays across service from anywhere",
      attribution: "Sam Rivera — Brightkiln Kitchen",
    },
  },
  featureSplit: {
    heading: "Day-to-day control without living at the till",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "owner-dashboard",
    features: [
      {
        icon: "gauge",
        title: "Live sales pulse",
        body: "See today's orders, ticket size, and trends without waiting for end-of-night exports.",
      },
      {
        icon: "gear",
        title: "Operate on the go",
        body: "Adjust hours, menus, and promos when something changes mid-service.",
      },
      {
        icon: "users",
        title: "Know your guests",
        body: "Spot regulars, new customers, and loyalty activity from one screen you actually own.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Built for operators who cannot be everywhere at once.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "cream-blue",
        label: "Clarity",
        title: "A clean dashboard that answers “how are we tracking?” instantly.",
        visual: "analytics-chart",
      },
      {
        layout: "half",
        theme: "white",
        label: "Alerts",
        title: "Stay informed when something needs a quick decision.",
        visual: "push-notif-stack",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Peace of mind",
        title: "Leave the floor without losing visibility of first-party orders.",
        visual: "owner-photo-fill",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Is the operator app separate from the guest app?",
        answer:
          "Yes. Guests get your branded ordering app. You get the Tuvi operator app for sales, operations, and controls.",
      },
      {
        question: "Can multiple managers use it?",
        answer:
          "Yes. Give access to trusted managers so the team can monitor service without sharing one login.",
      },
      {
        question: "Does it work with online ordering?",
        answer:
          "It sits on top of your Tuvi stack — ordering, loyalty, and reporting stay connected so you own the guest relationship end to end.",
      },
    ],
  },
};
