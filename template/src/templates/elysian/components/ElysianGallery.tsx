"use client";

import { useGalleryLightbox } from "../hooks/useGalleryLightbox";
import type { GalleryImage } from "@/data/types/gallery";
import ElysianImage from "./ElysianImage";

export default function ElysianGallery({ images }: { images: GalleryImage[] }) {
  const { open, src, alt, openImage, close } = useGalleryLightbox();

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
          <div className="masonry" id="masonryGrid">
            {images.map((img, i) => (
              <div
                key={img.url + i}
                className="masonry-item reveal fade-up"
                onClick={() => openImage(img.url, img.alt)}
              >
                <ElysianImage
                  src={img.url}
                  alt={img.alt}
                  width={900}
                  height={600}
                  sizes="(max-width: 768px) 50vw, 33vw"
                  className="masonry-photo"
                />
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
          <ElysianImage src={src} alt={alt} width={1200} height={800} className="lightbox-photo" id="lightboxImg" />
        ) : null}
      </div>
    </>
  );
}
