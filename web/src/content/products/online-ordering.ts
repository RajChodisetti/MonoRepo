import type { ProductPageConfig } from "@/content/products/types";

export const onlineOrdering: ProductPageConfig = {
  meta: {
    title: "Online Ordering | Tuvi",
    description:
      "Your online ordering should grow your business. Tuvi is built to grow direct orders — always testing what boosts your sales.",
  },
  hero: {
    heading: "Your online ordering should grow your business.",
    subheading:
      "Tuvi is the only ordering system that works to grow your orders. We're always running tests to see how we can boost your sales.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/mo-farraj.jpg",
      imageAlt: "Mo Farraj from Talkin' Tacos",
      title: "See how our online ordering helps Mo grow his business profitably",
      attribution: "Mo Farraj — Talkin' Tacos",
    },
  },
  featureSplit: {
    heading: "Online ordering built to grow your direct orders",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "ordering-phone-preview",
    features: [
      {
        icon: "bolt",
        title: "Always getting better",
        body: "Tuvi is built to grow your direct orders over time. We're always testing what gets more customers ordering from you.",
      },
      {
        icon: "wallet",
        title: "Keep more profit",
        body: "Keep more of every order. Direct orders through Tuvi come without the fees that eat into your margins.",
      },
      {
        icon: "users",
        title: "Your customers, yours to keep",
        body: "Every direct order builds your customer list. More data, more connections, and more chances to bring them back.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Online ordering that grows direct orders",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "cream-blue",
        label: "Built like the best",
        title: "Feels like the apps your customers use daily.",
        visual: "ordering-app-showcase",
      },
      {
        layout: "half",
        theme: "white",
        label: "Grow your customer list",
        title: "Own your data. Grow your list. Connect with your customers.",
        visual: "ordering-customer-list",
      },
      {
        layout: "half",
        theme: "sky",
        label: "Save on fees",
        title: "Direct orders: cheaper for guests, better for you.",
        visual: "fee-savings-toast",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How can I grow my direct orders?",
        answer:
          "Tuvi's ordering experience is continuously tested for conversion — clearer menus, smarter upsells, and a checkout guests trust — so more visitors order from you instead of delivery apps.",
      },
      {
        question: "What POS systems do you integrate with?",
        answer:
          "We integrate with the major restaurant POS systems. During onboarding we map your menu and order flow so tickets land where your kitchen already works.",
      },
      {
        question: "My POS comes with online ordering, why should I switch to Tuvi?",
        answer:
          "POS ordering is built for ops. Tuvi is built to grow sales — better guest UX, lower fees than marketplaces, and a customer list you own so you can bring guests back.",
      },
    ],
  },
};
