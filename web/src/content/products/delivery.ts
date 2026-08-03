import type { ProductPageConfig } from "@/content/products/types";

export const delivery: ProductPageConfig = {
  meta: {
    title: "Delivery | Tuvi",
    description:
      "Profitable delivery and a great guest experience. Get food to your customers on time, by top-rated drivers, at a fair price.",
  },
  hero: {
    heading: "Profitable delivery and a great guest experience.",
    subheading:
      "Get food to your customers on time, by top-rated drivers, at a fair price.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/gia-schultz.jpg",
      imageAlt: "Gia Schultz from The Modern Vegan",
      title: "See why Gia thinks food delivery via Tuvi is better",
      attribution: "Gia Schultz — The Modern Vegan",
    },
  },
  featureSplit: {
    heading: "Delivery that's better for you and your guests",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "delivery-map-phone",
    features: [
      {
        icon: "wallet",
        title: "Pay a flat rate for delivery",
        body: "Use third-party drivers at fixed rates. Guests pay delivery on small orders, you chip in for larger ones.",
      },
      {
        icon: "car",
        title: "Reach more guests with reliable delivery",
        body: "Top-rated drivers deliver across your service area, so you can serve more customers.",
      },
      {
        icon: "phone",
        title: "A direct line to your customers",
        body: "You can call customers directly. We'll cover refunds for any delivery issues.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Get your food delivered by top-rated drivers",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "green",
        label: "Third party drivers",
        title: "Top third-party drivers deliver for you. One flat fee. No markup from us.",
        visual: "delivery-tracking-card",
      },
      {
        layout: "half",
        theme: "light",
        label: "Control delivery options",
        title: "Use your in-house delivery drivers, third-party, or both.",
        visual: "delivery-control-map",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Better for your guests",
        title: "Call your customers if needed. We'll pay for any delivery problems.",
        visual: "delivery-guest-photo",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Why would customers order from my app instead of the third parties?",
        answer:
          "Lower fees, a better ordering experience, and delivery that still feels premium. Guests get your brand — not a marketplace — while you keep the relationship and more of the margin.",
      },
      {
        question: "Who pays for delivery, the guest or the restaurant?",
        answer:
          "You set the rules. Guests can pay delivery on smaller orders, and you can chip in on larger ones — with flat third-party rates so costs stay predictable.",
      },
      {
        question: "Why would the third-party apps go for this?",
        answer:
          "Drivers still get paid work through their networks. Tuvi routes delivery through top-rated partners at a fair flat fee — without marketplace markups on your menu.",
      },
    ],
  },
};
