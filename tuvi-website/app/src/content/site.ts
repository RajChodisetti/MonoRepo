export const siteContent = {
  brand: {
    name: "Tuvi Solutions",
    tagline: "Engineered for Strength. Built for Scale.",
    email: "contact@tuvisolutions.com",
  },

  nav: [
    { href: "#about", label: "About" },
    { href: "#testimonials", label: "Clients" },
    { href: "#guarantee", label: "Guarantee" },
    { href: "#team", label: "Team" },
    { href: "#contact", label: "Contact" },
  ],

  hero: {
    eyebrow: "Custom Software Solutions",
    headline: ["Engineered for Strength.", "Built for Scale."],
    subcopy:
      "We deliver custom software development solutions that streamline your business processes, reduce operational costs, and drive measurable growth through smart technology.",
    primaryCta: "Schedule a Free Consultation",
    secondaryCta: "Explore our approach",
    secondaryHref: "#about",
    trust: "4.7 out of 5 (30+ clients)",
    note: "No commitment required.",
  },

  stats: [
    { value: 10, suffix: "+", label: "Years Experience" },
    { value: 50, suffix: "+", label: "Projects Delivered" },
    { value: 3, suffix: "x", label: "Average Growth" },
    { value: 100, suffix: "%", label: "Client Satisfaction" },
  ],

  about: {
    id: "about",
    eyebrow: "What does Tuvi mean?",
    title: "Strength and Capability.",
    paragraphs: [
      "Tuvi is built on the concept of unyielding strength. We engineer robust software architectures and provide the deep, data-driven consulting needed to scale your operations.",
      "From complex system integrations to overarching business strategy, we build solutions strong enough to handle your most critical challenges.",
    ],
    highlights: [
      "Custom software architecture",
      "AI & machine learning systems",
      "Cross-platform app development",
      "Data-driven business strategy",
    ],
  },

  testimonials: {
    id: "testimonials",
    eyebrow: "What Our Clients Say",
    title: "Real results from real partnerships",
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
    eyebrow: "100% Risk-Free",
    title: "The Tuvi Confidence Guarantee",
    description:
      "We believe in proving our capability before asking for your commitment. We will complete the first $1,000 of product development work absolutely free.",
    pillars: [
      {
        title: "See Progress First",
        description: "See tangible progress before you pay a dime.",
      },
      {
        title: "No Risk",
        description: "If you don't love the direction, walk away. You pay nothing.",
      },
      {
        title: "Start Small",
        description: "Standard payments only begin after the initial phase.",
      },
    ],
    cta: "Claim Your Risk-Free Trial",
  },

  team: {
    id: "team",
    eyebrow: "Meet the Leadership Team",
    title: "Elite expertise meets authentic partnership",
    members: [
      {
        name: "Sri Raj",
        role: "CEO",
        initials: "SR",
        bio: "Over 10+ years of elite experience in software development, AI/Machine Learning, and cross-platform app development. Having engineered solutions for global giants like Bank of America alongside dynamic SMBs, Sri bridges complex technical systems and practical, revenue-driving business strategies.",
      },
    ],
  },

  contact: {
    id: "contact",
    eyebrow: "Let's build",
    title: "Ready to engineer your next phase of growth?",
    description:
      "Let's discuss your challenges. No sales pitch, just a deep dive into data-driven solutions.",
    primaryCta: "Schedule a Free Consultation",
  },

  footer: {
    legal: [
      { label: "Privacy Policy", href: "#" },
      { label: "Terms of Service", href: "#" },
      { label: "Contact", href: "#contact" },
    ],
  },
} as const;

export type SiteContent = typeof siteContent;
