"use client";

import { useState } from "react";
import Image, { type ImageProps } from "next/image";
import type { GalleryImage } from "@/data/types/gallery";

type Props = Omit<ImageProps, "src" | "alt" | "unoptimized"> & {
  media: Pick<GalleryImage, "url" | "alt" | "unoptimized">;
  fallbackClassName?: string;
};

export function mediaForURL(url: string, alt: string, type: GalleryImage["type"] = "other"): GalleryImage {
  let hostname = "";
  try {
    hostname = new URL(url).hostname;
  } catch {
    hostname = "";
  }
  const googleLive = hostname === "lh3.googleusercontent.com" || hostname.endsWith(".googleusercontent.com");
  return {
    url,
    alt,
    type,
    sourceKind: googleLive ? "google_places_live" : "legacy_public_url",
    unoptimized: googleLive,
  };
}

// Google Places media bypasses the Next.js optimizer so short-lived provider
// content is not cached or re-hosted. Owned/licensed assets retain optimization.
export default function SourceAwareImage({ media, fallbackClassName, ...props }: Props) {
  const [failed, setFailed] = useState(false);
  if (failed || !media.url) {
    return (
      <div
        className={fallbackClassName || "h-full w-full bg-gradient-to-br from-neutral-800 to-neutral-950"}
        role="img"
        aria-label={media.alt}
      />
    );
  }
  return (
    <Image
      {...props}
      src={media.url}
      alt={media.alt}
      unoptimized={Boolean(media.unoptimized)}
      onError={() => setFailed(true)}
    />
  );
}
