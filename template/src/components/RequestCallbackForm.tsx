"use client";

import { useState, type FormEvent } from "react";
import { telHref } from "@/lib/reservation";

type FormStatus = "idle" | "loading" | "success" | "error" | "queued";

export default function RequestCallbackForm({
  compact = false,
  isAurora = false,
  restaurantIndex = 0,
  restaurantName,
  restaurantPhone,
  onSuccess,
}: {
  compact?: boolean;
  isAurora?: boolean;
  restaurantIndex?: number;
  restaurantName?: string;
  restaurantPhone?: string;
  onSuccess?: () => void;
}) {
  const [phone, setPhone] = useState("");
  const [nameVal, setNameVal] = useState("");
  const [status, setStatus] = useState<FormStatus>("idle");
  const [message, setMessage] = useState("");

  const inputClass = isAurora
    ? "w-full rounded-xl border border-white/15 bg-white/5 px-3 py-2.5 text-sm text-white placeholder:text-white/40 outline-none transition focus:border-cyan-400/50"
    : "w-full rounded-xl border border-[#e8e0d4]/20 bg-black/20 px-3 py-2.5 text-sm text-[#f7f0e6] placeholder:text-[#a89f96]/70 outline-none transition focus:border-[#b88a44]/50";

  const buttonClass = isAurora
    ? "w-full rounded-xl bg-gradient-to-r from-violet-600 to-cyan-500 px-4 py-3 text-sm font-semibold text-white transition hover:from-violet-500 hover:to-cyan-400 disabled:opacity-60"
    : "w-full rounded-xl bg-[#b88a44] px-4 py-3 text-sm font-semibold text-[#1a1614] transition hover:bg-[#c99a54] disabled:opacity-60";

  const tel = telHref(restaurantPhone);
  const label = restaurantName || "our restaurant";

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setStatus("loading");
    setMessage("");
    try {
      const res = await fetch("/api/voice-agent/call", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          phone: phone.trim(),
          name: nameVal.trim() || undefined,
          restaurant_index: restaurantIndex,
        }),
      });
      const data = (await res.json()) as {
        status?: string;
        message?: string;
        caller_display?: string;
        caller_name?: string;
        from_verified?: boolean;
      };
      if (data.status === "calling") {
        setStatus("success");
        const callerNote =
          data.from_verified && data.caller_display
            ? `Answer — call from ${data.caller_name || label} (${data.caller_display}).`
            : data.caller_display
              ? `Answer — our AI receptionist for ${data.caller_name || label} is calling. Listed number: ${data.caller_display}.`
              : data.message || "We're calling you now — please answer your phone.";
        setMessage(data.message || callerNote);
        onSuccess?.();
        return;
      }
      if (data.status === "queued") {
        setStatus("queued");
        setMessage(
          data.message || "We're outside calling hours. Try again later or use voice chat.",
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

  return (
    <div className={compact ? "space-y-3" : "mx-auto mt-8 max-w-md space-y-3 text-left"}>
      {!compact && (
        <div>
          <p className={`text-sm font-semibold ${isAurora ? "text-white" : "text-[#f7f0e6]"}`}>
            Call me
          </p>
          <p className={`mt-1 text-xs ${isAurora ? "text-white/55" : "text-[#a89f96]"}`}>
            Our AI receptionist for {label} will call you now.
          </p>
        </div>
      )}

      {compact && (
        <p className={`text-xs leading-relaxed ${isAurora ? "text-white/60" : "text-[#a89f96]"}`}>
          {restaurantPhone
            ? `AI receptionist for ${label}. Caller may show as ${restaurantPhone}.`
            : `AI receptionist for ${label} will call you.`}
        </p>
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
        <button type="submit" disabled={status === "loading"} className={buttonClass}>
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

      {restaurantPhone && tel ? (
        <p className={`text-xs ${isAurora ? "text-white/50" : "text-[#a89f96]"}`}>
          Prefer to call the restaurant?{" "}
          <a
            href={tel}
            className={`font-medium ${isAurora ? "text-cyan-300 hover:text-cyan-200" : "text-[#b88a44] hover:text-[#c99a54]"}`}
          >
            {restaurantPhone}
          </a>
        </p>
      ) : null}
    </div>
  );
}
