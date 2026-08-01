"use client";

import { useEffect, useRef, useState } from "react";
import { gsap, registerGsap } from "@/lib/gsap";
import { usePrefersReducedMotion } from "@/hooks/usePrefersReducedMotion";
import { useIsMobile } from "@/hooks/useMediaQuery";
import type { StoryStep } from "@/data/types/restaurant";
import SourceAwareImage, { mediaForURL } from "@/components/SourceAwareImage";
import PhotoAttribution from "@/components/PhotoAttribution";

export default function RestaurantStorySection({
  steps,
  restaurantName,
}: {
  steps: StoryStep[];
  restaurantName: string;
}) {
  const sectionRef = useRef<HTMLElement>(null);
  const pinRef = useRef<HTMLDivElement>(null);
  const [active, setActive] = useState(0);
  const reduced = usePrefersReducedMotion();
  const isMobile = useIsMobile();

  useEffect(() => {
    if (reduced || isMobile || !steps.length) return;
    registerGsap();

    const section = sectionRef.current;
    const pin = pinRef.current;
    if (!section || !pin) return;

    const ctx = gsap.context(() => {
      gsap.timeline({
        scrollTrigger: {
          trigger: section,
          start: "top top",
          end: `+=${steps.length * 80}%`,
          scrub: true,
          pin: pin,
          onUpdate: (self) => {
            const idx = Math.min(
              steps.length - 1,
              Math.floor(self.progress * steps.length)
            );
            setActive(idx);
          },
        },
      });
    }, section);

    return () => ctx.revert();
  }, [reduced, isMobile, steps]);

  if (!steps.length) return null;

  return (
    <section ref={sectionRef} id="story" className="relative bg-charcoal py-24 md:py-0">
      <div ref={pinRef} className="mx-auto grid min-h-screen max-w-6xl grid-cols-1 items-center gap-12 px-6 md:grid-cols-2 md:py-24">
        <div className="relative aspect-[4/5] overflow-hidden rounded-xl">
          {steps.map((step, i) => {
            const media = step.imageMedia || mediaForURL(step.image, step.title);
            return (
              <div
                key={step.number}
                className={`absolute inset-0 transition-all duration-700 ${
                  i === active ? "opacity-100 scale-100" : "opacity-0 scale-105"
                }`}
              >
                <SourceAwareImage
                  media={media}
                  fill
                  loading="lazy"
                  className="object-cover"
                  sizes="(max-width: 768px) 100vw, 50vw"
                />
                {media.sourceKind === "google_places_live" ? (
                  <div className="absolute inset-x-3 bottom-3 z-10 rounded bg-black/65 px-2 py-1 text-white/80">
                    <PhotoAttribution media={media} compact />
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>

        <div className="space-y-8">
          <p className="text-xs uppercase tracking-[0.2em] text-brass">Our Story</p>
          <h2 className="font-display text-4xl text-cream md:text-5xl">
            The {restaurantName} experience
          </h2>
          {steps.map((step, i) => (
            <article
              key={step.number}
              className={`border-l-2 pl-6 transition-all duration-500 ${
                i === active ? "border-brass opacity-100" : "border-cream/10 opacity-40"
              }`}
            >
              <p className="text-xs uppercase tracking-widest text-brass/80">
                {step.number} / {String(steps.length).padStart(2, "0")}
              </p>
              <h3 className="mt-2 font-display text-2xl text-cream">{step.title}</h3>
              <p className="mt-2 text-cream/65">{step.description}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
