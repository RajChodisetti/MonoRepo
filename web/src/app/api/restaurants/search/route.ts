import { NextResponse } from "next/server";
import { searchSeoRestaurants } from "@/lib/places";

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
    const payload = await searchSeoRestaurants(q, location);
    return NextResponse.json(payload);
  } catch (err) {
    console.error("[restaurants/search]", err);
    return NextResponse.json(
      {
        error: "Search failed",
        results: [],
        meta: { placesEnabled: false, inventoryEnabled: false },
      },
      { status: 500 },
    );
  }
}
