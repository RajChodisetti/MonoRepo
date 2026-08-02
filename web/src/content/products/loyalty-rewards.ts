import type { ProductPageConfig } from "@/content/products/types";

export const loyaltyRewards: ProductPageConfig = {
  meta: {
    title: "Loyalty & Rewards | Tuvi",
    description:
      "Restaurant loyalty that guests actually use. Points, rewards, and perks that grow repeat orders under your brand.",
  },
  hero: {
    heading: "Loyalty that turns guests into regulars.",
    subheading:
      "Points, rewards, and VIP perks that live in your app and ordering flow — so guests come back for you, not a marketplace.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/enga-stanfield.jpg",
      imageAlt: "Restaurant owner celebrating loyal guests",
      title: "See how loyalty grows repeat weekly orders",
      attribution: "Enga Stanfield — Mattenga's Pizzeria",
    },
  },
  featureSplit: {
    heading: "Rewards that feel simple for guests and staff",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "loyalty-card",
    features: [
      {
        icon: "trophy",
        title: "Earn on every order",
        body: "Guests collect points online and in-app automatically — no punch cards to manage.",
      },
      {
        icon: "diamond",
        title: "Rewards they want",
        body: "Free items, discounts, and member-only offers that make coming back obvious.",
      },
      {
        icon: "users",
        title: "Know your best guests",
        body: "See who spends, who is slipping away, and who deserves a VIP nudge.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "A loyalty program that runs itself.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "green",
        label: "Guest-facing",
        title: "Clear progress toward the next reward every visit.",
        visual: "loyalty-rewards-grid",
      },
      {
        layout: "half",
        theme: "white",
        label: "Marketing",
        title: "Trigger offers when someone is close to redeeming.",
        visual: "email-sms-preview",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Retention",
        title: "Win back quiet guests before they forget you.",
        visual: "owner-photo-fill",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Does loyalty work with online ordering?",
        answer:
          "Yes. Points earn and redeem across website, app, and supported ordering flows so the program stays consistent.",
      },
      {
        question: "Can I customize rewards?",
        answer:
          "Absolutely. Set point values, reward items, and member tiers to match how your restaurant operates.",
      },
      {
        question: "How do staff handle redemptions?",
        answer:
          "Rewards show clearly on the order and kitchen ticket so your team can honor them without extra steps.",
      },
    ],
  },
};
