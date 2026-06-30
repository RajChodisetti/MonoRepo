"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import { motion, AnimatePresence } from "framer-motion";
import { gsap, registerGsap } from "@/lib/gsap";
import { usePrefersReducedMotion } from "@/hooks/usePrefersReducedMotion";
import type { GalleryImage } from "@/data/types/gallery";

const LAYOUTS = [
  "col-span-7 row-span-2",
  "col-span-5 row-span-2",
  "col-span-4 row-span-1",
  "col-span-4 row-span-1",
  "col-span-4 row-span-1",
  "col-span-6 row-span-1",
  "col-span-6 row-span-1",
];

export default function AtmosphereGallery({
  images,
}: {
  images: GalleryImage[];
}) {
  const gridRef = useRef<HTMLDivElement>(null);
  const [modal, setModal] = useState<GalleryImage | null>(null);
  const reduced = usePrefersReducedMotion();

  useEffect(() => {
    if (reduced || !images.length) return;
    registerGsap();
    const grid = gridRef.current;
    if (!grid) return;

    const items = grid.querySelectorAll("[data-gallery-item]");
    const ctx = gsap.context(() => {
      items.forEach((item) => {
        gsap.fromTo(
          item,
          { y: 40 },
          {
            y: -20,
            ease: "none",
            scrollTrigger: {
              trigger: item,
              start: "top bottom",
              end: "bottom top",
              scrub: true,
            },
          }
        );
      });
    }, grid);

    return () => ctx.revert();
  }, [reduced, images.length]);

  if (!images.length) return null;

  return (
    <section id="gallery" className="bg-charcoal py-24">
      <div className="mx-auto max-w-6xl px-6">
        <div className="text-center">
          <p className="text-xs uppercase tracking-[0.2em] text-brass">Ambience</p>
          <h2 className="font-display mt-3 text-4xl text-cream md:text-5xl">Gallery</h2>
          <p className="mt-3 text-cream/60">What does it feel like to eat here?</p>
        </div>

        <div
          ref={gridRef}
          className="mt-12 grid auto-rows-[180px] grid-cols-12 gap-3 md:auto-rows-[220px]"
        >
          {images.map((img, i) => (
            <button
              key={img.url}
              type="button"
              data-gallery-item
              onClick={() => setModal(img)}
              className={`relative overflow-hidden rounded-xl ${LAYOUTS[i % LAYOUTS.length]}`}
            >
              <Image
                src={img.url}
                alt={img.alt}
                fill
                loading="lazy"
                className="object-cover transition duration-500 hover:scale-105"
                sizes="(max-width: 768px) 100vw, 33vw"
              />
            </button>
          ))}
        </div>
      </div>

      <AnimatePresence>
        {modal && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-[100] flex items-center justify-center bg-charcoal/90 p-6"
            onClick={() => setModal(null)}
          >
            <motion.div
              initial={{ scale: 0.95 }}
              animate={{ scale: 1 }}
              exit={{ scale: 0.95 }}
              className="relative h-[70vh] w-full max-w-4xl"
              onClick={(e) => e.stopPropagation()}
            >
              <Image src={modal.url} alt={modal.alt} fill className="object-contain" sizes="90vw" />
              <button
                type="button"
                className="absolute -top-10 right-0 text-sm uppercase tracking-wider text-cream"
                onClick={() => setModal(null)}
              >
                Close
              </button>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </section>
  );
}
