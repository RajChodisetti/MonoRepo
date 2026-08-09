"use client";

import { FormEvent, useCallback, useEffect, useId, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import type { MediaCard, RestaurantDetails } from "@/lib/places";
import type { RestaurantReport } from "@/lib/report";
import { parsePreviewCoordinates } from "@/lib/report-preview";
import { normalizeScanPhotos, reportMapEmbedUrl, websiteCaptureEvidence } from "@/lib/report-scan";
import LiveCompetitorsCard from "@/components/report/LiveCompetitorsCard";
import LiveHealthCard from "@/components/report/LiveHealthCard";
import LiveIssuesCard from "@/components/report/LiveIssuesCard";
import LiveListingMedia from "@/components/report/LiveListingMedia";
import LockedBlur from "@/components/report/LockedBlur";
import ReportSection from "@/components/report/ReportSection";
import ScanExperience, { ReviewScroller } from "@/components/report/ScanExperience";

type Payload = {
  place: RestaurantDetails;
  report: RestaurantReport;
};

function mediaImageSrc(card: MediaCard): string | null {
  const direct = card.imageUrl?.trim();
  if (direct) return direct;
  const photoName = card.photoName?.trim();
  return photoName
    ? `/api/restaurants/photo?name=${encodeURIComponent(photoName)}&max=960`
    : null;
}

function ReportOverviewEvidence({ place }: { place: RestaurantDetails }) {
  const listingPhotos = useMemo(
    () => normalizeScanPhotos(
      undefined,
      place.name,
      [
        ...(place.media?.menuAndHighlights || []),
        ...(place.media?.photosAndVideos || []),
      ].flatMap((card) => {
        const src = mediaImageSrc(card);
        return src ? [{ src, label: card.label }] : [];
      }),
    ),
    [place.media, place.name],
  );
  const [readyPhoto, setReadyPhoto] = useState("");
  const [failedPhotos, setFailedPhotos] = useState<Set<string>>(() => new Set());
  const mapSrc = reportMapEmbedUrl({
    restaurantName: place.name,
    address: place.address,
    placeId: place.placeId,
    latitude: place.latitude,
    longitude: place.longitude,
  });
  const hasExactCoordinates =
    typeof place.latitude === "number" &&
    typeof place.longitude === "number" &&
    Number.isFinite(place.latitude) &&
    Number.isFinite(place.longitude) &&
    Math.abs(place.latitude) <= 90 &&
    Math.abs(place.longitude) <= 180;

  useEffect(() => {
    let cancelled = false;
    const images = listingPhotos.map((photo) => {
      const image = new window.Image();
      image.onload = () => {
        if (!cancelled) setReadyPhoto((current) => current || photo.src);
      };
      image.onerror = () => {
        if (!cancelled) setFailedPhotos((current) => new Set(current).add(photo.src));
      };
      image.src = photo.src;
      return image;
    });
    return () => {
      cancelled = true;
      for (const image of images) {
        image.onload = null;
        image.onerror = null;
      }
    };
  }, [listingPhotos]);

  const showPhoto = Boolean(readyPhoto && !failedPhotos.has(readyPhoto));
  if (!mapSrc && !showPhoto) return null;

  return (
    <div className={`mt-5 grid gap-3 border-t border-border pt-5 ${showPhoto && mapSrc ? "sm:grid-cols-2" : ""}`}>
      {mapSrc ? (
        <figure>
          <div className="relative h-[190px] overflow-hidden rounded-2xl border border-border bg-[#e8e4dc] sm:h-[220px]">
            <iframe
              title={hasExactCoordinates ? `Exact map location of ${place.name}` : `Google listing map for ${place.name}`}
              src={mapSrc}
              className="absolute inset-0 h-full w-full border-0"
              loading="lazy"
              referrerPolicy="no-referrer-when-downgrade"
            />
          </div>
          <figcaption className="mt-1.5 text-[11px] font-medium text-muted">
            {hasExactCoordinates ? "Exact Google listing location" : "Google listing map"}
          </figcaption>
        </figure>
      ) : null}
      {showPhoto ? (
        <figure>
          <div className="h-[190px] overflow-hidden rounded-2xl border border-border bg-[#efebe6] sm:h-[220px]">
            {/* Dynamic Google media cannot use next/image's fixed remote allow-list. */}
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={readyPhoto}
              alt={`${place.name} listing photo`}
              className="h-full w-full object-cover"
              onError={() => {
                setFailedPhotos((current) => new Set(current).add(readyPhoto));
                setReadyPhoto("");
              }}
            />
          </div>
          <figcaption className="mt-1.5 text-[11px] font-medium text-muted">Live listing photo</figcaption>
        </figure>
      ) : null}
    </div>
  );
}

function WebsiteReviewCaptures({
  restaurantName,
  desktopSrc,
  mobileSrc,
}: {
  restaurantName: string;
  desktopSrc?: string;
  mobileSrc?: string;
}) {
  const captures = websiteCaptureEvidence(desktopSrc, mobileSrc);
  const [readySources, setReadySources] = useState<Set<string>>(() => new Set());
  const [failedSources, setFailedSources] = useState<Set<string>>(() => new Set());

  useEffect(() => {
    let cancelled = false;
    const pending = websiteCaptureEvidence(desktopSrc, mobileSrc);
    const images = pending.map((capture) => {
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
      for (const image of images) {
        image.onload = null;
        image.onerror = null;
      }
    };
  }, [desktopSrc, mobileSrc]);

  const available = captures.filter(
    (capture) => readySources.has(capture.src) && !failedSources.has(capture.src),
  );
  if (available.length === 0) return null;

  return (
    <div className={`grid items-start gap-4 px-5 pb-5 sm:px-6 ${available.length === 2 ? "sm:grid-cols-[minmax(0,1fr)_12rem]" : ""}`}>
      {available.map((capture) => (
        <figure key={capture.kind} className={capture.kind === "mobile" ? "mx-auto w-full max-w-[12rem]" : "min-w-0"}>
          <div
            className={
              capture.kind === "desktop"
                ? "overflow-hidden rounded-xl border border-border bg-[#efebe6]"
                : "overflow-hidden rounded-[1.5rem] border-[5px] border-[#1a1a1a] bg-[#1a1a1a]"
            }
          >
            {capture.kind === "desktop" ? (
              <div className="flex h-6 items-center gap-1.5 border-b border-border bg-white px-2.5" aria-hidden="true">
                <span className="h-1.5 w-1.5 rounded-full bg-[#ee6a5f]" />
                <span className="h-1.5 w-1.5 rounded-full bg-[#f3bd4f]" />
                <span className="h-1.5 w-1.5 rounded-full bg-[#61c454]" />
              </div>
            ) : null}
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={capture.src}
              alt={`${restaurantName} ${capture.kind} website capture`}
              className={
                capture.kind === "desktop"
                  ? "block max-h-[460px] w-full object-cover object-top"
                  : "block aspect-[9/18] w-full object-cover object-top"
              }
              onError={() => setFailedSources((current) => new Set(current).add(capture.src))}
            />
          </div>
          <figcaption className="mt-2 text-center text-[12px] font-semibold capitalize text-muted">
            {capture.kind} view
          </figcaption>
        </figure>
      ))}
    </div>
  );
}

export default function ReportClient({ placeId }: { placeId: string }) {
  return <ReportClientContent key={placeId} placeId={placeId} />;
}

function ReportClientContent({ placeId }: { placeId: string }) {
  const searchParams = useSearchParams();
  const unlockFromLink = (searchParams.get("unlock") || "").trim();
  const nameFromQuery = (searchParams.get("name") || "").trim();
  const addressFromQuery = (searchParams.get("address") || "").trim();
  const { latitude: latPreview, longitude: lngPreview } = parsePreviewCoordinates(
    searchParams.get("lat"),
    searchParams.get("lng"),
  );

  const [data, setData] = useState<Payload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [phase, setPhase] = useState<"scan" | "ready">("scan");
  const [fetchComplete, setFetchComplete] = useState(false);
  const bootstrappingRef = useRef(true);
  const [email, setEmail] = useState("");
  const [otp, setOtp] = useState("");
  const [unlockToken, setUnlockToken] = useState(unlockFromLink);
  const [step, setStep] = useState<"email" | "otp" | "unlocked">(
    unlockFromLink ? "unlocked" : "email",
  );
  const [leadStatus, setLeadStatus] = useState<"idle" | "sending" | "ok" | "err">("idle");
  const [leadMessage, setLeadMessage] = useState("");
  const [devOtp, setDevOtp] = useState("");
  const leadRef = useRef<HTMLDivElement>(null);
  const emailId = useId();
  const otpId = useId();

  useEffect(() => {
    let cancelled = false;
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), 18_000);
    async function load() {
      const quiet = !bootstrappingRef.current;
      try {
        const qs = unlockToken
          ? `?unlock=${encodeURIComponent(unlockToken)}`
          : "";
        const res = await fetch(`/api/restaurants/${encodeURIComponent(placeId)}${qs}`, {
          signal: controller.signal,
        });
        const json = await res.json();
        if (!res.ok) throw new Error(json.error || "Failed to load report");
        if (!cancelled) {
          const payload = json as Payload;
          if (!payload?.report || !payload?.place) {
            throw new Error("Invalid report payload from server");
          }
          setData(payload);
          if (payload.report.fullReportLocked === false) {
            setStep("unlocked");
          }
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
        if (!cancelled) {
          const message =
            e instanceof DOMException && e.name === "AbortError"
              ? "The live report timed out after 18 seconds. Please try again."
              : e instanceof Error
                ? e.message
                : "Failed to load report";
          setError(message);
          setPhase("ready");
          bootstrappingRef.current = false;
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
  }, [placeId, unlockToken]);

  const handleScanReady = useCallback(() => {
    setPhase("ready");
  }, []);

  // Hard escape hatch — only after min scan window + buffer
  useEffect(() => {
    if (phase !== "scan") return;
    const t = window.setTimeout(() => setPhase("ready"), 90_000);
    return () => window.clearTimeout(t);
  }, [phase]);

  async function requestOtp(event: FormEvent) {
    event.preventDefault();
    if (!data) return;
    setLeadStatus("sending");
    setLeadMessage("");
    setDevOtp("");
    try {
      const res = await fetch("/api/leads", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email,
          placeId: data.place.placeId,
          restaurantName: data.place.name,
        }),
      });
      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.error?.message || json.error || json.message || "Could not send code");
      }
      setLeadStatus("ok");
      setLeadMessage(json.message || "Check your inbox for the verification code.");
      if (json.devOtp) setDevOtp(String(json.devOtp));
      setStep("otp");
    } catch (e) {
      setLeadStatus("err");
      setLeadMessage(e instanceof Error ? e.message : "Something went wrong");
    }
  }

  async function verifyOtp(event: FormEvent) {
    event.preventDefault();
    if (!data) return;
    setLeadStatus("sending");
    setLeadMessage("");
    try {
      const res = await fetch("/api/leads/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email,
          placeId: data.place.placeId,
          otp,
        }),
      });
      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.error?.message || json.error || "Invalid verification code");
      }
      setLeadStatus("ok");
      setLeadMessage("Email verified. Full report unlocked.");
      if (json.unlockToken) setUnlockToken(String(json.unlockToken));
      if (json.place && json.report) {
        setData({ place: json.place, report: json.report });
      }
      setStep("unlocked");
    } catch (e) {
      setLeadStatus("err");
      setLeadMessage(e instanceof Error ? e.message : "Something went wrong");
    }
  }

  if (phase === "scan" && !error) {
    const previewName = data?.place.name || nameFromQuery || "Your restaurant";
    const categoryRaw = data?.place.types?.[0]?.replace(/_/g, " ") || "Restaurant";
    const category = categoryRaw.replace(/\b\w/g, (c) => c.toUpperCase());

    const photoCards = [
      ...(data?.place.media?.menuAndHighlights || []),
      ...(data?.place.media?.photosAndVideos || []),
    ];
    const scanPhotos = photoCards
      .map((card) => {
        const src = card.imageUrl
          ? card.imageUrl
          : card.photoName
            ? `/api/restaurants/photo?name=${encodeURIComponent(card.photoName)}&max=480`
            : "";
        return src ? { src, label: card.label } : null;
      })
      .filter((p): p is { src: string; label: string } => Boolean(p));

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
        latitude={data?.place.latitude ?? latPreview}
        longitude={data?.place.longitude ?? lngPreview}
        photoUrl={heroPhoto}
        photos={scanPhotos}
        reviews={data?.report.recentReviews || []}
        desktopScreenshot={data?.report.websiteScreenshot || undefined}
        mobileScreenshot={data?.report.websiteMobileScreenshot || undefined}
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
        <Link href="/" className="mt-6 inline-block font-semibold text-primary underline">
          Search again
        </Link>
      </div>
    );
  }

  const { place, report } = data;
  const summaryLines = (report.aiSummary || "")
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);
  const unlocked = report.fullReportLocked === false || step === "unlocked";
  const aiAssisted = report.analysisSource === "ai-assisted";
  const partial = report.analysisStatus === "partial";
  const competitorRows = report.competitors.filter(
    (row) => !/^Nearby competitor\s+[A-Z]$/i.test(row.name.trim()),
  );

  return (
    <div className="hero-atmosphere min-h-[70vh] px-5 py-10 sm:px-8 md:px-10 md:py-14">
      <div className="report-reveal mx-auto w-full max-w-6xl space-y-5 md:space-y-6">
        {/* 1 — Overview */}
        <ReportSection
          eyebrow={aiAssisted ? "AI-assisted restaurant report" : "Restaurant digital-footprint report"}
          bodyClassName="px-5 py-6 sm:px-7 sm:py-7"
        >
          <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
            <div className="min-w-0 max-w-2xl">
              <h1 className="font-display text-[clamp(1.75rem,3.8vw,2.75rem)] font-semibold leading-[1.12] tracking-[-0.03em] text-ink">
                {report.restaurantName}
              </h1>
              {report.address ? (
                <p className="mt-2 text-[14px] leading-relaxed text-muted sm:text-[15px]">{report.address}</p>
              ) : null}
              <p className="mt-4 text-[15px] leading-relaxed text-muted sm:text-[16px]">
                We scored the available SEO, review, website, order-online, menu, contact, and listing
                evidence with the weights shown in the scorecard.
              </p>
              {partial ? (
                <p className="mt-3 rounded-xl border border-[#e8d7b6] bg-[#fff8e9] px-3.5 py-2.5 text-[12px] leading-relaxed text-[#7a5518]">
                  {report.analysisNotice || "Some live signals were unavailable, so this report is partial."}
                </p>
              ) : null}
            </div>

            <div className="shrink-0 rounded-2xl border border-border bg-[#f7f4ef] px-5 py-4 sm:min-w-[168px]">
              <p className="text-[11px] font-semibold uppercase tracking-[0.07em] text-muted">
                {partial ? "Verified points" : "Overall score"}
              </p>
              <p className="mt-1 flex items-baseline gap-2">
                <span className="font-display text-[2.4rem] font-semibold leading-none tracking-[-0.04em] text-ink">
                  {report.overallScore}<span className="ml-1 text-[1rem] text-muted">/100</span>
                </span>
                <span className="text-[13px] font-semibold" style={{ color: report.overallColor }}>
                  {report.overallLabel}
                </span>
              </p>
              <p className="mt-2 text-[12px] text-muted">
                {partial
                  ? `${report.metrics.filter((metric) => !metric.status.toLowerCase().includes("assessed")).reduce((total, metric) => total + (metric.max || 0), 0)}/100 fully assessed`
                  : aiAssisted
                    ? "AI-assisted weighted review"
                    : "Automated weighted review"}
              </p>
            </div>
          </div>

          {(place.website || place.mapsUri) && (
            <div className="mt-5 flex flex-wrap gap-x-5 gap-y-2 border-t border-border pt-4 text-[13px]">
              {place.website ? (
                <a
                  href={place.website}
                  className="font-medium text-primary underline"
                  target="_blank"
                  rel="noreferrer"
                >
                  Website
                </a>
              ) : null}
              {place.mapsUri ? (
                <a
                  href={place.mapsUri}
                  className="font-medium text-primary underline"
                  target="_blank"
                  rel="noreferrer"
                >
                  Google Maps
                </a>
              ) : null}
            </div>
          )}
          <ReportOverviewEvidence place={place} />
        </ReportSection>

        <div className="grid gap-5 lg:grid-cols-[minmax(0,1.05fr)_minmax(300px,0.95fr)] lg:items-start lg:gap-6">
          {/* Left column — narrative sections */}
          <div className="space-y-5">
            {summaryLines.length > 0 ? (
              <ReportSection eyebrow="01 · Findings" title={aiAssisted ? "AI-assisted summary" : "Automated summary"}>
                <div className="space-y-3 text-[15px] leading-relaxed text-ink">
                  {summaryLines.map((line) => (
                    <p key={line}>{line}</p>
                  ))}
                </div>
                {!unlocked ? (
                  <p className="mt-5 rounded-xl bg-[#f7f4ef] px-4 py-3 text-[13px] font-medium text-primary">
                    {report.unlockCta || "Unlock the full SEO report by verifying your email."}
                  </p>
                ) : (
                  <p className="mt-5 rounded-xl bg-[#eef6f1] px-4 py-3 text-[13px] font-medium text-accent">
                    Full report unlocked.
                  </p>
                )}
              </ReportSection>
            ) : null}

            {report.recentReviews && report.recentReviews.length > 0 ? (
              <ReportSection
                eyebrow="Google review evidence"
                title={`${report.recentReviews.length} available Google reviews`}
              >
                <ReviewScroller reviews={report.recentReviews} placeRating={place.rating} />
              </ReportSection>
            ) : null}

            <ReportSection
              eyebrow="02 · Access"
              title={unlocked ? "Report unlocked" : "Confirm your email"}
            >
              <div ref={leadRef}>
                {unlocked ? (
                  <p className="text-[15px] leading-relaxed text-ink">
                    You&apos;re on the list — we saved this restaurant as an interested lead. Full scoring
                    details and fix suggestions are open on the right.
                  </p>
                ) : step === "otp" ? (
                  <>
                    <p className="text-[15px] font-semibold text-ink">Enter the verification code</p>
                    <p className="mt-1 text-[13px] text-muted">We sent a 6-digit code to {email}.</p>
                    {devOtp ? (
                      <p className="mt-3 rounded-lg bg-[#f3f1ed] px-3 py-2 text-[13px] text-ink">
                        Local dev code: <strong>{devOtp}</strong>
                      </p>
                    ) : null}
                    <form className="mt-4 flex flex-col gap-2 sm:flex-row" onSubmit={verifyOtp}>
                      <label htmlFor={otpId} className="sr-only">
                        Verification code
                      </label>
                      <input
                        id={otpId}
                        inputMode="numeric"
                        pattern="[0-9]*"
                        maxLength={6}
                        required
                        value={otp}
                        onChange={(e) => setOtp(e.target.value.replace(/\D/g, "").slice(0, 6))}
                        placeholder="123456"
                        className="min-w-0 flex-1 rounded-xl border border-border bg-transparent px-4 py-3 text-[15px] tracking-[0.3em] text-ink outline-none placeholder:text-secondary focus:border-primary"
                      />
                      <button
                        type="submit"
                        disabled={leadStatus === "sending" || otp.length < 4}
                        className="cursor-pointer rounded-xl bg-primary px-5 py-3 text-[15px] font-semibold text-bg hover:bg-primary-dim disabled:opacity-60"
                      >
                        {leadStatus === "sending" ? "Verifying…" : "Unlock report"}
                      </button>
                    </form>
                    <button
                      type="button"
                      className="mt-3 text-[13px] font-medium text-primary underline"
                      onClick={() => {
                        setStep("email");
                        setLeadStatus("idle");
                        setLeadMessage("");
                      }}
                    >
                      Use a different email
                    </button>
                  </>
                ) : (
                  <>
                    <p className="text-[15px] leading-relaxed text-muted">
                      Free forever for this restaurant. Unlock the full scoring details and every fix
                      suggestion.
                    </p>
                    <form className="mt-4 flex flex-col gap-2 sm:flex-row" onSubmit={requestOtp}>
                      <label htmlFor={emailId} className="sr-only">
                        Email
                      </label>
                      <input
                        id={emailId}
                        type="email"
                        required
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        placeholder="you@restaurant.com"
                        className="min-w-0 flex-1 rounded-xl border border-border bg-transparent px-4 py-3 text-[15px] text-ink outline-none placeholder:text-secondary focus:border-primary"
                      />
                      <button
                        type="submit"
                        disabled={leadStatus === "sending"}
                        className="cursor-pointer rounded-xl bg-primary px-5 py-3 text-[15px] font-semibold text-bg hover:bg-primary-dim disabled:opacity-60"
                      >
                        {leadStatus === "sending" ? "Sending…" : "Send code"}
                      </button>
                    </form>
                  </>
                )}
                {leadMessage ? (
                  <p
                    className={`mt-3 text-[13px] ${
                      leadStatus === "ok"
                        ? "text-accent"
                        : leadStatus === "err"
                          ? "text-[#b42318]"
                          : "text-muted"
                    }`}
                  >
                    {leadMessage}
                  </p>
                ) : null}
              </div>
            </ReportSection>

            {report.websiteScreenshot || report.websiteMobileScreenshot || report.websiteReview ? (
              <ReportSection
                eyebrow="03 · Website"
                title="Website captures & visual review"
                bodyClassName="p-0"
              >
                <div className="border-b border-border px-5 py-4 sm:px-6">
                  {report.websiteQualityAssessed && typeof report.websiteQualityScore === "number" ? (
                    <p className="text-[15px] text-ink">
                      Visual quality <strong>{report.websiteQualityScore}/100</strong>{" "}
                      <span className="text-muted">(visual-quality score)</span>
                    </p>
                  ) : (
                    <p className="text-[14px] text-muted">
                      {report.websiteScreenshot || report.websiteMobileScreenshot
                        ? "Captured homepage evidence and design notes."
                        : "Website review notes; no homepage capture was available."}
                    </p>
                  )}
                </div>
                <LockedBlur
                  locked={!unlocked}
                  label="Confirm email to unlock the full website review"
                  className="min-h-[120px]"
                >
                  <div className="px-5 py-4 sm:px-6">
                    {report.websiteReview ? (
                      <p className="text-[14px] leading-relaxed text-muted">{report.websiteReview}</p>
                    ) : (
                      <p className="text-[14px] leading-relaxed text-muted">
                        Detailed website notes unlock after email verification.
                      </p>
                    )}
                  </div>
                  <WebsiteReviewCaptures
                    restaurantName={report.restaurantName}
                    desktopSrc={report.websiteScreenshot}
                    mobileSrc={report.websiteMobileScreenshot}
                  />
                </LockedBlur>
              </ReportSection>
            ) : null}
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
                partial={partial}
                locked={!unlocked}
              />
            </ReportSection>

            {competitorRows.length > 0 ? (
              <ReportSection
                eyebrow={competitorRows.some((row) => !row.highlight) ? "05 · Market" : "05 · Listing"}
                bodyClassName="bg-[#f4f0ea] p-3 sm:p-3.5"
              >
                <LiveCompetitorsCard
                  rows={competitorRows}
                  locked={!unlocked}
                  onUnlock={() => leadRef.current?.scrollIntoView({ behavior: "smooth", block: "center" })}
                />
              </ReportSection>
            ) : null}

            {report.issues.length > 0 ? (
              <ReportSection eyebrow="06 · Fixes" bodyClassName="bg-[#f4f0ea] p-3 sm:p-3.5">
                <LiveIssuesCard
                  issues={report.issues}
                  estimatedMonthlyLoss={report.estimatedMonthlyLoss}
                  locked={!unlocked}
                  onFix={() => leadRef.current?.scrollIntoView({ behavior: "smooth", block: "center" })}
                />
              </ReportSection>
            ) : null}
          </aside>
        </div>

        <LiveListingMedia media={place.media} />
      </div>
    </div>
  );
}
