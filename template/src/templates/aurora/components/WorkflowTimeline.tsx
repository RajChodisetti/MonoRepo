"use client";

import { useEffect, useRef, useState } from "react";
import { gsap, registerGsap } from "@/lib/gsap";
import { usePrefersReducedMotion } from "@/hooks/usePrefersReducedMotion";
import { useIsMobile } from "@/hooks/useMediaQuery";
import type { StoryStep } from "@/data/types/restaurant";
import BlurReveal from "./ui/BlurReveal";
import SourceAwareImage, { mediaForURL } from "@/components/SourceAwareImage";
import PhotoAttribution from "@/components/PhotoAttribution";

export default function WorkflowTimeline({ steps }: { steps: StoryStep[] }) {
  const sectionRef = useRef<HTMLElement>(null);
  const [active, setActive] = useState(0);
  const reduced = usePrefersReducedMotion();
  const isMobile = useIsMobile();

  useEffect(() => {
    if (reduced || isMobile || !steps.length) return;
    registerGsap();

    const section = sectionRef.current;
    if (!section) return;

    const ctx = gsap.context(() => {
      gsap.timeline({
        scrollTrigger: {
          trigger: section,
          start: "top top",
          end: `+=${steps.length * 60}%`,
          scrub: true,
          pin: true,
          onUpdate: (self) => {
            const idx = Math.min(steps.length - 1, Math.floor(self.progress * steps.length));
            setActive(idx);
          },
        },
      });
    }, section);

    return () => ctx.revert();
  }, [reduced, isMobile, steps.length]);

  if (!steps.length) return null;

  const step = steps[active];

  return (
    <section ref={sectionRef} id="workflow" className="aurora-section min-h-screen">
      <div className="aurora-container grid min-h-screen items-center gap-12 lg:grid-cols-2">
        <BlurReveal>
          <p className="text-xs uppercase tracking-[0.2em] text-blue-400">Our Journey</p>
          <h2 className="aurora-heading mt-3 text-4xl font-bold text-white md:text-5xl">
            Every plate tells a story
          </h2>
          <p className="mt-4 text-white/50">
            From the first ingredient to the final garnish — craft, tradition, and care in every step.
          </p>

          <div className="mt-10 space-y-4">
            {steps.map((s, i) => (
              <button
                key={s.number}
                type="button"
                onClick={() => setActive(i)}
                className={`block w-full rounded-xl border p-4 text-left transition ${
                  i === active
                    ? "border-purple-500/40 bg-purple-500/10"
                    : "border-white/10 bg-white/5 hover:border-white/20"
                }`}
              >
                <span className="text-xs text-purple-400">{s.number} / {String(steps.length).padStart(2, "0")}</span>
                <p className="aurora-heading mt-1 font-semibold text-white">{s.title}</p>
              </button>
            ))}
          </div>
        </BlurReveal>

        <div className="relative aspect-square overflow-hidden rounded-2xl border border-white/10">
          {steps.map((s, i) => {
            const media = s.imageMedia || mediaForURL(s.image, s.title);
            return (
              <div
                key={s.number}
                className={`absolute inset-0 transition-all duration-700 ${
                  i === active ? "opacity-100 scale-100" : "opacity-0 scale-105"
                }`}
              >
                <SourceAwareImage media={media} fill className="object-cover" sizes="50vw" />
                <div className="absolute inset-0 bg-gradient-to-t from-[#09090B] via-[#09090B]/40 to-transparent" />
                {media.sourceKind === "google_places_live" ? (
                  <div className="absolute inset-x-3 top-3 z-10 rounded bg-black/65 px-2 py-1 text-white/80">
                    <PhotoAttribution media={media} compact />
                  </div>
                ) : null}
                <div className="absolute bottom-0 p-8">
                  <h3 className="aurora-heading text-2xl font-bold text-white">{s.title}</h3>
                  <p className="mt-2 text-white/70">{s.description}</p>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
