"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import { gsap, registerGsap, ScrollTrigger } from "@/lib/gsap";
import { usePrefersReducedMotion } from "@/hooks/usePrefersReducedMotion";
import type { RestaurantContent } from "@/data/types/restaurant";

export default function HeroRestaurantVideo({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const sectionRef = useRef<HTMLElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const reduced = usePrefersReducedMotion();
  const [useVideo, setUseVideo] = useState(false);

  useEffect(() => {
    const video = videoRef.current;
    if (!video || reduced) return;

    const tryPlay = () => {
      video.play().then(() => setUseVideo(true)).catch(() => setUseVideo(false));
    };
    video.addEventListener("loadeddata", tryPlay);
    tryPlay();
    return () => video.removeEventListener("loadeddata", tryPlay);
  }, [reduced]);

  useEffect(() => {
    if (reduced) return;
    registerGsap();
    const section = sectionRef.current;
    const content = contentRef.current;
    if (!section || !content) return;

    const ctx = gsap.context(() => {
      gsap.to(content, {
        y: 80,
        opacity: 0.2,
        ease: "none",
        scrollTrigger: {
          trigger: section,
          start: "top top",
          end: "bottom top",
          scrub: true,
        },
      });
    }, section);

    return () => ctx.revert();
  }, [reduced]);

  return (
    <section ref={sectionRef} id="hero" className="relative flex min-h-screen items-center justify-center overflow-hidden">
      <div className="absolute inset-0">
        {useVideo && !reduced ? (
          <video
            ref={videoRef}
            className="h-full w-full object-cover scale-105"
            src={restaurant.videos.hero.src}
            poster={restaurant.heroPoster}
            muted
            loop
            playsInline
            preload="metadata"
            aria-hidden
          />
        ) : (
          <div className="hero-ken-burns relative h-full w-full">
            {restaurant.heroPoster && (
              <Image
                src={restaurant.heroPoster}
                alt=""
                fill
                priority
                className="object-cover object-center"
                sizes="100vw"
              />
            )}
          </div>
        )}
        <div className="absolute inset-0 bg-gradient-to-b from-[#10100e]/92 via-[#10100e]/72 to-[#10100e]" />
        <div className="absolute inset-x-0 top-0 h-32 bg-gradient-to-b from-[#10100e] to-transparent" />
      </div>

      <div ref={contentRef} className="relative z-10 mx-auto max-w-4xl px-6 pt-28 text-center">
        <p className="mb-4 text-xs font-medium uppercase tracking-[0.25em] text-[#b88a44] drop-shadow-md">
          {restaurant.cuisine}
        </p>
        <h1 className="font-display text-5xl leading-tight text-[#f7f0e6] drop-shadow-[0_2px_24px_rgba(0,0,0,0.8)] md:text-7xl lg:text-8xl">
          {restaurant.name}
        </h1>
        <p className="mx-auto mt-6 max-w-2xl text-lg text-[#f7f0e6]/85 drop-shadow-md">{restaurant.tagline}</p>
        <p className="mx-auto mt-3 max-w-xl text-sm text-[#f7f0e6]/65 drop-shadow-md">{restaurant.subheadline}</p>

        <div className="mt-6 flex flex-wrap items-center justify-center gap-3 text-[10px] uppercase tracking-[0.18em] text-cream/45">
          {restaurant.metadataLabels.map((label) => (
            <span key={label}>{label}</span>
          ))}
        </div>

        {restaurant.rating && (
          <p className="mt-4 text-sm text-brass">
            ★ {restaurant.rating} · {restaurant.reviewsCount || 0} reviews · {restaurant.priceLevel}
          </p>
        )}

        <div className="mt-10 flex flex-wrap justify-center gap-4">
          <a href={restaurant.primaryCTA.href} className="btn-primary">
            {restaurant.primaryCTA.label}
          </a>
          <a href={restaurant.secondaryCTA.href} className="btn-ghost">
            {restaurant.secondaryCTA.label}
          </a>
        </div>
      </div>

      <a href="#story" aria-label="Scroll down" className="absolute bottom-8 left-1/2 z-10 -translate-x-1/2">
        <span className="block h-12 w-px animate-pulse bg-gradient-to-b from-brass to-transparent" />
      </a>
    </section>
  );
}
