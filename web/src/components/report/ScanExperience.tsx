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

const MIN_SCAN_MS = 24_000;
const MIN_EVIDENCE_PREVIEW_MS = 16_500;
const STEP_INTERVAL_MS = 4_000;
const FINISH_HOLD_MS = 320;
const TARGET_SECONDS = 24;
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

function useReadyEvidence<T extends { src: string }>(items: T[]) {
  const [readySources, setReadySources] = useState<Set<string>>(() => new Set());
  const [failedSources, setFailedSources] = useState<Set<string>>(() => new Set());

  useEffect(() => {
    let cancelled = false;
    const loaders = items.map((item) => {
      const image = new window.Image();
      image.onload = () => {
        if (!cancelled) setReadySources((current) => new Set(current).add(item.src));
      };
      image.onerror = () => {
        if (!cancelled) setFailedSources((current) => new Set(current).add(item.src));
      };
      image.src = item.src;
      return image;
    });

    return () => {
      cancelled = true;
      for (const image of loaders) {
        image.onload = null;
        image.onerror = null;
      }
    };
  }, [items]);

  const available = useMemo(
    () => items.filter((item) => readySources.has(item.src) && !failedSources.has(item.src)),
    [failedSources, items, readySources],
  );

  return {
    available,
    markFailed: (src: string) => {
      setFailedSources((current) => new Set(current).add(src));
    },
  };
}

