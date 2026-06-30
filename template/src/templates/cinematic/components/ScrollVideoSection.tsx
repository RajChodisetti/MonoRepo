"use client";

import { useEffect, useRef } from "react";
import Image from "next/image";
import { gsap, registerGsap } from "@/lib/gsap";
import { usePrefersReducedMotion } from "@/hooks/usePrefersReducedMotion";
import { useIsMobile } from "@/hooks/useMediaQuery";

interface ScrollVideoSectionProps {
  videoSrc: string;
  posterSrc: string;
  eyebrow: string;
  title: string;
  description: string;
  ctaLabel: string;
  ctaHref: string;
  durationMultiplier?: number;
}

export default function ScrollVideoSection({
  videoSrc,
  posterSrc,
  eyebrow,
  title,
  description,
  ctaLabel,
  ctaHref,
  durationMultiplier = 2.5,
}: ScrollVideoSectionProps) {
  const sectionRef = useRef<HTMLElement>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const reduced = usePrefersReducedMotion();
  const isMobile = useIsMobile();

  useEffect(() => {
    if (reduced || isMobile) return;
    registerGsap();

    const section = sectionRef.current;
    const video = videoRef.current;
    if (!section || !video) return;

    let tween: gsap.core.Tween | null = null;

    const setup = () => {
      tween = gsap.to(video, {
        currentTime: video.duration || 0,
        ease: "none",
        scrollTrigger: {
          trigger: section,
          start: "top top",
          end: `+=${Math.round(durationMultiplier * 100)}%`,
          scrub: true,
          pin: true,
        },
      });
    };

    if (video.readyState >= 1) setup();
    else video.addEventListener("loadedmetadata", setup, { once: true });

    return () => {
      tween?.scrollTrigger?.kill();
      tween?.kill();
    };
  }, [reduced, isMobile, durationMultiplier]);

  const showPosterOnly = reduced || isMobile;

  return (
    <section ref={sectionRef} className="relative min-h-screen bg-charcoal">
      <div className="relative h-screen w-full overflow-hidden">
        {showPosterOnly ? (
          <Image src={posterSrc} alt="" fill loading="lazy" className="object-cover" sizes="100vw" />
        ) : (
          <video
            ref={videoRef}
            src={videoSrc}
            poster={posterSrc}
            muted
            playsInline
            preload="metadata"
            className="h-full w-full object-cover"
          />
        )}
        <div className="absolute inset-0 bg-gradient-to-r from-charcoal/90 via-charcoal/50 to-transparent" />
        <div className="absolute inset-0 flex items-center">
          <div className="mx-auto max-w-6xl px-6">
            <p className="mb-3 text-xs uppercase tracking-[0.2em] text-brass">{eyebrow}</p>
            <h2 className="font-display max-w-2xl text-4xl text-cream md:text-6xl">{title}</h2>
            <p className="mt-4 max-w-lg text-cream/70">{description}</p>
            <a href={ctaHref} className="btn-primary mt-8 inline-flex">
              {ctaLabel}
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
