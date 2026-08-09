"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { successStories, type SuccessStory } from "@/components/sections/ownerStories.config";

function StoryCard({ story, isDuplicate = false }: { story: SuccessStory; isDuplicate?: boolean }) {
  return (
    <Link
      href={story.href}
      tabIndex={isDuplicate ? -1 : undefined}
      className="relative block h-[440px] w-[320px] shrink-0 overflow-hidden rounded-[28px] outline-none transition-[transform,box-shadow] duration-300 hover:-translate-y-1 hover:shadow-[0_18px_40px_rgba(0,0,0,0.18)] focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 sm:h-[480px] sm:w-[340px]"
      aria-label={`Read case study: ${story.name} at ${story.business}`}
    >
      <Image src={story.imageUrl} alt="" fill sizes="340px" className="object-cover" />
      <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/25 to-transparent" />
      <div className="absolute inset-x-0 bottom-0 p-5 text-white sm:p-6">
        <p className="text-[28px] font-bold leading-none tracking-[-0.03em] sm:text-[32px]">
          {story.metricValue}
        </p>
        <p className="mt-1.5 text-[15px] font-medium leading-snug text-white/90">
          {story.metricDescription}
        </p>
        <div className="mt-5">
          <p className="text-[14px] font-semibold tracking-[-0.01em]">{story.name}</p>
          <p className="mt-0.5 text-[13px] text-white/75">{story.business}</p>
        </div>
        <p className="mt-4 text-[12px] font-semibold uppercase tracking-[0.14em] text-white/70">
          Read story →
        </p>
      </div>
    </Link>
  );
}

export default function OwnerStories() {
  const [isPlaying, setIsPlaying] = useState(false);

  useEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
    const followMotionPreference = () => setIsPlaying(!mediaQuery.matches);

    followMotionPreference();
    mediaQuery.addEventListener("change", followMotionPreference);
    return () => mediaQuery.removeEventListener("change", followMotionPreference);
  }, []);

  return (
    <section className="overflow-hidden bg-bg py-14 sm:py-20">
      <div className="mx-auto max-w-[1100px] px-4 sm:px-8 md:px-12">
        <div className="flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 className="font-display text-[clamp(2.15rem,4.2vw,3.35rem)] font-semibold tracking-[-0.03em] text-ink">
              Restaurant growth stories
            </h2>
            <p className="mt-3 max-w-[44ch] text-[15px] leading-relaxed text-muted sm:text-[16px]">
              Explore fictional, illustrative examples of how independent restaurants can grow
              direct sales and keep the guest. Results vary by restaurant.
            </p>
          </div>
          <button
            type="button"
            aria-controls="owner-stories-track"
            aria-label={isPlaying ? "Pause automatic story scrolling" : "Start automatic story scrolling"}
            onClick={() => setIsPlaying((playing) => !playing)}
            className="owner-stories-motion-control inline-flex min-h-11 shrink-0 items-center rounded-full border border-ink/15 bg-white px-4 py-2 text-[13px] font-semibold text-ink shadow-sm transition-colors hover:border-ink/30 hover:bg-ink/[0.03] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2"
          >
            {isPlaying ? "Pause auto-scroll" : "Start auto-scroll"}
          </button>
        </div>
      </div>

      <div className="owner-stories-marquee-viewport mt-8 overflow-hidden pl-4 motion-reduce:overflow-x-auto motion-reduce:pb-2 sm:mt-10 sm:pl-8 md:pl-12">
        <div
          id="owner-stories-track"
          data-owner-marquee-track
          data-playing={isPlaying}
          className="owner-stories-marquee-track flex w-max"
        >
          {[false, true].map((isDuplicate) => (
            <div
              key={isDuplicate ? "duplicate" : "original"}
              aria-hidden={isDuplicate || undefined}
              className="flex shrink-0 gap-4 pr-4 sm:gap-5 sm:pr-5"
            >
              {successStories.map((story) => (
                <StoryCard key={story.id} story={story} isDuplicate={isDuplicate} />
              ))}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
