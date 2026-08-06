export type CaseStudy = {
  slug: string;
  name: string;
  restaurant: string;
  role: string;
  metricValue: string;
  metricDescription: string;
  imageUrl: string;
  location: string;
  services: string[];
  summary: string;
  challenge: string;
  approach: string[];
  results: { value: string; label: string }[];
  quote: string;
};

/** Fictional demo venues only — not real businesses. */
export const caseStudies: CaseStudy[] = [
  {
    slug: "quillnest-kitchen",
    name: "Riley Quinn",
    restaurant: "Quillnest Kitchen",
    role: "Owner",
    metricValue: "+$72,000",
    metricDescription: "Online sales",
    imageUrl:
      "https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?auto=format&fit=crop&w=800&q=80",
    location: "Melbourne, VIC",
    services: ["Website", "Online ordering", "SEO", "Loyalty"],
    summary:
      "Quillnest Kitchen moved guests off marketplaces and into a branded ordering flow — lifting direct online sales without raising headcount.",
    challenge:
      "Most online orders sat on third-party apps. Fees ate margin, guest data never reached the restaurant, and weeknight covers stayed soft.",
    approach: [
      "Launched a fast Tuvi website with commission-free online ordering under the Quillnest brand.",
      "Synced Google listings and menu SEO so nearby searches pointed to the restaurant’s own site.",
      "Turned one-time buyers into regulars with loyalty stamps and timed email offers for slow midweeks.",
    ],
    results: [
      { value: "+$72,000", label: "Online sales lift" },
      { value: "31%", label: "Fewer marketplace orders" },
      { value: "2.1x", label: "Repeat order rate" },
    ],
    quote:
      "We finally own the guest relationship. Orders come through our site, and Tuvi keeps the quieter nights busier.",
  },
  {
    slug: "orzo-vale-kouzina",
    name: "Jordan Hale",
    restaurant: "Orzo Vale Kouzina",
    role: "Owner",
    metricValue: "+$40,000",
    metricDescription: "Monthly online sales",
    imageUrl:
      "https://images.unsplash.com/photo-1560250097-0b93528c311a?auto=format&fit=crop&w=800&q=80",
    location: "Sydney, NSW",
    services: ["Online ordering", "Upsells", "Email & SMS", "Analytics"],
    summary:
      "Orzo Vale Kouzina rebuilt first-party ordering and used smart upsells plus guest messaging to push monthly online sales past $40K.",
    challenge:
      "The old site looked dated, checkout dropped guests, and there was no way to remarket to people who had already ordered once.",
    approach: [
      "Replaced the legacy site with Tuvi ordering tuned for mobile checkout and clear menu photos.",
      "Added guided upsells at cart so add-ons became part of every ticket.",
      "Used email and SMS after first purchase to fill lunch gaps and promote weekly specials.",
    ],
    results: [
      { value: "+$40,000", label: "Monthly online sales" },
      { value: "+18%", label: "Average order value" },
      { value: "44%", label: "Orders from returning guests" },
    ],
    quote:
      "Online used to feel like a side channel. Now it is a reliable revenue line we can see and grow every week.",
  },
  {
    slug: "nonna-parcel-kitchen",
    name: "Avery Knox",
    restaurant: "Nonna Parcel Kitchen",
    role: "Owner",
    metricValue: "+180%",
    metricDescription: "Growth in online catering",
    imageUrl:
      "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=800&q=80",
    location: "Brisbane, QLD",
    services: ["Catering", "Website", "Campaigns", "Reviews"],
    summary:
      "Nonna Parcel Kitchen productized catering online — inquiries became bookable packages, and catering revenue more than doubled.",
    challenge:
      "Catering lived in email threads and phone tag. Teams could not browse packages, and big orders were easy to lose to competitors with clearer sites.",
    approach: [
      "Published catering packages with clear minimums, lead times, and add-ons on the Tuvi site.",
      "Captured leads into campaigns so follow-ups happened automatically after a quote request.",
      "Asked happy catering clients for Google reviews to strengthen local trust.",
    ],
    results: [
      { value: "+180%", label: "Online catering growth" },
      { value: "3.2x", label: "Qualified catering leads" },
      { value: "4.9★", label: "Average new reviews" },
    ],
    quote:
      "Catering stopped being chaos. Guests pick a package online and we just execute — Tuvi made that simple.",
  },
  {
    slug: "brightkiln-kitchen",
    name: "Sam Rivera",
    restaurant: "Brightkiln Kitchen",
    role: "Owner",
    metricValue: "+100%",
    metricDescription: "Growth in Google bookings",
    imageUrl:
      "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?auto=format&fit=crop&w=800&q=80",
    location: "Perth, WA",
    services: ["Restaurant SEO", "Listings", "Reviews", "Menu"],
    summary:
      "Brightkiln Kitchen doubled Google bookings by fixing listings, reviews, and SEO so nearby diners found them first.",
    challenge:
      "Search results showed incomplete hours, thin menus, and competitor sites ranking above Brightkiln for local dish queries.",
    approach: [
      "Cleaned Google Business and directory listings so hours, photos, and CTAs stayed accurate.",
      "Published SEO-ready menu and location pages that matched how guests actually searched.",
      "Built a steady review rhythm so fresh ratings kept the listing competitive.",
    ],
    results: [
      { value: "+100%", label: "Google bookings" },
      { value: "2x", label: "Local search clicks" },
      { value: "+38%", label: "Weeknight covers" },
    ],
    quote:
      "Weeknights used to be quiet. Once we showed up properly on Google, the tables started filling themselves.",
  },
  {
    slug: "steamrail-noodles",
    name: "Priya Mehta",
    restaurant: "Steamrail Noodles",
    role: "Owner",
    metricValue: "+65%",
    metricDescription: "Repeat guest bookings",
    imageUrl:
      "https://images.unsplash.com/photo-1580489944761-15a19d654956?auto=format&fit=crop&w=800&q=80",
    location: "Adelaide, SA",
    services: ["Loyalty", "Push & SMS", "Owner app", "Ordering"],
    summary:
      "Steamrail Noodles used loyalty and timed messaging to lift repeat bookings 65% while keeping every order on their own channels.",
    challenge:
      "Guests loved the food once — then disappeared. There was no owned list, no loyalty loop, and no signal when regulars went quiet.",
    approach: [
      "Launched a simple points loyalty program tied to first-party ordering.",
      "Sent push and SMS when favourite dishes returned or slow hours needed a nudge.",
      "Gave the owner live views of repeats and top guests in the Tuvi owner app.",
    ],
    results: [
      { value: "+65%", label: "Repeat guest bookings" },
      { value: "4,200+", label: "Loyalty members" },
      { value: "+$28,000", label: "Monthly first-party sales" },
    ],
    quote:
      "Our regulars finally have a reason to come back through us — not through another app’s homepage.",
  },
];

export function getCaseStudy(slug: string): CaseStudy | undefined {
  return caseStudies.find((study) => study.slug === slug);
}

export function caseStudyPath(slug: string): string {
  return `/resources/case-studies/${slug}`;
}
