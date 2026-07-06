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
    "w-full rounded-xl border border-white/15 bg-white/5 px-3 py-2.5 text-sm text-text placeholder:text-muted/70 outline-none transition focus:border-cyan/50";

  return (
    <div className={compact ? "space-y-3" : "mx-auto mt-8 max-w-md space-y-3 text-left"}>
      {!compact && (
        <div>
          <p className="text-sm font-semibold text-text">Get a callback</p>
          <p className="mt-1 text-xs text-muted">
            Drop your number and our AI will call you now. Example: +61 412 345 678
          </p>
        </div>
      )}

      <form onSubmit={onSubmit} className="space-y-2">
        {!compact && (
          <input
            type="text"
            name="name"
            autoComplete="name"
            placeholder="Name (optional)"
            value={nameVal}
            onChange={(e) => setNameVal(e.target.value)}
            className={inputClass}
          />
        )}
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
        <button
          type="submit"
          disabled={status === "loading"}
          className="w-full rounded-xl bg-gradient-to-r from-gold-dim to-gold px-4 py-3 text-sm font-semibold text-bg transition hover:opacity-90 disabled:opacity-60"
        >
          {status === "loading" ? "Calling…" : "Call me now"}
        </button>
      </form>

      {message && (
        <p
          className={`text-xs leading-relaxed ${
            status === "success"
              ? "text-emerald-300"
              : status === "queued"
                ? "text-amber-200"
                : "text-red-300"
          }`}
          role="status"
        >
          {message}
        </p>
      )}

      {callInDisplay && callInHref ? (
        <p className="text-xs text-muted">
          Prefer to dial us?{" "}
          <a href={callInHref} className="font-medium text-cyan transition hover:text-gold">
            {callInDisplay}
          </a>
        </p>
      ) : null}
    </div>
  );
}
