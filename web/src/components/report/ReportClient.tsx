"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import type { MediaCard, PlaceAttribution, RestaurantDetails } from "@/lib/places";
import type { RestaurantReport } from "@/lib/report";
import { parsePreviewCoordinates } from "@/lib/report-preview";
import LiveCompetitorsCard from "@/components/report/LiveCompetitorsCard";
import LiveHealthCard from "@/components/report/LiveHealthCard";
import LiveIssuesCard from "@/components/report/LiveIssuesCard";
import LiveListingMedia from "@/components/report/LiveListingMedia";
import LivePresenceEvidence from "@/components/report/LivePresenceEvidence";
import LockedBlur from "@/components/report/LockedBlur";
import ReportSection from "@/components/report/ReportSection";
import ScanExperience, { type ScanPhoto } from "@/components/report/ScanExperience";
import UnlockReportDialog from "@/components/report/UnlockReportDialog";

type Payload = {
  place: RestaurantDetails;
  report: RestaurantReport;
};

function boundedRetryAfterMs(raw: string | null): number {
  const seconds = Number.parseInt(raw || "", 10);
  if (!Number.isFinite(seconds) || seconds < 1) return 3_000;
  return Math.min(seconds * 1_000, 5_000);
}

function waitForRetry(ms: number, signal: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    const timer = window.setTimeout(() => {
      signal.removeEventListener("abort", handleAbort);
      resolve();
    }, ms);
    const handleAbort = () => {
      window.clearTimeout(timer);
      reject(new DOMException("The report retry was cancelled.", "AbortError"));
    };
    signal.addEventListener("abort", handleAbort, { once: true });
  });
}

function reportMapSrc(place: RestaurantDetails): string | null {
  const { latitude, longitude } = place;
  if (
    typeof latitude === "number" &&
    typeof longitude === "number" &&
    Number.isFinite(latitude) &&
    Number.isFinite(longitude)
  ) {
    const params = new URLSearchParams({
      q: `${latitude},${longitude}`,
      ll: `${latitude},${longitude}`,
      z: "17",
      hl: "en",
      output: "embed",
    });
    return `https://www.google.com/maps?${params.toString()}`;
  }

  const query = [place.name, place.address].filter(Boolean).join(", ");
  if (!query) return null;
  const params = new URLSearchParams({ q: query, z: "16", hl: "en", output: "embed" });
  if (place.placeId) params.set("query_place_id", place.placeId);
  return `https://www.google.com/maps?${params.toString()}`;
}

