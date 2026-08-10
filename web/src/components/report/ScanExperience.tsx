"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  EVIDENCE_CARD_ENTRY_MS,
  LISTING_CARD_COUNT,
  PHOTO_FLIP_MS,
  REVIEW_PAGE_SIZE,
  REVIEW_WALL_HOLD_MS,
  REVIEW_WALL_LIMIT,
  TARGET_SCAN_SECONDS,
  dealPhotosToCards,
  evidenceBatchPresentationMs,
  evidenceCardEntryDelayMs,
  isScanCompletionReady,
  nextEvidenceBatchStart,
  nextPhotoFaceIndex,
  photoFlipCycleMs,
  satellitePositions,
} from "@/lib/scan-timeline";
import type { AuthorAttribution, PlaceAttribution } from "@/lib/places";

// A floor under a step's dwell so an empty step (nothing to show) still gets
// a readable beat before the checklist moves on.
const MIN_STEP_MS = 2_200;
const FINISH_HOLD_MS = 320;
const REVIEW_RELEVANCE_NOTICE =
  "Google Maps supplies up to five reviews ordered by relevance. Tuvi shows that relevance-ordered sample and may visually shorten long text in this scan.";

export type ScanPhoto = {
  src: string;
  label?: string;
  sourceLabel?: string;
  sourceUrl?: string;
  authorAttributions?: AuthorAttribution[];
  googleMapsUri?: string;
  flagContentUri?: string;
  alt?: string;
};

export type ScanReview = {
  author?: string;
  authorUri?: string;
  authorPhotoUri?: string;
  googleMapsUri?: string;
  flagContentUri?: string;
  text?: string;
  rating?: number;
  relativeTime?: string;
  publishTime?: string;
  visitDate?: {
    year?: number;
    month?: number;
    day?: number;
  };
  sentiment?: string;
};

export type ScanExperienceProps = {
  restaurantName?: string;
  address?: string;
  rating?: number;
  category?: string;
  website?: string;
  placeId?: string;
  mapsUri?: string;
  placeAttributions?: PlaceAttribution[];
  latitude?: number;
  longitude?: number;
  photoUrl?: string;
  photos?: ScanPhoto[];
  /** Live Google reviews shown during the review-sentiment step. */
  reviews?: ScanReview[];
  /** Desktop viewport screenshot of the restaurant website (data URL or http). */
  desktopScreenshot?: string;
  /** Mobile viewport screenshot of the restaurant website (data URL or http). */
  mobileScreenshot?: string;
  /** Live restaurant website URL used only for source labels and unavailable states. */
  websiteUrl?: string;
  fetchComplete?: boolean;
  onReady?: () => void;
  className?: string;
};

