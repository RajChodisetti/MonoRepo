import { getVoiceAgentBaseUrl, type VoiceAgentKind } from "@/lib/voiceAgentConfig";

function parseAgent(raw: string | null): VoiceAgentKind {
  return raw === "real_estate" ? "real_estate" : "corporate";
}

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const kind = parseAgent(searchParams.get("agent"));
  const base = getVoiceAgentBaseUrl(kind);
  const label = kind === "real_estate" ? "Real estate agent" : "Tuvi agent";
  const startHint =
    kind === "real_estate"
      ? "Start it: cd MonoRepo/andre-voice-agent && uvicorn bot:app --host 0.0.0.0 --port 8001"
      : "Start it: cd voice-sales-agent && docker compose up -d";

  try {
    const res = await fetch(`${base}/readyz/browser`, {
      cache: "no-store",
      signal: AbortSignal.timeout(5000),
    });
    const data = (await res.json().catch(() => ({}))) as {
      status?: string;
      ready?: boolean;
      missing?: string[];
      message?: string;
      agent?: string;
      mode?: string;
    };

    // Andre returns { ready: true }; corporate returns { status: "ready" }.
    const ready =
      res.ok && (data.status === "ready" || data.ready === true || data.status === "ok");

    if (ready) {
      return Response.json(
        {
          status: "ready",
          agent: kind === "real_estate" ? "andre" : data.agent || "corporate",
          mode: data.mode || (kind === "real_estate" ? "browser" : "corporate"),
        },
        { status: 200 },
      );
    }

    return Response.json(
      {
        status: data.status || "unavailable",
        missing: data.missing,
        message: data.message || `${label} is not ready.`,
        agent: kind,
      },
      { status: res.status >= 400 ? res.status : 503 },
    );
  } catch {
    return Response.json(
      {
        status: "unavailable",
        missing: ["VOICE_AGENT_SERVER"],
        message: `${label} server is not running. ${startHint}`,
        agent: kind,
      },
      { status: 503 },
    );
  }
}