function listingPhotoSrc(card?: MediaCard): string | null {
  if (!card) return null;
  if (card.photoName) {
    return `/api/restaurants/photo?name=${encodeURIComponent(card.photoName)}&max=720`;
  }
  return card.imageUrl || null;
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

function ListingPhoto({ card, restaurantName }: { card?: MediaCard; restaurantName: string }) {
  const src = listingPhotoSrc(card);
  const sourceLabel = card?.subtitle || (card?.photoName ? "From Google listing" : "");
  const sourceHref = safeExternalHref(card?.googleMapsUri || card?.href);
  const flagContentHref = safeExternalHref(card?.flagContentUri);
  const attributions = card?.authorAttributions || [];
  return (
    <div className="relative h-full min-h-0 overflow-hidden rounded-2xl bg-[#e8e2da]">
      {src ? (
        <Image
          src={src}
          alt={`${restaurantName} listing photo${card?.label ? ` — ${card.label}` : ""}`}
          fill
          sizes="(max-width: 640px) 38vw, 280px"
          className="object-cover"
          unoptimized
          preload
        />
      ) : (
        <div className="flex h-full items-center justify-center px-3 text-center text-[12px] font-medium text-muted">
          Listing photo unavailable
        </div>
      )}
      {card?.label || sourceLabel ? (
        <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/80 to-transparent px-3 pb-2.5 pt-9 text-white">
          {card?.label ? <p className="truncate text-[12px] font-semibold">{card.label}</p> : null}
          {attributions.length > 0 ? (
            <div className="mt-0.5 flex flex-wrap items-center gap-x-1.5 gap-y-1 text-[12px] font-medium leading-4 text-white/85">
              <span>Photo by</span>
              {attributions.map((attribution, index) => {
                const href = safeExternalHref(attribution.uri);
                const photoHref = safeExternalHref(attribution.photoUri);
                const name = attribution.displayName || "Google contributor";
                return (
                  <span key={`${name}-${index}`} className="inline-flex items-center gap-1">
                    {index > 0 ? <span aria-hidden="true">·</span> : null}
                    {photoHref ? (
                      // Google photo contributor avatar remains live and unoptimized.
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={photoHref}
                        alt=""
                        className="h-4 w-4 rounded-full object-cover"
                        referrerPolicy="no-referrer"
                      />
                    ) : null}
                    {href ? (
                      <a
                        href={href}
                        target="_blank"
                        rel="noreferrer"
                        className="underline decoration-white/50 underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-white"
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
          ) : sourceLabel ? (
            <p className="mt-0.5 truncate text-[12px] font-medium text-white/80">{sourceLabel}</p>
          ) : null}
          {sourceHref || flagContentHref ? (
            <div className="mt-1.5 flex flex-wrap items-center gap-x-2 gap-y-1 text-[12px] font-semibold">
              {sourceHref ? (
                <a
                  href={sourceHref}
                  target="_blank"
                  rel="noreferrer"
                  className="underline decoration-white/50 underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-white"
                >
                  {card?.googleMapsUri ? "Google Maps" : "Photo source"}
                </a>
              ) : null}
              {flagContentHref ? (
                <a
                  href={flagContentHref}
                  target="_blank"
                  rel="noreferrer"
                  className="underline decoration-white/50 underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-white"
                >
                  Report photo
                </a>
              ) : null}
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

function PlaceDataAttribution({
  attributions,
  mapsUri,
  showGoogleMaps,
}: {
  attributions?: PlaceAttribution[];
  mapsUri?: string;
  showGoogleMaps: boolean;
}) {
  if (!showGoogleMaps && !attributions?.length) return null;
  const googleMapsHref = safeExternalHref(mapsUri) || "https://www.google.com/maps";
  return (
    <div
      className="mt-2 flex flex-wrap items-center gap-x-2 gap-y-1 text-[12px] font-normal leading-4 text-[#5e5e5e]"
      translate="no"
    >
      {showGoogleMaps ? (
        <a
          href={googleMapsHref}
          target="_blank"
          rel="noreferrer"
          className="underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
        >
          Google Maps
        </a>
      ) : null}
      {attributions?.map((attribution, index) => {
        const providerHref = safeExternalHref(attribution.providerUri);
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
  );
}

export default function ReportClient({ placeId }: { placeId: string }) {
  const searchParams = useSearchParams();
  const nameFromQuery = (searchParams.get("name") || "").trim();
  const addressFromQuery = (searchParams.get("address") || "").trim();
  const previewCoordinates = parsePreviewCoordinates(
    searchParams.get("lat"),
    searchParams.get("lng"),
  );
  const latPreview = previewCoordinates.latitude;
  const lngPreview = previewCoordinates.longitude;
  const [data, setData] = useState<Payload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [phase, setPhase] = useState<"scan" | "ready">("scan");
  const [fetchComplete, setFetchComplete] = useState(false);
  const bootstrappingRef = useRef(true);
  const activePlaceIdRef = useRef<string | null>(null);
  const [unlockOpen, setUnlockOpen] = useState(false);
  const [retryKey, setRetryKey] = useState(0);
  const [canRetry, setCanRetry] = useState(false);

  useEffect(() => {
    let cancelled = false;
    let timedOut = false;
    const controller = new AbortController();
    const timeout = window.setTimeout(() => {
      timedOut = true;
      controller.abort();
    }, 18_000);
    if (activePlaceIdRef.current !== placeId) {
      activePlaceIdRef.current = placeId;
      bootstrappingRef.current = true;
    }

    async function load() {
      const quiet = !bootstrappingRef.current;
      if (!quiet) {
        setPhase("scan");
        setFetchComplete(false);
        setData(null);
        setError(null);
        setCanRetry(false);
      }
      try {
        let res: Response | null = null;
        let json: unknown = {};
        for (let attempt = 0; attempt < 2; attempt += 1) {
          res = await fetch(`/api/restaurants/${encodeURIComponent(placeId)}`, {
            signal: controller.signal,
            cache: "no-store",
          });
          json = await res.json().catch(() => ({}));
          if (res.status !== 503 || attempt === 1) break;
          await waitForRetry(boundedRetryAfterMs(res.headers.get("retry-after")), controller.signal);
        }
        if (!res) throw new Error("Failed to load report");
        if (!res.ok) {
          const payload = json && typeof json === "object" ? (json as { error?: string }) : {};
          throw new Error(payload.error || "Failed to load report");
        }
        if (!cancelled) {
          const payload = json as Payload;
          if (!payload?.report || !payload?.place) {
            throw new Error("Invalid report payload from server");
          }
          setData(payload);
          if (quiet) {
            setPhase("ready");
          } else {
            // Defer so ScanExperience mounts/updates with fetchComplete=true reliably
            window.setTimeout(() => {
              if (!cancelled) setFetchComplete(true);
            }, 50);
          }
          bootstrappingRef.current = false;
        }
      } catch (e) {
        if (cancelled) return;
        if (timedOut) {
          setError("The live scan timed out before all signals finished. Please try again.");
          setPhase("ready");
          setFetchComplete(true);
          bootstrappingRef.current = false;
          setCanRetry(true);
          return;
        }
        if (controller.signal.aborted) return;
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "Failed to load report");
          setPhase("ready");
          bootstrappingRef.current = false;
          setCanRetry(true);
        }
      } finally {
        window.clearTimeout(timeout);
      }
    }
    load();
    return () => {
      cancelled = true;
      window.clearTimeout(timeout);
      controller.abort();
    };
  }, [placeId, retryKey]);

  const handleScanReady = useCallback(() => {
    setPhase("ready");
  }, []);

  const retryReport = useCallback(() => {
    bootstrappingRef.current = true;
    setPhase("scan");
    setFetchComplete(false);
    setData(null);
    setError(null);
    setCanRetry(false);
    setRetryKey((key) => key + 1);
  }, []);

  if (phase === "scan" && !error) {
    const previewName = data?.place.name || nameFromQuery || "Your restaurant";
    const categoryRaw = (
      data?.report.competitorScan?.cuisine ||
      data?.place.primaryType ||
      data?.place.types?.[0] ||
      "Restaurant"
    ).replace(/_/g, " ");
    const category = categoryRaw.replace(/\b\w/g, (c) => c.toUpperCase());

    const photoCards = [
      ...(data?.place.media?.menuAndHighlights || []),
      ...(data?.place.media?.photosAndVideos || []),
    ];
    const listingMapsUrl = data?.place.media?.mapsUri || data?.place.mapsUri;
    const scanPhotos = photoCards
      .map((card): ScanPhoto | null => {
        const src = card.imageUrl
          ? card.imageUrl
          : card.photoName
            ? `/api/restaurants/photo?name=${encodeURIComponent(card.photoName)}&max=480`
            : "";
        if (!src) return null;
        // Places photo metadata does not prove that an image is a menu. Keep
        // the scan label factual even when the legacy media bucket says menu.
        const label = card.photoName && card.kind === "menu" ? "Listing photo" : card.label;
        const sourceLabel = card.photoName
          ? [card.subtitle, "Google listing photo"].filter(Boolean).join(" · ")
          : card.subtitle || "Restaurant listing media";
        return {
          src,
          label,
          sourceLabel,
          authorAttributions: card.authorAttributions,
          googleMapsUri: card.googleMapsUri,
          flagContentUri: card.flagContentUri,
          sourceUrl: card.photoName
            ? card.googleMapsUri || card.href || listingMapsUrl
            : card.href,
          alt: `${previewName} listing photo${label ? ` — ${label}` : ""}`,
        };
      })
      .filter((photo): photo is ScanPhoto => photo !== null);

    const heroPhoto =
      scanPhotos[0]?.src ||
      (data?.place.media?.photosAndVideos?.[0]?.photoName
        ? `/api/restaurants/photo?name=${encodeURIComponent(data.place.media.photosAndVideos[0].photoName)}&max=480`
        : undefined);

    return (
      <ScanExperience
        restaurantName={previewName}
        address={data?.place.address || data?.report.address || addressFromQuery || undefined}
        rating={data?.place.rating}
        category={category}
        website={data?.place.website}
        placeId={placeId}
        mapsUri={data?.place.mapsUri || data?.place.media?.mapsUri}
        placeAttributions={data?.place.attributions}
        latitude={data?.place.latitude ?? latPreview}
        longitude={data?.place.longitude ?? lngPreview}
        photoUrl={heroPhoto}
        photos={scanPhotos}
        reviews={data?.report.recentReviews || []}
        desktopScreenshot={data?.report.websiteScreenshot || undefined}
        mobileScreenshot={data?.report.websiteMobileScreenshot || undefined}
        websiteUrl={data?.place.website}
        fetchComplete={fetchComplete}
        onReady={handleScanReady}
      />
    );
  }

  if (error || !data) {
    return (
      <div className="mx-auto max-w-3xl px-6 py-24 text-center">
        <p className="font-display text-2xl font-semibold text-ink">Report unavailable</p>
        <p className="mt-2 text-muted">{error || "Restaurant not found."}</p>
        <div className="mt-6 flex flex-wrap justify-center gap-3">
          {canRetry ? (
            <button
              type="button"
              onClick={retryReport}
              className="min-h-11 cursor-pointer rounded-full bg-primary px-5 font-semibold text-white hover:bg-primary-dim focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
            >
              Try the live scan again
            </button>
          ) : null}
          <Link
            href="/"
            className="inline-flex min-h-11 items-center rounded-full border border-primary px-5 font-semibold text-primary underline underline-offset-4"
          >
            Search again
          </Link>
        </div>
      </div>
    );
  }

  const { place, report } = data;
  const summaryLines = (report.aiSummary || "")
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);
  // Unlock state is authoritative only when the server validates the HttpOnly
  // report cookie and returns an explicitly unlocked payload.
  const unlocked = report.fullReportLocked === false;
  const aiAssisted = report.analysisSource === "ai-assisted";
  const partial = report.analysisStatus === "partial";
  const analysisLabel = aiAssisted ? "AI-assisted" : "Automated signals";
  const summaryTitle = aiAssisted ? "AI-assisted summary" : "Signal summary";
  const mapSrc = reportMapSrc(place);
  const listingCards = [
    ...(place.media?.photosAndVideos || []),
    ...(place.media?.menuAndHighlights || []),
  ].filter((card) => Boolean(listingPhotoSrc(card)));
  const heroPhoto = listingCards[0];
  const generatedSeconds =
    typeof report.generatedInMs === "number" && report.generatedInMs >= 0
      ? `${Math.max(0.1, report.generatedInMs / 1000).toFixed(1)}s`
      : "Ready";
  const openUnlock = () => setUnlockOpen(true);
  const pdfHref = `/api/restaurants/${encodeURIComponent(place.placeId)}/report.pdf`;

  return (
    <div className="hero-atmosphere min-h-[70vh] px-4 py-4 sm:px-8 sm:py-8 md:px-10 md:py-12">
      <div className="report-reveal mx-auto w-full max-w-6xl space-y-5 md:space-y-6">
        {/* Identity, exact map, real listing imagery and scan status stay above fold on mobile. */}
        <section className="overflow-hidden rounded-[24px] border border-border bg-bg/95 shadow-[0_14px_46px_rgba(15,39,31,0.08)]">
          <header className="px-5 pb-4 pt-5 sm:px-7 sm:pb-5 sm:pt-6">
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded-full bg-[#e8f1eb] px-2.5 py-1 text-[10px] font-bold uppercase tracking-[0.09em] text-primary">
                {analysisLabel}
              </span>
              <span
                className={`rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-[0.09em] ${
                  partial ? "bg-[#fff2dc] text-[#8a5200]" : "bg-[#e8f6ee] text-[#176b3a]"
                }`}
              >
                {partial ? "Partial snapshot" : "Scan complete"}
              </span>
            </div>
            <h1 className="mt-3 font-display text-[clamp(1.65rem,7vw,2.75rem)] font-semibold leading-[1.05] tracking-[-0.035em] text-ink">
              {report.restaurantName}
            </h1>
            {report.address ? (
              <p className="mt-1.5 line-clamp-2 text-[13px] leading-snug text-muted sm:text-[15px]">
                {report.address}
              </p>
            ) : null}
            <PlaceDataAttribution
              attributions={place.attributions}
              mapsUri={place.mapsUri}
              showGoogleMaps={place.source === "places"}
            />
          </header>

          <div className="px-3 pb-4 sm:px-5 sm:pb-5">
            <div className="grid h-[178px] grid-cols-[minmax(0,1.2fr)_minmax(96px,0.8fr)] gap-2 sm:h-[250px] sm:gap-3">
              <div className="relative overflow-hidden rounded-2xl bg-[#e8e4dc]">
                {mapSrc ? (
                  <iframe
                    title={`Map of ${report.restaurantName}`}
                    src={mapSrc}
                    className="absolute inset-0 h-full w-full border-0"
                    loading="eager"
                    referrerPolicy="no-referrer-when-downgrade"
                  />
                ) : (
                  <div className="flex h-full items-center justify-center px-4 text-center text-[12px] font-medium text-muted">
                    Map location unavailable
                  </div>
                )}
                <span
                  className="absolute bottom-2 left-2 rounded-full bg-white/95 px-2.5 py-1 text-[12px] font-normal text-[#5e5e5e] shadow-sm"
                  translate="no"
                >
                  Google Maps
                </span>
              </div>
              <ListingPhoto card={heroPhoto} restaurantName={report.restaurantName} />
            </div>

            <div className="mt-3 grid grid-cols-3 divide-x divide-border rounded-2xl border border-border bg-[#f7f4ef] px-2 py-3 text-center">
              <div className="px-1.5">
                <p className="font-display text-[1.55rem] font-semibold leading-none text-ink">{report.overallScore}</p>
                <p className="mt-1 text-[10px] font-semibold uppercase tracking-[0.06em] text-muted">Score</p>
              </div>
              <div className="px-1.5">
                <p className="truncate text-[13px] font-bold" style={{ color: report.overallColor }}>
                  {report.overallLabel}
                </p>
                <p className="mt-1 text-[10px] font-semibold uppercase tracking-[0.06em] text-muted">Visibility</p>
              </div>
              <div className="px-1.5">
                <p className="text-[13px] font-bold tabular-nums text-ink">{generatedSeconds}</p>
                <p className="mt-1 text-[10px] font-semibold uppercase tracking-[0.06em] text-muted">Scan time</p>
              </div>
            </div>

            {report.analysisNotice ? (
              <p className="mt-3 text-[12px] leading-relaxed text-muted">{report.analysisNotice}</p>
            ) : null}

            {(place.website || place.mapsUri) && (
              <div className="mt-2 flex flex-wrap gap-2 text-[13px]">
                {place.website ? (
                  <a
                    href={place.website}
                    className="inline-flex min-h-11 items-center rounded-full px-3 font-semibold text-primary underline decoration-primary/35 underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
                    target="_blank"
                    rel="noreferrer"
                  >
                    Visit website
                  </a>
                ) : null}
                {place.mapsUri ? (
                  <a
                    href={place.mapsUri}
                    className="inline-flex min-h-11 items-center rounded-full px-3 font-semibold text-primary underline decoration-primary/35 underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
                    target="_blank"
                    rel="noreferrer"
                  >
                    Open Google Maps
                  </a>
                ) : null}
              </div>
            )}
          </div>
        </section>

        <div className="grid gap-5 lg:grid-cols-[minmax(0,1.05fr)_minmax(300px,0.95fr)] lg:items-start lg:gap-6">
          {/* Left column — narrative sections */}
          <div className="space-y-5">
            {summaryLines.length > 0 || !unlocked ? (
              <ReportSection eyebrow="01 · Findings" title={summaryTitle}>
                <div className="space-y-3 text-[15px] leading-relaxed text-ink">
                  {summaryLines.length > 0 ? (
                    summaryLines.map((line) => <p key={line}>{line}</p>)
                  ) : (
                    <p>
                      Your headline score is ready. Confirm your email to reveal the evidence-backed observations and prioritized improvement plan.
                    </p>
                  )}
                </div>
                {!unlocked ? (
                  <button
                    type="button"
                    onClick={openUnlock}
                    className="mt-5 min-h-11 w-full cursor-pointer rounded-xl bg-[#f7f4ef] px-4 py-3 text-left text-[13px] font-semibold text-primary ring-1 ring-primary/10 transition-colors hover:bg-[#efe9e1] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
                  >
                    Confirm email for the full venue pulse →
                  </button>
                ) : (
                  <a
                    href={pdfHref}
                    className="mt-5 flex min-h-11 items-center justify-center rounded-xl bg-[#eef6f1] px-4 py-3 text-[13px] font-semibold text-accent ring-1 ring-accent/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
                  >
                    Full report unlocked · Download the 2-page PDF
                  </a>
                )}
              </ReportSection>
            ) : null}

            {report.websiteScreenshot || report.websiteReview ? (
              <ReportSection
                eyebrow="02 · Website"
                title={aiAssisted ? "Screenshot & AI-assisted design review" : "Screenshot & design signals"}
                bodyClassName="p-0"
              >
                <div className="border-b border-border px-5 py-4 sm:px-6">
                  {typeof report.websiteQualityScore === "number" && report.websiteQualityScore > 0 ? (
                    <p className="text-[15px] text-ink">
                      Visual quality <strong>{report.websiteQualityScore}</strong>{" "}
                      <span className="text-muted">
                        ({aiAssisted ? "AI-assisted design score" : "automated design estimate"})
                      </span>
                    </p>
                  ) : (
                    <p className="text-[14px] text-muted">Live homepage capture + design notes.</p>
                  )}
                </div>
                <LockedBlur locked={!unlocked} label="Confirm email for the full venue pulse" className="min-h-[100px]" onUnlock={openUnlock}>
                  <div className="px-5 py-4 sm:px-6">
                    {report.websiteReview ? (
                      <p className="text-[14px] leading-relaxed text-muted">{report.websiteReview}</p>
                    ) : (
                      <p className="text-[14px] leading-relaxed text-muted">
                        Detailed design notes unlock after email verification.
                      </p>
                    )}
                  </div>
                </LockedBlur>
                {report.websiteScreenshot || report.websiteMobileScreenshot ? (
                  <div className="grid gap-3 border-t border-border bg-[#f7f4ef] p-4 sm:grid-cols-[minmax(0,1fr)_180px] sm:p-5">
                    {report.websiteScreenshot ? (
                      <figure className="overflow-hidden rounded-xl border border-border bg-white">
                        <figcaption className="border-b border-border px-3 py-2 text-[11px] font-semibold uppercase tracking-[0.06em] text-muted">Desktop capture</figcaption>
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img src={report.websiteScreenshot} alt={`${report.restaurantName} website desktop capture`} className="block aspect-[16/9] w-full object-cover object-top" />
                      </figure>
                    ) : null}
                    {report.websiteMobileScreenshot ? (
                      <figure className="overflow-hidden rounded-xl border border-border bg-white">
                        <figcaption className="border-b border-border px-3 py-2 text-[11px] font-semibold uppercase tracking-[0.06em] text-muted">Mobile capture</figcaption>
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img src={report.websiteMobileScreenshot} alt={`${report.restaurantName} website mobile capture`} className="block aspect-[390/520] w-full object-cover object-top" />
                      </figure>
                    ) : null}
                  </div>
                ) : null}
              </ReportSection>
            ) : null}

            <ReportSection eyebrow="03 · Evidence" title="Menu & social presence">
              <LivePresenceEvidence
                menu={report.menuEvidence}
                social={report.socialPresence}
                locked={!unlocked}
                onUnlock={openUnlock}
              />
            </ReportSection>
          </div>

          {/* Right column — scorecard stack */}
          <aside className="report-phone-stage space-y-4 lg:sticky lg:top-24">
            <ReportSection eyebrow="04 · Scorecard" bodyClassName="bg-[#f4f0ea] p-3 sm:p-3.5">
              <LiveHealthCard
                restaurantName={report.restaurantName}
                overallScore={report.overallScore}
                overallLabel={report.overallLabel}
                overallColor={report.overallColor}
                metrics={report.metrics}
                locked={!unlocked}
                onUnlock={openUnlock}
                pdfHref={unlocked ? pdfHref : undefined}
              />
            </ReportSection>

            <ReportSection eyebrow="05 · Market" bodyClassName="bg-[#f4f0ea] p-3 sm:p-3.5">
              <LiveCompetitorsCard
                rows={report.competitors}
                scan={report.competitorScan}
                locked={!unlocked}
                onUnlock={openUnlock}
              />
            </ReportSection>

            <ReportSection eyebrow="06 · Fixes" bodyClassName="bg-[#f4f0ea] p-3 sm:p-3.5">
              <LiveIssuesCard
                issues={report.issues}
                locked={!unlocked}
                onFix={openUnlock}
              />
            </ReportSection>
          </aside>
        </div>

        <LiveListingMedia media={place.media} />
      </div>
      <UnlockReportDialog
        open={unlockOpen}
        placeId={place.placeId}
        restaurantName={report.restaurantName}
        onClose={() => setUnlockOpen(false)}
        onUnlocked={(payload) => {
          setData(payload);
        }}
      />
    </div>
  );
}
