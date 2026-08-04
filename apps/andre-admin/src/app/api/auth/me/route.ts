import { NextResponse } from "next/server";
import { readSession, setSession, touchSession } from "@/lib/session";

export async function GET() {
  const session = await readSession();
  if (!session) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
  await setSession(touchSession(session));
  return NextResponse.json({
    email: session.email,
    lastActive: session.lastActive,
  });
}
