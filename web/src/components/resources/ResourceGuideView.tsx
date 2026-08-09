import Image from "next/image";
import Link from "next/link";
import SiteFooter from "@/components/layout/SiteFooter";
import { resourceGuides, type ResourceGuide } from "@/content/resourceGuides";

export default function ResourceGuideView({ guide }: { guide: ResourceGuide }) {
  const others = resourceGuides
    .filter((g) => g.slug !== guide.slug && Boolean(g.heroImage))
    .slice(0, 3);

  return (
    <>
      <article>
        <section className="hero-atmosphere relative overflow-hidden px-4 pb-8 pt-14 sm:px-8 sm:pb-10 sm:pt-18 md:px-12">
          <div
            className="pointer-events-none absolute inset-0 hero-grid opacity-25 [mask-image:radial-gradient(40rem_24rem_at_50%_20%,black,transparent)]"
            aria-hidden="true"
          />
          <div className="relative z-10 mx-auto max-w-[920px]">
            <Link
              href="/resources"
              className="text-[13px] font-semibold text-primary transition-colors hover:text-primary-dim"
            >
              ← All resources
            </Link>
            <p className="mt-5 text-[11px] font-semibold uppercase tracking-[0.2em] text-primary">
              {guide.eyebrow}
            </p>
            <h1 className="mt-3 max-w-[18ch] font-display text-[clamp(2.15rem,4.8vw,3.6rem)] font-semibold leading-[1.06] tracking-[-0.03em] text-ink">
              {guide.title}
            </h1>
            <p className="mt-4 max-w-[52ch] text-[16px] leading-relaxed text-muted sm:text-[17px]">
              {guide.description}
            </p>
            <div className="mt-4 flex flex-wrap items-center gap-x-3 gap-y-1 text-[13px] font-medium text-secondary">
              <span>{guide.publishedLabel}</span>
              <span aria-hidden>·</span>
              <span>{guide.readTime}</span>
            </div>
          </div>
        </section>

        <section className="px-4 sm:px-8 md:px-12">
          <div className="relative mx-auto aspect-[16/9] max-w-[920px] overflow-hidden rounded-[28px] sm:rounded-[32px]">
            <Image
              src={guide.heroImage}
              alt={guide.heroAlt}
              fill
              priority
              className="object-cover"
              sizes="(max-width: 960px) 100vw, 920px"
            />
          </div>
        </section>

        <section className="px-4 py-10 sm:px-8 md:px-12">
          <div className="mx-auto grid max-w-[920px] gap-4 sm:grid-cols-3">
            {guide.takeaways.map((item) => (
              <div
                key={item}
                className="rounded-[20px] border border-black/8 bg-white px-4 py-4"
              >
                <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-primary">
                  Takeaway
                </p>
                <p className="mt-2 text-[14px] font-semibold leading-snug tracking-[-0.01em] text-ink">
                  {item}
                </p>
              </div>
            ))}
          </div>
        </section>

        <section className="px-4 pb-16 sm:px-8 md:px-12">
          <div className="mx-auto max-w-[720px] space-y-14">
            {guide.sections.map((section, index) => (
              <div key={section.heading}>
                <h2 className="font-display text-[clamp(1.45rem,2.4vw,1.85rem)] font-semibold tracking-[-0.02em] text-ink">
                  {section.heading}
                </h2>
                <p className="mt-3 text-[16px] leading-relaxed text-muted">{section.body}</p>
                {section.bullets?.length ? (
                  <ul className="mt-4 space-y-2">
                    {section.bullets.map((bullet) => (
                      <li key={bullet} className="flex gap-2.5 text-[15px] text-ink">
                        <span className="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" aria-hidden />
                        <span>{bullet}</span>
                      </li>
                    ))}
                  </ul>
                ) : null}
                {section.imageSrc ? (
                  <div className="relative mt-7 aspect-[4/3] overflow-hidden rounded-[24px]">
                    <Image
                      src={section.imageSrc}
                      alt={section.imageAlt ?? ""}
                      fill
                      className="object-cover"
                      sizes="(max-width: 768px) 100vw, 720px"
                    />
                  </div>
                ) : null}
                {guide.quote && index === 1 ? (
                  <blockquote className="mt-10 border-l-2 border-primary pl-5">
                    <p className="font-display text-[1.35rem] font-medium leading-snug tracking-[-0.02em] text-ink sm:text-[1.5rem]">
                      “{guide.quote.text}”
                    </p>
                    <footer className="mt-3 text-[13px] text-muted">— {guide.quote.attribution}</footer>
                  </blockquote>
                ) : null}
              </div>
            ))}

            <div className="overflow-hidden rounded-[28px] bg-ink px-6 py-8 text-bg sm:px-8">
              <h3 className="font-display text-[1.45rem] font-semibold tracking-[-0.02em]">
                Ready to put this into practice?
              </h3>
              <p className="mt-2 max-w-[40ch] text-[15px] text-bg/70">
                We’ll map these plays to your locations — website, SEO, ordering, and retention under your brand.
              </p>
              <div className="mt-5 flex flex-wrap gap-3">
                {guide.relatedHref ? (
                  <Link
                    href={guide.relatedHref}
                    className="inline-flex rounded-full bg-primary px-5 py-2.5 text-[14px] font-semibold text-bg hover:bg-primary-dim"
                  >
                    {guide.relatedLabel ?? "Learn more"}
                  </Link>
                ) : null}
                <Link
                  href="/book"
                  className="inline-flex rounded-full bg-white/10 px-5 py-2.5 text-[14px] font-semibold text-bg ring-1 ring-white/20 hover:bg-white/15"
                >
                  Get a free demo
                </Link>
              </div>
            </div>
          </div>
        </section>

        <section className="border-t border-border bg-white px-4 py-14 sm:px-8 md:px-12">
          <div className="mx-auto max-w-[920px]">
            <h2 className="font-display text-[1.5rem] font-semibold tracking-[-0.02em] text-ink">
              Keep reading
            </h2>
            <div className="mt-6 grid gap-5 sm:grid-cols-3">
              {others.map((item) => (
                <Link
                  key={item.slug}
                  href={`/resources/${item.slug}`}
                  className="group overflow-hidden rounded-[22px] border border-black/8 bg-bg transition-shadow hover:shadow-[0_16px_40px_rgba(0,0,0,0.08)]"
                >
                  <div className="relative aspect-[16/10] overflow-hidden">
                    <Image
                      src={item.heroImage}
                      alt=""
                      fill
                      className="object-cover transition-transform duration-500 group-hover:scale-[1.03]"
                      sizes="300px"
                    />
                  </div>
                  <div className="p-4">
                    <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-primary">
                      {item.eyebrow}
                    </p>
                    <p className="mt-2 text-[15px] font-semibold leading-snug tracking-[-0.01em] text-ink">
                      {item.title}
                    </p>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        </section>
      </article>
      <SiteFooter />
    </>
  );
}
