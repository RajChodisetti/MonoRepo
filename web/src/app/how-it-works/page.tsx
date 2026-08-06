import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import SiteFooter from "@/components/layout/SiteFooter";

export const metadata: Metadata = {
  title: "How it works | Tuvi",
  description:
    "See how Tuvi launches your restaurant online, grows Google discovery, takes first-party orders, and turns guests into regulars — under your brand.",
};

const steps = [
  {
    num: "01",
    title: "Launch your online foundation",
    body: "We stand up a fast website, structured menu, and commission-free ordering under your brand — so guests can find you and checkout without a marketplace tax.",
    detail:
      "Branding, hours, locations, and photos go live together. Kitchen tickets and owner visibility are wired from day one.",
    image: "/resources/resource-website-hero.png",
    imageAlt: "Restaurant website on a counter laptop",
  },
  {
    num: "02",
    title: "Get found where guests already search",
    body: "SEO, listings hygiene, and reviews work as one system so “near me” searches point to you — not a third-party homepage.",
    detail:
      "We clean Google Business details, publish menu pages that rank, and keep review momentum so your profile stays competitive.",
    image: "/resources/resource-seo-hero.png",
    imageAlt: "Phone on a cafe table during a local search",
  },
  {
    num: "03",
    title: "Take more first-party orders",
    body: "Online ordering, upsells, delivery, catering, and AI phone capture demand on channels you own — with tickets the kitchen can cook from.",
    detail:
      "Guests reorder faster. You keep the data. Margin stays with the restaurant instead of disappearing into fees.",
    image: "/resources/resource-ordering-hero.png",
    imageAlt: "Takeaway ready on a kitchen pass",
  },
  {
    num: "04",
    title: "Bring guests back on purpose",
    body: "Loyalty, email, SMS, and push fire at the right meal windows so one visit becomes a regular — timed to fill soft nights.",
    detail:
      "Campaigns use your order history, not a rented list. Points and offers live on your branded app and site.",
    image: "/resources/resource-app-hero.png",
    imageAlt: "Guest ordering from a phone over a meal",
  },
] as const;

const timeline = [
  {
    when: "Week 1",
    title: "Kickoff & build",
    body: "Brand assets, menu import, locations, and ordering rules. You approve the look; we handle the stack.",
  },
  {
    when: "Week 2",
    title: "Go live online",
    body: "Site + checkout live. Listings cleaned. Staff trained on tickets and the owner app.",
  },
  {
    when: "Week 3–4",
    title: "Discovery & retention",
    body: "SEO pages, review flow, and first campaigns. Soft dayparts get their first timed offers.",
  },
  {
    when: "Ongoing",
    title: "Tuvi keeps optimizing",
    body: "You cook. We keep tightening search, orders, and repeats with clear reporting.",
  },
] as const;

const pillars = [
  {
    title: "Get found online",
    body: "Website, SEO, menu, listings, and reviews — so nearby diners discover you first.",
    href: "/product/seo",
    image: "/resources/resource-marketing-hero.png",
  },
  {
    title: "Take more orders",
    body: "Ordering, delivery, catering, AI phone, and your branded app — commission-free.",
    href: "/product/ordering",
    image: "/resources/resource-ordering-hero.png",
  },
  {
    title: "Bring guests back",
    body: "Campaigns, email & SMS, push, and loyalty that turn one ticket into a habit.",
    href: "/product/loyalty",
    image: "/resources/resource-email-hero.png",
  },
  {
    title: "Run the floor",
    body: "Owner app, kitchen tablet, POS sync, and analytics so service stays clear.",
    href: "/product/owner-app",
    image: "/resources/resource-help-hero.png",
  },
] as const;

const faqs = [
  {
    q: "How long until we’re live?",
    a: "Most restaurants launch the foundation in about two weeks once menus and brand assets are ready. Discovery and retention layers ramp in the weeks after.",
  },
  {
    q: "Do we still need marketplaces?",
    a: "You can keep them if you want — Tuvi’s job is to grow the share of orders that happen on your site and app so fees stop eating the ticket.",
  },
  {
    q: "Who owns the guest data?",
    a: "You do. First-party orders, loyalty, and messaging stay under your brand — not locked inside a marketplace account.",
  },
  {
    q: "What do we need to prepare?",
    a: "Logo, photos, current menu, hours, and a point person for approvals. We’ll guide the rest on the kickoff call.",
  },
] as const;

