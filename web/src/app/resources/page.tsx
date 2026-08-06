import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import SiteFooter from "@/components/layout/SiteFooter";
import { blogLibrarySlugs, resourceGuides } from "@/content/resourceGuides";

export const metadata: Metadata = {
  title: "Resources & blog | Tuvi",
  description:
    "Guides for restaurant operators — marketing, SEO, ordering, apps, and websites — with practical plays you can run under your brand.",
};

const featured = resourceGuides.filter((g) =>
  (blogLibrarySlugs as readonly string[]).includes(g.slug),
);

export default function ResourcesIndexPage() {
  return (
    <>
      <section className="hero-atmosphere relative overflow-hidden px-4 pb-12 pt-14 sm:px-8 sm:pb-16 sm:pt-20 md:px-12">
        <div
          className="pointer-events-none absolute inset-0 hero-grid opacity-30 [mask-image:radial-gradient(40rem_24rem_at_50%_30%,black,transparent)]"
          aria-hidden="true"
        />
        <div className="relative z-10 mx-auto max-w-[860px] text-center">
          <p className="inline-flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-primary">
            <span className="h-1.5 w-1.5 rounded-full bg-accent" aria-hidden="true" />
            Resources
          </p>
          <h1 className="mt-4 font-display text-[clamp(2.2rem,5vw,3.75rem)] font-semibold leading-[1.05] tracking-[-0.03em] text-ink">
            The Tuvi blog for growing restaurants
          </h1>
          <p className="mx-auto mt-4 max-w-[46ch] text-[16px] leading-relaxed text-muted sm:text-[17px]">
            Long-form guides on discovery, ordering, and retention — original writing and imagery, built for operators.
          </p>
        </div>
      </section>

      <section className="px-4 pb-20 sm:px-8 md:px-12">
        <div className="mx-auto grid max-w-[1100px] gap-6 sm:grid-cols-2 lg:grid-cols-3">
          <Link
            href="/resources/case-studies"
            className="group overflow-hidden rounded-[24px] border border-black/8 bg-white sm:col-span-2 lg:col-span-1"
          >
            <div className="relative aspect-[16/10] overflow-hidden bg-ink">
              <Image
                src="/resources/resource-blog-hero.png"
                alt=""
                fill
                className="object-cover opacity-90 transition-transform duration-500 group-hover:scale-[1.03]"
                sizes="400px"
              />
            </div>
            <div className="p-5">
              <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-primary">
                Case studies
              </p>
              <p className="mt-2 font-display text-[1.25rem] font-semibold tracking-[-0.02em] text-ink">
                How restaurants grow with Tuvi
              </p>
            </div>
          </Link>

          {featured.map((guide) => (
            <Link
              key={guide.slug}
              href={`/resources/${guide.slug}`}
              className="group overflow-hidden rounded-[24px] border border-black/8 bg-white transition-shadow hover:shadow-[0_16px_40px_rgba(0,0,0,0.08)]"
            >
              <div className="relative aspect-[16/10] overflow-hidden">
                <Image
                  src={guide.heroImage}
                  alt=""
                  fill
                  className="object-cover transition-transform duration-500 group-hover:scale-[1.03]"
                  sizes="(max-width: 768px) 100vw, 360px"
                />
              </div>
              <div className="p-5">
                <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-primary">
                  {guide.eyebrow}
                </p>
                <p className="mt-2 text-[16px] font-semibold leading-snug tracking-[-0.015em] text-ink">
                  {guide.title}
                </p>
                <p className="mt-2 line-clamp-2 text-[14px] leading-relaxed text-muted">
                  {guide.description}
                </p>
                <p className="mt-3 text-[12px] font-medium text-secondary">{guide.readTime}</p>
              </div>
            </Link>
          ))}
        </div>
      </section>
      <SiteFooter />
    </>
  );
}
