import type { ProductPageConfig } from "@/content/products/types";

export const emailTextMarketing: ProductPageConfig = {
  meta: {
    title: "Email & Text Marketing | Tuvi",
    description:
      "Email and SMS for restaurants. Win back guests, announce specials, and drive commission-free direct orders.",
  },
  hero: {
    heading: "Email & text that bring guests back.",
    subheading:
      "Send the right message at the right time — specials, loyalty rewards, and win-backs that turn into first-party orders.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-email-hero.png",
      imageAlt: "Avery Knox from Nonna Parcel Kitchen",
      title: "See how Avery uses email and SMS to refill slow nights",
      attribution: "Avery Knox — Nonna Parcel Kitchen",
    },
  },
  featureSplit: {
    heading: "Stay top of mind without living in a spreadsheet",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "email-sms-preview",
    features: [
      {
        icon: "bolt",
        title: "Automations that work",
        body: "Welcome series, win-backs, and birthday offers run automatically from your guest list.",
      },
      {
        icon: "users",
        title: "Segments that matter",
        body: "Message by visit history, spend, or loyalty status so offers feel personal — not spammy.",
      },
      {
        icon: "chart",
        title: "Tied to sales",
        body: "See which messages drove orders so you double down on what works.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Two channels. One guest list. More return visits.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "cream-blue",
        label: "Email",
        title: "Messages that showcase your food and offers clearly.",
        visual: "campaign-promo",
      },
      {
        layout: "half",
        theme: "white",
        label: "SMS",
        title: "Short texts guests actually open — and act on.",
        visual: "push-notif-stack",
      },
      {
        layout: "half",
        theme: "light",
        label: "Compliance-ready",
        title: "Opt-ins and unsubscribes handled the right way.",
        visual: "loyalty-card",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Where does the guest list come from?",
        answer:
          "Online orders, loyalty signups, and app users flow into one Tuvi list you own and can message anytime.",
      },
      {
        question: "Can I send both email and SMS together?",
        answer:
          "Yes. Launch them separately or as part of a multi-channel campaign timed to the same offer.",
      },
      {
        question: "Do I write everything from scratch?",
        answer:
          "No. Start from restaurant-ready templates and customise the copy, offer, and timing.",
      },
    ],
  },
};
