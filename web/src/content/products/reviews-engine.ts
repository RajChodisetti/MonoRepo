import type { ProductPageConfig } from "@/content/products/types";

export const reviewsEngine: ProductPageConfig = {
  meta: {
    title: "Reviews Engine | Tuvi",
    description:
      "We'll grow your Google reviews. Tuvi automatically drives more reviews for your restaurant so your rating keeps growing.",
  },
  hero: {
    heading: "We'll grow your Google reviews.",
    subheading:
      "Tuvi automatically drives more reviews for your restaurant, so your Google rating keeps growing and new customers keep finding you.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/enga-stanfield.jpg",
      imageAlt: "Enga Stanfield from Mattenga's Pizzeria",
      title: "See how our Reviews Engine boosts Enga's online presence",
      attribution: "Enga Stanfield — Mattenga's Pizzeria",
    },
  },
  featureSplit: {
    heading: "The easiest way to get more Google reviews",
    headingTone: "dark",
    visualPanel: "none",
    visual: "reviews-owner-photo",
    features: [
      {
        icon: "diamond",
        title: "More Google reviews automatically.",
        body: "Tuvi drives a steady stream of Google reviews without you having to ask every single customer.",
      },
      {
        icon: "users",
        title: "Better reviews from happier guests",
        body: "A great guest experience leads to great reviews. Tuvi ensures every online interaction is seamless.",
      },
      {
        icon: "person",
        title: "More reviews means more new customers",
        body: "A stronger Google rating improves your SEO visibility and attracts more guests.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "An engine for driving more Google reviews.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "green",
        label: "More reviews",
        title: "Automatically drives a steady stream of Google reviews.",
        visual: "google-reviews-stack",
      },
      {
        layout: "half",
        theme: "white",
        label: "More visibility",
        title: "A stronger Google rating brings in more new customers.",
        visual: "customers-flow",
      },
      {
        layout: "half",
        theme: "white-green",
        label: "Better ratings",
        title: "Great guest experiences lead to higher-quality reviews.",
        visual: "reviews-phone-mockup",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How does Tuvi get me more reviews?",
        answer:
          "After a great visit, Tuvi prompts happy guests to leave a Google review at the right moment — automatically — so you get a steady stream without chasing every customer.",
      },
      {
        question: "Do I need to respond to my reviews?",
        answer:
          "Responding helps, and we make it easy. You can reply from one place, and our guidance helps you stay consistent so ratings and guest trust keep improving.",
      },
      {
        question: "How long does it take to start getting more reviews?",
        answer:
          "Most restaurants see new reviews within the first couple of weeks once the Reviews Engine is live. Momentum builds as more guests go through your ordering and dining flow.",
      },
    ],
  },
};
