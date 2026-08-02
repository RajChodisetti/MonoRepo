"use client";

import { FormEvent, useEffect, useId, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import type { RestaurantDetails } from "@/lib/places";
import type { RestaurantReport } from "@/lib/report";
import LiveCompetitorsCard from "@/components/report/LiveCompetitorsCard";
import LiveHealthCard from "@/components/report/LiveHealthCard";
import LiveIssuesCard from "@/components/report/LiveIssuesCard";

type Payload = {
  place: RestaurantDetails;
  report: RestaurantReport;
};

export default function ReportClient({ placeId }: { placeId: string }) {
  const searchParams = useSearchParams();
  const unlockFromLink = (searchParams.get("unlock") || "").trim();

  const [data, setData] = useState<Payload | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
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
    async function load() {
      setLoading(true);
      setError(null);
      try {
        const qs = unlockToken
          ? `?unlock=${encodeURIComponent(unlockToken)}`
          : "";
        const res = await fetch(`/api/restaurants/${encodeURIComponent(placeId)}${qs}`);
        const json = await res.json();
        if (!res.ok) throw new Error(json.error || "Failed to load report");
        if (!cancelled) {
          const payload = json as Payload;
          setData(payload);
          if (payload.report.fullReportLocked === false) {
            setStep("unlocked");
          }
        }
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : "Failed to load report");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    load();
    return () => {
      cancelled = true;
    };
  }, [placeId, unlockToken]);

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

  if (loading) {
    return (
      <div className="mx-auto max-w-3xl px-6 py-24 text-center">
        <p className="font-display text-2xl font-semibold text-ink">Building your AI report…</p>
        <p className="mt-2 text-muted">Pulling Google listing signals for this restaurant.</p>
      </div>
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
    <div className="hero-atmosphere min-h-[70vh] px-6 py-12 md:px-10 md:py-16">
      <div className="mx-auto grid w-full max-w-5xl gap-10 lg:grid-cols-[1.05fr_0.95fr] lg:items-start">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.08em] text-muted">AI restaurant report</p>
          <h1 className="mt-3 font-display text-[clamp(1.8rem,4vw,3rem)] font-semibold leading-[1.15] tracking-[-0.03em] text-ink">
            {report.restaurantName}
          </h1>
          {report.address ? <p className="mt-2 text-[15px] text-muted">{report.address}</p> : null}
          <p className="mt-6 max-w-md text-[16px] leading-relaxed text-muted">
            Score <strong className="text-ink">{report.overallScore}/100</strong> ({report.overallLabel}).
            We scored SEO keywords, recent reviews, website design (from a live screenshot + AI review),
            order-online, menu, and contact signals.
          </p>

          {summaryLines.length > 0 ? (
            <div className="mt-6 max-w-md rounded-2xl border border-border bg-bg/80 p-5">
              <p className="text-[13px] font-semibold uppercase tracking-[0.06em] text-muted">AI summary</p>
              <div className="mt-3 space-y-2 text-[15px] leading-relaxed text-ink">
                {summaryLines.map((line) => (
                  <p key={line}>{line}</p>
                ))}
              </div>
              {!unlocked ? (
                <p className="mt-4 text-[13px] font-medium text-primary">
                  {report.unlockCta || "Unlock the full SEO report by verifying your email."}
                </p>
              ) : (
                <p className="mt-4 text-[13px] font-medium text-accent">Full report unlocked.</p>
              )}
            </div>
          ) : null}

          <div ref={leadRef} className="mt-8 rounded-2xl border border-border bg-bg p-5 shadow-[0_10px_40px_rgba(15,39,31,0.08)]">
            {unlocked ? (
              <p className="text-[15px] font-semibold text-ink">
                You&apos;re on the list — we saved this restaurant as an interested lead.
              </p>
            ) : step === "otp" ? (
              <>
                <p className="text-[15px] font-semibold text-ink">Enter the verification code</p>
                <p className="mt-1 text-[13px] text-muted">We sent a 6-digit code to {email}.</p>
                {devOtp ? (
                  <p className="mt-2 rounded-lg bg-[#f3f1ed] px-3 py-2 text-[13px] text-ink">
                    Local dev code: <strong>{devOtp}</strong>
                  </p>
                ) : null}
                <form className="mt-3 flex flex-col gap-2 sm:flex-row" onSubmit={verifyOtp}>
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
                <p className="text-[15px] font-semibold text-ink">Verify email to unlock full report</p>
                <form className="mt-3 flex flex-col gap-2 sm:flex-row" onSubmit={requestOtp}>
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
                className={`mt-2 text-[13px] ${leadStatus === "ok" ? "text-accent" : leadStatus === "err" ? "text-[#b42318]" : "text-muted"}`}
              >
                {leadMessage}
              </p>
            ) : null}
          </div>

          {place.website || place.mapsUri ? (
            <p className="mt-4 text-[13px] text-muted">
              {place.website ? (
                <a href={place.website} className="mr-4 font-medium text-primary underline" target="_blank" rel="noreferrer">
                  Website
                </a>
              ) : null}
              {place.mapsUri ? (
                <a href={place.mapsUri} className="font-medium text-primary underline" target="_blank" rel="noreferrer">
                  Google Maps
                </a>
              ) : null}
            </p>
          ) : null}

          {report.websiteScreenshot || report.websiteReview ? (
            <div className="mt-8 max-w-md overflow-hidden rounded-2xl border border-border bg-bg/80">
              <div className="border-b border-border px-5 py-4">
                <p className="text-[13px] font-semibold uppercase tracking-[0.06em] text-muted">
                  Website screenshot review
                </p>
                {typeof report.websiteQualityScore === "number" && report.websiteQualityScore > 0 ? (
                  <p className="mt-2 text-[15px] text-ink">
                    Visual quality{" "}
                    <strong>
                      {report.websiteQualityScore}
                      /100
                    </strong>{" "}
                    <span className="text-muted">(strict design score)</span>
                  </p>
                ) : null}
                {report.websiteReview ? (
                  <p className="mt-2 text-[14px] leading-relaxed text-muted">{report.websiteReview}</p>
                ) : null}
              </div>
              {report.websiteScreenshot ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={report.websiteScreenshot}
                  alt={`${report.restaurantName} website homepage screenshot`}
                  className="block w-full bg-[#efebe6] object-cover object-top"
                />
              ) : null}
            </div>
          ) : null}
        </div>

        <div className="mx-auto w-full max-w-[340px] space-y-4 rounded-[28px] bg-[#efebe6] p-4 shadow-[0_20px_60px_rgba(15,39,31,0.12)]">
          <LiveHealthCard
            restaurantName={report.restaurantName}
            overallScore={report.overallScore}
            overallLabel={report.overallLabel}
            overallColor={report.overallColor}
            metrics={report.metrics}
          />
          <LiveCompetitorsCard rows={report.competitors} />
          <LiveIssuesCard
            issues={report.issues}
            estimatedMonthlyLoss={report.estimatedMonthlyLoss}
            onFix={() => leadRef.current?.scrollIntoView({ behavior: "smooth", block: "center" })}
          />
        </div>
      </div>
    </div>
  );
}
