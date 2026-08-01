"use client";

import { useMemo, useState } from "react";
import { useGalleryLightbox } from "../hooks/useGalleryLightbox";
import type { GalleryImage } from "@/data/types/gallery";
import ElysianImage from "./ElysianImage";
import PhotoAttribution from "@/components/PhotoAttribution";

export default function ElysianGallery({ images }: { images: GalleryImage[] }) {
  const { open, src, alt, openImage, close } = useGalleryLightbox();
  const [filter, setFilter] = useState<"all" | "food" | "ambience" | "other">("all");
  const visibleImages = useMemo(
    () => (filter === "all" ? images : images.filter((image) => image.type === filter)),
    [filter, images],
  );
  const activeImage = images.find((image) => image.url === src);

  if (!images.length) return null;

  return (
    <>
      <section className="gallery section" id="gallery">
        <div className="container">
          <div className="section-head reveal fade-up">
            <p className="eyebrow">Gallery</p>
            <h2 className="section-title">
              The <span className="gold-text">Atmosphere</span>
            </h2>
          </div>
          <div className="gallery-filters" role="group" aria-label="Gallery categories">
            {(["all", "food", "ambience", "other"] as const)
              .filter((item) => item === "all" || images.some((image) => image.type === item))
              .map((item) => (
                <button
                  key={item}
                  type="button"
                  className={`gallery-filter${filter === item ? " active" : ""}`}
                  aria-pressed={filter === item}
                  onClick={() => setFilter(item)}
                >
                  {item === "all" ? "All photos" : item === "food" ? "Food & drink" : item === "ambience" ? "Space & atmosphere" : "More"}
                </button>
              ))}
          </div>
          <div className="masonry" id="masonryGrid">
            {visibleImages.map((img, i) => (
              <div
                key={img.url + i}
                className="masonry-item reveal fade-up"
                onClick={() => openImage(img.url, img.alt)}
              >
                <ElysianImage
                  src={img.url}
                  alt={img.alt}
                  media={img}
                  width={900}
                  height={600}
                  sizes="(max-width: 768px) 50vw, 33vw"
                  className="masonry-photo"
                />
                {img.sourceKind === "google_places_live" ? (
                  <div className="gallery-card-attribution" onClick={(event) => event.stopPropagation()}>
                    <PhotoAttribution media={img} compact />
                  </div>
                ) : null}
              </div>
            ))}
          </div>
        </div>
      </section>

      <div className={`lightbox${open ? " open" : ""}`} id="lightbox">
        <button
          type="button"
          className="lightbox-close"
          id="lightboxClose"
          aria-label="Close gallery preview"
          onClick={close}
        >
          &times;
        </button>
        {src ? (
          <>
            <ElysianImage
              src={src}
              alt={alt}
              media={activeImage}
              width={1200}
              height={800}
              className="lightbox-photo"
              id="lightboxImg"
            />
            {activeImage ? (
              <div className="lightbox-attribution">
                <PhotoAttribution media={activeImage} />
              </div>
            ) : null}
          </>
        ) : null}
      </div>
    </>
  );
}
