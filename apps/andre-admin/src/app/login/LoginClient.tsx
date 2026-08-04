"use client";

import { FormEvent, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

export function LoginClient() {
  const router = useRouter();
  const params = useSearchParams();
  const preloadEmail = process.env.NEXT_PUBLIC_PRELOAD_ADMIN_EMAIL || "";
  const preloadPassword = process.env.NEXT_PUBLIC_PRELOAD_ADMIN_PASSWORD || "";
  const [email, setEmail] = useState(preloadEmail);
  const [password, setPassword] = useState(preloadPassword);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const reasonMessage = useMemo(() => {
    const reason = params.get("reason");
    if (reason === "idle") return "Signed out after 10 minutes of inactivity.";
    if (reason === "auth") return "Please sign in to continue.";
    return "";
  }, [params]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const data = (await res.json()) as { error?: string };
      if (!res.ok) {
        setError(data.error || "Login failed");
        return;
      }
      router.replace(params.get("next") || "/properties");
    } catch {
      setError("Could not reach login service");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen grid place-items-center px-4">
      <form
        onSubmit={onSubmit}
        className="w-full max-w-md rounded-2xl border border-line bg-panel p-8 shadow-[0_20px_60px_rgba(28,36,48,0.08)]"
      >
        <p
          className="text-3xl tracking-tight"
          style={{ fontFamily: "var(--font-fraunces), Georgia, serif" }}
        >
          Real Voice Agent Admin
        </p>
        <p className="mt-2 text-sm text-muted">
          Sign in to manage listings and run the voice agent.
        </p>

        {(reasonMessage || error) && (
          <p className={`mt-4 text-sm ${error ? "text-[var(--bad)]" : "text-[var(--warn)]"}`}>
            {error || reasonMessage}
          </p>
        )}

        <div className="mt-6 space-y-4">
          <div className="field">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              autoComplete="username"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className="field">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
        </div>

        <button type="submit" className="btn btn-primary mt-6 w-full" disabled={loading}>
          {loading ? "Signing in…" : "Sign in"}
        </button>
      </form>
    </div>
  );
}
