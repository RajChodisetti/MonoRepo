"use client";

import { useRef } from "react";
import type { MediaCard, PlaceMedia } from "@/lib/places";

function cardImageSrc(card: MediaCard): string | null {
  if (card.imageUrl) return card.imageUrl;
  if (card.photoName) {
    return `/api/restaurants/photo?name=${encodeURIComponent(card.photoName)}&max=720`;
  }
  return null;
}

function safeExternalHref(raw?: string): string | undefined {
  const value = (raw || "").trim();
  if (!value) return undefined;
  try {
    const parsed = new URL(value.startsWith("//") ? `https:${value}` : value);
    return parsed.protocol === "http:" || parsed.protocol === "https:"
      ? parsed.toString()
      : undefined;
  } catch {
    return undefined;
  }
}

function MediaTile({ card }: { card: MediaCard }) {
  const src = cardImageSrc(card);
  const sourceHref = safeExternalHref(card.googleMapsUri || card.href);
  const flagContentHref = safeExternalHref(card.flagContentUri);
  const attributions = card.authorAttributions || [];

  return (
    <article className="w-[148px] shrink-0 snap-start overflow-hidden rounded-xl border border-border bg-white sm:w-[168px]">
      <div className="relative h-[148px] w-full overflow-hidden bg-[#e8e2da] sm:h-[168px]">
        {src ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={src}
            alt={`${card.label}${card.subtitle ? ` by ${card.subtitle}` : ""}`}
            className="h-full w-full object-cover"
            loading="lazy"
            referrerPolicy="no-referrer"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center bg-[#ddd6cc] text-[13px] font-semibold text-[#6a635c]">
            {card.label}
          </div>
        )}
        <div className="pointer-events-none absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/75 via-black/35 to-transparent px-2.5 pb-2.5 pt-10">
          <p className="truncate text-[13px] font-semibold leading-tight text-white">{card.label}</p>
          {card.subtitle ? (
            <p className="mt-0.5 line-clamp-2 text-[12px] leading-snug text-white/85">{card.subtitle}</p>
          ) : null}
        </div>
      </div>
      <div className="min-h-[90px] px-2.5 py-2 text-[12px] leading-snug text-muted">
        {attributions.length > 0 ? (
          <div className="flex flex-wrap items-center gap-x-1.5 gap-y-1">
            <span>Photo by</span>
            {attributions.map((attribution, index) => {
              const authorHref = safeExternalHref(attribution.uri);
              const authorPhotoHref = safeExternalHref(attribution.photoUri);
              const name = attribution.displayName || "Google contributor";
              return (
                <span key={`${name}-${index}`} className="inline-flex items-center gap-1">
                  {index > 0 ? <span aria-hidden="true">·</span> : null}
                  {authorPhotoHref ? (
                    // Google photo contributor avatar stays live and unoptimized.
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={authorPhotoHref}
                      alt=""
                      className="h-4 w-4 rounded-full object-cover"
                      referrerPolicy="no-referrer"
                    />
                  ) : null}
                  {authorHref ? (
                    <a
                      href={authorHref}
                      target="_blank"
                      rel="noreferrer"
                      className="font-semibold text-primary underline decoration-primary/30 underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary"
                    >
                      {name}
                    </a>
                  ) : (
                    name
                  )}
                </span>
              );
            })}
          </div>
        ) : (
          <p>{card.photoName ? "Google contributor not supplied" : "Public listing source"}</p>
        )}
        <div className="mt-2 flex flex-wrap gap-x-2 gap-y-1 font-semibold">
          {sourceHref ? (
            <a
              href={sourceHref}
              target="_blank"
              rel="noreferrer"
              className="text-primary underline decoration-primary/30 underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary"
            >
              {card.googleMapsUri ? "Google Maps" : "View source"}
            </a>
          ) : null}
          {flagContentHref ? (
            <a
              href={flagContentHref}
              target="_blank"
              rel="noreferrer"
              className="underline underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary"
            >
              Report photo
            </a>
          ) : null}
        </div>
      </div>
    </article>
  );
}

function MediaRow({
  title,
  cards,
  seeMoreHref,
}: {
  title: string;
  cards: MediaCard[];
  seeMoreHref?: string;
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
          <MediaTile key={`${card.kind}-${card.label}-${i}`} card={card} />
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
  if (!media) return null;
  const highlights = media.menuAndHighlights || [];
  const photos = media.photosAndVideos || [];
  if (!highlights.length && !photos.length) return null;
  const googleMapsHref = safeExternalHref(media.mapsUri) || "https://www.google.com/maps";

  return (
    <section className="overflow-hidden rounded-[22px] border border-border bg-bg/90 shadow-[0_10px_36px_rgba(15,39,31,0.06)]">
      <header className="border-b border-border px-5 py-4 sm:px-6">
        <p className="text-[12px] font-semibold uppercase tracking-[0.08em] text-muted">07 · Listing media</p>
        <h2 className="mt-1.5 font-display text-[1.15rem] font-semibold tracking-[-0.02em] text-ink sm:text-[1.25rem]">
          Listing photos &amp; public media
        </h2>
        <p className="mt-1 text-[13px] text-muted">
          Live photo evidence from the venue listing. Google Places does not identify the Maps
          interface&apos;s Menu photo category, so only a confirmed public menu link earns menu points.
        </p>
        <a
          href={googleMapsHref}
          target="_blank"
          rel="noreferrer"
          className="mt-2 inline-block text-[12px] font-normal text-[#5e5e5e] underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
          translate="no"
        >
          Google Maps
        </a>
      </header>
      <div className="space-y-8 px-5 py-5 sm:px-6 sm:py-6">
        <MediaRow title="Listing highlights" cards={highlights} seeMoreHref={media.mapsUri} />
        <MediaRow title="Listing photos" cards={photos} seeMoreHref={media.mapsUri} />
      </div>
    </section>
  );
}