function PhotoBoard({
  photos,
  restaurantName,
  onImageError,
}: {
  photos: ScanPhoto[];
  restaurantName: string;
  onImageError: (src: string) => void;
}) {
  const [turn, setTurn] = useState(0);
  const [paused, setPaused] = useState(false);
  const [reducedMotion, setReducedMotion] = useState(false);
  const photoSlots = useMemo(
    () => buildScanPhotoSlots(photos, PHOTO_CARD_LIMIT),
    [photos],
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

  const rotations = [-2.4, 1.6, -1.2, 2.1, -1.8, 1.3];

  return (
    <section
      className="scan-stage-in relative mx-auto flex h-full w-full max-w-[880px] flex-col items-center justify-center px-4 pb-28 pt-16 sm:px-7 sm:pt-20"
      aria-label={`${photoSlots.length} photo cards showing ${photos.length} available restaurant listing photos`}
    >
      <div className="absolute inset-x-4 top-5 flex items-start justify-between gap-3 sm:inset-x-7 sm:top-7">
        <div className="rounded-full border border-black/5 bg-white/92 px-3 py-2 shadow-sm backdrop-blur">
          <p className="text-[10px] font-semibold uppercase tracking-[0.12em] text-muted">Restaurant photos</p>
          <p className="text-[12px] font-semibold text-ink">
            {photoSlots.length} live card{photoSlots.length === 1 ? "" : "s"} · {photos.length} photo{photos.length === 1 ? "" : "s"} found
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-1.5">
          {hasOverflow && !reducedMotion ? (
            <button
              type="button"
              aria-pressed={paused}
              onClick={() => setPaused((current) => !current)}
              className="min-h-9 rounded-full border border-black/10 bg-white/95 px-3 text-[10px] font-semibold text-primary shadow-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
            >
              {paused ? "Play" : "Pause"}
            </button>
          ) : null}
          {hasOverflow ? (
            <button
              type="button"
              onClick={() => setTurn((current) => current + 1)}
              className="min-h-9 rounded-full border border-black/10 bg-white/95 px-3 text-[10px] font-semibold text-primary shadow-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
            >
              Next
            </button>
          ) : null}
        </div>
      </div>
      <div className="grid w-full grid-cols-2 items-center gap-2.5 sm:grid-cols-3 sm:gap-4">
        {photoSlots.map((slotPhotos, slotIndex) => {
          const photoIndex = turn % slotPhotos.length;
          const currentPhoto = slotPhotos[photoIndex];
          const sourceIndex = photos.findIndex((photo) => photo.src === currentPhoto.src);
          const style = {
            "--scan-photo-delay": `${slotIndex * 65}ms`,
            "--scan-photo-rotation": `${rotations[slotIndex] || 0}deg`,
          } as CSSProperties;
          return (
            <figure
              key={`slot-${slotIndex}`}
              className="scan-photo-collage-card min-w-0 overflow-hidden rounded-[1.1rem] border-[5px] border-white bg-white shadow-[0_16px_38px_rgba(31,42,37,0.18)]"
              style={style}
            >
              <div
                key={currentPhoto.src}
                className={slotPhotos.length > 1 ? "scan-photo-flip" : "scan-photo-in"}
              >
                {/* Dynamic Google media cannot use next/image's fixed remote allow-list. */}
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={currentPhoto.src}
                  alt={currentPhoto.label || `${restaurantName} listing photo ${sourceIndex + 1}`}
                  className="h-[116px] w-full object-cover sm:h-[152px] lg:h-[168px] xl:h-[154px]"
                  onError={() => onImageError(currentPhoto.src)}
                />
                <figcaption className="flex items-center justify-between gap-1 bg-white px-2.5 py-2 text-[9px] font-semibold text-ink sm:text-[10px]">
                  <span className="truncate">{currentPhoto.label || "Listing photo"}</span>
                  <span className="shrink-0 tabular-nums text-muted">{sourceIndex + 1}/{photos.length}</span>
                </figcaption>
              </div>
            </figure>
          );
        })}
      </div>
    </section>
  );
}

function ReviewCard({ review, index, compact = false }: { review: ScanReview; index: number; compact?: boolean }) {
  const meta = review.sentiment ? sentimentMeta(review.sentiment) : null;
  return (
    <article className={`rounded-xl border border-black/5 bg-white shadow-sm ${compact ? "p-2.5" : "p-3"}`}>
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className={`truncate font-semibold text-ink ${compact ? "text-[11px]" : "text-[12px]"}`}>
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
        <p className={`${compact ? "mt-1.5 line-clamp-2 text-[10px]" : "mt-2 line-clamp-3 text-[11px]"} leading-snug text-ink/80`}>&ldquo;{review.text}&rdquo;</p>
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
  variant?: "default" | "board" | "stage";
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
        variant === "stage"
          ? "scan-stage-in h-[470px] max-w-[560px] shrink-0 sm:h-[500px]"
          : variant === "board"
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
          variant === "board" || variant === "stage" ? "flex-1" : "h-[126px] sm:h-[142px]"
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
                <ReviewCard review={item.review} index={item.sourceIndex} compact={variant === "stage"} />
              </div>
            ))}
          </div>
          {canScroll ? (
            <div className="scan-review-copy space-y-2 py-2.5" aria-hidden="true">
              {stream.map((item, index) => (
                <ReviewCard key={`review-copy-${index}`} review={item.review} index={item.sourceIndex} compact={variant === "stage"} />
              ))}
            </div>
          ) : null}
        </div>
      </div>
    </section>
  );
}

