import { NextRequest, NextResponse } from "next/server";
import { apiBaseUrl } from "@/lib/api";
import { getSessionToken } from "@/lib/session";

type Ctx = { params: Promise<{ path: string[] }> };

async function proxy(request: NextRequest, ctx: Ctx) {
  const token = await getSessionToken();
  if (!token) {
    return NextResponse.json(
      { error: { message: "Authentication required." } },
      { status: 401 },
    );
  }

  const { path } = await ctx.params;
  const targetPath = `/api/v1/${path.join("/")}`;
  const url = new URL(`${apiBaseUrl()}${targetPath}`);
  request.nextUrl.searchParams.forEach((v, k) => {
    url.searchParams.set(k, v);
  });

  const headers: HeadersInit = {
    Accept: "application/json",
    Authorization: `Bearer ${token}`,
  };

  const method = request.method.toUpperCase();
  let body: ArrayBuffer | undefined;
  if (method !== "GET" && method !== "HEAD") {
    const bytes = await request.arrayBuffer();
    if (bytes.byteLength > 0) {
      body = bytes;
      headers["Content-Type"] =
        request.headers.get("content-type") || "application/json";
    }
  }

  const upstream = await fetch(url.toString(), {
    method,
    headers,
    body,
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
export const POST = proxy;
export const PATCH = proxy;
export const PUT = proxy;
export const DELETE = proxy;
