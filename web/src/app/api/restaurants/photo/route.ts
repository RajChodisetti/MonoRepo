import { NextResponse } from "next/server";
import { seoForwardingHeaders } from "@/lib/seo-forwarding";

export const dynamic = "force-dynamic";

function monorepoBase(): string {
  return (process.env.MONOREPO_API_URL?.trim() || "http://localhost:8080").replace(/\/$/, "");
}

/** Proxies Google Places listing photos via MonoRepo (API key stays server-side). */
export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const name = (searchParams.get("name") || "").trim();
  if (!name || !name.startsWith("places/") || !name.includes("/photos/")) {
    return NextResponse.json(
      { error: "Invalid photo name" },
      { status: 400, headers: { "Cache-Control": "private, no-store" } },
    );
  }
  const max = (searchParams.get("max") || "720").trim();

  try {
    const qs = new URLSearchParams({ name, max });
    const res = await fetch(`${monorepoBase()}/api/public/v1/seo/photo?${qs.toString()}`, {
      cache: "no-store",
      headers: seoForwardingHeaders(request),
    });
    if (!res.ok) {
      const retryAfter = res.headers.get("retry-after")?.trim();
      return NextResponse.json(
        { error: res.status === 429 ? "Too many photo requests" : "Photo unavailable" },
        {
          status: res.status === 404 || res.status === 429 ? res.status : 502,
          headers: {
            "Cache-Control": "private, no-store",
            ...(res.status === 429 && retryAfter ? { "Retry-After": retryAfter } : {}),
          },
        },
      );
    }
    const buf = await res.arrayBuffer();
    const contentType = res.headers.get("content-type") || "image/jpeg";
    return new NextResponse(buf, {
      status: 200,
      headers: {
        "Content-Type": contentType,
        "Cache-Control": "private, no-store, max-age=0",
        Pragma: "no-cache",
      },
    });
  } catch (err) {
    console.error("[restaurants/photo]", err);
    return NextResponse.json(
      { error: "Failed to load photo" },
      { status: 500, headers: { "Cache-Control": "private, no-store" } },
    );
  }
}
