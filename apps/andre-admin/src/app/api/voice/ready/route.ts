import { NextResponse } from "next/server";
import { agentFetch } from "@/lib/agent";
import { requireSession, setSession, touchSession } from "@/lib/session";

export async function GET() {
  try {
    const session = await requireSession();
    await setSession(touchSession(session));
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const [browser, phone] = await Promise.all([
    agentFetch("/readyz/browser", { auth: false }),
    agentFetch("/readyz", { auth: false }),
  ]);

  return NextResponse.json({
    browserReady: browser.ok,
    phoneReady: phone.ok,
    browser: browser.data,
    phone: phone.data,
    browserError: browser.error,
    phoneError: phone.error,
  });
}
