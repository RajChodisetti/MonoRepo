export const siteContent = {
  brand: {
    name: "Tuvi Solutions",
    tagline: "We build software that grows businesses.",
    email: "contact@tuvisolutions.com",
  },

  oauthApp: {
    name: "tuvi",
    eyebrow: "Google Workspace application",
    title: "Reviewed email delivery with one narrow permission.",
    description:
      "tuvi is a private, company-operated application from Tuvi Solutions. Authorized Tuvi team members use it to send individually reviewed business email from their Google Workspace mailboxes through the Gmail API over HTTPS.",
    dataUse:
      "tuvi requests gmail.send on authorized sending mailboxes. A separate inbound mailbox may use gmail.readonly only to capture replies to outreach plus-addresses. Sending mailboxes cannot read inbox messages, contacts, attachments, message history, or Google Drive files.",
    access:
      "Mailbox access is limited to authorized Tuvi Solutions owners and administrators. There is no public signup.",
    metadataDescription:
      "tuvi uses Google OAuth and the Gmail API to send individually reviewed email from authorized Tuvi Solutions Google Workspace mailboxes.",
  },

  nav: [
    { href: "/#about", label: "Why Tuvi" },
    { href: "/#approach", label: "Our standard" },
    { href: "/#guarantee", label: "First build" },
    { href: "/#contact", label: "Contact" },
  ],

  servicesNav: {
    label: "Services",
    items: [
      {
        href: "/#services",
        label: "AI & software",
        description: "AI systems, websites, apps & connected integrations.",
      },
      {
        href: "/services/restaurants",
        label: "Restaurants",
        description: "QR ordering, rewards & guest growth — watch the live demos.",
      },
    ],
  },

  hero: {
    eyebrow: "AI & custom software for every stage",
    headline: ["The right", "Software."],
    subcopy:
      "Tuvi Solutions designs and builds websites, customer and team apps, AI assistants, and connected automation — from first idea through launch and growth.",
    primaryCta: "Start your project",
    secondaryCta: "See a live demo",
    secondaryHref: "/services/restaurants",
    trust: "One team from strategy to launch",
    note: "Free consultation · clear next steps",
  },

  about: {
    id: "about",
    eyebrow: "Why Tuvi",
    title: "Small team. Serious engineering.",
    paragraphs: [
      "Tuvi means strength — and that's the bar for everything we ship. Robust architecture, sharp design, and software that holds up when your business takes off.",
      "From your first landing page to AI that answers your phones, we shape the scope, build it, launch it, and support the agreed next stage.",
    ],
    highlights: [
      "Custom websites & web apps",
      "AI assistants & voice agents",
      "Mobile apps for iOS & Android",
      "Data, dashboards & integrations",
    ],
  },

  testimonials: {
    id: "approach",
    eyebrow: "Our standard",
    title: "Small team. High bar. No hand-offs.",
    items: [
      {
        title: "Outcome before output",
        description:
          "We begin with the business result, then choose the smallest product that can move it.",
        number: "01",
      },
      {
        title: "Working software, early",
        description:
          "You review real screens and real flows throughout the build—not a reveal at the end.",
        number: "02",
      },
      {
        title: "Built beyond launch",
        description:
          "Clean foundations, measurable behaviour, and a team that stays accountable after release.",
        number: "03",
      },
    ],
  },

  guarantee: {
    id: "guarantee",
    eyebrow: "A confident start",
    title: "Your first focused build can be on us.",
    description:
      "For qualifying new projects, we agree a focused first milestone valued up to $1,000 and complete it at no charge. Eligibility and scope are confirmed in writing before work begins.",
    pillars: [
      {
        title: "A defined milestone",
        description: "We agree the outcome, scope, and review point before we begin.",
      },
      {
        title: "Working progress",
        description: "You see something tangible—not a strategy deck or speculative promise.",
      },
      {
        title: "Continue by choice",
        description: "A larger engagement starts only after both sides agree on the next scope.",
      },
    ],
    cta: "Discuss your first milestone",
  },

  contact: {
    id: "contact",
    eyebrow: "Let's talk",
    title: "Tell us what you want to build.",
    description:
      "Book a call, talk to our AI assistant, or request a callback. No agency theatre—just a useful conversation and a clear next step.",
    primaryCta: "Book a free consultation",
  },

  footer: {
    legal: [
      { label: "tuvi Google Workspace app", href: "/google-workspace" },
      { label: "Privacy Policy", href: "/privacy" },
      { label: "Terms of Service", href: "/terms" },
      { label: "Contact", href: "/#contact" },
    ],
  },
} as const;

export type SiteContent = typeof siteContent;
