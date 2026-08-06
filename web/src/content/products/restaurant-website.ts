import type { ProductPageConfig } from "@/content/products/types";

export const restaurantWebsite: ProductPageConfig = {
  meta: {
    title: "Restaurant Website | Tuvi",
    description:
      "A restaurant website built to sell. Drive Google traffic, beat marketplace listings, and convert visitors into first-party orders.",
  },
  hero: {
    heading: "A website built to sell food — not just look pretty.",
    subheading:
      "Tuvi builds restaurant sites that win local search, showcase your brand, and turn visits into commission-free online orders.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-website-hero.png",
      imageAlt: "Sandy Sei from Lantern Ridge Kitchen",
      title:
        "“More clicks are landing on our site — and those clicks are turning into orders.”",
      attribution: "Sandy Sei — Lantern Ridge Kitchen",
    },
  },
  featureSplit: {
    heading: "Your website should be a growth channel",
    headingTone: "muted",
    visualPanel: "peach",
    visual: "gyro-preview",
    features: [
      {
        icon: "badge",
        title: "Proven restaurant layouts",
        body: "Designs shaped around menus, ordering, and local discovery — so guests know what to do in seconds.",
      },
      {
        icon: "percent",
        title: "Pages Google can understand",
        body: "Structure, speed, and local signals that help nearby searchers find you before they hit a marketplace.",
      },
      {
        icon: "gauge",
        title: "Always getting sharper",
        body: "As we learn what converts for venues like yours, those improvements flow into your site.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "A site that grows first-party orders.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "blue",
        label: "Local search",
        title: "Built to earn visibility when hungry locals search nearby.",
        visual: "ai-search-mock",
      },
      {
        layout: "half",
        theme: "beige",
        label: "Ordering built in",
        title: "A guest-ready order path that grows direct sales.",
        visual: "ordering-ravioli",
      },
      {
        layout: "half",
        theme: "white-green",
        label: "Keeps evolving",
        title: "You get ongoing refinements — not a static brochure site.",
        visual: "phone-improving",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "What happens to my current website?",
        answer:
          "We keep what already works and rebuild the rest for growth. Your domain, brand, and content can carry over while Tuvi upgrades search readiness, ordering, and conversion paths so you do not lose traffic during the switch.",
      },
      {
        question: "How much can I customise the design?",
        answer:
          "A lot. Colours, photos, menu layout, and key sections are tailored to your restaurant. You get a polished look that still feels on-brand, without needing a designer for every small change.",
      },
      {
        question: "How long will this take?",
        answer:
          "Most restaurants go live in a few weeks. Timeline depends on how quickly we get your menu, photos, and feedback — once those are in, we move fast and keep you updated.",
      },
    ],
  },
};
