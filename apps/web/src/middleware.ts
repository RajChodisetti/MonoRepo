import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const SESSION_COOKIE = "tuvi_admin_token";

// Empirically, in this Next.js version request.nextUrl.pathname in
// middleware includes the basePath prefix (it is NOT stripped), and
// request.nextUrl.clone()/`.basePath` do not reliably reapply it to
// NextResponse.redirect's Location header either. So every comparison and
// redirect target here is made explicitly basePath-aware using the same
// build-time env var next.config.ts uses, rather than relying on any
// implicit Next.js basePath handling.
const BASE = process.env.NEXT_PUBLIC_BASE_PATH || "";
const LOGIN_PATH = `${BASE}/login`;
const DASHBOARD_PATH = `${BASE}/dashboard`;
const ROOT_PATH = BASE || "/";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get(SESSION_COOKIE)?.value;

  if (pathname.startsWith(`${BASE}/api/`)) {
    return NextResponse.next();
  }

  const isAuthPage = pathname === LOGIN_PATH;
  const isRootPage = pathname === ROOT_PATH || pathname === `${ROOT_PATH}/`;

  if (!token && !isAuthPage && !isRootPage) {
    const url = request.nextUrl.clone();
    url.pathname = LOGIN_PATH;
    url.searchParams.set("next", pathname);
    return NextResponse.redirect(url);
  }

  if (token && (isAuthPage || isRootPage)) {
    const url = request.nextUrl.clone();
    url.pathname = DASHBOARD_PATH;
    return NextResponse.redirect(url);
  }

  if (!token && isRootPage) {
    const url = request.nextUrl.clone();
    url.pathname = LOGIN_PATH;
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
