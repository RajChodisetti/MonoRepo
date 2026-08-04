import { NextResponse } from "next/server";
import { agentFetch } from "@/lib/agent";
import { requireSession, setSession, touchSession } from "@/lib/session";

export async function POST(request: Request) {
  try {
    const session = await requireSession();
    await setSession(touchSession(session));
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  let body: { to?: string; phone?: string; language?: string };
  try {
    body = (await request.json()) as { to?: string; phone?: string; language?: string };
  } catch {
    return NextResponse.json({ error: "Invalid JSON" }, { status: 400 });
  }

  const to = (body.to || body.phone || "").trim();
  if (!to) {
    return NextResponse.json({ error: "Phone number is required" }, { status: 400 });
  }

  const result = await agentFetch("/call", {
    method: "POST",
    body: {
      to,
      language: body.language || "auto",
      campaign_id: "andre-admin",
      skip_compliance: true,
    },
  });

  if (!result.ok) {
    return NextResponse.json({ error: result.error }, { status: result.status });
  }
  return NextResponse.json(result.data);
}
