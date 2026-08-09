"use client";

import { useEffect, useMemo, useRef, useState } from "react";

const MIN_SCAN_MS = 800;
const STEP_INTERVAL_MS = 2_200;
const FINISH_HOLD_MS = 320;
const TARGET_SECONDS = 15;

export type ScanPhoto = {
  src: string;
  label?: string;
};

export type ScanReview = {
  author?: string;
  text?: string;
  rating?: number;
  relativeTime?: string;
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
  latitude?: number;
  longitude?: number;
  photoUrl?: string;
  photos?: ScanPhoto[];
  /** Live Google reviews shown during the review-sentiment step. */
  reviews?: ScanReview[];
  /** Mobile viewport screenshot of the restaurant website (data URL or http). */
  mobileScreenshot?: string;
  /** Live restaurant website URL for the phone mockup (never Maps). */
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

function sentimentMeta(sentiment?: string) {
  const s = (sentiment || "mixed").toLowerCase();
  if (s === "positive") return { label: "Positive", className: "bg-[#e8f6ee] text-[#1f7a45]" };
  if (s === "negative") return { label: "Negative", className: "bg-[#fdecea] text-[#b42318]" };
  return { label: "Mixed", className: "bg-[#fff6e5] text-[#a15c00]" };
}

function ReviewCardsOverlay({
  reviews,
  visibleCount,
  placeRating,
}: {
  reviews: ScanReview[];
  visibleCount: number;
  placeRating?: number;
}) {
  const shown = reviews.slice(0, visibleCount);
  if (shown.length === 0) return null;

  const pos = reviews.filter((r) => (r.sentiment || "").toLowerCase() === "positive").length;
  const neg = reviews.filter((r) => (r.sentiment || "").toLowerCase() === "negative").length;
  const mix = Math.max(0, reviews.length - pos - neg);
  const overall =
    pos >= neg && pos >= mix ? "Mostly positive" : neg > pos ? "Needs attention" : "Mixed sentiment";

  return (
    <>
      {/* Sentiment summary chip */}
      <div className="scan-photo-float pointer-events-none absolute left-1/2 top-[4.75rem] z-40 w-[min(360px,calc(100%-2rem))] -translate-x-1/2 sm:top-[5.25rem]">
        <div className="flex items-center justify-between gap-3 rounded-2xl border border-black/5 bg-white/95 px-3.5 py-2.5 shadow-[0_14px_40px_rgba(15,39,31,0.16)] backdrop-blur">
          <div className="min-w-0">
            <p className="text-[10px] font-semibold uppercase tracking-[0.12em] text-muted">Google review sentiment</p>
            <p className="truncate text-[13px] font-semibold text-ink">{overall}</p>
          </div>
          <div className="flex shrink-0 items-center gap-2">
            {typeof placeRating === "number" ? (
              <span className="rounded-full bg-[#f4f0ea] px-2 py-1 text-[11px] font-semibold tabular-nums text-ink">
                {placeRating.toFixed(1)}★
              </span>
            ) : null}
            <span className="rounded-full bg-[#e8f6ee] px-2 py-1 text-[10px] font-semibold text-[#1f7a45]">
              {pos}↑
            </span>
            {neg > 0 ? (
              <span className="rounded-full bg-[#fdecea] px-2 py-1 text-[10px] font-semibold text-[#b42318]">
                {neg}↓
              </span>
            ) : null}
          </div>
        </div>
      </div>

      {shown.map((review, i) => {
        const slot = REVIEW_SLOTS[i % REVIEW_SLOTS.length];
        const meta = sentimentMeta(review.sentiment);
        return (
          <div
            key={`review-${i}-${review.author || "anon"}`}
            className="scan-photo-float pointer-events-none absolute z-40 w-[min(220px,42vw)] rounded-2xl border border-white/90 bg-white/95 p-3 shadow-[0_18px_50px_rgba(15,39,31,0.2)] backdrop-blur"
            style={{
              top: "top" in slot ? slot.top : undefined,
              bottom: "bottom" in slot ? slot.bottom : undefined,
              left: "left" in slot ? slot.left : undefined,
              right: "right" in slot ? slot.right : undefined,
              transform: `rotate(${slot.rotate})`,
              animationDelay: `${0.12 + i * 0.12}s`,
            }}
          >
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <p className="truncate text-[12px] font-semibold text-ink">{review.author || "Google reviewer"}</p>
                <div className="mt-0.5 flex items-center gap-1.5">
                  <Stars rating={review.rating} />
                  {review.relativeTime ? (
                    <span className="truncate text-[10px] text-muted">{review.relativeTime}</span>
                  ) : null}
                </div>
              </div>
              <span className={`shrink-0 rounded-full px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wide ${meta.className}`}>
                {meta.label}
              </span>
            </div>
            {review.text ? (
              <p className="mt-2 line-clamp-4 text-[11px] leading-snug text-ink/80">&ldquo;{review.text}&rdquo;</p>
            ) : null}
          </div>
        );
      })}
    </>
  );
}

function RevealingReviewCards({
  reviews,
  placeRating,
}: {
  reviews: ScanReview[];
  placeRating?: number;
}) {
  const [visibleCount, setVisibleCount] = useState(1);

  useEffect(() => {
    const id = window.setInterval(() => {
      setVisibleCount((count) => Math.min(count + 1, Math.min(4, reviews.length)));
    }, 1800);
    return () => window.clearInterval(id);
  }, [reviews.length]);

  return (
    <ReviewCardsOverlay
      reviews={reviews}
      visibleCount={visibleCount}
      placeRating={placeRating}
    />
  );
}

function MobilePhoneMockup({
  screenshot,
  hostname,
  websiteUrl,
}: {
  screenshot?: string;
  hostname?: string;
  websiteUrl?: string;
}) {
  const host =
    hostname ||
    (() => {
      if (!websiteUrl) return "";
      try {
        return new URL(websiteUrl).hostname.replace(/^www\./, "");
      } catch {
        return websiteUrl.replace(/^https?:\/\//, "").split("/")[0] || "";
      }
    })();

  const httpsSite = (() => {
    if (!websiteUrl) return "";
    const raw = websiteUrl.trim();
    if (!raw) return "";
    if (raw.startsWith("http://")) return `https://${raw.slice("http://".length)}`;
    if (!raw.includes("://")) return `https://${raw}`;
    return raw;
  })();

  return (
    <div className="scan-phone-mockup pointer-events-none absolute bottom-6 right-4 z-40 sm:bottom-8 sm:right-8">
      <div className="relative mx-auto w-[148px] sm:w-[168px]">
        <div className="relative overflow-hidden rounded-[1.65rem] border-[5px] border-[#1a1a1a] bg-[#1a1a1a] shadow-[0_28px_60px_rgba(15,39,31,0.38)]">
          <div className="absolute left-1/2 top-2 z-20 h-[14px] w-[52px] -translate-x-1/2 rounded-full bg-black" />
          <div className="relative aspect-[9/19.5] overflow-hidden rounded-[1.25rem] bg-[#f4f1ed]">
            <div className="absolute inset-x-0 top-0 z-10 border-b border-black/5 bg-white/95 px-2 pb-1.5 pt-5">
              <div className="truncate rounded-full bg-[#efebe6] px-2.5 py-1 text-center text-[7px] font-medium text-ink/70">
                {host || "website"}
              </div>
            </div>

            {screenshot ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={screenshot}
                alt=""
                className="absolute inset-0 h-full w-full object-cover object-top pt-10"
              />
            ) : httpsSite ? (
              <div className="absolute inset-0 overflow-hidden pt-10">
                <iframe
                  title={`Mobile preview of ${host}`}
                  src={httpsSite}
                  className="pointer-events-none absolute left-0 top-10 origin-top-left border-0 bg-white"
                  style={{ width: 390, height: 844, transform: "scale(0.43)" }}
                  loading="eager"
                  sandbox="allow-scripts allow-same-origin"
                />
              </div>
            ) : (
              <div className="absolute inset-0 flex items-center justify-center pt-10 text-[10px] text-muted">
                Loading site…
              </div>
            )}

            <div
              className="pointer-events-none absolute inset-0 bg-gradient-to-b from-white/10 via-transparent to-black/10"
              aria-hidden="true"
            />
          </div>
          <div className="absolute bottom-1.5 left-1/2 z-20 h-1 w-10 -translate-x-1/2 rounded-full bg-white/35" />
        </div>
        <p className="mt-2 max-w-[168px] truncate text-center text-[10px] font-semibold tracking-wide text-ink/80 drop-shadow-sm">
          {host ? `Mobile · ${host}` : "Mobile experience"}
        </p>
      </div>
    </div>
  );
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

const PHOTO_SLOTS = [
  { top: "12%", left: "8%", rotate: "-8deg", delay: 0 },
  { top: "18%", right: "10%", rotate: "7deg", delay: 1 },
  { bottom: "22%", left: "12%", rotate: "5deg", delay: 2 },
  { bottom: "16%", right: "8%", rotate: "-6deg", delay: 3 },
  { top: "42%", left: "4%", rotate: "3deg", delay: 4 },
  { top: "48%", right: "5%", rotate: "-4deg", delay: 5 },
] as const;

const REVIEW_SLOTS = [
  { top: "28%", left: "6%", rotate: "-3deg" },
  { top: "34%", right: "7%", rotate: "4deg" },
  { bottom: "28%", left: "10%", rotate: "2deg" },
  { bottom: "24%", right: "12%", rotate: "-5deg" },
] as const;

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
  const [secondsLeft, setSecondsLeft] = useState(TARGET_SECONDS);
  const [finishing, setFinishing] = useState(false);
  const [visiblePhotoCount, setVisiblePhotoCount] = useState(1);
  const startedAtRef = useRef<number>(0);
  const onReadyRef = useRef(onReady);

  useEffect(() => {
    onReadyRef.current = onReady;
  }, [onReady]);

  const gallery = useMemo(() => {
    const seen = new Set<string>();
    const out: ScanPhoto[] = [];
    const push = (src?: string, label?: string) => {
      const s = (src || "").trim();
      if (!s || seen.has(s)) return;
      seen.add(s);
      out.push({ src: s, label });
    };
    push(photoUrl, restaurantName);
    for (const p of photos) push(p.src, p.label);
    // No stock-photo fallback — wait for real Google listing media
    return out.slice(0, PHOTO_SLOTS.length);
  }, [photoUrl, photos, restaurantName]);

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

  const photoStepIndex = 2; // "Photo quality and quantity"
  const reviewStepIndex = 3; // "Google review sentiment"
  const mobileStepIndex = steps.length - 1;
  const displayedPhotoCount =
    activeIndex >= photoStepIndex
      ? gallery.length
      : Math.min(visiblePhotoCount, gallery.length);
  const showMobileMockup =
    (Boolean(mobileScreenshot) || Boolean(websiteUrl) || Boolean(website)) &&
    (finishing || activeIndex >= mobileStepIndex);
  // Exclusive phases: photos first → then reviews (never mixed on map)
  const showReviews =
    reviews.length > 0 &&
    !showMobileMockup &&
    !finishing &&
    activeIndex >= reviewStepIndex &&
    activeIndex < mobileStepIndex;
  const showPhotosOverlay =
    gallery.length > 0 && !showReviews && !showMobileMockup && !finishing;

  // Countdown
  useEffect(() => {
    startedAtRef.current = Date.now();
    const id = window.setInterval(() => {
      const elapsed = Math.floor((Date.now() - startedAtRef.current) / 1000);
      setSecondsLeft(Math.max(0, TARGET_SECONDS - elapsed));
    }, 250);
    return () => window.clearInterval(id);
  }, []);

  // Advance checklist; photos fill during early steps, reviews only after photo step
  useEffect(() => {
    if (finishing) return;
    const id = window.setInterval(() => {
      setActiveIndex((i) => Math.min(i + 1, steps.length - 1));
      setVisiblePhotoCount((n) => Math.min(n + 1, gallery.length));
    }, STEP_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [finishing, steps.length, gallery.length]);

  // Hand off as soon as the real response is ready; the short floor prevents a
  // jarring flash on cached responses without manufacturing a long scan.
  useEffect(() => {
    let cancelled = false;
    let finishTimer: number | undefined;

    const tryFinish = () => {
      if (cancelled) return;
      const elapsed = Date.now() - startedAtRef.current;
      const minMet = elapsed >= MIN_SCAN_MS;
      if (!fetchComplete || !minMet) return;

      setFinishing(true);
      setActiveIndex(steps.length);
      setVisiblePhotoCount(gallery.length);

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
  }, [fetchComplete, steps.length, gallery.length, reviews.length]);

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
            Checking Google rankings, reviews, photos, website &amp; mobile signals — live.
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

          {/* Google reviews — only after photos phase */}
          {showReviews ? (
            <RevealingReviewCards
              reviews={reviews}
              placeRating={rating}
            />
          ) : null}

          {/* Scraped photos — hidden once reviews / mobile take over */}
          {showPhotosOverlay
            ? gallery.slice(0, displayedPhotoCount).map((photo, i) => {
                const slot = PHOTO_SLOTS[i % PHOTO_SLOTS.length];
                return (
                  <div
                    key={`${photo.src}-${i}`}
                    className="scan-photo-float pointer-events-none absolute z-30 overflow-hidden rounded-2xl border border-white/80 bg-white shadow-[0_18px_50px_rgba(15,39,31,0.22)]"
                    style={{
                      top: "top" in slot ? slot.top : undefined,
                      bottom: "bottom" in slot ? slot.bottom : undefined,
                      left: "left" in slot ? slot.left : undefined,
                      right: "right" in slot ? slot.right : undefined,
                      width: i % 3 === 0 ? "118px" : "102px",
                      transform: `rotate(${slot.rotate})`,
                      animationDelay: `${slot.delay * 0.08}s`,
                    }}
                  >
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={photo.src} alt="" className="h-[118px] w-full object-cover" />
                    {photo.label ? (
                      <p className="truncate bg-white/95 px-2 py-1 text-[10px] font-semibold text-ink">
                        {photo.label}
                      </p>
                    ) : null}
                  </div>
                );
              })
            : null}

          {/* Mobile experience — live website screenshot in phone mockup */}
          {showMobileMockup ? (
            <MobilePhoneMockup
              screenshot={mobileScreenshot}
              hostname={websiteLabel !== "Website / listing signals" ? websiteLabel : undefined}
              websiteUrl={websiteUrl || website}
            />
          ) : null}

          {/* Scan beam across map */}
          <div className="pointer-events-none absolute inset-x-0 top-0 z-10 h-full overflow-hidden" aria-hidden="true">
            <div className="scan-beam absolute left-0 right-0 h-[2px] opacity-70" />
          </div>

          <div className="absolute bottom-4 left-4 z-20 hidden rounded-full bg-white/90 px-3 py-1.5 text-[11px] font-medium text-muted shadow-sm lg:block">
            {showMobileMockup
              ? "Checking mobile experience…"
              : showReviews
                ? "Reading Google review sentiment…"
                : showPhotosOverlay
                  ? "Scanning listing photos…"
                  : hasExactPin
                    ? "Google Maps · pinned to listing"
                    : "Live Google listing scan"}
          </div>

          <div
            className="absolute inset-x-4 bottom-24 z-50 rounded-2xl border border-black/5 bg-white/95 p-4 shadow-[0_18px_50px_rgba(15,39,31,0.2)] backdrop-blur lg:hidden"
            role="status"
            aria-live="polite"
          >
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
          </div>
        </section>
      </div>
    </div>
  );
}
