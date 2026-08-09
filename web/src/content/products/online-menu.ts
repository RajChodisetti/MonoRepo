import type { ProductPageConfig } from "@/content/products/types";

export const onlineMenu: ProductPageConfig = {
  meta: {
    title: "Online Menu | Tuvi",
    description:
      "A menu built to convert browsers into first-party orders. Clear layout, strong photos, and a checkout guests trust.",
  },
  hero: {
    heading: "A menu that turns browsers into buyers.",
    subheading:
      "Tuvi's online menu is shaped for conversion — so more people who land on your site place a commission-free order with you.",
    primaryCta: { label: "Get a free demo", href: "/book" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-inline-dining.png",
      imageAlt: "Jordan Hale from Orzo Vale Kouzina",
      title: "See how Jordan's menu lifts direct online sales",
      attribution: "Jordan Hale — Orzo Vale Kouzina",
    },
  },
  featureSplit: {
    heading: "Designed so guests finish the order",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "menu-items-stack",
    features: [
      {
        icon: "users",
        title: "Convert more of the traffic you already earn",
        body: "Layout, modifiers, and flow are tuned so hungry locals check out instead of bouncing to a marketplace.",
      },
      {
        icon: "badge",
        title: "Clear, inviting menus guests enjoy using",
        body: "Easy navigation and strong food photography keep people engaged long enough to order.",
      },
      {
        icon: "chart",
        title: "Refined as you grow",
        body: "We keep testing and tightening the experience so your direct order rate climbs over time.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Always clear. Always converting.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "cream-blue",
        visualSide: "right",
        label: "Built to convert",
        title: "Every screen nudges guests toward a completed first-party order.",
        visual: "rewards-phone",
      },
      {
        layout: "half",
        theme: "beige",
        label: "Looks the part",
        title: "Menus that feel premium without confusing the guest.",
        visual: "pita-wraps-menu",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Keeps improving",
        title: "Small refinements that compound into more weekly orders.",
        visual: "order-tracking-phone",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How is Tuvi's online menu different from a basic menu page?",
        answer:
          "It is built to convert — not just list dishes. Layout, photos, guided add-ons, and checkout are designed so more visitors place first-party orders with you.",
      },
      {
        question: "Can I customise my online menu?",
        answer:
          "Yes. Categories, items, photos, prices, modifiers, and branding stay yours — while the structure stays conversion-focused.",
      },
      {
        question: "What if my POS already has a menu online?",
        answer:
          "Many venues keep the POS for kitchen ops and use Tuvi for a better guest experience and higher direct sales. We can replace or improve what you have today.",
      },
    ],
  },
};
