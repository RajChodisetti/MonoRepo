import { readFile } from "node:fs/promises";
import path from "node:path";
import { renderToBuffer } from "@react-pdf/renderer";
import { cookies } from "next/headers";
import { getSeoReport } from "@/lib/places";
import { RestaurantReportPDF } from "@/lib/report-pdf";
import { REPORT_UNLOCK_COOKIE } from "@/lib/report-unlock";
import { seoForwardingHeaders } from "@/lib/seo-forwarding";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

type Params = { params: Promise<{ placeId: string }> };

function safeFilename(name: string): string {
  const normalized = name
    .normalize("NFKD")
    .replace(/[^a-zA-Z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .toLowerCase();
  return `${normalized || "restaurant"}-tuvi-venue-pulse.pdf`;
}

export async function GET(request: Request, { params }: Params) {
  const { placeId: rawPlaceId } = await params;
  const placeId = decodeURIComponent(rawPlaceId || "").trim();
  if (!placeId) {
    return Response.json({ error: "Missing placeId" }, { status: 400 });
  }

  const cookieStore = await cookies();
  const unlockToken = cookieStore.get(REPORT_UNLOCK_COOKIE)?.value?.trim() || "";
  if (!unlockToken) {
    return Response.json(
      { error: "Verify your email to download this report." },
      {
        status: 401,
        headers: { "Cache-Control": "private, no-store" },
      },
    );
  }

  try {
    const payload = await getSeoReport(placeId, unlockToken, seoForwardingHeaders(request));
    if (payload.report.fullReportLocked !== false) {
      cookieStore.delete(REPORT_UNLOCK_COOKIE);
      return Response.json(
        { error: "This report unlock is invalid or expired." },
        {
          status: 403,
          headers: { "Cache-Control": "private, no-store" },
        },
      );
    }

    const logo = await readFile(
      path.join(process.cwd(), "public", "brand", "tuvi-solutions-logo.png"),
    );
    const generatedAt = new Intl.DateTimeFormat("en-AU", {
      dateStyle: "medium",
      timeZone: "Australia/Sydney",
    }).format(new Date());
    const pdf = await renderToBuffer(
      <RestaurantReportPDF
        place={payload.place}
        report={payload.report}
        logoDataUrl={`data:image/png;base64,${logo.toString("base64")}`}
        generatedAt={generatedAt}
      />,
    );

    return new Response(new Uint8Array(pdf), {
      status: 200,
      headers: {
        "Content-Type": "application/pdf",
        "Content-Disposition": `attachment; filename="${safeFilename(payload.report.restaurantName)}"`,
        "Cache-Control": "private, no-store, max-age=0",
        Pragma: "no-cache",
        "X-Content-Type-Options": "nosniff",
        "X-Robots-Tag": "noindex, nofollow, noarchive",
      },
    });
  } catch (error) {
    const status = (error as Error & { status?: number }).status;
    const retryAfter = (error as Error & { retryAfter?: string }).retryAfter;
    const retryable = status === 429 || status === 503;
    if (!retryable) console.error("[report.pdf] generation failed", error);
    return Response.json(
      {
        error:
          status === 429
            ? "Too many report requests. Please retry the PDF download shortly."
            : status === 503
              ? "The live report is busy. Please retry the PDF download shortly."
              : "The report PDF could not be generated.",
      },
      {
        status: retryable ? status : 500,
        headers: {
          "Cache-Control": "private, no-store",
          ...(retryable && retryAfter ? { "Retry-After": retryAfter } : {}),
        },
      },
    );
  }
}
