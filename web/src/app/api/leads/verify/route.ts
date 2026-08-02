import { NextResponse } from "next/server";

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
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      cache: "no-store",
    });
    const json = await res.json().catch(() => ({}));
    return NextResponse.json(json, { status: res.status });
  } catch (err) {
    console.error("[leads/verify]", err);
    return NextResponse.json({ error: "Failed to verify code" }, { status: 500 });
  }
}
