import { NextResponse } from "next/server";
import {
  REPORT_UNLOCK_COOKIE,
  REPORT_UNLOCK_MAX_AGE_SECONDS,
} from "@/lib/report-unlock";
import { seoForwardingHeaders } from "@/lib/seo-forwarding";

export const dynamic = "force-dynamic";

function monorepoBase(): string {
  return (process.env.MONOREPO_API_URL?.trim() || "http://localhost:8080").replace(/\/$/, "");
}

/** Proxy OTP verify to MonoRepo Go API. */
export async function POST(request: Request) {
  try {
    const body = await request.json();
    const res = await fetch(`${monorepoBase()}/api/public/v1/seo/unlock/verify`, {
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
    if (!res.ok) {
      return NextResponse.json(json, {
        status: res.status,
        headers: responseHeaders,
      });
    }

    const unlockToken =
      typeof json.unlockToken === "string" ? json.unlockToken.trim() : "";
    if (!unlockToken || json.report?.fullReportLocked !== false) {
      return NextResponse.json(
        { error: "The API did not confirm this report unlock." },
        { status: 502, headers: { "Cache-Control": "private, no-store" } },
      );
    }
    const { unlockToken: _privateUnlockToken, ...safePayload } = json;
    void _privateUnlockToken;
    const response = NextResponse.json(safePayload, {
      status: res.status,
      headers: { "Cache-Control": "private, no-store" },
    });
    if (unlockToken) {
      response.cookies.set(REPORT_UNLOCK_COOKIE, unlockToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        path: "/",
        maxAge: REPORT_UNLOCK_MAX_AGE_SECONDS,
        priority: "high",
      });
    }
    return response;
  } catch (err) {
    console.error("[leads/verify]", err);
    return NextResponse.json(
      { error: "Failed to verify code" },
      { status: 500, headers: { "Cache-Control": "private, no-store" } },
    );
  }
}
