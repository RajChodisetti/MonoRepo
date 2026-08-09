import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import SiteFooter from "@/components/layout/SiteFooter";
import { caseStudies, getCaseStudy } from "@/content/caseStudies";

type PageProps = {
  params: Promise<{ slug: string }>;
};

export function generateStaticParams() {
  return caseStudies.map((study) => ({ slug: study.slug }));
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params;
  const study = getCaseStudy(slug);
  if (!study) return { title: "Case study | Tuvi" };
  return {
    title: `${study.restaurant}: ${study.metricValue} ${study.metricDescription} | Tuvi`,
    description: study.summary,
  };
}

export default async function CaseStudyPage({ params }: PageProps) {
  const { slug } = await params;
  const study = getCaseStudy(slug);
  if (!study) notFound();

  return (
    <>
      <article>
        <section className="hero-atmosphere relative overflow-hidden px-4 pb-10 pt-14 sm:px-8 sm:pb-14 sm:pt-18 md:px-12">
          <div
            className="pointer-events-none absolute inset-0 hero-grid opacity-25 [mask-image:radial-gradient(40rem_24rem_at_50%_25%,black,transparent)]"
            aria-hidden="true"
          />
          <div className="relative z-10 mx-auto max-w-[920px]">
            <Link
              href="/resources/case-studies"
              className="text-[13px] font-semibold text-primary transition-colors hover:text-primary-dim"
            >
              ← All case studies
            </Link>
            <p className="mt-5 text-[12px] font-semibold uppercase tracking-[0.18em] text-muted">
              Fictional illustrative example · {study.location}
            </p>
            <h1 className="mt-3 font-display text-[clamp(2.1rem,4.6vw,3.5rem)] font-semibold leading-[1.08] tracking-[-0.03em] text-ink">
              How {study.name} achieved {study.metricValue} {study.metricDescription.toLowerCase()} at{" "}
              {study.restaurant}
            </h1>
            <p className="mt-4 max-w-[54ch] text-[16px] leading-relaxed text-muted sm:text-[17px]">
              {study.summary}
            </p>
            <div className="mt-6 flex flex-wrap gap-2">
              {study.services.map((service) => (
                <span
                  key={service}
                  className="rounded-full border border-black/10 bg-white/70 px-3 py-1 text-[12px] font-medium text-ink"
                >
                  {service}
                </span>
              ))}
            </div>
          </div>
        </section>

        <section className="px-4 pb-6 sm:px-8 md:px-12">
          <div className="relative mx-auto aspect-[16/9] max-w-[920px] overflow-hidden rounded-[28px]">
            <Image
              src={study.imageUrl}
              alt={`Illustrative portrait of ${study.name}, ${study.role.toLowerCase()} of ${study.restaurant}`}
              fill
              priority
              sizes="(max-width: 960px) 100vw, 920px"
              className="object-cover object-top"
            />
            <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/75 to-transparent p-6 text-white sm:p-8">
              <p className="text-[28px] font-bold tracking-[-0.03em] sm:text-[34px]">
                {study.metricValue}
              </p>
              <p className="mt-1 text-[15px] text-white/85">{study.metricDescription}</p>
              <p className="mt-3 text-[14px] font-semibold">
                {study.name} · {study.role} of {study.restaurant}
              </p>
            </div>
          </div>
        </section>

        <section className="px-4 py-12 sm:px-8 md:px-12">
          <div className="mx-auto grid max-w-[920px] gap-10 md:grid-cols-[1fr_1.35fr] md:gap-14">
            <div>
              <h2 className="font-display text-[1.65rem] font-semibold tracking-[-0.02em] text-ink">
                The challenge
              </h2>
              <p className="mt-3 text-[15px] leading-relaxed text-muted sm:text-[16px]">
                {study.challenge}
              </p>
            </div>
            <div>
              <h2 className="font-display text-[1.65rem] font-semibold tracking-[-0.02em] text-ink">
                How Tuvi helped
              </h2>
              <ol className="mt-4 space-y-3">
                {study.approach.map((step, index) => (
                  <li key={step} className="flex gap-3 text-[15px] leading-relaxed text-muted sm:text-[16px]">
                    <span className="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-[12px] font-semibold text-bg">
                      {index + 1}
                    </span>
                    <span>{step}</span>
                  </li>
                ))}
              </ol>
            </div>
          </div>
        </section>

        <section className="border-y border-black/6 bg-white px-4 py-12 sm:px-8 md:px-12">
          <div className="mx-auto max-w-[920px]">
            <h2 className="font-display text-[1.65rem] font-semibold tracking-[-0.02em] text-ink">
              Results
            </h2>
            <div className="mt-6 grid gap-4 sm:grid-cols-3">
              {study.results.map((result) => (
                <div key={result.label} className="rounded-[20px] bg-white px-5 py-5">
                  <p className="text-[28px] font-bold tracking-[-0.03em] text-ink">{result.value}</p>
                  <p className="mt-1 text-[14px] text-muted">{result.label}</p>
                </div>
              ))}
            </div>
            <blockquote className="mt-10 border-l-2 border-primary pl-5">
              <p className="font-display text-[1.35rem] font-medium leading-snug tracking-[-0.02em] text-ink sm:text-[1.55rem]">
                “{study.quote}”
              </p>
              <footer className="mt-3 text-[14px] text-muted">
                — {study.name}, {study.role} of {study.restaurant}
              </footer>
            </blockquote>
          </div>
        </section>

        <section className="px-4 py-16 sm:px-8 md:px-12">
          <div className="mx-auto flex max-w-[920px] flex-col items-start justify-between gap-6 rounded-[28px] bg-ink px-7 py-9 text-bg sm:flex-row sm:items-center sm:px-10">
            <div>
              <h2 className="font-display text-[clamp(1.5rem,3vw,2.1rem)] font-semibold tracking-[-0.03em]">
                Want a growth plan for your restaurant?
              </h2>
              <p className="mt-2 max-w-[40ch] text-[15px] text-bg/70">
                Book a free demo and we&apos;ll map the same playbook to your locations.
              </p>
            </div>
            <Link
              href="/book"
              className="inline-flex shrink-0 items-center justify-center rounded-full bg-primary px-6 py-3 text-[14px] font-semibold text-bg transition-colors hover:bg-primary-dim"
            >
              Get a free demo
            </Link>
          </div>
        </section>
      </article>
      <SiteFooter />
    </>
  );
}
