import type { ProductPageConfig } from "@/content/products/types";

export const reviewsEngine: ProductPageConfig = {
  meta: {
    title: "Reviews Engine | Tuvi",
    description:
      "Grow your Google reviews the right way. Tuvi prompts happy guests at the right moment so your rating and visibility keep climbing.",
  },
  hero: {
    heading: "More Google reviews — without chasing every guest.",
    subheading:
      "Tuvi asks satisfied diners for a review at the right moment, so your Google rating strengthens and new locals keep finding you.",
    primaryCta: { label: "Get a free demo", href: "/demo" },
    secondaryCta: { label: "See how it works", href: "/how-it-works" },
    testimonial: {
      imageSrc: "/resources/resource-support-hero.png",
      imageAlt: "Riley Quinn from Quillnest Kitchen",
      title: "See how Riley grew Google reviews with Tuvi",
      attribution: "Riley Quinn — Quillnest Kitchen",
    },
  },
  featureSplit: {
    heading: "A simple engine for a steadier stream of reviews",
    headingTone: "dark",
    visualPanel: "none",
    visual: "reviews-owner-photo",
    features: [
      {
        icon: "diamond",
        title: "Reviews on autopilot",
        body: "Tuvi drives a steady stream of Google reviews without you asking every single customer in person.",
      },
      {
        icon: "users",
        title: "Better reviews from happier guests",
        body: "A smooth first-party experience leads to better feedback. Tuvi keeps the digital journey tidy.",
      },
      {
        icon: "person",
        title: "More reviews mean more new guests",
        body: "A stronger Google rating supports local visibility and helps nearby diners choose you with confidence.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Built to grow authentic Google reviews.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "green",
        label: "More reviews",
        title: "A steady stream of prompts timed for happy guests.",
        visual: "google-reviews-stack",
      },
      {
        layout: "half",
        theme: "white",
        label: "More visibility",
        title: "A stronger rating helps new locals find and trust you.",
        visual: "customers-flow",
      },
      {
        layout: "half",
        theme: "white-green",
        label: "Better ratings",
        title: "Great guest experiences turn into higher-quality reviews.",
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
          "Responding helps, and we make it easy. You can reply from one place, and light guidance helps you stay consistent so ratings and guest trust keep improving.",
      },
      {
        question: "How long does it take to start getting more reviews?",
        answer:
          "Most restaurants see new reviews within the first couple of weeks once the Reviews Engine is live. Momentum builds as more guests go through your ordering and dining flow.",
      },
    ],
  },
};
