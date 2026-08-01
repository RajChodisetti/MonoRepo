import { NextRequest, NextResponse } from "next/server";
import { apiBaseUrl } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

type Ctx = { params: Promise<{ path: string[] }> };

/** Proxies authenticated admin session to public read endpoints (site images/content). */
async function proxy(request: NextRequest, ctx: Ctx) {
  const token = await getSessionToken();
  if (!token) {
    return NextResponse.json(
      { error: { message: "Authentication required." } },
      { status: 401 },
    );
  }

  const { path } = await ctx.params;
  const targetPath = `/api/public/v1/${path.join("/")}`;
  const url = new URL(`${apiBaseUrl()}${targetPath}`);
  request.nextUrl.searchParams.forEach((v, k) => {
    url.searchParams.set(k, v);
  });

  const upstream = await fetch(url.toString(), {
    method: "GET",
    headers: { Accept: "application/json" },
    cache: "no-store",
  });

  const respText = await upstream.text();
  return new NextResponse(respText || null, {
    status: upstream.status,
    headers: {
      "Content-Type":
        upstream.headers.get("content-type") || "application/json",
    },
  });
}

export const GET = proxy;
