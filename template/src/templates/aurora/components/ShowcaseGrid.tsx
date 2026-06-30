"use client";

import { useState } from "react";
import Image from "next/image";
import { motion, AnimatePresence } from "framer-motion";
import BlurReveal from "./ui/BlurReveal";
import type { GalleryImage } from "@/data/types/gallery";

export default function ShowcaseGrid({ images }: { images: GalleryImage[] }) {
  const [modal, setModal] = useState<GalleryImage | null>(null);

  if (!images.length) return null;

  return (
    <section id="gallery" className="aurora-section">
      <div className="aurora-container">
        <BlurReveal className="text-center">
          <p className="text-xs uppercase tracking-[0.2em] text-purple-400">Ambience</p>
          <h2 className="aurora-heading mt-3 text-4xl font-bold text-white">Showcase Grid</h2>
        </BlurReveal>

        <div className="mt-12 columns-1 gap-4 sm:columns-2 lg:columns-3">
          {images.map((img, i) => (
            <BlurReveal key={img.url} delay={(i % 6) * 0.05}>
              <button
                type="button"
                onClick={() => setModal(img)}
                className="mb-4 block w-full overflow-hidden rounded-xl border border-white/10 transition hover:border-purple-500/30"
              >
                <div className="relative aspect-[4/5]">
                  <Image
                    src={img.url}
                    alt={img.alt}
                    fill
                    className="object-cover"
                    sizes="(max-width: 768px) 100vw, 33vw"
                  />
                </div>
              </button>
            </BlurReveal>
          ))}
        </div>
      </div>

      <AnimatePresence>
        {modal && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-[100] flex items-center justify-center bg-black/90 p-6"
            onClick={() => setModal(null)}
          >
            <div className="relative h-[80vh] w-full max-w-4xl" onClick={(e) => e.stopPropagation()}>
              <Image src={modal.url} alt={modal.alt} fill className="object-contain" sizes="90vw" />
              <button
                type="button"
                className="absolute -top-10 right-0 text-sm text-white/70"
                onClick={() => setModal(null)}
              >
                Close
              </button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </section>
  );
}
