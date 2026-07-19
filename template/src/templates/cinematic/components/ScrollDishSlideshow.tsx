"use client";

import { useEffect, useRef, useState } from "react";
import { gsap, registerGsap } from "@/lib/gsap";
import { usePrefersReducedMotion } from "@/hooks/usePrefersReducedMotion";
import { useIsMobile } from "@/hooks/useMediaQuery";
import type { MenuItem } from "@/data/types/menu";
import SourceAwareImage, { mediaForURL } from "@/components/SourceAwareImage";
import { isFoodMenuImage } from "../lib/menuImages";

export default function ScrollDishSlideshow({ dishes: rawDishes }: { dishes: MenuItem[] }) {
  const dishes = rawDishes.filter((d) => d.image && isFoodMenuImage(d.image));
  const sectionRef = useRef<HTMLElement>(null);
  const [active, setActive] = useState(0);
  const reduced = usePrefersReducedMotion();
  const isMobile = useIsMobile();

  useEffect(() => {
    if (reduced || isMobile || dishes.length < 2) return;
    registerGsap();

    const section = sectionRef.current;
    if (!section) return;

    const ctx = gsap.context(() => {
      gsap.timeline({
        scrollTrigger: {
          trigger: section,
          start: "top top",
          end: `+=${dishes.length * 70}%`,
          scrub: true,
          pin: true,
          onUpdate: (self) => {
            const idx = Math.min(
              dishes.length - 1,
              Math.floor(self.progress * dishes.length)
            );
            setActive(idx);
          },
        },
      });
    }, section);

    return () => ctx.revert();
  }, [reduced, isMobile, dishes.length]);

  if (!dishes.length) return null;

  const dish = dishes[active];

  return (
    <section ref={sectionRef} id="signatures" className="relative min-h-screen bg-[#141210]">
      <div className="grid min-h-screen grid-cols-1 md:grid-cols-2">
        <div className="relative min-h-[50vh] overflow-hidden md:min-h-screen">
          {dishes.map((d, i) => (
            <div
              key={d.name + i}
              className={`absolute inset-0 transition-all duration-700 ${
                i === active ? "opacity-100 scale-100" : "opacity-0 scale-[1.08]"
              }`}
            >
              {d.image && (
                <SourceAwareImage
                  media={mediaForURL(d.image, d.name, "food")}
                  fill
                  loading="lazy"
                  className="object-contain bg-[#141210] p-6"
                  sizes="(max-width: 768px) 100vw, 50vw"
                />
              )}
            </div>
          ))}
        </div>

        <div className="flex flex-col justify-center px-8 py-16 md:px-16">
          <p className="text-xs uppercase tracking-[0.2em] text-brass">Signature Dishes</p>
          <p className="mt-4 text-sm text-cream/50">
            {String(active + 1).padStart(2, "0")} / {String(dishes.length).padStart(2, "0")}
          </p>
          <h2 className="font-display mt-6 text-4xl text-cream md:text-5xl">{dish.name}</h2>
          <p className="mt-4 max-w-md text-cream/70">{dish.description}</p>
          {dish.price && <p className="mt-4 text-lg font-semibold text-brass">{dish.price}</p>}
          {dish.tags && dish.tags.length > 0 && (
            <div className="mt-4 flex flex-wrap gap-2">
              {dish.tags.map((tag) => (
                <span key={tag} className="rounded-full border border-cream/15 px-3 py-1 text-[10px] uppercase tracking-wider text-cream/60">
                  {tag}
                </span>
              ))}
            </div>
          )}
          <a href="#menu" className="btn-ghost mt-8 inline-flex w-fit">
            View Menu
          </a>

          {isMobile && (
            <div className="mt-8 flex gap-2">
              {dishes.map((_, i) => (
                <button
                  key={i}
                  type="button"
                  aria-label={`Show dish ${i + 1}`}
                  onClick={() => setActive(i)}
                  className={`h-2 flex-1 rounded-full ${i === active ? "bg-brass" : "bg-cream/20"}`}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </section>
  );
}
