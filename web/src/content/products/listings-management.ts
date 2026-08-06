import type { ProductPageConfig } from "@/content/products/types";

export const listingsManagement: ProductPageConfig = {
  meta: {
    title: "Listings Management | Tuvi",
    description:
      "Keep restaurant listings accurate everywhere. Tuvi syncs your details so locals find you and first-party traffic stays strong.",
  },
  hero: {
    heading: "Accurate listings. Stronger local discovery.",
    subheading:
      "Tuvi keeps your name, hours, and address consistent across the web so hungry locals find the right venue — and land on you.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-marketing-hero.png",
      imageAlt: "Dev Kapoor from Masala Quill Kitchen",
      title:
        "“Across three locations, our details stay clean. Google is still how most new guests find us.”",
      attribution: "Dev Kapoor — Masala Quill Kitchen",
    },
  },
  featureSplit: {
    heading: "Consistent listings.\nClearer search. More direct traffic.",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "listing-map-card",
    features: [
      {
        icon: "gear",
        title: "Always consistent across platforms",
        body: "Your name, address, phone, and website stay aligned everywhere guests look you up.",
      },
      {
        icon: "diamond",
        title: "The small details that matter for search",
        body: "A typo in hours or suburb can confuse guests and hurt visibility. Tuvi keeps the details tight.",
      },
      {
        icon: "bolt",
        title: "Managed for you",
        body: "We monitor and update listings so you are not chasing directories between lunch and dinner service.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Accurate listings quietly support local search.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "indigo",
        label: "Always consistent",
        title: "Your info stays in sync across the directories that matter.",
        visual: "listings-synced",
      },
      {
        layout: "half",
        theme: "beige",
        label: "Search-critical details",
        title: "We catch the tiny mismatches that confuse Google and guests.",
        visual: "address-fix-cards",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Fully managed",
        title: "A team watching your listings so you can focus on service.",
        visual: "listings-experts-photo",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "What is listings management and why does it matter?",
        answer:
          "Listings management keeps your restaurant's name, address, phone, hours, and website accurate across Google, social, and directories. Consistent details support local search and help customers find the right location every time.",
      },
      {
        question: "Which platforms does Tuvi manage my listings on?",
        answer:
          "We sync across major platforms customers use — including Google, Facebook, and other key directories — so your info stays consistent everywhere locals look.",
      },
      {
        question: "Couldn't I just update my listings myself?",
        answer:
          "You could, but details drift across dozens of sites. Tuvi monitors and updates them for you so visibility and guest trust do not slip from small mistakes.",
      },
    ],
  },
};
