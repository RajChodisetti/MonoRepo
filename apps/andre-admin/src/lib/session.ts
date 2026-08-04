import { cookies } from "next/headers";
import { idleTimeoutMs, SESSION_COOKIE } from "@/lib/config";
import {
  decodeSession,
  encodeSession,
  isSessionActive,
  touchSession,
  type SessionPayload,
} from "@/lib/session-token";

export type { SessionPayload };
export { decodeSession, encodeSession, isSessionActive, touchSession };

export async function readSession(): Promise<SessionPayload | null> {
  const jar = await cookies();
  const session = await decodeSession(jar.get(SESSION_COOKIE)?.value);
  if (!session || !isSessionActive(session)) return null;
  return session;
}

export async function setSession(session: SessionPayload): Promise<void> {
  const jar = await cookies();
  const maxAge = Math.ceil(idleTimeoutMs() / 1000);
  jar.set(SESSION_COOKIE, await encodeSession(session), {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    maxAge,
  });
}

export async function clearSession(): Promise<void> {
  const jar = await cookies();
  jar.delete(SESSION_COOKIE);
}

export async function requireSession(): Promise<SessionPayload> {
  const session = await readSession();
  if (!session) {
    throw new Error("UNAUTHORIZED");
  }
  return session;
}
