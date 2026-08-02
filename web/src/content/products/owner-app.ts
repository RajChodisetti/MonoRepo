import type { ProductPageConfig } from "@/content/products/types";

export const ownerApp: ProductPageConfig = {
  meta: {
    title: "Owner App | Tuvi",
    description:
      "Run your restaurant from your phone. Orders, sales, and guest insights in one owner app.",
  },
  hero: {
    heading: "Your restaurant in your pocket.",
    subheading:
      "Check sales, manage orders, and stay on top of performance — whether you are in the dining room or away.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/alex-lambroulis.jpg",
      imageAlt: "Restaurant owner using the Tuvi owner app",
      title: "See how owners stay in control from anywhere",
      attribution: "Alex Lambroulis — Olive & Oak",
    },
  },
  featureSplit: {
    heading: "Everything you need to run the business day to day",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "owner-dashboard",
    features: [
      {
        icon: "gauge",
        title: "Live sales pulse",
        body: "See today's orders, ticket size, and trends without waiting for end-of-night reports.",
      },
      {
        icon: "gear",
        title: "Operate on the go",
        body: "Adjust hours, menus, and promos when something changes mid-service.",
      },
      {
        icon: "users",
        title: "Know your guests",
        body: "Spot regulars, new customers, and loyalty activity from one screen.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Built for owners who cannot be everywhere at once.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "cream-blue",
        label: "Clarity",
        title: "A clean dashboard that answers “how are we doing?” instantly.",
        visual: "analytics-chart",
      },
      {
        layout: "half",
        theme: "white",
        label: "Alerts",
        title: "Stay informed when something needs attention.",
        visual: "push-notif-stack",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Peace of mind",
        title: "Leave the floor without losing visibility.",
        visual: "owner-photo-fill",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Is the Owner App separate from the guest app?",
        answer:
          "Yes. Guests get your branded ordering app. You get the Owner App for operations, sales, and controls.",
      },
      {
        question: "Can multiple managers use it?",
        answer:
          "Yes. Give access to trusted managers so the team can monitor service without sharing one login.",
      },
      {
        question: "Does it work with online ordering?",
        answer:
          "It sits on top of your Tuvi stack — ordering, loyalty, and reporting stay connected.",
      },
    ],
  },
};
