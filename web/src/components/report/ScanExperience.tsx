"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import type { CSSProperties } from "react";
import {
  buildScanPhotoSlots,
  buildScanReviewStream,
  normalizeScanPhotos,
  normalizeScanReviews,
  reportMapEmbedUrl,
  websiteCaptureEvidence,
} from "@/lib/report-scan";
import type {
  ScanPhotoEvidence,
  ScanReviewEvidence,
  WebsiteCaptureEvidence,
} from "@/lib/report-scan";

const MIN_SCAN_MS = 6_000;
const MIN_EVIDENCE_PREVIEW_MS = 6_500;
const STEP_INTERVAL_MS = 2_200;
const FINISH_HOLD_MS = 320;
const TARGET_SECONDS = 15;
const PHOTO_CARD_LIMIT = 6;
const REVIEW_STREAM_CARD_COUNT = 10;

export type ScanPhoto = ScanPhotoEvidence;
export type ScanReview = ScanReviewEvidence;

export type ScanExperienceProps = {
  restaurantName?: string;
  address?: string;
  rating?: number;
  category?: string;
  website?: string;
  placeId?: string;
  mapsUri?: string;
  latitude?: number;
  longitude?: number;
  photoUrl?: string;
  photos?: ScanPhoto[];
  /** Available Google reviews shown during the review-sentiment step. */
  reviews?: ScanReview[];
  /** Desktop viewport screenshot of the restaurant website (data URL or http). */
  desktopScreenshot?: string;
  /** Mobile viewport screenshot of the restaurant website (data URL or http). */
  mobileScreenshot?: string;
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

function sentimentMeta(sentiment?: string) {
  const s = (sentiment || "mixed").toLowerCase();
  if (s === "positive") return { label: "Positive", className: "bg-[#e8f6ee] text-[#1f7a45]" };
  if (s === "negative") return { label: "Negative", className: "bg-[#fdecea] text-[#b42318]" };
  return { label: "Mixed", className: "bg-[#fff6e5] text-[#a15c00]" };
}

function PhotoBoard({
  photos,
  restaurantName,
}: {
  photos: ScanPhoto[];
  restaurantName: string;
}) {
  const [readySources, setReadySources] = useState<Set<string>>(() => new Set());
  const [failedSources, setFailedSources] = useState<Set<string>>(() => new Set());
  const [turn, setTurn] = useState(0);
  const [paused, setPaused] = useState(false);
  const [reducedMotion, setReducedMotion] = useState(false);

  useEffect(() => {
    let cancelled = false;
    const loaders = photos.map((photo) => {
      const image = new window.Image();
      image.onload = () => {
        if (cancelled) return;
        setReadySources((current) => new Set(current).add(photo.src));
      };
      image.onerror = () => {
        if (cancelled) return;
        setFailedSources((current) => new Set(current).add(photo.src));
      };
      image.src = photo.src;
      return image;
    });
    return () => {
      cancelled = true;
      for (const image of loaders) {
        image.onload = null;
        image.onerror = null;
      }
    };
  }, [photos]);

  const availablePhotos = useMemo(
    () => photos.filter(
      (photo) => readySources.has(photo.src) && !failedSources.has(photo.src),
    ),
    [failedSources, photos, readySources],
  );
  const photoSlots = useMemo(
    () => buildScanPhotoSlots(availablePhotos, PHOTO_CARD_LIMIT),
    [availablePhotos],
  );
  const hasOverflow = photoSlots.some((slot) => slot.length > 1);

  useEffect(() => {
    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    const syncPreference = () => setReducedMotion(media.matches);
    syncPreference();
    media.addEventListener("change", syncPreference);
    return () => media.removeEventListener("change", syncPreference);
  }, []);

  useEffect(() => {
    if (!hasOverflow || paused || reducedMotion) return;
    const interval = window.setInterval(() => {
      setTurn((current) => current + 1);
    }, 3_200);
    return () => window.clearInterval(interval);
  }, [hasOverflow, paused, reducedMotion]);

  if (photoSlots.length === 0) return null;

  return (
    <section
      className="scan-capture-in min-w-0 rounded-2xl border border-black/5 bg-white/90 p-2.5 shadow-[0_10px_26px_rgba(15,39,31,0.12)]"
      aria-label={`${photoSlots.length} photo cards showing ${availablePhotos.length} available restaurant listing photos`}
    >
      <div className="mb-2 flex items-center justify-between gap-2 px-0.5">
        <div>
          <p className="text-[10px] font-semibold uppercase tracking-[0.12em] text-muted">Listing photos</p>
          <p className="text-[11px] font-semibold text-ink">
            {photoSlots.length} card{photoSlots.length === 1 ? "" : "s"} on the board
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <span className="rounded-full bg-[#edf4f0] px-2 py-1 text-[9px] font-semibold tabular-nums text-primary">
            {availablePhotos.length} found
          </span>
          {hasOverflow && !reducedMotion ? (
            <button
              type="button"
              aria-pressed={paused}
              onClick={() => setPaused((current) => !current)}
              className="min-h-7 rounded-full border border-black/10 bg-white px-2 text-[9px] font-semibold text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
            >
              {paused ? "Play" : "Pause"}
            </button>
          ) : null}
          {hasOverflow ? (
            <button
              type="button"
              onClick={() => setTurn((current) => current + 1)}
              className="min-h-7 rounded-full border border-black/10 bg-white px-2 text-[9px] font-semibold text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
            >
              Next
            </button>
          ) : null}
        </div>
      </div>
      <div className="grid grid-cols-3 gap-2">
        {photoSlots.map((slotPhotos, slotIndex) => {
          const photoIndex = turn % slotPhotos.length;
          const currentPhoto = slotPhotos[photoIndex];
          const sourceIndex = availablePhotos.findIndex((photo) => photo.src === currentPhoto.src);
          const style = { "--scan-photo-delay": `${slotIndex * 55}ms` } as CSSProperties;
          return (
            <figure
              key={`slot-${slotIndex}`}
              className="min-w-0 overflow-hidden rounded-xl border border-black/5 bg-[#f1eee8] shadow-sm"
            >
              <div
                key={currentPhoto.src}
                className={slotPhotos.length > 1 ? "scan-photo-flip" : "scan-photo-in"}
                style={style}
              >
                {/* Dynamic Google media cannot use next/image's fixed remote allow-list. */}
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={currentPhoto.src}
                  alt={currentPhoto.label || `${restaurantName} listing photo ${sourceIndex + 1}`}
                  className="h-[68px] w-full object-cover sm:h-[76px] xl:h-[68px]"
                  onError={() => {
                    setFailedSources((current) => new Set(current).add(currentPhoto.src));
                  }}
                />
                <figcaption className="flex items-center justify-between gap-1 bg-white px-2 py-1.5 text-[9px] font-semibold text-ink">
                  <span className="truncate">{currentPhoto.label || "Listing photo"}</span>
                  <span className="shrink-0 tabular-nums text-muted">{sourceIndex + 1}/{availablePhotos.length}</span>
                </figcaption>
              </div>
            </figure>
          );
        })}
      </div>
    </section>
  );
}

function ReviewCard({ review, index }: { review: ScanReview; index: number }) {
  const meta = review.sentiment ? sentimentMeta(review.sentiment) : null;
  return (
    <article className="rounded-xl border border-black/5 bg-white p-3 shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="truncate text-[12px] font-semibold text-ink">
            {review.author || `Google review ${index + 1}`}
          </p>
          {typeof review.rating === "number" || review.relativeTime ? (
            <div className="mt-0.5 flex items-center gap-1.5">
              {typeof review.rating === "number" ? <Stars rating={review.rating} /> : null}
              {review.relativeTime ? (
                <span className="truncate text-[10px] text-muted">{review.relativeTime}</span>
              ) : null}
            </div>
          ) : null}
        </div>
        {meta ? (
          <span className={`shrink-0 rounded-full px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wide ${meta.className}`}>
            {meta.label}
          </span>
        ) : null}
      </div>
      {review.text ? (
        <p className="mt-2 line-clamp-3 text-[11px] leading-snug text-ink/80">&ldquo;{review.text}&rdquo;</p>
      ) : null}
    </article>
  );
}

export function ReviewScroller({
  reviews,
  placeRating,
  variant = "default",
}: {
  reviews: ScanReview[];
  placeRating?: number;
  variant?: "default" | "board";
}) {
  const [paused, setPaused] = useState(false);
  const [reducedMotion, setReducedMotion] = useState(false);
  const [motionOverride, setMotionOverride] = useState(false);
  const stream = buildScanReviewStream(reviews, REVIEW_STREAM_CARD_COUNT);
  const canScroll = reviews.length > 1;
  const duration = Math.max(28, stream.length * 3);
  const style = {
    "--scan-review-duration": `${duration}s`,
    animationPlayState: paused ? "paused" : undefined,
  } as CSSProperties;
  const streamLabel = reviews.length >= REVIEW_STREAM_CARD_COUNT
    ? `${REVIEW_STREAM_CARD_COUNT} Google reviews`
    : `Cycling ${reviews.length} available review${reviews.length === 1 ? "" : "s"} in a ${REVIEW_STREAM_CARD_COUNT}-card stream`;

  useEffect(() => {
    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    const syncPreference = () => {
      setReducedMotion(media.matches);
      if (media.matches) setMotionOverride(false);
    };
    syncPreference();
    media.addEventListener("change", syncPreference);
    return () => media.removeEventListener("change", syncPreference);
  }, []);

  const toggleMotion = () => {
    if (reducedMotion && !motionOverride) {
      setMotionOverride(true);
      setPaused(false);
      return;
    }
    setPaused((current) => !current);
  };
  const motionButtonLabel = reducedMotion && !motionOverride
    ? "Play"
    : paused
      ? "Resume"
      : "Pause";

  if (reviews.length === 0) return null;

  return (
    <section
      className={`scan-review-rail relative flex w-full flex-col overflow-hidden rounded-2xl border border-black/5 bg-white shadow-[0_10px_26px_rgba(15,39,31,0.12)] ${
        variant === "board"
          ? "min-h-[220px] flex-1"
          : "h-[174px] max-w-[560px] sm:h-[190px]"
      }`}
      aria-label={`${reviews.length} available Google reviews`}
    >
      <header className="relative z-10 flex items-center justify-between gap-3 border-b border-black/5 bg-white/95 px-3.5 py-2.5">
        <div className="min-w-0">
          <p className="text-[10px] font-semibold uppercase tracking-[0.12em] text-muted">Available Google reviews</p>
          <p className="truncate text-[11px] font-semibold text-ink">{streamLabel}</p>
        </div>
        <div className="flex shrink-0 items-center gap-1.5">
          {typeof placeRating === "number" ? (
            <span className="rounded-full bg-[#f4f0ea] px-2 py-1 text-[11px] font-semibold tabular-nums text-ink">
              {placeRating.toFixed(1)}★
            </span>
          ) : null}
          {canScroll ? (
            <button
              type="button"
              aria-pressed={paused}
              onClick={toggleMotion}
              className="min-h-8 rounded-full border border-black/10 bg-white px-2 text-[10px] font-semibold text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
            >
              {motionButtonLabel}
            </button>
          ) : null}
        </div>
      </header>
      <div
        className={`scan-review-viewport min-h-0 overflow-hidden px-2.5 focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary ${
          variant === "board" ? "flex-1" : "h-[126px] sm:h-[142px]"
        } ${reducedMotion && motionOverride ? "scan-review-force-viewport" : ""}`}
        tabIndex={0}
        aria-label="Scrollable Google reviews"
      >
        <div
          className={canScroll
            ? `scan-review-track ${reducedMotion && motionOverride ? "scan-review-force-motion" : ""}`
            : "py-2.5"}
          style={canScroll ? style : undefined}
        >
          <div className="space-y-2 py-2.5">
            {(canScroll ? stream : stream.slice(0, 1)).map((item, index) => (
              <div key={`review-${index}-${item.sourceIndex}`} aria-hidden={item.repeated || undefined}>
                <ReviewCard review={item.review} index={item.sourceIndex} />
              </div>
            ))}
          </div>
          {canScroll ? (
            <div className="scan-review-copy space-y-2 py-2.5" aria-hidden="true">
              {stream.map((item, index) => (
                <ReviewCard key={`review-copy-${index}`} review={item.review} index={item.sourceIndex} />
              ))}
            </div>
          ) : null}
        </div>
      </div>
    </section>
  );
}

function WebsiteCaptureBoard({
  captures,
  hostname,
}: {
  captures: WebsiteCaptureEvidence[];
  hostname?: string;
}) {
  const [readySources, setReadySources] = useState<Set<string>>(() => new Set());
  const [failedSources, setFailedSources] = useState<Set<string>>(() => new Set());

  useEffect(() => {
    let cancelled = false;
    const loaders = captures.map((capture) => {
      const image = new window.Image();
      image.onload = () => {
        if (!cancelled) setReadySources((current) => new Set(current).add(capture.src));
      };
      image.onerror = () => {
        if (!cancelled) setFailedSources((current) => new Set(current).add(capture.src));
      };
      image.src = capture.src;
      return image;
    });
    return () => {
      cancelled = true;
      for (const image of loaders) {
        image.onload = null;
        image.onerror = null;
      }
    };
  }, [captures]);

  const available = captures.filter(
    (capture) => readySources.has(capture.src) && !failedSources.has(capture.src),
  );
  if (available.length === 0) return null;

  return (
    <section className="scan-capture-in pointer-events-none relative w-full rounded-2xl border border-black/5 bg-white p-2.5 shadow-[0_10px_26px_rgba(15,39,31,0.12)]">
      <div className="mb-2 flex items-center justify-between gap-3 px-1">
        <div className="min-w-0">
          <p className="text-[10px] font-semibold uppercase tracking-[0.12em] text-muted">Website captures</p>
          {hostname ? <p className="truncate text-[11px] font-semibold text-ink">{hostname}</p> : null}
        </div>
        <span className="shrink-0 text-[10px] font-medium text-muted">
          {available.length === 2 ? "Desktop + mobile" : `${available[0].kind} view`}
        </span>
      </div>
      <div className="grid grid-cols-2 items-start gap-2">
        {available.slice(0, 2).map((capture) => (
          <figure
            key={`${capture.kind}-${capture.src}`}
            className={available.length === 1 ? "col-span-2 min-w-0" : "min-w-0"}
          >
            <div
              className={
                capture.kind === "desktop"
                  ? "overflow-hidden rounded-lg border border-black/10 bg-[#efebe6]"
                  : "mx-auto max-w-[92px] overflow-hidden rounded-[0.85rem] border-[3px] border-[#1a1a1a] bg-[#1a1a1a]"
              }
            >
              {capture.kind === "desktop" ? (
                <div className="flex h-3 items-center gap-1 border-b border-black/5 bg-white px-1.5" aria-hidden="true">
                  <span className="h-1 w-1 rounded-full bg-[#ee6a5f]" />
                  <span className="h-1 w-1 rounded-full bg-[#f3bd4f]" />
                  <span className="h-1 w-1 rounded-full bg-[#61c454]" />
                </div>
              ) : null}
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={capture.src}
                alt={`${capture.kind === "desktop" ? "Desktop" : "Mobile"} website capture`}
                className={
                  capture.kind === "desktop"
                    ? "h-[76px] w-full object-cover object-top sm:h-[86px] xl:h-[78px]"
                    : "h-[88px] w-full object-cover object-top sm:h-[96px] xl:h-[88px]"
                }
                onError={() => {
                  setFailedSources((current) => new Set(current).add(capture.src));
                }}
              />
            </div>
            <figcaption className="mt-1 text-center text-[9px] font-semibold capitalize text-ink/75">
              {capture.kind}
            </figcaption>
          </figure>
        ))}
      </div>
    </section>
  );
}

export default function ScanExperience({
  restaurantName = "Your restaurant",
  address,
  rating,
  website,
  placeId,
  latitude,
  longitude,
  photoUrl,
  photos = [],
  reviews = [],
  desktopScreenshot,
  mobileScreenshot,
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
      `${restaurantName} listing identity`,
      "Google business profile",
      "Photo quality and quantity",
      "Google review sentiment",
      websiteLabel.startsWith("http") ? websiteLabel : website || websiteLabel,
      "Mobile experience",
    ],
    [restaurantName, website, websiteLabel],
  );

  const [activeIndex, setActiveIndex] = useState(0);
  const [secondsLeft, setSecondsLeft] = useState(TARGET_SECONDS);
  const [finishing, setFinishing] = useState(false);
  const startedAtRef = useRef<number>(0);
  const evidenceReadyAtRef = useRef<number | null>(null);
  const onReadyRef = useRef(onReady);

  useEffect(() => {
    onReadyRef.current = onReady;
  }, [onReady]);

  const gallery = useMemo(
    () => normalizeScanPhotos(photoUrl, restaurantName, photos),
    [photoUrl, photos, restaurantName],
  );
  const availableReviews = useMemo(() => normalizeScanReviews(reviews), [reviews]);
  const websiteCaptures = useMemo(
    () => websiteCaptureEvidence(desktopScreenshot, mobileScreenshot),
    [desktopScreenshot, mobileScreenshot],
  );
  const hasEvidence = gallery.length > 0 || availableReviews.length > 0 || websiteCaptures.length > 0;

  useEffect(() => {
    if (hasEvidence && evidenceReadyAtRef.current === null) {
      evidenceReadyAtRef.current = Date.now();
    }
  }, [hasEvidence]);

  const searchQuery = useMemo(() => {
    const cityish = (address || "").split(",").slice(-2).join(",").trim();
    if (cityish) return `${restaurantName} in ${cityish}`;
    return restaurantName;
  }, [restaurantName, address]);

  const hasExactPin =
    typeof latitude === "number" &&
    typeof longitude === "number" &&
    Number.isFinite(latitude) &&
    Number.isFinite(longitude) &&
    Math.abs(latitude) <= 90 &&
    Math.abs(longitude) <= 180;

  const candidateEmbedSrc = useMemo(
    () => reportMapEmbedUrl({ restaurantName, address, placeId, latitude, longitude }),
    [restaurantName, address, placeId, latitude, longitude],
  );
  // Keep the first useful map stable, while still promoting exact coordinates once available.
  const [initialEmbedSrc] = useState(candidateEmbedSrc);
  const embedSrc =
    hasExactPin && candidateEmbedSrc
      ? candidateEmbedSrc
      : initialEmbedSrc ?? candidateEmbedSrc;

  // Countdown
  useEffect(() => {
    startedAtRef.current = Date.now();
    const id = window.setInterval(() => {
      const elapsed = Math.floor((Date.now() - startedAtRef.current) / 1000);
      setSecondsLeft(Math.max(0, TARGET_SECONDS - elapsed));
    }, 250);
    return () => window.clearInterval(id);
  }, []);

  // Advance the checklist while the evidence board stays visible throughout.
  useEffect(() => {
    if (finishing) return;
    const id = window.setInterval(() => {
      setActiveIndex((i) => Math.min(i + 1, steps.length - 1));
    }, STEP_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [finishing, steps.length, gallery.length]);

  // Once live evidence arrives, keep the board visible long enough for a photo
  // flip and a meaningful section of the review stream before handing off.
  useEffect(() => {
    let cancelled = false;
    let finishTimer: number | undefined;

    const tryFinish = () => {
      if (cancelled) return;
      const now = Date.now();
      const elapsed = now - startedAtRef.current;
      const minMet = elapsed >= MIN_SCAN_MS;
      const evidencePreviewMet = evidenceReadyAtRef.current === null ||
        now - evidenceReadyAtRef.current >= MIN_EVIDENCE_PREVIEW_MS;
      if (!fetchComplete || !minMet || !evidencePreviewMet) return;

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
  }, [fetchComplete, steps.length, gallery.length, availableReviews.length, websiteCaptures.length]);

  const progress = finishing
    ? 1
    : Math.min(0.96, ((TARGET_SECONDS - secondsLeft) / TARGET_SECONDS) * 0.96);

  const statusLine = finishing
    ? "Wrapping up your report…"
    : secondsLeft > 0
      ? `${secondsLeft} seconds remaining`
      : fetchComplete
        ? "Finalizing scores…"
        : "Almost done…";
  const evidenceSummary = [
    gallery.length > 0
      ? `${Math.min(PHOTO_CARD_LIMIT, gallery.length)} photo card${Math.min(PHOTO_CARD_LIMIT, gallery.length) === 1 ? "" : "s"}`
      : null,
    websiteCaptures.length > 0
      ? `${websiteCaptures.length} website view${websiteCaptures.length === 1 ? "" : "s"}`
      : null,
    availableReviews.length > 1
      ? `${REVIEW_STREAM_CARD_COUNT}-card review stream`
      : availableReviews.length === 1
        ? "1 available review"
        : null,
  ].filter(Boolean).join(" · ");

  return (
    <div className={`relative min-h-[calc(100dvh-4.5rem)] overflow-hidden bg-[#f3f1ed] xl:h-[calc(100dvh-4.5rem)] xl:min-h-0 ${className}`}>
      <div className="relative grid min-h-[calc(100dvh-4.5rem)] w-full lg:grid-cols-[minmax(260px,320px)_minmax(0,1fr)] xl:h-full xl:min-h-0">
        {/* Desktop — expanded checklist. Mobile keeps identity, map, media and progress above fold. */}
        <aside className="relative z-20 order-2 hidden flex-col border-b border-black/5 bg-white/95 px-6 py-7 backdrop-blur-md sm:px-7 lg:order-1 lg:flex lg:border-b-0 lg:border-r lg:py-10 xl:h-full xl:min-h-0">
          <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-muted">Digital footprint scan</p>
          <h1 className="mt-2 font-display text-[2rem] font-semibold leading-none tracking-[-0.04em] text-ink">
            Scanning…
          </h1>
          <p className="mt-3 text-[14px] leading-relaxed text-muted">
            Checking Google listing, reviews, photos, website &amp; mobile signals — live.
          </p>

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

        {/* Map and evidence share the canvas without overlaying the listing pin. */}
        <section className="relative order-1 grid min-h-[calc(100dvh-4.5rem)] w-full min-w-0 grid-cols-[minmax(0,1fr)] grid-rows-[auto_minmax(0,1fr)] overflow-hidden bg-[#e8e4dc] lg:order-2 lg:min-h-0 xl:h-full">
          <header className="relative z-20 border-b border-black/5 bg-white/95 px-4 py-3 backdrop-blur sm:px-6">
            <div className="mx-auto flex w-full max-w-xl items-center gap-3 rounded-full border border-black/5 bg-[#f7f4ef] px-4 py-2.5">
              <svg viewBox="0 0 20 20" className="h-4 w-4 shrink-0 text-muted" fill="none" aria-hidden="true">
                <circle cx="9" cy="9" r="5.5" stroke="currentColor" strokeWidth="1.8" />
                <path d="M13.5 13.5 17 17" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
              </svg>
              <p className="truncate text-[14px] font-medium text-ink">{searchQuery}</p>
            </div>
          </header>

          <div className="grid min-h-0 w-full min-w-0 grid-cols-1 xl:grid-cols-[minmax(0,1fr)_360px] 2xl:grid-cols-[minmax(0,1fr)_390px]">
            <div className="relative min-h-[340px] min-w-0 overflow-hidden sm:min-h-[420px] xl:min-h-0">
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
              <div
                className="pointer-events-none absolute inset-0 bg-gradient-to-b from-white/10 via-transparent to-white/10"
                aria-hidden="true"
              />
              <div className="pointer-events-none absolute inset-0 z-10 overflow-hidden" aria-hidden="true">
                <div className="scan-beam absolute left-0 right-0 h-[2px] opacity-55" />
              </div>
            </div>

            <aside className="relative z-20 flex min-h-0 min-w-0 flex-col gap-3 border-t border-black/5 bg-[#f7f4ef] px-3 py-3 xl:overflow-y-auto xl:border-l xl:border-t-0">
              <div className="flex items-center justify-between gap-3 px-1" role="status" aria-live="polite">
                <div className="min-w-0">
                  <p className="text-[10px] font-semibold uppercase tracking-[0.12em] text-muted">Live evidence board</p>
                  <p className="truncate text-[12px] font-semibold text-ink">
                    {evidenceSummary || (finishing
                      ? "Building your scorecard"
                      : hasExactPin
                        ? "Pinned listing · evidence is loading"
                        : "Gathering live listing signals")}
                  </p>
                </div>
                <span className="shrink-0 text-[11px] font-semibold tabular-nums text-primary lg:hidden">
                  <span className="sm:hidden">{secondsLeft > 0 ? `${secondsLeft}s` : finishing ? "Done" : "Finishing"}</span>
                  <span className="hidden sm:inline">{statusLine}</span>
                </span>
              </div>

              <div className="h-1 overflow-hidden rounded-full bg-[#e3ddd4] lg:hidden">
                <div
                  className="scan-progress-fill h-full rounded-full transition-[width] duration-500 ease-out"
                  style={{ width: `${Math.max(8, progress * 100)}%` }}
                />
              </div>

              {gallery.length > 0 ? <PhotoBoard photos={gallery} restaurantName={restaurantName} /> : null}
              {websiteCaptures.length > 0 ? (
                <WebsiteCaptureBoard
                  captures={websiteCaptures}
                  hostname={websiteLabel !== "Website / listing signals" ? websiteLabel : undefined}
                />
              ) : null}
              {availableReviews.length > 0 ? (
                <ReviewScroller reviews={availableReviews} placeRating={rating} variant="board" />
              ) : null}
              {!hasEvidence ? (
                <div className="flex min-h-[180px] flex-1 items-center justify-center rounded-2xl border border-dashed border-black/10 bg-white/60 px-5 text-center">
                  <div>
                    <Spinner />
                    <p className="mt-3 text-[12px] font-semibold text-ink">Verifying visual evidence</p>
                    <p className="mt-1 text-[11px] leading-relaxed text-muted">
                      Photo, website, and review cards appear only after their sources load.
                    </p>
                  </div>
                </div>
              ) : null}
            </aside>
          </div>
        </section>
      </div>
    </div>
  );
}
