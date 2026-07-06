"use client";

import GallerySlider from "@/components/GallerySlider";
import type { GalleryImage } from "@/data/types/gallery";

export default function AtmosphereGallery({ images }: { images: GalleryImage[] }) {
  return (
    <GallerySlider
      images={images}
      variant="cinematic"
      eyebrow="Ambience"
      title="Gallery"
      subtitle="What does it feel like to eat here?"
    />
  );
}
