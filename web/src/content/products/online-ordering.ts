import type { ProductPageConfig } from "@/content/products/types";

export const onlineOrdering: ProductPageConfig = {
  meta: {
    title: "Online Ordering | Tuvi",
    description:
      "First-party online ordering built to grow direct sales. Commission-free orders, a guest list you own, and a checkout guests trust.",
  },
  hero: {
    heading: "Online ordering that grows your business — not a marketplace.",
    subheading:
      "Tuvi turns website and app visitors into commission-free orders. You keep the margin, the data, and the relationship.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-ordering-hero.png",
      imageAlt: "Jordan Hale from Orzo Vale Kouzina",
      title: "See how Jordan grows profitable direct orders with Tuvi",
      attribution: "Jordan Hale — Orzo Vale Kouzina",
    },
  },
  featureSplit: {
    heading: "Built to grow first-party orders week after week",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "ordering-phone-preview",
    features: [
      {
        icon: "bolt",
        title: "Always getting better",
        body: "We keep refining the guest flow so more people who find you finish a direct order.",
      },
      {
        icon: "wallet",
        title: "Keep more profit",
        body: "Direct orders through Tuvi avoid the marketplace fees that quietly eat your margins.",
      },
      {
        icon: "users",
        title: "Your customers, yours to keep",
        body: "Every order strengthens your guest list — more data, more follow-ups, more repeats under your brand.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Ordering that puts you back in control",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "cream-blue",
        label: "Guest-ready",
        title: "A polished order experience that feels familiar and fast.",
        visual: "ordering-app-showcase",
      },
      {
        layout: "half",
        theme: "white",
        label: "Own the list",
        title: "Grow a first-party guest list you can message anytime.",
        visual: "ordering-customer-list",
      },
      {
        layout: "half",
        theme: "sky",
        label: "Save on fees",
        title: "Direct orders: fairer for guests, better for your P&L.",
        visual: "fee-savings-toast",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How can I grow my direct orders?",
        answer:
          "Tuvi's ordering experience is tuned for conversion — clearer menus, guided add-ons, and a checkout guests trust — so more visitors order from you instead of delivery apps.",
      },
      {
        question: "What POS systems do you integrate with?",
        answer:
          "We integrate with the major restaurant POS systems. During onboarding we map your menu and order flow so tickets land where your kitchen already works.",
      },
      {
        question: "My POS comes with online ordering — why switch to Tuvi?",
        answer:
          "POS ordering is built for ops. Tuvi is built to grow sales — better guest UX, lower fees than marketplaces, and a customer list you own so you can bring guests back.",
      },
    ],
  },
};
