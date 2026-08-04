import { NextResponse } from "next/server";
import { clearSession, readSession, setSession, touchSession } from "@/lib/session";

export async function POST() {
  const session = await readSession();
  if (!session) {
    await clearSession();
    return NextResponse.json({ error: "Session expired" }, { status: 401 });
  }
  const next = touchSession(session);
  await setSession(next);
  return NextResponse.json({ ok: true, lastActive: next.lastActive });
}
