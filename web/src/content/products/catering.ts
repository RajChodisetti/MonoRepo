import type { ProductPageConfig } from "@/content/products/types";

export const catering: ProductPageConfig = {
  meta: {
    title: "Catering | Tuvi",
    description:
      "Take catering orders directly on your website. Grow catering with a simple menu guests can book from in minutes — no marketplace fees.",
  },
  hero: {
    heading: "Take catering orders directly on your website.",
    subheading:
      "Grow your catering orders with a simple menu guests can book from in minutes. No marketplace fees.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/lisa-desisto.jpg",
      imageAlt: "Lisa Desisto from Rig A' Tony's",
      title: '"Every time someone orders via Tuvi, it\'s more money in my pocket."',
      attribution: "Lisa Desisto — Rig A' Tony's",
    },
  },
  featureSplit: {
    heading: "Catering built to win more\nhigh-margin orders",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "catering-menu-stack",
    features: [
      {
        icon: "bolt",
        title: "Earn more profit per order",
        body: "Catering is included with Tuvi, so more profit goes right to you. Stop paying crazy fees for catering orders.",
      },
      {
        icon: "card",
        title: "Make booking easy for guests",
        body: "A simple, app-like menu and request form on your site. No PDFs. No endless back-and-forth.",
      },
      {
        icon: "trophy",
        title: "Get found by local businesses",
        body: "Search-friendly pages on your domain help nearby groups and offices find and book with you.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Commission-free orders, under your brand",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "blue",
        label: "Grow catering orders",
        title: "Get found on Google and drive commission-free orders.",
        visual: "catering-search",
      },
      {
        layout: "half",
        theme: "white",
        label: "Built for big orders",
        title: "Booking that feels like the apps your guests use.",
        visual: "catering-food-collage",
      },
      {
        layout: "half",
        theme: "white-green",
        label: "Easy to find",
        title: "Guests order without leaving your website.",
        visual: "catering-phone-mockup",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How much does Tuvi charge for catering?",
        answer:
          "Catering is included with Tuvi — no marketplace commission on those orders. You keep more of every catering booking that comes through your site.",
      },
      {
        question: "Can I set minimums, notice times, and fees?",
        answer:
          "Yes. Set order minimums, lead times, delivery or pickup fees, and custom options so catering runs the way your kitchen needs.",
      },
      {
        question: "I use a marketplace or an inquiry form for catering. Why switch?",
        answer:
          "Marketplaces take fees and own the guest. Inquiry forms are slow. Tuvi lets guests book a clear catering menu on your site in minutes — under your brand, commission-free.",
      },
      {
        question: "How does Tuvi get me more direct catering orders?",
        answer:
          "Search-friendly catering pages, a simple booking flow, and your existing website traffic convert more local businesses and groups without sending them to a marketplace.",
      },
    ],
  },
};
