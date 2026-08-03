"use client";

import { useState, type FormEvent } from "react";
import { getCallInDisplay, getCallInTelHref } from "@/lib/env";

type FormStatus = "idle" | "loading" | "success" | "error" | "queued";

export default function RequestCallbackForm({
  compact = false,
  onSuccess,
}: {
  compact?: boolean;
  onSuccess?: () => void;
}) {
  const [phone, setPhone] = useState("");
  const [nameVal, setNameVal] = useState("");
  const [status, setStatus] = useState<FormStatus>("idle");
  const [message, setMessage] = useState("");
  const callInDisplay = getCallInDisplay();
  const callInHref = getCallInTelHref();

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setStatus("loading");
    setMessage("");
    try {
      const res = await fetch("/api/voice-agent/call", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ phone: phone.trim(), name: nameVal.trim() || undefined }),
      });
      const data = (await res.json()) as { status?: string; message?: string };
      if (data.status === "calling") {
        setStatus("success");
        setMessage(data.message || "We're calling you now — please answer your phone.");
        onSuccess?.();
        return;
      }
      if (data.status === "queued") {
        setStatus("queued");
        setMessage(
          data.message ||
            "We're outside calling hours. Try again later or book a consultation.",
        );
        return;
      }
      setStatus("error");
      setMessage(data.message || "Could not request a callback.");
    } catch {
      setStatus("error");
      setMessage("Network error. Please try again.");
    }
  }

  const inputClass =
    "w-full rounded-xl border border-border bg-bg px-3 py-2.5 text-sm text-ink placeholder:text-muted/70 outline-none transition focus:border-primary";

  return (
    <div className={compact ? "space-y-3" : "mx-auto max-w-md space-y-3 text-left"}>
      {!compact && (
        <div>
          <p className="text-sm font-semibold text-ink">Get a callback</p>
          <p className="mt-1 text-xs text-muted">
            Leave your number and we&apos;ll call when the service is available. Example: +61 412 345 678
          </p>
        </div>
      )}

      <form onSubmit={onSubmit} className="space-y-2">
        {!compact ? (
          <label className="block">
            <span className="mb-1.5 block text-[11px] font-bold uppercase tracking-[0.12em] text-muted">
              Name <span className="normal-case tracking-normal">(optional)</span>
            </span>
            <input
              type="text"
              name="name"
              autoComplete="name"
              placeholder="Jane Smith"
              value={nameVal}
              onChange={(e) => setNameVal(e.target.value)}
              className={inputClass}
            />
          </label>
        ) : null}
        <label className="block">
          <span
            className={`${compact ? "sr-only" : "mb-1.5 block text-[11px] font-bold uppercase tracking-[0.12em] text-muted"}`}
          >
            Phone number
          </span>
          <input
            type="tel"
            name="phone"
            required
            autoComplete="tel"
            placeholder="Phone (+61…)"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            className={inputClass}
          />
        </label>
        <button
          type="submit"
          disabled={status === "loading"}
          className="w-full cursor-pointer rounded-xl bg-ink px-4 py-3 text-sm font-semibold text-bg transition-colors hover:bg-primary-dim disabled:opacity-60"
        >
          {status === "loading" ? "Requesting…" : "Request a callback"}
        </button>
      </form>

      {message && (
        <p
          className={`text-xs leading-relaxed ${
            status === "success"
              ? "text-accent"
              : status === "queued"
                ? "text-[#b45309]"
                : "text-[#b42318]"
          }`}
          role="status"
        >
          {message}
        </p>
      )}

      {callInDisplay && callInHref ? (
        <p className="text-xs text-muted">
          Prefer to dial us?{" "}
          <a href={callInHref} className="font-medium text-primary transition-colors hover:text-ink">
            {callInDisplay}
          </a>
        </p>
      ) : null}
    </div>
  );
}
