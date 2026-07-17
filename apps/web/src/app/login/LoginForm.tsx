"use client";

import { FormEvent, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

export default function LoginForm() {
  const router = useRouter();
  const search = useSearchParams();
  const next = search.get("next") || "/dashboard";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await fetch("/api/admin/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data?.error?.message || "Login failed");
      }
      router.replace(next.startsWith("/") ? next : "/dashboard");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "grid",
        placeItems: "center",
        padding: "1.5rem",
        background:
          "radial-gradient(circle at top left, #d8ebe6 0%, transparent 45%), var(--bg)",
      }}
    >
      <form
        onSubmit={onSubmit}
        className="card"
        style={{ width: "100%", maxWidth: 420 }}
      >
        <p
          style={{
            margin: 0,
            fontFamily: "var(--font-fraunces), serif",
            fontSize: "1.75rem",
            fontWeight: 600,
          }}
        >
          Tuvi Admin
        </p>
        <p style={{ margin: "0.35rem 0 1.25rem", color: "var(--muted)" }}>
          Sign in with an internal_admin account
        </p>

        {error ? (
          <div className="alert alert-error" style={{ marginBottom: "1rem" }}>
            {error}
          </div>
        ) : null}

        <label style={{ display: "grid", gap: "0.35rem", marginBottom: "0.85rem" }}>
          <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>Email</span>
          <input
            className="input"
            type="email"
            autoComplete="username"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>

        <label style={{ display: "grid", gap: "0.35rem", marginBottom: "1.1rem" }}>
          <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>Password</span>
          <input
            className="input"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>

        <button
          className="btn btn-primary"
          type="submit"
          disabled={loading}
          style={{ width: "100%" }}
        >
          {loading ? "Signing in…" : "Sign in"}
        </button>
      </form>
    </div>
  );
}
