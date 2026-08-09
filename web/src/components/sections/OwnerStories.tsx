"use client";

import Image from "next/image";
import Link from "next/link";
import { successStories, type SuccessStory } from "@/components/sections/ownerStories.config";

function StoryCard({ story }: { story: SuccessStory }) {
  return (
    <Link
      href={story.href}
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
  const loop = [...successStories, ...successStories];

  return (
    <section className="overflow-hidden bg-bg py-14 sm:py-20">
      <div className="mx-auto max-w-[1100px] px-4 sm:px-8 md:px-12">
        <h2 className="font-display text-[clamp(2.15rem,4.2vw,3.35rem)] font-semibold tracking-[-0.03em] text-ink">
          Restaurant growth stories
        </h2>
        <p className="mt-3 max-w-[44ch] text-[15px] leading-relaxed text-muted sm:text-[16px]">
          Explore fictional, illustrative examples of how independent restaurants can grow direct
          sales and keep the guest. Results vary by restaurant.
        </p>
      </div>

      <div className="group/marquee mt-8 motion-reduce:overflow-x-auto motion-reduce:pb-2 sm:mt-10">
        <div className="owner-marquee-track flex w-max gap-4 pl-4 [animation-duration:63s] group-hover/marquee:[animation-play-state:paused] sm:gap-5 sm:pl-8 md:pl-12">
          {loop.map((story, index) => (
            <StoryCard key={`${story.id}-${index}`} story={story} />
          ))}
        </div>
      </div>
    </section>
  );
}
