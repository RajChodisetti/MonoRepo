import { NextResponse } from "next/server";

export const dynamic = "force-dynamic";

function monorepoBase(): string {
  return (process.env.MONOREPO_API_URL?.trim() || "http://localhost:8080").replace(/\/$/, "");
}

/** Proxies Google Places listing photos via MonoRepo (API key stays server-side). */
export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const name = (searchParams.get("name") || "").trim();
  if (!name || !name.startsWith("places/") || !name.includes("/photos/")) {
    return NextResponse.json({ error: "Invalid photo name" }, { status: 400 });
  }
  const max = (searchParams.get("max") || "720").trim();

  try {
    const qs = new URLSearchParams({ name, max });
    const res = await fetch(`${monorepoBase()}/api/public/v1/seo/photo?${qs.toString()}`, {
      cache: "no-store",
    });
    if (!res.ok) {
      return NextResponse.json({ error: "Photo unavailable" }, { status: res.status === 404 ? 404 : 502 });
    }
    const buf = await res.arrayBuffer();
    const contentType = res.headers.get("content-type") || "image/jpeg";
    return new NextResponse(buf, {
      status: 200,
      headers: {
        "Content-Type": contentType,
        "Cache-Control": "public, max-age=86400, stale-while-revalidate=604800",
      },
    });
  } catch (err) {
    console.error("[restaurants/photo]", err);
    return NextResponse.json({ error: "Failed to load photo" }, { status: 500 });
  }
}
