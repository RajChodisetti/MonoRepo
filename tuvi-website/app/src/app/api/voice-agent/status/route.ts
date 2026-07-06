import { getVoiceAgentBaseUrl } from "@/lib/voiceAgentConfig";

export async function GET() {
  const base = getVoiceAgentBaseUrl();

  try {
    const res = await fetch(`${base}/readyz/browser`, {
      cache: "no-store",
      signal: AbortSignal.timeout(5000),
    });
    const data = await res.json();
    return Response.json(data, { status: res.status });
  } catch {
    return Response.json(
      {
        status: "unavailable",
        missing: ["VOICE_AGENT_SERVER"],
        message:
          "Voice agent server is not running. Start it: cd voice-sales-agent && docker compose up -d",
      },
      { status: 503 },
    );
  }
}
