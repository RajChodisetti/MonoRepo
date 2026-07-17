import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const SESSION_COOKIE = "tuvi_admin_token";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get(SESSION_COOKIE)?.value;

  if (pathname.startsWith("/api/")) {
    return NextResponse.next();
  }

  const isAuthPage = pathname === "/login";

  // Neither request.nextUrl.clone() nor request.nextUrl.basePath reliably
  // reapplies/exposes basePath for NextResponse.redirect's Location header
  // in this Next.js version (verified empirically against a built
  // container) — read the same build-time env var next.config.ts uses
  // instead of relying on either.
  const base = process.env.NEXT_PUBLIC_BASE_PATH || "";

  if (!token && !isAuthPage && pathname !== "/") {
    const url = request.nextUrl.clone();
    url.pathname = `${base}/login`;
    url.searchParams.set("next", pathname);
    return NextResponse.redirect(url);
  }

  if (token && (pathname === "/login" || pathname === "/")) {
    const url = request.nextUrl.clone();
    url.pathname = `${base}/dashboard`;
    return NextResponse.redirect(url);
  }

  if (!token && pathname === "/") {
    const url = request.nextUrl.clone();
    url.pathname = `${base}/login`;
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
