import type { Metadata } from "next";
import Link from "next/link";
import SiteFooter from "@/components/layout/SiteFooter";
import DemoContactForm from "@/components/demo/DemoContactForm";

export const metadata: Metadata = {
  title: "Get a free demo | Tuvi",
  description:
    "Tell us about your restaurant and we will show how Tuvi grows discovery, first-party orders, and repeat guests.",
};

export default function DemoPage() {
  return (
    <>
      <section className="hero-atmosphere relative px-6 pb-16 pt-14 md:px-10 md:pb-20 md:pt-20 lg:px-12">
        <div
          className="pointer-events-none absolute inset-0 hero-grid opacity-25 [mask-image:radial-gradient(40rem_24rem_at_50%_20%,black,transparent)]"
          aria-hidden="true"
        />
        <div className="relative z-10 mx-auto grid w-full max-w-[1040px] gap-10 lg:grid-cols-[1fr_1.05fr] lg:items-start lg:gap-14">
          <div className="pt-2 text-center lg:pt-8 lg:text-left">
            <p className="text-[13px] font-semibold uppercase tracking-[0.1em] text-primary">
              Free demo
            </p>
            <h1 className="mt-3 font-display text-[clamp(1.9rem,4vw,3.25rem)] font-semibold leading-[1.12] tracking-[-0.04em] text-ink">
              See Tuvi on your restaurant — not a generic pitch deck
            </h1>
            <p className="mx-auto mt-4 max-w-[40ch] text-[16px] leading-relaxed text-muted lg:mx-0">
              Share a few details and we&apos;ll walk through website, SEO, ordering, and guest retention
              for your locations. Usually within one business day.
            </p>
            <ul className="mx-auto mt-8 flex max-w-[34ch] flex-col gap-2.5 text-left text-[14px] text-ink lg:mx-0">
              {[
                "No marketplace commissions",
                "Built for Australian restaurants",
                "Clear next steps after the call",
              ].map((item) => (
                <li key={item} className="flex items-start gap-2">
                  <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" aria-hidden />
                  {item}
                </li>
              ))}
            </ul>
            <Link
              href="/how-it-works"
              className="mt-6 inline-flex text-[14px] font-semibold text-primary transition-colors hover:text-primary-dim"
            >
              See how it works →
            </Link>
          </div>

          <DemoContactForm />
        </div>
      </section>
      <SiteFooter />
    </>
  );
}
