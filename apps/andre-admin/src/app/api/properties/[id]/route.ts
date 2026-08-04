import { NextResponse } from "next/server";
import { agentFetch } from "@/lib/agent";
import { requireSession, setSession, touchSession } from "@/lib/session";

async function gate() {
  const session = await requireSession();
  await setSession(touchSession(session));
}

type Ctx = { params: Promise<{ id: string }> };

export async function GET(_request: Request, ctx: Ctx) {
  try {
    await gate();
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
  const { id } = await ctx.params;
  const result = await agentFetch(`/api/properties/${encodeURIComponent(id)}`, { auth: false });
  if (!result.ok) {
    return NextResponse.json({ error: result.error }, { status: result.status });
  }
  return NextResponse.json(result.data);
}

export async function PUT(request: Request, ctx: Ctx) {
  try {
    await gate();
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
  const { id } = await ctx.params;
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON" }, { status: 400 });
  }
  const result = await agentFetch(`/api/properties/${encodeURIComponent(id)}`, {
    method: "PUT",
    body,
  });
  if (!result.ok) {
    return NextResponse.json({ error: result.error }, { status: result.status });
  }
  return NextResponse.json(result.data);
}

export async function DELETE(_request: Request, ctx: Ctx) {
  try {
    await gate();
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
  const { id } = await ctx.params;
  const result = await agentFetch(`/api/properties/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!result.ok) {
    return NextResponse.json({ error: result.error }, { status: result.status });
  }
  return NextResponse.json(result.data);
}
