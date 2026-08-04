import { NextResponse } from "next/server";
import { adminCredentials } from "@/lib/config";
import { setSession } from "@/lib/session";

export async function POST(request: Request) {
  let body: { email?: string; password?: string };
  try {
    body = (await request.json()) as { email?: string; password?: string };
  } catch {
    return NextResponse.json({ error: "Invalid JSON" }, { status: 400 });
  }

  const email = (body.email || "").trim().toLowerCase();
  const password = body.password || "";
  const expected = adminCredentials();

  if (email !== expected.email || password !== expected.password) {
    return NextResponse.json({ error: "Invalid email or password" }, { status: 401 });
  }

  await setSession({ email, lastActive: Date.now() });
  return NextResponse.json({ ok: true, email });
}
