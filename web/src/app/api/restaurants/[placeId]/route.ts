import { NextResponse } from "next/server";
import { getSeoReport } from "@/lib/places";

export const dynamic = "force-dynamic";

type Params = { params: Promise<{ placeId: string }> };

export async function GET(request: Request, { params }: Params) {
  const { placeId: raw } = await params;
  const placeId = decodeURIComponent(raw || "").trim();
  if (!placeId) {
    return NextResponse.json({ error: "Missing placeId" }, { status: 400 });
  }

  const unlock = new URL(request.url).searchParams.get("unlock") || "";

  try {
    const payload = await getSeoReport(placeId, unlock || undefined);
    return NextResponse.json(payload);
  } catch (err) {
    const status = (err as Error & { status?: number }).status;
    if (status === 404 || (err instanceof Error && err.message.includes("not found"))) {
      return NextResponse.json({ error: "Restaurant not found" }, { status: 404 });
    }
    console.error("[restaurants/details]", err);
    return NextResponse.json({ error: "Failed to load restaurant" }, { status: 500 });
  }
}
