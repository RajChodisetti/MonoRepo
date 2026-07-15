export function getVoiceAgentBaseUrl(): string {
  return (
    process.env.VOICE_AGENT_URL?.replace(/\/$/, "") ||
    process.env.NEXT_PUBLIC_VOICE_AGENT_URL?.replace(/\/$/, "") ||
    "http://localhost:8000"
  );
}

export function getVoiceAgentWsUrl(): string {
  const base = getVoiceAgentBaseUrl();
  const wsBase = base.replace(/^http/i, "ws");
  return `${wsBase}/browser-stream?agent=corporate`;
}
