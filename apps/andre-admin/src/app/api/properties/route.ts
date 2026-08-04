import { NextResponse } from "next/server";
import { agentFetch } from "@/lib/agent";
import { requireSession, setSession, touchSession } from "@/lib/session";

async function gate() {
  const session = await requireSession();
  await setSession(touchSession(session));
  return session;
}

export async function GET() {
  try {
    await gate();
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
  const result = await agentFetch<{ results?: unknown[]; count?: number }>("/api/properties");
  if (!result.ok) {
    return NextResponse.json({ error: result.error }, { status: result.status });
  }
  return NextResponse.json(result.data);
}

export async function POST(request: Request) {
  try {
    await gate();
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON" }, { status: 400 });
  }
  const result = await agentFetch("/api/properties", { method: "POST", body });
  if (!result.ok) {
    return NextResponse.json({ error: result.error }, { status: result.status });
  }
  return NextResponse.json(result.data, { status: 201 });
}
