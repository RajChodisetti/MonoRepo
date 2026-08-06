import type { ProductPageConfig } from "@/content/products/types";

export const catering: ProductPageConfig = {
  meta: {
    title: "Catering | Tuvi",
    description:
      "Take catering on your own site. Groups book a clear menu in minutes — commission-free, under your brand.",
  },
  hero: {
    heading: "Catering bookings on your website.",
    subheading:
      "Grow high-value group orders with a simple menu guests can book from in minutes. No marketplace fees. You own the guest.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-inline-street.png",
      imageAlt: "Priya Mehta from Steamrail Noodles",
      title: "“Every catering order via Tuvi stays ours — margin and guest.”",
      attribution: "Priya Mehta — Steamrail Noodles",
    },
  },
  featureSplit: {
    heading: "Built to win more high-margin group orders",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "catering-menu-stack",
    features: [
      {
        icon: "bolt",
        title: "Keep more profit per booking",
        body: "Catering sits inside Tuvi, so commission does not chew the ticket. More of every group order stays with you.",
      },
      {
        icon: "card",
        title: "Make booking easy for offices and events",
        body: "A clear menu and request flow on your site. No PDFs. No endless email chains.",
      },
      {
        icon: "trophy",
        title: "Get found by local groups",
        body: "Search-friendly catering pages on your domain help nearby businesses find and book with you directly.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Commission-free catering under your brand",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "blue",
        label: "Grow catering",
        title: "Get found locally and convert groups without marketplace fees.",
        visual: "catering-search",
      },
      {
        layout: "half",
        theme: "white",
        label: "Built for big tickets",
        title: "Booking that feels as smooth as everyday online ordering.",
        visual: "catering-food-collage",
      },
      {
        layout: "half",
        theme: "white-green",
        label: "On your domain",
        title: "Guests book without leaving your website or brand.",
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
        question: "I use a marketplace or an inquiry form. Why switch?",
        answer:
          "Marketplaces take fees and own the guest. Inquiry forms are slow. Tuvi lets groups book a clear catering menu on your site in minutes — under your brand, commission-free.",
      },
      {
        question: "How does Tuvi help me get more direct catering?",
        answer:
          "Search-friendly catering pages, a simple booking flow, and your existing website traffic convert more local businesses and groups without sending them to a marketplace.",
      },
    ],
  },
};
