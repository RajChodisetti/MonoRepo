import type { Metadata } from "next";
import Link from "next/link";
import SiteFooter from "@/components/layout/SiteFooter";
import GrowOnlineCta from "@/components/sections/GrowOnlineCta";
import TrustedByOwners from "@/components/sections/TrustedByOwners";

export const metadata: Metadata = {
  title: "How it works | Tuvi",
  description:
    "See how Tuvi helps restaurants get found, take more first-party orders, and bring guests back — under your brand.",
};

const steps = [
  {
    num: "01",
    title: "We launch your online foundation",
    body: "Website, menu, and ordering go live under your brand — so guests can find you and order without a marketplace tax.",
  },
  {
    num: "02",
    title: "We grow discovery where guests search",
    body: "SEO, listings, and reviews work together so more nearby hungry people see you first on Google.",
  },
  {
    num: "03",
    title: "We turn one-time guests into regulars",
    body: "Loyalty, email, SMS, and push keep your best customers coming back — with offers timed to fill slow hours.",
  },
  {
    num: "04",
    title: "You run the restaurant. Tuvi runs the stack.",
    body: "Owner app, kitchen tickets, analytics, and POS sync keep operations clear while online sales keep growing.",
  },
] as const;

const pillars = [
  {
    title: "Grow discovery",
    body: "Website, SEO, menu, reviews, and listings that help guests find you.",
    href: "/product/seo",
  },
  {
    title: "Grow online sales",
    body: "Ordering, upsells, delivery, catering, and AI phone — commission-free.",
    href: "/product/ordering",
  },
  {
    title: "Grow repeat orders",
    body: "App, campaigns, email/SMS, push, and loyalty that bring guests back.",
    href: "/product/loyalty",
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
        <div className="relative z-10 mx-auto max-w-[860px] text-center">
          <p className="inline-flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">
            <span className="h-1.5 w-1.5 rounded-full bg-accent" aria-hidden="true" />
            How it works
          </p>
          <h1 className="mt-4 font-display text-[clamp(2.2rem,5vw,3.75rem)] font-semibold leading-[1.05] tracking-[-0.03em] text-ink">
            From first Google search to regular guest — under your brand
          </h1>
          <p className="mx-auto mt-4 max-w-[46ch] text-[16px] leading-relaxed text-muted sm:text-[17px]">
            Tuvi is the AI platform restaurants use to get found, take more first-party orders, and
            grow repeats without giving the guest away.
          </p>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
            <Link
              href="/demo"
              className="inline-flex items-center justify-center rounded-full bg-primary px-6 py-3 text-[14px] font-semibold text-bg transition-colors hover:bg-primary-dim"
            >
              Get a free demo
            </Link>
            <Link
              href="/pricing"
              className="inline-flex items-center justify-center rounded-full bg-surface px-6 py-3 text-[14px] font-semibold text-ink transition-colors hover:bg-parchment"
            >
              View pricing
            </Link>
          </div>
        </div>
      </section>

      <section className="px-4 pb-6 sm:px-8 md:px-12">
        <div className="mx-auto max-w-[1040px]">
          <h2 className="max-w-[18ch] font-display text-[clamp(1.7rem,3.4vw,2.6rem)] font-semibold tracking-[-0.03em] text-ink">
            Four steps to growing online the Tuvi way
          </h2>
          <div className="mt-10 grid gap-4 sm:gap-5 md:grid-cols-2">
            {steps.map((step) => (
              <article
                key={step.num}
                className="rounded-[28px] bg-parchment p-6 ring-1 ring-border sm:p-7"
              >
                <p className="text-[13px] font-semibold tracking-[0.14em] text-primary">{step.num}</p>
                <h3 className="mt-3 font-display text-[clamp(1.25rem,2vw,1.55rem)] font-semibold leading-snug tracking-[-0.02em] text-ink">
                  {step.title}
                </h3>
                <p className="mt-2.5 text-[15px] leading-relaxed text-muted">{step.body}</p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="px-4 py-14 sm:px-8 sm:py-16 md:px-12 md:py-18">
        <div className="mx-auto max-w-[1040px]">
          <h2 className="max-w-[16ch] font-display text-[clamp(1.7rem,3.4vw,2.6rem)] font-semibold tracking-[-0.03em] text-ink">
            One platform. Three growth jobs.
          </h2>
          <div className="mt-8 grid gap-4 sm:mt-10 sm:gap-5 lg:grid-cols-3">
            {pillars.map((pillar) => (
              <Link
                key={pillar.title}
                href={pillar.href}
                className="group rounded-[28px] bg-bg p-6 ring-1 ring-border transition-colors hover:bg-sage/40 sm:p-7"
              >
                <h3 className="font-display text-[1.35rem] font-semibold tracking-[-0.02em] text-ink">
                  {pillar.title}
                </h3>
                <p className="mt-2 text-[14px] leading-relaxed text-muted">{pillar.body}</p>
                <span className="mt-5 inline-flex text-[13px] font-semibold text-primary transition-transform group-hover:translate-x-0.5">
                  Explore →
                </span>
              </Link>
            ))}
          </div>
        </div>
      </section>

      <section className="px-4 pb-10 sm:px-8 md:px-12">
        <div className="tuvi-forest-panel mx-auto max-w-[1040px] overflow-hidden rounded-[28px] px-6 py-10 sm:rounded-[36px] sm:px-10 sm:py-12 md:px-12">
          <div className="max-w-[36ch]">
            <h2 className="font-display text-[clamp(1.7rem,3.2vw,2.4rem)] font-semibold leading-[1.15] tracking-[-0.03em] text-bg">
              Ready in weeks, not months of agency backlog
            </h2>
            <p className="mt-3 text-[15px] leading-relaxed text-bg/75">
              Onboarding covers branding, menu, and go-live. Then Tuvi keeps optimizing discovery,
              orders, and repeats while you cook.
            </p>
            <Link
              href="/demo"
              className="mt-6 inline-flex items-center justify-center rounded-full bg-bg px-6 py-3 text-[14px] font-semibold text-ink transition-colors hover:bg-sage"
            >
              Book a walkthrough
            </Link>
          </div>
        </div>
      </section>

      <TrustedByOwners />
      <GrowOnlineCta variant="centered" />
      <SiteFooter />
    </div>
  );
}
