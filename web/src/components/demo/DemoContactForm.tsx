"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";

type FormState = {
  name: string;
  email: string;
  phone: string;
  restaurantName: string;
  city: string;
  message: string;
};

const empty: FormState = {
  name: "",
  email: "",
  phone: "",
  restaurantName: "",
  city: "",
  message: "",
};

export default function DemoContactForm() {
  const [form, setForm] = useState<FormState>(empty);
  const [submitting, setSubmitting] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function update<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      const res = await fetch("/api/contact", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ...form, source: "website-demo" }),
      });
      const json = await res.json().catch(() => ({}));
      if (!res.ok) {
        const msg =
          (typeof json.error === "string" && json.error) ||
          (typeof json.message === "string" && json.message) ||
          (json.error && typeof json.error.message === "string" && json.error.message) ||
          "Could not send your request. Try again.";
        throw new Error(msg);
      }
      setDone(true);
      setForm(empty);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong.");
    } finally {
      setSubmitting(false);
    }
  }

  if (done) {
    return (
      <div className="rounded-2xl border border-border bg-bg px-6 py-10 text-center shadow-[0_12px_40px_rgba(15,39,31,0.08)] sm:px-10">
        <p className="text-[13px] font-semibold uppercase tracking-[0.08em] text-primary">Request received</p>
        <h2 className="mt-3 font-display text-[1.75rem] font-semibold tracking-[-0.03em] text-ink">
          Thanks — we&apos;ll be in touch soon
        </h2>
        <p className="mx-auto mt-3 max-w-[36ch] text-[15px] leading-relaxed text-muted">
          Your demo request landed in our inbox. A Tuvi teammate will reply from contact@tuvisolutions.com.
        </p>
        <Link
          href="/"
          className="mt-8 inline-flex rounded-full bg-primary px-6 py-3 text-[15px] font-semibold text-bg transition-colors hover:bg-primary-dim"
        >
          Back to home
        </Link>
      </div>
    );
  }

  return (
    <form
      onSubmit={onSubmit}
      className="rounded-2xl border border-border bg-bg p-6 shadow-[0_12px_40px_rgba(15,39,31,0.08)] sm:p-8"
      noValidate
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Your name" required>
          <input
            required
            name="name"
            value={form.name}
            onChange={(e) => update("name", e.target.value)}
            className={inputClass}
            placeholder="Alex Chen"
            autoComplete="name"
          />
        </Field>
        <Field label="Work email" required>
          <input
            required
            type="email"
            name="email"
            value={form.email}
            onChange={(e) => update("email", e.target.value)}
            className={inputClass}
            placeholder="you@restaurant.com"
            autoComplete="email"
          />
        </Field>
        <Field label="Phone">
          <input
            type="tel"
            name="phone"
            value={form.phone}
            onChange={(e) => update("phone", e.target.value)}
            className={inputClass}
            placeholder="+61…"
            autoComplete="tel"
          />
        </Field>
        <Field label="City">
          <input
            name="city"
            value={form.city}
            onChange={(e) => update("city", e.target.value)}
            className={inputClass}
            placeholder="Sydney"
            autoComplete="address-level2"
          />
        </Field>
        <Field label="Restaurant name" required className="sm:col-span-2">
          <input
            required
            name="restaurantName"
            value={form.restaurantName}
            onChange={(e) => update("restaurantName", e.target.value)}
            className={inputClass}
            placeholder="Quillnest Kitchen"
          />
        </Field>
        <Field label="How can we help?" className="sm:col-span-2">
          <textarea
            name="message"
            rows={4}
            value={form.message}
            onChange={(e) => update("message", e.target.value)}
            className={`${inputClass} resize-y`}
            placeholder="Tell us about your locations, goals, or timeline…"
          />
        </Field>
      </div>

      {error ? <p className="mt-4 text-[14px] text-[#b42318]">{error}</p> : null}

      <button
        type="submit"
        disabled={submitting}
        className="mt-6 inline-flex w-full cursor-pointer items-center justify-center rounded-full bg-primary px-6 py-3.5 text-[15px] font-semibold text-bg transition-colors hover:bg-primary-dim disabled:opacity-60 sm:w-auto"
      >
        {submitting ? "Sending…" : "Request a free demo"}
      </button>
      <p className="mt-3 text-[12px] text-muted">
        We reply from contact@tuvisolutions.com. No spam — just a real follow-up.
      </p>
    </form>
  );
}

const inputClass =
  "mt-1.5 w-full rounded-xl border border-border bg-bg px-3.5 py-3 text-[15px] text-ink outline-none transition-colors placeholder:text-secondary focus:border-primary";

function Field({
  label,
  required,
  className = "",
  children,
}: {
  label: string;
  required?: boolean;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <label className={`block text-left text-[13px] font-medium text-ink ${className}`}>
      {label}
      {required ? <span className="text-primary"> *</span> : null}
      {children}
    </label>
  );
}
