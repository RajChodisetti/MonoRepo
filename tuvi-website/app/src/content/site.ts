export const siteContent = {
  brand: {
    name: "Tuvi Solutions",
    tagline: "We build software that grows businesses.",
    email: "contact@tuvisolutions.com",
  },

  nav: [
    { href: "/#about", label: "Why Tuvi" },
    { href: "/#testimonials", label: "Results" },
    { href: "/#guarantee", label: "Guarantee" },
    { href: "/#contact", label: "Contact" },
  ],

  servicesNav: {
    label: "Services",
    items: [
      {
        href: "/services/restaurants",
        label: "Restaurants",
        description: "QR ordering, rewards & guest growth — watch the live demos.",
      },
    ],
  },

  hero: {
    eyebrow: "Websites · Apps · AI — built for you",
    headline: ["We build websites, apps & AI", "that grow your business."],
    subcopy:
      "Tuvi is a software studio for ambitious businesses. Bring us the outcome you want — we design, build, and launch the product that gets you there. No jargon, no agency runaround.",
    primaryCta: "Start your project",
    secondaryCta: "See a live demo",
    secondaryHref: "/services/restaurants",
    trust: "4.7/5 from 30+ clients",
    note: "Free consultation. No commitment.",
  },

  stats: [
    { value: 10, suffix: "+", label: "Years building software" },
    { value: 50, suffix: "+", label: "Products shipped" },
    { value: 3, suffix: "x", label: "Average client growth" },
    { value: 100, suffix: "%", label: "Client satisfaction" },
  ],

  about: {
    id: "about",
    eyebrow: "Why Tuvi",
    title: "Small team. Serious engineering.",
    paragraphs: [
      "Tuvi means strength — and that's the bar for everything we ship. Robust architecture, sharp design, and software that holds up when your business takes off.",
      "From your first landing page to AI that answers your phones, we build it, launch it, and stay until it works.",
    ],
    highlights: [
      "Custom websites & web apps",
      "AI assistants & voice agents",
      "Mobile apps for iOS & Android",
      "Data, dashboards & integrations",
    ],
  },

  testimonials: {
    id: "testimonials",
    eyebrow: "Results",
    title: "In our clients' words.",
    items: [
      {
        quote:
          "Tuvi completely overhauled our legacy systems. Their data-driven approach to our system integrations didn't just solve our immediate issues; it gave us the capability to scale our operations by 3x. True professionals.",
        author: "Director of Operations",
        company: "Tamara Agro",
        initial: "T",
      },
      {
        quote:
          "The $1,000 risk-free guarantee showed us they meant business. The work they delivered in that initial phase was so robust and impressive, hiring them for the full build was the easiest decision we've made.",
        author: "Founder",
        company: "Nexus Logistics",
        initial: "N",
      },
      {
        quote:
          "When you are dealing with complex machine learning models, you need architecture that won't break. Tuvi Solutions delivered an AI infrastructure that is both powerful and remarkably easy for our team to use.",
        author: "VP of Engineering",
        company: "Vertex Health",
        initial: "V",
      },
    ],
  },

  guarantee: {
    id: "guarantee",
    eyebrow: "Zero risk",
    title: "Your first $1,000 of work is free.",
    description:
      "We'd rather prove it than pitch it. We complete the first $1,000 of development at no cost — if you don't love the direction, walk away and pay nothing.",
    pillars: [
      {
        title: "Watch it take shape",
        description: "See real, working progress before you spend a cent.",
      },
      {
        title: "Walk away anytime",
        description: "Not feeling it? You owe us nothing. No fine print.",
      },
      {
        title: "Scale when ready",
        description: "Payments only start once you're convinced.",
      },
    ],
    cta: "Claim your free build",
  },

  contact: {
    id: "contact",
    eyebrow: "Let's talk",
    title: "Tell us what you want to build.",
    description:
      "Book a call, talk to our AI assistant, or get a callback in minutes. No sales pitch — just a clear plan you can keep.",
    primaryCta: "Book a free consultation",
  },

  footer: {
    legal: [
      { label: "Privacy Policy", href: "#" },
      { label: "Terms of Service", href: "#" },
      { label: "Contact", href: "/#contact" },
    ],
  },
} as const;

export type SiteContent = typeof siteContent;
