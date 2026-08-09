"use client";

import { FormEvent, useCallback, useEffect, useId, useRef, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import type { MediaCard, RestaurantDetails } from "@/lib/places";
import type { RestaurantReport } from "@/lib/report";
import { parsePreviewCoordinates } from "@/lib/report-preview";
import LiveCompetitorsCard from "@/components/report/LiveCompetitorsCard";
import LiveHealthCard from "@/components/report/LiveHealthCard";
import LiveIssuesCard from "@/components/report/LiveIssuesCard";
import LiveListingMedia from "@/components/report/LiveListingMedia";
import LockedBlur from "@/components/report/LockedBlur";
import ReportSection from "@/components/report/ReportSection";
import ScanExperience, { type ScanPhoto } from "@/components/report/ScanExperience";

type Payload = {
  place: RestaurantDetails;
  report: RestaurantReport;
};

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

function ListingPhoto({ card, restaurantName }: { card?: MediaCard; restaurantName: string }) {
  const src = listingPhotoSrc(card);
  const sourceLabel = card?.subtitle || (card?.photoName ? "From Google listing" : "");
  const tile = (
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
          {card?.label ? <p className="truncate text-[11px] font-semibold">{card.label}</p> : null}
          {sourceLabel ? <p className="mt-0.5 truncate text-[9px] font-medium text-white/80">{sourceLabel}</p> : null}
        </div>
      ) : null}
    </div>
  );

  if (card?.href) {
    return (
      <a
        href={card.href}
        target="_blank"
        rel="noreferrer"
        aria-label={`Open ${card.label || restaurantName} photo source`}
        className="block h-full min-h-0 rounded-2xl focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
      >
        {tile}
      </a>
    );
  }
  return tile;
}

export default function ReportClient({ placeId }: { placeId: string }) {
  const searchParams = useSearchParams();
  const unlockFromLink = (searchParams.get("unlock") || "").trim();
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
      }
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
        if (cancelled) return;
        if (timedOut) {
          setError("The live scan timed out before all signals finished. Please try again.");
          setPhase("ready");
          setFetchComplete(true);
          bootstrappingRef.current = false;
          return;
        }
        if (controller.signal.aborted) return;
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "Failed to load report");
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
          sourceUrl: card.photoName
            ? card.href || listingMapsUrl
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
  const scrollToLead = () => {
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    leadRef.current?.scrollIntoView({ behavior: reducedMotion ? "auto" : "smooth", block: "center" });
  };

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
                <span className="absolute bottom-2 left-2 rounded-full bg-white/95 px-2.5 py-1 text-[10px] font-semibold text-ink shadow-sm">
                  Google listing
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
            {summaryLines.length > 0 ? (
              <ReportSection eyebrow="01 · Findings" title={summaryTitle}>
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

            <ReportSection
              eyebrow="02 · Access"
              title={unlocked ? "Report unlocked" : "Verify email"}
            >
              <div ref={leadRef}>
                {unlocked ? (
                  <p className="text-[15px] leading-relaxed text-ink">
                    You&apos;re on the list — we saved this restaurant as an interested lead. Full competitor
                    rankings and fix suggestions are open on the right.
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
                        className="min-w-0 flex-1 rounded-xl border border-border bg-transparent px-4 py-3 text-[15px] tracking-[0.3em] text-ink outline-none placeholder:text-secondary focus:border-primary focus-visible:ring-2 focus-visible:ring-primary/30"
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
                      Free forever for this restaurant. Unlock full competitor rankings and every fix
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
                        className="min-w-0 flex-1 rounded-xl border border-border bg-transparent px-4 py-3 text-[15px] text-ink outline-none placeholder:text-secondary focus:border-primary focus-visible:ring-2 focus-visible:ring-primary/30"
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

            {report.websiteScreenshot || report.websiteReview ? (
              <ReportSection
                eyebrow="03 · Website"
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
                <LockedBlur
                  locked={!unlocked}
                  label="Verify email to unlock the full website review"
                  className="min-h-[120px]"
                >
                  <div className="px-5 py-4 sm:px-6">
                    {report.websiteReview ? (
                      <p className="text-[14px] leading-relaxed text-muted">{report.websiteReview}</p>
                    ) : (
                      <p className="text-[14px] leading-relaxed text-muted">
                        Detailed design notes unlock after email verification.
                      </p>
                    )}
                  </div>
                  {report.websiteScreenshot ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={report.websiteScreenshot}
                      alt={`${report.restaurantName} website homepage screenshot`}
                      className="block w-full bg-[#efebe6] object-cover object-top"
                    />
                  ) : null}
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
                locked={!unlocked}
              />
            </ReportSection>

            <ReportSection eyebrow="05 · Market" bodyClassName="bg-[#f4f0ea] p-3 sm:p-3.5">
              <LiveCompetitorsCard
                rows={report.competitors}
                locked={!unlocked}
                onUnlock={scrollToLead}
              />
            </ReportSection>

            <ReportSection eyebrow="06 · Fixes" bodyClassName="bg-[#f4f0ea] p-3 sm:p-3.5">
              <LiveIssuesCard
                issues={report.issues}
                estimatedMonthlyLoss={report.estimatedMonthlyLoss}
                locked={!unlocked}
                onFix={scrollToLead}
              />
            </ReportSection>
          </aside>
        </div>

        <LiveListingMedia media={place.media} />
      </div>
    </div>
  );
}
