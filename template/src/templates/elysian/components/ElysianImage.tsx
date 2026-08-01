"use client";

import { forwardRef, type CSSProperties, type MouseEventHandler, type Ref } from "react";
import type { GalleryImage } from "@/data/types/gallery";
import SourceAwareImage, { mediaForURL } from "@/components/SourceAwareImage";

type ElysianImageProps = {
  src: string;
  alt: string;
  className?: string;
  style?: CSSProperties;
  width?: number;
  height?: number;
  fill?: boolean;
  sizes?: string;
  priority?: boolean;
  id?: string;
  onClick?: () => void;
  onMouseMove?: MouseEventHandler<HTMLDivElement>;
  media?: GalleryImage;
};

/** Source-aware remote image: live Google bypasses caching; durable CDN media is optimized. */
const ElysianImage = forwardRef(function ElysianImage(
  {
    src,
    alt,
    className,
    style,
    width,
    height,
    fill,
    sizes,
    priority,
    id,
    onClick,
    onMouseMove,
    media,
  }: ElysianImageProps,
  ref: Ref<HTMLDivElement>,
) {
  const resolvedMedia = media || mediaForURL(src, alt);
  if (fill) {
    return (
      <div
        ref={ref}
        className={className}
        style={{ position: "relative", width: "100%", height: "100%", ...style }}
        id={id}
        onClick={onClick}
        onMouseMove={onMouseMove}
      >
        <SourceAwareImage
          media={resolvedMedia}
          fill
          className="object-cover"
          sizes={sizes || "100vw"}
          priority={priority}
        />
      </div>
    );
  }

  return (
    <SourceAwareImage
      media={resolvedMedia}
      width={width!}
      height={height!}
      className={className}
      style={style}
      sizes={sizes}
      priority={priority}
      id={id}
      onClick={onClick}
    />
  );
});

export default ElysianImage;
