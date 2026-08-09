export type ResourceSection = {
  heading: string;
  body: string;
  bullets?: string[];
  imageSrc?: string;
  imageAlt?: string;
};

export type ResourceGuide = {
  slug: string;
  eyebrow: string;
  title: string;
  description: string;
  readTime: string;
  publishedLabel: string;
  heroImage: string;
  heroAlt: string;
  takeaways: string[];
  quote?: { text: string; attribution: string };
  sections: ResourceSection[];
  relatedHref?: string;
  relatedLabel?: string;
};

/** Original Tuvi blog guides — AI-written copy + generated imagery (not Owner.com). */
export const resourceGuides: ResourceGuide[] = [
  {
    slug: "marketing-guide",
    eyebrow: "Blog · Marketing",
    title: "The restaurant marketing guide for owners who want guests — not marketplace rent",
    description:
      "A field guide to discovery, offers, and retention that keeps every click on your brand. Built for independent restaurants that are done paying for their own customers twice.",
    readTime: "11 min read",
    publishedLabel: "Updated Aug 2026",
    heroImage: "/resources/resource-marketing-hero.png",
    heroAlt: "Warm evening dining room ready for service",
    takeaways: [
      "Own the channel before you buy the traffic",
      "Time offers to real meal windows, not random blasts",
      "Judge campaigns by covers and fee savings — not vanity clicks",
    ],
    quote: {
      text: "Marketing only works when the guest lands somewhere you control. Otherwise you’re renting attention you already paid for.",
      attribution: "Tuvi operator playbook",
    },
    relatedHref: "/product/campaigns",
    relatedLabel: "Explore marketing campaigns",
    sections: [
      {
        heading: "Why most restaurant marketing leaks money",
        body: "Boosted posts and marketplace promos feel busy — but they often send hungry people into someone else’s app. You pay to acquire the guest, then pay again when they reorder. First-party marketing flips that: website, Google presence, email, SMS, and a branded app become the funnel you own.",
        bullets: [
          "Stop measuring success by impressions alone",
          "Route every campaign to order-on-your-site (or book-a-table)",
          "Capture the guest once, market to them forever",
        ],
        imageSrc: "/resources/resource-inline-dining.png",
        imageAlt: "Plated pasta on a restaurant table",
      },
      {
        heading: "Build the owned stack first",
        body: "Before another ad dollar leaves the bank, make sure the landing experience is ready: fast mobile site, clear menu, commission-free checkout, and a way to join loyalty. Tuvi ships that foundation so campaigns have somewhere useful to land.",
      },
      {
        heading: "Match the message to the moment",
        body: "Lunch rush, quiet Tuesdays, and catering season each need different creative. A midday SMS about a 20-minute pickup window beats a generic weekend blast. Tuvi campaigns help you schedule email, SMS, and push around when guests are actually deciding where to eat.",
        imageSrc: "/resources/resource-inline-street.png",
        imageAlt: "City food street at dusk",
      },
      {
        heading: "A simple weekly rhythm",
        body: "Monday: check last week’s first-party orders and repeat rate. Wednesday: one offer for your softest daypart. Friday: a reminder for regulars with points to spend. Keep the cadence human — restaurants aren’t SaaS drip sequences.",
      },
      {
        heading: "What to measure instead of vanity metrics",
        body: "Track first-party order volume, average ticket, repeat rate within 30 days, and marketplace fee avoided. If a campaign doesn’t move those numbers in two cycles, kill it and try a sharper offer.",
      },
    ],
  },
  {
    slug: "seo-guide",
    eyebrow: "Blog · SEO",
    title: "SEO for restaurants: how to show up when someone searches “near me”",
    description:
      "Local search is the new front door. Here’s how listings, menus, and reviews work together so nearby diners find you — then order or book under your brand.",
    readTime: "10 min read",
    publishedLabel: "Updated Aug 2026",
    heroImage: "/resources/resource-seo-hero.png",
    heroAlt: "Phone on a cafe table during a local search",
    takeaways: [
      "“Near me” intent is the highest-value traffic you can earn",
      "Your site must answer the query — not just look pretty",
      "Fresh reviews keep rankings and trust compounding",
    ],
    quote: {
      text: "If Google can’t tell what you serve and where you are, the guest will pick whoever shows up first with a clear photo and a button.",
      attribution: "Tuvi local SEO notes",
    },
    relatedHref: "/product/seo",
    relatedLabel: "See restaurant SEO",
    sections: [
      {
        heading: "Win the searches that already have intent",
        body: "Restaurant SEO isn’t about ranking for “best food blog.” It’s dish + neighbourhood: noodles near Southbank, brunch in Fremantle, late-night pizza. Keep Google Business hours, categories, and photos accurate so those queries have somewhere to go.",
        imageSrc: "/resources/resource-inline-street.png",
        imageAlt: "Evening street dining corridor",
      },
      {
        heading: "Make your website finish the job",
        body: "Listings get the click; your site closes it. Location pages, structured menus, and obvious Order / Reserve CTAs help both guests and search engines. Tuvi builds SEO-ready pages so “near me” traffic doesn’t bounce to a PDF menu.",
        bullets: [
          "One page per location with unique copy",
          "Menu items as real pages, not images only",
          "Fast mobile load — hungry people don’t wait",
        ],
      },
      {
        heading: "Reviews are ranking fuel",
        body: "Volume, recency, and response quality all matter. A quiet profile looks abandoned. Pair listings hygiene with a review engine so happy diners leave feedback while the meal is still warm.",
        imageSrc: "/resources/resource-inline-dining.png",
        imageAlt: "Finished plate ready for a guest review moment",
      },
      {
        heading: "A 30-day SEO sprint for busy owners",
        body: "Week 1: fix hours, categories, and cover photos. Week 2: publish or refresh menu pages. Week 3: respond to every review. Week 4: check Search Console / insights for queries you almost rank for — and tighten those pages.",
      },
    ],
  },
  {
    slug: "email-marketing",
    eyebrow: "Blog · Email & SMS",
    title: "Restaurant email & SMS that bring guests back (without sounding like spam)",
    description:
      "Short messages, sharp offers, and first-party data — the retention loop independent restaurants actually stick with.",
    readTime: "9 min read",
    publishedLabel: "Updated Aug 2026",
    heroImage: "/resources/resource-email-hero.png",
    heroAlt: "Laptop and notes on a kitchen counter",
    takeaways: [
      "One offer, one link, one reason to open",
      "Segment by behaviour — not just “everyone who ordered once”",
      "SMS converts fast; respect quiet hours harder",
    ],
    relatedHref: "/product/email",
    relatedLabel: "Explore email & SMS",
    sections: [
      {
        heading: "Write like a host, not a billboard",
        body: "Guests skim between school runs and shift changes. Lead with the outcome (“Free garlic bread on your next midweek order”) and send them straight to first-party checkout. Long newsletters with six CTAs lose the plot.",
        imageSrc: "/resources/resource-inline-dining.png",
        imageAlt: "Inviting plated dish for a midweek offer",
      },
      {
        heading: "Segment with order history you own",
        body: "First-timers, lapsed regulars, and catering buyers need different nudges. Because Tuvi runs on your channels, those segments come from real tickets — not a rented marketplace list.",
        bullets: [
          "Welcome series after first online order",
          "Win-back after 21–30 quiet days",
          "Catering follow-up with packages, not coupons",
        ],
      },
      {
        heading: "SMS rules that keep you welcome",
        body: "Text for time-sensitive moments: soft Tuesday, weather wipeout, holiday special. Avoid late nights. Always make opt-out obvious. Earn the right to message by being useful.",
      },
      {
        heading: "Pair messaging with loyalty",
        body: "Points and stamps give people a reason to open the next email. “You’re two visits from a free side” beats another blast about your vibe.",
      },
    ],
  },
  {
    slug: "mobile-app",
    eyebrow: "Blog · App",
    title: "Why your restaurant needs a branded mobile app (and what guests actually use)",
    description:
      "Reorder, rewards, and push — the everyday features that keep tickets on your channels instead of a marketplace homepage.",
    readTime: "9 min read",
    publishedLabel: "Updated Aug 2026",
    heroImage: "/resources/resource-app-hero.png",
    heroAlt: "Guest ordering from a phone over a burger",
    takeaways: [
      "Convenience beats novelty — reorder is the killer feature",
      "Loyalty only compounds if it lives on your brand",
      "Launch lean: ordering + rewards, then deepen",
    ],
    quote: {
      text: "Guests don’t want another logo on their home screen unless it saves them taps. Make reorder and rewards undeniable.",
      attribution: "Tuvi product notes",
    },
    relatedHref: "/product/app",
    relatedLabel: "See branded restaurant app",
    sections: [
      {
        heading: "What a restaurant app is really for",
        body: "It’s not a mini Instagram. It’s the fastest path from craving to confirmed order under your brand — with points that make the next visit inevitable.",
        imageSrc: "/resources/resource-inline-dining.png",
        imageAlt: "Food ready for pickup after an app order",
      },
      {
        heading: "Features guests touch weekly",
        body: "Saved favourites, one-tap reorder, live order status, and a simple loyalty balance. Fancy AR menus can wait. Reliability cannot.",
        bullets: [
          "Push for “your usual is ready to reorder”",
          "Clear pickup vs delivery choice",
          "Rewards visible before checkout",
        ],
      },
      {
        heading: "How Tuvi approaches branded apps",
        body: "We design the experience around first-party ordering and retention — not marketplace clones with your logo slapped on. Your photography, your menu, your guest list.",
      },
    ],
  },
  {
    slug: "ordering-systems",
    eyebrow: "Blog · Ordering",
    title: "Online ordering systems: how to choose one that protects margin",
    description:
      "Checkout speed, kitchen tickets, and fee math — the checklist operators use before locking in a stack.",
    readTime: "10 min read",
    publishedLabel: "Updated Aug 2026",
    heroImage: "/resources/resource-ordering-hero.png",
    heroAlt: "Takeaway ready on a modern kitchen pass",
    takeaways: [
      "Own the checkout or own the fee forever",
      "Pretty menus fail if the pass is chaos",
      "Upsells should add taps, not friction",
    ],
    relatedHref: "/product/ordering",
    relatedLabel: "Explore online ordering",
    sections: [
      {
        heading: "Commission-free is the baseline",
        body: "If every digital order pays rent to a marketplace, online growth shrinks your margin. First-party ordering on your site or app keeps guest data and dollars with you.",
        imageSrc: "/resources/resource-inline-street.png",
        imageAlt: "Busy evening service outside the restaurant",
      },
      {
        heading: "Kitchen-ready by design",
        body: "Modifiers, allergens, and fire times need to arrive as clear tickets. Pair ordering with a kitchen tablet so the line isn’t decoding screenshots from three apps.",
        bullets: [
          "Item notes that cooks can scan in one glance",
          "Throttling when the board is slammed",
          "POS sync so reports match reality",
        ],
      },
      {
        heading: "Upsells without slowing mobile checkout",
        body: "Guided add-ons at cart lift ticket size when they’re relevant — drinks with spicy mains, garlic bread with pasta. Keep pay under a few taps.",
      },
      {
        heading: "Questions to ask any vendor",
        body: "Who owns the guest list? What’s the fee on a $40 ticket? How do refunds and partial refunds work? Can catering live on the same stack? Tuvi’s answers are built for independents who want one system, not five logins.",
      },
    ],
  },
  {
    slug: "website-builders",
    eyebrow: "Blog · Website",
    title: "Restaurant website builders: what actually matters in 2026",
    description:
      "Speed, menu clarity, SEO, and a direct path to order — why generic site builders leave money on the table for restaurants.",
    readTime: "9 min read",
    publishedLabel: "Updated Aug 2026",
    heroImage: "/resources/resource-website-hero.png",
    heroAlt: "Laptop on a counter with a restaurant site mockup",
    takeaways: [
      "Mobile-first isn’t optional — it’s most of your traffic",
      "Menus as pages beat PDF uploads",
      "Ordering and SEO should ship together",
    ],
    relatedHref: "/restaurant-website-ai",
    relatedLabel: "See restaurant websites",
    sections: [
      {
        heading: "Stop shipping brochure sites",
        body: "A beautiful homepage that dies on “Order” helps nobody. Guests arrive from Google with intent — give them hours, menu, and checkout in under ten seconds.",
        imageSrc: "/resources/resource-inline-dining.png",
        imageAlt: "Dish photography ready for a website menu",
      },
      {
        heading: "Menu architecture that ranks and converts",
        body: "Structured menu pages help search engines understand what you serve and help guests browse without pinch-zooming a photo of a chalkboard.",
        bullets: [
          "Categories guests recognise (not internal station names)",
          "Photos that load fast on cellular",
          "Allergens and modifiers without clutter",
        ],
      },
      {
        heading: "Why Tuvi websites are restaurant-native",
        body: "Hours, multi-location, catering, and commission-free ordering aren’t plugins bolted on later. They’re part of the same stack that powers SEO and retention.",
      },
    ],
  },
  {
    slug: "blog",
    eyebrow: "Blog",
    title: "Tuvi blog: notes for operators growing direct sales",
    description:
      "Longer reads on discovery, ordering, and retention — written for owners who want practical plays, not buzzwords.",
    readTime: "Library",
    publishedLabel: "Updated weekly",
    heroImage: "/resources/resource-blog-hero.png",
    heroAlt: "Editorial reading setup for restaurant operators",
    takeaways: [
      "Start with the six core guides below",
      "Pair every article with a product path you can demo",
      "Case studies show the same ideas in motion",
    ],
    sections: [
      {
        heading: "What’s live in the library",
        body: "Marketing, SEO, email & SMS, branded apps, ordering systems, and website builders — each guide is written for busy operators and links into the Tuvi product that executes the play.",
        imageSrc: "/resources/resource-inline-street.png",
        imageAlt: "City evening atmosphere for operator reading",
      },
      {
        heading: "Prefer stories over checklists?",
        body: "Visit case studies for fictional demo venues that put these ideas into practice — Quillnest Kitchen, Orzo Vale Kouzina, Brightkiln Kitchen, and more.",
      },
    ],
  },
  {
    slug: "help",
    eyebrow: "Help",
    title: "Tuvi help center: get unstuck fast",
    description:
      "Setup, ordering, and guest messaging answers — plus how to reach a human when you need one.",
    readTime: "Support",
    publishedLabel: "Always on",
    heroImage: "/resources/resource-help-hero.png",
    heroAlt: "Operator reviewing a tablet during service prep",
    takeaways: [
      "Most launches start with website + menu + ordering",
      "Demo calls map the stack to your locations",
      "Email us with restaurant name + city for fastest help",
    ],
    relatedHref: "/book",
    relatedLabel: "Get a free demo",
    sections: [
      {
        heading: "Getting started with Tuvi",
        body: "We launch your online foundation first — site, menu, checkout — then layer SEO, listings, and retention. Bring photos, hours, and your busiest dayparts to the first call.",
        imageSrc: "/resources/resource-inline-dining.png",
        imageAlt: "Service-ready plate during launch week",
      },
      {
        heading: "Talk to a person",
        body: "Email contact@tuvisolutions.com. Include your restaurant name, city, and what you’re trying to fix (discovery, orders, or repeats). We reply within one business day.",
      },
    ],
  },
  {
    slug: "grader",
    eyebrow: "Tool",
    title: "Restaurant grader: a fast read on your online presence",
    description:
      "See where you stand on search, site strength, and reviews — then fix the gaps that quietly cost covers.",
    readTime: "Free check",
    publishedLabel: "Try on homepage",
    heroImage: "/resources/resource-grader-hero.png",
    heroAlt: "Dining room mood for an online presence check",
    takeaways: [
      "Visibility gaps show up before sales gaps",
      "Fix listings and site speed before buying ads",
      "Use the homepage report for a live snapshot",
    ],
    relatedHref: "/",
    relatedLabel: "Run a free report",
    sections: [
      {
        heading: "What the grader looks at",
        body: "Google visibility signals, listings hygiene, website clarity, and review momentum — the same pillars behind Tuvi’s live report experience on the homepage.",
        imageSrc: "/resources/resource-inline-street.png",
        imageAlt: "Street-level discovery where guests search",
      },
      {
        heading: "How to use the results",
        body: "Prioritise anything that blocks a hungry guest: wrong hours, thin menus, slow mobile pages, stale reviews. Then book a demo if you want Tuvi to run the fix.",
      },
    ],
  },
  {
    slug: "support",
    eyebrow: "Support",
    title: "Tuvi support center",
    description:
      "For restaurants live on Tuvi — go-live questions, ordering quirks, and campaign help from people who know the floor.",
    readTime: "Contact",
    publishedLabel: "Business days",
    heroImage: "/resources/resource-support-hero.png",
    heroAlt: "Friendly support conversation over coffee",
    takeaways: [
      "Email is fastest with restaurant name + issue",
      "Demos are available if you want a live walkthrough",
      "We optimise for operators, not ticket bots",
    ],
    relatedHref: "/book",
    relatedLabel: "Book a free demo",
    sections: [
      {
        heading: "How to reach us",
        body: "Write contact@tuvisolutions.com with your restaurant name and a short description of what broke or what you want to improve. Screenshots help.",
        imageSrc: "/resources/resource-inline-dining.png",
        imageAlt: "Calm table setting while you wait for support",
      },
      {
        heading: "Prefer screenshare?",
        body: "Book a free demo and we’ll walk website, SEO, ordering, and retention on your locations — useful even if you’re already a customer scoping the next module.",
      },
    ],
  },
];

export function getResourceGuide(slug: string): ResourceGuide | undefined {
  return resourceGuides.find((guide) => guide.slug === slug);
}

export const resourceGuideSlugs = resourceGuides.map((guide) => guide.slug);

export const blogLibrarySlugs = [
  "marketing-guide",
  "seo-guide",
  "email-marketing",
  "mobile-app",
  "ordering-systems",
  "website-builders",
  "blog",
] as const;
