"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import type { MediaCard, PlaceMedia } from "@/lib/places";

function cardImageSrc(card: MediaCard): string | null {
  if (card.imageUrl) return card.imageUrl;
  if (card.photoName) {
    return `/api/restaurants/photo?name=${encodeURIComponent(card.photoName)}&max=720`;
  }
  return null;
}

function mediaCardKey(card: MediaCard): string {
  return `${card.kind}|${cardImageSrc(card) || ""}`;
}

function MediaTile({
  card,
  onImageError,
}: {
  card: MediaCard;
  onImageError: (card: MediaCard) => void;
}) {
  const src = cardImageSrc(card);
  if (!src) return null;
  const inner = (
    <div className="relative h-[148px] w-[132px] shrink-0 overflow-hidden rounded-xl bg-[#e8e2da] sm:h-[168px] sm:w-[148px]">
      {/* Dynamic Google media cannot use next/image's fixed remote allow-list. */}
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={src}
        alt={card.label || "Restaurant listing media"}
        className="h-full w-full object-cover"
        loading="lazy"
        onError={() => onImageError(card)}
      />
      <div className="pointer-events-none absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/75 via-black/35 to-transparent px-2.5 pb-2.5 pt-10">
        {card.kind === "menu" ? (
          <span className="inline-flex items-center gap-1.5 rounded-md bg-[#2b2b2b]/92 px-2 py-1 text-[12px] font-semibold text-white">
            <svg viewBox="0 0 16 16" className="h-3.5 w-3.5 fill-current" aria-hidden="true">
              <path d="M3 2.5h10a.5.5 0 0 1 .5.5v10a.5.5 0 0 1-.5.5H3a.5.5 0 0 1-.5-.5V3a.5.5 0 0 1 .5-.5Zm1 2v1.5h8V4.5H4Zm0 3v1.5h5V7.5H4Zm0 3V12h7v-1.5H4Z" />
            </svg>
            Menu
          </span>
        ) : (
          <>
            <p className="truncate text-[13px] font-semibold leading-tight text-white">{card.label}</p>
            {card.subtitle ? (
              <p className="mt-0.5 truncate text-[11px] text-white/80">{card.subtitle}</p>
            ) : null}
          </>
        )}
      </div>
    </div>
  );

  if (card.href) {
    return (
      <a href={card.href} target="_blank" rel="noreferrer" className="shrink-0 snap-start">
        {inner}
      </a>
    );
  }
  return <div className="shrink-0 snap-start">{inner}</div>;
}

function MediaRow({
  title,
  cards,
  seeMoreHref,
  onImageError,
}: {
  title: string;
  cards: MediaCard[];
  seeMoreHref?: string;
  onImageError: (card: MediaCard) => void;
}) {
  const scrollerRef = useRef<HTMLDivElement>(null);
  if (!cards.length) return null;

  function scrollBy(dir: 1 | -1) {
    const el = scrollerRef.current;
    if (!el) return;
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    el.scrollBy({ left: dir * 280, behavior: reducedMotion ? "auto" : "smooth" });
  }

  return (
    <div>
      <div className="mb-3 flex items-center justify-between gap-3">
        <h3 className="text-[17px] font-bold tracking-[-0.02em] text-ink sm:text-[18px]">{title}</h3>
        {cards.length > 3 ? (
          <button
            type="button"
            onClick={() => scrollBy(1)}
            className="hidden h-11 w-11 shrink-0 cursor-pointer items-center justify-center rounded-full border border-border bg-bg text-ink shadow-sm transition-colors duration-200 hover:bg-[#f7f4ef] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary sm:inline-flex"
            aria-label={`Scroll ${title}`}
          >
            ›
          </button>
        ) : null}
      </div>
      <div
        ref={scrollerRef}
        className="flex gap-3 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] snap-x snap-mandatory [&::-webkit-scrollbar]:hidden"
      >
        {cards.map((card, i) => (
          <MediaTile
            key={`${mediaCardKey(card)}-${i}`}
            card={card}
            onImageError={onImageError}
          />
        ))}
      </div>
      {seeMoreHref ? (
        <div className="mt-3 text-center">
          <a
            href={seeMoreHref}
            target="_blank"
            rel="noreferrer"
            className="inline-flex min-h-11 items-center rounded-full px-3 text-[14px] font-semibold text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
          >
            See more
          </a>
        </div>
      ) : null}
    </div>
  );
}

export default function LiveListingMedia({ media }: { media?: PlaceMedia | null }) {
  const candidateCards = useMemo(
    () => [
      ...(media?.menuAndHighlights || []),
      ...(media?.photosAndVideos || []),
    ].filter((card) => Boolean(cardImageSrc(card))),
    [media],
  );
  const [readyImages, setReadyImages] = useState<Set<string>>(() => new Set());
  const [failedImages, setFailedImages] = useState<Set<string>>(() => new Set());

  useEffect(() => {
    let cancelled = false;
    const loaders = candidateCards.map((card) => {
      const key = mediaCardKey(card);
      const src = cardImageSrc(card);
      if (!src) return null;
      const image = new window.Image();
      image.onload = () => {
        if (!cancelled) setReadyImages((current) => new Set(current).add(key));
      };
      image.onerror = () => {
        if (!cancelled) setFailedImages((current) => new Set(current).add(key));
      };
      image.src = src;
      return image;
    });
    return () => {
      cancelled = true;
      for (const image of loaders) {
        if (!image) continue;
        image.onload = null;
        image.onerror = null;
      }
    };
  }, [candidateCards]);

  if (!media) return null;
  const highlights = (media.menuAndHighlights || []).filter(
    (card) => readyImages.has(mediaCardKey(card)) && !failedImages.has(mediaCardKey(card)),
  );
  const photos = (media.photosAndVideos || []).filter(
    (card) => readyImages.has(mediaCardKey(card)) && !failedImages.has(mediaCardKey(card)),
  );
  if (!highlights.length && !photos.length) return null;

  const handleImageError = (card: MediaCard) => {
    setFailedImages((current) => new Set(current).add(mediaCardKey(card)));
  };

  return (
    <section className="overflow-hidden rounded-[22px] border border-border bg-bg/90 shadow-[0_10px_36px_rgba(15,39,31,0.06)]">
      <header className="border-b border-border px-5 py-4 sm:px-6">
        <p className="text-[12px] font-semibold uppercase tracking-[0.08em] text-muted">07 · Listing media</p>
        <h2 className="mt-1.5 font-display text-[1.15rem] font-semibold tracking-[-0.02em] text-ink sm:text-[1.25rem]">
          {highlights.length > 0 ? "Menu, highlights & photos" : "Listing photos"}
        </h2>
        <p className="mt-1 text-[13px] text-muted">
          Loaded from the live Google listing.
        </p>
      </header>
      <div className="space-y-8 px-5 py-5 sm:px-6 sm:py-6">
        <MediaRow
          title="Menu & highlights"
          cards={highlights}
          seeMoreHref={media.mapsUri}
          onImageError={handleImageError}
        />
        <MediaRow
          title={photos.some((card) => card.kind === "video") ? "Photos & videos" : "Photos"}
          cards={photos}
          seeMoreHref={media.mapsUri}
          onImageError={handleImageError}
        />
      </div>
    </section>
  );
}
