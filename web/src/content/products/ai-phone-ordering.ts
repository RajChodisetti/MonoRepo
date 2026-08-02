import type { ProductPageConfig } from "@/content/products/types";

export const aiPhoneOrdering: ProductPageConfig = {
  meta: {
    title: "AI Phone Ordering | Tuvi",
    description:
      "AI that answers every call you're currently missing. Tuvi answers every call, takes the order, and sends it to your kitchen.",
  },
  hero: {
    heading: "AI that answers every call you're currently missing.",
    subheading: "Tuvi answers every call, takes the order, and sends it to your kitchen.",
    primaryCta: { label: "Join the waitlist", href: "#" },
    testimonial: {
      imageSrc: "/product/reviews-owner.jpg",
      imageAlt: "Restaurant owner managing orders on a phone",
    },
  },
  featureSplit: {
    heading: "AI phone ordering that brings customers back.",
    headingTone: "dark",
    visualPanel: "green",
    visual: "ai-phone-mockup",
    features: [
      {
        icon: "trophy",
        title: "Turn callers into loyal regulars",
        body: "Tuvi does more than take the order. Every caller joins your loyalty program and customer list, and gets your marketing automatically, turning a one-time call into a regular.",
      },
      {
        icon: "bolt",
        title: "Every call answered, instantly",
        body: "Tuvi picks up on the first ring and handles it all: orders, hours, delivery questions. Even when the phone won't stop, your team can keep moving.",
      },
      {
        icon: "diamond",
        title: "Connected to everything you run",
        body: "Every order goes to your kitchen, every caller into your customer list, points and all. Most tools stop at the order. Tuvi does the rest.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Answers your phone. Grows your sales.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "sky",
        label: "Natural conversation",
        title: "Feels like talking to your best employee. Available 24/7.",
        visual: "ai-phone-conversation",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Bring phone callers back",
        title: "Callers join your list, earn points, and get follow-ups.",
        visual: "ai-phone-loyalty-photo",
      },
      {
        layout: "half",
        theme: "light",
        label: "Real orders, not redirects",
        title: "Real orders, taken on the phone. Not a text with a link.",
        visual: "ai-phone-food-tiles",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How does AI phone ordering work?",
        answer:
          "Tuvi answers your restaurant phone, takes the order in a natural conversation, confirms details, and sends it straight to your kitchen — while adding the caller to your customer list.",
      },
      {
        question: "What happens if the AI can't handle a call?",
        answer:
          "Complex or edge cases can transfer to your staff. Most routine orders, hours, and delivery questions are handled automatically so your team stays focused in the kitchen.",
      },
      {
        question: "How is this different from other AI phone bots?",
        answer:
          "Tuvi doesn't stop at taking the order. Callers join your loyalty and marketing flows, orders connect to your kitchen ops, and the experience is built specifically for restaurants — not generic phone bots.",
      },
    ],
  },
};
