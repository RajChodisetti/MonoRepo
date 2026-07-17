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

  // request.nextUrl.clone() does not reliably reapply basePath to the
  // Location header on NextResponse.redirect in this Next.js version, so
  // the basePath is prefixed explicitly here rather than relied on
  // implicitly.
  const base = request.nextUrl.basePath || "";

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
