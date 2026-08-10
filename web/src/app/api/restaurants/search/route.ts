import { NextResponse } from "next/server";
import { searchSeoRestaurants } from "@/lib/places";
import { seoForwardingHeaders } from "@/lib/seo-forwarding";

export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const q = (searchParams.get("q") || "").trim();
  const location = (searchParams.get("location") || "Australia").trim() || "Australia";
  if (q.length < 2) {
    return NextResponse.json({
      results: [],
      meta: { placesEnabled: false, inventoryEnabled: false },
    });
  }

  try {
    const payload = await searchSeoRestaurants(q, location, seoForwardingHeaders(request));
    return NextResponse.json(payload);
  } catch (err) {
    const status = (err as Error & { status?: number }).status;
    const retryAfter = (err as Error & { retryAfter?: string }).retryAfter;
    console.error("[restaurants/search]", err);
    return NextResponse.json(
      {
        error: "Search failed",
        results: [],
        meta: { placesEnabled: false, inventoryEnabled: false },
      },
      {
        status: status === 429 ? 429 : 500,
        headers: {
          "Cache-Control": "private, no-store",
          ...(status === 429 && retryAfter ? { "Retry-After": retryAfter } : {}),
        },
      },
    );
  }
}
