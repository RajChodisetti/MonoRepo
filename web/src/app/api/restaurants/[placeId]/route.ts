import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { getSeoReport } from "@/lib/places";
import { REPORT_UNLOCK_COOKIE } from "@/lib/report-unlock";
import { seoForwardingHeaders } from "@/lib/seo-forwarding";

export const dynamic = "force-dynamic";

type Params = { params: Promise<{ placeId: string }> };

export async function GET(request: Request, { params }: Params) {
  const { placeId: raw } = await params;
  const placeId = decodeURIComponent(raw || "").trim();
  if (!placeId) {
    return NextResponse.json(
      { error: "Missing placeId" },
      { status: 400, headers: { "Cache-Control": "private, no-store" } },
    );
  }

  const cookieStore = await cookies();
  const unlock = cookieStore.get(REPORT_UNLOCK_COOKIE)?.value?.trim() || "";

  try {
    const payload = await getSeoReport(
      placeId,
      unlock || undefined,
      seoForwardingHeaders(request),
    );
    if (unlock && payload.report.fullReportLocked !== false) {
      cookieStore.delete(REPORT_UNLOCK_COOKIE);
    }
    return NextResponse.json(payload, {
      headers: { "Cache-Control": "private, no-store" },
    });
  } catch (err) {
    const status = (err as Error & { status?: number }).status;
    const retryAfter = (err as Error & { retryAfter?: string }).retryAfter;
    if (status === 404 || (err instanceof Error && err.message.includes("not found"))) {
      return NextResponse.json(
        { error: "Restaurant not found" },
        { status: 404, headers: { "Cache-Control": "private, no-store" } },
      );
    }
    if (status === 429) {
      return NextResponse.json(
        { error: "Too many report requests. Please retry shortly." },
        {
          status: 429,
          headers: {
            "Cache-Control": "private, no-store",
            ...(retryAfter ? { "Retry-After": retryAfter } : {}),
          },
        },
      );
    }
    if (status === 503) {
      return NextResponse.json(
        { error: "The live scan is busy. Please retry shortly." },
        {
          status: 503,
          headers: {
            "Cache-Control": "private, no-store",
            ...(retryAfter ? { "Retry-After": retryAfter } : {}),
          },
        },
      );
    }
    console.error("[restaurants/details]", err);
    return NextResponse.json(
      { error: "Failed to load restaurant" },
      { status: 500, headers: { "Cache-Control": "private, no-store" } },
    );
  }
}
