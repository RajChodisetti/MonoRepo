"use client";

import { useEffect, useRef } from "react";
import { gsap, registerGsap, ScrollTrigger } from "@/lib/gsap";
import { usePrefersReducedMotion } from "@/hooks/usePrefersReducedMotion";
import type { RestaurantContent } from "@/data/types/restaurant";
import type { GalleryImage } from "@/data/types/gallery";
import SourceAwareImage from "@/components/SourceAwareImage";
import PhotoAttribution from "@/components/PhotoAttribution";

export default function HeroRestaurantVideo({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const sectionRef = useRef<HTMLElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const reduced = usePrefersReducedMotion();

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

  const heroMedia: GalleryImage = restaurant.heroMedia || {
    url: restaurant.heroPoster,
    alt: `${restaurant.name} restaurant`,
    type: "ambience",
  };
  const supportingMedia = restaurant.galleryImages
    .filter((image) => image.url !== heroMedia.url && image.sourceKind !== "google_places_live")
    .slice(0, 3);

  return (
    <section ref={sectionRef} id="hero" className="relative flex min-h-screen items-center justify-center overflow-hidden">
      <div className="absolute inset-0">
        <div className="hero-ken-burns relative h-full w-full">
          {heroMedia.url ? (
            <SourceAwareImage
              media={heroMedia}
              fill
              priority
              className="object-cover object-center"
              sizes="100vw"
            />
          ) : null}
        </div>
        <div className="absolute inset-0 bg-gradient-to-b from-[#10100e]/92 via-[#10100e]/72 to-[#10100e]" />
        <div className="absolute inset-x-0 top-0 h-32 bg-gradient-to-b from-[#10100e] to-transparent" />
      </div>

      {supportingMedia.length ? (
        <div className="absolute bottom-8 right-6 z-20 hidden gap-2 lg:flex">
          {supportingMedia.map((image) => (
            <div key={image.url} className="relative h-20 w-28 overflow-hidden rounded-lg border border-cream/20">
              <SourceAwareImage media={image} fill className="object-cover" sizes="112px" />
            </div>
          ))}
        </div>
      ) : null}

      {heroMedia.sourceKind === "google_places_live" ? (
        <div className="absolute bottom-8 left-6 z-20 rounded bg-black/55 px-3 py-2 text-cream/70">
          <PhotoAttribution media={heroMedia} compact />
        </div>
      ) : null}

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