function WebsiteCaptureStage({
  captures,
  hostname,
  mode,
  onImageError,
}: {
  captures: WebsiteCaptureEvidence[];
  hostname?: string;
  mode: "desktop" | "mobile";
  onImageError: (src: string) => void;
}) {
  const desktop = captures.find((capture) => capture.kind === "desktop");
  const mobile = captures.find((capture) => capture.kind === "mobile");
  const primary = mode === "mobile" ? mobile || desktop : desktop || mobile;
  if (!primary) return null;

  return (
    <section className="scan-stage-in relative mx-auto flex h-full w-full max-w-[780px] items-center justify-center px-5 pb-28 pt-16 sm:px-8 sm:pt-20">
      <div className="absolute left-5 top-5 rounded-full border border-black/5 bg-white/92 px-3 py-2 shadow-sm backdrop-blur sm:left-8 sm:top-7">
        <p className="text-[10px] font-semibold uppercase tracking-[0.12em] text-muted">
          {mode === "mobile" ? "Mobile experience" : "Website review"}
        </p>
        <p className="max-w-[230px] truncate text-[12px] font-semibold text-ink">
          {hostname || (primary.kind === "mobile" ? "Mobile website capture" : "Desktop website capture")}
        </p>
      </div>

      {mode === "desktop" && desktop ? (
        <figure className="relative w-[92%] max-w-[590px]">
          <div className="overflow-hidden rounded-xl border border-black/10 bg-white shadow-[0_24px_60px_rgba(31,42,37,0.2)]">
            <div className="flex h-6 items-center gap-1.5 border-b border-black/5 bg-white px-3" aria-hidden="true">
              <span className="h-2 w-2 rounded-full bg-[#ee6a5f]" />
              <span className="h-2 w-2 rounded-full bg-[#f3bd4f]" />
              <span className="h-2 w-2 rounded-full bg-[#61c454]" />
            </div>
            {/* Dynamic website captures can be data URLs and cannot use next/image. */}
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={desktop.src}
              alt="Desktop website capture"
              className="aspect-[16/9] w-full object-cover object-top"
              onError={() => onImageError(desktop.src)}
            />
          </div>
          {mobile ? (
            <div className="absolute -bottom-16 -left-3 w-[84px] overflow-hidden rounded-[1rem] border-[4px] border-[#202523] bg-[#202523] shadow-[0_18px_42px_rgba(31,42,37,0.28)] sm:-bottom-20 sm:-left-10 sm:w-[118px]">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={mobile.src}
                alt="Mobile website capture"
                className="aspect-[9/16] w-full object-cover object-top"
                onError={() => onImageError(mobile.src)}
              />
            </div>
          ) : null}
          <figcaption className="mt-3 text-center text-[11px] font-semibold text-muted">Desktop website view</figcaption>
        </figure>
      ) : (
        <figure className="w-[150px] sm:w-[190px]">
          <div className="overflow-hidden rounded-[1.5rem] border-[6px] border-[#202523] bg-[#202523] shadow-[0_24px_60px_rgba(31,42,37,0.24)]">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={primary.src}
              alt={`${primary.kind === "mobile" ? "Mobile" : "Desktop fallback"} website capture`}
              className="aspect-[9/16] w-full object-cover object-top"
              onError={() => onImageError(primary.src)}
            />
          </div>
          <figcaption className="mt-3 text-center text-[11px] font-semibold text-muted">
            {primary.kind === "mobile" ? "Mobile website view" : "Desktop capture · mobile view unavailable"}
          </figcaption>
        </figure>
      )}
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
      websiteLabel,
      "Mobile experience",
    ],
    [restaurantName, websiteLabel],
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
  const { available: readyGallery, markFailed: markPhotoFailed } = useReadyEvidence(gallery);
  const { available: readyWebsiteCaptures, markFailed: markCaptureFailed } = useReadyEvidence(websiteCaptures);
  const hasEvidence = readyGallery.length > 0 || availableReviews.length > 0 || readyWebsiteCaptures.length > 0;

  useEffect(() => {
    if (hasEvidence && evidenceReadyAtRef.current === null) {
      evidenceReadyAtRef.current = Date.now();
      setActiveIndex((current) => current > 2 ? 2 : current);
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

  // Give every phase a full dwell. Resetting a phase restarts its full timeout.
  useEffect(() => {
    if (finishing) return;
    const id = window.setTimeout(() => {
      setActiveIndex((i) => Math.min(i + 1, steps.length - 1));
    }, STEP_INTERVAL_MS);
    return () => window.clearTimeout(id);
  }, [activeIndex, finishing, steps.length]);

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
  }, [fetchComplete, steps.length, readyGallery.length, availableReviews.length, readyWebsiteCaptures.length]);

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
  const stageIndex = Math.min(activeIndex, steps.length - 1);
  const stageStatus = [
    `Scanning ${restaurantName} listing`,
    "Scanning Google business profile",
    "Scanning restaurant photos",
    "Scanning available Google reviews",
    "Scanning desktop website",
    "Scanning mobile experience",
  ][stageIndex];

  const evidenceSummary = [
    readyGallery.length > 0
      ? `${Math.min(PHOTO_CARD_LIMIT, readyGallery.length)} photo card${Math.min(PHOTO_CARD_LIMIT, readyGallery.length) === 1 ? "" : "s"}`
      : null,
    readyWebsiteCaptures.length > 0
      ? `${readyWebsiteCaptures.length} website view${readyWebsiteCaptures.length === 1 ? "" : "s"}`
      : null,
    availableReviews.length > 1
      ? `${REVIEW_STREAM_CARD_COUNT}-card review stream from ${availableReviews.length} available review${availableReviews.length === 1 ? "" : "s"}`
      : availableReviews.length === 1
        ? "1 available review"
        : null,
  ].filter(Boolean).join(" · ");

  const waitingStage = (title: string, detail: string) => (
    <div className="scan-stage-in flex h-full items-center justify-center px-6 pb-28 text-center">
      <div className="rounded-2xl border border-black/5 bg-white/92 px-6 py-5 shadow-[0_16px_42px_rgba(31,42,37,0.12)]">
        <div className="mx-auto flex w-fit items-center justify-center"><Spinner /></div>
        <p className="mt-3 text-[13px] font-semibold text-ink">{title}</p>
        <p className="mt-1 max-w-[320px] text-[11px] leading-relaxed text-muted">{detail}</p>
      </div>
    </div>
  );

  const stageContent = (() => {
    if (stageIndex <= 1) {
      return (
        <div className="relative h-full min-h-0 overflow-hidden bg-[#e8e4dc]">
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
            waitingStage("Locating on Google Maps", "The map appears as soon as a usable listing location is available.")
          )}
          <div className="absolute inset-x-4 top-4 z-20 sm:inset-x-6 sm:top-6">
            <div className="mx-auto flex w-full max-w-xl items-center gap-3 rounded-full border border-black/5 bg-white/95 px-4 py-3 shadow-[0_10px_28px_rgba(31,42,37,0.12)] backdrop-blur">
              <svg viewBox="0 0 20 20" className="h-4 w-4 shrink-0 text-muted" fill="none" aria-hidden="true">
                <circle cx="9" cy="9" r="5.5" stroke="currentColor" strokeWidth="1.8" />
                <path d="M13.5 13.5 17 17" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
              </svg>
              <p className="truncate text-[14px] font-medium text-ink">{searchQuery}</p>
            </div>
          </div>
        </div>
      );
    }

    if (stageIndex === 2) {
      return readyGallery.length > 0 ? (
        <PhotoBoard photos={readyGallery} restaurantName={restaurantName} onImageError={markPhotoFailed} />
      ) : waitingStage(
        fetchComplete ? "No usable listing photos found" : "Loading restaurant photos",
        "Photo cards appear only after a real image loads successfully.",
      );
    }

    if (stageIndex === 3) {
      return availableReviews.length > 0 ? (
        <div className="flex h-full items-center justify-center px-4 pb-28 pt-16 sm:px-8 sm:pt-20">
          <ReviewScroller reviews={availableReviews} placeRating={rating} variant="stage" />
        </div>
      ) : waitingStage(
        fetchComplete ? "No Google review text returned" : "Loading Google review evidence",
        "Tuvi shows only genuine review evidence returned for this listing.",
      );
    }

    if (stageIndex === 4) {
      return readyWebsiteCaptures.length > 0 ? (
        <WebsiteCaptureStage
          captures={readyWebsiteCaptures}
          hostname={websiteLabel !== "Website / listing signals" ? websiteLabel : undefined}
          mode="desktop"
          onImageError={markCaptureFailed}
        />
      ) : waitingStage(
        fetchComplete ? "Desktop website capture unavailable" : "Capturing the restaurant website",
        "The website card appears only after a genuine browser capture succeeds.",
      );
    }

    return readyWebsiteCaptures.length > 0 ? (
      <WebsiteCaptureStage
        captures={readyWebsiteCaptures}
        hostname={websiteLabel !== "Website / listing signals" ? websiteLabel : undefined}
        mode="mobile"
        onImageError={markCaptureFailed}
      />
    ) : waitingStage(
      fetchComplete ? "Mobile website capture unavailable" : "Capturing the mobile experience",
      "Tuvi will not show a placeholder when a mobile capture is unavailable.",
    );
  })();

  return (
    <div className={`relative min-h-[calc(100dvh-4.5rem)] overflow-hidden bg-[#fbf8f5] xl:h-[calc(100dvh-4.5rem)] xl:min-h-0 ${className}`}>
      <div className="relative grid min-h-[calc(100dvh-4.5rem)] w-full lg:grid-cols-[minmax(260px,320px)_minmax(0,1fr)] xl:h-full xl:min-h-0">
        <aside className="relative z-20 hidden flex-col border-r border-black/5 bg-white/95 px-6 py-9 backdrop-blur-md sm:px-7 lg:flex xl:h-full xl:min-h-0">
          <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-muted">Digital footprint scan</p>
          <h1 className="mt-2 font-display text-[2rem] font-semibold leading-none tracking-[-0.04em] text-ink">Scanning…</h1>
          <p className="mt-3 text-[14px] leading-relaxed text-muted">
            Reviewing the listing, real photos, Google reviews, and desktop and mobile website captures.
          </p>

          <ul className="mt-8 flex-1 space-y-3.5">
            {steps.map((label, index) => {
              const done = finishing || index < activeIndex;
              const active = !done && index === stageIndex;
              return (
                <li
                  key={`${index}-${label}`}
                  className={`flex items-start gap-3 rounded-xl px-2 py-1.5 transition-colors duration-300 ${active ? "bg-[#f4f7f5]" : ""}`}
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
                  <span className={`pt-0.5 text-[14px] leading-snug ${done || active ? "font-medium text-ink" : "text-muted/55"}`}>
                    {label}
                  </span>
                </li>
              );
            })}
          </ul>

          <div className="mt-7 border-t border-black/5 pt-5">
            <p className="text-[10px] font-semibold uppercase tracking-[0.12em] text-muted">Evidence loaded</p>
            <p className="mt-1 text-[12px] leading-relaxed text-ink/75">
              {evidenceSummary || (hasExactPin ? "Listing pinned · visual evidence is loading" : "Gathering live listing signals")}
            </p>
          </div>
        </aside>

        <main className="relative min-h-[calc(100dvh-4.5rem)] min-w-0 overflow-hidden bg-[#fbf8f5] lg:min-h-0 xl:h-full">
          <div key={stageIndex} className="h-full min-h-[calc(100dvh-4.5rem)] lg:min-h-0">
            {stageContent}
          </div>

          <div className="pointer-events-none absolute inset-0 z-20 overflow-hidden" aria-hidden="true">
            <div className="scan-beam absolute left-0 right-0 h-[2px] opacity-60" />
          </div>

          <div className="absolute inset-x-3 bottom-3 z-30 overflow-hidden rounded-2xl border border-black/5 bg-white/96 shadow-[0_14px_36px_rgba(31,42,37,0.16)] backdrop-blur sm:inset-x-5 sm:bottom-5">
            <div className="h-1 bg-[#e7e1d9]">
              <div
                className="scan-progress-fill h-full transition-[width] duration-700 ease-out"
                style={{ width: `${Math.max(6, progress * 100)}%` }}
              />
            </div>
            <div className="flex items-center gap-3 px-4 py-3.5 sm:px-5">
              <Spinner />
              <div className="min-w-0 flex-1" role="status" aria-live="polite">
                <p className="truncate text-[13px] font-semibold text-ink">{stageStatus}</p>
                <p className="mt-0.5 text-[11px] font-medium tabular-nums text-muted">{statusLine}</p>
              </div>
              <span className="shrink-0 rounded-full bg-[#edf4f0] px-2.5 py-1 text-[10px] font-semibold tabular-nums text-primary">
                {finishing ? "Done" : `${stageIndex + 1}/${steps.length}`}
              </span>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
