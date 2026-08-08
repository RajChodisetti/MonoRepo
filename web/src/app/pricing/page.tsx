import type { Metadata } from "next";
import Link from "next/link";
import SiteFooter from "@/components/layout/SiteFooter";
import GrowOnlineCta from "@/components/sections/GrowOnlineCta";
import ProductFaq from "@/components/product/ProductFaq";

export const metadata: Metadata = {
  title: "Pricing | Tuvi",
  description:
    "Simple Tuvi pricing for restaurants. Website, ordering, SEO, loyalty, and more — without marketplace commissions.",
};

const plans = [
  {
    name: "Starter",
    price: "$299",
    period: "/mo",
    blurb: "For single-location restaurants getting online the right way.",
    featured: false,
    cta: "Start with Starter",
    features: [
      "Restaurant website + online menu",
      "Commission-free online ordering",
      "Basic SEO setup",
      "Email support",
      "Guest customer list",
    ],
  },
  {
    name: "Growth",
    price: "$499",
    period: "/mo",
    blurb: "For operators who want discovery, sales, and repeat guests together.",
    featured: true,
    cta: "Get Growth",
    features: [
      "Everything in Starter",
      "Restaurant SEO + listings sync",
      "Reviews engine",
      "Email & SMS marketing",
      "Loyalty & rewards",
      "Priority onboarding",
    ],
  },
  {
    name: "Scale",
    price: "Custom",
    period: "",
    blurb: "For multi-location groups that need the full Tuvi stack.",
    featured: false,
    cta: "Talk to sales",
    features: [
      "Everything in Growth",
      "Branded restaurant app",
      "AI phone ordering",
      "POS integrations",
      "Kitchen tablet + owner app",
      "Dedicated success manager",
    ],
  },
] as const;

const pricingFaq = [
  {
    question: "Is there a long-term contract?",
    answer:
      "Most restaurants start month-to-month after onboarding. Multi-location Scale plans can include annual options for better rates — we will confirm on your demo.",
  },
  {
    question: "Are there marketplace-style commissions?",
    answer:
      "No. Tuvi is built for first-party ordering under your brand. You keep the guest relationship and more of the margin.",
  },
  {
    question: "Can I switch plans later?",
    answer:
      "Yes. Start with what you need and add SEO, loyalty, app, or AI phone as you grow. We will help migrate cleanly.",
  },
  {
    question: "What is included in onboarding?",
    answer:
      "We help with site setup, menu import, branding, and go-live. Growth and Scale include higher-touch onboarding and training.",
  },
];

export default function PricingPage() {
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
            Pricing
          </p>
          <h1 className="mt-4 font-display text-[clamp(2.2rem,5vw,3.75rem)] font-semibold leading-[1.05] tracking-[-0.03em] text-ink">
            Simple pricing for restaurants that want to own growth
          </h1>
          <p className="mx-auto mt-4 max-w-[48ch] text-[16px] leading-relaxed text-muted sm:text-[17px]">
            No marketplace commissions. Pick a plan that matches your stage — upgrade anytime as you
            add products.
          </p>
        </div>
      </section>

      <section className="px-4 pb-16 sm:px-8 sm:pb-20 md:px-12">
        <div className="mx-auto grid max-w-[1100px] gap-4 lg:grid-cols-3 lg:gap-5">
          {plans.map((plan) => (
            <article
              key={plan.name}
              className={`flex flex-col rounded-[28px] p-6 sm:p-7 ${
                plan.featured
                  ? "tuvi-forest-panel text-bg shadow-[0_24px_60px_rgba(15,39,31,0.28)]"
                  : "bg-parchment text-ink ring-1 ring-border"
              }`}
            >
              <div className="flex items-center justify-between gap-3">
                <h2 className="font-display text-[1.45rem] font-semibold tracking-[-0.02em]">
                  {plan.name}
                </h2>
                {plan.featured ? (
                  <span className="rounded-full bg-bg/15 px-2.5 py-1 text-[11px] font-semibold text-bg">
                    Most popular
                  </span>
                ) : null}
              </div>
              <p
                className={`mt-2 text-[14px] leading-relaxed ${
                  plan.featured ? "text-bg/75" : "text-muted"
                }`}
              >
                {plan.blurb}
              </p>
              <p className="mt-6 flex items-baseline gap-1">
                <span className="font-display text-[clamp(2rem,3vw,2.6rem)] font-semibold tracking-[-0.03em]">
                  {plan.price}
                </span>
                {plan.period ? (
                  <span
                    className={`text-[14px] ${plan.featured ? "text-bg/70" : "text-secondary"}`}
                  >
                    {plan.period}
                  </span>
                ) : null}
              </p>
              <ul className="mt-6 flex flex-1 flex-col gap-2.5">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex gap-2.5 text-[14px] leading-snug">
                    <span
                      className={`mt-1 h-1.5 w-1.5 shrink-0 rounded-full ${
                        plan.featured ? "bg-accent" : "bg-primary"
                      }`}
                      aria-hidden="true"
                    />
                    <span className={plan.featured ? "text-bg/90" : "text-ink"}>{feature}</span>
                  </li>
                ))}
              </ul>
              <Link
                href="/book"
                className={`mt-8 inline-flex items-center justify-center rounded-full px-5 py-3 text-[14px] font-semibold transition-colors ${
                  plan.featured
                    ? "bg-bg text-ink hover:bg-sage"
                    : "bg-primary text-bg hover:bg-primary-dim"
                }`}
              >
                {plan.cta}
              </Link>
            </article>
          ))}
        </div>
        <p className="mx-auto mt-8 max-w-[60ch] text-center text-[14px] text-muted">
          Prices shown are illustrative starting points. Final pricing depends on locations and
          modules — get a tailored quote on your demo.
        </p>
      </section>

      <ProductFaq items={pricingFaq} />
      <GrowOnlineCta variant="centered" />
      <SiteFooter />
    </div>
  );
}
