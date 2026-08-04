import { NextRequest, NextResponse } from "next/server";
import { idleTimeoutMs, SESSION_COOKIE } from "@/lib/config";
import { decodeSession, encodeSession, isSessionActive } from "@/lib/session-token";

const PUBLIC_PATHS = new Set(["/login"]);

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  if (
    pathname.startsWith("/_next") ||
    pathname.startsWith("/audio-processor") ||
    pathname === "/favicon.ico"
  ) {
    return NextResponse.next();
  }

  const token = request.cookies.get(SESSION_COOKIE)?.value;
  const session = await decodeSession(token);
  const active = session ? isSessionActive(session) : false;
  const isLoginApi = pathname === "/api/auth/login";
  const isPublic = PUBLIC_PATHS.has(pathname) || isLoginApi;

  if (!active && !isPublic) {
    if (pathname.startsWith("/api/")) {
      const res = NextResponse.json({ error: "Session expired" }, { status: 401 });
      res.cookies.delete(SESSION_COOKIE);
      return res;
    }
    const url = request.nextUrl.clone();
    url.pathname = "/login";
    url.searchParams.set("reason", session ? "idle" : "auth");
    url.searchParams.set("next", pathname);
    const res = NextResponse.redirect(url);
    res.cookies.delete(SESSION_COOKIE);
    return res;
  }

  if (active && (pathname === "/login" || pathname === "/")) {
    const url = request.nextUrl.clone();
    url.pathname = "/properties";
    return NextResponse.redirect(url);
  }

  const res = NextResponse.next();
  if (active && session && !pathname.startsWith("/api/")) {
    const maxAge = Math.ceil(idleTimeoutMs() / 1000);
    res.cookies.set(SESSION_COOKIE, await encodeSession(session), {
      httpOnly: true,
      sameSite: "lax",
      secure: process.env.NODE_ENV === "production",
      path: "/",
      maxAge,
    });
  }
  return res;
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
