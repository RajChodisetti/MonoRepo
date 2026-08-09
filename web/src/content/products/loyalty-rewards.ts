import type { ProductPageConfig } from "@/content/products/types";

export const loyaltyRewards: ProductPageConfig = {
  meta: {
    title: "Loyalty & Rewards | Tuvi",
    description:
      "Restaurant loyalty guests actually use. Points and perks that grow repeat first-party orders under your brand.",
  },
  hero: {
    heading: "Loyalty that turns guests into regulars.",
    subheading:
      "Points, rewards, and VIP perks that live in your app and ordering flow — so guests come back for you, not a marketplace.",
    primaryCta: { label: "Get a free demo", href: "/book" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-support-hero.png",
      imageAlt: "Sam Rivera from Brightkiln Kitchen",
      title: "See how loyalty grew Sam's midweek repeats",
      attribution: "Sam Rivera — Brightkiln Kitchen",
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
        body: "See who spends, who is slipping away, and who deserves a VIP nudge — on a list you own.",
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
          "Yes. Points earn and redeem across website, app, and supported ordering flows so the program stays consistent and first-party.",
      },
      {
        question: "Can I customise rewards?",
        answer:
          "Absolutely. Set point values, reward items, and member tiers to match how your restaurant operates.",
      },
      {
        question: "How do staff handle redemptions?",
        answer:
          "Rewards show clearly on the order and kitchen ticket so your team can honour them without extra steps.",
      },
    ],
  },
};
