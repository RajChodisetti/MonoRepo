import type { ProductPageConfig } from "@/content/products/types";

export const restaurantSeo: ProductPageConfig = {
  meta: {
    title: "Restaurant SEO | Tuvi",
    description:
      "Get more Google traffic with world-class SEO. Tuvi keeps your restaurant at the top of rankings, no matter how Google changes.",
  },
  hero: {
    heading: "Get more Google traffic with world-class SEO.",
    subheading:
      "Our SEO is built to keep you at the top of rankings, no matter how Google changes.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/sarkis-panossian.jpg",
      imageAlt: "Sarkis Panossian from Township Line Pizza",
      title: "See how our SEO helped Sarkis get top Google rankings",
      attribution: "Sarkis Panossian — Township Line Pizza",
    },
  },
  featureSplit: {
    heading: "World-class SEO that grows your online presence",
    headingTone: "dark",
    visualPanel: "green",
    visual: "seo-score",
    features: [
      {
        icon: "chart",
        title: "Get more traffic from places near you",
        body: "Tuvi's SEO is designed to get you as much traffic as possible from neighborhoods near you.",
      },
      {
        icon: "users",
        title: "Beat your competition on Google",
        body: "Our SEO studies your restaurant's competition and analyzes how to get you ahead.",
      },
      {
        icon: "pencil",
        title: "Never worry about Google updates",
        body: "Google's algorithm is always changing. Our experts handle every update so your rankings keep improving.",
      },
    ],
  },
  featureCards: {
    sectionBg: "white",
    cards: [
      {
        layout: "full",
        theme: "green",
        label: "AI SEO",
        title: "AI-optimized sites boost SEO and Google traffic.",
        visual: "seo-ai-search",
      },
      {
        layout: "half",
        theme: "sky",
        label: "Always up to date",
        title: "We track every Google update so your rankings never slip.",
        visual: "google-update",
      },
      {
        layout: "half",
        theme: "light",
        label: "Run by experts",
        title: "Our SEO experts constantly improve your restaurant's online visibility.",
        visual: "experts-avatars",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How is Tuvi's SEO different from other platforms?",
        answer:
          "Tuvi SEO is built specifically for restaurants — local search, menus, and nearby demand. We combine AI optimization with human experts who watch Google updates, so rankings keep climbing instead of stalling after a one-time audit.",
      },
      {
        question: "How long does it take to see results?",
        answer:
          "Most restaurants start seeing meaningful movement within a few weeks, with stronger gains over the first few months. Timing depends on competition in your area and how quickly we can optimize your site and listings.",
      },
      {
        question: "Do I need to do anything to maintain my SEO?",
        answer:
          "Very little. Our team tracks algorithm changes and keeps improving your visibility. You focus on running the restaurant — we handle the SEO maintenance.",
      },
    ],
  },
};