export default function HowItWorksPage() {
  return (
    <div className="bg-bg">
      <section className="hero-atmosphere relative overflow-hidden px-4 pb-12 pt-14 sm:px-8 sm:pb-16 sm:pt-20 md:px-12">
        <div
          className="pointer-events-none absolute inset-0 hero-grid opacity-30 [mask-image:radial-gradient(40rem_24rem_at_50%_30%,black,transparent)]"
          aria-hidden="true"
        />
        <div className="relative z-10 mx-auto max-w-[920px] text-center">
          <p className="inline-flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">
            <span className="h-1.5 w-1.5 rounded-full bg-accent" aria-hidden="true" />
            How it works
          </p>
          <h1 className="mt-4 font-display text-[clamp(2.2rem,5vw,3.75rem)] font-semibold leading-[1.05] tracking-[-0.03em] text-ink">
            From first Google search to regular guest — under your brand
          </h1>
          <p className="mx-auto mt-4 max-w-[48ch] text-[16px] leading-relaxed text-muted sm:text-[17px]">
            Tuvi is the AI platform restaurants use to get found, take more first-party orders, and
            grow repeats without giving the guest away to a marketplace.
          </p>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
            <Link
              href="/demo"
              className="inline-flex items-center justify-center rounded-full bg-primary px-6 py-3 text-[14px] font-semibold text-bg transition-colors hover:bg-primary-dim"
            >
              Get a free demo
            </Link>
            <Link
              href="#steps"
              className="inline-flex items-center justify-center rounded-full bg-surface px-6 py-3 text-[14px] font-semibold text-ink transition-colors hover:bg-parchment"
            >
              See the 4 steps
            </Link>
          </div>
        </div>
      </section>

      <section className="px-4 pb-6 sm:px-8 md:px-12">
        <div className="relative mx-auto aspect-[16/9] max-w-[1040px] overflow-hidden rounded-[28px] sm:rounded-[36px]">
          <video
            className="absolute inset-0 h-full w-full object-cover"
            autoPlay
            muted
            loop
            playsInline
            preload="metadata"
            aria-label="Guests dining and ordering with Tuvi"
          >
            <source
              src="/hf_20260727_055931_a989648e-ba15-4a67-919d-e2e758e351fe.mp4"
              type="video/mp4"
            />
          </video>
          <div
            className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/50 via-transparent to-transparent"
            aria-hidden
          />
          <p className="absolute bottom-6 left-6 max-w-[28ch] text-[15px] font-semibold text-white sm:bottom-8 sm:left-8 sm:text-[17px]">
            One stack for discovery, orders, and repeats — built for independents.
          </p>
        </div>
      </section>

      <section id="steps" className="scroll-mt-24 px-4 py-14 sm:px-8 sm:py-16 md:px-12">
        <div className="mx-auto max-w-[1040px]">
          <h2 className="max-w-[20ch] font-display text-[clamp(1.75rem,3.4vw,2.6rem)] font-semibold tracking-[-0.03em] text-ink">
            Four steps to growing online the Tuvi way
          </h2>
          <p className="mt-3 max-w-[48ch] text-[15px] leading-relaxed text-muted sm:text-[16px]">
            Each step compounds the last. You don’t pick a random tool — you launch a system that
            keeps guests on your brand.
          </p>

          <div className="mt-12 space-y-14 sm:mt-14 sm:space-y-16">
            {steps.map((step, index) => (
              <article
                key={step.num}
                className={`grid items-center gap-8 lg:grid-cols-2 lg:gap-12 ${
                  index % 2 === 1 ? "lg:[&>*:first-child]:order-2" : ""
                }`}
              >
                <div className="relative aspect-[16/11] overflow-hidden rounded-[24px] sm:rounded-[28px]">
                  <Image
                    src={step.image}
                    alt={step.imageAlt}
                    fill
                    className="object-cover"
                    sizes="(max-width: 1024px) 100vw, 520px"
                  />
                </div>
                <div>
                  <p className="text-[13px] font-semibold tracking-[0.16em] text-primary">
                    {step.num}
                  </p>
                  <h3 className="mt-3 font-display text-[clamp(1.45rem,2.4vw,2rem)] font-semibold leading-snug tracking-[-0.02em] text-ink">
                    {step.title}
                  </h3>
                  <p className="mt-3 text-[15px] leading-relaxed text-muted sm:text-[16px]">
                    {step.body}
                  </p>
                  <p className="mt-3 text-[15px] leading-relaxed text-ink/80 sm:text-[16px]">
                    {step.detail}
                  </p>
                </div>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="border-y border-border bg-white px-4 py-14 sm:px-8 sm:py-16 md:px-12">
        <div className="mx-auto max-w-[1040px]">
          <h2 className="max-w-[18ch] font-display text-[clamp(1.75rem,3.4vw,2.6rem)] font-semibold tracking-[-0.03em] text-ink">
            What the first month looks like
          </h2>
          <div className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {timeline.map((item) => (
              <div
                key={item.when}
                className="rounded-[22px] border border-black/8 bg-bg px-5 py-5"
              >
                <p className="text-[12px] font-semibold uppercase tracking-[0.14em] text-primary">
                  {item.when}
                </p>
                <h3 className="mt-2 text-[17px] font-semibold tracking-[-0.02em] text-ink">
                  {item.title}
                </h3>
                <p className="mt-2 text-[14px] leading-relaxed text-muted">{item.body}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="px-4 py-14 sm:px-8 sm:py-16 md:px-12">
        <div className="mx-auto max-w-[1040px]">
          <h2 className="max-w-[16ch] font-display text-[clamp(1.75rem,3.4vw,2.6rem)] font-semibold tracking-[-0.03em] text-ink">
            One platform. Four growth jobs.
          </h2>
          <p className="mt-3 max-w-[44ch] text-[15px] text-muted sm:text-[16px]">
            Explore the product areas — or book a demo and we’ll map them to your locations.
          </p>
          <div className="mt-10 grid gap-5 sm:grid-cols-2">
            {pillars.map((pillar) => (
              <Link
                key={pillar.title}
                href={pillar.href}
                className="group overflow-hidden rounded-[24px] border border-black/8 bg-white transition-shadow hover:shadow-[0_16px_40px_rgba(0,0,0,0.08)]"
              >
                <div className="relative aspect-[16/9] overflow-hidden">
                  <Image
                    src={pillar.image}
                    alt=""
                    fill
                    className="object-cover transition-transform duration-500 group-hover:scale-[1.03]"
                    sizes="(max-width: 768px) 100vw, 500px"
                  />
                </div>
                <div className="p-5 sm:p-6">
                  <h3 className="font-display text-[1.35rem] font-semibold tracking-[-0.02em] text-ink">
                    {pillar.title}
                  </h3>
                  <p className="mt-2 text-[14px] leading-relaxed text-muted">{pillar.body}</p>
                  <span className="mt-4 inline-flex text-[13px] font-semibold text-primary transition-transform group-hover:translate-x-0.5">
                    Explore →
                  </span>
                </div>
              </Link>
            ))}
          </div>
        </div>
      </section>

      <section className="border-t border-border bg-white px-4 py-14 sm:px-8 sm:py-16 md:px-12">
        <div className="mx-auto grid max-w-[920px] gap-10 lg:grid-cols-[0.9fr_1.1fr] lg:gap-14">
          <div>
            <h2 className="font-display text-[clamp(1.75rem,3vw,2.4rem)] font-semibold tracking-[-0.03em] text-ink">
              Questions owners ask before switching
            </h2>
            <p className="mt-3 text-[15px] leading-relaxed text-muted">
              Straight answers — then a demo if you want it on your menu and locations.
            </p>
          </div>
          <ul className="space-y-5">
            {faqs.map((item) => (
              <li key={item.q} className="border-b border-border pb-5">
                <h3 className="text-[16px] font-semibold tracking-[-0.01em] text-ink">{item.q}</h3>
                <p className="mt-2 text-[15px] leading-relaxed text-muted">{item.a}</p>
              </li>
            ))}
          </ul>
        </div>
      </section>

      <section className="px-4 py-14 sm:px-8 md:px-12 md:pb-20">
        <div className="tuvi-forest-panel mx-auto flex max-w-[1040px] flex-col items-start justify-between gap-6 overflow-hidden rounded-[28px] px-6 py-10 sm:flex-row sm:items-center sm:rounded-[36px] sm:px-10 sm:py-12 md:px-12">
          <div className="max-w-[40ch]">
            <h2 className="font-display text-[clamp(1.7rem,3.2vw,2.4rem)] font-semibold leading-[1.15] tracking-[-0.03em] text-bg">
              See Tuvi on your restaurant — not a generic deck
            </h2>
            <p className="mt-3 text-[15px] leading-relaxed text-bg/75">
              Book a free demo and we’ll walk website, SEO, ordering, and retention for your
              locations. Usually within one business day.
            </p>
          </div>
          <div className="flex flex-wrap gap-3">
            <Link
              href="/demo"
              className="inline-flex items-center justify-center rounded-full bg-bg px-6 py-3 text-[14px] font-semibold text-ink transition-colors hover:bg-sage"
            >
              Get a free demo
            </Link>
            <Link
              href="/resources"
              className="inline-flex items-center justify-center rounded-full bg-white/10 px-6 py-3 text-[14px] font-semibold text-bg ring-1 ring-white/20 transition-colors hover:bg-white/15"
            >
              Read the guides
            </Link>
          </div>
        </div>
      </section>

      <SiteFooter />
    </div>
  );
}
