"use client";

import Image from "next/image";
import { successStories, type SuccessStory } from "@/components/sections/ownerStories.config";

function StoryCard({ story }: { story: SuccessStory }) {
  return (
    <article className="relative h-[440px] w-[320px] shrink-0 overflow-hidden rounded-[28px] sm:h-[480px] sm:w-[340px]">
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
      </div>
    </article>
  );
}

export default function OwnerStories() {
  const loop = [...successStories, ...successStories];

  return (
    <section className="overflow-hidden bg-bg py-14 sm:py-20">
      <div className="mx-auto max-w-[1100px] px-4 sm:px-8 md:px-12">
        <h2 className="font-display text-[clamp(2.15rem,4.2vw,3.35rem)] font-semibold tracking-[-0.03em] text-ink">
          Grow sales like these owners
        </h2>
      </div>

      <div className="group/marquee mt-8 sm:mt-10">
        <div className="owner-marquee-track flex w-max gap-4 pl-4 group-hover/marquee:[animation-play-state:paused] sm:gap-5 sm:pl-8 md:pl-12">
          {loop.map((story, index) => (
            <StoryCard key={`${story.id}-${index}`} story={story} />
          ))}
        </div>
      </div>
    </section>
  );
}
