import Image from "next/image";
import Link from "next/link";
import type { ProductHeroConfig } from "@/content/products/types";

function PlayIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-3.5 w-3.5" aria-hidden="true">
      <path d="M9 7.2v9.6l8.2-4.8L9 7.2Z" fill="currentColor" />
    </svg>
  );
}

type ProductHeroProps = {
  hero: ProductHeroConfig;
};

export default function ProductHero({ hero }: ProductHeroProps) {
  const hasOverlay = Boolean(hero.testimonial.title);

  return (
    <section className="hero-atmosphere relative overflow-hidden px-4 pb-7 pt-7 sm:px-8 sm:pb-9 sm:pt-9 md:px-12 md:pt-11">
      <div
        className="pointer-events-none absolute inset-0 hero-grid opacity-30 [mask-image:radial-gradient(40rem_24rem_at_30%_40%,black,transparent)]"
        aria-hidden="true"
      />
      <div
        className="tuvi-drift tuvi-soft-pulse pointer-events-none absolute -right-20 top-4 h-64 w-64 rounded-full bg-sage/70 blur-3xl"
        aria-hidden="true"
      />

      <div className="relative z-10 mx-auto grid max-w-[1040px] items-center gap-7 lg:grid-cols-[1.15fr_0.85fr] lg:gap-11">
        <div>
          <h1 className="tuvi-rise max-w-[15ch] font-display text-[clamp(1.7rem,3.4vw,2.7rem)] font-semibold leading-[1.1] tracking-[-0.03em] text-ink">
            {hero.heading}
          </h1>
          <p
            className="tuvi-rise mt-3.5 max-w-[42ch] text-[15px] leading-relaxed text-muted sm:text-[16px]"
            style={{ animationDelay: "80ms" }}
          >
            {hero.subheading}
          </p>
          <div
            className="tuvi-rise mt-5 flex flex-wrap items-center gap-2.5 sm:mt-6"
            style={{ animationDelay: "140ms" }}
          >
            <Link
              href={hero.primaryCta.href}
              className="inline-flex items-center justify-center rounded-full bg-primary px-5 py-2.5 text-[14px] font-semibold text-bg transition-colors hover:bg-primary-dim"
            >
              {hero.primaryCta.label}
            </Link>
            {hero.secondaryCta ? (
              <Link
                href={hero.secondaryCta.href}
                className="inline-flex items-center justify-center rounded-full bg-surface px-5 py-2.5 text-[14px] font-semibold text-ink transition-colors hover:bg-parchment"
              >
                {hero.secondaryCta.label}
              </Link>
            ) : null}
          </div>
        </div>

        <div
          className="tuvi-rise relative mx-auto w-full max-w-[300px] sm:max-w-[330px] lg:max-w-[350px] lg:justify-self-end"
          style={{ animationDelay: "120ms" }}
        >
          <div className="relative overflow-hidden rounded-[24px] bg-accent p-[5px] sm:rounded-[26px]">
            <div className="relative h-[380px] overflow-hidden rounded-[20px] sm:h-[400px] sm:rounded-[22px]">
              <Image
                src={hero.testimonial.imageSrc}
                alt={hero.testimonial.imageAlt}
                fill
                priority
                quality={92}
                className="object-cover object-[center_20%]"
                sizes="(max-width: 640px) 300px, (max-width: 1024px) 330px, 400px"
              />
              {hasOverlay ? (
                <>
                  <div
                    className="absolute inset-0 bg-gradient-to-t from-ink/80 via-ink/20 to-transparent"
                    aria-hidden="true"
                  />
                  <div className="absolute inset-x-0 bottom-0 flex items-end justify-between gap-3 p-4 sm:p-5">
                    <div className="min-w-0 max-w-[82%]">
                      <p className="text-[14px] font-semibold leading-snug tracking-[-0.015em] text-bg">
                        {hero.testimonial.title}
                      </p>
                      {hero.testimonial.attribution ? (
                        <p className="mt-2 text-[12px] font-medium text-bg/85">
                          {hero.testimonial.attribution}
                        </p>
                      ) : null}
                    </div>
                    <button
                      type="button"
                      className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-ink/50 text-bg backdrop-blur-sm transition-transform hover:scale-105"
                      aria-label="Play testimonial video"
                    >
                      <PlayIcon />
                    </button>
                  </div>
                </>
              ) : null}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
