"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Image from "next/image";
import { AnimatePresence, motion } from "framer-motion";
import type { GalleryImage } from "@/data/types/gallery";

const PER_SLIDE = 3;

type Variant = "cinematic" | "aurora";

function chunkImages(images: GalleryImage[], size: number): GalleryImage[][] {
  const chunks: GalleryImage[][] = [];
  for (let i = 0; i < images.length; i += size) {
    chunks.push(images.slice(i, i + size));
  }
  return chunks;
}

function Chevron({ dir }: { dir: "left" | "right" }) {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      {dir === "left" ? <path d="M15 18 9 12l6-6" /> : <path d="M9 18 15 12 9 6" />}
    </svg>
  );
}

export default function GallerySlider({
  images,
  variant,
  eyebrow = "Ambience",
  title = "Gallery",
  subtitle,
  hideHeader = false,
  imageFit = "cover",
}: {
  images: GalleryImage[];
  variant: Variant;
  eyebrow?: string;
  title?: string;
  subtitle?: string;
  hideHeader?: boolean;
  imageFit?: "cover" | "contain";
}) {
  const slides = useMemo(() => chunkImages(images, PER_SLIDE), [images]);
  const [index, setIndex] = useState(0);
  const [modal, setModal] = useState<GalleryImage | null>(null);
  const [direction, setDirection] = useState(1);

  const go = useCallback(
    (delta: number) => {
      if (slides.length <= 1) return;
      setDirection(delta);
      setIndex((i) => (i + delta + slides.length) % slides.length);
    },
    [slides.length],
  );

  useEffect(() => {
    if (slides.length <= 1) return;
    const timer = setInterval(() => go(1), 6000);
    return () => clearInterval(timer);
  }, [go, slides.length]);

  if (!images.length) return null;

  const isCinematic = variant === "cinematic";
  const current = slides[index] ?? [];

  const sectionClass = isCinematic ? "bg-charcoal py-24" : "aurora-section";
  const containerClass = isCinematic ? "mx-auto max-w-6xl px-6" : "aurora-container";
  const eyebrowClass = isCinematic
    ? "text-xs uppercase tracking-[0.2em] text-brass"
    : "text-xs uppercase tracking-[0.2em] text-purple-400";
  const titleClass = isCinematic
    ? "font-display mt-3 text-4xl text-cream md:text-5xl"
    : "aurora-heading mt-3 text-4xl font-bold text-white";
  const subtitleClass = isCinematic ? "mt-3 text-cream/60" : "mt-3 text-white/50";
  const navBtnClass = isCinematic
    ? "flex h-11 w-11 items-center justify-center rounded-full border border-cream/20 text-cream transition hover:border-brass/50 hover:text-brass disabled:opacity-30"
    : "flex h-11 w-11 items-center justify-center rounded-full border border-white/15 text-white transition hover:border-purple-500/40 hover:text-purple-300 disabled:opacity-30";
  const dotActive = isCinematic ? "w-6 bg-brass" : "w-6 bg-purple-400";
  const dotIdle = isCinematic ? "bg-cream/25 hover:bg-cream/40" : "bg-white/20 hover:bg-white/35";
  const cardClass = isCinematic
    ? `relative aspect-[4/5] overflow-hidden rounded-xl border border-cream/10${
        imageFit === "contain" ? " bg-[#1a1614]" : ""
      }`
    : "relative aspect-[4/5] overflow-hidden rounded-xl border border-white/10 transition hover:border-purple-500/30";

  return (
    <section id="gallery" className={sectionClass}>
      <div className={containerClass}>
        {!hideHeader && (
          <div className="text-center">
            <p className={eyebrowClass}>{eyebrow}</p>
            <h2 className={titleClass}>{title}</h2>
            {subtitle ? <p className={subtitleClass}>{subtitle}</p> : null}
          </div>
        )}

        <div className={hideHeader ? "relative" : "relative mt-12"}>
          <div className="overflow-hidden">
            <AnimatePresence mode="wait" initial={false}>
              <motion.div
                key={index}
                initial={{ opacity: 0, x: direction > 0 ? 48 : -48 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: direction > 0 ? -48 : 48 }}
                transition={{ duration: 0.45, ease: [0.22, 1, 0.36, 1] }}
                className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3"
              >
                {current.map((img) => (
                  <button
                    key={img.url}
                    type="button"
                    onClick={() => setModal(img)}
                    className={`${cardClass} group`}
                  >
                    <Image
                      src={img.url}
                      alt={img.alt}
                      fill
                      loading="lazy"
                      className={
                        imageFit === "contain"
                          ? "object-contain p-2"
                          : `object-cover transition duration-500 ${isCinematic ? "group-hover:scale-105" : ""}`
                      }
                      sizes="(max-width: 768px) 100vw, 33vw"
                    />
                  </button>
                ))}
              </motion.div>
            </AnimatePresence>
          </div>

          {slides.length > 1 && (
            <div className="mt-8 flex items-center justify-center gap-4">
              <button
                type="button"
                aria-label="Previous images"
                onClick={() => go(-1)}
                className={navBtnClass}
              >
                <Chevron dir="left" />
              </button>

              <div className="flex items-center gap-2">
                {slides.map((_, i) => (
                  <button
                    key={i}
                    type="button"
                    aria-label={`Go to slide ${i + 1}`}
                    onClick={() => {
                      setDirection(i > index ? 1 : -1);
                      setIndex(i);
                    }}
                    className={`h-2 rounded-full transition-all ${i === index ? dotActive : `w-2 ${dotIdle}`}`}
                  />
                ))}
              </div>

              <button
                type="button"
                aria-label="Next images"
                onClick={() => go(1)}
                className={navBtnClass}
              >
                <Chevron dir="right" />
              </button>
            </div>
          )}

          <p className={`mt-4 text-center text-xs ${isCinematic ? "text-cream/40" : "text-white/35"}`}>
            Showing {index * PER_SLIDE + 1}–{Math.min((index + 1) * PER_SLIDE, images.length)} of{" "}
            {images.length}
          </p>
        </div>
      </div>

      <AnimatePresence>
        {modal && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className={`fixed inset-0 z-[100] flex items-center justify-center p-6 ${
              isCinematic ? "bg-charcoal/90" : "bg-black/90"
            }`}
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
                className={`absolute -top-10 right-0 text-sm uppercase tracking-wider ${
                  isCinematic ? "text-cream" : "text-white/70"
                }`}
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
