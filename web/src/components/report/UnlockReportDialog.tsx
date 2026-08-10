"use client";

import { FormEvent, useEffect, useId, useRef, useState } from "react";
import type { RestaurantDetails } from "@/lib/places";
import type { RestaurantReport } from "@/lib/report";

type UnlockedPayload = {
  place: RestaurantDetails;
  report: RestaurantReport;
};

function apiError(payload: unknown, fallback: string): string {
  if (!payload || typeof payload !== "object") return fallback;
  const value = payload as {
    error?: string | { message?: string };
    message?: string;
  };
  if (typeof value.error === "string" && value.error.trim()) return value.error;
  if (typeof value.error === "object" && value.error?.message?.trim()) {
    return value.error.message;
  }
  return value.message?.trim() || fallback;
}

export default function UnlockReportDialog({
  open,
  placeId,
  restaurantName,
  notice,
  onClose,
  onUnlocked,
}: {
  open: boolean;
  placeId: string;
  restaurantName: string;
  notice?: string;
  onClose: () => void;
  onUnlocked: (payload: UnlockedPayload) => void;
}) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const nameId = useId();
  const emailId = useId();
  const phoneId = useId();
  const otpId = useId();
  const [step, setStep] = useState<"contact" | "otp" | "unlocked">("contact");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [otp, setOtp] = useState("");
  const [status, setStatus] = useState<"idle" | "sending" | "success" | "error">("idle");
  const [message, setMessage] = useState("");
  const [devOtp, setDevOtp] = useState("");

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
    if (!open && dialog.open) dialog.close();
  }, [open]);

  async function requestCode(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatus("sending");
    setMessage("");
    setDevOtp("");
    try {
      const response = await fetch("/api/leads", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: name.trim(),
          email: email.trim(),
          phone: phone.trim(),
          placeId,
        }),
      });
      const payload = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(apiError(payload, "Could not send the code."));
      setStatus("success");
      setMessage(payload.message || "A 6-digit code is on its way to your inbox.");
      if (payload.devOtp) setDevOtp(String(payload.devOtp));
      setStep("otp");
    } catch (error) {
      setStatus("error");
      setMessage(error instanceof Error ? error.message : "Could not send the code.");
    }
  }

  async function verifyCode(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatus("sending");
    setMessage("");
    try {
      const response = await fetch("/api/leads/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email.trim(), placeId, otp }),
      });
      const payload = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(apiError(payload, "That code is invalid or expired."));
      if (!payload.place || !payload.report || payload.report.fullReportLocked !== false) {
        throw new Error("The server did not confirm this report unlock. Please request a new code.");
      }
      setStatus("success");
      setMessage("Verified. Your full report and download are ready.");
      setStep("unlocked");
      onUnlocked(payload as UnlockedPayload);
    } catch (error) {
      setStatus("error");
      setMessage(error instanceof Error ? error.message : "Could not verify the code.");
    }
  }

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby="unlock-report-title"
      aria-describedby="unlock-report-description"
      onCancel={(event) => {
        event.preventDefault();
        onClose();
      }}
      onClose={onClose}
      className="m-auto w-[min(94vw,560px)] max-w-none rounded-[28px] border border-border bg-white p-0 text-ink shadow-[0_28px_90px_rgba(10,35,27,0.34)] backdrop:bg-[#0b1d17]/55 backdrop:backdrop-blur-[2px]"
    >
      <div className="relative overflow-hidden rounded-[28px]">
        <div className="bg-primary px-6 py-6 text-white sm:px-8">
          <p className="text-[11px] font-bold uppercase tracking-[0.13em] text-white/70">
            Free venue pulse
          </p>
          <h2 id="unlock-report-title" className="mt-2 font-display text-[2rem] font-semibold leading-none">
            Unlock your improvement plan
          </h2>
          <p id="unlock-report-description" className="mt-3 max-w-[48ch] text-[14px] leading-relaxed text-white/80">
            Verify your email to reveal the full score breakdown, genuine nearby competitor snapshot,
            prioritized fixes, and the two-page Tuvi report for {restaurantName}.
          </p>
        </div>

        <button
          type="button"
          onClick={onClose}
          aria-label="Close verification"
          className="absolute right-4 top-4 flex h-11 w-11 cursor-pointer items-center justify-center rounded-full bg-white/12 text-xl text-white transition-colors hover:bg-white/20 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
        >
          ×
        </button>

        <div className="px-6 py-6 sm:px-8 sm:py-7">
          {step === "contact" ? (
            <form className="space-y-4" onSubmit={requestCode}>
              {notice ? (
                <p role="alert" className="rounded-xl bg-[#fff2dc] px-4 py-3 text-[13px] leading-relaxed text-[#8a5200]">
                  {notice}
                </p>
              ) : null}
              <div>
                <label htmlFor={nameId} className="text-[13px] font-semibold text-ink">
                  Your name
                </label>
                <input
                  id={nameId}
                  autoFocus
                  autoComplete="name"
                  minLength={2}
                  maxLength={100}
                  required
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="Restaurant owner or manager"
                  className="mt-1.5 min-h-12 w-full rounded-xl border border-border bg-[#fbfaf8] px-4 text-[15px] outline-none focus:border-primary focus-visible:ring-2 focus-visible:ring-primary/25"
                />
              </div>
              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <label htmlFor={emailId} className="text-[13px] font-semibold text-ink">
                    Work email
                  </label>
                  <input
                    id={emailId}
                    type="email"
                    autoComplete="email"
                    maxLength={254}
                    required
                    value={email}
                    onChange={(event) => setEmail(event.target.value)}
                    placeholder="you@restaurant.com"
                    className="mt-1.5 min-h-12 w-full rounded-xl border border-border bg-[#fbfaf8] px-4 text-[15px] outline-none focus:border-primary focus-visible:ring-2 focus-visible:ring-primary/25"
                  />
                </div>
                <div>
                  <label htmlFor={phoneId} className="text-[13px] font-semibold text-ink">
                    Phone number
                  </label>
                  <input
                    id={phoneId}
                    type="tel"
                    inputMode="tel"
                    autoComplete="tel"
                    minLength={7}
                    maxLength={40}
                    required
                    value={phone}
                    onChange={(event) => setPhone(event.target.value)}
                    placeholder="+61 4…"
                    className="mt-1.5 min-h-12 w-full rounded-xl border border-border bg-[#fbfaf8] px-4 text-[15px] outline-none focus:border-primary focus-visible:ring-2 focus-visible:ring-primary/25"
                  />
                </div>
              </div>
              <p className="text-[12px] leading-relaxed text-muted">
                We store these details with this report request. Verification is not marketing consent.
              </p>
              <button
                type="submit"
                disabled={status === "sending"}
                className="min-h-12 w-full cursor-pointer rounded-full bg-primary px-5 text-[15px] font-semibold text-white transition-colors hover:bg-primary-dim disabled:cursor-wait disabled:opacity-60"
              >
                {status === "sending" ? "Sending code…" : "Email my verification code"}
              </button>
            </form>
          ) : step === "otp" ? (
            <form className="space-y-4" onSubmit={verifyCode}>
              <div>
                <p className="text-[15px] font-semibold text-ink">Enter the 6-digit code</p>
                <p className="mt-1 text-[13px] text-muted">Sent to {email}. It expires in 15 minutes.</p>
              </div>
              {devOtp ? (
                <p className="rounded-xl bg-[#f4f1ec] px-4 py-3 text-[13px] text-ink">
                  Local-only code: <strong>{devOtp}</strong>
                </p>
              ) : null}
              <div>
                <label htmlFor={otpId} className="sr-only">
                  Six-digit verification code
                </label>
                <input
                  id={otpId}
                  autoFocus
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  pattern="[0-9]{6}"
                  maxLength={6}
                  required
                  value={otp}
                  onChange={(event) => setOtp(event.target.value.replace(/\D/g, "").slice(0, 6))}
                  placeholder="123456"
                  className="min-h-14 w-full rounded-xl border border-border bg-[#fbfaf8] px-4 text-center text-[24px] font-semibold tracking-[0.35em] outline-none focus:border-primary focus-visible:ring-2 focus-visible:ring-primary/25"
                />
              </div>
              <button
                type="submit"
                disabled={status === "sending" || otp.length !== 6}
                className="min-h-12 w-full cursor-pointer rounded-full bg-primary px-5 text-[15px] font-semibold text-white transition-colors hover:bg-primary-dim disabled:cursor-not-allowed disabled:opacity-60"
              >
                {status === "sending" ? "Verifying…" : "Verify and unlock report"}
              </button>
              <div className="flex flex-wrap justify-between gap-3 text-[13px]">
                <button
                  type="button"
                  className="cursor-pointer font-semibold text-primary underline underline-offset-4"
                  onClick={() => {
                    setStep("contact");
                    setOtp("");
                    setMessage("");
                    setStatus("idle");
                  }}
                >
                  Change details
                </button>
                <button
                  type="button"
                  className="cursor-pointer font-semibold text-primary underline underline-offset-4 disabled:opacity-50"
                  disabled={status === "sending"}
                  onClick={() => {
                    setStep("contact");
                    setOtp("");
                    setStatus("idle");
                    setMessage("Confirm your details to request a new code.");
                  }}
                >
                  Request a new code
                </button>
              </div>
            </form>
          ) : (
            <div className="text-center">
              <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-[#e8f6ee] text-2xl text-[#176b3a]">
                ✓
              </div>
              <p className="mt-4 font-display text-2xl font-semibold text-ink">Full report unlocked</p>
              <p className="mt-2 text-[14px] leading-relaxed text-muted">
                Your detailed criteria, improvement plan, and live competitor snapshot are ready.
              </p>
              <a
                href={`/api/restaurants/${encodeURIComponent(placeId)}/report.pdf`}
                className="mt-5 inline-flex min-h-12 items-center justify-center rounded-full bg-primary px-6 text-[15px] font-semibold text-white hover:bg-primary-dim focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
              >
                Download the 2-page PDF
              </a>
            </div>
          )}

          {message ? (
            <p
              role={status === "error" ? "alert" : "status"}
              className={`mt-4 rounded-xl px-4 py-3 text-[13px] ${
                status === "error" ? "bg-[#fff0ee] text-[#a72b20]" : "bg-[#eef6f1] text-accent"
              }`}
            >
              {message}
            </p>
          ) : null}
        </div>
      </div>
    </dialog>
  );
}
