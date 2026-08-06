"use client";

import Image from "next/image";
import {
  ratingReviewsRow1,
  ratingReviewsRow2,
  type RatingReview,
} from "@/components/sections/ratingsProof.config";

function StarIcon({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 20 20" className={className} aria-hidden="true">
      <path
        fill="currentColor"
        d="M10 1.8l2.35 4.76 5.25.76-3.8 3.7.9 5.22L10 13.7l-4.7 2.54.9-5.22-3.8-3.7 5.25-.76L10 1.8z"
      />
    </svg>
  );
}

function ReviewCard({ review }: { review: RatingReview }) {
  return (
    <article
      className="flex h-[200px] w-[280px] shrink-0 flex-col rounded-[22px] px-5 pb-5 pt-5 sm:h-[210px] sm:w-[300px] sm:rounded-[24px] sm:px-6"
      style={{ backgroundColor: "#dce6dd" }}
    >
      <div className="flex items-center gap-0.5 text-ink" aria-label="5 stars">
        {Array.from({ length: 5 }).map((_, i) => (
          <StarIcon key={i} className="h-3.5 w-3.5" />
        ))}
      </div>
      <p className="mt-3 flex-1 text-[15px] font-bold leading-snug tracking-[-0.02em] text-ink sm:text-[16px]">
        &ldquo;{review.title}&rdquo;
      </p>
      <div className="mt-4 flex items-center gap-2.5">
        <span className="relative h-8 w-8 shrink-0 overflow-hidden rounded-full">
          <Image src={review.avatar} alt="" fill className="object-cover" sizes="32px" />
        </span>
        <span className="text-[13px] font-semibold text-[#1a1a1a]">{review.name}</span>
      </div>
    </article>
  );
}

function MarqueeRow({
  reviews,
  reverse = false,
}: {
  reviews: RatingReview[];
  reverse?: boolean;
}) {
  const loop = [...reviews, ...reviews];
  return (
    <div className="group/marquee overflow-hidden">
      <div
        className={`flex w-max gap-3 sm:gap-3.5 ${
          reverse ? "owner-marquee-track-reverse" : "owner-marquee-track"
        } group-hover/marquee:[animation-play-state:paused]`}
      >
        {loop.map((review, index) => (
          <ReviewCard key={`${review.id}-${index}`} review={review} />
        ))}
      </div>
    </div>
  );
}

export default function RatingsProof() {
  return (
    <section className="bg-bg pt-10 sm:pt-14">
      <div className="tuvi-forest-panel relative w-full overflow-hidden rounded-t-[28px] sm:rounded-t-[36px] md:rounded-t-[44px]">
        <div className="relative pt-14 pb-10 sm:pt-16 sm:pb-12 md:pt-20 md:pb-14">
          <svg
            className="pointer-events-none absolute inset-0 h-full w-full"
            viewBox="0 0 1440 900"
            preserveAspectRatio="none"
            aria-hidden="true"
          >
            {[280, 420, 560, 720, 900, 1120].map((r, i) => (
              <circle
                key={r}
                cx="720"
                cy="940"
                r={r}
                fill="none"
                stroke="rgba(255,255,255,0.12)"
                strokeWidth={1.2 - i * 0.05}
              />
            ))}
          </svg>

          <div
            className="pointer-events-none absolute inset-0 opacity-[0.18] mix-blend-overlay"
            style={{
              backgroundImage:
                "url(\"data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E\")",
            }}
            aria-hidden="true"
          />

          <div className="relative z-10 mx-auto max-w-[820px] px-6 text-center sm:px-8">
            <h2 className="font-display text-[clamp(2.5rem,5vw,4rem)] font-semibold leading-[1.1] tracking-[-0.03em] text-bg">
              <span className="block">See why owners switch</span>
              <span className="block">to Tuvi</span>
            </h2>

            <p className="mt-6 flex flex-wrap items-center justify-center gap-x-2.5 gap-y-1 text-[16px] font-medium text-bg/75 sm:mt-8 sm:text-[18px]">
              <span>4.8</span>
              <span className="inline-flex items-center gap-1 text-bg" aria-label="4.8 stars">
                {Array.from({ length: 5 }).map((_, i) => (
                  <StarIcon key={i} className="h-4 w-4 sm:h-[18px] sm:w-[18px]" />
                ))}
              </span>
              <span>Trusted by 200+ restaurants growing direct sales with Tuvi</span>
            </p>
          </div>

          <div className="relative z-10 mt-12 flex flex-col gap-3 sm:mt-14 sm:gap-3.5">
            <MarqueeRow reviews={ratingReviewsRow1} />
            <MarqueeRow reviews={ratingReviewsRow2} reverse />
          </div>
        </div>
      </div>
    </section>
  );
}
