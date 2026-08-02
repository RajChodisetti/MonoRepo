import type { ProductPageConfig } from "@/content/products/types";

export const listingsManagement: ProductPageConfig = {
  meta: {
    title: "Listings Management | Tuvi",
    description:
      "Better restaurant listings. Better SEO. Tuvi syncs your restaurant's information everywhere online automatically.",
  },
  hero: {
    heading: "Better restaurant listings. Better SEO.",
    subheading:
      "Tuvi syncs your restaurant's information everywhere online automatically, so your SEO stays strong and customers always find you.",
    primaryCta: { label: "Get a free demo", href: "#" },
    secondaryCta: { label: "View pricing", href: "/pricing" },
    testimonial: {
      imageSrc: "/product/rahul-bhatia.jpg",
      imageAlt: "Rahul Bhatia from Saffron Indian Kitchen",
      title:
        '"SEO across all three locations is seamless. The number one way new customers find us is via Google."',
      attribution: "Rahul Bhatia — Saffron Indian Kitchen",
    },
  },
  featureSplit: {
    heading: "Perfectly consistent listings.\nBetter SEO. More Traffic.",
    headingTone: "dark",
    visualPanel: "peach",
    visual: "listing-map-card",
    features: [
      {
        icon: "gear",
        title: "Always consistent across every platform",
        body: "Your name, address, phone, and website: perfectly in sync everywhere online.",
      },
      {
        icon: "diamond",
        title: "The SEO details that really matter",
        body: "Even a small typo in your listing can hurt your Google ranking. Tuvi gets every detail exactly right.",
      },
      {
        icon: "bolt",
        title: "Managed by our experts.",
        body: "Our team monitors and updates your listings so you never have to think about it.",
      },
    ],
  },
  featureCards: {
    sectionHeading: "Consistent, accurate listings: the secret behind great SEO.",
    sectionBg: "beige",
    cards: [
      {
        layout: "full",
        theme: "indigo",
        label: "Always consistent",
        title: "With Tuvi, your info stays perfectly in sync across every directory.",
        visual: "listings-synced",
      },
      {
        layout: "half",
        theme: "beige",
        label: "SEO-critical details",
        title: "We fix the tiny details that hurt your Google rankings.",
        visual: "address-fix-cards",
      },
      {
        layout: "half",
        theme: "dark",
        label: "Fully managed",
        title: "A team of SEO experts that have your back.",
        visual: "listings-experts-photo",
      },
    ],
  },
  faq: {
    items: [
      {
        question: "What is listings management and why does it matter?",
        answer:
          "Listings management keeps your restaurant's name, address, phone, hours, and website accurate across Google, social, and directories. Consistent details boost SEO and help customers find the right location every time.",
      },
      {
        question: "Which platforms does Tuvi manage my listings on?",
        answer:
          "We sync across major platforms customers use — including Google, Facebook, Yelp, TripAdvisor, DoorDash, Uber Eats, and more — so your info stays consistent everywhere.",
      },
      {
        question: "Couldn't I just update my listings myself?",
        answer:
          "You could, but it's easy for details to drift across dozens of sites. Tuvi monitors and updates them for you so rankings and customer trust don't slip from small mistakes.",
      },
    ],
  },
};
