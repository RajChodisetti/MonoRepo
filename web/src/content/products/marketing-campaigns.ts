import type { ProductPageConfig } from "@/content/products/types";

export const marketingCampaigns: ProductPageConfig = {
  meta: {
    title: "Marketing Campaigns | Tuvi",
    description:
      "Run restaurant marketing campaigns across email, SMS, and push — from one place, timed to drive orders.",
  },
  hero: {
    heading: "Campaigns that fill seats and grow orders.",
    subheading:
      "Plan, launch, and measure restaurant promotions across every channel without juggling five tools.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/gia-schultz.jpg",
      imageAlt: "Restaurant marketer reviewing campaign results",
      title: "See how campaigns bring guests back midweek",
      attribution: "Gia Schultz — HillCrust Pizza",
    },
  },
  featureSplit: {
    heading: "Promote once. Reach guests everywhere they listen.",
    headingTone: "dark",
    visualPanel: "green",
    visual: "campaign-promo",
    features: [
      {
        icon: "bolt",
        title: "Multi-channel in one launch",
        body: "Email, SMS, and push from a single campaign so your offer hits every inbox and phone.",
      },
      {
        icon: "chart",
        title: "Built for restaurants",
        body: "Templates for happy hour, slow nights, holidays, and loyalty offers — ready to customize.",
      },
      {
        icon: "gauge",
        title: "See what drove orders",
        body: "Track opens, clicks, and attributed sales so you know which campaigns actually paid off.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Marketing that runs on autopilot when you are busy.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "sky",
        label: "Scheduled",
        title: "Set it once and let Tuvi send at the right time.",
        visual: "campaign-calendar",
      },
      {
        layout: "half",
        theme: "white",
        label: "Targeted",
        title: "Reach regulars, lapsed guests, or first-time buyers.",
        visual: "email-sms-preview",
      },
      {
        layout: "half",
        theme: "light",
        label: "Measurable",
        title: "Connect campaigns to real order volume — not vanity metrics.",
        visual: "analytics-chart",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Can I reuse campaigns?",
        answer:
          "Yes. Save winning campaigns as templates and relaunch for weekly specials, holidays, or seasonal menus in a few clicks.",
      },
      {
        question: "Do campaigns sync with my customer list?",
        answer:
          "They use your Tuvi guest list automatically — including loyalty members and past online orderers.",
      },
      {
        question: "Will this spam my guests?",
        answer:
          "You control frequency and segments. Tuvi helps you message the right guests without blasting everyone every day.",
      },
    ],
  },
};