function CheckIcon() {
  return (
    <svg viewBox="0 0 16 16" className="h-3.5 w-3.5" fill="none" aria-hidden="true">
      <path
        d="M3.5 8.2 6.4 11l6.1-6.5"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function Spinner() {
  return (
    <svg className="scan-spinner h-4 w-4 text-primary" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeOpacity="0.2" strokeWidth="2.5" />
      <path d="M21 12a9 9 0 0 0-9-9" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" />
    </svg>
  );
}

function Stars({ rating = 0 }: { rating?: number }) {
  const full = Math.max(0, Math.min(5, Math.round(rating)));
  return (
    <span className="inline-flex items-center gap-0.5 text-[#f4b400]" aria-label={`${rating} stars`}>
      {Array.from({ length: 5 }).map((_, i) => (
        <svg key={i} viewBox="0 0 12 12" className="h-2.5 w-2.5" fill={i < full ? "currentColor" : "none"} aria-hidden="true">
          <path
            d="M6 1.2 7.4 4.2 10.7 4.6 8.3 6.9 8.9 10.2 6 8.6 3.1 10.2 3.7 6.9 1.3 4.6 4.6 4.2 6 1.2Z"
            stroke="currentColor"
            strokeWidth="0.8"
          />
        </svg>
      ))}
    </span>
  );
}

function sentimentMeta(sentiment: string) {
  const s = sentiment.toLowerCase();
  if (s === "positive") return { label: "Positive", className: "bg-[#e8f6ee] text-[#1f7a45]" };
  if (s === "negative") return { label: "Negative", className: "bg-[#fdecea] text-[#b42318]" };
  return { label: "Mixed", className: "bg-[#fff6e5] text-[#a15c00]" };
}

type ScanEvidence = {
  id: string;
  kind: "listing" | "desktop" | "mobile" | "website" | "competitor";
  label: string;
  sourceLabel: string;
  alt: string;
  src?: string;
  sourceUrl?: string;
  authorAttributions?: AuthorAttribution[];
  googleMapsUri?: string;
  flagContentUri?: string;
  /**
   * Every decoded listing photo this one card turns through. Populated only for
   * listing evidence; a single-entry rotation simply never flips.
   */
  photoRotation?: ScanPhoto[];
  /** Live Google Maps nearby search shown inside the competitor card. */
  competitorMapSrc?: string;
};

type ImageLoadStatus = "loaded" | "failed";

function safeHttpUrl(raw?: string): string | undefined {
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

function reviewVisitLabel(visitDate?: ScanReview["visitDate"]): string | undefined {
  const year = visitDate?.year;
  const month = visitDate?.month;
  if (!Number.isInteger(year) || !year || year < 1) return undefined;
  if (!Number.isInteger(month) || !month || month < 1 || month > 12) {
    return `Visited ${year}`;
  }
  const value = new Date(Date.UTC(year, month - 1, 1));
  return `Visited ${new Intl.DateTimeFormat("en", {
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  }).format(value)}`;
}

function BrowserBar({ viewport }: { viewport: "desktop" | "mobile" | "neutral" }) {
  return (
    <div className="flex h-6 items-center gap-1.5 border-b border-black/5 bg-[#f4f1ed] px-2" aria-hidden="true">
      {viewport === "mobile" ? (
        <span className="mx-auto h-1.5 w-10 rounded-full bg-ink/20" />
      ) : viewport === "neutral" ? (
        <span className="mx-auto h-2.5 w-2/3 rounded-full bg-white" />
      ) : (
        <>
          <span className="h-1.5 w-1.5 rounded-full bg-[#ef7d74]" />
          <span className="h-1.5 w-1.5 rounded-full bg-[#e9bd63]" />
          <span className="h-1.5 w-1.5 rounded-full bg-[#73b881]" />
          <span className="ml-1 h-2.5 flex-1 rounded-full bg-white" />
        </>
      )}
    </div>
  );
}

/**
 * One card, many real photos. The face turning away keeps its photo until the
 * rotation lands, so a flip never flashes the next image early.
 */
function PhotoFlipStage({
  photos,
  active,
  onImageError,
}: {
  photos: ScanPhoto[];
  active: boolean;
  onImageError: (src: string) => void;
}) {
  const photoCount = photos.length;
  const [flip, setFlip] = useState(0);
  const [faces, setFaces] = useState<[number, number]>([0, photoCount > 1 ? 1 : 0]);

  useEffect(() => {
    if (photoCount < 2) return;
    const id = window.setInterval(() => setFlip((current) => current + 1), photoFlipCycleMs());
    return () => window.clearInterval(id);
  }, [photoCount]);

  // Refresh the face that just turned away, ready for the next flip.
  useEffect(() => {
    if (photoCount < 2) return;
    const id = window.setTimeout(() => {
      const upcoming = nextPhotoFaceIndex(flip, photoCount);
      setFaces((current) =>
        flip % 2 === 0 ? [current[0], upcoming] : [upcoming, current[1]],
      );
    }, PHOTO_FLIP_MS);
    return () => window.clearTimeout(id);
  }, [flip, photoCount]);

  const front = photos[faces[0] % photoCount];
  const back = photos[faces[1] % photoCount];
  const showing = (flip % photoCount) + 1;

  return (
    <div className={`${active ? "h-44 sm:h-52" : "h-24 sm:h-28"} scan-photo-flip-stage relative bg-[#e9e4dc]`}>
      <div
        className="scan-photo-flip-inner"
        style={
          {
            transform: `rotateY(${flip * 180}deg)`,
            "--scan-photo-flip-ms": `${PHOTO_FLIP_MS}ms`,
          } as React.CSSProperties
        }
      >
        <div className="scan-photo-flip-face">
          {/* Evidence is rendered only from URLs received in the report payload. */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={front.src}
            alt={front.alt || ""}
            className="h-full w-full object-cover object-top"
            onError={() => onImageError(front.src)}
          />
        </div>
        <div className="scan-photo-flip-face scan-photo-flip-face-back">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={back.src}
            alt=""
            className="h-full w-full object-cover object-top"
            onError={() => onImageError(back.src)}
          />
        </div>
      </div>
      {active && photoCount > 1 ? (
        <span className="absolute bottom-2 right-2 z-10 rounded-full bg-ink/70 px-2 py-0.5 text-[10px] font-semibold tabular-nums text-bg">
          {showing} / {photoCount}
        </span>
      ) : null}
    </div>
  );
}

/** Desktop captures keep browser chrome; mobile captures sit in a phone shell. */
function DeviceFrame({
  kind,
  active,
  children,
}: {
  kind: "desktop" | "mobile" | "website";
  active: boolean;
  children: React.ReactNode;
}) {
  if (kind === "mobile") {
    return (
      <div className={`${active ? "h-44 sm:h-52" : "h-24 sm:h-28"} flex items-center justify-center bg-[#eae5dd] py-2`}>
        <div className="relative h-full overflow-hidden rounded-[0.9rem] border-[3px] border-[#1a1a1a] bg-[#1a1a1a] shadow-[0_10px_24px_rgba(15,39,31,0.28)]">
          <span
            className="absolute left-1/2 top-1 z-10 h-[3px] w-6 -translate-x-1/2 rounded-full bg-white/40"
            aria-hidden="true"
          />
          <div className="h-full overflow-hidden rounded-[0.7rem] bg-white">{children}</div>
        </div>
      </div>
    );
  }
  return <div className={`${active ? "h-44 sm:h-52" : "h-24 sm:h-28"} bg-[#e9e4dc]`}>{children}</div>;
}

function EvidenceCard({
  evidence,
  active,
  onImageError,
}: {
  evidence: ScanEvidence;
  active: boolean;
  onImageError: (src: string) => void;
}) {
  const isWebsite =
    evidence.kind === "desktop" ||
    evidence.kind === "mobile" ||
    evidence.kind === "website";
  const sourceUrl = safeHttpUrl(evidence.sourceUrl);
  const googleMapsUri = safeHttpUrl(evidence.googleMapsUri);
  const flagContentUri = safeHttpUrl(evidence.flagContentUri);
  const sourceLinks = [googleMapsUri, sourceUrl].filter(
    (value, index, links): value is string => Boolean(value) && links.indexOf(value) === index,
  );
  const rotation = evidence.photoRotation;

  return (
    <article
      className={`overflow-hidden rounded-2xl border-[3px] bg-white shadow-[0_18px_50px_rgba(15,39,31,0.24)] transition-opacity duration-500 motion-reduce:transition-none ${
        active ? "border-white opacity-100" : "border-white/90 opacity-90"
      }`}
      aria-label={`${evidence.label}. ${evidence.sourceLabel}`}
      data-source-url={evidence.sourceUrl || undefined}
    >
      <div className="flex items-center justify-between gap-2 bg-white px-2.5 py-1.5">
        <span className="truncate text-[9px] font-bold uppercase tracking-[0.1em] text-primary">
          {active ? "Scanning now" : evidence.kind}
        </span>
      </div>

      {isWebsite ? (
        <BrowserBar
          viewport={
            evidence.kind === "mobile"
              ? "mobile"
              : evidence.kind === "desktop"
                ? "desktop"
                : "neutral"
          }
        />
      ) : null}

      {evidence.kind === "competitor" ? (
        <div className={`${active ? "h-44 sm:h-52" : "h-24 sm:h-28"} relative overflow-hidden bg-[#f4f0ea]`}>
          {evidence.competitorMapSrc ? (
            <iframe
              title={evidence.alt}
              src={evidence.competitorMapSrc}
              className="absolute inset-0 h-full w-full border-0"
              loading="lazy"
              referrerPolicy="no-referrer-when-downgrade"
            />
          ) : (
            <div className="absolute inset-0 flex items-center justify-center px-3 text-center">
              <p className={`${active ? "text-[11px]" : "text-[9px]"} font-medium leading-snug text-muted`}>
                Nearby same-cuisine restaurants are being located on Google Maps.
              </p>
            </div>
          )}
          <div className="pointer-events-none absolute inset-x-0 top-0 bg-gradient-to-b from-white/95 via-white/70 to-transparent px-2.5 pb-4 pt-1.5">
            <p className={`${active ? "text-[9px]" : "text-[7px]"} font-bold uppercase tracking-[0.12em] text-primary`}>
              10 km discovery radius
            </p>
          </div>
          {active ? (
            <p className="pointer-events-none absolute inset-x-0 bottom-0 bg-gradient-to-t from-white/95 to-transparent px-2.5 pb-1.5 pt-4 text-[9px] font-medium leading-snug text-muted">
              Tuvi&rsquo;s ranking and scores stay hidden until the report is unlocked.
            </p>
          ) : null}
        </div>
      ) : rotation && rotation.length > 0 ? (
        <PhotoFlipStage photos={rotation} active={active} onImageError={onImageError} />
      ) : (
        <DeviceFrame
          kind={evidence.kind === "listing" ? "website" : evidence.kind}
          active={active}
        >
          {/* Every source here decoded before the card was allowed on stage. */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={evidence.src}
            alt={evidence.alt}
            className={`h-full w-full object-top ${
              evidence.kind === "listing" ? "object-cover" : "object-contain"
            }`}
            onError={() => onImageError(evidence.src!)}
          />
        </DeviceFrame>
      )}

      <div className={`${active ? "px-3 py-2.5" : "px-2 py-1.5"} flex items-end justify-between gap-2 bg-white`}>
        <div className="min-w-0">
          <p className={`${active ? "text-[12px]" : "text-[9px]"} truncate font-semibold text-ink`}>{evidence.label}</p>
          <p className="mt-0.5 truncate text-[12px] font-medium text-muted">{evidence.sourceLabel}</p>
          {evidence.authorAttributions?.length ? (
            <div className="mt-1 flex flex-wrap items-center gap-x-1.5 gap-y-1 text-[12px] leading-4 text-muted">
              <span>Photo by</span>
              {evidence.authorAttributions.map((attribution, index) => {
                const authorUri = safeHttpUrl(attribution.uri);
                const authorPhotoUri = safeHttpUrl(attribution.photoUri);
                const name = attribution.displayName || "Google contributor";
                return (
                  <span key={`${name}-${index}`} className="inline-flex items-center gap-1">
                    {index > 0 ? <span aria-hidden="true">·</span> : null}
                    {authorPhotoUri ? (
                      // Google photo contributor avatar; never proxied or optimized.
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={authorPhotoUri}
                        alt=""
                        className="h-4 w-4 rounded-full object-cover"
                        referrerPolicy="no-referrer"
                      />
                    ) : null}
                    {authorUri ? (
                      <a
                        href={authorUri}
                        target="_blank"
                        rel="noreferrer"
                        className="pointer-events-auto text-primary underline decoration-primary/30 underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary"
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
          ) : null}
        </div>
        {sourceLinks.length > 0 || flagContentUri ? (
          <div className="flex shrink-0 flex-col items-end gap-1">
            {sourceLinks.map((href, index) => (
              <a
                key={href}
                href={href}
                target="_blank"
                rel="noreferrer"
                className="pointer-events-auto rounded-full border border-ink/15 px-2 py-1 text-[12px] font-semibold text-primary underline decoration-primary/30 underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
              >
                {index === 0 && googleMapsUri ? "Google Maps" : "View source"}
              </a>
            ))}
            {flagContentUri ? (
              <a
                href={flagContentUri}
                target="_blank"
                rel="noreferrer"
                className="pointer-events-auto text-[12px] font-semibold text-muted underline underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary"
              >
                Report content
              </a>
            ) : null}
          </div>
        ) : null}
      </div>
    </article>
  );
}

const SATELLITE_WIDTH_CLASS = "w-[100px] sm:w-[140px]";

/**
 * Whatever a step needs to show — one lone competitor card, six listing
 * photos — gets exactly that: a centred hero plus its satellites, positioned
 * by `satellitePositions` rather than picked from a fixed table, so the
 * count on stage always determines the spacing instead of squeezing into
 * slots sized for a different composition.
 */
function EvidenceCollage({
  evidence,
  onImageError,
}: {
  evidence: ScanEvidence[];
  onImageError: (src: string) => void;
}) {
  const [hero, ...satellites] = evidence;
  const positions = useMemo(() => satellitePositions(satellites.length), [satellites.length]);
  if (!hero) return null;

  return (
    <div className="pointer-events-none absolute inset-x-3 bottom-40 top-20 z-30 sm:inset-x-5 sm:bottom-28 sm:top-24 lg:bottom-12">
      <p className="sr-only" role="status" aria-live="polite">
        {`Reviewing ${hero.label}`}
      </p>

      <div
        className="absolute left-1/2 top-1/2 w-[min(250px,68vw)] -translate-x-1/2 -translate-y-1/2 sm:w-[320px]"
        style={{ zIndex: 50 }}
      >
        <div
          className="scan-evidence-card-enter"
          style={{ animationDelay: "0ms", animationDuration: `${EVIDENCE_CARD_ENTRY_MS}ms` }}
        >
          <div style={{ transform: "rotate(-1.5deg)" }}>
            <EvidenceCard evidence={hero} active onImageError={onImageError} />
          </div>
        </div>
      </div>

      {satellites.map((item, i) => {
        const pos = positions[i];
        const entryDelayMs = evidenceCardEntryDelayMs(i + 1);
        return (
          <div
            key={item.id}
            className={`absolute -translate-x-1/2 -translate-y-1/2 transition-[left,top] duration-500 motion-reduce:transition-none ${SATELLITE_WIDTH_CLASS}`}
            style={{ left: `${pos.leftPercent}%`, top: `${pos.topPercent}%`, zIndex: 40 - i }}
          >
            <div
              className="scan-evidence-card-enter"
              style={{
                animationDelay: `${entryDelayMs}ms`,
                animationDuration: `${EVIDENCE_CARD_ENTRY_MS}ms`,
              }}
            >
              <div style={{ transform: `rotate(${pos.rotateDeg}deg)` }}>
                <EvidenceCard evidence={item} active={false} onImageError={onImageError} />
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function ReviewCard({ review, mapsUri }: { review: ScanReview; mapsUri?: string }) {
  const sentiment = review.sentiment ? sentimentMeta(review.sentiment) : null;
  const authorUri = safeHttpUrl(review.authorUri);
  const authorPhotoUri = safeHttpUrl(review.authorPhotoUri);
  const visitLabel = reviewVisitLabel(review.visitDate);
  const sourceUri = safeHttpUrl(review.googleMapsUri) || safeHttpUrl(mapsUri);
  const flagContentUri = safeHttpUrl(review.flagContentUri);

  return (
    <article className="scan-review-slide min-h-[88px] rounded-[18px] bg-[#dce6dd] px-3.5 py-3 sm:min-h-[100px] sm:px-4 sm:py-3.5">
      <div className="flex items-start justify-between gap-2">
        <div className="flex min-w-0 items-start gap-2">
          {authorPhotoUri ? (
            // Google reviewer avatar stays live from the supplied attribution URI.
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={authorPhotoUri}
              alt=""
              className="h-7 w-7 shrink-0 rounded-full object-cover"
              referrerPolicy="no-referrer"
            />
          ) : null}
          <div className="min-w-0">
            {authorUri ? (
              <a
                href={authorUri}
                target="_blank"
                rel="noreferrer"
                className="block truncate text-[13px] font-semibold text-ink underline decoration-ink/25 underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
              >
                {review.author || "Google reviewer"}
              </a>
            ) : (
              <p className="truncate text-[13px] font-semibold text-ink">
                {review.author || "Google reviewer"}
              </p>
            )}
            <div className="mt-0.5 flex flex-wrap items-center gap-x-1.5">
              {typeof review.rating === "number" ? <Stars rating={review.rating} /> : null}
              {review.relativeTime ? (
                <span className="truncate text-[11px] text-ink/60">{review.relativeTime}</span>
              ) : null}
              {visitLabel ? (
                <span className="truncate text-[11px] text-ink/60">· {visitLabel}</span>
              ) : null}
            </div>
          </div>
        </div>
        {sentiment ? (
          <span
            className={`shrink-0 rounded-full px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wide ${sentiment.className}`}
          >
            {sentiment.label}
          </span>
        ) : null}
      </div>
      <p className="mt-2 line-clamp-2 text-[12px] font-medium leading-snug text-ink/80 sm:line-clamp-3">
        {review.text ? `“${review.text}”` : "This reviewer left a rating without written feedback."}
      </p>
      {sourceUri || flagContentUri ? (
        <div className="mt-1.5 flex items-center gap-2 text-[11px] font-semibold">
          {sourceUri ? (
            <a
              href={sourceUri}
              target="_blank"
              rel="noreferrer"
              className="text-primary underline decoration-primary/30 underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
            >
              Google Maps
            </a>
          ) : null}
          {flagContentUri ? (
            <a
              href={flagContentUri}
              target="_blank"
              rel="noreferrer"
              className="font-normal text-ink/50 underline underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary"
            >
              Report content
            </a>
          ) : null}
        </div>
      ) : null}
    </article>
  );
}

/**
 * Recent Google reviews in the same rounded tile the marketing pages use, so
 * the scan reads as one product. Two reviews share the wall at once, holding
 * for the same beat as a photo card before sliding to the next pair.
 */
function ReviewWall({
  reviews,
  placeRating,
  mapsUri,
}: {
  reviews: ScanReview[];
  placeRating?: number;
  mapsUri?: string;
}) {
  const shown = useMemo(() => reviews.slice(0, REVIEW_WALL_LIMIT), [reviews]);
  const [pageStart, setPageStart] = useState(0);
  const [paused, setPaused] = useState(false);

  useEffect(() => {
    if (paused || shown.length <= REVIEW_PAGE_SIZE) return;
    const id = window.setInterval(
      () => setPageStart((current) => nextEvidenceBatchStart(current, shown.length, REVIEW_PAGE_SIZE)),
      REVIEW_WALL_HOLD_MS,
    );
    return () => window.clearInterval(id);
  }, [paused, shown.length]);

  if (shown.length === 0) return null;

  const safePageStart = pageStart < shown.length ? pageStart : 0;
  const page = shown.slice(safePageStart, safePageStart + REVIEW_PAGE_SIZE);
  const pageCount = Math.ceil(shown.length / REVIEW_PAGE_SIZE);
  const pageNumber = Math.floor(safePageStart / REVIEW_PAGE_SIZE) + 1;

  // Anchored top-right, below the centred search chip at every breakpoint —
  // on a narrow phone the chip spans nearly the full width, so "top right"
  // means stacked underneath it rather than beside it, and that same offset
  // still reads as a clean corner box once the canvas is wide enough for the
  // chip to leave room beside it. A capped height keeps it off short screens.
  return (
    <section
      className="pointer-events-auto absolute right-3 top-24 z-40 max-h-[52vh] w-[min(320px,calc(100%-1.5rem))] overflow-hidden rounded-[22px] border border-black/5 bg-white/95 p-3 shadow-[0_18px_50px_rgba(15,39,31,0.2)] backdrop-blur sm:right-4 sm:top-28 sm:w-[360px] sm:rounded-[24px] sm:p-3.5 lg:right-6 lg:w-[400px] lg:p-4"
      aria-label="Recent Google reviews"
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
      onFocusCapture={() => setPaused(true)}
      onBlurCapture={() => setPaused(false)}
    >
      <div className="flex items-center justify-between gap-2">
        <p className="text-[10px] font-bold uppercase tracking-[0.12em] text-primary">
          Recent Google reviews
        </p>
        <div className="flex shrink-0 items-center gap-1.5">
          {typeof placeRating === "number" ? (
            <span className="rounded-full bg-[#f4f0ea] px-2 py-0.5 text-[11px] font-semibold tabular-nums text-ink">
              {placeRating.toFixed(1)}★
            </span>
          ) : null}
          <span className="text-[11px] font-semibold tabular-nums text-muted">
            {pageNumber} / {pageCount}
          </span>
        </div>
      </div>

      <div key={`review-page-${safePageStart}`} className="mt-2.5 flex flex-col gap-2 sm:mt-3">
        {page.map((review, i) => (
          <ReviewCard
            key={`${safePageStart}-${i}-${review.author || "anon"}`}
            review={review}
            mapsUri={mapsUri}
          />
        ))}
      </div>

      {pageCount > 1 ? (
        <div className="mt-2.5 flex items-center gap-1" aria-hidden="true">
          {Array.from({ length: pageCount }, (_, dot) => (
            <span
              key={`review-dot-${dot}`}
              className={`h-1.5 rounded-full transition-all duration-300 motion-reduce:transition-none ${
                dot === pageNumber - 1 ? "w-4 bg-primary" : "w-1.5 bg-ink/15"
              }`}
            />
          ))}
        </div>
      ) : null}
    </section>
  );
}

/**
 * A live Google Maps nearby search for the same cuisine. This is Google's own
 * public result set, not Tuvi's scored competitor rows, which stay redacted
 * server-side until the report is unlocked.
 */
function nearbyCompetitorMapSrc(opts: {
  cuisine: string;
  address?: string;
  latitude?: number;
  longitude?: number;
}): string | null {
  const term = `${opts.cuisine} restaurants`.trim();
  const lat = opts.latitude;
  const lng = opts.longitude;
  if (
    typeof lat === "number" &&
    typeof lng === "number" &&
    Number.isFinite(lat) &&
    Number.isFinite(lng) &&
    Math.abs(lat) <= 90 &&
    Math.abs(lng) <= 180
  ) {
    const params = new URLSearchParams({
      q: term,
      ll: `${lat},${lng}`,
      z: "13",
      hl: "en",
      output: "embed",
    });
    return `https://www.google.com/maps?${params.toString()}`;
  }
  const near = (opts.address || "").trim();
  if (!near) return null;
  const params = new URLSearchParams({
    q: `${term} near ${near}`,
    z: "13",
    hl: "en",
    output: "embed",
  });
  return `https://www.google.com/maps?${params.toString()}`;
}

function mapEmbedSrc(opts: {
  restaurantName: string;
  address?: string;
  placeId?: string;
  latitude?: number;
  longitude?: number;
}): string | null {
  const lat = opts.latitude;
  const lng = opts.longitude;
  if (
    typeof lat === "number" &&
    typeof lng === "number" &&
    Number.isFinite(lat) &&
    Number.isFinite(lng) &&
    Math.abs(lat) <= 90 &&
    Math.abs(lng) <= 180
  ) {
    // Exact WGS84 pin — Google Maps native marker at the listing location
    const params = new URLSearchParams({
      q: `${lat},${lng}`,
      ll: `${lat},${lng}`,
      z: "17",
      hl: "en",
      output: "embed",
    });
    return `https://www.google.com/maps?${params.toString()}`;
  }
  if (opts.placeId) {
    // Pin via Places ID from the first frame — no generic world map
    const label = [opts.restaurantName, opts.address].filter(Boolean).join(", ");
    const params = new URLSearchParams({
      q: label ? `${label}` : `place_id:${opts.placeId}`,
      query_place_id: opts.placeId,
      z: "17",
      hl: "en",
      output: "embed",
    });
    return `https://www.google.com/maps?${params.toString()}`;
  }
  const q = [opts.restaurantName, opts.address].filter(Boolean).join(" ");
  if (!q || q === "Your restaurant") return null;
  const params = new URLSearchParams({
    q,
    z: "16",
    hl: "en",
    output: "embed",
  });
  return `https://www.google.com/maps?${params.toString()}`;
}

/**
 * What the map shows follows the checklist step exactly, positionally
 * matched to the `steps` array built below: the competitor card only while
 * "& competitors" is active, listing photos only during "Photo quality and
 * quantity", and so on. "Google business profile" has no evidence card of
 * its own — the map and pin already speak for that step — so the collage
 * simply goes empty for its beat. The review step is deliberately empty
 * here too: reviews live in the separate `ReviewWall`, gated the same way.
 */
const STEP_EVIDENCE_KINDS: readonly (readonly ScanEvidence["kind"][])[] = [
  ["competitor"],
  [],
  ["listing"],
  [],
  ["desktop", "website"],
  ["mobile"],
];
const REVIEW_STEP_INDEX = 3;
// Stable reference so an out-of-range step (e.g. once `finishing` pushes
// activeIndex past the last real step) doesn't recreate an empty array — and
// with it invalidate the memo below — on every render.
const NO_EVIDENCE_KINDS: readonly ScanEvidence["kind"][] = [];

export default function ScanExperience({
  restaurantName = "Your restaurant",
  address,
  rating,
  category = "Restaurant",
  website,
  placeId,
  mapsUri,
  placeAttributions = [],
  latitude,
  longitude,
  photoUrl,
  photos = [],
  reviews = [],
  desktopScreenshot,
  mobileScreenshot,
  websiteUrl,
  fetchComplete = false,
  onReady,
  className = "",
}: ScanExperienceProps) {
  const websiteLabel = useMemo(() => {
    if (!website) return "Website / listing signals";
    try {
      return new URL(website).hostname.replace(/^www\./, "");
    } catch {
      return website.replace(/^https?:\/\//, "").split("/")[0] || "Website signals";
    }
  }, [website]);

  const steps = useMemo(
    () => [
      `${restaurantName} & competitors`,
      "Google business profile",
      "Photo quality and quantity",
      "Google review sentiment",
      websiteLabel.startsWith("http") ? websiteLabel : website || websiteLabel,
      "Mobile experience",
    ],
    [restaurantName, website, websiteLabel],
  );

  const [activeIndex, setActiveIndex] = useState(0);
  const [secondsLeft, setSecondsLeft] = useState(TARGET_SCAN_SECONDS);
  const [finishing, setFinishing] = useState(false);
  const [imageLoadStates, setImageLoadStates] = useState<Map<string, ImageLoadStatus>>(
    () => new Map(),
  );
  const startedAtRef = useRef<number>(0);
  const evidenceReadyAtElapsedRef = useRef<number | null>(null);
  const requestedSrcRef = useRef<Set<string>>(new Set());
  const onReadyRef = useRef(onReady);

  useEffect(() => {
    onReadyRef.current = onReady;
  }, [onReady]);

  const markEvidenceReady = useCallback(() => {
    if (evidenceReadyAtElapsedRef.current !== null) return;
    evidenceReadyAtElapsedRef.current =
      startedAtRef.current > 0 ? Math.max(0, Date.now() - startedAtRef.current) : 0;
  }, []);

  const setImageLoadStatus = useCallback(
    (src: string, status: ImageLoadStatus) => {
      if (status === "loaded") markEvidenceReady();
      setImageLoadStates((current) => {
        if (current.get(src) === status) return current;
        const next = new Map(current);
        next.set(src, status);
        return next;
      });
    },
    [markEvidenceReady],
  );

  const gallery = useMemo(() => {
    const seen = new Set<string>();
    const out: ScanPhoto[] = [];
    const push = (photo: ScanPhoto) => {
      const s = (photo.src || "").trim();
      if (!s || seen.has(s)) return;
      seen.add(s);
      out.push({ ...photo, src: s });
    };
    for (const photo of photos) push(photo);
    if (photoUrl) {
      push({
        src: photoUrl,
        label: `${restaurantName} listing photo`,
        sourceLabel: "Google listing photo",
        alt: `${restaurantName} listing photo`,
      });
    }
    // No stock-photo fallback. Keep enough real media for multiple paced rotations.
    return out.slice(0, 12);
  }, [photoUrl, photos, restaurantName]);

  // Cards must never appear before their pixels exist. Every candidate source is
  // decoded off-stage first, so the collage only ever shows real, loaded media.
  //
  // Deliberately no per-run cancellation here. A URL's decode outcome is a
  // permanent fact, not something that goes stale when this effect re-runs for
  // an unrelated prop change — and `requestedSrcRef` already stops it firing
  // twice for the same source. An earlier version tied a `cancelled` flag to
  // each run's cleanup, which discarded every in-flight result the instant the
  // effect re-ran (guaranteed at least once under Strict Mode in development,
  // and again in production if the caller ever passes a new `photos` array
  // identity before the first pass finishes) — so no photo ever reached
  // "loaded" and the collage stayed empty for its entire scan.
  useEffect(() => {
    const candidates = [
      ...gallery.map((photo) => photo.src),
      desktopScreenshot,
      mobileScreenshot,
    ].filter((src): src is string => Boolean(src));
    const pending = Array.from(new Set(candidates)).filter(
      (src) => !requestedSrcRef.current.has(src),
    );

    for (const src of pending) {
      requestedSrcRef.current.add(src);
      // No referrerPolicy override here: listing photos and website captures
      // used the browser default before this preload existed, and Google's
      // photo CDN can reject a stripped referrer. "no-referrer" is reserved
      // for reviewer/contributor avatars, which are rendered directly and
      // never routed through this preload.
      const loader = new window.Image();
      loader.decoding = "async";
      loader.onload = () => setImageLoadStatus(src, loader.naturalWidth > 0 ? "loaded" : "failed");
      loader.onerror = () => setImageLoadStatus(src, "failed");
      loader.src = src;
    }
  }, [gallery, desktopScreenshot, mobileScreenshot, setImageLoadStatus]);

  const isDecoded = useCallback(
    (src?: string) => Boolean(src) && imageLoadStates.get(src!) === "loaded",
    [imageLoadStates],
  );

  const readyPhotos = useMemo(
    () => gallery.filter((photo) => isDecoded(photo.src)),
    [gallery, isDecoded],
  );

  const reviewWallItems = useMemo(
    () =>
      reviews
        .filter(
          (review) =>
            Boolean(review.author?.trim()) ||
            Boolean(review.text?.trim()) ||
            typeof review.rating === "number",
        )
        .slice(0, REVIEW_WALL_LIMIT),
    [reviews],
  );

  const hasReviewEvidence = reviewWallItems.length > 0;

  const evidenceItems = useMemo<ScanEvidence[]>(() => {
    const host = websiteLabel === "Website / listing signals" ? "Restaurant website" : websiteLabel;
    const items: ScanEvidence[] = [];

    // Decoded photos spread across up to five cards. Each card turns through
    // its own photos, so extra media deepens the cards instead of hiding.
    for (const [cardIndex, cardPhotos] of dealPhotosToCards(
      readyPhotos,
      LISTING_CARD_COUNT,
    ).entries()) {
      const lead = cardPhotos[0];
      items.push({
        id: `listing-${cardIndex}`,
        kind: "listing",
        label: lead.label || `Listing photo ${cardIndex + 1}`,
        sourceLabel: lead.sourceLabel || "Google listing photo",
        sourceUrl: lead.sourceUrl,
        authorAttributions: lead.authorAttributions,
        googleMapsUri: lead.googleMapsUri,
        flagContentUri: lead.flagContentUri,
        alt: lead.alt || `${restaurantName} listing photo ${cardIndex + 1}`,
        src: lead.src,
        photoRotation: cardPhotos,
      });
    }

    const desktopReady = isDecoded(desktopScreenshot);
    const mobileReady = isDecoded(mobileScreenshot);
    const duplicateViewportCapture = Boolean(
      desktopReady && mobileReady && desktopScreenshot === mobileScreenshot,
    );

    if (duplicateViewportCapture) {
      items.push({
        id: "website-neutral",
        kind: "website",
        label: "Website viewport capture",
        sourceLabel: `Website capture · ${host}`,
        alt: `${restaurantName} website screenshot; viewport could not be distinguished`,
        src: desktopScreenshot,
        sourceUrl: websiteUrl || website,
      });
    } else {
      if (desktopReady) {
        items.push({
          id: "website-desktop",
          kind: "desktop",
          label: "Desktop website view",
          sourceLabel: `Website capture · ${host}`,
          alt: `${restaurantName} website desktop screenshot`,
          src: desktopScreenshot,
          sourceUrl: websiteUrl || website,
        });
      }
      if (mobileReady) {
        items.push({
          id: "website-mobile",
          kind: "mobile",
          label: "Mobile website view",
          sourceLabel: `Website capture · ${host}`,
          alt: `${restaurantName} website mobile screenshot`,
          src: mobileScreenshot,
          sourceUrl: websiteUrl || website,
        });
      }
    }

    const cuisineLabel = category.trim() || "Restaurant";
    items.push({
      id: "competitor-discovery",
      kind: "competitor",
      label: "Nearby same-cuisine restaurants",
      sourceLabel: `Google Maps discovery · ${cuisineLabel} · within 10 km`,
      sourceUrl: mapsUri,
      alt: `Nearby ${cuisineLabel.toLowerCase()} restaurants on Google Maps`,
      competitorMapSrc:
        nearbyCompetitorMapSrc({
          cuisine: cuisineLabel,
          address,
          latitude,
          longitude,
        }) ?? undefined,
    });

    return items;
  }, [
    address,
    category,
    desktopScreenshot,
    isDecoded,
    latitude,
    longitude,
    mapsUri,
    mobileScreenshot,
    readyPhotos,
    restaurantName,
    website,
    websiteLabel,
    websiteUrl,
  ]);

  const hasLoadedVisualEvidence = evidenceItems.some(
    (item) => Boolean(item.src) && imageLoadStates.get(item.src!) === "loaded",
  );
  const hasReadyEvidence = hasReviewEvidence || hasLoadedVisualEvidence;

  useEffect(() => {
    if (hasReviewEvidence) markEvidenceReady();
  }, [hasReviewEvidence, markEvidenceReady]);

  // Only the cards relevant to the checklist's current step are on stage —
  // the competitor card while "& competitors" is active, listing photos only
  // during "Photo quality and quantity", and so on. Moving to the next step
  // hides the previous step's cards outright, freeing the stage for the next
  // ones rather than leaving a growing pile on screen for the whole scan.
  const visibleKinds = STEP_EVIDENCE_KINDS[activeIndex] ?? NO_EVIDENCE_KINDS;
  const visibleEvidence = useMemo(
    () => evidenceItems.filter((item) => visibleKinds.includes(item.kind)),
    [evidenceItems, visibleKinds],
  );
  const visibleSignature = useMemo(
    () =>
      visibleEvidence
        .map((item) => `${item.id}:${item.photoRotation?.length ?? 0}`)
        .join("|"),
    [visibleEvidence],
  );
  const showReviewWall = activeIndex === REVIEW_STEP_INDEX && !finishing;

  const searchQuery = useMemo(() => {
    const cityish = (address || "").split(",").slice(-2).join(",").trim();
    if (cityish) return `${restaurantName} in ${cityish}`;
    return restaurantName;
  }, [restaurantName, address]);

  const hasExactPin =
    typeof latitude === "number" &&
    typeof longitude === "number" &&
    Number.isFinite(latitude) &&
    Number.isFinite(longitude);

  const candidateEmbedSrc = useMemo(
    () => mapEmbedSrc({ restaurantName, address, placeId, latitude, longitude }),
    [restaurantName, address, placeId, latitude, longitude],
  );
  // Keep the first useful map stable, while still promoting exact coordinates once available.
  const [initialEmbedSrc] = useState(candidateEmbedSrc);
  const embedSrc =
    hasExactPin && candidateEmbedSrc
      ? candidateEmbedSrc
      : initialEmbedSrc ?? candidateEmbedSrc;

  const activeEvidence = visibleEvidence[0];

  // How many cards each step WOULD show, independent of which step is
  // currently active — used only to size that step's own dwell time.
  const stepItemCounts = useMemo(
    () =>
      STEP_EVIDENCE_KINDS.map(
        (kinds) => evidenceItems.filter((item) => kinds.includes(item.kind)).length,
      ),
    [evidenceItems],
  );
  const reviewPageCount = Math.max(1, Math.ceil(reviewWallItems.length / REVIEW_PAGE_SIZE));

  // Countdown
  useEffect(() => {
    startedAtRef.current = Date.now();
    const id = window.setInterval(() => {
      const elapsed = Math.floor((Date.now() - startedAtRef.current) / 1000);
      setSecondsLeft(Math.max(0, TARGET_SCAN_SECONDS - elapsed));
    }, 250);
    return () => window.clearInterval(id);
  }, []);

  // Each step dwells for as long as its own evidence needs, not a fixed
  // interval: the photo step needs its six cards to finish entering (a
  // second apart) and hold before it's fair to hide them, and the review
  // step needs a few page-slides to actually show the "next two" behaviour,
  // while an empty step (no card of its own, e.g. business profile) still
  // gets a minimum readable beat rather than a zero-length flash.
  useEffect(() => {
    if (finishing || activeIndex >= steps.length - 1) return;
    const durationMs =
      activeIndex === REVIEW_STEP_INDEX
        ? reviewWallItems.length === 0
          ? MIN_STEP_MS
          : Math.max(Math.min(reviewPageCount, 3) * REVIEW_WALL_HOLD_MS, MIN_STEP_MS)
        : Math.max(evidenceBatchPresentationMs(stepItemCounts[activeIndex] ?? 0), MIN_STEP_MS);
    const id = window.setTimeout(() => {
      setActiveIndex((i) => Math.min(i + 1, steps.length - 1));
    }, durationMs);
    return () => window.clearTimeout(id);
  }, [activeIndex, finishing, reviewPageCount, reviewWallItems.length, stepItemCounts, steps.length]);

  // Cached responses still receive the full 15-second review. Evidence arriving
  // late gets one complete staggered batch, including the final-card reading hold.
  useEffect(() => {
    let cancelled = false;
    let completing = false;
    let finishTimer: number | undefined;

    const tryFinish = () => {
      if (cancelled || completing) return;
      const elapsed = Date.now() - startedAtRef.current;
      if (
        !isScanCompletionReady({
          elapsedMs: elapsed,
          fetchComplete,
          evidenceReadyAtElapsedMs: evidenceReadyAtElapsedRef.current,
        })
      ) {
        return;
      }

      completing = true;
      setFinishing(true);
      setActiveIndex(steps.length);

      finishTimer = window.setTimeout(() => {
        if (!cancelled) onReadyRef.current?.();
      }, FINISH_HOLD_MS);
    };

    const pollTimer = window.setInterval(tryFinish, 400);
    tryFinish();

    return () => {
      cancelled = true;
      if (finishTimer) window.clearTimeout(finishTimer);
      if (pollTimer) window.clearInterval(pollTimer);
    };
  }, [fetchComplete, hasReadyEvidence, steps.length]);

  const progress = finishing
    ? 1
    : Math.min(
        0.96,
        ((TARGET_SCAN_SECONDS - secondsLeft) / TARGET_SCAN_SECONDS) * 0.96,
      );

  const statusLine = finishing
    ? "Wrapping up your report…"
    : secondsLeft > 0
      ? `${secondsLeft} seconds remaining`
      : fetchComplete
        ? "Reviewing final evidence…"
        : "Almost done…";

  return (
    <div className={`relative min-h-[calc(100dvh-4.5rem)] overflow-hidden bg-[#f3f1ed] ${className}`}>
      <div className="relative grid min-h-[calc(100dvh-4.5rem)] w-full lg:grid-cols-[minmax(260px,320px)_minmax(0,1fr)]">
        {/* Desktop — expanded checklist. Mobile keeps identity, map, media and progress above fold. */}
        <aside className="relative z-20 order-2 hidden flex-col border-b border-black/5 bg-white/95 px-6 py-7 backdrop-blur-md sm:px-7 lg:order-1 lg:flex lg:border-b-0 lg:border-r lg:py-10">
          <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-muted">Digital footprint scan</p>
          <h1 className="mt-2 font-display text-[2rem] font-semibold leading-none tracking-[-0.04em] text-ink">
            Scanning…
          </h1>
          <p className="mt-3 text-[14px] leading-relaxed text-muted">
            Checking Google listing visibility, reviews, photos, website &amp; mobile signals — live.
          </p>
          {hasReviewEvidence ? (
            <p className="mt-3 text-[12px] leading-relaxed text-muted">
              {REVIEW_RELEVANCE_NOTICE}
            </p>
          ) : null}

          <ul className="mt-8 flex-1 space-y-3.5">
            {steps.map((label, index) => {
              const done = finishing || index < activeIndex;
              const active = !done && index === activeIndex;
              return (
                <li
                  key={`${index}-${label}`}
                  className={`flex items-start gap-3 rounded-xl px-2 py-1.5 transition-colors duration-300 ${
                    active ? "bg-[#f4f7f5]" : ""
                  }`}
                >
                  <span
                    className={`mt-0.5 flex h-[22px] w-[22px] shrink-0 items-center justify-center rounded-full transition-all duration-300 ${
                      done
                        ? "bg-ink text-bg shadow-sm"
                        : active
                          ? "scan-step-active border-2 border-primary bg-white text-primary"
                          : "border border-[#d5d0c8] bg-transparent"
                    }`}
                  >
                    {done ? <CheckIcon /> : active ? <Spinner /> : null}
                  </span>
                  <span
                    className={`pt-0.5 text-[14px] leading-snug transition-colors duration-300 ${
                      done || active ? "font-medium text-ink" : "text-muted/55"
                    }`}
                  >
                    {label}
                  </span>
                </li>
              );
            })}
          </ul>

          <div className="mt-8 border-t border-black/5 pt-5">
            <div className="h-1.5 overflow-hidden rounded-full bg-[#ebe6de]">
              <div
                className="scan-progress-fill h-full rounded-full transition-[width] duration-700 ease-out"
                style={{ width: `${Math.max(8, progress * 100)}%` }}
              />
            </div>
            <div className="mt-3 flex items-center gap-2 text-[13px] text-muted">
              <Spinner />
              <span className="font-medium tabular-nums">{statusLine}</span>
            </div>
            {fetchComplete && !finishing ? (
              <p className="mt-2 text-[12px] text-accent">Signals found — polishing your scorecard…</p>
            ) : null}
          </div>
        </aside>

        {/* Right — Google Map stage + floating scraped photos */}
        <section className="relative order-1 min-h-[calc(100dvh-4.5rem)] overflow-hidden bg-[#e8e4dc] lg:order-2 lg:min-h-0">
          {embedSrc ? (
            <iframe
              title={`Map of ${restaurantName}`}
              src={embedSrc}
              className="absolute inset-0 h-full w-full border-0"
              loading="eager"
              referrerPolicy="no-referrer-when-downgrade"
              allowFullScreen
            />
          ) : (
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="flex items-center gap-2 rounded-full bg-white/90 px-4 py-2 text-[13px] font-medium text-muted shadow-sm">
                <Spinner />
                Locating on Google Maps…
              </div>
            </div>
          )}

          {/* Soft veil so photos read clearly — keep light so Google's pin stays visible */}
          <div
            className="pointer-events-none absolute inset-0 bg-gradient-to-b from-white/15 via-transparent to-white/20"
            aria-hidden="true"
          />

          {/* Owner-style search chip */}
          <div className="absolute left-1/2 top-5 z-20 w-[min(420px,calc(100%-2rem))] -translate-x-1/2 sm:top-7">
            <div className="flex items-center gap-3 rounded-full border border-black/5 bg-white/95 px-4 py-3 shadow-[0_12px_40px_rgba(15,39,31,0.14)] backdrop-blur">
              <svg viewBox="0 0 20 20" className="h-4 w-4 shrink-0 text-muted" fill="none" aria-hidden="true">
                <circle cx="9" cy="9" r="5.5" stroke="currentColor" strokeWidth="1.8" />
                <path d="M13.5 13.5 17 17" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
              </svg>
              <p className="truncate text-[14px] font-medium text-ink">{searchQuery}</p>
            </div>
          </div>

          {/* Only the current checklist step's evidence is on stage. */}
          <EvidenceCollage
            key={`evidence-collage-${activeIndex}-${visibleSignature}`}
            evidence={visibleEvidence}
            onImageError={(src) => setImageLoadStatus(src, "failed")}
          />

          {/* Reviews take the top-right corner only during their own step. */}
          {showReviewWall ? (
            <ReviewWall reviews={reviewWallItems} placeRating={rating} mapsUri={mapsUri} />
          ) : null}

          {/* Scan beam across map */}
          <div className="pointer-events-none absolute inset-x-0 top-0 z-10 h-full overflow-hidden" aria-hidden="true">
            <div className="scan-beam absolute left-0 right-0 h-[2px] opacity-70" />
          </div>

          <div className="absolute bottom-4 left-4 z-20 hidden rounded-full bg-white/90 px-3 py-1.5 text-[12px] font-medium text-muted shadow-sm lg:block">
            {finishing
              ? "Evidence collected — building report…"
              : activeEvidence
                ? `Scanning ${activeEvidence.label.toLowerCase()}…`
                : hasExactPin
                  ? "Google Maps · pinned to listing"
                  : "Live Google listing scan"}
          </div>

          <div
            className="pointer-events-auto absolute bottom-4 right-4 z-50 flex max-w-[calc(100%-2rem)] flex-wrap justify-end gap-x-2 gap-y-1 rounded-xl bg-white/95 px-3 py-2 text-[12px] font-normal leading-4 text-[#5e5e5e] shadow-sm"
            translate="no"
          >
            <a
              href={safeHttpUrl(mapsUri) || "https://www.google.com/maps"}
              target="_blank"
              rel="noreferrer"
              className="underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
            >
              Google Maps
            </a>
            {placeAttributions.map((attribution, index) => {
              const providerHref = safeHttpUrl(attribution.providerUri);
              const provider = attribution.provider?.trim() || "Data source";
              return providerHref ? (
                <a
                  key={`${provider}-${index}`}
                  href={providerHref}
                  target="_blank"
                  rel="noreferrer"
                  className="underline underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
                >
                  Source: {provider}
                </a>
              ) : (
                <span key={`${provider}-${index}`}>Source: {provider}</span>
              );
            })}
          </div>

          <div className="absolute inset-x-4 bottom-24 z-50 rounded-2xl border border-black/5 bg-white/95 p-4 shadow-[0_18px_50px_rgba(15,39,31,0.2)] backdrop-blur lg:hidden">
            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <p className="text-[10px] font-semibold uppercase tracking-[0.12em] text-muted">
                  Digital footprint scan
                </p>
                <p className="mt-1 truncate text-[14px] font-semibold text-ink">
                  {finishing ? "Building your scorecard" : steps[Math.min(activeIndex, steps.length - 1)]}
                </p>
              </div>
              <span className="shrink-0 text-[12px] font-semibold tabular-nums text-primary">{statusLine}</span>
            </div>
            <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-[#ebe6de]">
              <div
                className="scan-progress-fill h-full rounded-full transition-[width] duration-500 ease-out"
                style={{ width: `${Math.max(8, progress * 100)}%` }}
              />
            </div>
            {hasReviewEvidence ? (
              <p className="mt-2 text-[12px] leading-snug text-muted">
                {REVIEW_RELEVANCE_NOTICE}
              </p>
            ) : null}
          </div>
        </section>
      </div>
    </div>
  );
}
