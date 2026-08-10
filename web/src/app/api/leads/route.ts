import { NextResponse } from "next/server";
import { seoForwardingHeaders } from "@/lib/seo-forwarding";

export const dynamic = "force-dynamic";

function monorepoBase(): string {
  return (process.env.MONOREPO_API_URL?.trim() || "http://localhost:8080").replace(/\/$/, "");
}

/** Proxy OTP unlock request to MonoRepo Go API. */
export async function POST(request: Request) {
  try {
    const body = await request.json();
    const res = await fetch(`${monorepoBase()}/api/public/v1/seo/unlock/request`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...seoForwardingHeaders(request),
      },
      body: JSON.stringify(body),
      cache: "no-store",
    });
    const json = await res.json().catch(() => ({}));
    const responseHeaders: Record<string, string> = { "Cache-Control": "private, no-store" };
    const retryAfter = res.headers.get("retry-after")?.trim();
    if (retryAfter) responseHeaders["Retry-After"] = retryAfter;
    return NextResponse.json(json, {
      status: res.status,
      headers: responseHeaders,
    });
  } catch (err) {
    console.error("[leads/unlock-request]", err);
    return NextResponse.json(
      { error: "Failed to send verification code" },
      { status: 500, headers: { "Cache-Control": "private, no-store" } },
    );
  }
}
