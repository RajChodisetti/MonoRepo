import type { ProductPageConfig } from "@/content/products/types";

export const pushNotifications: ProductPageConfig = {
  meta: {
    title: "Push Notifications Marketing | Tuvi",
    description:
      "Push notifications that drive orders. Reach app guests with offers, updates, and reminders that feel timely.",
  },
  hero: {
    heading: "Push that gets guests to open and order.",
    subheading:
      "Send timely notifications through your branded app — happy hour, order status, and loyalty wins.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/matt-miller.jpg",
      imageAlt: "Restaurant owner sending push offers",
      title: "See how push notifications lift same-day orders",
      attribution: "Matt Miller — Metro Bistro",
    },
  },
  featureSplit: {
    heading: "Reach guests the moment they should order",
    headingTone: "dark",
    visualPanel: "green",
    visual: "push-notif-stack",
    features: [
      {
        icon: "bolt",
        title: "Instant reach",
        body: "Notifications land on the lock screen — perfect for flash specials and slow-hour fills.",
      },
      {
        icon: "phone",
        title: "Tied to your app",
        body: "Pushes deepen engagement with your branded app, not a third-party feed.",
      },
      {
        icon: "diamond",
        title: "Smart timing",
        body: "Trigger on order events, loyalty milestones, or schedules you set once.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Notifications guests welcome — not mute.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "sky",
        label: "Offers",
        title: "Promote specials when guests are deciding what to eat.",
        visual: "campaign-promo",
      },
      {
        layout: "half",
        theme: "white",
        label: "Order updates",
        title: "Keep pickup and delivery guests informed automatically.",
        visual: "kitchen-ticket",
      },
      {
        layout: "half",
        theme: "light",
        label: "Loyalty",
        title: "Celebrate points and rewards so guests come back sooner.",
        visual: "loyalty-rewards-grid",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "Do guests need the branded app?",
        answer:
          "Push notifications work with your Tuvi restaurant app. Guests who install it can opt in to receive offers and updates.",
      },
      {
        question: "Can I schedule pushes ahead of time?",
        answer:
          "Yes. Schedule for happy hour, weekends, or holidays — or trigger them from campaigns and automations.",
      },
      {
        question: "How do I avoid sending too many?",
        answer:
          "Frequency caps and segments help you message the right guests without overwhelming everyone.",
      },
    ],
  },
};
