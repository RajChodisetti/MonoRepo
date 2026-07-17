"use client";

import { withBasePath } from "@/lib/base-path";

export async function adminFetch<T = unknown>(
  path: string,
  options: {
    method?: string;
    body?: unknown;
    query?: Record<string, string | number | boolean | undefined | null>;
  } = {},
): Promise<T> {
  const url = new URL(
    withBasePath(
      path.startsWith("/api/admin/")
        ? path
        : `/api/admin/proxy/${path.replace(/^\//, "")}`,
    ),
    window.location.origin,
  );
  if (options.query) {
    for (const [k, v] of Object.entries(options.query)) {
      if (v !== undefined && v !== null && v !== "") {
        url.searchParams.set(k, String(v));
      }
    }
  }

  const res = await fetch(url.pathname + url.search, {
    method: options.method || "GET",
    headers:
      options.body !== undefined
        ? { "Content-Type": "application/json", Accept: "application/json" }
        : { Accept: "application/json" },
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
    credentials: "same-origin",
  });

  const text = await res.text();
  let data: unknown = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { error: { message: text.slice(0, 300) } };
    }
  }

  if (!res.ok) {
    const message =
      typeof data === "object" &&
      data !== null &&
      "error" in data &&
      typeof (data as { error?: { message?: string } }).error?.message ===
        "string"
        ? (data as { error: { message: string } }).error.message
        : `Request failed (${res.status})`;
    const err = new Error(message) as Error & { status: number; body: unknown };
    err.status = res.status;
    err.body = data;
    throw err;
  }

  return data as T;
}
