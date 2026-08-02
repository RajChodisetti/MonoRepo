import type { ProductPageConfig } from "@/content/products/types";

export const restaurantWebsite: ProductPageConfig = {
  meta: {
    title: "Restaurant Website AI | Tuvi",
    description:
      "Restaurant websites built for sales first, style second. Tuvi builds your website to drive sales, grow Google traffic, and beat the competition.",
  },
  hero: {
    heading: "Restaurant websites built for sales first, style second.",
    subheading:
      "Tuvi builds your website to drive sales. Our proven design grows Google traffic, outranks delivery apps, and beats your competition.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/matt-miller.jpg",
      imageAlt: "Matt Miller from Mr M's Sandwich Shop",
      title:
        "“We're getting more clicks, and we're getting more traffic. And those clicks are turning into sales.”",
      attribution: "Matt Miller — Mr M's Sandwich Shop",
    },
  },
  featureSplit: {
    heading: "Your website could be driving more online sales",
    headingTone: "muted",
    visualPanel: "peach",
    visual: "gyro-preview",
    features: [
      {
        icon: "badge",
        title: "Upgrade your website to our proven design",
        body: "Thousands of restaurants use Tuvi. We know exactly how to design restaurant websites to drive more online orders.",
      },
      {
        icon: "percent",
        title: "We build websites that Google loves",
        body: "We've analyzed the Google algorithm. We build your website with world-class SEO that gets you top rankings.",
      },
      {
        icon: "gauge",
        title: "A website that's always getting better",
        body: "We're always studying the science of online sales. When we learn something new, we add it to your website right away.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "A website that's built to grow your business.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "blue",
        label: "AI SEO",
        title: "We use AI to grow your SEO and Google traffic.",
        visual: "ai-search-mock",
      },
      {
        layout: "half",
        theme: "beige",
        label: "Online ordering built-in",
        title: "Great ordering experience that grows your online sales.",
        visual: "ordering-ravioli",
      },
      {
        layout: "half",
        theme: "white-green",
        label: "Never stops improving",
        title: "You'll always get the latest best practices.",
        visual: "phone-improving",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "What happens to my current website?",
        answer:
          "We keep what already works and rebuild the rest for growth. Your domain, brand, and content can carry over while Tuvi upgrades SEO, ordering, and conversion paths so you do not lose traffic during the switch.",
      },
      {
        question: "How much can I customize my design?",
        answer:
          "A lot. Colors, photos, menu layout, and key sections are tailored to your restaurant. You get a polished look that still feels on-brand, without needing a designer for every small change.",
      },
      {
        question: "How long will this take?",
        answer:
          "Most restaurants go live in a few weeks. Timeline depends on how quickly we get your menu, photos, and feedback — once those are in, we move fast and keep you updated at every step.",
      },
    ],
  },
};
