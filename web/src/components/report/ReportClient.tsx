"use client";

import { FormEvent, useCallback, useEffect, useId, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import type { RestaurantDetails } from "@/lib/places";
import type { RestaurantReport } from "@/lib/report";
import LiveCompetitorsCard from "@/components/report/LiveCompetitorsCard";
import LiveHealthCard from "@/components/report/LiveHealthCard";
import LiveIssuesCard from "@/components/report/LiveIssuesCard";
import LiveListingMedia from "@/components/report/LiveListingMedia";
import LockedBlur from "@/components/report/LockedBlur";
import ReportSection from "@/components/report/ReportSection";
import ScanExperience from "@/components/report/ScanExperience";

type Payload = {
  place: RestaurantDetails;
  report: RestaurantReport;
};

export default function ReportClient({ placeId }: { placeId: string }) {
  const searchParams = useSearchParams();
  const unlockFromLink = (searchParams.get("unlock") || "").trim();
  const nameFromQuery = (searchParams.get("name") || "").trim();
  const addressFromQuery = (searchParams.get("address") || "").trim();
  const latFromQuery = Number(searchParams.get("lat") || "");
  const lngFromQuery = Number(searchParams.get("lng") || "");
  const latPreview = Number.isFinite(latFromQuery) ? latFromQuery : undefined;
  const lngPreview = Number.isFinite(lngFromQuery) ? lngFromQuery : undefined;

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
    bootstrappingRef.current = true;
    setPhase("scan");
    setFetchComplete(false);
    setData(null);
    setError(null);
  }, [placeId]);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      const quiet = !bootstrappingRef.current;
      if (!quiet) {
        setPhase("scan");
        setFetchComplete(false);
        setError(null);
      }
      try {
        const qs = unlockToken
          ? `?unlock=${encodeURIComponent(unlockToken)}`
          : "";
        const res = await fetch(`/api/restaurants/${encodeURIComponent(placeId)}${qs}`);
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
          setError(e instanceof Error ? e.message : "Failed to load report");
          setPhase("ready");
          bootstrappingRef.current = false;
        }
      }
    }
    load();
    return () => {
      cancelled = true;
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
        mobileScreenshot={
          data?.report.websiteMobileScreenshot ||
          data?.report.websiteScreenshot ||
          undefined
        }
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
        <a href="/" className="mt-6 inline-block font-semibold text-primary underline">
          Search again
        </a>
      </div>
    );
  }

  const { place, report } = data;
  const summaryLines = (report.aiSummary || "")
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);
  const unlocked = report.fullReportLocked === false || step === "unlocked";

  return (
    <div className="hero-atmosphere min-h-[70vh] px-5 py-10 sm:px-8 md:px-10 md:py-14">
      <div className="report-reveal mx-auto w-full max-w-6xl space-y-5 md:space-y-6">
        {/* 1 — Overview */}
        <ReportSection eyebrow="AI restaurant report" bodyClassName="px-5 py-6 sm:px-7 sm:py-7">
          <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
            <div className="min-w-0 max-w-2xl">
              <h1 className="font-display text-[clamp(1.75rem,3.8vw,2.75rem)] font-semibold leading-[1.12] tracking-[-0.03em] text-ink">
                {report.restaurantName}
              </h1>
              {report.address ? (
                <p className="mt-2 text-[14px] leading-relaxed text-muted sm:text-[15px]">{report.address}</p>
              ) : null}
              <p className="mt-4 text-[15px] leading-relaxed text-muted sm:text-[16px]">
                We scored SEO keywords, recent reviews, website design (live screenshot + AI), then deeper
                order-online, menu, and contact signals after you verify.
              </p>
            </div>

            <div className="shrink-0 rounded-2xl border border-border bg-[#f7f4ef] px-5 py-4 sm:min-w-[168px]">
              <p className="text-[11px] font-semibold uppercase tracking-[0.07em] text-muted">Overall score</p>
              <p className="mt-1 flex items-baseline gap-2">
                <span className="font-display text-[2.4rem] font-semibold leading-none tracking-[-0.04em] text-ink">
                  {report.overallScore}
                </span>
                <span className="text-[13px] font-semibold" style={{ color: report.overallColor }}>
                  {report.overallLabel}
                </span>
              </p>
              <p className="mt-2 text-[12px] text-muted">AI scoring review</p>
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
        </ReportSection>

        <div className="grid gap-5 lg:grid-cols-[minmax(0,1.05fr)_minmax(300px,0.95fr)] lg:items-start lg:gap-6">
          {/* Left column — narrative sections */}
          <div className="space-y-5">
            {summaryLines.length > 0 ? (
              <ReportSection eyebrow="01 · Findings" title="AI summary">
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

            {report.websiteScreenshot || report.websiteReview ? (
              <ReportSection eyebrow="03 · Website" title="Screenshot & AI design review" bodyClassName="p-0">
                <div className="border-b border-border px-5 py-4 sm:px-6">
                  {typeof report.websiteQualityScore === "number" && report.websiteQualityScore > 0 ? (
                    <p className="text-[15px] text-ink">
                      Visual quality <strong>{report.websiteQualityScore}</strong>{" "}
                      <span className="text-muted">(strict AI design score)</span>
                    </p>
                  ) : (
                    <p className="text-[14px] text-muted">Live homepage capture + design notes.</p>
                  )}
                </div>
                <LockedBlur
                  locked={!unlocked}
                  label="Verify email to unlock full website AI review"
                  className="min-h-[120px]"
                >
                  <div className="px-5 py-4 sm:px-6">
                    {report.websiteReview ? (
                      <p className="text-[14px] leading-relaxed text-muted">{report.websiteReview}</p>
                    ) : (
                      <p className="text-[14px] leading-relaxed text-muted">
                        Detailed AI design notes unlock after email verification.
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
                onUnlock={() => leadRef.current?.scrollIntoView({ behavior: "smooth", block: "center" })}
              />
            </ReportSection>

            <ReportSection eyebrow="06 · Fixes" bodyClassName="bg-[#f4f0ea] p-3 sm:p-3.5">
              <LiveIssuesCard
                issues={report.issues}
                estimatedMonthlyLoss={report.estimatedMonthlyLoss}
                locked={!unlocked}
                onFix={() => leadRef.current?.scrollIntoView({ behavior: "smooth", block: "center" })}
              />
            </ReportSection>
          </aside>
        </div>

        <LiveListingMedia media={place.media} />
      </div>
    </div>
  );
}
