"use client";

import GallerySlider from "@/components/GallerySlider";
import type { GalleryImage } from "@/data/types/gallery";

export default function ShowcaseGrid({ images }: { images: GalleryImage[] }) {
  return (
    <GallerySlider
      images={images}
      variant="aurora"
      eyebrow="Ambience"
      title="In the spotlight"
      subtitle="A glimpse of the atmosphere, plates, and moments that define us."
    />
  );
}
