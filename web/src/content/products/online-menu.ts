import type { ProductPageConfig } from "@/content/products/types";

export const onlineMenu: ProductPageConfig = {
  meta: {
    title: "Online Menu | Tuvi",
    description:
      "An online menu that turns visitors into customers. Designed by experts and constantly optimized to grow your online orders.",
  },
  hero: {
    heading: "An online menu that turns visitors into customers.",
    subheading:
      "Tuvi's online menu is designed by experts and constantly optimized for one thing: to grow your online orders.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/alex-lambroulis.jpg",
      imageAlt: "Alex Lambroulis from Karv Greek Kouzina",
      title: "See how our online menus help Alex grow his online sales",
      attribution: "Alex Lambroulis — Karv Greek Kouzina",
    },
  },
  featureSplit: {
    heading: "A menu that gets more guests to order.",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "menu-items-stack",
    features: [
      {
        icon: "users",
        title: "Turn more visitors into paying customers",
        body: "Tuvi's menu is designed to turn browsers into buyers, just like the top restaurant brands.",
      },
      {
        icon: "badge",
        title: "Beautiful, familiar designs that guests love",
        body: "A great-looking menu that's easy to navigate keeps guests engaged and drives more orders.",
      },
      {
        icon: "chart",
        title: "Continuously optimized by our experts",
        body: "Our team constantly tests and refines your online menu on your behalf.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Always optimized. Always converting.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "cream-blue",
        visualSide: "right",
        label: "Built to convert",
        title: "Designed to turn menu browsers into buyers.",
        visual: "rewards-phone",
      },
      {
        layout: "half",
        theme: "beige",
        label: "Great-looking menus",
        title: "Beautiful, easy menus guests enjoy using every time.",
        visual: "pita-wraps-menu",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Always improving",
        title: "We're always optimizing to grow your orders.",
        visual: "order-tracking-phone",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "How is Tuvi's online menu different from a regular menu?",
        answer:
          "It's built to convert — not just list food. Layout, photos, upsells, and ordering flow are designed and continuously tested so more visitors place orders.",
      },
      {
        question: "Can I customize my online menu?",
        answer:
          "Yes. Categories, items, photos, prices, modifiers, and branding can all be tailored to your restaurant while keeping a proven conversion-focused design.",
      },
      {
        question: "What if I already have an online menu through my POS?",
        answer:
          "We can replace or improve what you have. Many restaurants keep their POS for kitchen ops and use Tuvi's menu for a better guest experience and higher online sales.",
      },
    ],
  },
};
