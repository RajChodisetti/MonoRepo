import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import SiteFooter from "@/components/layout/SiteFooter";
import { caseStudies, caseStudyPath } from "@/content/caseStudies";

export const metadata: Metadata = {
  title: "Restaurant growth examples | Tuvi",
  description:
    "Explore fictional, illustrative examples of restaurants growing discovery, first-party orders, and repeat guests under their own brand.",
};

export default function CaseStudiesIndexPage() {
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
            Case studies
          </p>
          <h1 className="mt-4 font-display text-[clamp(2.2rem,5vw,3.75rem)] font-semibold leading-[1.05] tracking-[-0.03em] text-ink">
            What restaurant growth can look like
          </h1>
          <p className="mx-auto mt-4 max-w-[46ch] text-[16px] leading-relaxed text-muted sm:text-[17px]">
            Fictional, illustrative stories built around common restaurant goals — each one shows
            a different path to more discovery, orders, or repeats. Results vary.
          </p>
        </div>
      </section>

      <section className="px-4 pb-20 sm:px-8 md:px-12">
        <div className="mx-auto grid max-w-[1100px] gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {caseStudies.map((study) => (
            <Link
              key={study.slug}
              href={caseStudyPath(study.slug)}
              className="group overflow-hidden rounded-[24px] border border-black/8 bg-white transition-shadow hover:shadow-[0_16px_40px_rgba(0,0,0,0.08)]"
            >
              <div className="relative aspect-[4/3] overflow-hidden">
                <Image
                  src={study.imageUrl}
                  alt=""
                  fill
                  sizes="(max-width: 768px) 100vw, 360px"
                  className="object-cover transition-transform duration-500 group-hover:scale-[1.03]"
                />
              </div>
              <div className="p-5">
                <p className="text-[22px] font-bold tracking-[-0.03em] text-ink">
                  {study.metricValue}
                </p>
                <p className="mt-1 text-[14px] text-muted">{study.metricDescription}</p>
                <p className="mt-4 text-[15px] font-semibold text-ink">{study.name}</p>
                <p className="mt-0.5 text-[13px] text-muted">
                  {study.role} of {study.restaurant}
                </p>
              </div>
            </Link>
          ))}
        </div>
      </section>

      <SiteFooter />
    </>
  );
}
